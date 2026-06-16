package heartbeat

import (
	"context"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
)

type stubPlatform struct{}

func (stubPlatform) Name() string { return constants.OperatingSystemWindows }

func (stubPlatform) Hostname(context.Context) (string, error) { return "study-laptop", nil }

func (stubPlatform) Capabilities() platform.Capabilities {
	return platform.Capabilities{OperatingSystem: constants.OperatingSystemWindows}
}

func (stubPlatform) ForegroundApp(context.Context) (platform.ForegroundApp, error) {
	return platform.ForegroundApp{}, platform.ErrNoForegroundApp
}

func TestCollectorEmitsMetadataOnlyAgentHeartbeat(t *testing.T) {
	t.Parallel()

	collector := New(stubPlatform{}, Options{
		ContinuousMode:     true,
		CollectionInterval: "10m",
		ArchiveEnabled:     true,
		ArchiveDue:         true,
		BackendSyncEnabled: true,
		AlertsEnabled:      true,
	})

	events, err := collector.Collect(context.Background(), &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Profile:  constants.DefaultProfile,
	})
	if err != nil {
		t.Fatalf("collect heartbeat: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one heartbeat event, got %d", len(events))
	}

	evt := events[0]
	if evt.Type != constants.EventTypeAgentHeartbeat || evt.Source != constants.EventSourceHeartbeat {
		t.Fatalf("unexpected heartbeat identity: %+v", evt)
	}
	if evt.HostName != "study-laptop" || evt.AppName != constants.AppName {
		t.Fatalf("expected host and agent app labels, got %+v", evt)
	}
	if evt.Metadata[constants.EventMetadataAgentHealthy] != "true" ||
		evt.Metadata[constants.EventMetadataAgentVersion] != constants.AppVersion ||
		evt.Metadata[constants.EventMetadataCollectionMode] != constants.HeartbeatCollectionModeContinuous ||
		evt.Metadata[constants.EventMetadataArchiveDue] != "true" ||
		evt.Metadata[constants.EventMetadataBackendSync] != "true" {
		t.Fatalf("expected heartbeat readiness metadata, got %+v", evt.Metadata)
	}
	for _, forbidden := range []string{"password", "screenshot", "raw_url", "page_title", "cookie", "token"} {
		if _, ok := evt.Metadata[forbidden]; ok {
			t.Fatalf("heartbeat should not expose forbidden metadata %q: %+v", forbidden, evt.Metadata)
		}
	}
}
