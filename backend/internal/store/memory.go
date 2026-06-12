package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
)

var (
	ErrDeviceNotFound = errors.New("device not found")
	ErrTenantNotFound = errors.New("tenant not found")
)

type Memory struct {
	mu              sync.RWMutex
	path            string
	devices         map[string]model.Device
	tenants         map[string]model.Tenant
	auditEvents     []model.AuditEvent
	policyEvents    map[string][]model.RiskEvent
	anomalyEvents   map[string][]model.RiskEvent
	tamperEvents    map[string][]model.RiskEvent
	alertDeliveries map[string][]model.AlertDelivery
	healthScores    map[string]model.DeviceHealth
}

func NewMemory() *Memory {
	return &Memory{
		devices:         make(map[string]model.Device),
		tenants:         make(map[string]model.Tenant),
		policyEvents:    make(map[string][]model.RiskEvent),
		anomalyEvents:   make(map[string][]model.RiskEvent),
		tamperEvents:    make(map[string][]model.RiskEvent),
		alertDeliveries: make(map[string][]model.AlertDelivery),
		healthScores:    make(map[string]model.DeviceHealth),
	}
}

func NewPersistent(path string) (*Memory, error) {
	memory := NewMemory()
	memory.path = strings.TrimSpace(path)
	if memory.path == "" {
		return memory, nil
	}
	if err := memory.load(); err != nil {
		return nil, err
	}
	return memory, nil
}

func (m *Memory) EnrollDevice(_ context.Context, req model.EnrollDeviceRequest) (model.Device, error) {
	now := time.Now().UTC()
	device := model.Device{
		TenantID:   strings.TrimSpace(req.TenantID),
		DeviceID:   strings.TrimSpace(req.DeviceID),
		HostName:   strings.TrimSpace(req.HostName),
		Profile:    strings.TrimSpace(req.Profile),
		OSName:     strings.TrimSpace(req.OSName),
		EnrolledAt: now,
		LastSeenAt: now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if current, ok := m.devices[device.DeviceID]; ok {
		device.EnrolledAt = current.EnrolledAt
	}
	m.devices[device.DeviceID] = device
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return model.Device{}, err
	}
	return device, nil
}

func (m *Memory) CreateTenant(_ context.Context, req model.CreateTenantRequest) (model.Tenant, error) {
	now := time.Now().UTC()
	tenantID := strings.TrimSpace(req.TenantID)
	tenant := model.Tenant{
		TenantID:        tenantID,
		Name:            strings.TrimSpace(req.Name),
		PlanID:          strings.TrimSpace(req.PlanID),
		RetentionTierID: strings.TrimSpace(req.RetentionTierID),
		PrimaryProfile:  strings.TrimSpace(req.PrimaryProfile),
		DeviceLimit:     planDeviceLimit(req.PlanID),
		Status:          constants.TenantStatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	if current, ok := m.tenants[tenantID]; ok {
		tenant.CreatedAt = current.CreatedAt
	}
	m.tenants[tenantID] = tenant
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryTenant,
		Action:    constants.AuditActionTenantCreated,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "tenant readiness profile created",
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.Tenant{}, err
	}
	return tenant, nil
}

func (m *Memory) ListTenants(_ context.Context) []model.Tenant {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenants := make([]model.Tenant, 0, len(m.tenants))
	for _, tenant := range m.tenants {
		tenants = append(tenants, tenant)
	}
	sort.Slice(tenants, func(i, j int) bool {
		return tenants[i].TenantID < tenants[j].TenantID
	})
	return tenants
}

func (m *Memory) GetTenant(_ context.Context, tenantID string) (model.Tenant, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, ok := m.tenants[strings.TrimSpace(tenantID)]
	if !ok {
		return model.Tenant{}, ErrTenantNotFound
	}
	return tenant, nil
}

