package model

import "time"

type Health struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	StartedAt time.Time `json:"started_at"`
}

type Version struct {
	Service string `json:"service"`
	Version string `json:"version"`
}

type Device struct {
	TenantID   string    `json:"tenant_id"`
	DeviceID   string    `json:"device_id"`
	HostName   string    `json:"host_name"`
	Profile    string    `json:"profile"`
	OSName     string    `json:"os_name"`
	EnrolledAt time.Time `json:"enrolled_at"`
	LastSeenAt time.Time `json:"last_seen_at"`
}

type Tenant struct {
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	PlanID          string    `json:"plan_id"`
	RetentionTierID string    `json:"retention_tier_id"`
	PrimaryProfile  string    `json:"primary_profile"`
	DeviceLimit     int       `json:"device_limit"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateTenantRequest struct {
	TenantID        string `json:"tenant_id"`
	Name            string `json:"name"`
	PlanID          string `json:"plan_id"`
	RetentionTierID string `json:"retention_tier_id"`
	PrimaryProfile  string `json:"primary_profile"`
}

type DeviceGroup struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Profile          string    `json:"profile"`
	DeviceIDs        []string  `json:"device_ids"`
	PolicyTemplateID string    `json:"policy_template_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreateDeviceGroupRequest struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Profile          string   `json:"profile"`
	DeviceIDs        []string `json:"device_ids"`
	PolicyTemplateID string   `json:"policy_template_id"`
}

type PolicyAssignment struct {
	ID               string    `json:"id"`
	TenantID         string    `json:"tenant_id"`
	Name             string    `json:"name"`
	TargetType       string    `json:"target_type"`
	TargetID         string    `json:"target_id"`
	PolicyTemplateID string    `json:"policy_template_id"`
	AlertRuleIDs     []string  `json:"alert_rule_ids"`
	Mode             string    `json:"mode"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type CreatePolicyAssignmentRequest struct {
	Name             string   `json:"name"`
	TargetType       string   `json:"target_type"`
	TargetID         string   `json:"target_id"`
	PolicyTemplateID string   `json:"policy_template_id"`
	AlertRuleIDs     []string `json:"alert_rule_ids"`
	Mode             string   `json:"mode"`
}

type TenantDataExport struct {
	ID            string     `json:"id"`
	TenantID      string     `json:"tenant_id"`
	Format        string     `json:"format"`
	Scope         string     `json:"scope"`
	Status        string     `json:"status"`
	ResourceCount int        `json:"resource_count"`
	StorageKey    string     `json:"storage_key"`
	RequestedBy   string     `json:"requested_by"`
	CreatedAt     time.Time  `json:"created_at"`
	CompletedAt   time.Time  `json:"completed_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
}

type CreateTenantDataExportRequest struct {
	Format string `json:"format"`
	Scope  string `json:"scope"`
}

