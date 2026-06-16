package activewindow

import (
	"context"
	"errors"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
)

type fakePlatform struct {
	foreground platform.ForegroundApp
	err        error
}

func (fakePlatform) Name() string { return constants.OperatingSystemWindows }

func (fakePlatform) Hostname(context.Context) (string, error) { return "study-host", nil }

func (fakePlatform) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		OperatingSystem:   constants.OperatingSystemWindows,
		ProcessCollection: true,
		LocalStorage:      true,
		Features: []platform.CapabilitySupport{{
			ID:     constants.PlatformCapabilityForegroundApp,
			Status: constants.PlatformSupportSupported,
		}},
	}
}

func (f fakePlatform) ForegroundApp(context.Context) (platform.ForegroundApp, error) {
	return f.foreground, f.err
}

func TestCollectorEmitsMetadataOnlyForegroundAppEvent(t *testing.T) {
	t.Parallel()

	events, err := New(fakePlatform{
		foreground: platform.ForegroundApp{
			AppName:        "Code.exe",
			ProcessID:      42,
			ExecutablePath: `C:\Users\student\AppData\Local\Programs\Microsoft VS Code\Code.exe`,
		},
	}).Collect(context.Background(), foregroundPolicy(true))
	if err != nil {
		t.Fatalf("collect foreground app: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one foreground app event, got %d", len(events))
	}

	evt := events[0]
	if evt.Type != constants.EventTypeForegroundAppObserved || evt.Source != constants.EventSourceForegroundAppCollector {
		t.Fatalf("unexpected foreground event identity: %+v", evt)
	}
	if evt.AppName != "Code.exe" || evt.ProcessID != 42 || evt.HostName != "study-host" {
		t.Fatalf("unexpected foreground app event: %+v", evt)
	}
	if evt.PathHash == "" || evt.PathHash == `C:\Users\student\AppData\Local\Programs\Microsoft VS Code\Code.exe` {
		t.Fatalf("expected hashed executable path only, got %q", evt.PathHash)
	}
	if evt.Metadata[constants.EventMetadataForegroundState] != constants.ForegroundStateActive ||
		evt.Metadata[constants.EventMetadataWindowTitleMode] != constants.WindowTitleModeNone ||
		evt.Metadata[constants.EventMetadataPathMode] != constants.PathModeHashOnly ||
		evt.Metadata[constants.EventMetadataProfile] != constants.DefaultProfile {
		t.Fatalf("expected typed foreground metadata, got %+v", evt.Metadata)
	}
	for _, forbidden := range []string{"password", "screenshot", "raw_url", "page_title", "window_title", "cookie", "token"} {
		if _, ok := evt.Metadata[forbidden]; ok {
			t.Fatalf("foreground metadata contains forbidden key %q: %+v", forbidden, evt.Metadata)
		}
	}
}

func TestCollectorClassifiesRiskyForegroundApps(t *testing.T) {
	t.Parallel()

	events, err := New(fakePlatform{
		foreground: platform.ForegroundApp{
			AppName:        "steam.exe",
			ProcessID:      99,
			ExecutablePath: `C:\Program Files (x86)\Steam\steam.exe`,
		},
	}).Collect(context.Background(), foregroundPolicy(true))
	if err != nil {
		t.Fatalf("collect foreground app: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected one foreground app event, got %d", len(events))
	}
	if events[0].Metadata[constants.EventMetadataSoftwareRiskCategory] != constants.SoftwareRiskCategoryGameLauncher {
		t.Fatalf("expected game launcher risk metadata, got %+v", events[0].Metadata)
	}
}

func TestCollectorSkipsDisabledOrUnsupportedForegroundCollection(t *testing.T) {
	t.Parallel()

	disabled, err := New(fakePlatform{
		foreground: platform.ForegroundApp{AppName: "Code.exe", ProcessID: 42},
	}).Collect(context.Background(), foregroundPolicy(false))
	if err != nil {
		t.Fatalf("disabled foreground collection should not fail: %v", err)
	}
	if len(disabled) != 0 {
		t.Fatalf("expected disabled collector to emit no events, got %d", len(disabled))
	}

	unsupported, err := New(fakePlatform{
		err: platform.CapabilityError{
			OperatingSystem: constants.OperatingSystemMacOS,
			CapabilityID:    constants.PlatformCapabilityForegroundApp,
			Status:          constants.PlatformSupportRequiresPermission,
			Reason:          "permission not granted",
		},
	}).Collect(context.Background(), foregroundPolicy(true))
	if err != nil {
		t.Fatalf("unsupported foreground collection should not fail: %v", err)
	}
	if len(unsupported) != 0 {
		t.Fatalf("expected unsupported foreground adapter to emit no events, got %d", len(unsupported))
	}
}

func TestCollectorSurfacesUnexpectedForegroundErrors(t *testing.T) {
	t.Parallel()

	_, err := New(fakePlatform{err: errors.New("desktop api failed")}).Collect(context.Background(), foregroundPolicy(true))
	if err == nil {
		t.Fatal("expected unexpected foreground errors to fail collection")
	}
}

func foregroundPolicy(enabled bool) *config.Policy {
	return &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Profile:  constants.DefaultProfile,
		Collection: config.CollectionPolicy{
			ForegroundApp: config.ForegroundAppCollection{
				Enabled:         enabled,
				WindowTitleMode: config.WindowTitleMode(constants.WindowTitleModeNone),
			},
		},
	}
}
