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

type TenantDeliveryTimelineFilter struct {
	DeviceID string `json:"device_id"`
	Channel  string `json:"channel"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
	Query    string `json:"query"`
	Limit    int    `json:"limit"`
}

type TenantDeliveryTimelineSummary struct {
	Total               int        `json:"total"`
	Delivered           int        `json:"delivered"`
	Retrying            int        `json:"retrying"`
	Failed              int        `json:"failed"`
	Suppressed          int        `json:"suppressed"`
	Email               int        `json:"email"`
	Push                int        `json:"push"`
	Dashboard           int        `json:"dashboard"`
	SourceHostCount     int        `json:"source_host_count"`
	RouteProofGaps      int        `json:"route_proof_gaps"`
	NotificationScore   int        `json:"notification_score"`
	NextRetryAt         *time.Time `json:"next_retry_at,omitempty"`
	LastDeliveredAt     *time.Time `json:"last_delivered_at,omitempty"`
	RecommendedPaidTier string     `json:"recommended_paid_tier"`
}

type TenantDeliveryTimelineItem struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	DeviceID         string     `json:"device_id"`
	HostName         string     `json:"host_name"`
	EventID          string     `json:"event_id"`
	Channel          string     `json:"channel"`
	Provider         string     `json:"provider"`
	Recipient        string     `json:"recipient"`
	Status           string     `json:"status"`
	Attempts         int        `json:"attempts"`
	Summary          string     `json:"summary"`
	NextAction       string     `json:"next_action"`
	PaidTier         string     `json:"paid_tier"`
	LastAttemptAt    time.Time  `json:"last_attempt_at"`
	NextRetryAt      *time.Time `json:"next_retry_at,omitempty"`
	LastError        string     `json:"last_error,omitempty"`
	SuppressedReason string     `json:"suppressed_reason,omitempty"`
}

type TenantDeliveryTimeline struct {
	TenantID        string                        `json:"tenant_id"`
	TenantName      string                        `json:"tenant_name"`
	Filters         TenantDeliveryTimelineFilter  `json:"filters"`
	Summary         TenantDeliveryTimelineSummary `json:"summary"`
	Items           []TenantDeliveryTimelineItem  `json:"items"`
	GeneratedAt     time.Time                     `json:"generated_at"`
	PrivacyBoundary string                        `json:"privacy_boundary"`
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

type TenantAlertDeliveryProof struct {
	Channel       string     `json:"channel"`
	Status        string     `json:"status"`
	Provider      string     `json:"provider"`
	Recipient     string     `json:"recipient"`
	Attempts      int        `json:"attempts"`
	LastAttemptAt time.Time  `json:"last_attempt_at"`
	NextRetryAt   *time.Time `json:"next_retry_at,omitempty"`
	Proof         string     `json:"proof"`
}

type TenantAlertInboxSummary struct {
	Total             int `json:"total"`
	Open              int `json:"open"`
	HighOrCritical    int `json:"high_or_critical"`
	WithEmail         int `json:"with_email"`
	WithPush          int `json:"with_push"`
	WithDashboard     int `json:"with_dashboard"`
	DeliveryRetrying  int `json:"delivery_retrying"`
	DeliveryFailed    int `json:"delivery_failed"`
	SourceHostCount   int `json:"source_host_count"`
	NotificationReady int `json:"notification_ready"`
}

type TenantAlertInboxItem struct {
	ID             string                     `json:"id"`
	TenantID       string                     `json:"tenant_id"`
	DeviceID       string                     `json:"device_id"`
	HostName       string                     `json:"host_name"`
	EventID        string                     `json:"event_id"`
	Type           string                     `json:"type"`
	Severity       string                     `json:"severity"`
	Category       string                     `json:"category"`
	Status         string                     `json:"status"`
	Title          string                     `json:"title"`
	Detail         string                     `json:"detail"`
	Recommendation string                     `json:"recommendation"`
	Source         string                     `json:"source"`
	DeliveryState  string                     `json:"delivery_state"`
	DeliveryProof  []TenantAlertDeliveryProof `json:"delivery_proof"`
	NextAction     string                     `json:"next_action"`
	ObservedAt     time.Time                  `json:"observed_at"`
}

type TenantAlertInbox struct {
	TenantID        string                  `json:"tenant_id"`
	TenantName      string                  `json:"tenant_name"`
	Summary         TenantAlertInboxSummary `json:"summary"`
	Items           []TenantAlertInboxItem  `json:"items"`
	GeneratedAt     time.Time               `json:"generated_at"`
	PrivacyBoundary string                  `json:"privacy_boundary"`
}

type TenantNotificationCommandCenterSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	NotificationScore      int    `json:"notification_score"`
	MonetizationReadiness  int    `json:"monetization_readiness"`
	TrustScore             int    `json:"trust_score"`
	OpenAlerts             int    `json:"open_alerts"`
	HighPriorityAlerts     int    `json:"high_priority_alerts"`
	PolicyViolations       int    `json:"policy_violations"`
	Anomalies              int    `json:"anomalies"`
	TamperSignals          int    `json:"tamper_signals"`
	EmailDelivered         int    `json:"email_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	DashboardDelivered     int    `json:"dashboard_delivered"`
	DeliveryFailed         int    `json:"delivery_failed"`
	DeliveryRetrying       int    `json:"delivery_retrying"`
	RoutesTotal            int    `json:"routes_total"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	RemediationOpen        int    `json:"remediation_open"`
	RemediationPlanned     int    `json:"remediation_planned"`
	RemediationSLAWatch    int    `json:"remediation_sla_watch"`
	WeeklyReportReady      bool   `json:"weekly_report_ready"`
	ArchiveBacklog         int    `json:"archive_backlog"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
}

type TenantNotificationCommandCenterChannel struct {
	Channel              string     `json:"channel"`
	Provider             string     `json:"provider"`
	Recipient            string     `json:"recipient"`
	Enabled              bool       `json:"enabled"`
	RouteStatus          string     `json:"route_status"`
	ProofState           string     `json:"proof_state"`
	LatestDeliveryStatus string     `json:"latest_delivery_status"`
	Attempts             int        `json:"attempts"`
	LastDeliveryAt       *time.Time `json:"last_delivery_at,omitempty"`
	NextRetryAt          *time.Time `json:"next_retry_at,omitempty"`
	SLA                  string     `json:"sla"`
	Evidence             string     `json:"evidence"`
	NextAction           string     `json:"next_action"`
	PaidTier             string     `json:"paid_tier"`
}

