//go:build linux

package platform

import (
	"bufio"
	"context"
	"os"
	"strings"

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

func (linuxAdapter) SoftwareInventory(ctx context.Context) ([]InstalledSoftware, error) {
	items, err := linuxDPKGInventory(ctx, "/var/lib/dpkg/status")
	if err != nil {
		return unsupportedSoftwareInventory(LinuxCapabilities())
	}
	return items, nil
}

func linuxDPKGInventory(ctx context.Context, statusPath string) ([]InstalledSoftware, error) {
	file, err := os.Open(statusPath)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

	var out []InstalledSoftware
	current := InstalledSoftware{Source: constants.SoftwareSourceLinuxDPKG}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			if current.Name != "" {
				out = append(out, current)
			}
			current = InstalledSoftware{Source: constants.SoftwareSourceLinuxDPKG}
			continue
		}
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = strings.TrimSpace(value)
		switch strings.TrimSpace(key) {
		case "Package":
			current.ID = value
			current.Name = value
		case "Version":
			current.Version = value
		case "Maintainer":
			current.Publisher = value
		}
	}
	if current.Name != "" {
		out = append(out, current)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
