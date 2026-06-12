package software

import (
	"path/filepath"
	"strings"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type Risk struct {
	Category string
	Reason   string
}

func ClassifyProcess(processName string, executablePath string) (Risk, bool) {
	name := normalizeProcessName(processName)
	switch {
	case contains(constants.RiskyTorrentProcessNames, name):
		return Risk{Category: constants.SoftwareRiskCategoryTorrentClient, Reason: constants.SoftwareRiskReasonTorrentClient}, true
	case contains(constants.RiskyVPNProxyProcessNames, name):
		return Risk{Category: constants.SoftwareRiskCategoryVPNProxy, Reason: constants.SoftwareRiskReasonVPNProxy}, true
	case contains(constants.RiskyGameLauncherProcessNames, name):
		return Risk{Category: constants.SoftwareRiskCategoryGameLauncher, Reason: constants.SoftwareRiskReasonGameLauncher}, true
	case contains(constants.UnknownBrowserProcessNames, name) && !contains(constants.KnownBrowserProcessNames, name):
		return Risk{Category: constants.SoftwareRiskCategoryUnknownBrowser, Reason: constants.SoftwareRiskReasonUnknownBrowser}, true
	case isDownloadsInstaller(name, executablePath):
		return Risk{Category: constants.SoftwareRiskCategoryDownloadsInstaller, Reason: constants.SoftwareRiskReasonDownloadsInstaller}, true
	default:
		return Risk{}, false
	}
}

func normalizeProcessName(value string) string {
	return strings.ToLower(strings.TrimSpace(filepath.Base(value)))
}

func contains(values []string, candidate string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), candidate) {
			return true
		}
	}
	return false
}

func isDownloadsInstaller(processName string, executablePath string) bool {
	path := strings.ToLower(strings.ReplaceAll(executablePath, "/", `\`))
	if !strings.Contains(path, `\downloads\`) {
		return false
	}
	return strings.Contains(processName, "setup") ||
		strings.Contains(processName, "install") ||
		strings.Contains(processName, "installer")
}
