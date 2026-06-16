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
		"browser-activity-button",
		"/browser-activity",
		"theme-toggle-button",
		"server-status-light",
		"sourceBadge",
		"dashboard-page-nav",
		"Phase 82 product polish",
		"<span class=\"brand-mark\" aria-hidden=\"true\"><span></span><span></span><span></span></span>",
		"data-page-target=\"notifications\"",
		"Premium Operations Hub",
		"Premium Value Tiles",
		"Anomaly Notification Wall",
		"Mail And Push Delivery Ops",
		"Premium Owner Actions",
		"data-jump-target=\"premium-operations-section\"",
		"Monetisation Overview",
		"Anomaly Notification Proof",
		"Package And Revenue Fit",
		"Owner Action Queue",
		"Trust And Delivery Guard",
		"Tenant Onboarding Center",
		"Setup Checklist",
		"Role Handoff",
		"Onboarding Proof",
		"Onboarding Owner Actions",
		"data-jump-target=\"onboarding-center-section\"",
		"Customer Settings Center",
		"Settings Matrix",
		"Plan And Retention Options",
		"Notification Channel Settings",
		"Settings Owner Actions",
		"data-jump-target=\"customer-settings-section\"",
		"Revenue Operations Center",
		"Revenue Signals",
		"Anomaly And Delivery Wall",
		"Mail, Push, Dashboard Proof",
		"Commercial Levers",
		"Revenue Owner Actions",
		"data-jump-target=\"revenue-operations-section\"",
		"Runtime Status Center",
		"Runtime Proof",
		"Operator Actions",
		"runtime-status-badge",
		"runtime-proof-list",
		"runtime-action-list",
		"data-jump-target=\"runtime-status-section\"",
		"Deployment Readiness Center",
		"Platform Service Proof",
		"Service Manifest Proof",
		"Boot And Replay Proof",
		"Deployment Owner Actions",
		"data-jump-target=\"deployment-readiness-section\"",
		"Workspace Navigator",
		"class=\"command-label\">Premium Operations</span>",
		"class=\"command-label\">Onboarding Center</span>",
		"class=\"command-label\">Customer Settings</span>",
		"class=\"command-label\">Revenue Operations</span>",
		"class=\"command-label\">Runtime Status</span>",
		"class=\"command-label\">Deployment Readiness</span>",
		"class=\"command-label\">Customer Control Room</span>",
		"class=\"command-label\">Customer Success Packet</span>",
		"class=\"command-label\">Provider Setup</span>",
		"class=\"command-label\">Paid Operations</span>",
		"class=\"command-label\">Delivery Assurance</span>",
		"class=\"command-label\">Trust &amp; Consent</span>",
		"Customer Control Room",
		"Customer Value Tiles",
		"Anomaly Command Wall",
		"Mail And Push Delivery",
		"Owner Monetisation Actions",
		"data-jump-target=\"customer-control-section\"",
		"Customer Success Packet",
		"Success Proof Stack",
		"Buyer Objection Answers",
		"Success Packet Actions",
		"Delivery And Trust Promise",
		"data-jump-target=\"customer-success-section\"",
		"Push Activation Center",
		"Push Route Proof",
		"Anomaly Push And Mail Scenarios",
		"Push Owner Actions",
		"Push Privacy Guard",
		"data-jump-target=\"push-activation-section\"",
		"Portfolio Center",
		"Portfolio Alert Notifications",
		"Portfolio Delivery Proof",
		"Host Portfolio Rows",
		"Portfolio Segments",
		"Portfolio Owner Actions",
		"Portfolio Privacy Guard",
		"data-jump-target=\"portfolio-center-section\"",
		"Account Portfolio Index",
		"Account Tenant Rows",
		"Account Proof Cards",
		"Account Owner Actions",
		"data-jump-target=\"account-portfolio-section\"",
		"Executive Notification Console",
		"Value Tiles",
		"Anomaly Alert Stream",
		"Mail And Push Proof",
		"Owner Action Board",
		"data-jump-target=\"executive-console-section\"",
		"Notification Revenue Cockpit",
		"Revenue KPI Proof",
		"Anomaly Delivery Scenarios",
		"Channel Proof Matrix",
		"Upgrade Action Levers",
		"data-jump-target=\"notification-revenue-section\"",
		"Provider Simulation Lab",
		"Simulation Route Proof",
		"Simulation Scenarios",
		"Simulation Action Queue",
		"Provider Privacy Proof",
		"data-jump-target=\"provider-simulation-section\"",
		"Notification Provider Setup Center",
		"Provider Channel Setup",
		"Provider Setup Checklist",
		"Provider Setup Actions",
		"data-jump-target=\"notification-provider-setup-section\"",
		"Package Billing Readiness",
		"Plan Fit Matrix",
		"Feature Gate Proof",
		"Billing Milestones",
		"Upgrade Actions",
		"data-jump-target=\"package-billing-section\"",
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
		"Delivery Assurance Center",
		"Route Truth Matrix",
		"Delivery Truth Events",
		"Provider Proof Readiness",
		"data-jump-target=\"delivery-assurance-section\"",
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

	for _, forbidden := range []string{
		"<span class=\"brand-mark\" aria-hidden=\"true\">TD</span>",
		"Browser{",
		"Center{",
		"[B]",
		"{C}",
	} {
		if strings.Contains(html, forbidden) {
			t.Fatalf("dashboard contains stale visual marker %q", forbidden)
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

	ids, duplicates := dashboardElementIDs(html)
	if len(duplicates) > 0 {
		t.Fatalf("browser activity page contains duplicate DOM ids: %s", strings.Join(duplicates, ", "))
	}

	referenced := dashboardReferencedIDs(html)
	var missing []string
	for _, id := range referenced {
		if _, ok := ids[id]; !ok {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("browser activity JavaScript references missing DOM ids: %s", strings.Join(missing, ", "))
	}

	for _, marker := range []string{
		"TraceDeck Browser Activity",
		"Browser Activity Viewer",
		"Phase 82 product polish",
		"<span class=\"brand-mark\" aria-hidden=\"true\"><span></span><span></span><span></span></span>",
		"theme-toggle-button",
		"server-status-light",
		"Chrome",
		"Edge",
		"Brave",
		"Non-Study YouTube",
		"Notification Proof",
		"Host Breakdown",
		"Browser Domain Activity",
		"<th>Source</th>",
		"sourceBadge",
		"metadata-only guard",
	} {
		if !strings.Contains(html, marker) {
			t.Fatalf("browser activity page is missing marker %q", marker)
		}
	}

	for _, forbidden := range []string{
		"raw_url",
		"page_title",
		"screenshot_bytes",
		"password_value",
		"cookie_value",
		"token_value",
		"<span class=\"brand-mark\" aria-hidden=\"true\">TD</span>",
		"Browser{",
		"Center{",
		"[B]",
		"{C}",
	} {
		if strings.Contains(strings.ToLower(html), strings.ToLower(forbidden)) {
			t.Fatalf("browser activity page contains forbidden marker %q", forbidden)
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
