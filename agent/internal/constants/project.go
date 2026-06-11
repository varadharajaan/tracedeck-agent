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
	DefaultProcessLimit = 256
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
