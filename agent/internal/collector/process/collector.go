package process

import (
	"context"
	"fmt"
	"time"

	gopsprocess "github.com/shirou/gopsutil/v4/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/redaction"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/software"
)

type Collector struct {
	limit           int
	platformAdapter platform.Adapter
}

func New(limit int, platformAdapter platform.Adapter) *Collector {
	if limit <= 0 {
		limit = constants.DefaultProcessLimit
	}
	if platformAdapter == nil {
		platformAdapter = platform.Current()
	}
	return &Collector{limit: limit, platformAdapter: platformAdapter}
}

func (c *Collector) Collect(ctx context.Context, policy *config.Policy) ([]event.Event, error) {
	processes, err := gopsprocess.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	hostName, err := c.platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	capabilities := c.platformAdapter.Capabilities()
	now := time.Now().UTC()
	events := make([]event.Event, 0, min(c.limit, len(processes)))

	for _, proc := range processes {
		if len(events) >= c.limit {
			break
		}

		name, err := proc.NameWithContext(ctx)
		if err != nil || name == "" {
			continue
		}

		exe, _ := proc.ExeWithContext(ctx)
		metadata := map[string]string{
			constants.EventMetadataProfile:         policy.Profile,
			constants.EventMetadataOperatingSystem: capabilities.OperatingSystem,
		}
		if risk, ok := software.ClassifyProcess(name, exe); ok {
			metadata[constants.EventMetadataSoftwareRiskCategory] = risk.Category
			metadata[constants.EventMetadataSoftwareRiskReason] = risk.Reason
		}
		events = append(events, event.Event{
			Type:      constants.EventTypeProcessObserved,
			Source:    constants.EventSourceProcessCollector,
			Timestamp: now,
			TenantID:  policy.TenantID,
			DeviceID:  policy.DeviceID,
			HostName:  hostName,
			AppName:   name,
			ProcessID: proc.Pid,
			PathHash:  redaction.HashPath(exe),
			Metadata:  metadata,
		})
	}

	return events, nil
}
