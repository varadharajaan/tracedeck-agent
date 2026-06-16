package constants

const (
	BackendName    = "tracedeck-backend"
	BackendVersion = "0.1.0-dev"
)

const (
	DefaultBackendAddr = "127.0.0.1:18080"
	DefaultLogDir      = "logs/local/backend"
	DefaultLogLevel    = "info"
	DefaultDataPath    = "data/local/backend/backend-state.json"
)

const (
	DefaultRuntimeSummaryPath = "data/local/output/runtime-summary.json"
	RuntimeSummaryCommand     = "python ./devctl.py summary"
	RuntimeTaskRestartCommand = "python ./devctl.py server task-restart"
)

const (
	APIPrefix               = "/api/v1"
	RouteDashboard          = "/"
	RouteBrowserActivity    = "/browser-activity"
	RouteHealth             = "/health"
	RouteVersion            = APIPrefix + "/version"
	RouteDevices            = APIPrefix + "/devices"
	RouteDeviceEnroll       = APIPrefix + "/devices/enroll"
	RouteTenants            = APIPrefix + "/tenants"
	RoutePlans              = APIPrefix + "/plans"
	RouteRoles              = APIPrefix + "/roles"
	RouteRetentionTiers     = APIPrefix + "/retention-tiers"
	RouteAuditEvents        = APIPrefix + "/audit-events"
	RoutePolicyTemplates    = APIPrefix + "/policy-templates"
	RouteAlertRuleTemplates = APIPrefix + "/alert-rule-templates"
	RouteAccountPortfolio   = APIPrefix + "/account-portfolio-index"
	RouteRuntimeStatus      = APIPrefix + "/runtime-status-center"
	RouteArchiveStatus      = APIPrefix + "/archive/status"
)

const (
	RouteSegmentSummary          = "summary"
	RouteSegmentDaily            = "daily"
	RouteSegmentReports          = "reports"
	RouteSegmentWeekly           = "weekly"
	RouteSegmentPDF              = "pdf"
	RouteSegmentOverview         = "overview"
	RouteSegmentAuditEvents      = "audit-events"
	RouteSegmentAlertRules       = "alert-rules"
	RouteSegmentConsentCenter    = "consent-center"
	RouteSegmentAlertInbox       = "alert-inbox"
	RouteSegmentOperations       = "operations-summary"
	RouteSegmentMonetization     = "monetization-summary"
	RouteSegmentBusinessDash     = "business-dashboard"
	RouteSegmentOnboardingCenter = "onboarding-center"
	RouteSegmentCustomerSettings = "customer-settings-center"
	RouteSegmentRevenueOps       = "revenue-operations-center"
	RouteSegmentDeploymentReady  = "deployment-readiness-center"
	RouteSegmentPremiumOps       = "premium-operations-hub"
	RouteSegmentRoleExperience   = "role-experiences"
	RouteSegmentCustomerControl  = "customer-control-room"
	RouteSegmentSuccessPacket    = "customer-success-packet"
	RouteSegmentPushActivation   = "push-activation-center"
	RouteSegmentPortfolioCenter  = "portfolio-center"
	RouteSegmentExecutiveConsole = "executive-console"
	RouteSegmentNotificationRev  = "notification-revenue-cockpit"
	RouteSegmentProviderSim      = "provider-simulation-lab"
	RouteSegmentProviderSetup    = "notification-provider-setup"
	RouteSegmentPackageBilling   = "package-billing-readiness"
	RouteSegmentNotificationCmd  = "notification-command-center"
	RouteSegmentDeliveryTimeline = "delivery-timeline"
	RouteSegmentDeliveryAssure   = "delivery-assurance"
	RouteSegmentDeliveryDrill    = "delivery-drilldown"
	RouteSegmentDeliveryRemedy   = "delivery-remediation"
	RouteSegmentSyncHealth       = "sync-health"
	RouteSegmentActivityFeed     = "activity-feed"
	RouteSegmentBrowserActivity  = "browser-activity"
	RouteSegmentActivityViews    = "activity-views"
	RouteSegmentNotifications    = "notification-routes"
	RouteSegmentNotificationPref = "notification-preferences"
	RouteSegmentDataExports      = "data-exports"
	RouteSegmentDeleteRequests   = "delete-requests"
	RouteSegmentDeviceGroups     = "device-groups"
	RouteSegmentPolicyAssign     = "policy-assignments"
	RouteSegmentPolicyEvents     = "policy-violations"
	RouteSegmentAnomalies        = "anomalies"
	RouteSegmentTamperEvents     = "tamper-events"
	RouteSegmentAlertDelivery    = "alert-deliveries"
	RouteSegmentHealth           = "health"
	RouteSegmentTelemetry        = "telemetry-events"
	RouteSegmentTelemetryStatus  = "telemetry-status"
)

