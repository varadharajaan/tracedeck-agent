package app

import (
	"strings"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestParseDurationOrDefault(t *testing.T) {
	t.Parallel()

	duration, err := parseDurationOrDefault("", constants.DefaultCollectionInterval)
	if err != nil {
		t.Fatalf("parse default duration: %v", err)
	}
	if duration != 10*time.Minute {
		t.Fatalf("expected 10m, got %s", duration)
	}

	duration, err = parseDurationOrDefault("1s", constants.DefaultCollectionInterval)
	if err != nil {
		t.Fatalf("parse explicit duration: %v", err)
	}
	if duration != time.Second {
		t.Fatalf("expected 1s, got %s", duration)
	}
}

func TestParseDurationOrDefaultRejectsInvalidDuration(t *testing.T) {
	t.Parallel()

	if _, err := parseDurationOrDefault("soon", constants.DefaultCollectionInterval); err == nil {
		t.Fatal("expected invalid duration to fail")
	}
	if _, err := parseDurationOrDefault("0s", constants.DefaultCollectionInterval); err == nil {
		t.Fatal("expected zero duration to fail")
	}
}

func TestRunResultMerge(t *testing.T) {
	t.Parallel()

	result := RunResult{
		Cycles:           1,
		CollectedEvents:  5,
		StoredEvents:     5,
		ForegroundEvents: 1,
		HealthEvents:     1,
		HeartbeatEvents:  1,
	}
	result.merge(RunResult{
		Cycles:           1,
		CollectedEvents:  7,
		StoredEvents:     12,
		ArchiveBatch:     "archive.jsonl.gz",
		AlertsRaised:     2,
		AlertOutboxPath:  "alert.json",
		ForegroundEvents: 1,
		HealthEvents:     1,
		HeartbeatEvents:  1,
		TelemetrySynced:  true,
		TelemetryEvents:  4,
		TelemetryBacklog: 3,
		OTelExported:     true,
		OTelEvents:       4,
		OTelDropped:      1,
		OTelAttempts:     2,
		OTelBacklog:      0,
	})

	if result.Cycles != 2 {
		t.Fatalf("expected 2 cycles, got %d", result.Cycles)
	}
	if result.CollectedEvents != 12 {
		t.Fatalf("expected 12 collected events, got %d", result.CollectedEvents)
	}
	if result.StoredEvents != 12 {
		t.Fatalf("expected 12 stored events, got %d", result.StoredEvents)
	}
	if result.ArchiveBatch == "" || result.AlertOutboxPath == "" {
		t.Fatalf("expected latest archive and alert paths to be preserved: %+v", result)
	}
	if result.AlertsRaised != 2 {
		t.Fatalf("expected 2 alerts, got %d", result.AlertsRaised)
	}
	if result.HealthEvents != 2 {
		t.Fatalf("expected 2 health events, got %d", result.HealthEvents)
	}
	if result.ForegroundEvents != 2 {
		t.Fatalf("expected 2 foreground events, got %d", result.ForegroundEvents)
	}
	if result.HeartbeatEvents != 2 {
		t.Fatalf("expected 2 heartbeat events, got %d", result.HeartbeatEvents)
	}
	if !result.TelemetrySynced || result.TelemetryEvents != 4 {
		t.Fatalf("expected telemetry sync merge, got %+v", result)
	}
	if result.TelemetryBacklog != 3 {
		t.Fatalf("expected telemetry backlog to reflect latest cycle, got %+v", result)
	}
	if !result.OTelExported || result.OTelEvents != 4 || result.OTelDropped != 1 || result.OTelAttempts != 2 || result.OTelBacklog != 0 {
		t.Fatalf("expected opentelemetry export merge, got %+v", result)
	}
	formatted := FormatRunResult(result)
	for _, expected := range []string{"foreground_events=2", "otel_exported=true", "otel_events=4", "otel_dropped=1", "otel_attempts=2", "otel_backlog=0"} {
		if !strings.Contains(formatted, expected) {
			t.Fatalf("expected %q in formatted result: %s", expected, formatted)
		}
	}
}
