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
