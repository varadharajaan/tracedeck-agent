package heartbeat

import (
	"context"
	"strconv"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
)

type Options struct {
	ContinuousMode     bool
	CollectionInterval string
	ArchiveEnabled     bool
	ArchiveDue         bool
	BackendSyncEnabled bool
	AlertsEnabled      bool
}

type Collector struct {
	platformAdapter platform.Adapter
	options         Options
}

func New(platformAdapter platform.Adapter, options Options) *Collector {
	if platformAdapter == nil {
		platformAdapter = platform.Current()
	}
	return &Collector{platformAdapter: platformAdapter, options: options}
}

func (c *Collector) Collect(ctx context.Context, policy *config.Policy) ([]event.Event, error) {
	hostName, err := c.platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	capabilities := c.platformAdapter.Capabilities()
	collectionMode := constants.HeartbeatCollectionModeOnce
	if c.options.ContinuousMode {
		collectionMode = constants.HeartbeatCollectionModeContinuous
	}
	collectionInterval := c.options.CollectionInterval
	if collectionInterval == "" {
		collectionInterval = constants.DefaultCollectionInterval
	}

	return []event.Event{
		{
			Type:      constants.EventTypeAgentHeartbeat,
			Source:    constants.EventSourceHeartbeat,
			Timestamp: time.Now().UTC(),
			TenantID:  policy.TenantID,
			DeviceID:  policy.DeviceID,
			HostName:  hostName,
			AppName:   constants.AppName,
			Metadata: map[string]string{
				constants.EventMetadataProfile:         policy.Profile,
				constants.EventMetadataOperatingSystem: capabilities.OperatingSystem,
				constants.EventMetadataAgentHealthy:    strconv.FormatBool(true),
				constants.EventMetadataAgentVersion:    constants.AppVersion,
				constants.EventMetadataCollectionMode:  collectionMode,
				constants.EventMetadataCollectionEvery: collectionInterval,
				constants.EventMetadataArchiveEnabled:  strconv.FormatBool(c.options.ArchiveEnabled),
				constants.EventMetadataArchiveDue:      strconv.FormatBool(c.options.ArchiveDue),
				constants.EventMetadataBackendSync:     strconv.FormatBool(c.options.BackendSyncEnabled),
				constants.EventMetadataAlertsEnabled:   strconv.FormatBool(c.options.AlertsEnabled),
			},
		},
	}, nil
}
