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
