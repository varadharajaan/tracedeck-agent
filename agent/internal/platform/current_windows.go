//go:build windows

package platform

import (
	"context"
	"fmt"
	"strings"
	"unsafe"

	gopsprocess "github.com/shirou/gopsutil/v4/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type windowsAdapter struct{}

var (
	user32DLL                    = windows.NewLazySystemDLL("user32.dll")
	procGetForegroundWindow      = user32DLL.NewProc("GetForegroundWindow")
	procGetWindowThreadProcessID = user32DLL.NewProc("GetWindowThreadProcessId")
)

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

func (windowsAdapter) ForegroundApp(ctx context.Context) (ForegroundApp, error) {
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ForegroundApp{}, fmt.Errorf("%w: no foreground window handle", ErrNoForegroundApp)
	}

	var pid uint32
	procGetWindowThreadProcessID.Call(hwnd, uintptr(unsafe.Pointer(&pid)))
	if pid == 0 {
		return ForegroundApp{}, fmt.Errorf("%w: no foreground process id", ErrNoForegroundApp)
	}

	proc, err := gopsprocess.NewProcessWithContext(ctx, int32(pid))
	if err != nil {
		return ForegroundApp{}, fmt.Errorf("%w: open foreground process: %v", ErrNoForegroundApp, err)
	}
	name, err := proc.NameWithContext(ctx)
	if err != nil || name == "" {
		return ForegroundApp{}, fmt.Errorf("%w: read foreground process name: %v", ErrNoForegroundApp, err)
	}
	exe, _ := proc.ExeWithContext(ctx)

	return ForegroundApp{
		AppName:        name,
		ProcessID:      int32(pid),
		ExecutablePath: exe,
	}, nil
}

func (windowsAdapter) SoftwareInventory(ctx context.Context) ([]InstalledSoftware, error) {
	return windowsSoftwareInventory(ctx)
}

func windowsSoftwareInventory(ctx context.Context) ([]InstalledSoftware, error) {
	roots := []struct {
		key  registry.Key
		path string
	}{
		{key: registry.LOCAL_MACHINE, path: `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
		{key: registry.LOCAL_MACHINE, path: `SOFTWARE\WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall`},
		{key: registry.CURRENT_USER, path: `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`},
	}

	var out []InstalledSoftware
	seen := map[string]struct{}{}
	for _, root := range roots {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		items, err := readWindowsUninstallRoot(ctx, root.key, root.path)
		if err != nil {
			continue
		}
		for _, item := range items {
			key := strings.ToLower(strings.TrimSpace(item.ID))
			if key == "" {
				key = strings.ToLower(strings.TrimSpace(item.Name + "\x00" + item.Version + "\x00" + item.Publisher))
			}
			if key == "" {
				continue
			}
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, item)
		}
	}
	return out, nil
}

func readWindowsUninstallRoot(ctx context.Context, root registry.Key, rootPath string) ([]InstalledSoftware, error) {
	key, err := registry.OpenKey(root, rootPath, registry.READ)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = key.Close()
	}()

	names, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, err
	}

	out := make([]InstalledSoftware, 0, len(names))
	for _, subkeyName := range names {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		subkey, err := registry.OpenKey(key, subkeyName, registry.READ)
		if err != nil {
			continue
		}
		name, _, _ := subkey.GetStringValue("DisplayName")
		version, _, _ := subkey.GetStringValue("DisplayVersion")
		publisher, _, _ := subkey.GetStringValue("Publisher")
		_ = subkey.Close()

		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		out = append(out, InstalledSoftware{
			ID:        rootPath + `\` + subkeyName,
			Name:      name,
			Version:   strings.TrimSpace(version),
			Publisher: strings.TrimSpace(publisher),
			Source:    constants.SoftwareSourceWindowsRegistry,
		})
	}
	return out, nil
}
