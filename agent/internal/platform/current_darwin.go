//go:build darwin

package platform

import (
	"context"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type darwinAdapter struct{}

func Current() Adapter {
	return darwinAdapter{}
}

func (darwinAdapter) Name() string {
	return constants.OperatingSystemMacOS
}

func (darwinAdapter) Hostname(ctx context.Context) (string, error) {
	return osHostname(ctx)
}

func (darwinAdapter) Capabilities() Capabilities {
	return Capabilities{
		OperatingSystem:   constants.OperatingSystemMacOS,
		ProcessCollection: true,
		LocalStorage:      true,
	}
}
