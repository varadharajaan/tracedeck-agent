package config

type Policy struct {
	TenantID           string              `json:"tenant_id" yaml:"tenant_id" jsonschema:"minLength=1"`
	DeviceID           string              `json:"device_id" yaml:"device_id" jsonschema:"minLength=1"`
	Profile            string              `json:"profile" yaml:"profile" jsonschema:"minLength=1"`
	Collection         CollectionPolicy    `json:"collection" yaml:"collection"`
	Retention          RetentionPolicy     `json:"retention" yaml:"retention"`
	Archive            ArchivePolicy       `json:"archive" yaml:"archive"`
	BackendSync        BackendSyncPolicy   `json:"backend_sync" yaml:"backend_sync"`
	Observability      ObservabilityPolicy `json:"observability" yaml:"observability"`
	Alerts             AlertPolicy         `json:"alerts" yaml:"alerts"`
	StudyApps          []string            `json:"study_apps" yaml:"study_apps"`
	BlockedApps        []string            `json:"blocked_apps" yaml:"blocked_apps"`
	IgnoredApps        []string            `json:"ignored_apps" yaml:"ignored_apps"`
	AllowedDomains     []string            `json:"allowed_domains" yaml:"allowed_domains"`
	BlockedDomains     []string            `json:"blocked_domains" yaml:"blocked_domains"`
	AllowedCategories  []string            `json:"allowed_categories" yaml:"allowed_categories"`
	WarnCategories     []string            `json:"warn_categories" yaml:"warn_categories"`
	CriticalCategories []string            `json:"critical_categories" yaml:"critical_categories"`
	Thresholds         ThresholdPolicy     `json:"thresholds" yaml:"thresholds"`
	YouTubeKeywords    []string            `json:"youtube_study_keywords" yaml:"youtube_study_keywords"`
	AlertRules         map[string]RuleSpec `json:"alert_rules" yaml:"alert_rules"`
}

type CollectionPolicy struct {
	TransparencyMode      TransparencyMode        `json:"transparency_mode" yaml:"transparency_mode"`
	Browser               BrowserCollection       `json:"browser" yaml:"browser"`
	ForegroundApp         ForegroundAppCollection `json:"foreground_app" yaml:"foreground_app"`
	Media                 MediaCollection         `json:"media" yaml:"media"`
	SensitiveCapabilities SensitiveCapabilities   `json:"sensitive_capabilities" yaml:"sensitive_capabilities"`
}

type BrowserCollection struct {
	URLMode               URLMode     `json:"url_mode" yaml:"url_mode"`
	CollectPageTitle      bool        `json:"collect_page_title" yaml:"collect_page_title"`
	YouTubeClassification FeatureMode `json:"youtube_classification" yaml:"youtube_classification"`
	YouTubeVideoIDMode    VideoIDMode `json:"youtube_video_id_mode" yaml:"youtube_video_id_mode"`
}

type ForegroundAppCollection struct {
	Enabled         bool            `json:"enabled" yaml:"enabled"`
	WindowTitleMode WindowTitleMode `json:"window_title_mode" yaml:"window_title_mode"`
}

type MediaCollection struct {
	CollectFileName bool     `json:"collect_file_name" yaml:"collect_file_name"`
	CollectFilePath bool     `json:"collect_file_path" yaml:"collect_file_path"`
	PathMode        PathMode `json:"path_mode" yaml:"path_mode"`
}

type SensitiveCapabilities struct {
	Credentials     SensitiveCapabilityMode `json:"credentials" yaml:"credentials"`
	Keystrokes      SensitiveCapabilityMode `json:"keystrokes" yaml:"keystrokes"`
	Cookies         SensitiveCapabilityMode `json:"cookies" yaml:"cookies"`
	Tokens          SensitiveCapabilityMode `json:"tokens" yaml:"tokens"`
	PrivateMessages SensitiveCapabilityMode `json:"private_messages" yaml:"private_messages"`
	Screenshots     SensitiveCapabilityMode `json:"screenshots" yaml:"screenshots"`
}

type RetentionPolicy struct {
	LocalTTLDays      int `json:"local_ttl_days" yaml:"local_ttl_days" jsonschema:"minimum=1"`
	MaxLocalStorageMB int `json:"max_local_storage_mb" yaml:"max_local_storage_mb" jsonschema:"minimum=1"`
}