const (
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html; charset=utf-8"
	ContentTypePDF  = "application/pdf"
)

const (
	HeaderAPIKey   = "X-TraceDeck-API-Key" // #nosec G101 -- HTTP header name only; the API key value is runtime input, never a hardcoded secret.
	HeaderTenantID = "X-TraceDeck-Tenant-ID"
	HeaderActorID  = "X-TraceDeck-Actor-ID"
	HeaderCache    = "Cache-Control"
	HeaderPragma   = "Pragma"
	HeaderExpires  = "Expires"
)

const (
	CacheNoStore  = "no-store, no-cache, must-revalidate, max-age=0"
	PragmaNoCache = "no-cache"
	ExpiresNow    = "0"
)

const (
	QueryIncludeDemo = "include_demo"
	QueryValueTrue   = "true"
	QueryValueOne    = "1"
	QueryValueYes    = "yes"
)

const (
	RoleParent          = "parent"
	RoleStudent         = "student"
	RoleSchoolAdmin     = "school_admin"
	RoleBusinessManager = "business_manager"
)

const (
	PlanFree       = "free"
	PlanFamilyPro  = "family_pro"
	PlanSchool     = "school"
	PlanBusiness   = "business"
	PlanEnterprise = "enterprise"
)

const (
	RetentionLocalOnly   = "local_only_7_days"
	RetentionFamilyCloud = "family_cloud_90_365_archive"
	RetentionSchoolYear  = "school_year_archive"
	RetentionBusiness    = "business_compliance"
)

const (
	TenantStatusActive = "active"
)

const (
	AuditCategoryTenant = "tenant"
	AuditCategoryDevice = "device"
	AuditCategoryPolicy = "policy"
	AuditCategoryAccess = "access"
	AuditCategorySystem = "system"
)

const (
	AuditActionTenantCreated        = "tenant.created"
	AuditActionAlertRuleCreated     = "alert_rule.created"
	AuditActionNotificationRoute    = "notification_route.created"
	AuditActionNotificationPref     = "notification_preferences.updated"
	AuditActionDeliveryDrillRun     = "delivery_drilldown.rehearsed"
	AuditActionDeliveryRemediation  = "delivery_remediation.planned"
	AuditActionProviderSimulation   = "provider_simulation.rehearsed"
	AuditActionActivityViewCreated  = "activity_view.created"
	AuditActionDeviceGroupCreated   = "device_group.created"
	AuditActionDataExportCreated    = "data_export.created"
	AuditActionDeleteRequestCreated = "delete_request.created"
	AuditActionPolicyAssigned       = "policy_assignment.created"
	AuditActionTelemetryIngested    = "telemetry.ingested"
	AuditActorLocalAPI              = "local_backend"
)

const (
	DataExportFormatJSON = "json"
	DataExportFormatPDF  = "pdf"
)

const (
	DataExportScopeTenant = "tenant"
	DataExportScopeDevice = "device"
)

const (
	DataExportStatusReady = "ready"
)

const (
	DeleteRequestScopeTenant = "tenant"
	DeleteRequestScopeDevice = "device"
)

const (
	DeleteRequestStatusQueued = "queued"
)

const (
	PolicyAssignmentTargetTenant      = "tenant"
	PolicyAssignmentTargetDeviceGroup = "device_group"
	PolicyAssignmentTargetDevice      = "device"
)

const (
	PolicyAssignmentModeAudit  = "audit"
	PolicyAssignmentModeActive = "active"
)

const (
	PolicyAssignmentStatusActive = "active"
)

