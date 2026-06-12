package constants

const (
	EventMetadataSoftwareRiskCategory = "software_risk_category"
	EventMetadataSoftwareRiskReason   = "software_risk_reason"
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
	AlertReasonRiskySoftwareProcess   = "risky software process observed"
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
