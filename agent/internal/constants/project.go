package constants

const (
	AppName       = "tracedeck-agent"
	AppVersion    = "0.1.0-dev"
	RepoModule    = "github.com/varadharajaan/tracedeck-agent"
	DefaultConfig = "./examples/policies/ai-btech-student.yaml"
	UnknownHost   = "unknown-host"
)

const (
	CommandValidateConfig = "validate-config"
	CommandSchema         = "schema"
	CommandRun            = "run"
)

const (
	DefaultLogDir       = "logs/local/agent"
	DefaultLogFileName  = "tracedeck-agent.log"
	DefaultLogLevel     = LogLevelInfo
	DefaultDataDir      = "data/local"
	DefaultSQLiteFile   = "tracedeck-agent.sqlite"
	DefaultOutboxDir    = "data/local/outbox"
	DefaultProcessLimit = 256
	DefaultBrowserLimit = DefaultBrowserHistoryLimit
	DefaultMaxCycles    = 0
)

const (
	DefaultCollectionInterval    = "10m"
	DefaultBackendSyncBatchLimit = 100
	DefaultBackendSyncTimeout    = "10s"
	BackendSyncCursorName        = "backend_telemetry"
	BackendSyncEventIDPrefix     = "local-event-"
)

const (
	LogLevelTrace = "trace"
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

const (
	LogRotationMaxSizeMB = 10
	LogRotationMaxFiles  = 5
	LogRotationMaxAgeDay = 30
)

const (
	SQLiteMigrationGlob = "migrations/*.sql"
)

const (
	ArchiveOutboxDirName = "archive"
	AlertOutboxDirName   = "alerts"
	JSONLinesGzipExt     = ".jsonl.gz"
	JSONExt              = ".json"
)

const (
	TemplateTenantID = "{tenant_id}"
	TemplateDeviceID = "{device_id}"
	TemplateHostName = "{host_name}"
	TemplateYear     = "{yyyy}"
	TemplateMonth    = "{mm}"
	TemplateDay      = "{dd}"
	TemplateHour     = "{hh}"
)
