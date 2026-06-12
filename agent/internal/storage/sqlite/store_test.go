package sqlite

import (
	"context"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

func TestStoreSaveAndCountEvent(t *testing.T) {
	t.Parallel()

	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	}()

	err = store.SaveEvent(context.Background(), event.Event{
		Type:      constants.EventTypeProcessObserved,
		Source:    constants.EventSourceProcessCollector,
		Timestamp: time.Now().UTC(),
		TenantID:  constants.DefaultTenantID,
		DeviceID:  constants.DefaultDeviceID,
		HostName:  constants.DefaultDeviceID,
		AppName:   constants.AppName,
		ProcessID: 42,
		PathHash:  "test-path-hash",
		Metadata: map[string]string{
			constants.EventMetadataProfile:         constants.DefaultProfile,
			constants.EventMetadataOperatingSystem: constants.OperatingSystemWindows,
		},
	})
	if err != nil {
		t.Fatalf("save event: %v", err)
	}

	count, err := store.CountEvents(context.Background())
	if err != nil {
		t.Fatalf("count events: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 event, got %d", count)
	}
}

func TestBackendSyncCursorAndPendingEvents(t *testing.T) {
	t.Parallel()

	store, err := Open(t.TempDir())
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Fatalf("close store: %v", err)
		}
	}()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err = store.SaveEvent(ctx, event.Event{
			Type:      constants.EventTypeProcessObserved,
			Source:    constants.EventSourceProcessCollector,
			Timestamp: time.Date(2026, 6, 12, 8, i, 0, 0, time.UTC),
			TenantID:  constants.DefaultTenantID,
			DeviceID:  constants.DefaultDeviceID,
			HostName:  constants.DefaultDeviceID,
			AppName:   constants.AppName,
			ProcessID: int32(100 + i),
			PathHash:  "hash-only",
			Metadata: map[string]string{
				constants.EventMetadataProfile: constants.DefaultProfile,
			},
		})
		if err != nil {
			t.Fatalf("save event %d: %v", i, err)
		}
	}

	cursor, err := store.BackendSyncCursor(ctx, constants.BackendSyncCursorName)
	if err != nil {
		t.Fatalf("read initial cursor: %v", err)
	}
	if cursor != 0 {
		t.Fatalf("expected empty cursor, got %d", cursor)
	}

	pending, err := store.PendingBackendSyncEvents(ctx, cursor, 2)
	if err != nil {
		t.Fatalf("pending events: %v", err)
	}
	if len(pending) != 2 {
		t.Fatalf("expected bounded pending events, got %d", len(pending))
	}
	if pending[0].LocalID != 1 || pending[0].Event.ID != constants.BackendSyncEventIDPrefix+"1" {
		t.Fatalf("expected first local event id, got %+v", pending[0])
	}
	if pending[1].Event.ProcessID != 101 || pending[1].Event.Metadata[constants.EventMetadataProfile] != constants.DefaultProfile {
		t.Fatalf("expected decoded event fields, got %+v", pending[1])
	}

	if err := store.MarkBackendSyncCursor(ctx, constants.BackendSyncCursorName, pending[1].LocalID); err != nil {
		t.Fatalf("mark cursor: %v", err)
	}
	cursor, err = store.BackendSyncCursor(ctx, constants.BackendSyncCursorName)
	if err != nil {
		t.Fatalf("read updated cursor: %v", err)
	}
	if cursor != 2 {
		t.Fatalf("expected cursor 2, got %d", cursor)
	}

	pending, err = store.PendingBackendSyncEvents(ctx, cursor, 10)
	if err != nil {
		t.Fatalf("pending after cursor: %v", err)
	}
	if len(pending) != 1 || pending[0].LocalID != 3 {
		t.Fatalf("expected one remaining event, got %+v", pending)
	}

	if err := store.MarkBackendSyncCursor(ctx, constants.BackendSyncCursorName, 1); err != nil {
		t.Fatalf("mark lower cursor: %v", err)
	}
	cursor, err = store.BackendSyncCursor(ctx, constants.BackendSyncCursorName)
	if err != nil {
		t.Fatalf("read cursor after lower mark: %v", err)
	}
	if cursor != 2 {
		t.Fatalf("expected cursor not to move backwards, got %d", cursor)
	}
}