func (m *Memory) ListAuditEvents(_ context.Context, tenantID string) []model.AuditEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantID = strings.TrimSpace(tenantID)
	events := make([]model.AuditEvent, 0, len(m.auditEvents))
	for _, event := range m.auditEvents {
		if tenantID == "" || event.TenantID == tenantID {
			events = append(events, event)
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].CreatedAt.Before(events[j].CreatedAt)
	})
	return events
}

func (m *Memory) ListDevices(_ context.Context) []model.Device {
	m.mu.RLock()
	defer m.mu.RUnlock()

	devices := make([]model.Device, 0, len(m.devices))
	for _, device := range m.devices {
		devices = append(devices, device)
	}
	sort.Slice(devices, func(i, j int) bool {
		return devices[i].DeviceID < devices[j].DeviceID
	})
	return devices
}

func (m *Memory) GetDevice(_ context.Context, deviceID string) (model.Device, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return model.Device{}, ErrDeviceNotFound
	}
	return device, nil
}

func (m *Memory) DailySummary(ctx context.Context, deviceID string, date string) (model.DeviceSummary, error) {
	device, err := m.GetDevice(ctx, deviceID)
	if err != nil {
		return model.DeviceSummary{}, err
	}
	if strings.TrimSpace(date) == "" {
		date = time.Now().UTC().Format(time.DateOnly)
	}
	overview, err := m.HostOverview(ctx, device.DeviceID)
	if err != nil {
		return model.DeviceSummary{}, err
	}
	overview.Summary.Date = date
	return overview.Summary, nil
}

func (m *Memory) HostOverview(_ context.Context, deviceID string) (model.HostOverview, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return model.HostOverview{}, ErrDeviceNotFound
	}
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return model.HostOverview{}, err
	}
	return m.hostOverviewLocked(device), nil
}

func (m *Memory) ListPolicyViolations(_ context.Context, deviceID string) ([]model.RiskEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return nil, err
	}
	return cloneRiskEvents(m.policyEvents[device.DeviceID]), nil
}

func (m *Memory) DeviceHealth(_ context.Context, deviceID string) (model.DeviceHealth, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return model.DeviceHealth{}, ErrDeviceNotFound
	}
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return model.DeviceHealth{}, err
	}
	return m.healthScores[device.DeviceID], nil
}

func (m *Memory) ListAnomalies(_ context.Context, deviceID string) ([]model.RiskEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return nil, err
	}
	return cloneRiskEvents(m.anomalyEvents[device.DeviceID]), nil
}

func (m *Memory) ListTamperEvents(_ context.Context, deviceID string) ([]model.RiskEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return nil, err
	}
	return cloneRiskEvents(m.tamperEvents[device.DeviceID]), nil
}

func (m *Memory) ListAlertDeliveries(_ context.Context, deviceID string) ([]model.AlertDelivery, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return nil, ErrDeviceNotFound
	}
	m.seedDashboardForDeviceLocked(device)
	if err := m.persistLocked(); err != nil {
		return nil, err
	}
	return cloneAlertDeliveries(m.alertDeliveries[device.DeviceID]), nil
}

func (m *Memory) hostOverviewLocked(device model.Device) model.HostOverview {
	policyEvents := cloneRiskEvents(m.policyEvents[device.DeviceID])
	anomalies := cloneRiskEvents(m.anomalyEvents[device.DeviceID])
	tamperEvents := cloneRiskEvents(m.tamperEvents[device.DeviceID])
	deliveries := cloneAlertDeliveries(m.alertDeliveries[device.DeviceID])
	generatedAt := time.Now().UTC()
	summary := model.DeviceSummary{
		DeviceID:            device.DeviceID,
		Date:                generatedAt.Format(time.DateOnly),
		StudyMinutes:        246,
		CodingMinutes:       118,
		EntertainmentMins:   42,
		PolicyViolations:    len(policyEvents),
		ComplianceScore:     82,
		ArchiveBacklog:      2,
		AlertsRaised:        len(deliveries),
		DataCompletenessPct: 93,
	}

	return model.HostOverview{
		Device:    device,
		Summary:   summary,
		RiskLevel: constants.RiskLevelMedium,
		RiskScore: 64,
		Health:    m.healthScores[device.DeviceID],
		Archive: model.ArchiveStatus{
			Status:          constants.StatusPending,
			Provider:        constants.ArchiveProviderS3,
			PendingBatches:  summary.ArchiveBacklog,
			LastUploadedKey: archiveKey(device, generatedAt.Add(-1*time.Hour)),
		},
		PolicyViolations: policyEvents,
		Anomalies:        anomalies,
		TamperEvents:     tamperEvents,
		AlertDeliveries:  deliveries,
		GeneratedAt:      generatedAt,
	}
}

