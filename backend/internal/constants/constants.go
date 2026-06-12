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
	APIPrefix               = "/api/v1"
	RouteDashboard          = "/"
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
	RouteSegmentRoleExperience   = "role-experiences"
	RouteSegmentCustomerControl  = "customer-control-room"
	RouteSegmentSuccessPacket    = "customer-success-packet"
	RouteSegmentPushActivation   = "push-activation-center"
	RouteSegmentPortfolioCenter  = "portfolio-center"
	RouteSegmentExecutiveConsole = "executive-console"
	RouteSegmentNotificationRev  = "notification-revenue-cockpit"
	RouteSegmentProviderSim      = "provider-simulation-lab"
	RouteSegmentPackageBilling   = "package-billing-readiness"
	RouteSegmentNotificationCmd  = "notification-command-center"
	RouteSegmentDeliveryTimeline = "delivery-timeline"
	RouteSegmentDeliveryDrill    = "delivery-drilldown"
	RouteSegmentDeliveryRemedy   = "delivery-remediation"
	RouteSegmentSyncHealth       = "sync-health"
	RouteSegmentActivityFeed     = "activity-feed"
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
	HeaderAPIKey   = "X-TraceDeck-API-Key"
	HeaderTenantID = "X-TraceDeck-Tenant-ID"
	HeaderActorID  = "X-TraceDeck-Actor-ID"
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
	DeliveryDrillModeDryRun  = "dry_run"
	DeliveryDrillPrivacyNote = "metadata-only dry run: no provider secrets, alert bodies, screenshots, tokens, cookies, passwords, or endpoint payloads are collected or stored"
)

const (
	ProviderSimulationModeDryRun  = "dry_run"
	ProviderSimulationPrivacyNote = "metadata-only provider simulation: route labels, channel, provider type, delivery status, SLA result, retry posture, and buyer value only; no provider secrets, SMTP passwords, push endpoint payloads, alert bodies, screenshots, raw URLs, tokens, cookies, private content, or endpoint payloads are collected or stored"
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
	BusinessDashboardPrivacyNote      = "metadata-only business dashboard: customer health, notification route proof, anomaly categories, paid-plan packaging, archive/report readiness, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	DeliveryTimelinePrivacyNote       = "metadata-only delivery timeline: channel, provider label, recipient label, status, retry timing, host label, event id, and safe summary only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	RoleExperiencePrivacyNote         = "metadata-only role experience center: role labels, dashboard scope, onboarding status, notification proof, archive/report readiness, consent controls, and paid-tier packaging only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	CustomerControlPrivacyNote        = "metadata-only customer control room: host labels, anomaly categories, notification route status, mail and push proof, report readiness, archive posture, package fit, provider simulation state, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, endpoint payloads, or payment card data are collected or stored"
	CustomerSuccessPacketPrivacyNote  = "metadata-only customer success packet: customer-ready scores, anomaly categories, host labels, mail and push delivery proof, report/archive readiness, package fit, provider simulation state, privacy assurances, objections, and next actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, endpoint payloads, invoices, or payment card data are collected or stored"
	PushActivationPrivacyNote         = "metadata-only push activation center: push route labels, subscription labels, proof state, retry posture, simulation readiness, preference coverage, escalation metadata, and owner actions only; no push endpoints, provider secrets, alert bodies, screenshots, raw URLs, tokens, cookies, no passwords, private content, endpoint payloads, invoices, or payment card data are collected or stored"
	PortfolioCenterPrivacyNote        = "metadata-only portfolio center: host labels, profiles, OS labels, health scores, risk counts, notification status, archive backlog, sync posture, paid-tier labels, and owner actions only; no passwords, no screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, tokens, cookies, private content, endpoint payloads, invoices, payment card data, or raw provider payloads are collected or stored"
	ExecutiveConsolePrivacyNote       = "metadata-only executive console: product readiness, anomaly categories, host labels, email/push/dashboard delivery proof, weekly report readiness, archive status, role packaging, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	NotificationRevenuePrivacyNote    = "metadata-only notification revenue cockpit: anomaly SLA categories, mail/push/dashboard delivery proof, buyer demo readiness, escalation state, paid-package levers, weekly report readiness, and owner actions only; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored"
	PackageBillingPrivacyNote         = "metadata-only package billing readiness: plan labels, feature gates, seat counts, retention tier, billing setup status, report/archive value, notification proof, and upgrade actions only; no payment card data, no invoices, no provider secrets, no passwords, no screenshots, no raw URLs, no page titles, no alert bodies, no tokens, no cookies, no private content, and no endpoint payloads are collected or stored"
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
