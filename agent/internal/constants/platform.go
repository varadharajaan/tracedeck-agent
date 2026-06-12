package constants

const (
	OperatingSystemWindows = "windows"
	OperatingSystemMacOS   = "darwin"
	OperatingSystemLinux   = "linux"
	OperatingSystemOther   = "other"
)

const (
	PlatformCapabilityProcessCollection = "process_collection"
	PlatformCapabilityLocalStorage      = "local_storage"
	PlatformCapabilityServiceManager    = "service_manager"
	PlatformCapabilityForegroundApp     = "foreground_app"
	PlatformCapabilitySoftwareInventory = "software_inventory"
	PlatformCapabilityMediaMetadata     = "media_metadata"
	PlatformCapabilityBrowserHistory    = "browser_history"
	PlatformCapabilityLocalIndicator    = "local_indicator"
)

const (
	PlatformSupportSupported          = "supported"
	PlatformSupportRequiresPermission = "requires_permission"
	PlatformSupportPartial            = "partial"
	PlatformSupportPlanned            = "planned"
	PlatformSupportUnsupported        = "unsupported"
)

const (
	ServiceManagerWindowsService = "windows_service"
	ServiceManagerLaunchd        = "launchd"
	ServiceManagerSystemd        = "systemd"
	ServiceManagerNone           = "none"
)
