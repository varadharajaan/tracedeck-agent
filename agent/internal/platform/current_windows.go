//go:build windows

package platform

import (
	"context"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type windowsAdapter struct{}

func Current() Adapter {
	return windowsAdapter{}
}

func (windowsAdapter) Name() string {
	return constants.OperatingSystemWindows
}

func (windowsAdapter) Hostname(ctx context.Context) (string, error) {
	return osHostname(ctx)
}

func (windowsAdapter) Capabilities() Capabilities {
	return WindowsCapabilities()
}
