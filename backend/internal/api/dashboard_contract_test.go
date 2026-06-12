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

	ids, duplicates := dashboardElementIDs(html)
	if len(duplicates) > 0 {
		t.Fatalf("dashboard contains duplicate DOM ids: %s", strings.Join(duplicates, ", "))
	}

	referenced := dashboardReferencedIDs(html)
	var missing []string
	for _, id := range referenced {
		if _, ok := ids[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("dashboard JavaScript references missing DOM ids: %s", strings.Join(missing, ", "))
	}

	for _, id := range dashboardJumpTargets(html) {
		if _, ok := ids[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("dashboard command navigation references missing DOM ids: %s", strings.Join(missing, ", "))
	}

	for _, marker := range []string{
		"Command Navigation",
		"Business Dashboard",
		"Anomaly Notification Inbox",
		"Push And Mail Proof",
		"Paid Package Value",
		"Customer Owner Actions",
		"Growth Cockpit",
		"Anomaly Notification Ops",
		"Notification Delivery Proof",
		"Monetisation Owner Actions",
		"Notification Preference Center",
		"Preference Rule Matrix",
		"Study-Safe Suppression",
		"Role Experience Center",
		"Paid Onboarding Checklist",
		"Monetisation Command Center",
		"Anomaly And Notification Inbox",
		"Delivery And Mail Proof",
		"Owner Action Queue",
		"data-jump-target=\"premium-notification-section\"",
		"data-jump-target=\"paid-ops-section\"",
		"data-jump-target=\"revenue-section\"",
		"data-jump-target=\"notification-proof-section\"",
		"data-jump-target=\"mail-report-section\"",
		"data-jump-target=\"archive-proof-section\"",
		"data-jump-target=\"trust-proof-section\"",
		"data-jump-target=\"host-detail-section\"",
		"Paid Ops Console",
		"Commercial Control Room",
		"Revenue Command Center",
		"Premium Notification Command Center",
		"Notification Assurance Funnel",
		"Mail And Push Delivery Proof",
		"Customer Action SLAs",
		"Notification Proof Rail",
		"Notification Evidence Timeline",
		"Delivery Audit Trail",
		"Buyer Demo Checklist",
		"Mail Delivery Center",
		"Push Notification Center",
		"Provider-Safe Delivery Drilldown",
		"Delivery Rehearsal Actions",
		"Delivery Remediation Center",
		"Remediation Action Ledger",
		"Remediation SLA",
		"Archive Retention",
		"Tamper Trust",
		"Backend Alert Inbox",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("dashboard is missing monetisation marker %q", marker)
		}
	}
}

func dashboardJumpTargets(html string) []string {
	pattern := regexp.MustCompile(`data-jump-target="([^"]+)"`)
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
