package constants

const (
	EventTypeProcessObserved = "process.observed"
	EventTypeBrowserObserved = "browser.domain.observed"
	EventTypeAlertRaised     = "alert.raised"
)

const (
	EventSourceProcessCollector = "collector.process"
	EventSourceBrowserCollector = "collector.browser.history"
)

const (
	EventMetadataProfile         = "profile"
	EventMetadataOperatingSystem = "operating_system"
	EventMetadataBrowserName     = "browser_name"
	EventMetadataDomain          = "domain"
	EventMetadataCategory        = "category"
	EventMetadataURLMode         = "url_mode"
	EventMetadataStoredURLMode   = "stored_url_mode"
	EventMetadataVisitCount      = "visit_count"
	EventMetadataYouTubeStudy    = "youtube_study_match"
	EventMetadataYouTubeVideoID  = "youtube_video_id_hash"
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
