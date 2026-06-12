package health

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"time"

	gopscpu "github.com/shirou/gopsutil/v4/cpu"
	gopsdisk "github.com/shirou/gopsutil/v4/disk"
	gopshost "github.com/shirou/gopsutil/v4/host"
	gopsmem "github.com/shirou/gopsutil/v4/mem"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
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
	hostName, err := c.platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	capabilities := c.platformAdapter.Capabilities()

	cpuPercent := 0.0
	cpuValues, err := gopscpu.PercentWithContext(ctx, 0, false)
	if err == nil && len(cpuValues) > 0 {
		cpuPercent = cpuValues[0]
	}

	memoryPercent := 0.0
	if memory, err := gopsmem.VirtualMemoryWithContext(ctx); err == nil && memory != nil {
		memoryPercent = memory.UsedPercent
	}

	diskPercent := 0.0
	if usage, err := gopsdisk.UsageWithContext(ctx, healthDiskPath()); err == nil && usage != nil {
		diskPercent = usage.UsedPercent
	}

	bootTimeUnix := uint64(0)
	uptimeSeconds := uint64(0)
	if info, err := gopshost.InfoWithContext(ctx); err == nil && info != nil {
		bootTimeUnix = info.BootTime
		uptimeSeconds = info.Uptime
	}

	score := healthScore(cpuPercent, memoryPercent, diskPercent)
	status := healthStatus(score)
	now := time.Now().UTC()
	return []event.Event{
		{
			Type:      constants.EventTypeDeviceHealth,
			Source:    constants.EventSourceHealthCollector,
			Timestamp: now,
			TenantID:  policy.TenantID,
			DeviceID:  policy.DeviceID,
			HostName:  hostName,
			Metadata: map[string]string{
				constants.EventMetadataProfile:         policy.Profile,
				constants.EventMetadataOperatingSystem: capabilities.OperatingSystem,
				constants.EventMetadataHealthScore:     strconv.Itoa(score),
				constants.EventMetadataCPUPercent:      formatPercent(cpuPercent),
				constants.EventMetadataMemoryPercent:   formatPercent(memoryPercent),
				constants.EventMetadataDiskPercent:     formatPercent(diskPercent),
				constants.EventMetadataBootTimeUnix:    strconv.FormatUint(bootTimeUnix, 10),
				constants.EventMetadataUptimeSeconds:   strconv.FormatUint(uptimeSeconds, 10),
				constants.EventMetadataHealthStatus:    status,
			},
		},
	}, nil
}

func healthDiskPath() string {
	if runtime.GOOS == "windows" {
		if systemDrive := os.Getenv("SystemDrive"); systemDrive != "" {
			return systemDrive + `\`
		}
		return `C:\`
	}
	return "/"
}

func healthScore(cpuPercent float64, memoryPercent float64, diskPercent float64) int {
	score := 100
	score -= int(cpuPercent / 5)
	score -= int(memoryPercent / 4)
	score -= int(diskPercent / 5)
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func healthStatus(score int) string {
	switch {
	case score >= 85:
		return constants.HealthStatusHealthy
	case score >= 65:
		return constants.HealthStatusWatch
	default:
		return constants.HealthStatusAttention
	}
}

func formatPercent(value float64) string {
	return fmt.Sprintf("%.1f", value)
}
