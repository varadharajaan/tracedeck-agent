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
