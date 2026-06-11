package alert

import (
	"context"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

type Alert struct {
	RuleName   string            `json:"rule_name"`
	Severity   string            `json:"severity"`
	Reason     string            `json:"reason"`
	TenantID   string            `json:"tenant_id"`
	DeviceID   string            `json:"device_id"`
	HostName   string            `json:"host_name"`
	AppName    string            `json:"app_name"`
	ObservedAt time.Time         `json:"observed_at"`
	Metadata   map[string]string `json:"metadata"`
}

type Evaluator struct{}

func NewEvaluator() *Evaluator {
	return &Evaluator{}
}

func (e *Evaluator) Evaluate(_ context.Context, policy *config.Policy, events []event.Event) []Alert {
	rule, ok := policy.AlertRules[constants.AlertRuleBlockedAppOpened]
	if !ok || !rule.Enabled {
		return nil
	}

	blockedApps := normalizedSet(policy.BlockedApps)
	if len(blockedApps) == 0 {
		return nil
	}

	seenApps := make(map[string]struct{})
	alerts := make([]Alert, 0)
	for _, evt := range events {
		if evt.Type != constants.EventTypeProcessObserved {
			continue
		}
		normalizedApp := normalize(evt.AppName)
		if _, blocked := blockedApps[normalizedApp]; !blocked {
			continue
		}
		if _, seen := seenApps[normalizedApp]; seen {
			continue
		}
		seenApps[normalizedApp] = struct{}{}

		alerts = append(alerts, Alert{
			RuleName:   constants.AlertRuleBlockedAppOpened,
			Severity:   string(rule.Severity),
			Reason:     constants.AlertReasonBlockedAppObserved,
			TenantID:   evt.TenantID,
			DeviceID:   evt.DeviceID,
			HostName:   evt.HostName,
			AppName:    evt.AppName,
			ObservedAt: evt.Timestamp,
			Metadata: map[string]string{
				constants.AlertMetadataRuleName: constants.AlertRuleBlockedAppOpened,
				constants.AlertMetadataReason:   constants.AlertReasonBlockedAppObserved,
				constants.AlertMetadataSeverity: string(rule.Severity),
				constants.AlertMetadataAppName:  evt.AppName,
			},
		})
	}

	return filterBySeverity(alerts, policy.Alerts.Email.MinSeverity)
}

func normalizedSet(values []string) map[string]struct{} {
	out := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalize(value)
		if normalized != "" {
			out[normalized] = struct{}{}
		}
	}
	return out
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func filterBySeverity(alerts []Alert, minSeverity config.Severity) []Alert {
	if minSeverity == "" {
		return alerts
	}

	out := alerts[:0]
	minRank := severityRank(string(minSeverity))
	for _, candidate := range alerts {
		if severityRank(candidate.Severity) >= minRank {
			out = append(out, candidate)
		}
	}
	return out
}

func severityRank(severity string) int {
	switch severity {
	case constants.SeverityCritical:
		return 4
	case constants.SeverityHigh:
		return 3
	case constants.SeverityMedium:
		return 2
	case constants.SeverityLow:
		return 1
	default:
		return 0
	}
}
