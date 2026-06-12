package constants

const (
	BackendName    = "tracedeck-backend"
	BackendVersion = "0.1.0-dev"
)

const (
	DefaultBackendAddr = "127.0.0.1:18080"
	DefaultLogDir      = "logs/local/backend"
	DefaultLogLevel    = "info"
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
	RouteSegmentSummary      = "summary"
	RouteSegmentDaily        = "daily"
	RouteSegmentReports      = "reports"
	RouteSegmentWeekly       = "weekly"
	RouteSegmentAuditEvents  = "audit-events"
	RouteSegmentPolicyEvents = "policy-violations"
	RouteSegmentAnomalies    = "anomalies"
	RouteSegmentTamperEvents = "tamper-events"
)

const (
	ContentTypeJSON = "application/json"
	ContentTypeHTML = "text/html; charset=utf-8"
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
	StatusOK    = "ok"
	StatusEmpty = "empty"
)
