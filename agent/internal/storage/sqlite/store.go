package sqlite

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	_ "modernc.org/sqlite"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

type Store struct {
	db *sql.DB
}

type StoredEvent struct {
	LocalID int64
	Event   event.Event
}

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Open(dataDir string) (*Store, error) {
	if dataDir == "" {
		dataDir = constants.DefaultDataDir
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		return nil, fmt.Errorf("create data dir: %w", err)
	}

	db, err := sql.Open("sqlite", filepath.Join(dataDir, constants.DefaultSQLiteFile))
	if err != nil {
		return nil, fmt.Errorf("open sqlite store: %w", err)
	}

	store := &Store{db: db}
	if err := store.Migrate(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Migrate(ctx context.Context) error {
	paths, err := fs.Glob(migrationFiles, constants.SQLiteMigrationGlob)
	if err != nil {
		return fmt.Errorf("load sqlite migration manifest: %w", err)
	}
	sort.Strings(paths)

	for _, path := range paths {
		migration, err := migrationFiles.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read sqlite migration %s: %w", path, err)
		}
		if _, err := s.db.ExecContext(ctx, string(migration)); err != nil {
			return fmt.Errorf("apply sqlite migration %s: %w", path, err)
		}
	}

	return nil
}

func (s *Store) SaveEvent(ctx context.Context, evt event.Event) error {
	metadata, err := json.Marshal(evt.Metadata)
	if err != nil {
		return fmt.Errorf("marshal event metadata: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
INSERT INTO events (
  event_type, source, observed_at, tenant_id, device_id, host_name,
  app_name, process_id, path_hash, metadata_json
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		evt.Type,
		evt.Source,
		evt.Timestamp.UTC().Format(time.RFC3339Nano),
		evt.TenantID,
		evt.DeviceID,
		evt.HostName,
		evt.AppName,
		evt.ProcessID,
		evt.PathHash,
		string(metadata),
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	return nil
}

func (s *Store) CountEvents(ctx context.Context) (int, error) {
	var count int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count events: %w", err)
	}
	return count, nil
}

func (s *Store) BackendSyncCursor(ctx context.Context, syncName string) (int64, error) {
	if syncName == "" {
		syncName = constants.BackendSyncCursorName
	}
	var cursor int64
	err := s.db.QueryRowContext(ctx, `SELECT last_event_id FROM backend_sync_state WHERE sync_name = ?`, syncName).Scan(&cursor)
	if err == nil {
		return cursor, nil
	}
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return 0, fmt.Errorf("read backend sync cursor: %w", err)
}

func (s *Store) PendingBackendSyncEvents(ctx context.Context, afterID int64, limit int) ([]StoredEvent, error) {
	if limit <= 0 {
		limit = constants.DefaultBackendSyncBatchLimit
	}
	rows, err := s.db.QueryContext(ctx, `
SELECT id, event_type, source, observed_at, tenant_id, device_id, host_name,
       app_name, process_id, path_hash, metadata_json
FROM events
WHERE id > ?
ORDER BY id ASC
LIMIT ?`, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("query pending backend sync events: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	events := make([]StoredEvent, 0, limit)
	for rows.Next() {
		var stored StoredEvent
		var observedAt string
		var processID sql.NullInt64
		var pathHash sql.NullString
		var metadataJSON string
		if err := rows.Scan(
			&stored.LocalID,
			&stored.Event.Type,
			&stored.Event.Source,
			&observedAt,
			&stored.Event.TenantID,
			&stored.Event.DeviceID,
			&stored.Event.HostName,
			&stored.Event.AppName,
			&processID,
			&pathHash,
			&metadataJSON,
		); err != nil {
			return nil, fmt.Errorf("scan pending backend sync event: %w", err)
		}
		parsedAt, err := time.Parse(time.RFC3339Nano, observedAt)
		if err != nil {
			return nil, fmt.Errorf("parse pending event observed_at: %w", err)
		}
		metadata := map[string]string{}
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("decode pending event metadata: %w", err)
		}
		stored.Event.ID = fmt.Sprintf("%s%d", constants.BackendSyncEventIDPrefix, stored.LocalID)
		stored.Event.Timestamp = parsedAt.UTC()
		if processID.Valid {
			stored.Event.ProcessID = int32(processID.Int64)
		}
		if pathHash.Valid {
			stored.Event.PathHash = pathHash.String
		}
		stored.Event.Metadata = metadata
		events = append(events, stored)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate pending backend sync events: %w", err)
	}
	return events, nil
}

func (s *Store) MarkBackendSyncCursor(ctx context.Context, syncName string, lastEventID int64) error {
	if syncName == "" {
		syncName = constants.BackendSyncCursorName
	}
	if lastEventID <= 0 {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `
INSERT INTO backend_sync_state (sync_name, last_event_id, updated_at)
VALUES (?, ?, ?)
ON CONFLICT(sync_name) DO UPDATE SET
  last_event_id = excluded.last_event_id,
  updated_at = excluded.updated_at
WHERE excluded.last_event_id > backend_sync_state.last_event_id`,
		syncName,
		lastEventID,
		time.Now().UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return fmt.Errorf("mark backend sync cursor: %w", err)
	}
	return nil
}

func (s *Store) EnforceRetention(ctx context.Context, ttlDays int) error {
	if ttlDays <= 0 {
		return nil
	}
	cutoff := time.Now().UTC().AddDate(0, 0, -ttlDays).Format(time.RFC3339Nano)
	if _, err := s.db.ExecContext(ctx, `DELETE FROM events WHERE observed_at < ?`, cutoff); err != nil {
		return fmt.Errorf("enforce retention: %w", err)
	}
	return nil
}