const (
	ArchiveProviderS3 = "s3"
)

const (
	PlatformWindows = "windows"
	PlatformDarwin  = "darwin"
	PlatformLinux   = "linux"
)

const (
	ServiceManagerTaskScheduler = "task_scheduler"
	ServiceManagerLaunchd       = "launchd"
	ServiceManagerSystemd       = "systemd"
)

const (
	WindowsTaskTemplatePath = "deployments/service/windows/tracedeck-agent-task.xml.tmpl"
	WindowsTaskOutputPath   = "data/local/service-manifests/phase66/windows/tracedeck-agent-task.xml"
	DarwinLaunchdTemplate   = "deployments/service/darwin/io.tracedeck.agent.plist.tmpl"
	DarwinLaunchdOutput     = "data/local/service-manifests/phase66/darwin/io.tracedeck.agent.plist"
	LinuxSystemdTemplate    = "deployments/service/linux/tracedeck-agent.service.tmpl"
	LinuxSystemdOutput      = "data/local/service-manifests/phase66/linux/tracedeck-agent.service"
)

const (
	StatusOK        = "ok"
	StatusEmpty     = "empty"
	StatusPending   = "pending"
	StatusHealthy   = "healthy"
	StatusWatch     = "watch"
	StatusAttention = "attention"
)

const (
	RiskTypePolicyViolation = "policy_violation"
	RiskTypeAnomaly         = "anomaly"
	RiskTypeTamper          = "tamper"
)

const (
	SeverityInfo     = "info"
	SeverityLow      = "low"
	SeverityMedium   = "medium"
	SeverityHigh     = "high"
	SeverityCritical = "critical"
)

const (
	RiskLevelLow    = "low"
	RiskLevelMedium = "medium"
	RiskLevelHigh   = "high"
)

const (
	RiskScoreNone            = 0
	RiskScoreInfo            = 10
	RiskScoreLow             = 25
	RiskScoreMedium          = 50
	RiskScoreHigh            = 75
	RiskScoreCritical        = 95
	RiskScoreMediumThreshold = 40
	RiskScoreHighThreshold   = 75
	RiskScoreCountPenalty    = 5
	RiskScoreMaximum         = 100
	ComplianceScoreClean     = 100
)

const (
	SummaryMetricNone          = 0
	DataCompletenessUnknownPct = 0
)

const (
	HealthStatusHealthy   = "healthy"
	HealthStatusWatch     = "watch"
	HealthStatusAttention = "attention"
)

const (
	RiskStatusOpen         = "open"
	RiskStatusAcknowledged = "acknowledged"
	RiskStatusResolved     = "resolved"
)

const (
	RiskSourceBrowser = "browser"
	RiskSourceProcess = "process"
	RiskSourceAgent   = "agent"
	RiskSourceArchive = "archive"
)

const (
	RiskCategoryMediaPlayback     = "media_playback"
	RiskCategoryNonStudyYouTube   = "non_study_youtube"
	RiskCategoryEntertainment     = "entertainment"
	RiskCategoryRiskySoftware     = "risky_software"
	RiskCategoryArchiveHealth     = "archive_health"
	RiskCategoryAgentHealth       = "agent_health"
	RiskCategoryPolicyChange      = "policy_change"
	RiskCategoryProductivityShift = "productivity_shift"
)

const (
	DeliveryChannelEmail     = "email"
	DeliveryChannelPush      = "push"
	DeliveryChannelDashboard = "dashboard"
)

const (
	DeliveryStatusDelivered  = "delivered"
	DeliveryStatusPending    = "pending"
	DeliveryStatusRetrying   = "retrying"
	DeliveryStatusFailed     = "failed"
	DeliveryStatusSuppressed = "suppressed"
)

const (
	DeliveryProviderSMTP      = "smtp"
	DeliveryProviderWebPush   = "web_push"
	DeliveryProviderLocalFeed = "local_dashboard"
)

