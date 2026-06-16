package exporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

func TestOTLPHTTPLogExporterPostsMetadataOnlyLogs(t *testing.T) {
	t.Parallel()

	var received map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/logs" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != constants.OpenTelemetryContentTypeJSON {
			t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	exporter, err := NewOTLPHTTPLogExporter(config.OpenTelemetryPolicy{
		Endpoint:       server.URL + "/v1/logs",
		RequestTimeout: "2s",
		Retry:          config.OpenTelemetryRetryPolicy{MaxAttempts: 1},
	})
	if err != nil {
		t.Fatalf("new exporter: %v", err)
	}

	result, err := exporter.Export(context.Background(), Request{
		Policy:   samplePolicy(),
		HostName: "study-laptop",
		OSName:   "windows",
		Events: []event.Event{
			{
				ID:        "otel-event-1",
				Type:      constants.EventTypeBrowserObserved,
				Source:    constants.EventSourceBrowserCollector,
				Timestamp: time.Date(2026, 6, 16, 9, 0, 0, 0, time.UTC),
				TenantID:  "family-varadha",
				DeviceID:  "laptop-cousin-001",
				HostName:  "study-laptop",
				AppName:   "chrome.exe",
				ProcessID: 1234,
				PathHash:  "sha256:abc",
				Metadata: map[string]string{
					constants.EventMetadataDomain:         "youtube.com",
					constants.EventMetadataCategory:       "coding",
					constants.EventMetadataURLMode:        constants.URLModeDomainOnly,
					"raw_url":                             "https://youtube.com/watch?v=private",
					"page_title":                          "private title",
					"session_token":                       "secret",
					constants.EventMetadataYouTubeStudy:   "true",
					constants.EventMetadataYouTubeVideoID: "sha256:video",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("export logs: %v", err)
	}
	if result.ExportedEvents != 1 || result.DroppedEvents != 0 || result.Attempts != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}

	payload, err := json.Marshal(received)
	if err != nil {
		t.Fatalf("marshal received payload: %v", err)
	}
	text := string(payload)
	for _, expected := range []string{
		constants.OpenTelemetryPrivacyBoundary,
		constants.EventTypeBrowserObserved,
		"youtube.com",
		"coding",
		"sha256:video",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected %q in payload: %s", expected, text)
		}
	}
	for _, forbidden := range []string{
		"https://youtube.com",
		"private title",
		"session_token",
		"secret",
		"raw_url",
		"page_title",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("payload leaked forbidden value %q: %s", forbidden, text)
		}
	}
}

func TestOTLPHTTPLogExporterDropsAfterBoundedAttempts(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		http.Error(w, "collector unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	exporter, err := NewOTLPHTTPLogExporter(config.OpenTelemetryPolicy{
		Endpoint:       server.URL + "/v1/logs",
		RequestTimeout: "2s",
		Retry:          config.OpenTelemetryRetryPolicy{MaxAttempts: 2},
	})
	if err != nil {
		t.Fatalf("new exporter: %v", err)
	}

	result, err := exporter.Export(context.Background(), Request{
		Policy:   samplePolicy(),
		HostName: "study-laptop",
		OSName:   "windows",
		Events: []event.Event{
			{ID: "otel-event-1", Type: constants.EventTypeAgentHeartbeat, Timestamp: time.Now().UTC()},
			{ID: "otel-event-2", Type: constants.EventTypeDeviceHealth, Timestamp: time.Now().UTC()},
		},
	})
	if err == nil {
		t.Fatal("expected failed collector to return an error")
	}
	if attempts != 2 || result.Attempts != 2 {
		t.Fatalf("expected two bounded attempts, attempts=%d result=%+v", attempts, result)
	}
	if result.DroppedEvents != 2 || result.ExportedEvents != 0 {
		t.Fatalf("expected failed batch to be counted as dropped: %+v", result)
	}
}

func samplePolicy() *config.Policy {
	return &config.Policy{
		TenantID: "family-varadha",
		DeviceID: "laptop-cousin-001",
		Profile:  "ai-btech-student",
	}
}