func (m *Memory) seedDashboardForDeviceLocked(device model.Device) {
	_, policyOK := m.policyEvents[device.DeviceID]
	_, healthOK := m.healthScores[device.DeviceID]
	_, anomaliesOK := m.anomalyEvents[device.DeviceID]
	_, tamperOK := m.tamperEvents[device.DeviceID]
	_, deliveriesOK := m.alertDeliveries[device.DeviceID]
	if policyOK && anomaliesOK && tamperOK && deliveriesOK && healthOK {
		return
	}

	base := device.LastSeenAt
	if base.IsZero() {
		base = time.Now().UTC()
	}

	policyMediaID := riskID(device.DeviceID, constants.RiskTypePolicyViolation, 1)
	policyYouTubeID := riskID(device.DeviceID, constants.RiskTypePolicyViolation, 2)
	anomalySoftwareID := riskID(device.DeviceID, constants.RiskTypeAnomaly, 1)
	anomalyProductivityID := riskID(device.DeviceID, constants.RiskTypeAnomaly, 2)
	tamperBacklogID := riskID(device.DeviceID, constants.RiskTypeTamper, 1)

	if !policyOK {
		m.policyEvents[device.DeviceID] = []model.RiskEvent{
			{
				ID:             policyMediaID,
				DeviceID:       device.DeviceID,
				Type:           constants.RiskTypePolicyViolation,
				Severity:       constants.SeverityHigh,
				Category:       constants.RiskCategoryMediaPlayback,
				Source:         constants.RiskSourceProcess,
				AppName:        "VLC media player",
				ResourceLabel:  "sample-movie-file.mkv",
				Reason:         "Entertainment media playback during study policy hours.",
				Recommendation: "Review the usage window and tighten Exam Mode if this repeats.",
				Status:         constants.RiskStatusOpen,
				ObservedAt:     base.Add(-38 * time.Minute),
			},
			{
				ID:             policyYouTubeID,
				DeviceID:       device.DeviceID,
				Type:           constants.RiskTypePolicyViolation,
				Severity:       constants.SeverityMedium,
				Category:       constants.RiskCategoryNonStudyYouTube,
				Source:         constants.RiskSourceBrowser,
				Domain:         "youtube.com",
				ResourceLabel:  "non-study video category",
				Reason:         "YouTube activity was categorized outside coding, math, system design, or coursework.",
				Recommendation: "Suppress study videos automatically and alert only on repeated non-study sessions.",
				Status:         constants.RiskStatusOpen,
				ObservedAt:     base.Add(-64 * time.Minute),
			},
		}
	}

	if !anomaliesOK {
		m.anomalyEvents[device.DeviceID] = []model.RiskEvent{
			{
				ID:             anomalySoftwareID,
				DeviceID:       device.DeviceID,
				Type:           constants.RiskTypeAnomaly,
				Severity:       constants.SeverityMedium,
				Category:       constants.RiskCategoryRiskySoftware,
				Source:         constants.RiskSourceProcess,
				AppName:        "Unknown installer",
				ResourceLabel:  "Downloads installer source",
				Reason:         "New executable activity appeared from a downloads location.",
				Recommendation: "Add signed publisher inventory and approval workflow in the software phase.",
				Status:         constants.RiskStatusAcknowledged,
				ObservedAt:     base.Add(-2 * time.Hour),
			},
			{
				ID:             anomalyProductivityID,
				DeviceID:       device.DeviceID,
				Type:           constants.RiskTypeAnomaly,
				Severity:       constants.SeverityLow,
				Category:       constants.RiskCategoryProductivityShift,
				Source:         constants.RiskSourceAgent,
				ResourceLabel:  "late-night usage pattern",
				Reason:         "Entertainment minutes increased compared with the study baseline.",
				Recommendation: "Use weekly AI report thresholds before escalating low severity drift.",
				Status:         constants.RiskStatusOpen,
				ObservedAt:     base.Add(-4 * time.Hour),
			},
		}
	}

	if !tamperOK {
		m.tamperEvents[device.DeviceID] = []model.RiskEvent{
			{
				ID:             tamperBacklogID,
				DeviceID:       device.DeviceID,
				Type:           constants.RiskTypeTamper,
				Severity:       constants.SeverityLow,
				Category:       constants.RiskCategoryArchiveHealth,
				Source:         constants.RiskSourceArchive,
				ResourceLabel:  "S3 upload backlog",
				Reason:         "Two archive batches are waiting for the next online upload window.",
				Recommendation: "Keep retry visible; alert only if backlog age crosses policy threshold.",
				Status:         constants.RiskStatusOpen,
				ObservedAt:     base.Add(-21 * time.Minute),
			},
		}
	}

	retryAt := base.Add(12 * time.Minute)
	if !deliveriesOK {
		m.alertDeliveries[device.DeviceID] = []model.AlertDelivery{
			{
				ID:            deliveryID(device.DeviceID, constants.DeliveryChannelEmail, 1),
				DeviceID:      device.DeviceID,
				EventID:       policyMediaID,
				Channel:       constants.DeliveryChannelEmail,
				Recipient:     "varathu09@gmail.com",
				Provider:      constants.DeliveryProviderSMTP,
				Status:        constants.DeliveryStatusDelivered,
				Attempts:      1,
				LastAttemptAt: base.Add(-36 * time.Minute),
				Summary:       "High severity media playback alert delivered by email.",
			},
			{
				ID:            deliveryID(device.DeviceID, constants.DeliveryChannelPush, 1),
				DeviceID:      device.DeviceID,
				EventID:       policyYouTubeID,
				Channel:       constants.DeliveryChannelPush,
				Recipient:     "parent mobile push subscription",
				Provider:      constants.DeliveryProviderWebPush,
				Status:        constants.DeliveryStatusRetrying,
				Attempts:      2,
				LastAttemptAt: base.Add(-5 * time.Minute),
				NextRetryAt:   &retryAt,
				LastError:     "push endpoint unavailable during demo retry window",
				Summary:       "Non-study YouTube push alert is retrying.",
			},
			{
				ID:            deliveryID(device.DeviceID, constants.DeliveryChannelDashboard, 1),
				DeviceID:      device.DeviceID,
				EventID:       tamperBacklogID,
				Channel:       constants.DeliveryChannelDashboard,
				Recipient:     "local dashboard",
				Provider:      constants.DeliveryProviderLocalFeed,
				Status:        constants.DeliveryStatusDelivered,
				Attempts:      1,
				LastAttemptAt: base.Add(-20 * time.Minute),
				Summary:       "Archive backlog trust event is visible in dashboard.",
			},
		}
	}

	if !healthOK {
		m.healthScores[device.DeviceID] = model.DeviceHealth{
			DeviceID:             device.DeviceID,
			Score:                78,
			Status:               constants.HealthStatusWatch,
			CPUPercent:           38.5,
			MemoryPercent:        64.2,
			DiskPercent:          71.8,
			BatteryStatus:        "charging",
			BatteryPercent:       86,
			StartupApps:          11,
			AppCrashes24h:        1,
			AgentHealthy:         true,
			AgentLastHeartbeatAt: base.Add(-3 * time.Minute),
			ObservedAt:           base,
			Recommendation:       "Review startup apps and disk usage if the score stays below 80.",
		}
	}
}

