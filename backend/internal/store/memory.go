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
	alertRules      map[string][]model.AlertRule
	deviceGroups    map[string][]model.DeviceGroup
	policyAssigns   map[string][]model.PolicyAssignment
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
		alertRules:      make(map[string][]model.AlertRule),
		deviceGroups:    make(map[string][]model.DeviceGroup),
		policyAssigns:   make(map[string][]model.PolicyAssignment),
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
	m.seedAlertRulesForTenantLocked(tenant)
	m.seedDeviceGroupsForTenantLocked(tenant)
	m.seedPolicyAssignmentsForTenantLocked(tenant)
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

func (m *Memory) CreateAlertRule(_ context.Context, tenantID string, req model.CreateAlertRuleRequest) (model.AlertRule, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[tenantID]; !ok {
		return model.AlertRule{}, ErrTenantNotFound
	}
	rule := model.AlertRule{
		ID:         alertRuleID(tenantID, len(m.alertRules[tenantID])+1, now),
		TenantID:   tenantID,
		TemplateID: strings.TrimSpace(req.TemplateID),
		Name:       strings.TrimSpace(req.Name),
		Trigger:    strings.TrimSpace(req.Trigger),
		Severity:   strings.TrimSpace(req.Severity),
		Channels:   normalizeStrings(req.Channels),
		Condition: model.AlertRuleCondition{
			Subject:       strings.TrimSpace(req.Condition.Subject),
			Operator:      strings.TrimSpace(req.Condition.Operator),
			Value:         strings.TrimSpace(req.Condition.Value),
			WindowMinutes: req.Condition.WindowMinutes,
			Threshold:     req.Condition.Threshold,
		},
		Enabled:   req.Enabled,
		CreatedAt: now,
		UpdatedAt: now,
	}
	m.alertRules[tenantID] = append(m.alertRules[tenantID], rule)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryPolicy,
		Action:    constants.AuditActionAlertRuleCreated,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleParent,
		Summary:   "alert rule created: " + rule.Name,
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.AlertRule{}, err
	}
	return rule, nil
}

func (m *Memory) ListAlertRules(_ context.Context, tenantID string) []model.AlertRule {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantID = strings.TrimSpace(tenantID)
	rules := make([]model.AlertRule, 0)
	if tenantID == "" {
		for _, tenantRules := range m.alertRules {
			rules = append(rules, tenantRules...)
		}
	} else {
		rules = append(rules, m.alertRules[tenantID]...)
	}
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].TenantID == rules[j].TenantID {
			return rules[i].CreatedAt.Before(rules[j].CreatedAt)
		}
		return rules[i].TenantID < rules[j].TenantID
	})
	return rules
}

func (m *Memory) CreateDeviceGroup(_ context.Context, tenantID string, req model.CreateDeviceGroupRequest) (model.DeviceGroup, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[tenantID]; !ok {
		return model.DeviceGroup{}, ErrTenantNotFound
	}
	group := model.DeviceGroup{
		ID:               deviceGroupID(tenantID, len(m.deviceGroups[tenantID])+1, now),
		TenantID:         tenantID,
		Name:             strings.TrimSpace(req.Name),
		Description:      strings.TrimSpace(req.Description),
		Profile:          strings.TrimSpace(req.Profile),
		DeviceIDs:        normalizeStrings(req.DeviceIDs),
		PolicyTemplateID: strings.TrimSpace(req.PolicyTemplateID),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.deviceGroups[tenantID] = append(m.deviceGroups[tenantID], group)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryPolicy,
		Action:    constants.AuditActionDeviceGroupCreated,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "device group created: " + group.Name,
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.DeviceGroup{}, err
	}
	return group, nil
}

func (m *Memory) ListDeviceGroups(_ context.Context, tenantID string) []model.DeviceGroup {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantID = strings.TrimSpace(tenantID)
	groups := make([]model.DeviceGroup, 0)
	if tenantID == "" {
		for _, tenantGroups := range m.deviceGroups {
			groups = append(groups, tenantGroups...)
		}
	} else {
		groups = append(groups, m.deviceGroups[tenantID]...)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].TenantID == groups[j].TenantID {
			return groups[i].CreatedAt.Before(groups[j].CreatedAt)
		}
		return groups[i].TenantID < groups[j].TenantID
	})
	return cloneDeviceGroups(groups)
}

