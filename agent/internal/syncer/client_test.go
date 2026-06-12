package syncer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

func TestClientIngestEventsPostsMetadataOnlyPayload(t *testing.T) {
	t.Parallel()

	var captured ingestRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/devices/device-1/telemetry-events" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(IngestResult{
			TenantID:       captured.TenantID,
			DeviceID:       captured.DeviceID,
			AcceptedEvents: 1,
			StoredEvents:   1,
			LastIngestedAt: time.Now().UTC(),
		})
	}))
	defer server.Close()

	client, err := NewClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	result, err := client.IngestEvents(context.Background(), &config.Policy{
		TenantID: "tenant-1",
		DeviceID: "device-1",
		Profile:  "student",
	}, "host-1", "windows", []event.Event{{
		Type:      "process_snapshot",
		Source:    "process",
		Timestamp: time.Date(2026, 6, 12, 8, 0, 0, 0, time.UTC),
		AppName:   "Code.exe",
		ProcessID: 42,
		PathHash:  "hash-only",
		Metadata:  map[string]string{"category": "coding"},
	}})
	if err != nil {
		t.Fatalf("ingest events: %v", err)
	}
	if result.AcceptedEvents != 1 {
		t.Fatalf("expected accepted event count, got %+v", result)
	}
	if captured.TenantID != "tenant-1" || captured.DeviceID != "device-1" || captured.HostName != "host-1" {
		t.Fatalf("unexpected captured request: %+v", captured)
	}
	if len(captured.Events) != 1 || captured.Events[0].PathHash != "hash-only" || captured.Events[0].Metadata["category"] != "coding" {
		t.Fatalf("expected metadata-only event payload: %+v", captured.Events)
	}
}

func TestNewClientRejectsInvalidURL(t *testing.T) {
	t.Parallel()

	if _, err := NewClient("not a url", time.Second); err == nil {
		t.Fatal("expected invalid URL to fail")
	}
}
