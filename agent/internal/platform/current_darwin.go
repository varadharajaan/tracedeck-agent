//go:build darwin

package platform

import (
	"context"
	"os"
	"path/filepath"
	"strings"

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
	return DarwinCapabilities()
}

func (darwinAdapter) ForegroundApp(context.Context) (ForegroundApp, error) {
	return unsupportedForegroundApp(DarwinCapabilities())
}

func (darwinAdapter) SoftwareInventory(ctx context.Context) ([]InstalledSoftware, error) {
	return darwinSoftwareInventory(ctx)
}

func darwinSoftwareInventory(ctx context.Context) ([]InstalledSoftware, error) {
	home, _ := os.UserHomeDir()
	roots := []string{"/Applications"}
	if home != "" {
		roots = append(roots, filepath.Join(home, "Applications"))
	}

	var out []InstalledSoftware
	seen := map[string]struct{}{}
	for _, root := range roots {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if !entry.IsDir() || !strings.HasSuffix(strings.ToLower(entry.Name()), ".app") {
				continue
			}
			name := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			key := strings.ToLower(name)
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, InstalledSoftware{
				ID:     filepath.Join(root, entry.Name()),
				Name:   name,
				Source: constants.SoftwareSourceMacOSApplications,
			})
		}
	}
	return out, nil
}
