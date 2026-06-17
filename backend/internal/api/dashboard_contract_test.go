package api

import (
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"testing"
)

func TestDashboardDOMContract(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(dashboardFS, "web/dashboard.html")
	if err != nil {
		t.Fatalf("read dashboard asset: %v", err)
	}
	html := string(data)

	assertNoDuplicateIDs(t, "dashboard", html)
	assertReferencedIDsExist(t, "dashboard", html)
	assertPageTargetsExist(t, html)

	for _, marker := range []string{
		"TraceDeck Console",
		"dashboard-page-nav",
		"command-navigation",
		"Command Center",
		"Host Portfolio",
		"Signal Queue",
		"Browser Intelligence",
		"Delivery Assurance",
		"Revenue Packaging",
		"Trust Center",
		"browser-activity-button",
		"legacy-dashboard-button",
		"/browser-activity",
		"/v1-old",
		"theme-toggle-button",
		"server-status-light",
		"include-demo-toggle",
		"Demo proof",
		"live evidence",
		"mode-badge",
		"Risk Posture",
		"Browser Mix",
		"Evidence Pipeline",
		"Executive Brief",
		"pipeline-status",
		"operator-brief-list",
		"Domain Rows",
		"Route Proof",
		"Commercial Proof",
		"Archive Retention",
		"Runtime Proof",
		"Metadata-only privacy boundary",
		"<span class=\"brand-mark\" aria-hidden=\"true\"><span></span><span></span><span></span></span>",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("dashboard is missing modern console marker %q", marker)
		}
	}

	assertNoLegacyVisualMarkers(t, "dashboard", html)
}

func TestDashboardV1OldAssetContract(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(dashboardFS, "web/dashboard_v1_old.html")
	if err != nil {
		t.Fatalf("read legacy dashboard asset: %v", err)
	}
	html := string(data)

	assertNoDuplicateIDs(t, "legacy dashboard", html)
	for _, marker := range []string{
		"TraceDeck Dashboard",
		"v1-old legacy containment",
		"dashboard-page-nav",
		"Workspace Navigator",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("legacy dashboard is missing marker %q", marker)
		}
	}
}

func TestBrowserActivityDOMContract(t *testing.T) {
	t.Parallel()

	data, err := fs.ReadFile(dashboardFS, "web/browser_activity.html")
	if err != nil {
		t.Fatalf("read browser activity asset: %v", err)
	}
	html := string(data)

	assertNoDuplicateIDs(t, "browser activity page", html)
	assertReferencedIDsExist(t, "browser activity page", html)

	for _, marker := range []string{
		"TraceDeck Browser Intelligence",
		"Browser Intelligence",
		"theme-toggle-button",
		"server-status-light",
		"include-demo-toggle",
		"Demo proof",
		"source-mode-pill",
		"live evidence",
		"metadata-only guard",
		"Chrome",
		"Edge",
		"Brave",
		"YouTube Review",
		"Route Proof",
		"Host Coverage",
		"Browser Mix",
		"Domain Activity",
		"<th>Source</th>",
		"sourceBadge",
		"class=\"signal-cell\"",
		"<span class=\"brand-mark\" aria-hidden=\"true\"><span></span><span></span><span></span></span>",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("browser activity page is missing marker %q", marker)
		}
	}

	assertNoLegacyVisualMarkers(t, "browser activity page", html)
	for _, forbidden := range []string{
		"raw_url",
		"page_title",
		"screenshot_bytes",
		"password_value",
		"cookie_value",
		"token_value",
	} {
		if strings.Contains(strings.ToLower(html), strings.ToLower(forbidden)) {
			t.Fatalf("browser activity page contains forbidden marker %q", forbidden)
		}
	}
}

func assertNoDuplicateIDs(t *testing.T, label string, html string) {
	t.Helper()
	_, duplicates := dashboardElementIDs(html)
	if len(duplicates) > 0 {
		t.Fatalf("%s contains duplicate DOM ids: %s", label, strings.Join(duplicates, ", "))
	}
}

func assertReferencedIDsExist(t *testing.T, label string, html string) {
	t.Helper()
	ids, _ := dashboardElementIDs(html)
	var missing []string
	for _, id := range dashboardReferencedIDs(html) {
		if _, ok := ids[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("%s JavaScript references missing DOM ids: %s", label, strings.Join(missing, ", "))
	}
}

func assertPageTargetsExist(t *testing.T, html string) {
	t.Helper()
	ids, _ := dashboardElementIDs(html)
	var missing []string
	for _, target := range dashboardPageTargets(html) {
		pageID := target + "-page"
		if _, ok := ids[pageID]; !ok {
			missing = append(missing, pageID)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("dashboard page navigation references missing page ids: %s", strings.Join(missing, ", "))
	}
}

func assertNoLegacyVisualMarkers(t *testing.T, label string, html string) {
	t.Helper()
	for _, forbidden := range []string{
		"<span class=\"brand-mark\" aria-hidden=\"true\">TD</span>",
		"Browser{",
		"Center{",
		"[B]",
		"{C}",
		"Phase 82 product polish",
		"Workspace Navigator",
		"Premium Operations",
		"data-jump-target",
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("%s contains stale visual marker %q", label, forbidden)
		}
	}
}

func dashboardPageTargets(html string) []string {
	pattern := regexp.MustCompile(`data-page-target="([^"]+)"`)
	seen := map[string]struct{}{}
	for _, match := range pattern.FindAllStringSubmatch(html, -1) {
		seen[match[1]] = struct{}{}
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

func dashboardElementIDs(html string) (map[string]struct{}, []string) {
	html = regexp.MustCompile(`(?s)<script.*?</script>`).ReplaceAllString(html, "")
	idPattern := regexp.MustCompile(`\sid="([^"]+)"`)
	matches := idPattern.FindAllStringSubmatch(html, -1)
	ids := make(map[string]struct{}, len(matches))
	counts := make(map[string]int, len(matches))
	for _, match := range matches {
		counts[match[1]]++
		ids[match[1]] = struct{}{}
	}

	var duplicates []string
	for id, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}
	sort.Strings(duplicates)
	return ids, duplicates
}

func dashboardReferencedIDs(html string) []string {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`document\.getElementById\("([^"]+)"\)`),
		regexp.MustCompile(`setText\("([^"]+)"`),
		regexp.MustCompile(`setMetric\("([^"]+)"`),
		regexp.MustCompile(`setBar\("([^"]+)"`),
		regexp.MustCompile(`setPercentBar\("([^"]+)"`),
		regexp.MustCompile(`replace\("<span", '<span id="([^"]+)"`),
	}
	seen := map[string]struct{}{}
	for _, pattern := range patterns {
		for _, match := range pattern.FindAllStringSubmatch(html, -1) {
			seen[match[1]] = struct{}{}
		}
	}

	ids := make([]string, 0, len(seen))
	for id := range seen {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