const (
	DeliveryAssuranceProviderConfirmed = "provider_confirmed"
	DeliveryAssuranceDryRunRehearsed   = "dry_run_rehearsed"
	DeliveryAssuranceDashboardVisible  = "dashboard_visible"
	DeliveryAssuranceDemoOnly          = "demo_only"
	DeliveryAssuranceRetrying          = "retrying"
	DeliveryAssuranceFailed            = "failed"
	DeliveryAssuranceRouteDisabled     = "route_disabled"
	DeliveryAssurancePendingProvider   = "pending_provider"
)

const (
	DeliveryDrillModeDryRun  = "dry_run"
	DeliveryDrillPrivacyNote = "metadata-only dry run: no provider secrets, alert bodies, screenshots, tokens, cookies, passwords, or endpoint payloads are collected or stored"
)

const (
	ProviderSimulationModeDryRun  = "dry_run"
	ProviderSimulationPrivacyNote = "metadata-only provider simulation: route labels, channel, provider type, delivery status, SLA result, retry posture, and buyer value only; no provider secrets, SMTP passwords, push endpoint payloads, alert bodies, screenshots, raw URLs, tokens, cookies, private content, or endpoint payloads are collected or stored"
	ProviderSetupPrivacyNote      = "metadata-only notification provider setup: channel labels, provider labels, recipient labels, route status, proof state, checklist state, owner action, and setup readiness only; no provider secrets, SMTP passwords, push endpoints, raw provider payloads, alert bodies, screenshots, raw URLs, page titles, tokens, cookies, private content, endpoint payloads, or passwords are collected or stored"
)

const (
	DeliveryProofStateCustomer      = "customer_proof"
	DeliveryProofStateRehearsed     = "rehearsed"
	DeliveryProofStateDisabled      = "disabled"
	DeliveryProofStateMismatch      = "provider_mismatch"
	DeliveryProofStateNeedsProvider = "needs_provider_attention"
	DeliveryProofStateNeedsProof    = "needs_delivery_proof"
)

const (
	DeliveryRemediationModeDryRun      = "dry_run"
	DeliveryRemediationActionRetryPlan = "retry_plan"
	DeliveryRemediationActionOwnerAck  = "owner_ack"
	DeliveryRemediationActionSLAWatch  = "sla_watch"
	DeliveryRemediationActionEnable    = "enable_route"
	DeliveryRemediationActionFix       = "fix_provider"
	DeliveryRemediationActionRehearsal = "run_rehearsal"
	DeliveryRemediationActionMaintain  = "maintain_proof"
	DeliveryRemediationStatusOpen      = "open"
	DeliveryRemediationStatusPlanned   = "planned"
	DeliveryRemediationStatusAcked     = "owner_acknowledged"
	DeliveryRemediationStatusHealthy   = "healthy"
	DeliveryRemediationPrivacyNote     = "metadata-only remediation: plans and audit proof are recorded without live provider sends, alert bodies, screenshots, tokens, cookies, passwords, raw URLs, or provider secrets"
)

