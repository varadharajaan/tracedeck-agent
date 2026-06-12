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

type WeeklyReport struct {
	DeviceID      string   `json:"device_id"`
	Week          string   `json:"week"`
	Highlights    []string `json:"highlights"`
	Risks         []string `json:"risks"`
	Generated     bool     `json:"generated"`
	GeneratedNote string   `json:"generated_note"`
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