type TenantNotificationCommandCenterAlert struct {
	ID              string    `json:"id"`
	EventID         string    `json:"event_id"`
	DeviceID        string    `json:"device_id"`
	HostName        string    `json:"host_name"`
	Type            string    `json:"type"`
	Severity        string    `json:"severity"`
	Category        string    `json:"category"`
	Status          string    `json:"status"`
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	Recommendation  string    `json:"recommendation"`
	DeliveryState   string    `json:"delivery_state"`
	EmailStatus     string    `json:"email_status"`
	PushStatus      string    `json:"push_status"`
	DashboardStatus string    `json:"dashboard_status"`
	NextAction      string    `json:"next_action"`
	PaidTier        string    `json:"paid_tier"`
	ObservedAt      time.Time `json:"observed_at"`
}

type TenantNotificationCommandCenterAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Severity   string    `json:"severity"`
	Channel    string    `json:"channel"`
	Status     string    `json:"status"`
	Owner      string    `json:"owner"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantNotificationCommandCenter struct {
	TenantID        string                                   `json:"tenant_id"`
	TenantName      string                                   `json:"tenant_name"`
	PlanID          string                                   `json:"plan_id"`
	PlanName        string                                   `json:"plan_name"`
	Audience        string                                   `json:"audience"`
	Summary         TenantNotificationCommandCenterSummary   `json:"summary"`
	Channels        []TenantNotificationCommandCenterChannel `json:"channels"`
	Alerts          []TenantNotificationCommandCenterAlert   `json:"alerts"`
	Actions         []TenantNotificationCommandCenterAction  `json:"actions"`
	PrivacyBoundary string                                   `json:"privacy_boundary"`
	GeneratedAt     time.Time                                `json:"generated_at"`
}

type TenantNotificationRevenueSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	RevenueReadiness       int    `json:"revenue_readiness"`
	NotificationScore      int    `json:"notification_score"`
	AlertSLAReady          int    `json:"alert_sla_ready"`
	OpenAnomalies          int    `json:"open_anomalies"`
	HighPriorityAlerts     int    `json:"high_priority_alerts"`
	EmailDelivered         int    `json:"email_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	DashboardDelivered     int    `json:"dashboard_delivered"`
	DeliveryFailed         int    `json:"delivery_failed"`
	DeliveryRetrying       int    `json:"delivery_retrying"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	WeeklyReportReady      bool   `json:"weekly_report_ready"`
	EscalationReady        bool   `json:"escalation_ready"`
	BuyerDemoReady         bool   `json:"buyer_demo_ready"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	NextBestAction         string `json:"next_best_action"`
}

type TenantNotificationRevenueKPI struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	PaidTier string `json:"paid_tier"`
}

type TenantNotificationRevenueChannel struct {
	Channel              string     `json:"channel"`
	Provider             string     `json:"provider"`
	RecipientLabel       string     `json:"recipient_label"`
	Status               string     `json:"status"`
	ProofState           string     `json:"proof_state"`
	LatestDeliveryStatus string     `json:"latest_delivery_status"`
	Attempts             int        `json:"attempts"`
	LastDeliveryAt       *time.Time `json:"last_delivery_at,omitempty"`
	SLA                  string     `json:"sla"`
	BusinessValue        string     `json:"business_value"`
	NextAction           string     `json:"next_action"`
	PaidTier             string     `json:"paid_tier"`
}

type TenantNotificationRevenueScenario struct {
	Title          string   `json:"title"`
	Detail         string   `json:"detail"`
	Trigger        string   `json:"trigger"`
	Channels       []string `json:"channels"`
	Severity       string   `json:"severity"`
	Status         string   `json:"status"`
	ExampleOutcome string   `json:"example_outcome"`
	PaidTier       string   `json:"paid_tier"`
}

type TenantNotificationRevenueAction struct {
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	Owner           string    `json:"owner"`
	Status          string    `json:"status"`
	Severity        string    `json:"severity"`
	SLA             string    `json:"sla"`
	ConversionLever string    `json:"conversion_lever"`
	Source          string    `json:"source"`
	ObservedAt      time.Time `json:"observed_at"`
}

type TenantNotificationRevenueCockpit struct {
	TenantID        string                              `json:"tenant_id"`
	TenantName      string                              `json:"tenant_name"`
	PlanID          string                              `json:"plan_id"`
	PlanName        string                              `json:"plan_name"`
	Audience        string                              `json:"audience"`
	Summary         TenantNotificationRevenueSummary    `json:"summary"`
	KPIs            []TenantNotificationRevenueKPI      `json:"kpis"`
	Channels        []TenantNotificationRevenueChannel  `json:"channels"`
	Scenarios       []TenantNotificationRevenueScenario `json:"scenarios"`
	Actions         []TenantNotificationRevenueAction   `json:"actions"`
	PrivacyBoundary string                              `json:"privacy_boundary"`
	GeneratedAt     time.Time                           `json:"generated_at"`
}

type TenantCustomerControlSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	ProductScore           int    `json:"product_score"`
	NotificationScore      int    `json:"notification_score"`
	PackageScore           int    `json:"package_score"`
	TrustScore             int    `json:"trust_score"`
	CustomerHealth         string `json:"customer_health"`
	OpenAlerts             int    `json:"open_alerts"`
	HighPriorityAlerts     int    `json:"high_priority_alerts"`
	HostsTotal             int    `json:"hosts_total"`
	HostsAttention         int    `json:"hosts_attention"`
	MailDelivered          int    `json:"mail_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	DashboardDelivered     int    `json:"dashboard_delivered"`
	DeliveryFailed         int    `json:"delivery_failed"`
	DeliveryRetrying       int    `json:"delivery_retrying"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	WeeklyReportReady      bool   `json:"weekly_report_ready"`
	ArchiveBacklog         int    `json:"archive_backlog"`
	BillingReady           bool   `json:"billing_ready"`
	ProviderReady          bool   `json:"provider_ready"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	NextBestAction         string `json:"next_best_action"`
}

type TenantCustomerControlTile struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	Channel  string `json:"channel,omitempty"`
	PaidTier string `json:"paid_tier"`
}

type TenantCustomerControlAlert struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"`
	HostName        string    `json:"host_name"`
	Category        string    `json:"category"`
	EmailStatus     string    `json:"email_status"`
	PushStatus      string    `json:"push_status"`
	DashboardStatus string    `json:"dashboard_status"`
	NextAction      string    `json:"next_action"`
	PaidTier        string    `json:"paid_tier"`
	ObservedAt      time.Time `json:"observed_at"`
}

