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
	RouteSegmentSummary         = "summary"
	RouteSegmentDaily           = "daily"
	RouteSegmentReports         = "reports"
	RouteSegmentWeekly          = "weekly"
	RouteSegmentPDF             = "pdf"
	RouteSegmentOverview        = "overview"
	RouteSegmentAuditEvents     = "audit-events"
	RouteSegmentAlertRules      = "alert-rules"
	RouteSegmentConsentCenter   = "consent-center"
	RouteSegmentAlertInbox      = "alert-inbox"
	RouteSegmentOperations      = "operations-summary"
	RouteSegmentMonetization    = "monetization-summary"
	RouteSegmentNotificationCmd = "notification-command-center"
	RouteSegmentDeliveryDrill   = "delivery-drilldown"
	RouteSegmentDeliveryRemedy  = "delivery-remediation"
	RouteSegmentSyncHealth      = "sync-health"
	RouteSegmentActivityFeed    = "activity-feed"
	RouteSegmentActivityViews   = "activity-views"
	RouteSegmentNotifications   = "notification-routes"
	RouteSegmentDataExports     = "data-exports"
	RouteSegmentDeleteRequests  = "delete-requests"
	RouteSegmentDeviceGroups    = "device-groups"
	RouteSegmentPolicyAssign    = "policy-assignments"
	RouteSegmentPolicyEvents    = "policy-violations"
	RouteSegmentAnomalies       = "anomalies"
	RouteSegmentTamperEvents    = "tamper-events"
	RouteSegmentAlertDelivery   = "alert-deliveries"
	RouteSegmentHealth          = "health"
	RouteSegmentTelemetry       = "telemetry-events"
	RouteSegmentTelemetryStatus = "telemetry-status"
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
	AuditActionDeliveryDrillRun     = "delivery_drilldown.rehearsed"
	AuditActionDeliveryRemediation  = "delivery_remediation.planned"
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
	NotificationCommandPrivacyNote = "metadata-only notification command center: no passwords, credentials, screenshots, raw URLs, page titles, tokens, cookies, private content, alert bodies, or provider secrets are collected or stored"
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
