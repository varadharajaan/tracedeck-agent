package alert

import (
	"context"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

func TestEvaluatorDetectsBlockedApp(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID:    constants.DefaultTenantID,
		DeviceID:    constants.DefaultDeviceID,
		BlockedApps: []string{"vlc.exe"},
		Alerts: config.AlertPolicy{Email: config.EmailPolicy{
			MinSeverity: config.Severity(constants.SeverityHigh),
		}},
		AlertRules: map[string]config.RuleSpec{
			constants.AlertRuleBlockedAppOpened: {
				Enabled:  true,
				Severity: config.Severity(constants.SeverityHigh),
			},
		},
	}

	alerts := NewEvaluator().Evaluate(context.Background(), policy, []event.Event{{
		Type:      constants.EventTypeProcessObserved,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   "VLC.EXE",
	}})

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RuleName != constants.AlertRuleBlockedAppOpened {
		t.Fatalf("unexpected rule: %s", alerts[0].RuleName)
	}
}

func TestEvaluatorDetectsBlockedForegroundApp(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID:    constants.DefaultTenantID,
		DeviceID:    constants.DefaultDeviceID,
		BlockedApps: []string{"vlc.exe"},
		Alerts: config.AlertPolicy{Email: config.EmailPolicy{
			MinSeverity: config.Severity(constants.SeverityHigh),
		}},
		AlertRules: map[string]config.RuleSpec{
			constants.AlertRuleBlockedAppOpened: {
				Enabled:  true,
				Severity: config.Severity(constants.SeverityHigh),
			},
		},
	}

	alerts := NewEvaluator().Evaluate(context.Background(), policy, []event.Event{{
		Type:      constants.EventTypeForegroundAppObserved,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   "VLC.EXE",
	}})

	if len(alerts) != 1 {
		t.Fatalf("expected 1 foreground app alert, got %d", len(alerts))
	}
	if alerts[0].RuleName != constants.AlertRuleBlockedAppOpened {
		t.Fatalf("unexpected rule: %s", alerts[0].RuleName)
	}
}

func TestEvaluatorDetectsNonStudyYouTube(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Alerts: config.AlertPolicy{Email: config.EmailPolicy{
			MinSeverity: config.Severity(constants.SeverityMedium),
		}},
		AlertRules: map[string]config.RuleSpec{
			constants.AlertRuleNonStudyYouTube: {
				Enabled:  true,
				Severity: config.Severity(constants.SeverityMedium),
			},
		},
	}

	alerts := NewEvaluator().Evaluate(context.Background(), policy, []event.Event{{
		Type:      constants.EventTypeBrowserObserved,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   constants.BrowserNameChrome,
		Metadata: map[string]string{
			constants.EventMetadataDomain:   constants.DomainYouTubeLong,
			constants.EventMetadataCategory: constants.CategoryVideoStreaming,
		},
	}})

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RuleName != constants.AlertRuleNonStudyYouTube {
		t.Fatalf("unexpected rule: %s", alerts[0].RuleName)
	}
	if alerts[0].Metadata[constants.AlertMetadataDomain] != constants.DomainYouTubeLong {
		t.Fatalf("expected domain metadata: %+v", alerts[0].Metadata)
	}
}

func TestEvaluatorIgnoresStudyYouTube(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Alerts: config.AlertPolicy{Email: config.EmailPolicy{
			MinSeverity: config.Severity(constants.SeverityMedium),
		}},
		AlertRules: map[string]config.RuleSpec{
			constants.AlertRuleNonStudyYouTube: {
				Enabled:  true,
				Severity: config.Severity(constants.SeverityMedium),
			},
		},
	}

	alerts := NewEvaluator().Evaluate(context.Background(), policy, []event.Event{{
		Type:      constants.EventTypeBrowserObserved,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   constants.BrowserNameChrome,
		Metadata: map[string]string{
			constants.EventMetadataDomain:       constants.DomainYouTubeLong,
			constants.EventMetadataCategory:     constants.CategoryStudy,
			constants.EventMetadataYouTubeStudy: "true",
		},
	}})

	if len(alerts) != 0 {
		t.Fatalf("expected no alerts, got %d", len(alerts))
	}
}

func TestEvaluatorDetectsBlockedDomain(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID:       constants.DefaultTenantID,
		DeviceID:       constants.DefaultDeviceID,
		BlockedDomains: []string{"example.com"},
		Alerts: config.AlertPolicy{Email: config.EmailPolicy{
			MinSeverity: config.Severity(constants.SeverityHigh),
		}},
		AlertRules: map[string]config.RuleSpec{
			constants.AlertRuleBlockedDomainOpen: {
				Enabled:  true,
				Severity: config.Severity(constants.SeverityHigh),
			},
		},
	}

	alerts := NewEvaluator().Evaluate(context.Background(), policy, []event.Event{{
		Type:      constants.EventTypeBrowserObserved,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   constants.BrowserNameChrome,
		Metadata: map[string]string{
			constants.EventMetadataDomain:   "media.example.com",
			constants.EventMetadataCategory: constants.CategoryBlocked,
		},
	}})

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RuleName != constants.AlertRuleBlockedDomainOpen {
		t.Fatalf("unexpected rule: %s", alerts[0].RuleName)
	}
}

func TestEvaluatorDetectsRiskySoftware(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Alerts: config.AlertPolicy{Email: config.EmailPolicy{
			MinSeverity: config.Severity(constants.SeverityHigh),
		}},
		AlertRules: map[string]config.RuleSpec{
			constants.AlertRuleRiskySoftwareDetected: {
				Enabled:  true,
				Severity: config.Severity(constants.SeverityHigh),
			},
		},
	}

	alerts := NewEvaluator().Evaluate(context.Background(), policy, []event.Event{{
		Type:      constants.EventTypeProcessObserved,
		Timestamp: time.Now().UTC(),
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   "qbittorrent.exe",
		Metadata: map[string]string{
			constants.EventMetadataSoftwareRiskCategory: constants.SoftwareRiskCategoryTorrentClient,
			constants.EventMetadataSoftwareRiskReason:   constants.SoftwareRiskReasonTorrentClient,
		},
	}})

	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].RuleName != constants.AlertRuleRiskySoftwareDetected {
		t.Fatalf("unexpected rule: %s", alerts[0].RuleName)
	}
	if alerts[0].Metadata[constants.AlertMetadataSoftwareRiskCategory] != constants.SoftwareRiskCategoryTorrentClient {
		t.Fatalf("expected software risk category metadata: %+v", alerts[0].Metadata)
	}
}