type TenantCustomerControlDelivery struct {
	Channel        string     `json:"channel"`
	Provider       string     `json:"provider"`
	RecipientLabel string     `json:"recipient_label"`
	Status         string     `json:"status"`
	ProofState     string     `json:"proof_state"`
	Attempts       int        `json:"attempts"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	SLA            string     `json:"sla"`
	Evidence       string     `json:"evidence"`
	NextAction     string     `json:"next_action"`
	PaidTier       string     `json:"paid_tier"`
}

type TenantCustomerControlAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"`
	Owner      string    `json:"owner"`
	Channel    string    `json:"channel"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantCustomerControlRoom struct {
	TenantID        string                          `json:"tenant_id"`
	TenantName      string                          `json:"tenant_name"`
	PlanID          string                          `json:"plan_id"`
	PlanName        string                          `json:"plan_name"`
	Audience        string                          `json:"audience"`
	Summary         TenantCustomerControlSummary    `json:"summary"`
	Tiles           []TenantCustomerControlTile     `json:"tiles"`
	Alerts          []TenantCustomerControlAlert    `json:"alerts"`
	Deliveries      []TenantCustomerControlDelivery `json:"deliveries"`
	Actions         []TenantCustomerControlAction   `json:"actions"`
	PrivacyBoundary string                          `json:"privacy_boundary"`
	GeneratedAt     time.Time                       `json:"generated_at"`
}

type TenantCustomerSuccessPacketSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	ReadinessScore         int    `json:"readiness_score"`
	NotificationScore      int    `json:"notification_score"`
	PackageScore           int    `json:"package_score"`
	TrustScore             int    `json:"trust_score"`
	OpenAlerts             int    `json:"open_alerts"`
	HighPriorityAlerts     int    `json:"high_priority_alerts"`
	HostsTotal             int    `json:"hosts_total"`
	MailDelivered          int    `json:"mail_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	WeeklyReportReady      bool   `json:"weekly_report_ready"`
	ArchiveBacklog         int    `json:"archive_backlog"`
	ProviderReady          bool   `json:"provider_ready"`
	BillingReady           bool   `json:"billing_ready"`
	RolesReady             int    `json:"roles_ready"`
	RolesTotal             int    `json:"roles_total"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	OwnerNextStep          string `json:"owner_next_step"`
}

type TenantCustomerSuccessPacketProof struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Value       string `json:"value"`
	Detail      string `json:"detail"`
	Status      string `json:"status"`
	Evidence    string `json:"evidence"`
	PaidTier    string `json:"paid_tier"`
	BuyerImpact string `json:"buyer_impact"`
}

type TenantCustomerSuccessPacketObjection struct {
	ID       string `json:"id"`
	Concern  string `json:"concern"`
	Answer   string `json:"answer"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
	Owner    string `json:"owner"`
}

type TenantCustomerSuccessPacketAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Owner      string    `json:"owner"`
	Status     string    `json:"status"`
	Severity   string    `json:"severity"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantCustomerSuccessPacket struct {
	TenantID        string                                 `json:"tenant_id"`
	TenantName      string                                 `json:"tenant_name"`
	PlanID          string                                 `json:"plan_id"`
	PlanName        string                                 `json:"plan_name"`
	Audience        string                                 `json:"audience"`
	Summary         TenantCustomerSuccessPacketSummary     `json:"summary"`
	Proofs          []TenantCustomerSuccessPacketProof     `json:"proofs"`
	Objections      []TenantCustomerSuccessPacketObjection `json:"objections"`
	Actions         []TenantCustomerSuccessPacketAction    `json:"actions"`
	PrivacyBoundary string                                 `json:"privacy_boundary"`
	GeneratedAt     time.Time                              `json:"generated_at"`
}

type TenantPushActivationSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	ActivationScore        int    `json:"activation_score"`
	NotificationScore      int    `json:"notification_score"`
	MailDelivered          int    `json:"mail_delivered"`
	DashboardDelivered     int    `json:"dashboard_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	PushRetrying           int    `json:"push_retrying"`
	PushFailed             int    `json:"push_failed"`
	PushPending            int    `json:"push_pending"`
	PushRoutesTotal        int    `json:"push_routes_total"`
	PushRoutesReady        int    `json:"push_routes_ready"`
	PushRoutesNeedingProof int    `json:"push_routes_needing_proof"`
	AlertRulesUsingPush    int    `json:"alert_rules_using_push"`
	AlertsWithPush         int    `json:"alerts_with_push"`
	PushPreferenceEnabled  bool   `json:"push_preference_enabled"`
	PushEscalationEnabled  bool   `json:"push_escalation_enabled"`
	QuietHoursProtected    bool   `json:"quiet_hours_protected"`
	PushSimulationReady    bool   `json:"push_simulation_ready"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	OwnerNextStep          string `json:"owner_next_step"`
}

type TenantPushActivationRoute struct {
	RouteID              string     `json:"route_id"`
	Provider             string     `json:"provider"`
	SubscriptionLabel    string     `json:"subscription_label"`
	Status               string     `json:"status"`
	ProofState           string     `json:"proof_state"`
	LatestDeliveryStatus string     `json:"latest_delivery_status"`
	LatestDeliveryAt     *time.Time `json:"latest_delivery_at,omitempty"`
	Attempts             int        `json:"attempts"`
	NextRetryAt          *time.Time `json:"next_retry_at,omitempty"`
	SimulationStatus     string     `json:"simulation_status"`
	SLATarget            string     `json:"sla_target"`
	EndpointStorage      string     `json:"endpoint_storage"`
	Evidence             string     `json:"evidence"`
	NextAction           string     `json:"next_action"`
	PaidTier             string     `json:"paid_tier"`
}

type TenantPushActivationScenario struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Trigger    string   `json:"trigger"`
	Channels   []string `json:"channels"`
	Status     string   `json:"status"`
	BuyerValue string   `json:"buyer_value"`
	StudySafe  bool     `json:"study_safe"`
	PaidTier   string   `json:"paid_tier"`
}

type TenantPushActivationAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Owner      string    `json:"owner"`
	Status     string    `json:"status"`
	Severity   string    `json:"severity"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantPushActivationCenter struct {
	TenantID        string                         `json:"tenant_id"`
	TenantName      string                         `json:"tenant_name"`
	PlanID          string                         `json:"plan_id"`
	PlanName        string                         `json:"plan_name"`
	Audience        string                         `json:"audience"`
	Summary         TenantPushActivationSummary    `json:"summary"`
	Routes          []TenantPushActivationRoute    `json:"routes"`
	Scenarios       []TenantPushActivationScenario `json:"scenarios"`
	Actions         []TenantPushActivationAction   `json:"actions"`
	PrivacyBoundary string                         `json:"privacy_boundary"`
	GeneratedAt     time.Time                      `json:"generated_at"`
}

type TenantPortfolioSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	PortfolioScore         int    `json:"portfolio_score"`
	NotificationScore      int    `json:"notification_score"`
	TrustScore             int    `json:"trust_score"`
	RiskScore              int    `json:"risk_score"`
	HostsTotal             int    `json:"hosts_total"`
	HostsAttention         int    `json:"hosts_attention"`
	HostsReporting         int    `json:"hosts_reporting"`
	HostsPending           int    `json:"hosts_pending"`
	OpenAlerts             int    `json:"open_alerts"`
	HighPriorityAlerts     int    `json:"high_priority_alerts"`
	MailDelivered          int    `json:"mail_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	PushRetrying           int    `json:"push_retrying"`
	DashboardDelivered     int    `json:"dashboard_delivered"`
	ArchiveBacklog         int    `json:"archive_backlog"`
	StoredTelemetryEvents  int    `json:"stored_telemetry_events"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	OwnerNextStep          string `json:"owner_next_step"`
}

type TenantPortfolioHost struct {
	DeviceID             string    `json:"device_id"`
	HostName             string    `json:"host_name"`
	Profile              string    `json:"profile"`
	OSName               string    `json:"os_name"`
	Status               string    `json:"status"`
	RiskLevel            string    `json:"risk_level"`
	RiskScore            int       `json:"risk_score"`
	HealthScore          int       `json:"health_score"`
	ComplianceScore      int       `json:"compliance_score"`
	PolicyViolations     int       `json:"policy_violations"`
	Anomalies            int       `json:"anomalies"`
	TamperSignals        int       `json:"tamper_signals"`
	ArchiveBacklog       int       `json:"archive_backlog"`
	DataCompletenessPct  int       `json:"data_completeness_pct"`
	EmailStatus          string    `json:"email_status"`
	PushStatus           string    `json:"push_status"`
	DashboardStatus      string    `json:"dashboard_status"`
	LastSeenAt           time.Time `json:"last_seen_at"`
	LastDeliveryAt       time.Time `json:"last_delivery_at"`
	NextAction           string    `json:"next_action"`
	PaidTier             string    `json:"paid_tier"`
	MetadataProofSummary string    `json:"metadata_proof_summary"`
}

type TenantPortfolioSegment struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	PaidTier string `json:"paid_tier"`
}

type TenantPortfolioAlertNotification struct {
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"`
	HostName        string    `json:"host_name"`
	Category        string    `json:"category"`
	EmailStatus     string    `json:"email_status"`
	PushStatus      string    `json:"push_status"`
	DashboardStatus string    `json:"dashboard_status"`
	NextAction      string    `json:"next_action"`
	PaidTier        string    `json:"paid_tier"`
	ObservedAt      time.Time `json:"observed_at"`
}

type TenantPortfolioDeliveryProof struct {
	Label      string     `json:"label"`
	Value      string     `json:"value"`
	Detail     string     `json:"detail"`
	Channel    string     `json:"channel"`
	Status     string     `json:"status"`
	ProofState string     `json:"proof_state"`
	PaidTier   string     `json:"paid_tier"`
	NextAction string     `json:"next_action"`
	ObservedAt *time.Time `json:"observed_at,omitempty"`
}

type TenantPortfolioAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Owner      string    `json:"owner"`
	Status     string    `json:"status"`
	Severity   string    `json:"severity"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantPortfolioCenter struct {
	TenantID           string                             `json:"tenant_id"`
	TenantName         string                             `json:"tenant_name"`
	PlanID             string                             `json:"plan_id"`
	PlanName           string                             `json:"plan_name"`
	Audience           string                             `json:"audience"`
	Summary            TenantPortfolioSummary             `json:"summary"`
	Hosts              []TenantPortfolioHost              `json:"hosts"`
	Segments           []TenantPortfolioSegment           `json:"segments"`
	AlertNotifications []TenantPortfolioAlertNotification `json:"alert_notifications"`
	DeliveryProof      []TenantPortfolioDeliveryProof     `json:"delivery_proof"`
	Actions            []TenantPortfolioAction            `json:"actions"`
	PrivacyBoundary    string                             `json:"privacy_boundary"`
	GeneratedAt        time.Time                          `json:"generated_at"`
}

type TenantActivityView struct {
	ID          string                   `json:"id"`
	TenantID    string                   `json:"tenant_id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Filter      TenantActivityFeedFilter `json:"filter"`
	PaidTier    string                   `json:"paid_tier"`
	SortOrder   int                      `json:"sort_order"`
	CreatedAt   time.Time                `json:"created_at"`
	UpdatedAt   time.Time                `json:"updated_at"`
}

type CreateTenantActivityViewRequest struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Description string                   `json:"description"`
	Filter      TenantActivityFeedFilter `json:"filter"`
	PaidTier    string                   `json:"paid_tier"`
	SortOrder   int                      `json:"sort_order"`
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