func cloneRiskEvents(events []model.RiskEvent) []model.RiskEvent {
	cloned := append([]model.RiskEvent(nil), events...)
	sort.Slice(cloned, func(i, j int) bool {
		return cloned[i].ObservedAt.After(cloned[j].ObservedAt)
	})
	return cloned
}

func cloneAlertDeliveries(deliveries []model.AlertDelivery) []model.AlertDelivery {
	cloned := append([]model.AlertDelivery(nil), deliveries...)
	sort.Slice(cloned, func(i, j int) bool {
		return cloned[i].LastAttemptAt.After(cloned[j].LastAttemptAt)
	})
	return cloned
}

func riskID(deviceID string, riskType string, sequence int) string {
	return strings.Join([]string{strings.TrimSpace(deviceID), riskType, fmt.Sprintf("%03d", sequence)}, "-")
}

func deliveryID(deviceID string, channel string, sequence int) string {
	return strings.Join([]string{strings.TrimSpace(deviceID), channel, "delivery", fmt.Sprintf("%03d", sequence)}, "-")
}

func archiveKey(device model.Device, uploadedAt time.Time) string {
	parts := []string{
		"tenant=" + strings.TrimSpace(device.TenantID),
		"device=" + strings.TrimSpace(device.DeviceID),
		"date=" + uploadedAt.Format(time.DateOnly),
		"hour=" + uploadedAt.Format("15"),
		"batch.json.gz",
	}
	return strings.Join(parts, "/")
}

