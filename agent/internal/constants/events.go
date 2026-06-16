package constants

const (
	EventTypeProcessObserved = "process.observed"
	EventTypeBrowserObserved = "browser.domain.observed"
	EventTypeAlertRaised     = "alert.raised"
	EventTypeDeviceHealth    = "device.health.observed"
	EventTypeAgentHeartbeat  = "agent.health.heartbeat"
)

const (
	EventSourceProcessCollector = "collector.process"
	EventSourceBrowserCollector = "collector.browser.history"
	EventSourceHealthCollector  = "collector.device.health"
	EventSourceHeartbeat        = "collector.agent.heartbeat"
)

const (
	OpenTelemetryProtocolOTLPHTTPJSON = "otlp_http_json"
	OpenTelemetryContentTypeJSON      = "application/json"
	OpenTelemetryServiceNameKey       = "service.name"
	OpenTelemetryServiceVersionKey    = "service.version"
	OpenTelemetryScopeName            = "github.com/varadharajaan/tracedeck-agent/agent/internal/exporter"
	OpenTelemetryLogBody              = "tracedeck metadata event"
	OpenTelemetryPrivacyBoundary      = "metadata_only_otlp_logs"
)

const (
	OpenTelemetryAttrTenantID        = "tracedeck.tenant_id"
	OpenTelemetryAttrDeviceID        = "tracedeck.device_id"
	OpenTelemetryAttrHostName        = "host.name"
	OpenTelemetryAttrOSName          = "os.type"
	OpenTelemetryAttrProfile         = "tracedeck.profile"
	OpenTelemetryAttrPrivacyBoundary = "tracedeck.privacy_boundary"
	OpenTelemetryAttrEventID         = "tracedeck.event.id"
	OpenTelemetryAttrEventType       = "event.name"
	OpenTelemetryAttrEventSource     = "event.source"
	OpenTelemetryAttrAppName         = "process.executable.name"
	OpenTelemetryAttrProcessID       = "process.pid"
	OpenTelemetryAttrPathHash        = "tracedeck.path_hash"
	OpenTelemetryAttrMetadataPrefix  = "tracedeck.metadata."
)

const (
	OpenTelemetrySensitiveKeyPassword       = "password"
	OpenTelemetrySensitiveKeyCredential     = "credential"
	OpenTelemetrySensitiveKeyCookie         = "cookie"
	OpenTelemetrySensitiveKeyToken          = "token"
	OpenTelemetrySensitiveKeyScreenshot     = "screenshot"
	OpenTelemetrySensitiveKeyKeystroke      = "keystroke"
	OpenTelemetrySensitiveKeyPrivateMessage = "private_message"
	OpenTelemetrySensitiveKeyPageTitle      = "page_title"
	OpenTelemetrySensitiveKeyRawURL         = "raw_url"
	OpenTelemetrySensitiveKeyFullURL        = "full_url"
	OpenTelemetrySensitiveKeyProviderSecret = "provider_secret"
	OpenTelemetrySensitiveKeyPayment        = "payment"
	OpenTelemetrySensitiveKeyCard           = "card"
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
	EventMetadataHealthScore     = "health_score"
	EventMetadataCPUPercent      = "cpu_percent"
	EventMetadataMemoryPercent   = "memory_percent"
	EventMetadataDiskPercent     = "disk_percent"
	EventMetadataBootTimeUnix    = "boot_time_unix"
	EventMetadataUptimeSeconds   = "uptime_seconds"
	EventMetadataHealthStatus    = "health_status"
	EventMetadataAgentHealthy    = "agent_healthy"
	EventMetadataAgentVersion    = "agent_version"
	EventMetadataCollectionMode  = "collection_mode"
	EventMetadataCollectionEvery = "collection_interval"
	EventMetadataArchiveEnabled  = "archive_enabled"
	EventMetadataArchiveDue      = "archive_due"
	EventMetadataBackendSync     = "backend_sync_enabled"
	EventMetadataAlertsEnabled   = "alerts_enabled"
)

const (
	HeartbeatCollectionModeOnce       = "once"
	HeartbeatCollectionModeContinuous = "continuous"
)

const (
	HealthStatusHealthy   = "healthy"
	HealthStatusWatch     = "watch"
	HealthStatusAttention = "attention"
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
