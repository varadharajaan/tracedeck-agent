package platform

import (
	"context"
	"errors"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestCurrentAdapterReportsCapabilities(t *testing.T) {
	t.Parallel()

	adapter := Current()
	if adapter.Name() == "" {
		t.Fatal("platform adapter name is required")
	}

	caps := adapter.Capabilities()
	if caps.OperatingSystem == "" {
		t.Fatal("operating system capability is required")
	}
	if !caps.LocalStorage {
		t.Fatal("local storage must be supported for the local agent")
	}

	if _, err := adapter.Hostname(context.Background()); err != nil {
		t.Fatalf("hostname: %v", err)
	}
}

func TestPlatformCapabilityCatalogs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		caps           Capabilities
		serviceManager string
	}{
		{
			name:           constants.OperatingSystemWindows,
			caps:           WindowsCapabilities(),
			serviceManager: constants.ServiceManagerWindowsService,
		},
		{
			name:           constants.OperatingSystemMacOS,
			caps:           DarwinCapabilities(),
			serviceManager: constants.ServiceManagerLaunchd,
		},
		{
			name:           constants.OperatingSystemLinux,
			caps:           LinuxCapabilities(),
			serviceManager: constants.ServiceManagerSystemd,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.caps.ServiceManager != tt.serviceManager {
				t.Fatalf("expected service manager %q, got %q", tt.serviceManager, tt.caps.ServiceManager)
			}
			if err := tt.caps.Require(constants.PlatformCapabilityProcessCollection); err != nil {
				t.Fatalf("process collection should be supported: %v", err)
			}
			if err := tt.caps.Require(constants.PlatformCapabilityLocalStorage); err != nil {
				t.Fatalf("local storage should be supported: %v", err)
			}
		})
	}
}

func TestUnsupportedCapabilityErrorsAreTyped(t *testing.T) {
	t.Parallel()

	darwinCaps := DarwinCapabilities()
	err := darwinCaps.Require(constants.PlatformCapabilityForegroundApp)
	if err == nil {
		t.Fatal("expected macOS foreground app capability to require permission")
	}
	if !errors.Is(err, ErrUnsupportedCapability) {
		t.Fatalf("expected typed unsupported capability error, got %v", err)
	}

	var capabilityErr CapabilityError
	if !errors.As(err, &capabilityErr) {
		t.Fatalf("expected capability error details, got %T", err)
	}
	if capabilityErr.Status != constants.PlatformSupportRequiresPermission {
		t.Fatalf("expected requires_permission status, got %s", capabilityErr.Status)
	}

	linuxCaps := LinuxCapabilities()
	foreground, ok := linuxCaps.SupportFor(constants.PlatformCapabilityForegroundApp)
	if !ok {
		t.Fatal("expected linux foreground app capability to be declared")
	}
	if foreground.Status != constants.PlatformSupportPartial {
		t.Fatalf("expected linux foreground support to be partial, got %s", foreground.Status)
	}
}