func (m *Memory) CreatePolicyAssignment(_ context.Context, tenantID string, req model.CreatePolicyAssignmentRequest) (model.PolicyAssignment, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[tenantID]; !ok {
		return model.PolicyAssignment{}, ErrTenantNotFound
	}
	assignment := model.PolicyAssignment{
		ID:               policyAssignmentID(tenantID, len(m.policyAssigns[tenantID])+1, now),
		TenantID:         tenantID,
		Name:             strings.TrimSpace(req.Name),
		TargetType:       strings.TrimSpace(req.TargetType),
		TargetID:         strings.TrimSpace(req.TargetID),
		PolicyTemplateID: strings.TrimSpace(req.PolicyTemplateID),
		AlertRuleIDs:     normalizeStrings(req.AlertRuleIDs),
		Mode:             strings.TrimSpace(req.Mode),
		Status:           constants.PolicyAssignmentStatusActive,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	m.policyAssigns[tenantID] = append(m.policyAssigns[tenantID], assignment)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryPolicy,
		Action:    constants.AuditActionPolicyAssigned,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleSchoolAdmin,
		Summary:   "policy assigned: " + assignment.Name,
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.PolicyAssignment{}, err
	}
	return assignment, nil
}

func (m *Memory) ListPolicyAssignments(_ context.Context, tenantID string) []model.PolicyAssignment {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantID = strings.TrimSpace(tenantID)
	assignments := make([]model.PolicyAssignment, 0)
	if tenantID == "" {
		for _, tenantAssignments := range m.policyAssigns {
			assignments = append(assignments, tenantAssignments...)
		}
	} else {
		assignments = append(assignments, m.policyAssigns[tenantID]...)
	}
	sort.Slice(assignments, func(i, j int) bool {
		if assignments[i].TenantID == assignments[j].TenantID {
			return assignments[i].CreatedAt.Before(assignments[j].CreatedAt)
		}
		return assignments[i].TenantID < assignments[j].TenantID
	})
	return clonePolicyAssignments(assignments)
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

func (m *Memory) seedAlertRulesForTenantLocked(tenant model.Tenant) {
	tenantID := strings.TrimSpace(tenant.TenantID)
	if tenantID == "" || len(m.alertRules[tenantID]) > 0 {
		return
	}
	now := time.Now().UTC()
	m.alertRules[tenantID] = []model.AlertRule{
		{
			ID:         alertRuleID(tenantID, 1, now),
			TenantID:   tenantID,
			TemplateID: constants.AlertRuleTemplateMediaAfterHours,
			Name:       "Alert on VLC or media playback after 10 PM",
			Trigger:    constants.AlertTriggerMediaPlayback,
			Severity:   constants.SeverityHigh,
			Channels:   []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Condition: model.AlertRuleCondition{
				Subject:  constants.AlertConditionSubjectApp,
				Operator: constants.AlertConditionOperatorAfterLocal,
				Value:    "22:00",
			},
			Enabled:   true,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:         alertRuleID(tenantID, 2, now.Add(time.Millisecond)),
			TenantID:   tenantID,
			TemplateID: constants.AlertRuleTemplateNonStudyYouTube,
			Name:       "Alert when non-study YouTube crosses 30 minutes",
			Trigger:    constants.AlertTriggerNonStudyYouTube,
			Severity:   constants.SeverityMedium,
			Channels:   []string{constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Condition: model.AlertRuleCondition{
				Subject:       constants.AlertConditionSubjectUsageMinutes,
				Operator:      constants.AlertConditionOperatorGreaterThan,
				Value:         "30",
				WindowMinutes: 60,
				Threshold:     30,
			},
			Enabled:   true,
			CreatedAt: now.Add(time.Millisecond),
			UpdatedAt: now.Add(time.Millisecond),
		},
	}
}

func (m *Memory) seedDeviceGroupsForTenantLocked(tenant model.Tenant) {
	tenantID := strings.TrimSpace(tenant.TenantID)
	if tenantID == "" || len(m.deviceGroups[tenantID]) > 0 {
		return
	}
	now := time.Now().UTC()
	profile := strings.TrimSpace(tenant.PrimaryProfile)
	if profile == "" {
		profile = "ai-btech-student"
	}
	m.deviceGroups[tenantID] = []model.DeviceGroup{
		{
			ID:               deviceGroupID(tenantID, 1, now),
			TenantID:         tenantID,
			Name:             "Primary study devices",
			Description:      "Default group for the tenant primary policy profile.",
			Profile:          profile,
			DeviceIDs:        []string{},
			PolicyTemplateID: profile,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
}

func (m *Memory) seedPolicyAssignmentsForTenantLocked(tenant model.Tenant) {
	tenantID := strings.TrimSpace(tenant.TenantID)
	if tenantID == "" || len(m.policyAssigns[tenantID]) > 0 {
		return
	}
	now := time.Now().UTC()
	profile := strings.TrimSpace(tenant.PrimaryProfile)
	if profile == "" {
		profile = "ai-btech-student"
	}
	groupID := ""
	if groups := m.deviceGroups[tenantID]; len(groups) > 0 {
		groupID = groups[0].ID
	}
	m.policyAssigns[tenantID] = []model.PolicyAssignment{
		{
			ID:               policyAssignmentID(tenantID, 1, now),
			TenantID:         tenantID,
			Name:             "Primary profile assignment",
			TargetType:       constants.PolicyAssignmentTargetDeviceGroup,
			TargetID:         groupID,
			PolicyTemplateID: profile,
			AlertRuleIDs:     alertRuleIDs(m.alertRules[tenantID]),
			Mode:             constants.PolicyAssignmentModeActive,
			Status:           constants.PolicyAssignmentStatusActive,
			CreatedAt:        now,
			UpdatedAt:        now,
		},
	}
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

func alertRuleID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"alert-rule",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func deviceGroupID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"device-group",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func policyAssignmentID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"policy-assignment",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func alertRuleIDs(rules []model.AlertRule) []string {
	ids := make([]string, 0, len(rules))
	for _, rule := range rules {
		if clean := strings.TrimSpace(rule.ID); clean != "" {
			ids = append(ids, clean)
		}
	}
	return ids
}

func normalizeStrings(values []string) []string {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" || seen[clean] {
			continue
		}
		seen[clean] = true
		normalized = append(normalized, clean)
	}
	return normalized
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
	Version         string                              `json:"version"`
	Tenants         map[string]model.Tenant             `json:"tenants"`
	Devices         map[string]model.Device             `json:"devices"`
	AuditEvents     []model.AuditEvent                  `json:"audit_events"`
	AlertRules      map[string][]model.AlertRule        `json:"alert_rules"`
	DeviceGroups    map[string][]model.DeviceGroup      `json:"device_groups"`
	PolicyAssigns   map[string][]model.PolicyAssignment `json:"policy_assignments"`
	PolicyEvents    map[string][]model.RiskEvent        `json:"policy_events"`
	AnomalyEvents   map[string][]model.RiskEvent        `json:"anomaly_events"`
	TamperEvents    map[string][]model.RiskEvent        `json:"tamper_events"`
	AlertDeliveries map[string][]model.AlertDelivery    `json:"alert_deliveries"`
	HealthScores    map[string]model.DeviceHealth       `json:"health_scores"`
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
	m.alertRules = cloneAlertRuleMap(state.AlertRules)
	m.deviceGroups = cloneDeviceGroupMap(state.DeviceGroups)
	m.policyAssigns = clonePolicyAssignmentMap(state.PolicyAssigns)
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
		AlertRules:      cloneAlertRuleMap(m.alertRules),
		DeviceGroups:    cloneDeviceGroupMap(m.deviceGroups),
		PolicyAssigns:   clonePolicyAssignmentMap(m.policyAssigns),
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

func cloneAlertRuleMap(input map[string][]model.AlertRule) map[string][]model.AlertRule {
	output := make(map[string][]model.AlertRule, len(input))
	for key, value := range input {
		output[key] = append([]model.AlertRule(nil), value...)
	}
	return output
}

func cloneDeviceGroups(input []model.DeviceGroup) []model.DeviceGroup {
	output := append([]model.DeviceGroup(nil), input...)
	for index := range output {
		output[index].DeviceIDs = append([]string(nil), output[index].DeviceIDs...)
	}
	return output
}

func cloneDeviceGroupMap(input map[string][]model.DeviceGroup) map[string][]model.DeviceGroup {
	output := make(map[string][]model.DeviceGroup, len(input))
	for key, value := range input {
		output[key] = cloneDeviceGroups(value)
	}
	return output
}

func clonePolicyAssignments(input []model.PolicyAssignment) []model.PolicyAssignment {
	output := append([]model.PolicyAssignment(nil), input...)
	for index := range output {
		output[index].AlertRuleIDs = append([]string(nil), output[index].AlertRuleIDs...)
	}
	return output
}

func clonePolicyAssignmentMap(input map[string][]model.PolicyAssignment) map[string][]model.PolicyAssignment {
	output := make(map[string][]model.PolicyAssignment, len(input))
	for key, value := range input {
		output[key] = clonePolicyAssignments(value)
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

func AlertRuleTemplates() []model.AlertRuleTemplate {
	return []model.AlertRuleTemplate{
		{
			ID:              constants.AlertRuleTemplateNonStudyYouTube,
			Name:            "Non-study YouTube over limit",
			Trigger:         constants.AlertTriggerNonStudyYouTube,
			Description:     "Alert when YouTube usage is not categorized as coding, math, system design, or coursework and crosses a time threshold.",
			DefaultSeverity: constants.SeverityMedium,
			Channels:        []string{constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Example:         "If non-study YouTube is greater than 30 minutes in 60 minutes, send push and dashboard alert.",
			PaidTier:        constants.PlanFamilyPro,
		},
		{
			ID:              constants.AlertRuleTemplateMediaAfterHours,
			Name:            "Media playback after hours",
			Trigger:         constants.AlertTriggerMediaPlayback,
			Description:     "Alert when VLC, media player, or other entertainment playback appears during restricted hours.",
			DefaultSeverity: constants.SeverityHigh,
			Channels:        []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Example:         "If media playback starts after 10 PM, email the parent and add it to the dashboard feed.",
			PaidTier:        constants.PlanFamilyPro,
		},
		{
			ID:              constants.AlertRuleTemplateRiskySoftware,
			Name:            "Risky software detected",
			Trigger:         constants.AlertTriggerRiskySoftware,
			Description:     "Alert when torrent clients, VPN/proxy tools, game launchers, unknown browsers, or downloads installers are detected.",
			DefaultSeverity: constants.SeverityHigh,
			Channels:        []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Example:         "If risky software category equals torrent client, email and record the dashboard event.",
			PaidTier:        constants.PlanBusiness,
		},
		{
			ID:              constants.AlertRuleTemplateTamperBacklog,
			Name:            "Archive backlog over limit",
			Trigger:         constants.AlertTriggerArchiveBacklog,
			Description:     "Alert when S3 archive upload backlog waits beyond the configured online retry window.",
			DefaultSeverity: constants.SeverityMedium,
			Channels:        []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Example:         "If archive backlog is greater than 2 batches for 60 minutes, email and show a trust event.",
			PaidTier:        constants.PlanSchool,
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

func KnownAlertRuleTemplateID(templateID string) bool {
	for _, template := range AlertRuleTemplates() {
		if template.ID == strings.TrimSpace(templateID) {
			return true
		}
	}
	return false
}

func KnownPolicyTemplateID(templateID string) bool {
	for _, template := range PolicyTemplates() {
		if template.ID == strings.TrimSpace(templateID) {
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
