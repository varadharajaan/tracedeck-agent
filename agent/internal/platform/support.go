package platform

import "github.com/varadharajaan/tracedeck-agent/agent/internal/constants"

func WindowsCapabilities() Capabilities {
	return Capabilities{
		OperatingSystem:   constants.OperatingSystemWindows,
		ServiceManager:    constants.ServiceManagerWindowsService,
		ProcessCollection: true,
		LocalStorage:      true,
		Features: []CapabilitySupport{
			supported(constants.PlatformCapabilityProcessCollection, "process snapshots are available through the current Windows collector path"),
			supported(constants.PlatformCapabilityLocalStorage, "SQLite local buffer is supported"),
			supported(constants.PlatformCapabilityServiceManager, "Windows Service packaging is planned around the native service manager"),
			supported(constants.PlatformCapabilityForegroundApp, "foreground app support uses Windows desktop APIs"),
			planned(constants.PlatformCapabilitySoftwareInventory, "Windows software inventory collector remains a later hardening slice"),
			planned(constants.PlatformCapabilityMediaMetadata, "media metadata remains policy-gated and collector-specific"),
			planned(constants.PlatformCapabilityBrowserHistory, "browser history collector is Chromium-domain-only and policy bounded"),
			planned(constants.PlatformCapabilityLocalIndicator, "visible local indicator is required before interactive monitoring expansion"),
		},
	}
}

func DarwinCapabilities() Capabilities {
	return Capabilities{
		OperatingSystem:   constants.OperatingSystemMacOS,
		ServiceManager:    constants.ServiceManagerLaunchd,
		ProcessCollection: true,
		LocalStorage:      true,
		Features: []CapabilitySupport{
			supported(constants.PlatformCapabilityProcessCollection, "process snapshots are available through portable process collection"),
			supported(constants.PlatformCapabilityLocalStorage, "SQLite local buffer is supported"),
			supported(constants.PlatformCapabilityServiceManager, "launchd manifest packaging is generated in Phase 7"),
			requiresPermission(constants.PlatformCapabilityForegroundApp, "foreground app collection requires macOS Accessibility permission"),
			planned(constants.PlatformCapabilitySoftwareInventory, "macOS software inventory needs native adapter hardening"),
			planned(constants.PlatformCapabilityMediaMetadata, "media metadata remains policy-gated and requires native process inspection"),
			planned(constants.PlatformCapabilityBrowserHistory, "browser history support remains domain-only and profile-bounded"),
			planned(constants.PlatformCapabilityLocalIndicator, "menu bar or notification indicator remains a later UI slice"),
		},
	}
}

func LinuxCapabilities() Capabilities {
	return Capabilities{
		OperatingSystem:   constants.OperatingSystemLinux,
		ServiceManager:    constants.ServiceManagerSystemd,
		ProcessCollection: true,
		LocalStorage:      true,
		Features: []CapabilitySupport{
			supported(constants.PlatformCapabilityProcessCollection, "process snapshots are available through portable process collection"),
			supported(constants.PlatformCapabilityLocalStorage, "SQLite local buffer is supported"),
			supported(constants.PlatformCapabilityServiceManager, "systemd unit packaging is generated in Phase 7"),
			partial(constants.PlatformCapabilityForegroundApp, "foreground app support differs between X11 and Wayland compositors"),
			planned(constants.PlatformCapabilitySoftwareInventory, "Linux package inventory needs distro-aware adapters"),
			planned(constants.PlatformCapabilityMediaMetadata, "media metadata remains policy-gated and desktop-environment dependent"),
			planned(constants.PlatformCapabilityBrowserHistory, "browser history support remains domain-only and profile-bounded"),
			planned(constants.PlatformCapabilityLocalIndicator, "tray or desktop notification indicator remains a later UI slice"),
		},
	}
}

func OtherCapabilities() Capabilities {
	return Capabilities{
		OperatingSystem:   constants.OperatingSystemOther,
		ServiceManager:    constants.ServiceManagerNone,
		ProcessCollection: false,
		LocalStorage:      true,
		Features: []CapabilitySupport{
			unsupported(constants.PlatformCapabilityProcessCollection, "unsupported operating systems do not declare process collection"),
			supported(constants.PlatformCapabilityLocalStorage, "SQLite local buffer can run where Go and SQLite are available"),
			unsupported(constants.PlatformCapabilityServiceManager, "no service manager is declared for unsupported operating systems"),
			unsupported(constants.PlatformCapabilityForegroundApp, "foreground app collection requires a platform adapter"),
			unsupported(constants.PlatformCapabilitySoftwareInventory, "software inventory requires a platform adapter"),
			unsupported(constants.PlatformCapabilityMediaMetadata, "media metadata requires a platform adapter and explicit policy"),
			unsupported(constants.PlatformCapabilityBrowserHistory, "browser history collection requires a supported platform profile adapter"),
			unsupported(constants.PlatformCapabilityLocalIndicator, "local indicator requires platform UI integration"),
		},
	}
}

func supported(id string, notes string) CapabilitySupport {
	return capability(id, constants.PlatformSupportSupported, false, notes)
}

func requiresPermission(id string, notes string) CapabilitySupport {
	return capability(id, constants.PlatformSupportRequiresPermission, true, notes)
}

func partial(id string, notes string) CapabilitySupport {
	return capability(id, constants.PlatformSupportPartial, false, notes)
}

func planned(id string, notes string) CapabilitySupport {
	return capability(id, constants.PlatformSupportPlanned, false, notes)
}

func unsupported(id string, notes string) CapabilitySupport {
	return capability(id, constants.PlatformSupportUnsupported, false, notes)
}

func capability(id string, status string, permissionRequired bool, notes string) CapabilitySupport {
	return CapabilitySupport{
		ID:                 id,
		Status:             status,
		PermissionRequired: permissionRequired,
		Notes:              notes,
	}
}
