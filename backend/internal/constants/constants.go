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
	RoutePolicyTemplates = APIPrefix + "/policy-templates"
	RouteArchiveStatus   = APIPrefix + "/archive/status"
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
	StatusOK    = "ok"
	StatusEmpty = "empty"
)
