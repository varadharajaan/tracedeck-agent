//go:build linux

package platform

import (
	"context"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type linuxAdapter struct{}

func Current() Adapter {
	return linuxAdapter{}
}

func (linuxAdapter) Name() string {
	return constants.OperatingSystemLinux
}

func (linuxAdapter) Hostname(ctx context.Context) (string, error) {
	return osHostname(ctx)
}

func (linuxAdapter) Capabilities() Capabilities {
	return LinuxCapabilities()
}

func (linuxAdapter) ForegroundApp(context.Context) (ForegroundApp, error) {
	return unsupportedForegroundApp(LinuxCapabilities())
}
