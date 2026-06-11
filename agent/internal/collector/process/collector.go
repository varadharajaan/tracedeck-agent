package process

import (
	"context"
	"fmt"
	"os"
	"time"

	gopsprocess "github.com/shirou/gopsutil/v4/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/redaction"
)

type Collector struct {
	limit int
}

func New(limit int) *Collector {
	if limit <= 0 {
		limit = constants.DefaultProcessLimit
	}
	return &Collector{limit: limit}
}

func (c *Collector) Collect(ctx context.Context, policy *config.Policy) ([]event.Event, error) {
	processes, err := gopsprocess.ProcessesWithContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("list processes: %w", err)
	}

	hostName, _ := os.Hostname()
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
			Metadata: map[string]string{
				constants.EventMetadataProfile: policy.Profile,
			},
		})
	}

	return events, nil
}
