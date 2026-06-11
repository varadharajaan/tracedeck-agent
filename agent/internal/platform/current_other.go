//go:build !windows && !darwin && !linux

package platform

import (
	"context"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type otherAdapter struct{}

func Current() Adapter {
	return otherAdapter{}
}

func (otherAdapter) Name() string {
	return constants.OperatingSystemOther
}

func (otherAdapter) Hostname(ctx context.Context) (string, error) {
	return osHostname(ctx)
}

func (otherAdapter) Capabilities() Capabilities {
	return Capabilities{
		OperatingSystem:   constants.OperatingSystemOther,
		ProcessCollection: false,
		LocalStorage:      true,
	}
}
