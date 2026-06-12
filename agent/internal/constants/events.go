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
	AlertRuleBlockedAppOpened  = "blocked_app_opened"
	AlertRuleBlockedDomainOpen = "blocked_domain_opened"
	AlertRuleNonStudyYouTube   = "non_study_youtube"
)

const (
	AlertMetadataRuleName = "rule_name"
	AlertMetadataReason   = "reason"
	AlertMetadataSeverity = "severity"
	AlertMetadataAppName  = "app_name"
	AlertMetadataDomain   = "domain"
	AlertMetadataCategory = "category"
)

const (
	AlertReasonBlockedAppObserved      = "blocked app observed in process snapshot"
	AlertReasonBlockedDomainObserved   = "blocked domain observed in browser activity"
	AlertReasonNonStudyYouTubeObserved = "non-study youtube activity observed"
)