type DeleteRequest struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Scope       string    `json:"scope"`
	Reason      string    `json:"reason"`
	Status      string    `json:"status"`
	RequestedBy string    `json:"requested_by"`
	DueAt       time.Time `json:"due_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateDeleteRequestRequest struct {
	Scope  string `json:"scope"`
	Reason string `json:"reason"`
}

type EnrollDeviceRequest struct {
	TenantID string `json:"tenant_id"`
	DeviceID string `json:"device_id"`
	HostName string `json:"host_name"`
	Profile  string `json:"profile"`
	OSName   string `json:"os_name"`
}

type DeviceSummary struct {
	DeviceID            string `json:"device_id"`
	Date                string `json:"date"`
	StudyMinutes        int    `json:"study_minutes"`
	CodingMinutes       int    `json:"coding_minutes"`
	EntertainmentMins   int    `json:"entertainment_minutes"`
	PolicyViolations    int    `json:"policy_violations"`
	ComplianceScore     int    `json:"compliance_score"`
	ArchiveBacklog      int    `json:"archive_backlog"`
	AlertsRaised        int    `json:"alerts_raised"`
	DataCompletenessPct int    `json:"data_completeness_pct"`
}

type HostOverview struct {
	Device           Device          `json:"device"`
	Summary          DeviceSummary   `json:"summary"`
	RiskLevel        string          `json:"risk_level"`
	RiskScore        int             `json:"risk_score"`
	Health           DeviceHealth    `json:"health"`
	Archive          ArchiveStatus   `json:"archive"`
	PolicyViolations []RiskEvent     `json:"policy_violations"`
	Anomalies        []RiskEvent     `json:"anomalies"`
	TamperEvents     []RiskEvent     `json:"tamper_events"`
	AlertDeliveries  []AlertDelivery `json:"alert_deliveries"`
	GeneratedAt      time.Time       `json:"generated_at"`
}

type RiskEvent struct {
	ID             string    `json:"id"`
	DeviceID       string    `json:"device_id"`
	Type           string    `json:"type"`
	Severity       string    `json:"severity"`
	Category       string    `json:"category"`
	Source         string    `json:"source"`
	AppName        string    `json:"app_name"`
	Domain         string    `json:"domain"`
	ResourceLabel  string    `json:"resource_label"`
	Reason         string    `json:"reason"`
	Recommendation string    `json:"recommendation"`
	Status         string    `json:"status"`
	ObservedAt     time.Time `json:"observed_at"`
}

type AlertDelivery struct {
	ID               string     `json:"id"`
	DeviceID         string     `json:"device_id"`
	EventID          string     `json:"event_id"`
	Channel          string     `json:"channel"`
	Recipient        string     `json:"recipient"`
	Provider         string     `json:"provider"`
	Status           string     `json:"status"`
	Attempts         int        `json:"attempts"`
	LastAttemptAt    time.Time  `json:"last_attempt_at"`
	NextRetryAt      *time.Time `json:"next_retry_at,omitempty"`
	LastError        string     `json:"last_error,omitempty"`
	SuppressedReason string     `json:"suppressed_reason,omitempty"`
	Summary          string     `json:"summary"`
}

type TelemetryEvent struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Source     string            `json:"source"`
	ObservedAt time.Time         `json:"observed_at"`
	TenantID   string            `json:"tenant_id"`
	DeviceID   string            `json:"device_id"`
	HostName   string            `json:"host_name"`
	AppName    string            `json:"app_name"`
	ProcessID  int32             `json:"process_id"`
	PathHash   string            `json:"path_hash"`
	Metadata   map[string]string `json:"metadata"`
}

type IngestTelemetryRequest struct {
	TenantID string           `json:"tenant_id"`
	DeviceID string           `json:"device_id"`
	HostName string           `json:"host_name"`
	Profile  string           `json:"profile"`
	OSName   string           `json:"os_name"`
	Events   []TelemetryEvent `json:"events"`
}

type IngestTelemetryResponse struct {
	TenantID           string    `json:"tenant_id"`
	DeviceID           string    `json:"device_id"`
	AcceptedEvents     int       `json:"accepted_events"`
	StoredEvents       int       `json:"stored_events"`
	LastObservedAt     time.Time `json:"last_observed_at"`
	LastIngestedAt     time.Time `json:"last_ingested_at"`
	PrivacyBoundary    string    `json:"privacy_boundary"`
	BackendVisibleHost bool      `json:"backend_visible_host"`
}

type TelemetryIngestStatus struct {
	TenantID        string           `json:"tenant_id"`
	DeviceID        string           `json:"device_id"`
	HostName        string           `json:"host_name"`
	StoredEvents    int              `json:"stored_events"`
	CountsByType    map[string]int   `json:"counts_by_type"`
	CountsBySource  map[string]int   `json:"counts_by_source"`
	LastObservedAt  time.Time        `json:"last_observed_at"`
	LastIngestedAt  time.Time        `json:"last_ingested_at"`
	RecentEvents    []TelemetryEvent `json:"recent_events"`
	PrivacyBoundary string           `json:"privacy_boundary"`
}

type TenantActivityFeedFilter struct {
	DeviceID string `json:"device_id"`
	Kind     string `json:"kind"`
	Severity string `json:"severity"`
	Channel  string `json:"channel"`
	Status   string `json:"status"`
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
}

type TenantActivityFeedSummary struct {
	Total           int `json:"total"`
	RiskItems       int `json:"risk_items"`
	DeliveryItems   int `json:"delivery_items"`
	TelemetryItems  int `json:"telemetry_items"`
	HighRiskOpen    int `json:"high_risk_open"`
	EmailDelivered  int `json:"email_delivered"`
	PushNeedsRetry  int `json:"push_needs_retry"`
	ReportingHosts  int `json:"reporting_hosts"`
	SourceHostCount int `json:"source_host_count"`
}

type TenantActivityFeedItem struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	DeviceID       string    `json:"device_id"`
	HostName       string    `json:"host_name"`
	Kind           string    `json:"kind"`
	Type           string    `json:"type"`
	Severity       string    `json:"severity"`
	Category       string    `json:"category"`
	Channel        string    `json:"channel"`
	Status         string    `json:"status"`
	Title          string    `json:"title"`
	Detail         string    `json:"detail"`
	Recommendation string    `json:"recommendation"`
	Source         string    `json:"source"`
	Provider       string    `json:"provider"`
	Recipient      string    `json:"recipient"`
	EventID        string    `json:"event_id"`
	ObservedAt     time.Time `json:"observed_at"`
}

type TenantActivityFeed struct {
	TenantID        string                    `json:"tenant_id"`
	TenantName      string                    `json:"tenant_name"`
	Filters         TenantActivityFeedFilter  `json:"filters"`
	Summary         TenantActivityFeedSummary `json:"summary"`
	Items           []TenantActivityFeedItem  `json:"items"`
	GeneratedAt     time.Time                 `json:"generated_at"`
	PrivacyBoundary string                    `json:"privacy_boundary"`
}

type DeviceSyncHealth struct {
	TenantID          string    `json:"tenant_id"`
	DeviceID          string    `json:"device_id"`
	HostName          string    `json:"host_name"`
	Status            string    `json:"status"`
	StoredEvents      int       `json:"stored_events"`
	LastLocalEventID  int64     `json:"last_local_event_id"`
	LastObservedAt    time.Time `json:"last_observed_at"`
	LastIngestedAt    time.Time `json:"last_ingested_at"`
	ProcessEvents     int       `json:"process_events"`
	HealthEvents      int       `json:"health_events"`
	BrowserEvents     int       `json:"browser_events"`
	RecentEventIDs    []string  `json:"recent_event_ids"`
	Recommendation    string    `json:"recommendation"`
	PrivacyBoundary   string    `json:"privacy_boundary"`
	BackendVisible    bool      `json:"backend_visible"`
	OfflineReplayHint string    `json:"offline_replay_hint"`
}

type TenantSyncHealth struct {
	TenantID             string             `json:"tenant_id"`
	TenantName           string             `json:"tenant_name"`
	Status               string             `json:"status"`
	HostsTotal           int                `json:"hosts_total"`
	HostsReporting       int                `json:"hosts_reporting"`
	HostsPending         int                `json:"hosts_pending"`
	StoredEvents         int                `json:"stored_events"`
	LastLocalEventID     int64              `json:"last_local_event_id"`
	LastIngestedAt       time.Time          `json:"last_ingested_at"`
	BackendVisible       bool               `json:"backend_visible"`
	PrivacyBoundary      string             `json:"privacy_boundary"`
	OfflineReplayReady   bool               `json:"offline_replay_ready"`
	OfflineReplaySummary string             `json:"offline_replay_summary"`
	Devices              []DeviceSyncHealth `json:"devices"`
	GeneratedAt          time.Time          `json:"generated_at"`
}

type DeviceHealth struct {
	DeviceID             string    `json:"device_id"`
	Score                int       `json:"score"`
	Status               string    `json:"status"`
	CPUPercent           float64   `json:"cpu_percent"`
	MemoryPercent        float64   `json:"memory_percent"`
	DiskPercent          float64   `json:"disk_percent"`
	BatteryStatus        string    `json:"battery_status"`
	BatteryPercent       int       `json:"battery_percent"`
	StartupApps          int       `json:"startup_apps"`
	AppCrashes24h        int       `json:"app_crashes_24h"`
	AgentHealthy         bool      `json:"agent_healthy"`
	AgentLastHeartbeatAt time.Time `json:"agent_last_heartbeat_at"`
	ObservedAt           time.Time `json:"observed_at"`
	Recommendation       string    `json:"recommendation"`
}

type WeeklyReport struct {
	DeviceID      string    `json:"device_id"`
	Week          string    `json:"week"`
	Summary       string    `json:"summary"`
	Highlights    []string  `json:"highlights"`
	Risks         []string  `json:"risks"`
	Generated     bool      `json:"generated"`
	GeneratedNote string    `json:"generated_note"`
	EmailReady    bool      `json:"email_ready"`
	EmailSubject  string    `json:"email_subject"`
	EmailPreview  string    `json:"email_preview"`
	PDFReady      bool      `json:"pdf_ready"`
	GeneratedAt   time.Time `json:"generated_at"`
}

type PolicyTemplate struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Audience    string   `json:"audience"`
	Description string   `json:"description"`
	Roles       []string `json:"roles"`
}

type AlertRuleTemplate struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Trigger         string   `json:"trigger"`
	Description     string   `json:"description"`
	DefaultSeverity string   `json:"default_severity"`
	Channels        []string `json:"channels"`
	Example         string   `json:"example"`
	PaidTier        string   `json:"paid_tier"`
}

type AlertRuleCondition struct {
	Subject       string `json:"subject"`
	Operator      string `json:"operator"`
	Value         string `json:"value"`
	WindowMinutes int    `json:"window_minutes"`
	Threshold     int    `json:"threshold"`
}

type AlertRule struct {
	ID         string             `json:"id"`
	TenantID   string             `json:"tenant_id"`
	TemplateID string             `json:"template_id"`
	Name       string             `json:"name"`
	Trigger    string             `json:"trigger"`
	Severity   string             `json:"severity"`
	Channels   []string           `json:"channels"`
	Condition  AlertRuleCondition `json:"condition"`
	Enabled    bool               `json:"enabled"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type CreateAlertRuleRequest struct {
	TemplateID string             `json:"template_id"`
	Name       string             `json:"name"`
	Trigger    string             `json:"trigger"`
	Severity   string             `json:"severity"`
	Channels   []string           `json:"channels"`
	Condition  AlertRuleCondition `json:"condition"`
	Enabled    bool               `json:"enabled"`
}

