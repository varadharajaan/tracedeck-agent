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