type NotificationQuietHours struct {
	Enabled    bool   `json:"enabled"`
	StartLocal string `json:"start_local"`
	EndLocal   string `json:"end_local"`
	Timezone   string `json:"timezone"`
}

type NotificationEscalationPolicy struct {
	Enabled         bool     `json:"enabled"`
	AfterMinutes    int      `json:"after_minutes"`
	RepeatEveryMins int      `json:"repeat_every_minutes"`
	MaxRepeats      int      `json:"max_repeats"`
	Channels        []string `json:"channels"`
	Owner           string   `json:"owner"`
}

type NotificationPreferenceRule struct {
	ID                string    `json:"id"`
	TenantID          string    `json:"tenant_id"`
	Name              string    `json:"name"`
	EventType         string    `json:"event_type"`
	Severity          string    `json:"severity"`
	Channels          []string  `json:"channels"`
	Mode              string    `json:"mode"`
	RecipientGroup    string    `json:"recipient_group"`
	SuppressionLabel  string    `json:"suppression_label"`
	StudySafe         bool      `json:"study_safe"`
	QuietHoursBypass  bool      `json:"quiet_hours_bypass"`
	PaidTier          string    `json:"paid_tier"`
	DeliverySLA       string    `json:"delivery_sla"`
	NextAction        string    `json:"next_action"`
	RetentionEvidence string    `json:"retention_evidence"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type NotificationPreferenceCenterSummary struct {
	Status                string `json:"status"`
	PreferenceScore       int    `json:"preference_score"`
	RulesTotal            int    `json:"rules_total"`
	ImmediateRules        int    `json:"immediate_rules"`
	DigestRules           int    `json:"digest_rules"`
	SilentRules           int    `json:"silent_rules"`
	EmailEnabled          bool   `json:"email_enabled"`
	PushEnabled           bool   `json:"push_enabled"`
	DashboardEnabled      bool   `json:"dashboard_enabled"`
	QuietHoursEnabled     bool   `json:"quiet_hours_enabled"`
	EscalationEnabled     bool   `json:"escalation_enabled"`
	StudySuppressionRules int    `json:"study_suppression_rules"`
	RoutesNeedingProof    int    `json:"routes_needing_proof"`
	RecommendedPaidTier   string `json:"recommended_paid_tier"`
}

type NotificationPreferenceCenter struct {
	TenantID        string                              `json:"tenant_id"`
	TenantName      string                              `json:"tenant_name"`
	PlanID          string                              `json:"plan_id"`
	PlanName        string                              `json:"plan_name"`
	Audience        string                              `json:"audience"`
	DigestCadence   string                              `json:"digest_cadence"`
	QuietHours      NotificationQuietHours              `json:"quiet_hours"`
	Escalation      NotificationEscalationPolicy        `json:"escalation"`
	Summary         NotificationPreferenceCenterSummary `json:"summary"`
	Rules           []NotificationPreferenceRule        `json:"rules"`
	PrivacyBoundary string                              `json:"privacy_boundary"`
	GeneratedAt     time.Time                           `json:"generated_at"`
	UpdatedAt       time.Time                           `json:"updated_at"`
}

type UpdateNotificationPreferencesRequest struct {
	DigestCadence string                       `json:"digest_cadence"`
	QuietHours    NotificationQuietHours       `json:"quiet_hours"`
	Escalation    NotificationEscalationPolicy `json:"escalation"`
	Rules         []NotificationPreferenceRule `json:"rules"`
}

type RunDeliveryDrilldownRequest struct {
	Mode    string `json:"mode"`
	Channel string `json:"channel"`
	Reason  string `json:"reason"`
}

type RunDeliveryRemediationRequest struct {
	Mode    string `json:"mode"`
	Channel string `json:"channel"`
	RouteID string `json:"route_id"`
	Action  string `json:"action"`
	Reason  string `json:"reason"`
	Owner   string `json:"owner"`
}

type TenantDeliveryDrilldownSummary struct {
	RoutesTotal        int        `json:"routes_total"`
	EnabledRoutes      int        `json:"enabled_routes"`
	HealthyRoutes      int        `json:"healthy_routes"`
	RoutesNeedingProof int        `json:"routes_needing_proof"`
	EmailReady         bool       `json:"email_ready"`
	PushReady          bool       `json:"push_ready"`
	DashboardReady     bool       `json:"dashboard_ready"`
	DeliveryScore      int        `json:"delivery_score"`
	RehearsalMode      string     `json:"rehearsal_mode"`
	LastRehearsedAt    *time.Time `json:"last_rehearsed_at,omitempty"`
}

type TenantDeliveryDrilldownRoute struct {
	RouteID              string     `json:"route_id"`
	Channel              string     `json:"channel"`
	Provider             string     `json:"provider"`
	RecipientLabel       string     `json:"recipient_label"`
	Enabled              bool       `json:"enabled"`
	RouteStatus          string     `json:"route_status"`
	LastVerifiedAt       *time.Time `json:"last_verified_at,omitempty"`
	LatestDeliveryStatus string     `json:"latest_delivery_status"`
	LatestDeliveryAt     *time.Time `json:"latest_delivery_at,omitempty"`
	Attempts             int        `json:"attempts"`
	ProofState           string     `json:"proof_state"`
	RehearsalResult      string     `json:"rehearsal_result"`
	SLA                  string     `json:"sla"`
	NextAction           string     `json:"next_action"`
	Evidence             string     `json:"evidence"`
}

type TenantDeliveryDrilldown struct {
	TenantID        string                         `json:"tenant_id"`
	GeneratedAt     time.Time                      `json:"generated_at"`
	PrivacyBoundary string                         `json:"privacy_boundary"`
	Summary         TenantDeliveryDrilldownSummary `json:"summary"`
	Routes          []TenantDeliveryDrilldownRoute `json:"routes"`
	Actions         []TenantOperationsSignal       `json:"actions"`
}

type TenantDeliveryRemediationSummary struct {
	RoutesTotal        int        `json:"routes_total"`
	ProblemsOpen       int        `json:"problems_open"`
	PlannedActions     int        `json:"planned_actions"`
	OwnerAcknowledged  int        `json:"owner_acknowledged"`
	SLAWatch           int        `json:"sla_watch"`
	RemediationScore   int        `json:"remediation_score"`
	EmailProtected     bool       `json:"email_protected"`
	PushProtected      bool       `json:"push_protected"`
	DashboardProtected bool       `json:"dashboard_protected"`
	NextRetryAt        *time.Time `json:"next_retry_at,omitempty"`
	LastPlannedAt      *time.Time `json:"last_planned_at,omitempty"`
}

type TenantDeliveryRemediationAction struct {
	ID                   string     `json:"id"`
	TenantID             string     `json:"tenant_id"`
	RouteID              string     `json:"route_id"`
	Channel              string     `json:"channel"`
	Provider             string     `json:"provider"`
	RecipientLabel       string     `json:"recipient_label"`
	Action               string     `json:"action"`
	Status               string     `json:"status"`
	Owner                string     `json:"owner"`
	Problem              string     `json:"problem"`
	Plan                 string     `json:"plan"`
	SLATarget            string     `json:"sla_target"`
	LatestDeliveryStatus string     `json:"latest_delivery_status"`
	LatestDeliveryAt     *time.Time `json:"latest_delivery_at,omitempty"`
	NextRetryAt          *time.Time `json:"next_retry_at,omitempty"`
	AuditState           string     `json:"audit_state"`
	PrivacyBoundary      string     `json:"privacy_boundary"`
	CreatedAt            time.Time  `json:"created_at"`
}

type TenantDeliveryRemediation struct {
	TenantID        string                            `json:"tenant_id"`
	GeneratedAt     time.Time                         `json:"generated_at"`
	PrivacyBoundary string                            `json:"privacy_boundary"`
	Summary         TenantDeliveryRemediationSummary  `json:"summary"`
	Actions         []TenantDeliveryRemediationAction `json:"actions"`
	RecentPlans     []TenantDeliveryRemediationAction `json:"recent_plans"`
}

type RunProviderSimulationRequest struct {
	Mode     string `json:"mode"`
	Channel  string `json:"channel"`
	Scenario string `json:"scenario"`
	Reason   string `json:"reason"`
}

type TenantProviderSimulationSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	ReadinessScore         int    `json:"readiness_score"`
	SimulationScore        int    `json:"simulation_score"`
	RoutesTotal            int    `json:"routes_total"`
	SimulatedRoutes        int    `json:"simulated_routes"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	ProviderRisks          int    `json:"provider_risks"`
	EmailReady             bool   `json:"email_ready"`
	PushReady              bool   `json:"push_ready"`
	DashboardReady         bool   `json:"dashboard_ready"`
	SLAReady               bool   `json:"sla_ready"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	NextBestAction         string `json:"next_best_action"`
}

type TenantProviderSimulationRoute struct {
	RouteID              string     `json:"route_id"`
	Channel              string     `json:"channel"`
	Provider             string     `json:"provider"`
	RecipientLabel       string     `json:"recipient_label"`
	SimulationStatus     string     `json:"simulation_status"`
	ProofState           string     `json:"proof_state"`
	Scenario             string     `json:"scenario"`
	SLATarget            string     `json:"sla_target"`
	SimulatedLatency     string     `json:"simulated_latency"`
	LatestDeliveryStatus string     `json:"latest_delivery_status"`
	LastSimulatedAt      *time.Time `json:"last_simulated_at,omitempty"`
	BusinessValue        string     `json:"business_value"`
	Evidence             string     `json:"evidence"`
	NextAction           string     `json:"next_action"`
	PaidTier             string     `json:"paid_tier"`
}

type TenantProviderSimulationScenario struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Trigger    string   `json:"trigger"`
	Channels   []string `json:"channels"`
	Severity   string   `json:"severity"`
	Outcome    string   `json:"outcome"`
	BuyerValue string   `json:"buyer_value"`
	PaidTier   string   `json:"paid_tier"`
	StudySafe  bool     `json:"study_safe"`
}

type TenantProviderSimulationAction struct {
	Title           string `json:"title"`
	Detail          string `json:"detail"`
	Owner           string `json:"owner"`
	Channel         string `json:"channel"`
	Status          string `json:"status"`
	SLA             string `json:"sla"`
	ConversionLever string `json:"conversion_lever"`
	PaidTier        string `json:"paid_tier"`
}

type TenantProviderSimulationLab struct {
	TenantID        string                             `json:"tenant_id"`
	TenantName      string                             `json:"tenant_name"`
	PlanID          string                             `json:"plan_id"`
	PlanName        string                             `json:"plan_name"`
	Audience        string                             `json:"audience"`
	Summary         TenantProviderSimulationSummary    `json:"summary"`
	Routes          []TenantProviderSimulationRoute    `json:"routes"`
	Scenarios       []TenantProviderSimulationScenario `json:"scenarios"`
	Actions         []TenantProviderSimulationAction   `json:"actions"`
	PrivacyBoundary string                             `json:"privacy_boundary"`
	GeneratedAt     time.Time                          `json:"generated_at"`
}

type TenantPackageBillingSummary struct {
	Status             string `json:"status"`
	Headline           string `json:"headline"`
	Detail             string `json:"detail"`
	PackageScore       int    `json:"package_score"`
	BillingStatus      string `json:"billing_status"`
	RevenueStage       string `json:"revenue_stage"`
	CurrentPlan        string `json:"current_plan"`
	RecommendedPackage string `json:"recommended_package"`
	SeatsUsed          int    `json:"seats_used"`
	SeatsIncluded      int    `json:"seats_included"`
	SeatUtilization    int    `json:"seat_utilization"`
	FeatureGatesReady  int    `json:"feature_gates_ready"`
	FeatureGatesTotal  int    `json:"feature_gates_total"`
	UpgradeReady       bool   `json:"upgrade_ready"`
	BillingReady       bool   `json:"billing_ready"`
	RetentionReady     bool   `json:"retention_ready"`
	ArchiveReady       bool   `json:"archive_ready"`
	WeeklyReportReady  bool   `json:"weekly_report_ready"`
	NotificationReady  bool   `json:"notification_ready"`
	ProviderReady      bool   `json:"provider_ready"`
	TrustScore         int    `json:"trust_score"`
	NextBestAction     string `json:"next_best_action"`
}

type TenantPackageBillingPlan struct {
	PlanID      string   `json:"plan_id"`
	Name        string   `json:"name"`
	Audience    string   `json:"audience"`
	PriceModel  string   `json:"price_model"`
	Status      string   `json:"status"`
	Current     bool     `json:"current"`
	Recommended bool     `json:"recommended"`
	FitScore    int      `json:"fit_score"`
	Features    []string `json:"features"`
	Value       string   `json:"value"`
	NextAction  string   `json:"next_action"`
}

type TenantPackageBillingFeatureGate struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Status     string `json:"status"`
	Enabled    bool   `json:"enabled"`
	Evidence   string `json:"evidence"`
	BuyerValue string `json:"buyer_value"`
	PaidTier   string `json:"paid_tier"`
}

type TenantPackageBillingMilestone struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Detail     string `json:"detail"`
	Status     string `json:"status"`
	Owner      string `json:"owner"`
	Evidence   string `json:"evidence"`
	NextAction string `json:"next_action"`
	SLA        string `json:"sla"`
	PaidTier   string `json:"paid_tier"`
}

type TenantPackageBillingAction struct {
	Title           string `json:"title"`
	Detail          string `json:"detail"`
	Owner           string `json:"owner"`
	Status          string `json:"status"`
	PaidTier        string `json:"paid_tier"`
	ConversionLever string `json:"conversion_lever"`
	NextAction      string `json:"next_action"`
}

type TenantPackageBillingReadiness struct {
	TenantID        string                            `json:"tenant_id"`
	TenantName      string                            `json:"tenant_name"`
	PlanID          string                            `json:"plan_id"`
	PlanName        string                            `json:"plan_name"`
	Audience        string                            `json:"audience"`
	RetentionTierID string                            `json:"retention_tier_id"`
	RetentionName   string                            `json:"retention_name"`
	Summary         TenantPackageBillingSummary       `json:"summary"`
	Plans           []TenantPackageBillingPlan        `json:"plans"`
	FeatureGates    []TenantPackageBillingFeatureGate `json:"feature_gates"`
	Milestones      []TenantPackageBillingMilestone   `json:"milestones"`
	Actions         []TenantPackageBillingAction      `json:"actions"`
	PrivacyBoundary string                            `json:"privacy_boundary"`
	GeneratedAt     time.Time                         `json:"generated_at"`
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

type TenantBusinessDashboardSummary struct {
	Status             string `json:"status"`
	Headline           string `json:"headline"`
	Detail             string `json:"detail"`
	ProductScore       int    `json:"product_score"`
	CustomerHealth     string `json:"customer_health"`
	RevenueStage       string `json:"revenue_stage"`
	RecommendedPackage string `json:"recommended_package"`
	HostsTotal         int    `json:"hosts_total"`
	HostsAttention     int    `json:"hosts_attention"`
	OpenAlerts         int    `json:"open_alerts"`
	HighPriorityAlerts int    `json:"high_priority_alerts"`
	NotificationScore  int    `json:"notification_score"`
	PreferenceScore    int    `json:"preference_score"`
	TrustScore         int    `json:"trust_score"`
	MailDelivered      int    `json:"mail_delivered"`
	PushDelivered      int    `json:"push_delivered"`
	DashboardDelivered int    `json:"dashboard_delivered"`
	RoutesNeedingProof int    `json:"routes_needing_proof"`
	ArchiveBacklog     int    `json:"archive_backlog"`
	WeeklyReportReady  bool   `json:"weekly_report_ready"`
	ConsentVisible     bool   `json:"consent_visible"`
	DataRightsReady    bool   `json:"data_rights_ready"`
}

type TenantBusinessDashboardMetric struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	PaidTier string `json:"paid_tier"`
}

type TenantBusinessDashboardAlert struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"`
	HostName        string    `json:"host_name"`
	Category        string    `json:"category"`
	EmailStatus     string    `json:"email_status"`
	PushStatus      string    `json:"push_status"`
	DashboardStatus string    `json:"dashboard_status"`
	NextAction      string    `json:"next_action"`
	PaidTier        string    `json:"paid_tier"`
	ObservedAt      time.Time `json:"observed_at"`
}