const (
	NotificationCommandPrivacyNote    = "metadata-only notification command center: no passwords, credentials, screenshots, raw URLs, page titles, tokens, cookies, private content, alert bodies, or provider secrets are collected or stored"
	NotificationPreferencePrivacyNote = "metadata-only notification preferences: recipient labels, channel policy, quiet hours, report cadence, and escalation metadata only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, or private content are collected or stored"
	OnboardingCenterPrivacyNote       = "metadata-only onboarding center: setup step labels, host counts, service/autostart readiness labels, role readiness, notification proof, archive posture, package labels, owner actions, and privacy guardrails only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, or raw provider payloads are collected or stored"
	CustomerSettingsPrivacyNote       = "metadata-only customer settings center: plan labels, retention tier labels, notification preference status, route proof state, role settings, archive posture, autostart readiness, data-rights state, and owner actions only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, or raw provider payloads are collected or stored"
	RevenueOperationsPrivacyNote      = "metadata-only revenue operations center: tenant labels, host counts, anomaly categories, notification delivery proof, mail/push/dashboard route state, weekly report readiness, archive backlog, setup readiness, package labels, commercial levers, owner actions, and trust guardrails only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, or raw provider payloads are collected or stored"
	DeploymentReadinessPrivacyNote    = "metadata-only deployment readiness center: tenant labels, host counts, platform names, service manager labels, manifest paths, dry-run command labels, autostart status, live boot proof labels, offline replay/backlog counts, owner actions, and setup evidence only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, raw provider payloads, or hidden collection bypasses are collected or stored"
	RuntimeStatusPrivacyNote          = "metadata-only runtime status center: backend health, Scheduler readback, runtime doctor status, frontend URL presence, git diff hygiene, log file paths, operator actions, and local summary timestamps only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, payment data, raw provider payloads, keylogging, or hidden collection bypasses are collected or stored"
	PremiumOperationsPrivacyNote      = "metadata-only premium operations hub: product scores, host counts, anomaly categories, mail/push/dashboard delivery proof, weekly report readiness, archive backlog, deployment readiness, package labels, owner actions, and buyer-facing trust proof only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, raw provider payloads, keylogging, or hidden collection bypasses are collected or stored"
	BusinessDashboardPrivacyNote      = "metadata-only business dashboard: customer health, notification route proof, anomaly categories, paid-plan packaging, archive/report readiness, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	DeliveryTimelinePrivacyNote       = "metadata-only delivery timeline: channel, provider label, recipient label, status, retry timing, host label, event id, and safe summary only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	DeliveryAssurancePrivacyNote      = "metadata-only delivery assurance: channel, provider label, recipient label, source kind, evidence scope, route status, retry state, and operator truth labels only; no provider secrets, push endpoints, SMTP passwords, alert bodies, screenshots, raw URLs, page titles, tokens, cookies, private content, endpoint payloads, or raw provider payloads are collected or stored"
	RoleExperiencePrivacyNote         = "metadata-only role experience center: role labels, dashboard scope, onboarding status, notification proof, archive/report readiness, consent controls, and paid-tier packaging only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	CustomerControlPrivacyNote        = "metadata-only customer control room: host labels, anomaly categories, notification route status, mail and push proof, report readiness, archive posture, package fit, provider simulation state, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, endpoint payloads, or payment card data are collected or stored"
	CustomerSuccessPacketPrivacyNote  = "metadata-only customer success packet: customer-ready scores, anomaly categories, host labels, mail and push delivery proof, report/archive readiness, package fit, provider simulation state, privacy assurances, objections, and next actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, endpoint payloads, invoices, or payment card data are collected or stored"
	PushActivationPrivacyNote         = "metadata-only push activation center: push route labels, subscription labels, proof state, retry posture, simulation readiness, preference coverage, escalation metadata, and owner actions only; no push endpoints, provider secrets, alert bodies, screenshots, raw URLs, tokens, cookies, no passwords, private content, endpoint payloads, invoices, or payment card data are collected or stored"
	PortfolioCenterPrivacyNote        = "metadata-only portfolio center: host labels, profiles, OS labels, health scores, risk counts, notification status, archive backlog, sync posture, paid-tier labels, and owner actions only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, or raw provider payloads are collected or stored"
	AccountPortfolioPrivacyNote       = "metadata-only account portfolio index: tenant labels, plan labels, host counts, score summaries, alert counts, notification proof, archive backlog, package readiness, and owner actions only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, or raw provider payloads are collected or stored"
	ExecutiveConsolePrivacyNote       = "metadata-only executive console: product readiness, anomaly categories, host labels, email/push/dashboard delivery proof, weekly report readiness, archive status, role packaging, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	NotificationRevenuePrivacyNote    = "metadata-only notification revenue cockpit: anomaly SLA categories, mail/push/dashboard delivery proof, buyer demo readiness, escalation state, paid-package levers, weekly report readiness, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	PackageBillingPrivacyNote         = "metadata-only package billing readiness: plan labels, feature gates, seat counts, retention tier, billing setup status, report/archive value, notification proof, and upgrade actions only; no payment card data, no invoices, no provider secrets, no passwords, no screenshots, no raw URLs, no page titles, no alert bodies, no tokens, no cookies, no private content, and no endpoint payloads are collected or stored"
)

