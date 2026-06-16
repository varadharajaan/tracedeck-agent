package software

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/redaction"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
	softwarerisk "github.com/varadharajaan/tracedeck-agent/agent/internal/software"
)

type Collector struct {
	platformAdapter platform.Adapter
	snapshotDir     string
	now             func() time.Time
}

type snapshot struct {
	SchemaVersion string                   `json:"schema_version"`
	CapturedAt    string                   `json:"captured_at"`
	Entries       map[string]snapshotEntry `json:"entries"`
}

type snapshotEntry struct {
	IDHash    string `json:"id_hash"`
	Name      string `json:"name"`
	NameHash  string `json:"name_hash"`
	Version   string `json:"version,omitempty"`
	Publisher string `json:"publisher,omitempty"`
	Source    string `json:"source"`
}

func New(platformAdapter platform.Adapter, snapshotDir string) *Collector {
	if platformAdapter == nil {
		platformAdapter = platform.Current()
	}
	if snapshotDir == "" {
		snapshotDir = filepath.Join(constants.DefaultDataDir, constants.SoftwareCacheDirName)
	}
	return &Collector{
		platformAdapter: platformAdapter,
		snapshotDir:     snapshotDir,
		now:             func() time.Time { return time.Now().UTC() },
	}
}

func (c *Collector) Collect(ctx context.Context, policy *config.Policy) ([]event.Event, error) {
	if policy == nil || !policy.Collection.Software.Enabled {
		return nil, nil
	}

	inventory, err := c.platformAdapter.SoftwareInventory(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrUnsupportedCapability) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect software inventory: %w", err)
	}
	if len(inventory) == 0 {
		return nil, nil
	}

	current := c.buildSnapshot(inventory)
	if len(current.Entries) == 0 {
		return nil, nil
	}

	previous, found, err := c.loadSnapshot()
	if err != nil {
		return nil, err
	}
	if err := c.saveSnapshot(current); err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}

	hostName, err := c.platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	capabilities := c.platformAdapter.Capabilities()
	observedAt := c.now().UTC()
	events := make([]event.Event, 0)

	for _, id := range sortedNewKeys(current.Entries, previous.Entries) {
		events = append(events, c.changeEvent(policy, hostName, capabilities.OperatingSystem, current.Entries[id], constants.SoftwareChangeInstalled, observedAt))
	}
	for _, id := range sortedNewKeys(previous.Entries, current.Entries) {
		events = append(events, c.changeEvent(policy, hostName, capabilities.OperatingSystem, previous.Entries[id], constants.SoftwareChangeUninstalled, observedAt))
	}

	return events, nil
}

func (c *Collector) buildSnapshot(inventory []platform.InstalledSoftware) snapshot {
	entries := make(map[string]snapshotEntry, len(inventory))
	for _, item := range inventory {
		entry := newSnapshotEntry(item)
		if entry.IDHash == "" || entry.Name == "" {
			continue
		}
		entries[entry.IDHash] = entry
	}
	return snapshot{
		SchemaVersion: constants.PolicySchemaVersionV1Alpha1,
		CapturedAt:    c.now().UTC().Format(time.RFC3339Nano),
		Entries:       entries,
	}
}

func newSnapshotEntry(item platform.InstalledSoftware) snapshotEntry {
	name := strings.TrimSpace(item.Name)
	version := strings.TrimSpace(item.Version)
	publisher := strings.TrimSpace(item.Publisher)
	source := strings.TrimSpace(item.Source)
	if source == "" {
		source = constants.PlatformCapabilitySoftwareInventory
	}
	fingerprint := strings.Join([]string{
		strings.TrimSpace(item.ID),
		name,
		version,
		publisher,
		source,
	}, "\x00")
	return snapshotEntry{
		IDHash:    redaction.HashValue(fingerprint),
		Name:      name,
		NameHash:  redaction.HashValue(name),
		Version:   version,
		Publisher: publisher,
		Source:    source,
	}
}

func (c *Collector) changeEvent(policy *config.Policy, hostName string, operatingSystem string, item snapshotEntry, change string, observedAt time.Time) event.Event {
	eventType := constants.EventTypeSoftwareInstalled
	if change == constants.SoftwareChangeUninstalled {
		eventType = constants.EventTypeSoftwareUninstalled
	}
	metadata := map[string]string{
		constants.EventMetadataProfile:               policy.Profile,
		constants.EventMetadataOperatingSystem:       operatingSystem,
		constants.EventMetadataSoftwareChange:        change,
		constants.EventMetadataSoftwareInventoryMode: constants.SoftwareInventoryModeMetadataOnly,
		constants.EventMetadataSoftwareNameHash:      item.NameHash,
		constants.EventMetadataSoftwareSource:        item.Source,
		constants.EventMetadataSoftwareSnapshotID:    item.IDHash,
		constants.EventMetadataPathMode:              constants.PathModeNone,
	}
	if item.Version != "" {
		metadata[constants.EventMetadataSoftwareVersion] = item.Version
	}
	if item.Publisher != "" {
		metadata[constants.EventMetadataSoftwarePublisher] = item.Publisher
	}
	if risk, ok := softwarerisk.ClassifyProcess(item.Name, ""); ok {
		metadata[constants.EventMetadataSoftwareRiskCategory] = risk.Category
		metadata[constants.EventMetadataSoftwareRiskReason] = risk.Reason
	}

	return event.Event{
		Type:      eventType,
		Source:    constants.EventSourceSoftwareCollector,
		Timestamp: observedAt,
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  hostName,
		AppName:   item.Name,
		PathHash:  item.IDHash,
		Metadata:  metadata,
	}
}

func (c *Collector) loadSnapshot() (snapshot, bool, error) {
	data, err := os.ReadFile(c.snapshotPath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return snapshot{}, false, nil
		}
		return snapshot{}, false, fmt.Errorf("read software inventory snapshot: %w", err)
	}
	var out snapshot
	if err := json.Unmarshal(data, &out); err != nil {
		return snapshot{}, false, fmt.Errorf("decode software inventory snapshot: %w", err)
	}
	if out.Entries == nil {
		out.Entries = map[string]snapshotEntry{}
	}
	return out, true, nil
}

func (c *Collector) saveSnapshot(next snapshot) error {
	if err := os.MkdirAll(c.snapshotDir, 0o750); err != nil {
		return fmt.Errorf("create software inventory snapshot dir: %w", err)
	}
	data, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal software inventory snapshot: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(c.snapshotPath(), data, 0o600); err != nil {
		return fmt.Errorf("write software inventory snapshot: %w", err)
	}
	return nil
}

func (c *Collector) snapshotPath() string {
	return filepath.Join(c.snapshotDir, constants.SoftwareSnapshotFileName)
}

func sortedNewKeys(left map[string]snapshotEntry, right map[string]snapshotEntry) []string {
	keys := make([]string, 0)
	for key := range left {
		if _, ok := right[key]; !ok {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	return keys
}