type NotificationRoute struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	Channel        string     `json:"channel"`
	Provider       string     `json:"provider"`
	RecipientLabel string     `json:"recipient_label"`
	Status         string     `json:"status"`
	Enabled        bool       `json:"enabled"`
	LastVerifiedAt *time.Time `json:"last_verified_at,omitempty"`
	LastSummary    string     `json:"last_summary"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateNotificationRouteRequest struct {
	Channel        string `json:"channel"`
	Provider       string `json:"provider"`
	RecipientLabel string `json:"recipient_label"`
	Status         string `json:"status"`
	Enabled        bool   `json:"enabled"`
	LastSummary    string `json:"last_summary"`
}

type ConsentCollectionItem struct {
	Name        string `json:"name"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Retention   string `json:"retention"`
}

type ConsentCenter struct {
	TenantID           string                  `json:"tenant_id"`
	MonitoringVisible  bool                    `json:"monitoring_visible"`
	PauseControls      string                  `json:"pause_controls"`
	DataExportReady    bool                    `json:"data_export_ready"`
	DeleteRequestReady bool                    `json:"delete_request_ready"`
	AlertRecipients    []string                `json:"alert_recipients"`
	Collection         []ConsentCollectionItem `json:"collection"`
	AuditEvents        []AuditEvent            `json:"audit_events"`
	UpdatedAt          time.Time               `json:"updated_at"`
}