const (
	RuntimeStatusSourcePhase97Summary = "phase97_runtime_summary"
	RuntimeStatusProofBackendID       = "backend-runtime"
	RuntimeStatusProofSchedulerID     = "scheduler-readback"
	RuntimeStatusProofDoctorID        = "runtime-doctor"
	RuntimeStatusProofFrontendID      = "frontend-url"
	RuntimeStatusProofGitID           = "git-hygiene"
	RuntimeStatusProofPrivacyID       = "privacy-boundary"
	RuntimeStatusActionSummaryID      = "generate-runtime-summary"
	RuntimeStatusActionRestartID      = "restart-backend-task"
	RuntimeStatusActionReviewID       = "review-runtime-summary"
)

const (
	DeploymentAdvisoryLiveBootID       = "live-boot-proof"
	DeploymentAdvisoryAutostartID      = "native-autostart"
	DeploymentAdvisorySilentStartID    = "background-start"
	DeploymentAdvisoryOfflineReplayID  = "offline-replay"
	DeploymentAdvisoryArchiveBacklogID = "archive-backlog"
	DeploymentAdvisoryReadyID          = "deployment-ready"

	DeploymentAdvisoryLiveBootCode       = "live_boot_proof_missing"
	DeploymentAdvisoryAutostartCode      = "native_service_autostart_pending"
	DeploymentAdvisorySilentStartCode    = "background_start_not_fully_verified"
	DeploymentAdvisoryOfflineReplayCode  = "offline_replay_not_verified"
	DeploymentAdvisoryArchiveBacklogCode = "archive_backlog_waiting"
	DeploymentAdvisoryReadyCode          = "deployment_ready"
)

const (
	NotificationPreferenceModeImmediate = "immediate"
	NotificationPreferenceModeDigest    = "digest"
	NotificationPreferenceModeSilent    = "silent"
)

const (
	NotificationDigestCadenceDaily  = "daily"
	NotificationDigestCadenceWeekly = "weekly"
)

const (
	AlertRuleTemplateNonStudyYouTube = "non_study_youtube_over_limit"
	AlertRuleTemplateMediaAfterHours = "media_after_hours"
	AlertRuleTemplateRiskySoftware   = "risky_software_detected"
	AlertRuleTemplateTamperBacklog   = "archive_backlog_over_limit"
)

const (
	AlertTriggerNonStudyYouTube = "non_study_youtube"
	AlertTriggerMediaPlayback   = "media_playback"
	AlertTriggerRiskySoftware   = "risky_software"
	AlertTriggerArchiveBacklog  = "archive_backlog"
)

const (
	AlertConditionSubjectDomain         = "domain"
	AlertConditionSubjectApp            = "app"
	AlertConditionSubjectCategory       = "category"
	AlertConditionSubjectArchiveBacklog = "archive_backlog"
	AlertConditionSubjectUsageMinutes   = "usage_minutes"
)

const (
	AlertConditionOperatorEquals      = "equals"
	AlertConditionOperatorContains    = "contains"
	AlertConditionOperatorGreaterThan = "greater_than"
	AlertConditionOperatorAfterLocal  = "after_local_time"
)

const (
	ConsentStatusCollected    = "collected"
	ConsentStatusDerived      = "derived"
	ConsentStatusDenied       = "denied"
	ConsentStatusNotCollected = "not_collected"
)

const (
	ConsentCollectionAppUsage       = "Application usage metadata"
	ConsentCollectionBrowserDomains = "Browser domain and category activity"
	ConsentCollectionDeviceHealth   = "Device health score"
	ConsentCollectionArchiveHealth  = "Archive and upload health"
	ConsentCollectionPasswords      = "Passwords and credentials"
	ConsentCollectionScreenshots    = "Screenshots"
	ConsentCollectionPrivateContent = "Private messages, cookies, tokens, camera, and microphone"
)

const (
	MonetizationStageProofGap        = "proof_gap"
	MonetizationStagePilotReady      = "pilot_ready"
	MonetizationStageConversionReady = "conversion_ready"
	MonetizationStageExpansionReady  = "expansion_ready"
)

const (
	TelemetryIngestMaxEvents    = 500
	TelemetryStatusRecentEvents = 8
	TelemetryPrivacyBoundary    = "metadata-only: no passwords, credentials, screenshots, tokens, cookies, keylogs, private messages, raw URLs, or page titles"
)