type persistentState struct {
	Version         string                           `json:"version"`
	Tenants         map[string]model.Tenant          `json:"tenants"`
	Devices         map[string]model.Device          `json:"devices"`
	AuditEvents     []model.AuditEvent               `json:"audit_events"`
	PolicyEvents    map[string][]model.RiskEvent     `json:"policy_events"`
	AnomalyEvents   map[string][]model.RiskEvent     `json:"anomaly_events"`
	TamperEvents    map[string][]model.RiskEvent     `json:"tamper_events"`
	AlertDeliveries map[string][]model.AlertDelivery `json:"alert_deliveries"`
	HealthScores    map[string]model.DeviceHealth    `json:"health_scores"`
}

func (m *Memory) load() error {
	if strings.TrimSpace(m.path) == "" {
		return nil
	}
	data, err := os.ReadFile(m.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read backend state: %w", err)
	}

	var state persistentState
	if err := json.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("decode backend state: %w", err)
	}

	m.tenants = cloneTenantMap(state.Tenants)
	m.devices = cloneDeviceMap(state.Devices)
	m.auditEvents = append([]model.AuditEvent(nil), state.AuditEvents...)
	m.policyEvents = cloneRiskMap(state.PolicyEvents)
	m.anomalyEvents = cloneRiskMap(state.AnomalyEvents)
	m.tamperEvents = cloneRiskMap(state.TamperEvents)
	m.alertDeliveries = cloneDeliveryMap(state.AlertDeliveries)
	m.healthScores = cloneHealthMap(state.HealthScores)
	return nil
}