type TenantBusinessDashboardChannel struct {
	Channel        string     `json:"channel"`
	Provider       string     `json:"provider"`
	Status         string     `json:"status"`
	ProofState     string     `json:"proof_state"`
	RecipientLabel string     `json:"recipient_label"`
	Attempts       int        `json:"attempts"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	NextAction     string     `json:"next_action"`
	PaidTier       string     `json:"paid_tier"`
}

type TenantBusinessDashboardPackage struct {
	Name       string   `json:"name"`
	Tier       string   `json:"tier"`
	Audience   string   `json:"audience"`
	PriceModel string   `json:"price_model"`
	Status     string   `json:"status"`
	Included   []string `json:"included"`
	Value      string   `json:"value"`
	NextAction string   `json:"next_action"`
}

type TenantBusinessDashboardAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"`
	Owner      string    `json:"owner"`
	Channel    string    `json:"channel"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantBusinessDashboard struct {
	TenantID        string                           `json:"tenant_id"`
	TenantName      string                           `json:"tenant_name"`
	PlanID          string                           `json:"plan_id"`
	PlanName        string                           `json:"plan_name"`
	Audience        string                           `json:"audience"`
	Summary         TenantBusinessDashboardSummary   `json:"summary"`
	Metrics         []TenantBusinessDashboardMetric  `json:"metrics"`
	Alerts          []TenantBusinessDashboardAlert   `json:"alerts"`
	Channels        []TenantBusinessDashboardChannel `json:"channels"`
	Packages        []TenantBusinessDashboardPackage `json:"packages"`
	Actions         []TenantBusinessDashboardAction  `json:"actions"`
	PrivacyBoundary string                           `json:"privacy_boundary"`
	GeneratedAt     time.Time                        `json:"generated_at"`
}

type TenantRoleExperienceSummary struct {
	Status             string `json:"status"`
	Headline           string `json:"headline"`
	Detail             string `json:"detail"`
	ReadinessScore     int    `json:"readiness_score"`
	RolesTotal         int    `json:"roles_total"`
	RolesReady         int    `json:"roles_ready"`
	OwnerActions       int    `json:"owner_actions"`
	NotificationScore  int    `json:"notification_score"`
	TrustScore         int    `json:"trust_score"`
	PrivacyVisible     bool   `json:"privacy_visible"`
	RecommendedPackage string `json:"recommended_package"`
}

type TenantRoleExperienceMetric struct {
	Label  string `json:"label"`
	Value  string `json:"value"`
	Detail string `json:"detail"`
	Status string `json:"status"`
}

type TenantRoleExperienceRole struct {
	RoleID               string                       `json:"role_id"`
	Name                 string                       `json:"name"`
	Audience             string                       `json:"audience"`
	ViewMode             string                       `json:"view_mode"`
	Status               string                       `json:"status"`
	ReadinessScore       int                          `json:"readiness_score"`
	PrimaryGoal          string                       `json:"primary_goal"`
	VisiblePanels        []string                     `json:"visible_panels"`
	NotificationPromise  string                       `json:"notification_promise"`
	ArchiveReportPromise string                       `json:"archive_report_promise"`
	ConsentControls      string                       `json:"consent_controls"`
	PaidTier             string                       `json:"paid_tier"`
	NextAction           string                       `json:"next_action"`
	Metrics              []TenantRoleExperienceMetric `json:"metrics"`
}

type TenantRoleOnboardingItem struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Detail   string `json:"detail"`
	Owner    string `json:"owner"`
	Status   string `json:"status"`
	PaidTier string `json:"paid_tier"`
	Evidence string `json:"evidence"`
}

type TenantRoleExperience struct {
	TenantID        string                      `json:"tenant_id"`
	TenantName      string                      `json:"tenant_name"`
	PlanID          string                      `json:"plan_id"`
	PlanName        string                      `json:"plan_name"`
	Audience        string                      `json:"audience"`
	Summary         TenantRoleExperienceSummary `json:"summary"`
	Roles           []TenantRoleExperienceRole  `json:"roles"`
	Onboarding      []TenantRoleOnboardingItem  `json:"onboarding"`
	PrivacyBoundary string                      `json:"privacy_boundary"`
	GeneratedAt     time.Time                   `json:"generated_at"`
}

type TenantExecutiveConsoleSummary struct {
	Status                 string `json:"status"`
	Headline               string `json:"headline"`
	Detail                 string `json:"detail"`
	ReadinessScore         int    `json:"readiness_score"`
	NotificationScore      int    `json:"notification_score"`
	TrustScore             int    `json:"trust_score"`
	OpenAlerts             int    `json:"open_alerts"`
	HighPriorityAlerts     int    `json:"high_priority_alerts"`
	HostsTotal             int    `json:"hosts_total"`
	HostsAttention         int    `json:"hosts_attention"`
	EmailDelivered         int    `json:"email_delivered"`
	PushDelivered          int    `json:"push_delivered"`
	DashboardDelivered     int    `json:"dashboard_delivered"`
	DeliveryFailed         int    `json:"delivery_failed"`
	DeliveryRetrying       int    `json:"delivery_retrying"`
	RoutesNeedingProof     int    `json:"routes_needing_proof"`
	WeeklyReportReady      bool   `json:"weekly_report_ready"`
	ArchiveBacklog         int    `json:"archive_backlog"`
	RolesReady             int    `json:"roles_ready"`
	RolesTotal             int    `json:"roles_total"`
	RecommendedPaidPackage string `json:"recommended_paid_package"`
	NextBestAction         string `json:"next_best_action"`
}

type TenantExecutiveConsoleTile struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Value    string `json:"value"`
	Detail   string `json:"detail"`
	Status   string `json:"status"`
	PaidTier string `json:"paid_tier"`
}

type TenantExecutiveConsoleAlert struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Detail          string    `json:"detail"`
	Severity        string    `json:"severity"`
	Status          string    `json:"status"`
	HostName        string    `json:"host_name"`
	Category        string    `json:"category"`
	EmailStatus     string    `json:"email_status"`
	PushStatus      string    `json:"push_status"`
	DashboardStatus string    `json:"dashboard_status"`
	NextAction      string    `json:"next_action"`
	PaidTier        string    `json:"paid_tier"`
	ObservedAt      time.Time `json:"observed_at"`
}

type TenantExecutiveConsoleDelivery struct {
	Channel        string     `json:"channel"`
	Provider       string     `json:"provider"`
	Status         string     `json:"status"`
	ProofState     string     `json:"proof_state"`
	RecipientLabel string     `json:"recipient_label"`
	Attempts       int        `json:"attempts"`
	LastDeliveryAt *time.Time `json:"last_delivery_at,omitempty"`
	SLA            string     `json:"sla"`
	Evidence       string     `json:"evidence"`
	NextAction     string     `json:"next_action"`
	PaidTier       string     `json:"paid_tier"`
}

type TenantExecutiveConsoleAction struct {
	Title      string    `json:"title"`
	Detail     string    `json:"detail"`
	Severity   string    `json:"severity"`
	Status     string    `json:"status"`
	Owner      string    `json:"owner"`
	Channel    string    `json:"channel"`
	SLA        string    `json:"sla"`
	PaidTier   string    `json:"paid_tier"`
	Source     string    `json:"source"`
	ObservedAt time.Time `json:"observed_at"`
}

type TenantExecutiveConsole struct {
	TenantID        string                           `json:"tenant_id"`
	TenantName      string                           `json:"tenant_name"`
	PlanID          string                           `json:"plan_id"`
	PlanName        string                           `json:"plan_name"`
	Audience        string                           `json:"audience"`
	Summary         TenantExecutiveConsoleSummary    `json:"summary"`
	Tiles           []TenantExecutiveConsoleTile     `json:"tiles"`
	Alerts          []TenantExecutiveConsoleAlert    `json:"alerts"`
	Deliveries      []TenantExecutiveConsoleDelivery `json:"deliveries"`
	Actions         []TenantExecutiveConsoleAction   `json:"actions"`
	PrivacyBoundary string                           `json:"privacy_boundary"`
	GeneratedAt     time.Time                        `json:"generated_at"`
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