type TenantDeliverySnapshot struct {
	Channel       string    `json:"channel"`
	Status        string    `json:"status"`
	Recipient     string    `json:"recipient"`
	Provider      string    `json:"provider"`
	LastAttemptAt time.Time `json:"last_attempt_at"`
	Summary       string    `json:"summary"`
}

type TenantOperationsSignal struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Severity   string    `json:"severity"`
	Channel    string    `json:"channel"`
	Status     string    `json:"status"`
	Owner      string    `json:"owner"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantOperationsSummary struct {
	TenantID              string                   `json:"tenant_id"`
	TenantName            string                   `json:"tenant_name"`
	PlanID                string                   `json:"plan_id"`
	PlanName              string                   `json:"plan_name"`
	CustomerHealth        string                   `json:"customer_health"`
	MonetizationReadiness int                      `json:"monetization_readiness"`
	HostsTotal            int                      `json:"hosts_total"`
	HostsAttention        int                      `json:"hosts_attention"`
	RiskScore             int                      `json:"risk_score"`
	OpenPolicyViolations  int                      `json:"open_policy_violations"`
	OpenAnomalies         int                      `json:"open_anomalies"`
	TamperSignals         int                      `json:"tamper_signals"`
	ArchiveBacklog        int                      `json:"archive_backlog"`
	NotificationScore     int                      `json:"notification_score"`
	DeliveryTotal         int                      `json:"delivery_total"`
	DeliveryDelivered     int                      `json:"delivery_delivered"`
	DeliveryRetrying      int                      `json:"delivery_retrying"`
	DeliveryFailed        int                      `json:"delivery_failed"`
	EmailDelivered        int                      `json:"email_delivered"`
	PushDelivered         int                      `json:"push_delivered"`
	DashboardDelivered    int                      `json:"dashboard_delivered"`
	LastEmail             *TenantDeliverySnapshot  `json:"last_email,omitempty"`
	LastPush              *TenantDeliverySnapshot  `json:"last_push,omitempty"`
	PrioritySignals       []TenantOperationsSignal `json:"priority_signals"`
	UpgradeSignals        []TenantOperationsSignal `json:"upgrade_signals"`
	GeneratedAt           time.Time                `json:"generated_at"`
}

type TenantMonetizationSummary struct {
	TenantID            string                    `json:"tenant_id"`
	TenantName          string                    `json:"tenant_name"`
	PlanID              string                    `json:"plan_id"`
	PlanName            string                    `json:"plan_name"`
	Audience            string                    `json:"audience"`
	ConversionStage     string                    `json:"conversion_stage"`
	RevenueHealth       string                    `json:"revenue_health"`
	SeatsUsed           int                       `json:"seats_used"`
	SeatsIncluded       int                       `json:"seats_included"`
	ReadinessScore      int                       `json:"readiness_score"`
	NotificationScore   int                       `json:"notification_score"`
	TrustScore          int                       `json:"trust_score"`
	NotificationPromise TenantNotificationPromise `json:"notification_promise"`
	NotificationRoutes  []TenantNotificationRoute `json:"notification_routes"`
	ValuePanels         []TenantValuePanel        `json:"value_panels"`
	PaidCapabilities    []TenantPaidCapability    `json:"paid_capabilities"`
	ConversionActions   []TenantOperationsSignal  `json:"conversion_actions"`
	GeneratedAt         time.Time                 `json:"generated_at"`
}

type TenantNotificationPromise struct {
	Status    string `json:"status"`
	Summary   string `json:"summary"`
	Email     string `json:"email"`
	Push      string `json:"push"`
	Dashboard string `json:"dashboard"`
}

type TenantNotificationRoute struct {
	Channel       string     `json:"channel"`
	Provider      string     `json:"provider"`
	Status        string     `json:"status"`
	Recipient     string     `json:"recipient"`
	Attempts      int        `json:"attempts"`
	LastAttemptAt time.Time  `json:"last_attempt_at"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	Proof         string     `json:"proof"`
	NextAction    string     `json:"next_action"`
}

