package constants

const (
	EventTypeProcessObserved = "process.observed"
	EventTypeAlertRaised     = "alert.raised"
)

const (
	EventSourceProcessCollector = "collector.process"
)

const (
	EventMetadataProfile         = "profile"
	EventMetadataOperatingSystem = "operating_system"
)

const (
	AlertRuleBlockedAppOpened = "blocked_app_opened"
)

const (
	AlertMetadataRuleName = "rule_name"
	AlertMetadataReason   = "reason"
	AlertMetadataSeverity = "severity"
	AlertMetadataAppName  = "app_name"
)

const (
	AlertReasonBlockedAppObserved = "blocked app observed in process snapshot"
)
