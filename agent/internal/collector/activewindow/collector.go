package activewindow

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/redaction"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/software"
)

type Collector struct {
	platformAdapter platform.Adapter
}

func New(platformAdapter platform.Adapter) *Collector {
	if platformAdapter == nil {
		platformAdapter = platform.Current()
	}
	return &Collector{platformAdapter: platformAdapter}
}

func (c *Collector) Collect(ctx context.Context, policy *config.Policy) ([]event.Event, error) {
	if policy == nil || !policy.Collection.ForegroundApp.Enabled {
		return nil, nil
	}

	foreground, err := c.platformAdapter.ForegroundApp(ctx)
	if err != nil {
		if errors.Is(err, platform.ErrUnsupportedCapability) || errors.Is(err, platform.ErrNoForegroundApp) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect foreground app: %w", err)
	}
	if foreground.AppName == "" {
		return nil, nil
	}

	hostName, err := c.platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	capabilities := c.platformAdapter.Capabilities()
	titleMode := string(policy.Collection.ForegroundApp.WindowTitleMode)
	if titleMode == "" {
		titleMode = constants.WindowTitleModeNone
	}
	metadata := map[string]string{
		constants.EventMetadataProfile:         policy.Profile,
		constants.EventMetadataOperatingSystem: capabilities.OperatingSystem,
		constants.EventMetadataForegroundState: constants.ForegroundStateActive,
		constants.EventMetadataWindowTitleMode: titleMode,
		constants.EventMetadataPathMode:        constants.PathModeHashOnly,
	}
	if risk, ok := software.ClassifyProcess(foreground.AppName, foreground.ExecutablePath); ok {
		metadata[constants.EventMetadataSoftwareRiskCategory] = risk.Category
		metadata[constants.EventMetadataSoftwareRiskReason] = risk.Reason
	}

	return []event.Event{{
		Type:      constants.EventTypeForegroundAppObserved,
		Source:    constants.EventSourceForegroundAppCollector,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  hostName,
		AppName:   foreground.AppName,
		ProcessID: foreground.ProcessID,
		PathHash:  redaction.HashPath(foreground.ExecutablePath),
		Metadata:  metadata,
	}}, nil
}