func (m *Memory) persistLocked() error {
	if strings.TrimSpace(m.path) == "" {
		return nil
	}
	state := persistentState{
		Version:         constants.BackendVersion,
		Tenants:         cloneTenantMap(m.tenants),
		Devices:         cloneDeviceMap(m.devices),
		AuditEvents:     append([]model.AuditEvent(nil), m.auditEvents...),
		PolicyEvents:    cloneRiskMap(m.policyEvents),
		AnomalyEvents:   cloneRiskMap(m.anomalyEvents),
		TamperEvents:    cloneRiskMap(m.tamperEvents),
		AlertDeliveries: cloneDeliveryMap(m.alertDeliveries),
		HealthScores:    cloneHealthMap(m.healthScores),
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("encode backend state: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(m.path), 0o750); err != nil {
		return fmt.Errorf("create backend state dir: %w", err)
	}
	tmpPath := m.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return fmt.Errorf("write backend state temp: %w", err)
	}
	if err := os.Rename(tmpPath, m.path); err != nil {
		if removeErr := os.Remove(m.path); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
			return fmt.Errorf("replace backend state: %w", removeErr)
		}
		if renameErr := os.Rename(tmpPath, m.path); renameErr != nil {
			return fmt.Errorf("commit backend state: %w", renameErr)
		}
	}
	return nil
}