type TenantValuePanel struct {
	Title    string `json:"title"`
	Metric   string `json:"metric"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	PaidTier string `json:"paid_tier"`
}

type TenantPaidCapability struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Tier     string `json:"tier"`
	Evidence string `json:"evidence"`
}

type Plan struct {
	ID                 string   `json:"id"`
	Name               string   `json:"name"`
	Audience           string   `json:"audience"`
	DeviceLimit        int      `json:"device_limit"`
	CloudArchive       bool     `json:"cloud_archive"`
	WeeklyReports      bool     `json:"weekly_reports"`
	RoleBasedDashboard bool     `json:"role_based_dashboard"`
	PriceModel         string   `json:"price_model"`
	Features           []string `json:"features"`
}

type Role struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
}

type RetentionTier struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	LocalTTLDays       int    `json:"local_ttl_days"`
	S3StandardDays     int    `json:"s3_standard_days"`
	S3StandardIAUntil  int    `json:"s3_standard_ia_until_days"`
	S3ArchiveAfterDays int    `json:"s3_archive_after_days"`
	ComplianceExport   bool   `json:"compliance_export"`
	Description        string `json:"description"`
}

type AuditEvent struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Category  string    `json:"category"`
	Action    string    `json:"action"`
	Actor     string    `json:"actor"`
	ActorRole string    `json:"actor_role"`
	Summary   string    `json:"summary"`
	CreatedAt time.Time `json:"created_at"`
}

type ArchiveStatus struct {
	Status          string `json:"status"`
	Provider        string `json:"provider"`
	PendingBatches  int    `json:"pending_batches"`
	LastUploadedKey string `json:"last_uploaded_key"`
}

type ListResponse[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