type ThresholdPolicy struct {
	MaxVideoMinutesPerDay      int    `json:"max_video_minutes_per_day" yaml:"max_video_minutes_per_day" jsonschema:"minimum=0"`
	MaxSocialMinutesPerDay     int    `json:"max_social_minutes_per_day" yaml:"max_social_minutes_per_day" jsonschema:"minimum=0"`
	MaxUnknownAppMinutesPerDay int    `json:"max_unknown_app_minutes_per_day" yaml:"max_unknown_app_minutes_per_day" jsonschema:"minimum=0"`
	LateNightUsageStart        string `json:"late_night_usage_start" yaml:"late_night_usage_start" jsonschema:"pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$"`
	LateNightUsageEnd          string `json:"late_night_usage_end" yaml:"late_night_usage_end" jsonschema:"pattern=^([01][0-9]|2[0-3]):[0-5][0-9]$"`
}

type ArchivePolicy struct {
	Enabled          bool             `json:"enabled" yaml:"enabled"`
	Provider         ArchiveProvider  `json:"provider" yaml:"provider"`
	Bucket           string           `json:"bucket" yaml:"bucket"`
	PrefixTemplate   string           `json:"prefix_template" yaml:"prefix_template"`
	UploadInterval   string           `json:"upload_interval" yaml:"upload_interval"`
	RetryWhenOnline  bool             `json:"retry_when_online" yaml:"retry_when_online"`
	StorageClassDays StorageClassDays `json:"storage_class_days" yaml:"storage_class_days"`
}

type BackendSyncPolicy struct {
	Enabled        bool   `json:"enabled" yaml:"enabled"`
	BaseURL        string `json:"base_url" yaml:"base_url" jsonschema:"format=uri"`
	BatchLimit     int    `json:"batch_limit" yaml:"batch_limit" jsonschema:"minimum=1"`
	RequestTimeout string `json:"request_timeout" yaml:"request_timeout"`
}

type ObservabilityPolicy struct {
	OpenTelemetry OpenTelemetryPolicy `json:"opentelemetry" yaml:"opentelemetry"`
}

type OpenTelemetryPolicy struct {
	Enabled        bool                     `json:"enabled" yaml:"enabled"`
	Protocol       OpenTelemetryProtocol    `json:"protocol" yaml:"protocol"`
	Endpoint       string                   `json:"endpoint" yaml:"endpoint" jsonschema:"format=uri"`
	BatchLimit     int                      `json:"batch_limit" yaml:"batch_limit" jsonschema:"minimum=1"`
	RequestTimeout string                   `json:"request_timeout" yaml:"request_timeout"`
	Retry          OpenTelemetryRetryPolicy `json:"retry" yaml:"retry"`
}

type OpenTelemetryRetryPolicy struct {
	MaxAttempts int `json:"max_attempts" yaml:"max_attempts" jsonschema:"minimum=1"`
}

type StorageClassDays struct {
	Standard        int `json:"standard" yaml:"standard" jsonschema:"minimum=0"`
	StandardIAUntil int `json:"standard_ia_until" yaml:"standard_ia_until" jsonschema:"minimum=1"`
	ArchiveAfter    int `json:"archive_after" yaml:"archive_after" jsonschema:"minimum=1"`
}

type AlertPolicy struct {
	Enabled bool        `json:"enabled" yaml:"enabled"`
	Email   EmailPolicy `json:"email" yaml:"email"`
}

type EmailPolicy struct {
	Provider        EmailProvider `json:"provider" yaml:"provider"`
	From            string        `json:"from" yaml:"from" jsonschema:"format=email"`
	To              []string      `json:"to" yaml:"to"`
	MinSeverity     Severity      `json:"min_severity" yaml:"min_severity"`
	CooldownMinutes int           `json:"cooldown_minutes" yaml:"cooldown_minutes" jsonschema:"minimum=1"`
}

type RuleSpec struct {
	Enabled                  bool     `json:"enabled" yaml:"enabled"`
	Severity                 Severity `json:"severity" yaml:"severity"`
	IncludeMediaFileMetadata bool     `json:"include_media_file_metadata,omitempty" yaml:"include_media_file_metadata,omitempty"`
	ThresholdMinutesPerDay   int      `json:"threshold_minutes_per_day,omitempty" yaml:"threshold_minutes_per_day,omitempty"`
}