func cloneTenantMap(input map[string]model.Tenant) map[string]model.Tenant {
	output := make(map[string]model.Tenant, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneDeviceMap(input map[string]model.Device) map[string]model.Device {
	output := make(map[string]model.Device, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneRiskMap(input map[string][]model.RiskEvent) map[string][]model.RiskEvent {
	output := make(map[string][]model.RiskEvent, len(input))
	for key, value := range input {
		output[key] = append([]model.RiskEvent(nil), value...)
	}
	return output
}

func cloneDeliveryMap(input map[string][]model.AlertDelivery) map[string][]model.AlertDelivery {
	output := make(map[string][]model.AlertDelivery, len(input))
	for key, value := range input {
		output[key] = append([]model.AlertDelivery(nil), value...)
	}
	return output
}

func cloneHealthMap(input map[string]model.DeviceHealth) map[string]model.DeviceHealth {
	output := make(map[string]model.DeviceHealth, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func PolicyTemplates() []model.PolicyTemplate {
	return []model.PolicyTemplate{
		{
			ID:          "ai-btech-student",
			Name:        "AI BTech Student",
			Audience:    "family",
			Description: "Study-focused endpoint policy for coding, AI, and coursework devices.",
			Roles:       []string{constants.RoleParent, constants.RoleStudent},
		},
		{
			ID:          "school-laptop",
			Name:        "School Laptop",
			Audience:    "school",
			Description: "Managed learning device policy with role-based admin visibility.",
			Roles:       []string{constants.RoleSchoolAdmin, constants.RoleStudent},
		},
		{
			ID:          "small-business-productivity",
			Name:        "Small Business Productivity",
			Audience:    "business",
			Description: "Productivity and endpoint risk observability for managed workstations.",
			Roles:       []string{constants.RoleBusinessManager},
		},
	}
}

func Plans() []model.Plan {
	return []model.Plan{
		{
			ID:            constants.PlanFree,
			Name:          "Free",
			Audience:      "individual",
			DeviceLimit:   1,
			PriceModel:    "local-only starter",
			CloudArchive:  false,
			WeeklyReports: false,
			Features: []string{
				"one device",
				"local-only basic app usage",
				"starter policy templates",
			},
		},
		{
			ID:                 constants.PlanFamilyPro,
			Name:               "Family Pro",
			Audience:           "family",
			DeviceLimit:        5,
			PriceModel:         "per family",
			CloudArchive:       true,
			WeeklyReports:      true,
			RoleBasedDashboard: true,
			Features: []string{
				"weekly AI reports",
				"S3 archive readiness",
				"parent and student views",
				"policy anomaly alerts",
			},
		},
		{
			ID:                 constants.PlanSchool,
			Name:               "School",
			Audience:           "education",
			DeviceLimit:        500,
			PriceModel:         "per student device",
			CloudArchive:       true,
			WeeklyReports:      true,
			RoleBasedDashboard: true,
			Features: []string{
				"managed policy templates",
				"school admin dashboard",
				"audit history",
				"retention controls",
			},
		},
		{
			ID:                 constants.PlanBusiness,
			Name:               "Business",
			Audience:           "business",
			DeviceLimit:        250,
			PriceModel:         "per endpoint",
			CloudArchive:       true,
			WeeklyReports:      true,
			RoleBasedDashboard: true,
			Features: []string{
				"productivity analytics",
				"risky software detection",
				"manager dashboard",
				"compliance exports",
			},
		},
		{
			ID:                 constants.PlanEnterprise,
			Name:               "Enterprise",
			Audience:           "enterprise",
			DeviceLimit:        0,
			PriceModel:         "custom contract",
			CloudArchive:       true,
			WeeklyReports:      true,
			RoleBasedDashboard: true,
			Features: []string{
				"custom retention",
				"SSO readiness",
				"SIEM export roadmap",
				"custom anomaly rules",
			},
		},
	}
}

func Roles() []model.Role {
	return []model.Role{
		{
			ID:          constants.RoleParent,
			Name:        "Parent",
			Scope:       "family",
			Description: "Can review family device summaries, reports, alerts, and policy templates.",
		},
		{
			ID:          constants.RoleStudent,
			Name:        "Student",
			Scope:       "self",
			Description: "Can view transparent monitoring status and personal productivity summaries.",
		},
		{
			ID:          constants.RoleSchoolAdmin,
			Name:        "School Admin",
			Scope:       "education",
			Description: "Can manage school laptop templates, enrollment, and audit history.",
		},
		{
			ID:          constants.RoleBusinessManager,
			Name:        "Business Manager",
			Scope:       "business",
			Description: "Can review business endpoint productivity, risk, and retention settings.",
		},
	}
}

func RetentionTiers() []model.RetentionTier {
	return []model.RetentionTier{
		{
			ID:                 constants.RetentionLocalOnly,
			Name:               "Local Only",
			LocalTTLDays:       7,
			Description:        "Starter retention for free local-only devices.",
			S3StandardDays:     0,
			S3StandardIAUntil:  0,
			S3ArchiveAfterDays: 0,
		},
		{
			ID:                 constants.RetentionFamilyCloud,
			Name:               "Family Cloud Archive",
			LocalTTLDays:       90,
			S3StandardDays:     90,
			S3StandardIAUntil:  365,
			S3ArchiveAfterDays: 365,
			Description:        "Default family archive lifecycle with 90-day standard storage.",
		},
		{
			ID:                 constants.RetentionSchoolYear,
			Name:               "School Year Archive",
			LocalTTLDays:       90,
			S3StandardDays:     90,
			S3StandardIAUntil:  365,
			S3ArchiveAfterDays: 365,
			ComplianceExport:   true,
			Description:        "School retention with compliance export readiness.",
		},
		{
			ID:                 constants.RetentionBusiness,
			Name:               "Business Compliance",
			LocalTTLDays:       90,
			S3StandardDays:     90,
			S3StandardIAUntil:  365,
			S3ArchiveAfterDays: 365,
			ComplianceExport:   true,
			Description:        "Business retention tier for audit and compliance packaging.",
		},
	}
}

func ArchiveStatus() model.ArchiveStatus {
	return model.ArchiveStatus{
		Status:         constants.StatusEmpty,
		Provider:       constants.ArchiveProviderS3,
		PendingBatches: 0,
	}
}

func KnownPlanID(planID string) bool {
	for _, plan := range Plans() {
		if plan.ID == strings.TrimSpace(planID) {
			return true
		}
	}
	return false
}

func KnownRetentionTierID(tierID string) bool {
	for _, tier := range RetentionTiers() {
		if tier.ID == strings.TrimSpace(tierID) {
			return true
		}
	}
	return false
}

func planDeviceLimit(planID string) int {
	for _, plan := range Plans() {
		if plan.ID == strings.TrimSpace(planID) {
			return plan.DeviceLimit
		}
	}
	return 0
}

func auditID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		createdAt.Format("20060102T150405Z"),
		fmt.Sprintf("%04d", sequence),
	}, "-")
}
