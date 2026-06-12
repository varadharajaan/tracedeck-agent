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
	APIPrefix            = "/api/v1"
	RouteDashboard       = "/"
	RouteHealth          = "/health"
	RouteVersion         = APIPrefix + "/version"
	RouteDevices         = APIPrefix + "/devices"
	RouteDeviceEnroll    = APIPrefix + "/devices/enroll"
	RouteTenants         = APIPrefix + "/tenants"
	RoutePlans           = APIPrefix + "/plans"
	RouteRoles           = APIPrefix + "/roles"
	RouteRetentionTiers  = APIPrefix + "/retention-tiers"
	RouteAuditEvents     = APIPrefix + "/audit-events"
	RoutePolicyTemplates = APIPrefix + "/policy-templates"
	RouteArchiveStatus   = APIPrefix + "/archive/status"
)

const (
	RouteSegmentSummary       = "summary"
	RouteSegmentDaily         = "daily"
	RouteSegmentReports       = "reports"
	RouteSegmentWeekly        = "weekly"
	RouteSegmentOverview      = "overview"
	RouteSegmentAuditEvents   = "audit-events"
	RouteSegmentPolicyEvents  = "policy-violations"
	RouteSegmentAnomalies     = "anomalies"
	RouteSegmentTamperEvents  = "tamper-events"
	RouteSegmentAlertDelivery = "alert-deliveries"
	RouteSegmentHealth        = "health"
)

const (
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html; charset=utf-8"
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
	AuditActionTenantCreated = "tenant.created"
	AuditActorLocalAPI       = "local_backend"
)

const (
	ArchiveProviderS3 = "s3"
)

const (
	StatusOK      = "ok"
	StatusEmpty   = "empty"
	StatusPending = "pending"
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
