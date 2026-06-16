package constants

const (
	EventMetadataSoftwareRiskCategory  = "software_risk_category"
	EventMetadataSoftwareRiskReason    = "software_risk_reason"
	EventMetadataSoftwareChange        = "software_change"
	EventMetadataSoftwareInventoryMode = "software_inventory_mode"
	EventMetadataSoftwareNameHash      = "software_name_hash"
	EventMetadataSoftwareVersion       = "software_version"
	EventMetadataSoftwarePublisher     = "software_publisher"
	EventMetadataSoftwareSource        = "software_source"
	EventMetadataSoftwareSnapshotID    = "software_snapshot_id"
)

const (
	SoftwareInventoryModeMetadataOnly = "metadata_only"
	SoftwareChangeInstalled           = "installed"
	SoftwareChangeUninstalled         = "uninstalled"
	SoftwareSourceWindowsRegistry     = "windows_registry_uninstall"
	SoftwareSourceMacOSApplications   = "macos_applications"
	SoftwareSourceLinuxDPKG           = "linux_dpkg_status"
	SoftwareSnapshotFileName          = "software-inventory-snapshot.json"
)

const (
	SoftwareRiskCategoryTorrentClient      = "torrent_client"
	SoftwareRiskCategoryVPNProxy           = "vpn_proxy"
	SoftwareRiskCategoryGameLauncher       = "game_launcher"
	SoftwareRiskCategoryUnknownBrowser     = "unknown_browser"
	SoftwareRiskCategoryDownloadsInstaller = "downloads_installer"
)

const (
	SoftwareRiskReasonTorrentClient      = "torrent client process observed"
	SoftwareRiskReasonVPNProxy           = "vpn or proxy tool process observed"
	SoftwareRiskReasonGameLauncher       = "game launcher process observed"
	SoftwareRiskReasonUnknownBrowser     = "non-standard browser process observed"
	SoftwareRiskReasonDownloadsInstaller = "installer launched from downloads location"
)

const (
	AlertRuleRiskySoftwareDetected    = "risky_software_detected"
	AlertRuleUnknownSoftwareInstalled = "unknown_software_installed"
	AlertReasonRiskySoftwareProcess   = "risky software process observed"
	AlertReasonSoftwareInstalled      = "new software install observed"
	AlertMetadataSoftwareRiskCategory = "software_risk_category"
	AlertMetadataSoftwareRiskReason   = "software_risk_reason"
)

var RiskyTorrentProcessNames = []string{
	"qbittorrent.exe",
	"utorrent.exe",
	"bittorrent.exe",
	"transmission-qt.exe",
}

var RiskyVPNProxyProcessNames = []string{
	"nordvpn.exe",
	"protonvpn.exe",
	"openvpn.exe",
	"wireguard.exe",
	"outline-client.exe",
	"psiphon3.exe",
	"tor.exe",
	"torbrowser.exe",
}

var RiskyGameLauncherProcessNames = []string{
	"steam.exe",
	"epicgameslauncher.exe",
	"riotclientservices.exe",
	"battle.net.exe",
	"goggalaxy.exe",
}

var KnownBrowserProcessNames = []string{
	"chrome.exe",
	"msedge.exe",
	"firefox.exe",
	"brave.exe",
}

var UnknownBrowserProcessNames = []string{
	"opera.exe",
	"vivaldi.exe",
	"duckduckgo.exe",
	"torbrowser.exe",
}
