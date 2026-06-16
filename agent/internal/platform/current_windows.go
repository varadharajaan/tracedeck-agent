//go:build windows

package platform

import (
	"context"
	"fmt"
	"unsafe"

	gopsprocess "github.com/shirou/gopsutil/v4/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"golang.org/x/sys/windows"
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