const (
	ActivityFeedDefaultLimit  = 12
	ActivityFeedMaxLimit      = 50
	ActivityFeedKindRisk      = "risk"
	ActivityFeedKindDelivery  = "delivery"
	ActivityFeedKindTelemetry = "telemetry"
)

const (
	BrowserActivityDefaultLimit = 25
	BrowserActivityMaxLimit     = 100
	BrowserActivityPrivacyNote  = "metadata-only browser activity viewer: browser name, host label, domain, category, study-safe flag, visit count, notification proof, and timestamps only; no passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, or hidden collection bypasses are collected or stored"
)

const (
	BrowserNameChrome  = "chrome"
	BrowserNameEdge    = "edge"
	BrowserNameBrave   = "brave"
	BrowserNameUnknown = "unknown-browser"
)

const (
	BrowserCategoryStudy          = "study"
	BrowserCategoryVideoStreaming = "video-streaming"
	BrowserCategorySocialMedia    = "social-media"
	BrowserCategoryGaming         = "gaming"
	BrowserCategoryShopping       = "shopping"
	BrowserCategoryBlocked        = "blocked"
	BrowserCategoryUnknown        = "unknown"
)

const (
	TelemetryTypeBrowserDomainObserved = "browser.domain.observed"
	TelemetryTypeAgentHeartbeat        = "agent.health.heartbeat"
	TelemetrySourceBrowserHistory      = "collector.browser.history"
	TelemetrySourceAgentHeartbeat      = "collector.agent.heartbeat"
)

const (
	BrowserMetadataBrowserName       = "browser_name"
	BrowserMetadataCategory          = "category"
	BrowserMetadataDomain            = "domain"
	BrowserMetadataEvidenceDetail    = "evidence_detail"
	BrowserMetadataEvidenceScope     = "evidence_scope"
	BrowserMetadataSourceKind        = "source_kind"
	BrowserMetadataStoredURLMode     = "stored_url_mode"
	BrowserMetadataURLMode           = "url_mode"
	BrowserMetadataVisitCount        = "visit_count"
	BrowserMetadataYouTubeStudyMatch = "youtube_study_match"
)

const (
	EvidenceSourceDemoSeed   = "demo_seed"
	EvidenceSourceDryRun     = "dry_run"
	EvidenceSourceLiveIngest = "live_ingested"
	EvidenceSourceS3Sample   = "s3_sample"
	EvidenceSourceS3Archive  = "s3_archive"
)

const (
	EvidenceScopeDemo          = "demo"
	EvidenceScopeLive          = "live"
	EvidenceScopeMetadataOnly  = "metadata_only"
	EvidenceScopeDeliveryProof = "delivery_proof"
)

const (
	EvidenceDetailDemoRisk      = "Seeded local dashboard demo signal; use live telemetry ingest for production evidence."
	EvidenceDetailDemoDelivery  = "Seeded dashboard delivery proof only; no SMTP, SES, or web-push provider send was attempted."
	EvidenceDetailDemoBrowser   = "Seeded browser activity demo row; live rows come from browser telemetry ingest or S3 archive sampling."
	EvidenceDetailLiveTelemetry = "Accepted through local backend telemetry ingest from the endpoint agent."
	EvidenceDetailS3Sample      = "Sampled from S3 archive metadata for cloud admin rendering."
)

const (
	DemoRiskMediaAppName       = "VLC media player"
	DemoRiskMediaResourceLabel = "sample-movie-file.mkv"
	DemoRiskMediaReason        = "Entertainment media playback during study policy hours."
)

const (
	ActivityViewHighRiskOpen = "high-risk-open"
	ActivityViewEmailProof   = "email-proof"
	ActivityViewPushRetry    = "push-retry"
	ActivityViewSyncProof    = "sync-proof"
)

const (
	PortfolioProofAlertInbox        = "alert-notifications"
	PortfolioProofMailDelivery      = "mail-delivery"
	PortfolioProofPushNotifications = "push-notifications"
	PortfolioProofDashboardFallback = "dashboard-fallback"
	PortfolioProofWeeklyArchive     = "weekly-report-archive"
	PortfolioProofHostCoverage      = "host-coverage"
)
