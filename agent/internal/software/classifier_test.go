package software

import (
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestClassifyProcessDetectsRiskySoftware(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		path     string
		category string
	}{
		{name: "qBittorrent.exe", category: constants.SoftwareRiskCategoryTorrentClient},
		{name: "nordvpn.exe", category: constants.SoftwareRiskCategoryVPNProxy},
		{name: "steam.exe", category: constants.SoftwareRiskCategoryGameLauncher},
		{name: "opera.exe", category: constants.SoftwareRiskCategoryUnknownBrowser},
		{name: "setup.exe", path: `C:\Users\student\Downloads\setup.exe`, category: constants.SoftwareRiskCategoryDownloadsInstaller},
	} {
		risk, ok := ClassifyProcess(tc.name, tc.path)
		if !ok {
			t.Fatalf("expected %s to be risky", tc.name)
		}
		if risk.Category != tc.category {
			t.Fatalf("category for %s = %q, want %q", tc.name, risk.Category, tc.category)
		}
		if risk.Reason == "" {
			t.Fatalf("expected reason for %s", tc.name)
		}
	}
}

func TestClassifyProcessIgnoresStandardStudyBrowser(t *testing.T) {
	t.Parallel()

	if risk, ok := ClassifyProcess("chrome.exe", `C:\Program Files\Google\Chrome\Application\chrome.exe`); ok {
		t.Fatalf("expected chrome to be allowed, got %+v", risk)
	}
}
