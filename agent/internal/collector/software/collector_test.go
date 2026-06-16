package software

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
)

type fakePlatform struct {
	inventory []platform.InstalledSoftware
	err       error
}

func (fakePlatform) Name() string { return constants.OperatingSystemWindows }

func (fakePlatform) Hostname(context.Context) (string, error) { return "study-host", nil }

func (fakePlatform) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		OperatingSystem: constants.OperatingSystemWindows,
		Features: []platform.CapabilitySupport{{
			ID:     constants.PlatformCapabilitySoftwareInventory,
			Status: constants.PlatformSupportSupported,
		}},
	}
}

func (fakePlatform) ForegroundApp(context.Context) (platform.ForegroundApp, error) {
	return platform.ForegroundApp{}, platform.ErrNoForegroundApp
}

func (f fakePlatform) SoftwareInventory(context.Context) ([]platform.InstalledSoftware, error) {
	return f.inventory, f.err
}

func TestCollectorBaselinesMetadataOnlyInventory(t *testing.T) {
	t.Parallel()

	rawID := `C:\Program Files\TraceDeck Fixture\fixture.exe`
	cacheDir := t.TempDir()
	collector := New(fakePlatform{inventory: []platform.InstalledSoftware{{
		ID:        rawID,
		Name:      "TraceDeck Fixture",
		Version:   "1.0.0",
		Publisher: "TraceDeck Labs",
		Source:    constants.SoftwareSourceWindowsRegistry,
	}}}, cacheDir)
	collector.now = fixedNow

	events, err := collector.Collect(context.Background(), softwarePolicy(true))
	if err != nil {
		t.Fatalf("collect software baseline: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("first run should baseline without change events, got %d", len(events))
	}

	snapshotData, err := os.ReadFile(filepath.Join(cacheDir, constants.SoftwareSnapshotFileName))
	if err != nil {
		t.Fatalf("read snapshot: %v", err)
	}
	if strings.Contains(string(snapshotData), rawID) {
		t.Fatalf("snapshot leaked raw platform identifier: %s", snapshotData)
	}
	if !strings.Contains(string(snapshotData), "TraceDeck Fixture") {
		t.Fatalf("snapshot should retain software display metadata: %s", snapshotData)
	}
}

func TestCollectorEmitsInstallAndUninstallEvents(t *testing.T) {
	t.Parallel()

	cacheDir := t.TempDir()
	policy := softwarePolicy(true)

	baseline := New(fakePlatform{inventory: []platform.InstalledSoftware{
		softwareItem("steam.exe", "1.0.0"),
		softwareItem("Code.exe", "1.0.0"),
	}}, cacheDir)
	baseline.now = fixedNow
	if events, err := baseline.Collect(context.Background(), policy); err != nil || len(events) != 0 {
		t.Fatalf("baseline events=%d err=%v", len(events), err)
	}

	collector := New(fakePlatform{inventory: []platform.InstalledSoftware{
		softwareItem("Code.exe", "1.0.0"),
		softwareItem("qbittorrent.exe", "5.0.0"),
	}}, cacheDir)
	collector.now = fixedNow
	events, err := collector.Collect(context.Background(), policy)
	if err != nil {
		t.Fatalf("collect software changes: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected install and uninstall events, got %d: %+v", len(events), events)
	}

	seenInstalled := false
	seenUninstalled := false
	for _, evt := range events {
		if evt.Source != constants.EventSourceSoftwareCollector {
			t.Fatalf("unexpected source: %+v", evt)
		}
		if evt.PathHash == "" || strings.Contains(evt.PathHash, `C:\`) {
			t.Fatalf("expected hashed software id only, got %q", evt.PathHash)
		}
		if evt.Metadata[constants.EventMetadataSoftwareInventoryMode] != constants.SoftwareInventoryModeMetadataOnly ||
			evt.Metadata[constants.EventMetadataPathMode] != constants.PathModeNone {
			t.Fatalf("expected metadata-only software event, got %+v", evt.Metadata)
		}
		switch evt.Type {
		case constants.EventTypeSoftwareInstalled:
			seenInstalled = true
			if evt.AppName != "qbittorrent.exe" {
				t.Fatalf("expected qbittorrent install event, got %+v", evt)
			}
			if evt.Metadata[constants.EventMetadataSoftwareRiskCategory] != constants.SoftwareRiskCategoryTorrentClient {
				t.Fatalf("expected torrent risk metadata, got %+v", evt.Metadata)
			}
		case constants.EventTypeSoftwareUninstalled:
			seenUninstalled = true
			if evt.AppName != "steam.exe" {
				t.Fatalf("expected steam uninstall event, got %+v", evt)
			}
		default:
			t.Fatalf("unexpected event type: %s", evt.Type)
		}
	}
	if !seenInstalled || !seenUninstalled {
		t.Fatalf("expected both install and uninstall events: %+v", events)
	}
}

func TestCollectorSkipsDisabledOrUnsupportedInventory(t *testing.T) {
	t.Parallel()

	disabled, err := New(fakePlatform{inventory: []platform.InstalledSoftware{
		softwareItem("Code.exe", "1.0.0"),
	}}, t.TempDir()).Collect(context.Background(), softwarePolicy(false))
	if err != nil {
		t.Fatalf("disabled software collection should not fail: %v", err)
	}
	if len(disabled) != 0 {
		t.Fatalf("expected disabled software collection to emit no events, got %d", len(disabled))
	}

	unsupported, err := New(fakePlatform{
		err: platform.CapabilityError{
			OperatingSystem: constants.OperatingSystemOther,
			CapabilityID:    constants.PlatformCapabilitySoftwareInventory,
			Status:          constants.PlatformSupportUnsupported,
			Reason:          "not available",
		},
	}, t.TempDir()).Collect(context.Background(), softwarePolicy(true))
	if err != nil {
		t.Fatalf("unsupported software collection should not fail: %v", err)
	}
	if len(unsupported) != 0 {
		t.Fatalf("expected unsupported software collection to emit no events, got %d", len(unsupported))
	}
}

func softwareItem(name string, version string) platform.InstalledSoftware {
	return platform.InstalledSoftware{
		ID:        `C:\Program Files\` + name,
		Name:      name,
		Version:   version,
		Publisher: "TraceDeck Test Publisher",
		Source:    constants.SoftwareSourceWindowsRegistry,
	}
}

func softwarePolicy(enabled bool) *config.Policy {
	return &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Profile:  constants.DefaultProfile,
		Collection: config.CollectionPolicy{
			Software: config.SoftwareCollection{
				Enabled:       enabled,
				InventoryMode: config.SoftwareInventoryMode(constants.SoftwareInventoryModeMetadataOnly),
			},
		},
	}
}

func fixedNow() time.Time {
	return time.Date(2026, 6, 16, 10, 30, 0, 0, time.UTC)
}
