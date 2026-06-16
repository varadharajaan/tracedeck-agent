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
	alerts := make([]Alert, 0)
	alerts = append(alerts, e.evaluateBlockedApps(policy, events)...)
	alerts = append(alerts, e.evaluateRiskySoftware(policy, events)...)
	alerts = append(alerts, e.evaluateUnknownSoftwareInstalled(policy, events)...)
	alerts = append(alerts, e.evaluateBlockedDomains(policy, events)...)
	alerts = append(alerts, e.evaluateNonStudyYouTube(policy, events)...)
	return filterBySeverity(alerts, policy.Alerts.Email.MinSeverity)
}

func (e *Evaluator) evaluateBlockedApps(policy *config.Policy, events []event.Event) []Alert {
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
		if !isAppUsageEvent(evt) && !isSoftwareInventoryEvent(evt) {
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

	return alerts
}

func (e *Evaluator) evaluateRiskySoftware(policy *config.Policy, events []event.Event) []Alert {
	rule, ok := policy.AlertRules[constants.AlertRuleRiskySoftwareDetected]
	if !ok || !rule.Enabled {
		return nil
	}

	seenApps := make(map[string]struct{})
	alerts := make([]Alert, 0)
	for _, evt := range events {
		if !isAppUsageEvent(evt) && !isSoftwareInventoryEvent(evt) {
			continue
		}
		category := normalize(evt.Metadata[constants.EventMetadataSoftwareRiskCategory])
		if category == "" {
			continue
		}
		normalizedApp := normalize(evt.AppName)
		if _, seen := seenApps[normalizedApp]; seen {
			continue
		}
		seenApps[normalizedApp] = struct{}{}

		reason := evt.Metadata[constants.EventMetadataSoftwareRiskReason]
		if reason == "" {
			reason = constants.AlertReasonRiskySoftwareProcess
		}
		alerts = append(alerts, Alert{
			RuleName:   constants.AlertRuleRiskySoftwareDetected,
			Severity:   string(rule.Severity),
			Reason:     reason,
			TenantID:   evt.TenantID,
			DeviceID:   evt.DeviceID,
			HostName:   evt.HostName,
			AppName:    evt.AppName,
			ObservedAt: evt.Timestamp,
			Metadata: map[string]string{
				constants.AlertMetadataRuleName:             constants.AlertRuleRiskySoftwareDetected,
				constants.AlertMetadataReason:               reason,
				constants.AlertMetadataSeverity:             string(rule.Severity),
				constants.AlertMetadataAppName:              evt.AppName,
				constants.AlertMetadataSoftwareRiskCategory: category,
				constants.AlertMetadataSoftwareRiskReason:   reason,
			},
		})
	}

	return alerts
}

func (e *Evaluator) evaluateUnknownSoftwareInstalled(policy *config.Policy, events []event.Event) []Alert {
	rule, ok := policy.AlertRules[constants.AlertRuleUnknownSoftwareInstalled]
	if !ok || !rule.Enabled {
		return nil
	}

	seenApps := make(map[string]struct{})
	alerts := make([]Alert, 0)
	for _, evt := range events {
		if evt.Type != constants.EventTypeSoftwareInstalled {
			continue
		}
		normalizedApp := normalize(evt.AppName)
		if normalizedApp == "" {
			continue
		}
		if _, seen := seenApps[normalizedApp]; seen {
			continue
		}
		seenApps[normalizedApp] = struct{}{}

		alerts = append(alerts, Alert{
			RuleName:   constants.AlertRuleUnknownSoftwareInstalled,
			Severity:   string(rule.Severity),
			Reason:     constants.AlertReasonSoftwareInstalled,
			TenantID:   evt.TenantID,
			DeviceID:   evt.DeviceID,
			HostName:   evt.HostName,
			AppName:    evt.AppName,
			ObservedAt: evt.Timestamp,
			Metadata: map[string]string{
				constants.AlertMetadataRuleName: constants.AlertRuleUnknownSoftwareInstalled,
				constants.AlertMetadataReason:   constants.AlertReasonSoftwareInstalled,
				constants.AlertMetadataSeverity: string(rule.Severity),
				constants.AlertMetadataAppName:  evt.AppName,
			},
		})
	}

	return alerts
}

func (e *Evaluator) evaluateBlockedDomains(policy *config.Policy, events []event.Event) []Alert {
	rule, ok := policy.AlertRules[constants.AlertRuleBlockedDomainOpen]
	if !ok || !rule.Enabled {
		return nil
	}
	if len(policy.BlockedDomains) == 0 {
		return nil
	}

	seenDomains := make(map[string]struct{})
	alerts := make([]Alert, 0)
	for _, evt := range events {
		if evt.Type != constants.EventTypeBrowserObserved {
			continue
		}
		domain := normalize(evt.Metadata[constants.EventMetadataDomain])
		if domain == "" || !domainMatchesAny(domain, policy.BlockedDomains) {
			continue
		}
		if _, seen := seenDomains[domain]; seen {
			continue
		}
		seenDomains[domain] = struct{}{}
		alerts = append(alerts, browserAlert(constants.AlertRuleBlockedDomainOpen, rule, constants.AlertReasonBlockedDomainObserved, evt))
	}
	return alerts
}

func (e *Evaluator) evaluateNonStudyYouTube(policy *config.Policy, events []event.Event) []Alert {
	rule, ok := policy.AlertRules[constants.AlertRuleNonStudyYouTube]
	if !ok || !rule.Enabled {
		return nil
	}

	seenDomains := make(map[string]struct{})
	alerts := make([]Alert, 0)
	for _, evt := range events {
		if evt.Type != constants.EventTypeBrowserObserved {
			continue
		}
		domain := normalize(evt.Metadata[constants.EventMetadataDomain])
		category := normalize(evt.Metadata[constants.EventMetadataCategory])
		studyMatch := normalize(evt.Metadata[constants.EventMetadataYouTubeStudy])
		if !isYouTubeDomain(domain) || category != constants.CategoryVideoStreaming || studyMatch == "true" {
			continue
		}
		if _, seen := seenDomains[domain]; seen {
			continue
		}
		seenDomains[domain] = struct{}{}
		alerts = append(alerts, browserAlert(constants.AlertRuleNonStudyYouTube, rule, constants.AlertReasonNonStudyYouTubeObserved, evt))
	}
	return alerts
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

func isAppUsageEvent(evt event.Event) bool {
	return evt.Type == constants.EventTypeProcessObserved || evt.Type == constants.EventTypeForegroundAppObserved
}

func isSoftwareInventoryEvent(evt event.Event) bool {
	return evt.Type == constants.EventTypeSoftwareInstalled || evt.Type == constants.EventTypeSoftwareUninstalled
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func browserAlert(ruleName string, rule config.RuleSpec, reason string, evt event.Event) Alert {
	domain := evt.Metadata[constants.EventMetadataDomain]
	category := evt.Metadata[constants.EventMetadataCategory]
	return Alert{
		RuleName:   ruleName,
		Severity:   string(rule.Severity),
		Reason:     reason,
		TenantID:   evt.TenantID,
		DeviceID:   evt.DeviceID,
		HostName:   evt.HostName,
		AppName:    evt.AppName,
		ObservedAt: evt.Timestamp,
		Metadata: map[string]string{
			constants.AlertMetadataRuleName: ruleName,
			constants.AlertMetadataReason:   reason,
			constants.AlertMetadataSeverity: string(rule.Severity),
			constants.AlertMetadataDomain:   domain,
			constants.AlertMetadataCategory: category,
		},
	}
}

func domainMatchesAny(domain string, candidates []string) bool {
	domain = normalize(strings.TrimPrefix(domain, "www."))
	for _, candidate := range candidates {
		candidate = normalize(strings.TrimPrefix(candidate, "www."))
		if candidate == "" {
			continue
		}
		if domain == candidate || strings.HasSuffix(domain, "."+candidate) {
			return true
		}
	}
	return false
}

func isYouTubeDomain(domain string) bool {
	return domain == constants.DomainYouTubeLong ||
		strings.HasSuffix(domain, "."+constants.DomainYouTubeLong) ||
		domain == constants.DomainYouTubeShort
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
