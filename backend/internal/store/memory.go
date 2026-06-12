package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
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
	mu                 sync.RWMutex
	path               string
	devices            map[string]model.Device
	tenants            map[string]model.Tenant
	auditEvents        []model.AuditEvent
	alertRules         map[string][]model.AlertRule
	notificationRoutes map[string][]model.NotificationRoute
	notificationPrefs  map[string]model.NotificationPreferenceCenter
	deliveryRemedies   map[string][]model.TenantDeliveryRemediationAction
	activityViews      map[string][]model.TenantActivityView
	dataExports        map[string][]model.TenantDataExport
	deleteRequests     map[string][]model.DeleteRequest
	deviceGroups       map[string][]model.DeviceGroup
	policyAssigns      map[string][]model.PolicyAssignment
	policyEvents       map[string][]model.RiskEvent
	anomalyEvents      map[string][]model.RiskEvent
	tamperEvents       map[string][]model.RiskEvent
	alertDeliveries    map[string][]model.AlertDelivery
	healthScores       map[string]model.DeviceHealth
	telemetryEvents    map[string][]model.TelemetryEvent
}

func NewMemory() *Memory {
	return &Memory{
		devices:            make(map[string]model.Device),
		tenants:            make(map[string]model.Tenant),
		alertRules:         make(map[string][]model.AlertRule),
		notificationRoutes: make(map[string][]model.NotificationRoute),
		notificationPrefs:  make(map[string]model.NotificationPreferenceCenter),
		deliveryRemedies:   make(map[string][]model.TenantDeliveryRemediationAction),
		activityViews:      make(map[string][]model.TenantActivityView),
		dataExports:        make(map[string][]model.TenantDataExport),
		deleteRequests:     make(map[string][]model.DeleteRequest),
		deviceGroups:       make(map[string][]model.DeviceGroup),
		policyAssigns:      make(map[string][]model.PolicyAssignment),
		policyEvents:       make(map[string][]model.RiskEvent),
		anomalyEvents:      make(map[string][]model.RiskEvent),
		tamperEvents:       make(map[string][]model.RiskEvent),
		alertDeliveries:    make(map[string][]model.AlertDelivery),
		healthScores:       make(map[string]model.DeviceHealth),
		telemetryEvents:    make(map[string][]model.TelemetryEvent),
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
	m.seedNotificationRoutesForTenantLocked(tenant)
	m.seedNotificationPreferencesForTenantLocked(tenant)
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

func (m *Memory) CreateNotificationRoute(_ context.Context, tenantID string, req model.CreateNotificationRouteRequest) (model.NotificationRoute, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.NotificationRoute{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)

	status := strings.TrimSpace(req.Status)
	if status == "" {
		status = constants.StatusWatch
	}
	route := model.NotificationRoute{
		ID:             notificationRouteID(tenantID, len(m.notificationRoutes[tenantID])+1, now),
		TenantID:       tenantID,
		Channel:        strings.TrimSpace(req.Channel),
		Provider:       strings.TrimSpace(req.Provider),
		RecipientLabel: strings.TrimSpace(req.RecipientLabel),
		Status:         status,
		Enabled:        req.Enabled,
		LastSummary:    strings.TrimSpace(req.LastSummary),
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if route.Status == constants.StatusHealthy {
		route.LastVerifiedAt = &now
	}
	m.notificationRoutes[tenantID] = append(m.notificationRoutes[tenantID], route)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategorySystem,
		Action:    constants.AuditActionNotificationRoute,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "notification route created: " + route.Channel + "/" + route.Provider,
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.NotificationRoute{}, err
	}
	return route, nil
}

func (m *Memory) ListNotificationRoutes(_ context.Context, tenantID string) []model.NotificationRoute {
	m.mu.Lock()
	defer m.mu.Unlock()

	tenantID = strings.TrimSpace(tenantID)
	routes := make([]model.NotificationRoute, 0)
	if tenantID == "" {
		for currentTenantID := range m.tenants {
			if tenant, ok := m.tenants[currentTenantID]; ok {
				m.seedNotificationRoutesForTenantLocked(tenant)
			}
			routes = append(routes, m.notificationRoutes[currentTenantID]...)
		}
	} else {
		if tenant, ok := m.tenants[tenantID]; ok {
			m.seedNotificationRoutesForTenantLocked(tenant)
		}
		routes = append(routes, m.notificationRoutes[tenantID]...)
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].TenantID == routes[j].TenantID {
			return routes[i].CreatedAt.Before(routes[j].CreatedAt)
		}
		return routes[i].TenantID < routes[j].TenantID
	})
	return append([]model.NotificationRoute(nil), routes...)
}

func (m *Memory) TenantNotificationPreferences(_ context.Context, tenantID string) (model.NotificationPreferenceCenter, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.NotificationPreferenceCenter{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)
	m.seedNotificationPreferencesForTenantLocked(tenant)
	center := buildNotificationPreferenceCenter(tenant, m.notificationPrefs[tenantID], m.notificationRoutes[tenantID], now)
	if err := m.persistLocked(); err != nil {
		return model.NotificationPreferenceCenter{}, err
	}
	return center, nil
}

func (m *Memory) UpdateTenantNotificationPreferences(_ context.Context, tenantID string, req model.UpdateNotificationPreferencesRequest) (model.NotificationPreferenceCenter, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.NotificationPreferenceCenter{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)
	m.seedNotificationPreferencesForTenantLocked(tenant)

	current := m.notificationPrefs[tenantID]
	if strings.TrimSpace(req.DigestCadence) != "" {
		current.DigestCadence = strings.TrimSpace(req.DigestCadence)
	}
	if strings.TrimSpace(req.QuietHours.StartLocal) != "" || strings.TrimSpace(req.QuietHours.EndLocal) != "" || strings.TrimSpace(req.QuietHours.Timezone) != "" {
		current.QuietHours = req.QuietHours
	}
	if req.Escalation.AfterMinutes > 0 || req.Escalation.RepeatEveryMins > 0 || req.Escalation.MaxRepeats > 0 || len(req.Escalation.Channels) > 0 || strings.TrimSpace(req.Escalation.Owner) != "" {
		current.Escalation = req.Escalation
	}
	if len(req.Rules) > 0 {
		current.Rules = normalizeNotificationPreferenceRules(tenantID, req.Rules, now)
	}
	current.UpdatedAt = now
	m.notificationPrefs[tenantID] = current
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategorySystem,
		Action:    constants.AuditActionNotificationPref,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "notification preference center updated",
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.NotificationPreferenceCenter{}, err
	}
	return buildNotificationPreferenceCenter(tenant, current, m.notificationRoutes[tenantID], now), nil
}

func (m *Memory) TenantDeliveryDrilldown(_ context.Context, tenantID string) (model.TenantDeliveryDrilldown, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantDeliveryDrilldown{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)
	return buildTenantDeliveryDrilldown(tenantID, m.notificationRoutes[tenantID], m.deliveriesForTenantLocked(tenantID), now, constants.DeliveryDrillModeDryRun), nil
}

func (m *Memory) RunTenantDeliveryDrilldown(_ context.Context, tenantID string, req model.RunDeliveryDrilldownRequest) (model.TenantDeliveryDrilldown, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)
	channel := strings.TrimSpace(req.Channel)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantDeliveryDrilldown{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)

	rehearsed := 0
	routes := m.notificationRoutes[tenantID]
	for index := range routes {
		if channel != "" && routes[index].Channel != channel {
			continue
		}
		if !routes[index].Enabled {
			routes[index].Status = constants.StatusWatch
			routes[index].LastSummary = "Dry-run rehearsal skipped because the route is disabled."
			routes[index].UpdatedAt = now
			continue
		}
		if !deliveryProviderMatchesChannel(routes[index].Provider, routes[index].Channel) {
			routes[index].Status = constants.StatusAttention
			routes[index].LastSummary = "Dry-run rehearsal detected a provider/channel mismatch."
			routes[index].UpdatedAt = now
			rehearsed++
			continue
		}
		routes[index].Status = constants.StatusHealthy
		routes[index].LastVerifiedAt = &now
		routes[index].LastSummary = "Dry-run rehearsal passed without sending provider payloads or storing message content."
		routes[index].UpdatedAt = now
		rehearsed++
	}
	m.notificationRoutes[tenantID] = routes
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategorySystem,
		Action:    constants.AuditActionDeliveryDrillRun,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   fmt.Sprintf("delivery drilldown dry-run rehearsed %d route(s)", rehearsed),
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.TenantDeliveryDrilldown{}, err
	}
	return buildTenantDeliveryDrilldown(tenantID, m.notificationRoutes[tenantID], m.deliveriesForTenantLocked(tenantID), now, constants.DeliveryDrillModeDryRun), nil
}

func (m *Memory) TenantDeliveryRemediation(_ context.Context, tenantID string) (model.TenantDeliveryRemediation, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantDeliveryRemediation{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)
	return buildTenantDeliveryRemediation(
		tenantID,
		m.notificationRoutes[tenantID],
		m.deliveriesForTenantLocked(tenantID),
		m.deliveryRemedies[tenantID],
		now,
	), nil
}

func (m *Memory) RunTenantDeliveryRemediation(_ context.Context, tenantID string, req model.RunDeliveryRemediationRequest) (model.TenantDeliveryRemediation, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantDeliveryRemediation{}, ErrTenantNotFound
	}
	m.seedNotificationRoutesForTenantLocked(tenant)

	route, delivery := selectDeliveryRemediationRoute(
		m.notificationRoutes[tenantID],
		m.deliveriesForTenantLocked(tenantID),
		strings.TrimSpace(req.RouteID),
		strings.TrimSpace(req.Channel),
	)
	action := deliveryRemediationAction(route, delivery, now)
	action.ID = deliveryRemediationID(tenantID, len(m.deliveryRemedies[tenantID])+1, now)
	action.TenantID = tenantID
	action.Action = firstNonEmpty(strings.TrimSpace(req.Action), action.Action)
	action.Owner = firstNonEmpty(strings.TrimSpace(req.Owner), action.Owner)
	action.Status = deliveryRemediationStatusForAction(action.Action)
	action.Plan = firstNonEmpty(strings.TrimSpace(req.Reason), action.Plan)
	action.AuditState = constants.AuditActionDeliveryRemediation
	action.CreatedAt = now
	m.deliveryRemedies[tenantID] = append(m.deliveryRemedies[tenantID], action)

	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategorySystem,
		Action:    constants.AuditActionDeliveryRemediation,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   fmt.Sprintf("delivery remediation %s planned for %s route", action.Action, action.Channel),
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.TenantDeliveryRemediation{}, err
	}
	return buildTenantDeliveryRemediation(
		tenantID,
		m.notificationRoutes[tenantID],
		m.deliveriesForTenantLocked(tenantID),
		m.deliveryRemedies[tenantID],
		now,
	), nil
}

func (m *Memory) deliveriesForTenantLocked(tenantID string) []model.AlertDelivery {
	deliveries := make([]model.AlertDelivery, 0)
	for _, device := range m.devices {
		if device.TenantID == tenantID {
			deliveries = append(deliveries, m.alertDeliveries[device.DeviceID]...)
		}
	}
	return deliveries
}

func (m *Memory) CreateTenantActivityView(_ context.Context, tenantID string, req model.CreateTenantActivityViewRequest) (model.TenantActivityView, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantActivityView{}, ErrTenantNotFound
	}
	m.seedActivityViewsForTenantLocked(tenant)

	view := model.TenantActivityView{
		ID:          activityViewID(req.ID, tenantID, len(m.activityViews[tenantID])+1),
		TenantID:    tenantID,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
		Filter:      normalizeActivityFeedFilter(req.Filter),
		PaidTier:    fallbackString(req.PaidTier, constants.PlanFamilyPro),
		SortOrder:   req.SortOrder,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if view.SortOrder <= 0 {
		view.SortOrder = len(m.activityViews[tenantID]) + 1
	}
	m.activityViews[tenantID] = append(m.activityViews[tenantID], view)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryTenant,
		Action:    constants.AuditActionActivityViewCreated,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "tenant activity view created",
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.TenantActivityView{}, err
	}
	return view, nil
}

func (m *Memory) ListTenantActivityViews(_ context.Context, tenantID string) []model.TenantActivityView {
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if tenant, ok := m.tenants[tenantID]; ok {
		m.seedActivityViewsForTenantLocked(tenant)
		_ = m.persistLocked()
	}
	views := append([]model.TenantActivityView(nil), m.activityViews[tenantID]...)
	sort.Slice(views, func(i, j int) bool {
		if views[i].SortOrder != views[j].SortOrder {
			return views[i].SortOrder < views[j].SortOrder
		}
		return views[i].Name < views[j].Name
	})
	return views
}

func (m *Memory) TenantOperationsSummary(_ context.Context, tenantID string) (model.TenantOperationsSummary, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantOperationsSummary{}, ErrTenantNotFound
	}

	var allEvents []model.RiskEvent
	var allDeliveries []model.AlertDelivery
	riskTotal := 0
	hostsTotal := 0
	hostsAttention := 0
	openPolicy := 0
	openAnomalies := 0
	tamperSignals := 0
	archiveBacklog := 0

	for _, device := range m.devices {
		if device.TenantID != tenantID {
			continue
		}
		hostsTotal++
		m.seedDashboardForDeviceLocked(device)
		overview := m.hostOverviewLocked(device)
		riskTotal += overview.RiskScore
		openPolicy += len(overview.PolicyViolations)
		openAnomalies += len(overview.Anomalies)
		tamperSignals += len(overview.TamperEvents)
		archiveBacklog += overview.Archive.PendingBatches
		allEvents = append(allEvents, overview.PolicyViolations...)
		allEvents = append(allEvents, overview.Anomalies...)
		allEvents = append(allEvents, overview.TamperEvents...)
		allDeliveries = append(allDeliveries, overview.AlertDeliveries...)
		if overview.RiskScore >= 60 || overview.Health.Status != constants.HealthStatusHealthy || overview.Archive.PendingBatches > 0 {
			hostsAttention++
		}
	}
	if err := m.persistLocked(); err != nil {
		return model.TenantOperationsSummary{}, err
	}

	deliveryTotal := len(allDeliveries)
	delivered := 0
	retrying := 0
	failed := 0
	emailDelivered := 0
	pushDelivered := 0
	dashboardDelivered := 0
	for _, delivery := range allDeliveries {
		switch delivery.Status {
		case constants.DeliveryStatusDelivered:
			delivered++
			switch delivery.Channel {
			case constants.DeliveryChannelEmail:
				emailDelivered++
			case constants.DeliveryChannelPush:
				pushDelivered++
			case constants.DeliveryChannelDashboard:
				dashboardDelivered++
			}
		case constants.DeliveryStatusRetrying:
			retrying++
		case constants.DeliveryStatusFailed:
			failed++
		}
	}

	riskScore := 0
	if hostsTotal > 0 {
		riskScore = riskTotal / hostsTotal
	}
	notificationScore := 0
	if deliveryTotal > 0 {
		notificationScore = (delivered * 100) / deliveryTotal
	}
	plan := planByID(tenant.PlanID)
	monetizationReadiness := tenantReadinessScore(tenant, plan, hostsTotal, emailDelivered, pushDelivered, dashboardDelivered, len(m.alertRules[tenantID]), len(m.deviceGroups[tenantID]), len(m.policyAssigns[tenantID]))
	customerHealth := constants.StatusHealthy
	if failed > 0 || riskScore >= 75 {
		customerHealth = constants.StatusAttention
	} else if retrying > 0 || openAnomalies > 0 || archiveBacklog > 0 || hostsAttention > 0 {
		customerHealth = constants.StatusWatch
	}

	return model.TenantOperationsSummary{
		TenantID:              tenant.TenantID,
		TenantName:            tenant.Name,
		PlanID:                tenant.PlanID,
		PlanName:              plan.Name,
		CustomerHealth:        customerHealth,
		MonetizationReadiness: monetizationReadiness,
		HostsTotal:            hostsTotal,
		HostsAttention:        hostsAttention,
		RiskScore:             riskScore,
		OpenPolicyViolations:  openPolicy,
		OpenAnomalies:         openAnomalies,
		TamperSignals:         tamperSignals,
		ArchiveBacklog:        archiveBacklog,
		NotificationScore:     notificationScore,
		DeliveryTotal:         deliveryTotal,
		DeliveryDelivered:     delivered,
		DeliveryRetrying:      retrying,
		DeliveryFailed:        failed,
		EmailDelivered:        emailDelivered,
		PushDelivered:         pushDelivered,
		DashboardDelivered:    dashboardDelivered,
		LastEmail:             deliverySnapshot(latestTenantDelivery(allDeliveries, constants.DeliveryChannelEmail)),
		LastPush:              deliverySnapshot(latestTenantDelivery(allDeliveries, constants.DeliveryChannelPush)),
		PrioritySignals:       tenantPrioritySignals(allEvents, allDeliveries, now),
		UpgradeSignals:        tenantUpgradeSignals(tenant, plan, monetizationReadiness, now),
		GeneratedAt:           now,
	}, nil
}

func (m *Memory) TenantMonetizationSummary(ctx context.Context, tenantID string) (model.TenantMonetizationSummary, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantMonetizationSummary{}, err
	}

	m.mu.RLock()
	tenant, ok := m.tenants[tenantID]
	if !ok {
		m.mu.RUnlock()
		return model.TenantMonetizationSummary{}, ErrTenantNotFound
	}
	plan := planByID(tenant.PlanID)
	auditCount := 0
	for _, event := range m.auditEvents {
		if event.TenantID == tenantID {
			auditCount++
		}
	}
	rulesCount := len(m.alertRules[tenantID])
	groupsCount := len(m.deviceGroups[tenantID])
	assignmentsCount := len(m.policyAssigns[tenantID])
	exportsCount := len(m.dataExports[tenantID])
	deletesCount := len(m.deleteRequests[tenantID])
	deliveries := make([]model.AlertDelivery, 0)
	for _, device := range m.devices {
		if device.TenantID == tenantID {
			deliveries = append(deliveries, m.alertDeliveries[device.DeviceID]...)
		}
	}
	m.mu.RUnlock()

	email := latestTenantDelivery(deliveries, constants.DeliveryChannelEmail)
	push := latestTenantDelivery(deliveries, constants.DeliveryChannelPush)
	dashboard := latestTenantDelivery(deliveries, constants.DeliveryChannelDashboard)
	trustScore := tenantTrustScore(operations, auditCount, exportsCount, deletesCount)

	return model.TenantMonetizationSummary{
		TenantID:            tenant.TenantID,
		TenantName:          tenant.Name,
		PlanID:              tenant.PlanID,
		PlanName:            plan.Name,
		Audience:            plan.Audience,
		ConversionStage:     monetizationStage(tenant, operations, trustScore),
		RevenueHealth:       monetizationHealth(operations, trustScore),
		SeatsUsed:           operations.HostsTotal,
		SeatsIncluded:       tenant.DeviceLimit,
		ReadinessScore:      operations.MonetizationReadiness,
		NotificationScore:   operations.NotificationScore,
		TrustScore:          trustScore,
		NotificationPromise: notificationPromise(operations, trustScore, email, push, dashboard),
		NotificationRoutes: []model.TenantNotificationRoute{
			notificationRoute(constants.DeliveryChannelEmail, email),
			notificationRoute(constants.DeliveryChannelPush, push),
			notificationRoute(constants.DeliveryChannelDashboard, dashboard),
		},
		ValuePanels:       tenantValuePanels(operations, plan, trustScore),
		PaidCapabilities:  tenantPaidCapabilities(operations, rulesCount, groupsCount, assignmentsCount, auditCount, exportsCount),
		ConversionActions: tenantConversionActions(operations, tenant, plan, trustScore, now),
		GeneratedAt:       now,
	}, nil
}

func (m *Memory) TenantBusinessDashboard(ctx context.Context, tenantID string) (model.TenantBusinessDashboard, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}
	inbox, err := m.TenantAlertInbox(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}
	commandCenter, err := m.TenantNotificationCommandCenter(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}
	preferences, err := m.TenantNotificationPreferences(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}
	drilldown, err := m.TenantDeliveryDrilldown(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}
	remediation, err := m.TenantDeliveryRemediation(ctx, tenantID)
	if err != nil {
		return model.TenantBusinessDashboard{}, err
	}

	return buildTenantBusinessDashboard(operations, monetization, inbox, commandCenter, preferences, drilldown, remediation, generatedAt), nil
}

func (m *Memory) TenantAlertInbox(_ context.Context, tenantID string) (model.TenantAlertInbox, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantAlertInbox{}, ErrTenantNotFound
	}

	events := make([]model.TenantAlertInboxItem, 0)
	sourceHosts := make(map[string]bool)
	for _, device := range m.devices {
		if device.TenantID != tenantID {
			continue
		}
		sourceHosts[device.DeviceID] = true
		m.seedDashboardForDeviceLocked(device)
		deliveriesByEvent := deliveriesByEventID(m.alertDeliveries[device.DeviceID])
		deviceEvents := make([]model.RiskEvent, 0)
		deviceEvents = append(deviceEvents, m.policyEvents[device.DeviceID]...)
		deviceEvents = append(deviceEvents, m.anomalyEvents[device.DeviceID]...)
		deviceEvents = append(deviceEvents, m.tamperEvents[device.DeviceID]...)
		for _, event := range deviceEvents {
			proof := alertDeliveryProof(deliveriesByEvent[event.ID])
			events = append(events, model.TenantAlertInboxItem{
				ID:             strings.Join([]string{tenantID, device.DeviceID, event.ID}, ":"),
				TenantID:       tenantID,
				DeviceID:       device.DeviceID,
				HostName:       device.HostName,
				EventID:        event.ID,
				Type:           event.Type,
				Severity:       event.Severity,
				Category:       event.Category,
				Status:         event.Status,
				Title:          eventTitle(event),
				Detail:         event.Reason,
				Recommendation: event.Recommendation,
				Source:         event.Source,
				DeliveryState:  alertDeliveryState(proof),
				DeliveryProof:  proof,
				NextAction:     alertInboxNextAction(event, proof),
				ObservedAt:     event.ObservedAt,
			})
		}
	}
	if err := m.persistLocked(); err != nil {
		return model.TenantAlertInbox{}, err
	}

	sort.Slice(events, func(i, j int) bool {
		statusDelta := riskStatusRank(events[j].Status) - riskStatusRank(events[i].Status)
		if statusDelta != 0 {
			return statusDelta < 0
		}
		severityDelta := severityRank(events[j].Severity) - severityRank(events[i].Severity)
		if severityDelta != 0 {
			return severityDelta < 0
		}
		return events[i].ObservedAt.After(events[j].ObservedAt)
	})

	return model.TenantAlertInbox{
		TenantID:        tenant.TenantID,
		TenantName:      tenant.Name,
		Summary:         tenantAlertInboxSummary(events, len(sourceHosts)),
		Items:           append([]model.TenantAlertInboxItem(nil), events...),
		GeneratedAt:     now,
		PrivacyBoundary: constants.TelemetryPrivacyBoundary,
	}, nil
}

func (m *Memory) TenantNotificationCommandCenter(ctx context.Context, tenantID string) (model.TenantNotificationCommandCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationCommandCenter{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationCommandCenter{}, err
	}
	inbox, err := m.TenantAlertInbox(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationCommandCenter{}, err
	}
	drilldown, err := m.TenantDeliveryDrilldown(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationCommandCenter{}, err
	}
	remediation, err := m.TenantDeliveryRemediation(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationCommandCenter{}, err
	}

	return buildTenantNotificationCommandCenter(operations, monetization, inbox, drilldown, remediation, generatedAt), nil
}

func (m *Memory) TenantSyncHealth(_ context.Context, tenantID string) (model.TenantSyncHealth, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.RLock()
	defer m.mu.RUnlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantSyncHealth{}, ErrTenantNotFound
	}

	devices := make([]model.DeviceSyncHealth, 0)
	storedEvents := 0
	hostsReporting := 0
	var lastLocalEventID int64
	var lastIngestedAt time.Time
	for _, device := range m.devices {
		if device.TenantID != tenantID {
			continue
		}
		events := cloneTelemetryEvents(m.telemetryEvents[device.DeviceID])
		summary := deviceSyncHealth(device, events)
		if summary.StoredEvents > 0 {
			hostsReporting++
		}
		storedEvents += summary.StoredEvents
		if summary.LastLocalEventID > lastLocalEventID {
			lastLocalEventID = summary.LastLocalEventID
		}
		if summary.LastIngestedAt.After(lastIngestedAt) {
			lastIngestedAt = summary.LastIngestedAt
		}
		devices = append(devices, summary)
	}
	sort.Slice(devices, func(i, j int) bool {
		if devices[i].Status != devices[j].Status {
			return devices[i].Status < devices[j].Status
		}
		return devices[i].HostName < devices[j].HostName
	})
	hostsTotal := len(devices)
	hostsPending := hostsTotal - hostsReporting
	status := constants.StatusHealthy
	if hostsTotal == 0 || hostsReporting == 0 {
		status = constants.StatusPending
	} else if hostsPending > 0 {
		status = constants.StatusWatch
	}
	return model.TenantSyncHealth{
		TenantID:             tenant.TenantID,
		TenantName:           tenant.Name,
		Status:               status,
		HostsTotal:           hostsTotal,
		HostsReporting:       hostsReporting,
		HostsPending:         hostsPending,
		StoredEvents:         storedEvents,
		LastLocalEventID:     lastLocalEventID,
		LastIngestedAt:       lastIngestedAt,
		BackendVisible:       hostsReporting > 0,
		PrivacyBoundary:      constants.TelemetryPrivacyBoundary,
		OfflineReplayReady:   true,
		OfflineReplaySummary: "Agent stores metadata locally first, then replays unsynced SQLite rows when backend sync is available.",
		Devices:              devices,
		GeneratedAt:          now,
	}, nil
}

func (m *Memory) TenantActivityFeed(_ context.Context, tenantID string, filter model.TenantActivityFeedFilter) (model.TenantActivityFeed, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)
	filter = normalizeActivityFeedFilter(filter)

	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantActivityFeed{}, ErrTenantNotFound
	}

	items := make([]model.TenantActivityFeedItem, 0)
	sourceHosts := make(map[string]bool)
	reportingHosts := make(map[string]bool)
	for _, device := range m.devices {
		if device.TenantID != tenantID {
			continue
		}
		if filter.DeviceID != "" && filter.DeviceID != device.DeviceID {
			continue
		}
		sourceHosts[device.DeviceID] = true
		m.seedDashboardForDeviceLocked(device)
		items = append(items, riskFeedItems(tenantID, device, m.policyEvents[device.DeviceID])...)
		items = append(items, riskFeedItems(tenantID, device, m.anomalyEvents[device.DeviceID])...)
		items = append(items, riskFeedItems(tenantID, device, m.tamperEvents[device.DeviceID])...)
		items = append(items, deliveryFeedItems(tenantID, device, m.alertDeliveries[device.DeviceID])...)
		telemetryEvents := cloneTelemetryEvents(m.telemetryEvents[device.DeviceID])
		if len(telemetryEvents) > 0 {
			reportingHosts[device.DeviceID] = true
		}
		items = append(items, telemetryFeedItems(tenantID, device, telemetryEvents)...)
	}
	if err := m.persistLocked(); err != nil {
		return model.TenantActivityFeed{}, err
	}

	matched := make([]model.TenantActivityFeedItem, 0, len(items))
	for _, item := range items {
		if activityFeedItemMatches(item, filter) {
			matched = append(matched, item)
		}
	}
	sort.Slice(matched, func(i, j int) bool {
		if matched[i].ObservedAt.Equal(matched[j].ObservedAt) {
			return severityRank(matched[i].Severity) > severityRank(matched[j].Severity)
		}
		return matched[i].ObservedAt.After(matched[j].ObservedAt)
	})

	summary := activityFeedSummary(matched, len(sourceHosts), len(reportingHosts))
	limited := matched
	if len(limited) > filter.Limit {
		limited = limited[:filter.Limit]
	}

	return model.TenantActivityFeed{
		TenantID:        tenant.TenantID,
		TenantName:      tenant.Name,
		Filters:         filter,
		Summary:         summary,
		Items:           append([]model.TenantActivityFeedItem(nil), limited...),
		GeneratedAt:     now,
		PrivacyBoundary: constants.TelemetryPrivacyBoundary,
	}, nil
}

func normalizeActivityFeedFilter(filter model.TenantActivityFeedFilter) model.TenantActivityFeedFilter {
	filter.DeviceID = strings.TrimSpace(filter.DeviceID)
	filter.Kind = strings.ToLower(strings.TrimSpace(filter.Kind))
	filter.Severity = strings.ToLower(strings.TrimSpace(filter.Severity))
	filter.Channel = strings.ToLower(strings.TrimSpace(filter.Channel))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.Query = strings.ToLower(strings.TrimSpace(filter.Query))
	if filter.Limit <= 0 {
		filter.Limit = constants.ActivityFeedDefaultLimit
	}
	if filter.Limit > constants.ActivityFeedMaxLimit {
		filter.Limit = constants.ActivityFeedMaxLimit
	}
	return filter
}

func riskFeedItems(tenantID string, device model.Device, events []model.RiskEvent) []model.TenantActivityFeedItem {
	items := make([]model.TenantActivityFeedItem, 0, len(events))
	for _, event := range events {
		title := fallbackString(event.AppName, fallbackString(event.Domain, fallbackString(event.ResourceLabel, event.Category)))
		items = append(items, model.TenantActivityFeedItem{
			ID:             event.ID,
			TenantID:       tenantID,
			DeviceID:       device.DeviceID,
			HostName:       device.HostName,
			Kind:           constants.ActivityFeedKindRisk,
			Type:           event.Type,
			Severity:       event.Severity,
			Category:       event.Category,
			Status:         event.Status,
			Title:          title,
			Detail:         event.Reason,
			Recommendation: event.Recommendation,
			Source:         event.Source,
			ObservedAt:     event.ObservedAt,
		})
	}
	return items
}

func deliveryFeedItems(tenantID string, device model.Device, deliveries []model.AlertDelivery) []model.TenantActivityFeedItem {
	items := make([]model.TenantActivityFeedItem, 0, len(deliveries))
	for _, delivery := range deliveries {
		items = append(items, model.TenantActivityFeedItem{
			ID:         delivery.ID,
			TenantID:   tenantID,
			DeviceID:   device.DeviceID,
			HostName:   device.HostName,
			Kind:       constants.ActivityFeedKindDelivery,
			Type:       constants.ActivityFeedKindDelivery,
			Channel:    delivery.Channel,
			Status:     delivery.Status,
			Title:      fmt.Sprintf("%s delivery %s", titleWord(delivery.Channel), delivery.Status),
			Detail:     fallbackString(delivery.LastError, delivery.Summary),
			Source:     delivery.Provider,
			Provider:   delivery.Provider,
			Recipient:  delivery.Recipient,
			EventID:    delivery.EventID,
			ObservedAt: delivery.LastAttemptAt,
		})
	}
	return items
}

func telemetryFeedItems(tenantID string, device model.Device, events []model.TelemetryEvent) []model.TenantActivityFeedItem {
	items := make([]model.TenantActivityFeedItem, 0, len(events))
	for _, event := range events {
		title := fallbackString(event.AppName, fallbackString(event.Type, "metadata event"))
		items = append(items, model.TenantActivityFeedItem{
			ID:         event.ID,
			TenantID:   tenantID,
			DeviceID:   device.DeviceID,
			HostName:   device.HostName,
			Kind:       constants.ActivityFeedKindTelemetry,
			Type:       event.Type,
			Category:   event.Metadata["category"],
			Status:     constants.StatusOK,
			Title:      title,
			Detail:     fmt.Sprintf("%s metadata with %d redacted fields", event.Source, len(event.Metadata)),
			Source:     event.Source,
			ObservedAt: event.ObservedAt,
		})
	}
	return items
}

func activityFeedItemMatches(item model.TenantActivityFeedItem, filter model.TenantActivityFeedFilter) bool {
	if filter.Kind != "" && strings.ToLower(item.Kind) != filter.Kind {
		return false
	}
	if filter.Severity != "" && strings.ToLower(item.Severity) != filter.Severity {
		return false
	}
	if filter.Channel != "" && strings.ToLower(item.Channel) != filter.Channel {
		return false
	}
	if filter.Status != "" && strings.ToLower(item.Status) != filter.Status {
		return false
	}
	if filter.Query != "" && !strings.Contains(activityFeedSearchText(item), filter.Query) {
		return false
	}
	return true
}

func activityFeedSearchText(item model.TenantActivityFeedItem) string {
	parts := []string{
		item.ID,
		item.DeviceID,
		item.HostName,
		item.Kind,
		item.Type,
		item.Severity,
		item.Category,
		item.Channel,
		item.Status,
		item.Title,
		item.Detail,
		item.Recommendation,
		item.Source,
		item.Provider,
		item.Recipient,
		item.EventID,
	}
	return strings.ToLower(strings.Join(parts, " "))
}

func activityFeedSummary(items []model.TenantActivityFeedItem, sourceHostCount int, reportingHosts int) model.TenantActivityFeedSummary {
	summary := model.TenantActivityFeedSummary{
		Total:           len(items),
		SourceHostCount: sourceHostCount,
		ReportingHosts:  reportingHosts,
	}
	for _, item := range items {
		switch item.Kind {
		case constants.ActivityFeedKindRisk:
			summary.RiskItems++
			if item.Status == constants.RiskStatusOpen && severityRank(item.Severity) >= severityRank(constants.SeverityHigh) {
				summary.HighRiskOpen++
			}
		case constants.ActivityFeedKindDelivery:
			summary.DeliveryItems++
			if item.Channel == constants.DeliveryChannelEmail && item.Status == constants.DeliveryStatusDelivered {
				summary.EmailDelivered++
			}
			if item.Channel == constants.DeliveryChannelPush && item.Status != constants.DeliveryStatusDelivered {
				summary.PushNeedsRetry++
			}
		case constants.ActivityFeedKindTelemetry:
			summary.TelemetryItems++
		}
	}
	return summary
}

func deviceSyncHealth(device model.Device, events []model.TelemetryEvent) model.DeviceSyncHealth {
	sort.Slice(events, func(i, j int) bool {
		return events[i].ObservedAt.After(events[j].ObservedAt)
	})

	var lastObserved time.Time
	var lastLocalEventID int64
	processEvents := 0
	healthEvents := 0
	browserEvents := 0
	recentIDs := make([]string, 0, 5)
	for _, event := range events {
		if event.ObservedAt.After(lastObserved) {
			lastObserved = event.ObservedAt
		}
		if localID := stableLocalEventID(event.ID); localID > lastLocalEventID {
			lastLocalEventID = localID
		}
		if eventSourceMatches(event, "process") {
			processEvents++
		}
		if eventSourceMatches(event, "health") {
			healthEvents++
		}
		if eventSourceMatches(event, "browser") {
			browserEvents++
		}
		if len(recentIDs) < 5 && strings.TrimSpace(event.ID) != "" {
			recentIDs = append(recentIDs, strings.TrimSpace(event.ID))
		}
	}

	status := constants.StatusHealthy
	recommendation := "Backend has replay-safe metadata sync proof for this host."
	if len(events) == 0 {
		status = constants.StatusPending
		recommendation = "Run the agent with backend_sync enabled so this host can report metadata to the dashboard."
	} else if !device.LastSeenAt.IsZero() && time.Since(device.LastSeenAt.UTC()) > 24*time.Hour {
		status = constants.StatusWatch
		recommendation = "Host has stored telemetry but has not checked in recently; confirm the laptop is online and the agent is scheduled."
	}

	return model.DeviceSyncHealth{
		TenantID:          device.TenantID,
		DeviceID:          device.DeviceID,
		HostName:          device.HostName,
		Status:            status,
		StoredEvents:      len(events),
		LastLocalEventID:  lastLocalEventID,
		LastObservedAt:    lastObserved,
		LastIngestedAt:    device.LastSeenAt,
		ProcessEvents:     processEvents,
		HealthEvents:      healthEvents,
		BrowserEvents:     browserEvents,
		RecentEventIDs:    recentIDs,
		Recommendation:    recommendation,
		PrivacyBoundary:   constants.TelemetryPrivacyBoundary,
		BackendVisible:    len(events) > 0,
		OfflineReplayHint: "Stable local-event IDs are idempotent, so offline laptop batches can replay without duplicate backend rows.",
	}
}

func stableLocalEventID(value string) int64 {
	clean := strings.TrimSpace(value)
	clean = strings.TrimPrefix(clean, "local-event-")
	parsed, err := strconv.ParseInt(clean, 10, 64)
	if err != nil || parsed < 0 {
		return 0
	}
	return parsed
}

func eventSourceMatches(event model.TelemetryEvent, token string) bool {
	token = strings.ToLower(strings.TrimSpace(token))
	source := strings.ToLower(strings.TrimSpace(event.Source))
	eventType := strings.ToLower(strings.TrimSpace(event.Type))
	return strings.Contains(source, token) || strings.Contains(eventType, token)
}

func (m *Memory) CreateTenantDataExport(_ context.Context, tenantID string, req model.CreateTenantDataExportRequest) (model.TenantDataExport, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[tenantID]; !ok {
		return model.TenantDataExport{}, ErrTenantNotFound
	}
	expiresAt := now.Add(7 * 24 * time.Hour)
	export := model.TenantDataExport{
		ID:            dataExportID(tenantID, len(m.dataExports[tenantID])+1, now),
		TenantID:      tenantID,
		Format:        strings.TrimSpace(req.Format),
		Scope:         strings.TrimSpace(req.Scope),
		Status:        constants.DataExportStatusReady,
		ResourceCount: tenantResourceCountLocked(m, tenantID),
		StorageKey:    dataExportKey(tenantID, now, strings.TrimSpace(req.Format)),
		RequestedBy:   constants.AuditActorLocalAPI,
		CreatedAt:     now,
		CompletedAt:   now,
		ExpiresAt:     &expiresAt,
	}
	m.dataExports[tenantID] = append(m.dataExports[tenantID], export)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryAccess,
		Action:    constants.AuditActionDataExportCreated,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "tenant data export created",
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.TenantDataExport{}, err
	}
	return export, nil
}

func (m *Memory) ListTenantDataExports(_ context.Context, tenantID string) []model.TenantDataExport {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantID = strings.TrimSpace(tenantID)
	exports := make([]model.TenantDataExport, 0)
	if tenantID == "" {
		for _, tenantExports := range m.dataExports {
			exports = append(exports, tenantExports...)
		}
	} else {
		exports = append(exports, m.dataExports[tenantID]...)
	}
	sort.Slice(exports, func(i, j int) bool {
		if exports[i].TenantID == exports[j].TenantID {
			return exports[i].CreatedAt.Before(exports[j].CreatedAt)
		}
		return exports[i].TenantID < exports[j].TenantID
	})
	return append([]model.TenantDataExport(nil), exports...)
}

func (m *Memory) CreateDeleteRequest(_ context.Context, tenantID string, req model.CreateDeleteRequestRequest) (model.DeleteRequest, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.tenants[tenantID]; !ok {
		return model.DeleteRequest{}, ErrTenantNotFound
	}
	deleteRequest := model.DeleteRequest{
		ID:          deleteRequestID(tenantID, len(m.deleteRequests[tenantID])+1, now),
		TenantID:    tenantID,
		Scope:       strings.TrimSpace(req.Scope),
		Reason:      strings.TrimSpace(req.Reason),
		Status:      constants.DeleteRequestStatusQueued,
		RequestedBy: constants.AuditActorLocalAPI,
		DueAt:       now.Add(30 * 24 * time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	m.deleteRequests[tenantID] = append(m.deleteRequests[tenantID], deleteRequest)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategoryAccess,
		Action:    constants.AuditActionDeleteRequestCreated,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   "delete request queued: " + deleteRequest.Scope,
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.DeleteRequest{}, err
	}
	return deleteRequest, nil
}

func (m *Memory) ListDeleteRequests(_ context.Context, tenantID string) []model.DeleteRequest {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tenantID = strings.TrimSpace(tenantID)
	requests := make([]model.DeleteRequest, 0)
	if tenantID == "" {
		for _, tenantRequests := range m.deleteRequests {
			requests = append(requests, tenantRequests...)
		}
	} else {
		requests = append(requests, m.deleteRequests[tenantID]...)
	}
	sort.Slice(requests, func(i, j int) bool {
		if requests[i].TenantID == requests[j].TenantID {
			return requests[i].CreatedAt.Before(requests[j].CreatedAt)
		}
		return requests[i].TenantID < requests[j].TenantID
	})
	return append([]model.DeleteRequest(nil), requests...)
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

func (m *Memory) seedNotificationRoutesForTenantLocked(tenant model.Tenant) {
	tenantID := strings.TrimSpace(tenant.TenantID)
	if tenantID == "" || len(m.notificationRoutes[tenantID]) > 0 {
		return
	}
	now := time.Now().UTC()
	verifiedAt := now.Add(-10 * time.Minute)
	m.notificationRoutes[tenantID] = []model.NotificationRoute{
		{
			ID:             notificationRouteID(tenantID, 1, now),
			TenantID:       tenantID,
			Channel:        constants.DeliveryChannelEmail,
			Provider:       constants.DeliveryProviderSMTP,
			RecipientLabel: "configured parent email recipient",
			Status:         constants.StatusHealthy,
			Enabled:        true,
			LastVerifiedAt: &verifiedAt,
			LastSummary:    "SMTP route has delivered critical alert proof.",
			CreatedAt:      now,
			UpdatedAt:      now,
		},
		{
			ID:             notificationRouteID(tenantID, 2, now.Add(time.Millisecond)),
			TenantID:       tenantID,
			Channel:        constants.DeliveryChannelPush,
			Provider:       constants.DeliveryProviderWebPush,
			RecipientLabel: "parent mobile push subscription",
			Status:         constants.StatusWatch,
			Enabled:        true,
			LastSummary:    "Push route is configured but needs a delivered retry proof.",
			CreatedAt:      now.Add(time.Millisecond),
			UpdatedAt:      now.Add(time.Millisecond),
		},
		{
			ID:             notificationRouteID(tenantID, 3, now.Add(2*time.Millisecond)),
			TenantID:       tenantID,
			Channel:        constants.DeliveryChannelDashboard,
			Provider:       constants.DeliveryProviderLocalFeed,
			RecipientLabel: "local dashboard feed",
			Status:         constants.StatusHealthy,
			Enabled:        true,
			LastVerifiedAt: &verifiedAt,
			LastSummary:    "Dashboard route is visible in the local command center.",
			CreatedAt:      now.Add(2 * time.Millisecond),
			UpdatedAt:      now.Add(2 * time.Millisecond),
		},
	}
}

func (m *Memory) seedNotificationPreferencesForTenantLocked(tenant model.Tenant) {
	tenantID := strings.TrimSpace(tenant.TenantID)
	if tenantID == "" {
		return
	}
	if _, ok := m.notificationPrefs[tenantID]; ok {
		return
	}
	now := time.Now().UTC()
	m.notificationPrefs[tenantID] = model.NotificationPreferenceCenter{
		TenantID:      tenantID,
		TenantName:    tenant.Name,
		PlanID:        tenant.PlanID,
		PlanName:      planByID(tenant.PlanID).Name,
		Audience:      tenant.PrimaryProfile,
		DigestCadence: constants.NotificationDigestCadenceWeekly,
		QuietHours: model.NotificationQuietHours{
			Enabled:    true,
			StartLocal: "22:30",
			EndLocal:   "06:30",
			Timezone:   "local endpoint timezone",
		},
		Escalation: model.NotificationEscalationPolicy{
			Enabled:         true,
			AfterMinutes:    15,
			RepeatEveryMins: 30,
			MaxRepeats:      2,
			Channels:        []string{constants.DeliveryChannelEmail, constants.DeliveryChannelPush},
			Owner:           "parent or account owner",
		},
		Rules: []model.NotificationPreferenceRule{
			{
				ID:                notificationPreferenceRuleID(tenantID, 1, now),
				TenantID:          tenantID,
				Name:              "Critical tamper alerts",
				EventType:         constants.RiskTypeTamper,
				Severity:          constants.SeverityCritical,
				Channels:          []string{constants.DeliveryChannelEmail, constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
				Mode:              constants.NotificationPreferenceModeImmediate,
				RecipientGroup:    "account owner",
				QuietHoursBypass:  true,
				PaidTier:          constants.PlanFamilyPro,
				DeliverySLA:       "15 minutes",
				NextAction:        "Keep email and push proof current for tamper signals.",
				RetentionEvidence: "audit event and delivery metadata retained by tenant retention tier",
				UpdatedAt:         now,
			},
			{
				ID:                notificationPreferenceRuleID(tenantID, 2, now.Add(time.Millisecond)),
				TenantID:          tenantID,
				Name:              "Non-study entertainment digest",
				EventType:         constants.RiskCategoryEntertainment,
				Severity:          constants.SeverityMedium,
				Channels:          []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
				Mode:              constants.NotificationPreferenceModeDigest,
				RecipientGroup:    "parent weekly report",
				SuppressionLabel:  "study-safe YouTube and coding content suppressed",
				StudySafe:         true,
				QuietHoursBypass:  false,
				PaidTier:          constants.PlanFamilyPro,
				DeliverySLA:       "weekly report",
				NextAction:        "Review digest only when entertainment crosses policy threshold.",
				RetentionEvidence: "weekly report metadata and alert summary retained by tenant tier",
				UpdatedAt:         now.Add(time.Millisecond),
			},
			{
				ID:                notificationPreferenceRuleID(tenantID, 3, now.Add(2*time.Millisecond)),
				TenantID:          tenantID,
				Name:              "Study-safe learning activity",
				EventType:         constants.AlertTriggerNonStudyYouTube,
				Severity:          constants.SeverityLow,
				Channels:          []string{constants.DeliveryChannelDashboard},
				Mode:              constants.NotificationPreferenceModeSilent,
				RecipientGroup:    "dashboard archive",
				SuppressionLabel:  "coding, mathematics, system design, and study topics",
				StudySafe:         true,
				QuietHoursBypass:  false,
				PaidTier:          constants.PlanFree,
				DeliverySLA:       "dashboard only",
				NextAction:        "Suppress alerts when classifier marks the session as study-safe.",
				RetentionEvidence: "category metadata only; no raw URLs or page titles",
				UpdatedAt:         now.Add(2 * time.Millisecond),
			},
		},
		PrivacyBoundary: constants.NotificationPreferencePrivacyNote,
		GeneratedAt:     now,
		UpdatedAt:       now,
	}
}

func (m *Memory) seedActivityViewsForTenantLocked(tenant model.Tenant) {
	tenantID := strings.TrimSpace(tenant.TenantID)
	if tenantID == "" || len(m.activityViews[tenantID]) > 0 {
		return
	}
	now := time.Now().UTC()
	m.activityViews[tenantID] = []model.TenantActivityView{
		{
			ID:          constants.ActivityViewHighRiskOpen,
			TenantID:    tenantID,
			Name:        "Open high-risk anomalies",
			Description: "Prioritise open policy, anomaly, and tamper signals that need a human decision.",
			Filter: model.TenantActivityFeedFilter{
				Kind:     constants.ActivityFeedKindRisk,
				Severity: constants.SeverityHigh,
				Status:   constants.RiskStatusOpen,
				Limit:    constants.ActivityFeedDefaultLimit,
			},
			PaidTier:  constants.PlanFamilyPro,
			SortOrder: 1,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:          constants.ActivityViewEmailProof,
			TenantID:    tenantID,
			Name:        "Mail delivery proof",
			Description: "Show delivered email evidence for anomaly alerts and weekly report packaging.",
			Filter: model.TenantActivityFeedFilter{
				Kind:    constants.ActivityFeedKindDelivery,
				Channel: constants.DeliveryChannelEmail,
				Status:  constants.DeliveryStatusDelivered,
				Limit:   constants.ActivityFeedDefaultLimit,
			},
			PaidTier:  constants.PlanFamilyPro,
			SortOrder: 2,
			CreatedAt: now.Add(time.Millisecond),
			UpdatedAt: now.Add(time.Millisecond),
		},
		{
			ID:          constants.ActivityViewPushRetry,
			TenantID:    tenantID,
			Name:        "Push retry watch",
			Description: "Surface push routes that are configured but still need delivered anomaly proof.",
			Filter: model.TenantActivityFeedFilter{
				Kind:    constants.ActivityFeedKindDelivery,
				Channel: constants.DeliveryChannelPush,
				Status:  constants.DeliveryStatusRetrying,
				Limit:   constants.ActivityFeedDefaultLimit,
			},
			PaidTier:  constants.PlanFamilyPro,
			SortOrder: 3,
			CreatedAt: now.Add(2 * time.Millisecond),
			UpdatedAt: now.Add(2 * time.Millisecond),
		},
		{
			ID:          constants.ActivityViewSyncProof,
			TenantID:    tenantID,
			Name:        "Sync and archive proof",
			Description: "Verify metadata replay, dashboard visibility, and S3-backed archive readiness.",
			Filter: model.TenantActivityFeedFilter{
				Kind:  constants.ActivityFeedKindTelemetry,
				Limit: constants.ActivityFeedDefaultLimit,
			},
			PaidTier:  constants.PlanSchool,
			SortOrder: 4,
			CreatedAt: now.Add(3 * time.Millisecond),
			UpdatedAt: now.Add(3 * time.Millisecond),
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

func (m *Memory) IngestTelemetryEvents(_ context.Context, deviceID string, req model.IngestTelemetryRequest) (model.IngestTelemetryResponse, error) {
	now := time.Now().UTC()
	deviceID = strings.TrimSpace(deviceID)
	req.DeviceID = strings.TrimSpace(req.DeviceID)
	if req.DeviceID == "" {
		req.DeviceID = deviceID
	}
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.HostName = strings.TrimSpace(req.HostName)
	req.Profile = strings.TrimSpace(req.Profile)
	req.OSName = strings.TrimSpace(req.OSName)

	m.mu.Lock()
	defer m.mu.Unlock()

	if req.DeviceID != deviceID {
		return model.IngestTelemetryResponse{}, ErrDeviceNotFound
	}
	tenant, ok := m.tenants[req.TenantID]
	if !ok {
		return model.IngestTelemetryResponse{}, ErrTenantNotFound
	}

	device, ok := m.devices[deviceID]
	if !ok {
		device = model.Device{
			TenantID:   tenant.TenantID,
			DeviceID:   deviceID,
			HostName:   fallbackString(req.HostName, deviceID),
			Profile:    fallbackString(req.Profile, tenant.PrimaryProfile),
			OSName:     fallbackString(req.OSName, constants.StatusEmpty),
			EnrolledAt: now,
			LastSeenAt: now,
		}
	} else if device.TenantID != tenant.TenantID {
		return model.IngestTelemetryResponse{}, ErrDeviceNotFound
	}
	device.HostName = fallbackString(req.HostName, device.HostName)
	device.Profile = fallbackString(req.Profile, device.Profile)
	device.OSName = fallbackString(req.OSName, device.OSName)
	device.LastSeenAt = now
	m.devices[deviceID] = device

	limit := len(req.Events)
	if limit > constants.TelemetryIngestMaxEvents {
		limit = constants.TelemetryIngestMaxEvents
	}
	seenTelemetryIDs := telemetryEventIDs(m.telemetryEvents[deviceID])
	accepted := make([]model.TelemetryEvent, 0, limit)
	acceptedPayloadEvents := 0
	var lastObserved time.Time
	for i := 0; i < limit; i++ {
		evt := normalizeTelemetryEvent(req.Events[i], tenant.TenantID, device)
		if evt.ObservedAt.After(lastObserved) {
			lastObserved = evt.ObservedAt
		}
		acceptedPayloadEvents++
		if evt.ID != "" {
			if seenTelemetryIDs[evt.ID] {
				continue
			}
			seenTelemetryIDs[evt.ID] = true
		}
		accepted = append(accepted, evt)
	}
	m.telemetryEvents[deviceID] = append(m.telemetryEvents[deviceID], accepted...)
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenant.TenantID, len(m.auditEvents)+1, now),
		TenantID:  tenant.TenantID,
		Category:  constants.AuditCategorySystem,
		Action:    constants.AuditActionTelemetryIngested,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   fmt.Sprintf("telemetry ingest accepted %d metadata events for %s", acceptedPayloadEvents, deviceID),
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		return model.IngestTelemetryResponse{}, err
	}

	return model.IngestTelemetryResponse{
		TenantID:           tenant.TenantID,
		DeviceID:           deviceID,
		AcceptedEvents:     acceptedPayloadEvents,
		StoredEvents:       len(m.telemetryEvents[deviceID]),
		LastObservedAt:     lastObserved,
		LastIngestedAt:     now,
		PrivacyBoundary:    constants.TelemetryPrivacyBoundary,
		BackendVisibleHost: true,
	}, nil
}

func (m *Memory) TelemetryIngestStatus(_ context.Context, deviceID string) (model.TelemetryIngestStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	device, ok := m.devices[strings.TrimSpace(deviceID)]
	if !ok {
		return model.TelemetryIngestStatus{}, ErrDeviceNotFound
	}
	events := cloneTelemetryEvents(m.telemetryEvents[device.DeviceID])
	countsByType := make(map[string]int)
	countsBySource := make(map[string]int)
	var lastObserved time.Time
	for _, evt := range events {
		countsByType[evt.Type]++
		countsBySource[evt.Source]++
		if evt.ObservedAt.After(lastObserved) {
			lastObserved = evt.ObservedAt
		}
	}
	sort.Slice(events, func(i, j int) bool {
		return events[i].ObservedAt.After(events[j].ObservedAt)
	})
	recent := events
	if len(recent) > constants.TelemetryStatusRecentEvents {
		recent = recent[:constants.TelemetryStatusRecentEvents]
	}
	return model.TelemetryIngestStatus{
		TenantID:        device.TenantID,
		DeviceID:        device.DeviceID,
		HostName:        device.HostName,
		StoredEvents:    len(events),
		CountsByType:    countsByType,
		CountsBySource:  countsBySource,
		LastObservedAt:  lastObserved,
		LastIngestedAt:  device.LastSeenAt,
		RecentEvents:    cloneTelemetryEvents(recent),
		PrivacyBoundary: constants.TelemetryPrivacyBoundary,
	}, nil
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

func notificationRouteID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"notification-route",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func notificationPreferenceRuleID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"notification-pref",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func deliveryRemediationID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"delivery-remediation",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func activityViewID(requestedID string, tenantID string, sequence int) string {
	if id := strings.TrimSpace(requestedID); id != "" {
		return id
	}
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"activity-view",
		fmt.Sprintf("%03d", sequence),
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

func dataExportID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"data-export",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func deleteRequestID(tenantID string, sequence int, createdAt time.Time) string {
	return strings.Join([]string{
		strings.TrimSpace(tenantID),
		"delete-request",
		fmt.Sprintf("%03d", sequence),
		createdAt.Format("20060102T150405Z"),
	}, "-")
}

func dataExportKey(tenantID string, createdAt time.Time, format string) string {
	return strings.Join([]string{
		"tenant=" + strings.TrimSpace(tenantID),
		"exports",
		createdAt.Format(time.DateOnly),
		createdAt.Format("150405") + "." + strings.TrimSpace(format),
	}, "/")
}

func tenantResourceCountLocked(m *Memory, tenantID string) int {
	count := 0
	if _, ok := m.tenants[tenantID]; ok {
		count++
	}
	for _, device := range m.devices {
		if device.TenantID == tenantID {
			count++
		}
	}
	for _, event := range m.auditEvents {
		if event.TenantID == tenantID {
			count++
		}
	}
	count += len(m.alertRules[tenantID])
	count += len(m.notificationRoutes[tenantID])
	count += len(m.dataExports[tenantID])
	count += len(m.deleteRequests[tenantID])
	count += len(m.deviceGroups[tenantID])
	count += len(m.policyAssigns[tenantID])
	return count
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

func fallbackString(value string, fallback string) string {
	clean := strings.TrimSpace(value)
	if clean != "" {
		return clean
	}
	return strings.TrimSpace(fallback)
}

func normalizeTelemetryEvent(evt model.TelemetryEvent, tenantID string, device model.Device) model.TelemetryEvent {
	evt.ID = strings.TrimSpace(evt.ID)
	evt.Type = strings.TrimSpace(evt.Type)
	evt.Source = strings.TrimSpace(evt.Source)
	evt.TenantID = tenantID
	evt.DeviceID = device.DeviceID
	evt.HostName = device.HostName
	evt.AppName = strings.TrimSpace(evt.AppName)
	evt.PathHash = strings.TrimSpace(evt.PathHash)
	if evt.ObservedAt.IsZero() {
		evt.ObservedAt = time.Now().UTC()
	} else {
		evt.ObservedAt = evt.ObservedAt.UTC()
	}
	metadata := make(map[string]string, len(evt.Metadata))
	for key, value := range evt.Metadata {
		cleanKey := strings.TrimSpace(key)
		if cleanKey == "" {
			continue
		}
		metadata[cleanKey] = strings.TrimSpace(value)
	}
	evt.Metadata = metadata
	return evt
}

func telemetryEventIDs(events []model.TelemetryEvent) map[string]bool {
	ids := make(map[string]bool, len(events))
	for _, event := range events {
		id := strings.TrimSpace(event.ID)
		if id != "" {
			ids[id] = true
		}
	}
	return ids
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
	Version            string                                             `json:"version"`
	Tenants            map[string]model.Tenant                            `json:"tenants"`
	Devices            map[string]model.Device                            `json:"devices"`
	AuditEvents        []model.AuditEvent                                 `json:"audit_events"`
	AlertRules         map[string][]model.AlertRule                       `json:"alert_rules"`
	NotificationRoutes map[string][]model.NotificationRoute               `json:"notification_routes"`
	NotificationPrefs  map[string]model.NotificationPreferenceCenter      `json:"notification_preferences"`
	DeliveryRemedies   map[string][]model.TenantDeliveryRemediationAction `json:"delivery_remedies"`
	ActivityViews      map[string][]model.TenantActivityView              `json:"activity_views"`
	DataExports        map[string][]model.TenantDataExport                `json:"data_exports"`
	DeleteRequests     map[string][]model.DeleteRequest                   `json:"delete_requests"`
	DeviceGroups       map[string][]model.DeviceGroup                     `json:"device_groups"`
	PolicyAssigns      map[string][]model.PolicyAssignment                `json:"policy_assignments"`
	PolicyEvents       map[string][]model.RiskEvent                       `json:"policy_events"`
	AnomalyEvents      map[string][]model.RiskEvent                       `json:"anomaly_events"`
	TamperEvents       map[string][]model.RiskEvent                       `json:"tamper_events"`
	AlertDeliveries    map[string][]model.AlertDelivery                   `json:"alert_deliveries"`
	HealthScores       map[string]model.DeviceHealth                      `json:"health_scores"`
	TelemetryEvents    map[string][]model.TelemetryEvent                  `json:"telemetry_events"`
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
	m.notificationRoutes = cloneNotificationRouteMap(state.NotificationRoutes)
	m.notificationPrefs = cloneNotificationPreferenceMap(state.NotificationPrefs)
	m.deliveryRemedies = cloneDeliveryRemediationMap(state.DeliveryRemedies)
	m.activityViews = cloneActivityViewMap(state.ActivityViews)
	m.dataExports = cloneDataExportMap(state.DataExports)
	m.deleteRequests = cloneDeleteRequestMap(state.DeleteRequests)
	m.deviceGroups = cloneDeviceGroupMap(state.DeviceGroups)
	m.policyAssigns = clonePolicyAssignmentMap(state.PolicyAssigns)
	m.policyEvents = cloneRiskMap(state.PolicyEvents)
	m.anomalyEvents = cloneRiskMap(state.AnomalyEvents)
	m.tamperEvents = cloneRiskMap(state.TamperEvents)
	m.alertDeliveries = cloneDeliveryMap(state.AlertDeliveries)
	m.healthScores = cloneHealthMap(state.HealthScores)
	m.telemetryEvents = cloneTelemetryMap(state.TelemetryEvents)
	return nil
}

func (m *Memory) persistLocked() error {
	if strings.TrimSpace(m.path) == "" {
		return nil
	}
	state := persistentState{
		Version:            constants.BackendVersion,
		Tenants:            cloneTenantMap(m.tenants),
		Devices:            cloneDeviceMap(m.devices),
		AuditEvents:        append([]model.AuditEvent(nil), m.auditEvents...),
		AlertRules:         cloneAlertRuleMap(m.alertRules),
		NotificationRoutes: cloneNotificationRouteMap(m.notificationRoutes),
		NotificationPrefs:  cloneNotificationPreferenceMap(m.notificationPrefs),
		DeliveryRemedies:   cloneDeliveryRemediationMap(m.deliveryRemedies),
		ActivityViews:      cloneActivityViewMap(m.activityViews),
		DataExports:        cloneDataExportMap(m.dataExports),
		DeleteRequests:     cloneDeleteRequestMap(m.deleteRequests),
		DeviceGroups:       cloneDeviceGroupMap(m.deviceGroups),
		PolicyAssigns:      clonePolicyAssignmentMap(m.policyAssigns),
		PolicyEvents:       cloneRiskMap(m.policyEvents),
		AnomalyEvents:      cloneRiskMap(m.anomalyEvents),
		TamperEvents:       cloneRiskMap(m.tamperEvents),
		AlertDeliveries:    cloneDeliveryMap(m.alertDeliveries),
		HealthScores:       cloneHealthMap(m.healthScores),
		TelemetryEvents:    cloneTelemetryMap(m.telemetryEvents),
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

func cloneNotificationRouteMap(input map[string][]model.NotificationRoute) map[string][]model.NotificationRoute {
	output := make(map[string][]model.NotificationRoute, len(input))
	for key, value := range input {
		output[key] = append([]model.NotificationRoute(nil), value...)
	}
	return output
}

func cloneNotificationPreferenceMap(input map[string]model.NotificationPreferenceCenter) map[string]model.NotificationPreferenceCenter {
	output := make(map[string]model.NotificationPreferenceCenter, len(input))
	for key, value := range input {
		value.Rules = append([]model.NotificationPreferenceRule(nil), value.Rules...)
		value.Escalation.Channels = append([]string(nil), value.Escalation.Channels...)
		output[key] = value
	}
	return output
}

func cloneDeliveryRemediations(input []model.TenantDeliveryRemediationAction) []model.TenantDeliveryRemediationAction {
	return append([]model.TenantDeliveryRemediationAction(nil), input...)
}

func cloneDeliveryRemediationMap(input map[string][]model.TenantDeliveryRemediationAction) map[string][]model.TenantDeliveryRemediationAction {
	output := make(map[string][]model.TenantDeliveryRemediationAction, len(input))
	for key, value := range input {
		output[key] = cloneDeliveryRemediations(value)
	}
	return output
}

func cloneActivityViews(input []model.TenantActivityView) []model.TenantActivityView {
	return append([]model.TenantActivityView(nil), input...)
}

func cloneActivityViewMap(input map[string][]model.TenantActivityView) map[string][]model.TenantActivityView {
	output := make(map[string][]model.TenantActivityView, len(input))
	for key, value := range input {
		output[key] = cloneActivityViews(value)
	}
	return output
}

func cloneDataExportMap(input map[string][]model.TenantDataExport) map[string][]model.TenantDataExport {
	output := make(map[string][]model.TenantDataExport, len(input))
	for key, value := range input {
		output[key] = append([]model.TenantDataExport(nil), value...)
	}
	return output
}

func cloneDeleteRequestMap(input map[string][]model.DeleteRequest) map[string][]model.DeleteRequest {
	output := make(map[string][]model.DeleteRequest, len(input))
	for key, value := range input {
		output[key] = append([]model.DeleteRequest(nil), value...)
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

func cloneTelemetryEvents(input []model.TelemetryEvent) []model.TelemetryEvent {
	output := append([]model.TelemetryEvent(nil), input...)
	for index := range output {
		output[index].Metadata = cloneStringMap(output[index].Metadata)
	}
	return output
}

func cloneTelemetryMap(input map[string][]model.TelemetryEvent) map[string][]model.TelemetryEvent {
	output := make(map[string][]model.TelemetryEvent, len(input))
	for key, value := range input {
		output[key] = cloneTelemetryEvents(value)
	}
	return output
}

func cloneStringMap(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
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

func planByID(planID string) model.Plan {
	for _, plan := range Plans() {
		if plan.ID == strings.TrimSpace(planID) {
			return plan
		}
	}
	return model.Plan{ID: strings.TrimSpace(planID), Name: strings.TrimSpace(planID)}
}

func tenantReadinessScore(tenant model.Tenant, plan model.Plan, hostsTotal int, emailDelivered int, pushDelivered int, dashboardDelivered int, alertRules int, groups int, assignments int) int {
	checks := []bool{
		strings.TrimSpace(tenant.PlanID) != "",
		strings.TrimSpace(tenant.RetentionTierID) != "",
		hostsTotal > 0,
		plan.CloudArchive,
		plan.WeeklyReports,
		plan.RoleBasedDashboard,
		emailDelivered > 0,
		pushDelivered > 0,
		dashboardDelivered > 0,
		alertRules > 0,
		groups > 0,
		assignments > 0,
	}
	passed := 0
	for _, check := range checks {
		if check {
			passed++
		}
	}
	return (passed * 100) / len(checks)
}

func deliverySnapshot(delivery *model.AlertDelivery) *model.TenantDeliverySnapshot {
	if delivery == nil {
		return nil
	}
	return &model.TenantDeliverySnapshot{
		Channel:       delivery.Channel,
		Status:        delivery.Status,
		Recipient:     delivery.Recipient,
		Provider:      delivery.Provider,
		LastAttemptAt: delivery.LastAttemptAt,
		Summary:       delivery.Summary,
	}
}

func latestTenantDelivery(deliveries []model.AlertDelivery, channel string) *model.AlertDelivery {
	var latest *model.AlertDelivery
	for index := range deliveries {
		if deliveries[index].Channel != channel {
			continue
		}
		if latest == nil || deliveries[index].LastAttemptAt.After(latest.LastAttemptAt) {
			current := deliveries[index]
			latest = &current
		}
	}
	return latest
}

func tenantPrioritySignals(events []model.RiskEvent, deliveries []model.AlertDelivery, observedAt time.Time) []model.TenantOperationsSignal {
	signals := make([]model.TenantOperationsSignal, 0, 4)
	if delivery := topTenantDeliveryProblem(deliveries); delivery != nil {
		signals = append(signals, model.TenantOperationsSignal{
			Title:      titleWord(delivery.Channel) + " delivery needs attention",
			Detail:     firstNonEmpty(delivery.LastError, delivery.Summary, "Review provider route health and retry policy."),
			Severity:   constants.SeverityMedium,
			Channel:    delivery.Channel,
			Status:     delivery.Status,
			Owner:      delivery.Recipient,
			ObservedAt: delivery.LastAttemptAt,
		})
	}
	for _, event := range topTenantRiskEvents(events, 3) {
		signals = append(signals, model.TenantOperationsSignal{
			Title:      eventTitle(event),
			Detail:     firstNonEmpty(event.Recommendation, event.Reason, "Review this signal."),
			Severity:   event.Severity,
			Channel:    event.Source,
			Status:     event.Status,
			Owner:      event.Category,
			ObservedAt: event.ObservedAt,
		})
	}
	if len(signals) == 0 {
		signals = append(signals, model.TenantOperationsSignal{
			Title:      "No immediate escalation",
			Detail:     "Tenant routes and host signals are ready for command-center review.",
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			ObservedAt: observedAt,
		})
	}
	return signals
}

func tenantUpgradeSignals(tenant model.Tenant, plan model.Plan, readiness int, observedAt time.Time) []model.TenantOperationsSignal {
	signals := []model.TenantOperationsSignal{
		{
			Title:      "Family Pro proof pack",
			Detail:     "Weekly report, email alert, push route, dashboard feed, and S3 archive value are visible in one customer view.",
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelEmail,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleParent,
			ObservedAt: observedAt,
		},
		{
			Title:      "School rollout packaging",
			Detail:     "Device groups, policy assignments, consent center, audit history, and data rights workflows support managed cohorts.",
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleSchoolAdmin,
			ObservedAt: observedAt,
		},
		{
			Title:      "Business risk observability",
			Detail:     "Risky software, device health, archive backlog, tamper signals, and notification reliability support paid endpoint plans.",
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			ObservedAt: observedAt,
		},
	}
	if readiness < 80 {
		signals = append([]model.TenantOperationsSignal{{
			Title:      "Readiness gap before upgrade pitch",
			Detail:     "Improve route delivery, enrollment, or plan packaging before presenting this tenant as production-ready.",
			Severity:   constants.SeverityMedium,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusWatch,
			Owner:      plan.Name,
			ObservedAt: observedAt,
		}}, signals...)
	}
	if tenant.PlanID == constants.PlanFree {
		signals = append([]model.TenantOperationsSignal{{
			Title:      "Upgrade candidate",
			Detail:     "Cloud archive, weekly reports, role views, and notification proof are paid-plan conversion levers.",
			Severity:   constants.SeverityLow,
			Channel:    constants.DeliveryChannelEmail,
			Status:     constants.StatusWatch,
			Owner:      constants.PlanFamilyPro,
			ObservedAt: observedAt,
		}}, signals...)
	}
	return signals
}

func tenantTrustScore(operations model.TenantOperationsSummary, auditCount int, exportsCount int, deletesCount int) int {
	checks := []bool{
		operations.TamperSignals == 0 || operations.ArchiveBacklog <= 2,
		operations.DeliveryFailed == 0,
		operations.DashboardDelivered > 0,
		auditCount > 0,
		exportsCount > 0 || deletesCount > 0 || auditCount > 0,
		operations.HostsTotal > 0,
	}
	passed := 0
	for _, check := range checks {
		if check {
			passed++
		}
	}
	return (passed * 100) / len(checks)
}

func monetizationStage(tenant model.Tenant, operations model.TenantOperationsSummary, trustScore int) string {
	switch {
	case tenant.PlanID == constants.PlanFree && operations.MonetizationReadiness >= 60:
		return constants.MonetizationStageConversionReady
	case operations.MonetizationReadiness >= 85 && trustScore >= 80 && operations.HostsTotal > 1:
		return constants.MonetizationStageExpansionReady
	case operations.MonetizationReadiness >= 70 && trustScore >= 65:
		return constants.MonetizationStagePilotReady
	default:
		return constants.MonetizationStageProofGap
	}
}

func monetizationHealth(operations model.TenantOperationsSummary, trustScore int) string {
	switch {
	case operations.MonetizationReadiness >= 80 && operations.NotificationScore >= 65 && trustScore >= 75:
		return constants.StatusHealthy
	case operations.MonetizationReadiness >= 55 && operations.NotificationScore >= 50:
		return constants.StatusWatch
	default:
		return constants.StatusAttention
	}
}

func notificationPromise(operations model.TenantOperationsSummary, trustScore int, email *model.AlertDelivery, push *model.AlertDelivery, dashboard *model.AlertDelivery) model.TenantNotificationPromise {
	return model.TenantNotificationPromise{
		Status:    monetizationHealth(operations, trustScore),
		Summary:   fmt.Sprintf("%d/%d notification routes delivered with %d retrying", operations.DeliveryDelivered, operations.DeliveryTotal, operations.DeliveryRetrying),
		Email:     notificationPromiseLine(email),
		Push:      notificationPromiseLine(push),
		Dashboard: notificationPromiseLine(dashboard),
	}
}

func notificationPromiseLine(delivery *model.AlertDelivery) string {
	if delivery == nil {
		return "route not configured"
	}
	return strings.Join([]string{
		delivery.Status,
		delivery.Provider,
		firstNonEmpty(delivery.Recipient, "recipient pending"),
	}, " / ")
}

func notificationRoute(channel string, delivery *model.AlertDelivery) model.TenantNotificationRoute {
	if delivery == nil {
		return model.TenantNotificationRoute{
			Channel:    channel,
			Status:     constants.DeliveryStatusPending,
			Proof:      "No delivery proof has been recorded for this route.",
			NextAction: "Configure the route and send a demo alert.",
		}
	}
	return model.TenantNotificationRoute{
		Channel:       delivery.Channel,
		Provider:      delivery.Provider,
		Status:        delivery.Status,
		Recipient:     delivery.Recipient,
		Attempts:      delivery.Attempts,
		LastAttemptAt: delivery.LastAttemptAt,
		NextRetryAt:   delivery.NextRetryAt,
		Proof:         firstNonEmpty(delivery.Summary, "Route attempt is visible in dashboard delivery history."),
		NextAction:    notificationNextAction(delivery),
	}
}

func notificationNextAction(delivery *model.AlertDelivery) string {
	switch delivery.Status {
	case constants.DeliveryStatusDelivered:
		return "Use this route as customer proof."
	case constants.DeliveryStatusRetrying:
		return "Watch retry timing and provider health."
	case constants.DeliveryStatusFailed:
		return "Fix provider credentials or endpoint subscription."
	case constants.DeliveryStatusSuppressed:
		return "Review suppression policy before demo."
	default:
		return "Send a proof notification."
	}
}

func buildNotificationPreferenceCenter(tenant model.Tenant, stored model.NotificationPreferenceCenter, routes []model.NotificationRoute, generatedAt time.Time) model.NotificationPreferenceCenter {
	plan := planByID(tenant.PlanID)
	rules := normalizeNotificationPreferenceRules(tenant.TenantID, stored.Rules, generatedAt)
	if len(rules) == 0 {
		rules = normalizeNotificationPreferenceRules(tenant.TenantID, defaultNotificationPreferenceRules(tenant.TenantID, generatedAt), generatedAt)
	}
	quietHours := stored.QuietHours
	if strings.TrimSpace(quietHours.StartLocal) == "" {
		quietHours = model.NotificationQuietHours{
			Enabled:    true,
			StartLocal: "22:30",
			EndLocal:   "06:30",
			Timezone:   "local endpoint timezone",
		}
	}
	escalation := stored.Escalation
	if escalation.AfterMinutes == 0 {
		escalation = model.NotificationEscalationPolicy{
			Enabled:         true,
			AfterMinutes:    15,
			RepeatEveryMins: 30,
			MaxRepeats:      2,
			Channels:        []string{constants.DeliveryChannelEmail, constants.DeliveryChannelPush},
			Owner:           "parent or account owner",
		}
	}
	digestCadence := strings.TrimSpace(stored.DigestCadence)
	if digestCadence == "" {
		digestCadence = constants.NotificationDigestCadenceWeekly
	}
	summary := notificationPreferenceSummary(rules, routes, quietHours, escalation)
	return model.NotificationPreferenceCenter{
		TenantID:        tenant.TenantID,
		TenantName:      tenant.Name,
		PlanID:          tenant.PlanID,
		PlanName:        plan.Name,
		Audience:        firstNonEmpty(plan.Audience, tenant.PrimaryProfile),
		DigestCadence:   digestCadence,
		QuietHours:      quietHours,
		Escalation:      escalation,
		Summary:         summary,
		Rules:           rules,
		PrivacyBoundary: constants.NotificationPreferencePrivacyNote,
		GeneratedAt:     generatedAt,
		UpdatedAt:       stored.UpdatedAt,
	}
}

func notificationPreferenceSummary(rules []model.NotificationPreferenceRule, routes []model.NotificationRoute, quietHours model.NotificationQuietHours, escalation model.NotificationEscalationPolicy) model.NotificationPreferenceCenterSummary {
	channels := map[string]bool{}
	routesNeedingProof := 0
	for _, route := range routes {
		if route.Enabled {
			channels[route.Channel] = true
		}
		if !route.Enabled || route.Status != constants.StatusHealthy || route.LastVerifiedAt == nil {
			routesNeedingProof++
		}
	}
	immediate := 0
	digest := 0
	silent := 0
	studySafe := 0
	for _, rule := range rules {
		switch rule.Mode {
		case constants.NotificationPreferenceModeImmediate:
			immediate++
		case constants.NotificationPreferenceModeDigest:
			digest++
		case constants.NotificationPreferenceModeSilent:
			silent++
		}
		if rule.StudySafe || strings.TrimSpace(rule.SuppressionLabel) != "" {
			studySafe++
		}
	}
	checks := []bool{
		len(rules) > 0,
		immediate > 0,
		digest > 0,
		silent > 0,
		channels[constants.DeliveryChannelEmail],
		channels[constants.DeliveryChannelPush],
		channels[constants.DeliveryChannelDashboard],
		quietHours.Enabled,
		escalation.Enabled,
		studySafe > 0,
		routesNeedingProof == 0,
	}
	score := (countTrue(checks) * 100) / len(checks)
	status := constants.StatusHealthy
	if routesNeedingProof > 0 || !channels[constants.DeliveryChannelPush] {
		status = constants.StatusWatch
	}
	if len(rules) == 0 || !channels[constants.DeliveryChannelEmail] {
		status = constants.StatusAttention
	}
	return model.NotificationPreferenceCenterSummary{
		Status:                status,
		PreferenceScore:       score,
		RulesTotal:            len(rules),
		ImmediateRules:        immediate,
		DigestRules:           digest,
		SilentRules:           silent,
		EmailEnabled:          channels[constants.DeliveryChannelEmail],
		PushEnabled:           channels[constants.DeliveryChannelPush],
		DashboardEnabled:      channels[constants.DeliveryChannelDashboard],
		QuietHoursEnabled:     quietHours.Enabled,
		EscalationEnabled:     escalation.Enabled,
		StudySuppressionRules: studySafe,
		RoutesNeedingProof:    routesNeedingProof,
		RecommendedPaidTier:   constants.PlanFamilyPro,
	}
}

func defaultNotificationPreferenceRules(tenantID string, now time.Time) []model.NotificationPreferenceRule {
	return []model.NotificationPreferenceRule{
		{
			ID:                notificationPreferenceRuleID(tenantID, 1, now),
			TenantID:          tenantID,
			Name:              "Critical tamper alerts",
			EventType:         constants.RiskTypeTamper,
			Severity:          constants.SeverityCritical,
			Channels:          []string{constants.DeliveryChannelEmail, constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Mode:              constants.NotificationPreferenceModeImmediate,
			RecipientGroup:    "account owner",
			QuietHoursBypass:  true,
			PaidTier:          constants.PlanFamilyPro,
			DeliverySLA:       "15 minutes",
			NextAction:        "Keep email and push proof current for tamper signals.",
			RetentionEvidence: "audit event and delivery metadata retained by tenant retention tier",
			UpdatedAt:         now,
		},
	}
}

func normalizeNotificationPreferenceRules(tenantID string, rules []model.NotificationPreferenceRule, now time.Time) []model.NotificationPreferenceRule {
	normalized := make([]model.NotificationPreferenceRule, 0, len(rules))
	for index, rule := range rules {
		rule.ID = strings.TrimSpace(rule.ID)
		if rule.ID == "" {
			rule.ID = notificationPreferenceRuleID(tenantID, index+1, now)
		}
		rule.TenantID = tenantID
		rule.Name = strings.TrimSpace(rule.Name)
		rule.EventType = strings.TrimSpace(rule.EventType)
		rule.Severity = strings.TrimSpace(rule.Severity)
		rule.Mode = strings.TrimSpace(rule.Mode)
		rule.RecipientGroup = strings.TrimSpace(rule.RecipientGroup)
		rule.SuppressionLabel = strings.TrimSpace(rule.SuppressionLabel)
		rule.PaidTier = strings.TrimSpace(rule.PaidTier)
		rule.DeliverySLA = strings.TrimSpace(rule.DeliverySLA)
		rule.NextAction = strings.TrimSpace(rule.NextAction)
		rule.RetentionEvidence = strings.TrimSpace(rule.RetentionEvidence)
		if rule.Mode == "" {
			rule.Mode = constants.NotificationPreferenceModeImmediate
		}
		if rule.Severity == "" {
			rule.Severity = constants.SeverityMedium
		}
		rule.Channels = normalizeStringSlice(rule.Channels)
		if len(rule.Channels) == 0 {
			rule.Channels = []string{constants.DeliveryChannelDashboard}
		}
		if rule.PaidTier == "" {
			rule.PaidTier = constants.PlanFamilyPro
		}
		if rule.DeliverySLA == "" {
			rule.DeliverySLA = "tenant policy"
		}
		if rule.UpdatedAt.IsZero() {
			rule.UpdatedAt = now
		}
		normalized = append(normalized, rule)
	}
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Mode == normalized[j].Mode {
			return normalized[i].Name < normalized[j].Name
		}
		return normalized[i].Mode < normalized[j].Mode
	})
	return normalized
}

func normalizeStringSlice(values []string) []string {
	output := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		clean := strings.TrimSpace(value)
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		output = append(output, clean)
	}
	return output
}

func countTrue(values []bool) int {
	count := 0
	for _, value := range values {
		if value {
			count++
		}
	}
	return count
}

func deliveriesByEventID(deliveries []model.AlertDelivery) map[string][]model.AlertDelivery {
	output := make(map[string][]model.AlertDelivery)
	for _, delivery := range deliveries {
		eventID := strings.TrimSpace(delivery.EventID)
		if eventID == "" {
			continue
		}
		output[eventID] = append(output[eventID], delivery)
	}
	return output
}

func alertDeliveryProof(deliveries []model.AlertDelivery) []model.TenantAlertDeliveryProof {
	proof := make([]model.TenantAlertDeliveryProof, 0, len(deliveries))
	sort.Slice(deliveries, func(i, j int) bool {
		if deliveries[i].Channel != deliveries[j].Channel {
			return deliveries[i].Channel < deliveries[j].Channel
		}
		return deliveries[i].LastAttemptAt.After(deliveries[j].LastAttemptAt)
	})
	for _, delivery := range deliveries {
		proof = append(proof, model.TenantAlertDeliveryProof{
			Channel:       delivery.Channel,
			Status:        delivery.Status,
			Provider:      delivery.Provider,
			Recipient:     delivery.Recipient,
			Attempts:      delivery.Attempts,
			LastAttemptAt: delivery.LastAttemptAt,
			NextRetryAt:   delivery.NextRetryAt,
			Proof:         firstNonEmpty(delivery.LastError, delivery.Summary, "Delivery attempt is visible in TraceDeck."),
		})
	}
	return proof
}

func alertDeliveryState(proof []model.TenantAlertDeliveryProof) string {
	if len(proof) == 0 {
		return constants.DeliveryStatusPending
	}
	hasDelivered := false
	for _, item := range proof {
		switch item.Status {
		case constants.DeliveryStatusFailed:
			return constants.DeliveryStatusFailed
		case constants.DeliveryStatusRetrying:
			return constants.DeliveryStatusRetrying
		case constants.DeliveryStatusDelivered:
			hasDelivered = true
		}
	}
	if hasDelivered {
		return constants.DeliveryStatusDelivered
	}
	return proof[0].Status
}

func alertInboxNextAction(event model.RiskEvent, proof []model.TenantAlertDeliveryProof) string {
	state := alertDeliveryState(proof)
	switch state {
	case constants.DeliveryStatusDelivered:
		return firstNonEmpty(event.Recommendation, "Review this alert with the customer and keep proof visible.")
	case constants.DeliveryStatusRetrying:
		return "Watch notification retry timing before escalating the alert."
	case constants.DeliveryStatusFailed:
		return "Fix the delivery provider route, then resend proof for this alert."
	default:
		return "Route this alert through email, push, or dashboard before customer review."
	}
}

func tenantAlertInboxSummary(items []model.TenantAlertInboxItem, sourceHostCount int) model.TenantAlertInboxSummary {
	summary := model.TenantAlertInboxSummary{
		Total:           len(items),
		SourceHostCount: sourceHostCount,
	}
	for _, item := range items {
		if item.Status == constants.RiskStatusOpen {
			summary.Open++
		}
		if severityRank(item.Severity) >= severityRank(constants.SeverityHigh) {
			summary.HighOrCritical++
		}
		hasEmail := false
		hasPush := false
		hasDashboard := false
		for _, proof := range item.DeliveryProof {
			switch proof.Channel {
			case constants.DeliveryChannelEmail:
				hasEmail = true
			case constants.DeliveryChannelPush:
				hasPush = true
			case constants.DeliveryChannelDashboard:
				hasDashboard = true
			}
			switch proof.Status {
			case constants.DeliveryStatusRetrying:
				summary.DeliveryRetrying++
			case constants.DeliveryStatusFailed:
				summary.DeliveryFailed++
			}
		}
		if hasEmail {
			summary.WithEmail++
		}
		if hasPush {
			summary.WithPush++
		}
		if hasDashboard {
			summary.WithDashboard++
		}
	}
	if summary.Total > 0 {
		summary.NotificationReady = ((summary.WithEmail + summary.WithPush + summary.WithDashboard) * 100) / (summary.Total * 3)
	}
	return summary
}

func tenantValuePanels(operations model.TenantOperationsSummary, plan model.Plan, trustScore int) []model.TenantValuePanel {
	return []model.TenantValuePanel{
		{
			Title:    "Anomaly Notifications",
			Metric:   fmt.Sprintf("%d active", operations.OpenAnomalies+operations.OpenPolicyViolations),
			Detail:   "Policy, non-study YouTube, risky software, and media playback signals are routed into customer actions.",
			Status:   statusFromCount(operations.OpenAnomalies + operations.OpenPolicyViolations),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			Title:    "Mail Delivery",
			Metric:   fmt.Sprintf("%d delivered", operations.EmailDelivered),
			Detail:   "Critical alert and weekly report email proof is visible for customer trust.",
			Status:   deliveryValueStatus(operations.EmailDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			Title:    "Push Notification",
			Metric:   fmt.Sprintf("%d delivered", operations.PushDelivered),
			Detail:   "Mobile/web push routing makes anomalies feel immediate and premium.",
			Status:   deliveryValueStatus(operations.PushDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			Title:    "Archive And Retention",
			Metric:   fmt.Sprintf("%d backlog", operations.ArchiveBacklog),
			Detail:   "S3 lifecycle readiness supports Family Pro, school, and business retention packaging.",
			Status:   archiveValueStatus(operations.ArchiveBacklog, plan.CloudArchive),
			PaidTier: constants.PlanSchool,
		},
		{
			Title:    "Trust And Audit",
			Metric:   fmt.Sprintf("%d%%", trustScore),
			Detail:   "Visible monitoring, audit events, policy changes, exports, and delete workflows support legitimate rollout.",
			Status:   scoreStatus(trustScore),
			PaidTier: constants.PlanBusiness,
		},
	}
}

func tenantPaidCapabilities(operations model.TenantOperationsSummary, rulesCount int, groupsCount int, assignmentsCount int, auditCount int, exportsCount int) []model.TenantPaidCapability {
	return []model.TenantPaidCapability{
		{
			Name:     "Weekly AI report",
			Status:   constants.StatusHealthy,
			Tier:     constants.PlanFamilyPro,
			Evidence: "Generated report and PDF route are available from host overview.",
		},
		{
			Name:     "Alert rules builder",
			Status:   countStatus(rulesCount),
			Tier:     constants.PlanFamilyPro,
			Evidence: fmt.Sprintf("%d saved alert rules", rulesCount),
		},
		{
			Name:     "Role-based dashboard",
			Status:   constants.StatusHealthy,
			Tier:     constants.PlanSchool,
			Evidence: "Parent, student, school admin, and business manager views are modeled.",
		},
		{
			Name:     "Managed rollout",
			Status:   countStatus(groupsCount + assignmentsCount),
			Tier:     constants.PlanSchool,
			Evidence: fmt.Sprintf("%d groups and %d assignments", groupsCount, assignmentsCount),
		},
		{
			Name:     "Notification proof",
			Status:   scoreStatus(operations.NotificationScore),
			Tier:     constants.PlanFamilyPro,
			Evidence: fmt.Sprintf("%d/%d routes delivered", operations.DeliveryDelivered, operations.DeliveryTotal),
		},
		{
			Name:     "Compliance export",
			Status:   countStatus(exportsCount + auditCount),
			Tier:     constants.PlanBusiness,
			Evidence: fmt.Sprintf("%d exports and %d audit events", exportsCount, auditCount),
		},
	}
}

func tenantConversionActions(operations model.TenantOperationsSummary, tenant model.Tenant, plan model.Plan, trustScore int, observedAt time.Time) []model.TenantOperationsSignal {
	actions := make([]model.TenantOperationsSignal, 0, 5)
	if operations.PushDelivered == 0 {
		actions = append(actions, model.TenantOperationsSignal{
			Title:      "Finish push notification proof",
			Detail:     "A delivered push route makes anomaly monitoring feel immediate in Family Pro demos.",
			Severity:   constants.SeverityMedium,
			Channel:    constants.DeliveryChannelPush,
			Status:     constants.StatusWatch,
			Owner:      constants.RoleParent,
			ObservedAt: observedAt,
		})
	}
	if operations.EmailDelivered == 0 {
		actions = append(actions, model.TenantOperationsSignal{
			Title:      "Send email proof",
			Detail:     "Send one critical alert or weekly report email before pitching paid monitoring.",
			Severity:   constants.SeverityHigh,
			Channel:    constants.DeliveryChannelEmail,
			Status:     constants.StatusAttention,
			Owner:      constants.RoleParent,
			ObservedAt: observedAt,
		})
	}
	if operations.ArchiveBacklog > 0 {
		actions = append(actions, model.TenantOperationsSignal{
			Title:      "Clear archive backlog story",
			Detail:     "Show retry behavior and S3 lifecycle policy so archive retention looks reliable.",
			Severity:   constants.SeverityLow,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusWatch,
			Owner:      constants.PlanSchool,
			ObservedAt: observedAt,
		})
	}
	if trustScore < 80 {
		actions = append(actions, model.TenantOperationsSignal{
			Title:      "Strengthen consent and audit proof",
			Detail:     "Keep collection disclosure, recipients, policy changes, exports, and delete workflows visible.",
			Severity:   constants.SeverityMedium,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusWatch,
			Owner:      constants.RoleBusinessManager,
			ObservedAt: observedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantOperationsSignal{
			Title:      "Ready for paid-plan demo",
			Detail:     fmt.Sprintf("%s has notification, archive, report, and dashboard proof for %s packaging.", tenant.Name, plan.Name),
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			ObservedAt: observedAt,
		})
	}
	return actions
}

func buildTenantNotificationCommandCenter(
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	inbox model.TenantAlertInbox,
	drilldown model.TenantDeliveryDrilldown,
	remediation model.TenantDeliveryRemediation,
	generatedAt time.Time,
) model.TenantNotificationCommandCenter {
	summary := model.TenantNotificationCommandCenterSummary{
		Status:                 notificationCommandStatus(operations, inbox, remediation),
		Headline:               notificationCommandHeadline(inbox, operations, monetization),
		NotificationScore:      operations.NotificationScore,
		MonetizationReadiness:  operations.MonetizationReadiness,
		TrustScore:             monetization.TrustScore,
		OpenAlerts:             inbox.Summary.Open,
		HighPriorityAlerts:     inbox.Summary.HighOrCritical,
		PolicyViolations:       operations.OpenPolicyViolations,
		Anomalies:              operations.OpenAnomalies,
		TamperSignals:          operations.TamperSignals,
		EmailDelivered:         operations.EmailDelivered,
		PushDelivered:          operations.PushDelivered,
		DashboardDelivered:     operations.DashboardDelivered,
		DeliveryFailed:         operations.DeliveryFailed,
		DeliveryRetrying:       operations.DeliveryRetrying,
		RoutesTotal:            drilldown.Summary.RoutesTotal,
		RoutesNeedingProof:     drilldown.Summary.RoutesNeedingProof,
		RemediationOpen:        remediation.Summary.ProblemsOpen,
		RemediationPlanned:     remediation.Summary.PlannedActions,
		RemediationSLAWatch:    remediation.Summary.SLAWatch,
		WeeklyReportReady:      true,
		ArchiveBacklog:         operations.ArchiveBacklog,
		RecommendedPaidPackage: firstNonEmpty(monetization.PlanName, operations.PlanName, constants.PlanFamilyPro),
	}

	return model.TenantNotificationCommandCenter{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        monetization.Audience,
		Summary:         summary,
		Channels:        notificationCommandChannels(drilldown.Routes),
		Alerts:          notificationCommandAlerts(inbox.Items),
		Actions:         notificationCommandActions(inbox.Items, remediation.Actions, monetization.ConversionActions, generatedAt),
		PrivacyBoundary: constants.NotificationCommandPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func buildTenantBusinessDashboard(
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	inbox model.TenantAlertInbox,
	commandCenter model.TenantNotificationCommandCenter,
	preferences model.NotificationPreferenceCenter,
	drilldown model.TenantDeliveryDrilldown,
	remediation model.TenantDeliveryRemediation,
	generatedAt time.Time,
) model.TenantBusinessDashboard {
	productScore := averageScore(operations.MonetizationReadiness, operations.NotificationScore, monetization.TrustScore, preferences.Summary.PreferenceScore)
	status := businessDashboardStatus(operations, inbox, drilldown, remediation, productScore)
	headline, detail := businessDashboardNarrative(operations, monetization, inbox, drilldown, status)
	summary := model.TenantBusinessDashboardSummary{
		Status:             status,
		Headline:           headline,
		Detail:             detail,
		ProductScore:       productScore,
		CustomerHealth:     operations.CustomerHealth,
		RevenueStage:       monetization.ConversionStage,
		RecommendedPackage: firstNonEmpty(commandCenter.Summary.RecommendedPaidPackage, monetization.PlanName, operations.PlanName),
		HostsTotal:         operations.HostsTotal,
		HostsAttention:     operations.HostsAttention,
		OpenAlerts:         inbox.Summary.Open,
		HighPriorityAlerts: inbox.Summary.HighOrCritical,
		NotificationScore:  operations.NotificationScore,
		PreferenceScore:    preferences.Summary.PreferenceScore,
		TrustScore:         monetization.TrustScore,
		MailDelivered:      operations.EmailDelivered,
		PushDelivered:      operations.PushDelivered,
		DashboardDelivered: operations.DashboardDelivered,
		RoutesNeedingProof: drilldown.Summary.RoutesNeedingProof,
		ArchiveBacklog:     operations.ArchiveBacklog,
		WeeklyReportReady:  commandCenter.Summary.WeeklyReportReady,
		ConsentVisible:     true,
		DataRightsReady:    monetization.TrustScore >= 65,
	}

	return model.TenantBusinessDashboard{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        monetization.Audience,
		Summary:         summary,
		Metrics:         businessDashboardMetrics(summary, operations, monetization, preferences, drilldown),
		Alerts:          businessDashboardAlerts(commandCenter.Alerts),
		Channels:        businessDashboardChannels(commandCenter.Channels),
		Packages:        businessDashboardPackages(monetization),
		Actions:         businessDashboardActions(commandCenter.Actions, remediation.Actions, monetization.ConversionActions, generatedAt),
		PrivacyBoundary: constants.BusinessDashboardPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func businessDashboardStatus(operations model.TenantOperationsSummary, inbox model.TenantAlertInbox, drilldown model.TenantDeliveryDrilldown, remediation model.TenantDeliveryRemediation, productScore int) string {
	switch {
	case operations.DeliveryFailed > 0 || inbox.Summary.HighOrCritical > 0:
		return constants.StatusAttention
	case drilldown.Summary.RoutesNeedingProof > 0 || remediation.Summary.ProblemsOpen > 0 || operations.HostsAttention > 0:
		return constants.StatusWatch
	case productScore >= 75:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func businessDashboardNarrative(operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, inbox model.TenantAlertInbox, drilldown model.TenantDeliveryDrilldown, status string) (string, string) {
	if len(inbox.Items) > 0 {
		top := inbox.Items[0]
		return fmt.Sprintf("%s needs %s notification assurance", top.Title, top.Severity),
			fmt.Sprintf("%s on %s has email, push, and dashboard status visible for paid customer review.", top.Category, top.HostName)
	}
	if drilldown.Summary.RoutesNeedingProof > 0 {
		return fmt.Sprintf("%d notification routes need proof before paid demo", drilldown.Summary.RoutesNeedingProof),
			"Mail, push, dashboard, and weekly report routes must show provider-safe evidence before onboarding."
	}
	if operations.ArchiveBacklog > 0 {
		return "Archive backlog is the current trust story",
			"Show retry behavior, S3 lifecycle readiness, and owner action proof before selling longer retention."
	}
	if status == constants.StatusHealthy {
		return fmt.Sprintf("%s is ready for a paid product demo", firstNonEmpty(monetization.PlanName, operations.PlanName, "TraceDeck")),
			"Anomaly alerts, notification delivery, archive retention, weekly reports, and trust proof are visible in one cockpit."
	}
	return "Business dashboard is waiting for stronger proof",
		"Enroll hosts, verify notification routes, and keep consent, archive, and report evidence current."
}

func businessDashboardMetrics(summary model.TenantBusinessDashboardSummary, operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, preferences model.NotificationPreferenceCenter, drilldown model.TenantDeliveryDrilldown) []model.TenantBusinessDashboardMetric {
	return []model.TenantBusinessDashboardMetric{
		{
			ID:       "customer-health",
			Label:    "Customer Health",
			Value:    titleWord(summary.CustomerHealth),
			Detail:   fmt.Sprintf("%d/%d hosts need attention", summary.HostsAttention, summary.HostsTotal),
			Status:   summary.CustomerHealth,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "anomaly-alerts",
			Label:    "Anomaly Alerts",
			Value:    fmt.Sprintf("%d open", summary.OpenAlerts),
			Detail:   fmt.Sprintf("%d high-priority signals with delivery proof", summary.HighPriorityAlerts),
			Status:   countStatus(summary.OpenAlerts),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "mail-delivery",
			Label:    "Mail Delivery",
			Value:    fmt.Sprintf("%d delivered", summary.MailDelivered),
			Detail:   monetization.NotificationPromise.Email,
			Status:   deliveryValueStatus(summary.MailDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "push-reach",
			Label:    "Push Reach",
			Value:    fmt.Sprintf("%d delivered", summary.PushDelivered),
			Detail:   monetization.NotificationPromise.Push,
			Status:   deliveryValueStatus(summary.PushDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "route-proof",
			Label:    "Route Proof",
			Value:    fmt.Sprintf("%d/%d ready", drilldown.Summary.HealthyRoutes, drilldown.Summary.RoutesTotal),
			Detail:   fmt.Sprintf("%d routes need proof", summary.RoutesNeedingProof),
			Status:   scoreStatus(operations.NotificationScore),
			PaidTier: constants.PlanBusiness,
		},
		{
			ID:       "preference-policy",
			Label:    "Preference Policy",
			Value:    fmt.Sprintf("%d%%", summary.PreferenceScore),
			Detail:   fmt.Sprintf("%d typed rules, %s digest", preferences.Summary.RulesTotal, preferences.DigestCadence),
			Status:   scoreStatus(summary.PreferenceScore),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "archive-report",
			Label:    "Archive And Reports",
			Value:    boolReady(summary.WeeklyReportReady),
			Detail:   fmt.Sprintf("%d archive batches waiting", summary.ArchiveBacklog),
			Status:   archiveValueStatus(summary.ArchiveBacklog, true),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "paid-stage",
			Label:    "Paid Stage",
			Value:    titleWord(summary.RevenueStage),
			Detail:   fmt.Sprintf("%d%% readiness, %d%% trust", monetization.ReadinessScore, monetization.TrustScore),
			Status:   monetization.RevenueHealth,
			PaidTier: constants.PlanBusiness,
		},
	}
}

func businessDashboardAlerts(alerts []model.TenantNotificationCommandCenterAlert) []model.TenantBusinessDashboardAlert {
	items := make([]model.TenantBusinessDashboardAlert, 0, len(alerts))
	for _, alert := range alerts {
		items = append(items, model.TenantBusinessDashboardAlert{
			ID:              alert.ID,
			Title:           alert.Title,
			Detail:          firstNonEmpty(alert.Detail, alert.Recommendation),
			Severity:        alert.Severity,
			Status:          alert.Status,
			HostName:        alert.HostName,
			Category:        alert.Category,
			EmailStatus:     alert.EmailStatus,
			PushStatus:      alert.PushStatus,
			DashboardStatus: alert.DashboardStatus,
			NextAction:      alert.NextAction,
			PaidTier:        alert.PaidTier,
			ObservedAt:      alert.ObservedAt,
		})
	}
	if len(items) > 6 {
		return items[:6]
	}
	return items
}

func businessDashboardChannels(channels []model.TenantNotificationCommandCenterChannel) []model.TenantBusinessDashboardChannel {
	items := make([]model.TenantBusinessDashboardChannel, 0, len(channels))
	for _, channel := range channels {
		status := firstNonEmpty(channel.LatestDeliveryStatus, channel.RouteStatus)
		items = append(items, model.TenantBusinessDashboardChannel{
			Channel:        channel.Channel,
			Provider:       channel.Provider,
			Status:         status,
			ProofState:     channel.ProofState,
			RecipientLabel: channel.Recipient,
			Attempts:       channel.Attempts,
			LastDeliveryAt: channel.LastDeliveryAt,
			NextAction:     channel.NextAction,
			PaidTier:       channel.PaidTier,
		})
	}
	return items
}

func businessDashboardPackages(monetization model.TenantMonetizationSummary) []model.TenantBusinessDashboardPackage {
	plans := []model.Plan{planByID(constants.PlanFamilyPro), planByID(constants.PlanSchool), planByID(constants.PlanBusiness)}
	packages := make([]model.TenantBusinessDashboardPackage, 0, len(plans))
	for _, plan := range plans {
		status := constants.StatusWatch
		nextAction := "Keep collecting product proof before upgrade."
		if plan.ID == monetization.PlanID {
			status = constants.StatusHealthy
			nextAction = "Use current tenant evidence as the paid package proof."
		}
		packages = append(packages, model.TenantBusinessDashboardPackage{
			Name:       plan.Name,
			Tier:       plan.ID,
			Audience:   plan.Audience,
			PriceModel: plan.PriceModel,
			Status:     status,
			Included:   append([]string(nil), plan.Features...),
			Value:      businessPackageValue(plan.ID),
			NextAction: nextAction,
		})
	}
	return packages
}

func businessDashboardActions(commandActions []model.TenantNotificationCommandCenterAction, remedies []model.TenantDeliveryRemediationAction, conversion []model.TenantOperationsSignal, generatedAt time.Time) []model.TenantBusinessDashboardAction {
	actions := make([]model.TenantBusinessDashboardAction, 0, 8)
	for _, action := range commandActions {
		if len(actions) >= 4 {
			break
		}
		actions = append(actions, model.TenantBusinessDashboardAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Severity:   action.Severity,
			Status:     action.Status,
			Owner:      action.Owner,
			Channel:    action.Channel,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     "notification command center",
			ObservedAt: action.ObservedAt,
		})
	}
	for _, remedy := range remedies {
		if len(actions) >= 6 {
			break
		}
		if remedy.Status == constants.DeliveryRemediationStatusHealthy {
			continue
		}
		actions = append(actions, model.TenantBusinessDashboardAction{
			Title:      titleWord(remedy.Channel) + " delivery route",
			Detail:     firstNonEmpty(remedy.Plan, remedy.Problem),
			Severity:   constants.SeverityMedium,
			Status:     remedy.Status,
			Owner:      remedy.Owner,
			Channel:    remedy.Channel,
			SLA:        remedy.SLATarget,
			PaidTier:   notificationCommandChannelTier(remedy.Channel),
			Source:     "delivery remediation",
			ObservedAt: remedy.CreatedAt,
		})
	}
	for _, signal := range conversion {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantBusinessDashboardAction{
			Title:      signal.Title,
			Detail:     signal.Detail,
			Severity:   signal.Severity,
			Status:     signal.Status,
			Owner:      signal.Owner,
			Channel:    signal.Channel,
			SLA:        "before paid review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "monetisation summary",
			ObservedAt: signal.ObservedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantBusinessDashboardAction{
			Title:      "Business dashboard ready",
			Detail:     "Customer health, anomaly alerts, mail delivery, push reach, archive retention, and paid packaging are visible.",
			Severity:   constants.SeverityInfo,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			Channel:    constants.DeliveryChannelDashboard,
			SLA:        "weekly review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "business dashboard",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func averageScore(scores ...int) int {
	total := 0
	count := 0
	for _, score := range scores {
		if score < 0 {
			continue
		}
		if score > 100 {
			score = 100
		}
		total += score
		count++
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func boolReady(ready bool) string {
	if ready {
		return "ready"
	}
	return "pending"
}

func businessPackageValue(planID string) string {
	switch planID {
	case constants.PlanFamilyPro:
		return "Family-ready weekly reports, anomaly notifications, role views, and cloud archive proof."
	case constants.PlanSchool:
		return "School rollout packaging with cohorts, audit history, retention controls, and admin workflows."
	case constants.PlanBusiness:
		return "Endpoint productivity, risky software, delivery assurance, and compliance export value."
	default:
		return "Paid packaging proof for customer review."
	}
}

func notificationCommandStatus(operations model.TenantOperationsSummary, inbox model.TenantAlertInbox, remediation model.TenantDeliveryRemediation) string {
	switch {
	case operations.DeliveryFailed > 0 || inbox.Summary.HighOrCritical > 0:
		return constants.StatusAttention
	case remediation.Summary.ProblemsOpen > 0 || operations.DeliveryRetrying > 0 || inbox.Summary.Open > 0:
		return constants.StatusWatch
	case operations.NotificationScore >= 70:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func notificationCommandHeadline(inbox model.TenantAlertInbox, operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary) string {
	if len(inbox.Items) > 0 {
		top := inbox.Items[0]
		return fmt.Sprintf("%s on %s needs %s delivery proof", top.Title, top.HostName, top.Severity)
	}
	if operations.DeliveryFailed > 0 || operations.DeliveryRetrying > 0 {
		return fmt.Sprintf("%d notification routes need delivery assurance", operations.DeliveryFailed+operations.DeliveryRetrying)
	}
	return fmt.Sprintf("%s notification proof is ready for %s", firstNonEmpty(monetization.PlanName, operations.PlanName, "TraceDeck"), firstNonEmpty(monetization.Audience, "paid customer"))
}

func notificationCommandChannels(routes []model.TenantDeliveryDrilldownRoute) []model.TenantNotificationCommandCenterChannel {
	channels := make([]model.TenantNotificationCommandCenterChannel, 0, len(routes))
	for _, route := range routes {
		channels = append(channels, model.TenantNotificationCommandCenterChannel{
			Channel:              route.Channel,
			Provider:             route.Provider,
			Recipient:            route.RecipientLabel,
			Enabled:              route.Enabled,
			RouteStatus:          route.RouteStatus,
			ProofState:           route.ProofState,
			LatestDeliveryStatus: route.LatestDeliveryStatus,
			Attempts:             route.Attempts,
			LastDeliveryAt:       route.LatestDeliveryAt,
			NextRetryAt:          nil,
			SLA:                  route.SLA,
			Evidence:             route.Evidence,
			NextAction:           route.NextAction,
			PaidTier:             notificationCommandChannelTier(route.Channel),
		})
	}
	return channels
}

func notificationCommandAlerts(items []model.TenantAlertInboxItem) []model.TenantNotificationCommandCenterAlert {
	alerts := make([]model.TenantNotificationCommandCenterAlert, 0, len(items))
	for _, item := range items {
		alerts = append(alerts, model.TenantNotificationCommandCenterAlert{
			ID:              item.ID,
			EventID:         item.EventID,
			DeviceID:        item.DeviceID,
			HostName:        item.HostName,
			Type:            item.Type,
			Severity:        item.Severity,
			Category:        item.Category,
			Status:          item.Status,
			Title:           item.Title,
			Detail:          item.Detail,
			Recommendation:  item.Recommendation,
			DeliveryState:   item.DeliveryState,
			EmailStatus:     alertDeliveryProofStatus(item.DeliveryProof, constants.DeliveryChannelEmail),
			PushStatus:      alertDeliveryProofStatus(item.DeliveryProof, constants.DeliveryChannelPush),
			DashboardStatus: alertDeliveryProofStatus(item.DeliveryProof, constants.DeliveryChannelDashboard),
			NextAction:      item.NextAction,
			PaidTier:        notificationCommandAlertTier(item),
			ObservedAt:      item.ObservedAt,
		})
	}
	if len(alerts) > 8 {
		return alerts[:8]
	}
	return alerts
}

func notificationCommandActions(
	alerts []model.TenantAlertInboxItem,
	remedies []model.TenantDeliveryRemediationAction,
	conversion []model.TenantOperationsSignal,
	generatedAt time.Time,
) []model.TenantNotificationCommandCenterAction {
	actions := make([]model.TenantNotificationCommandCenterAction, 0, 8)
	for _, alert := range alerts {
		if len(actions) >= 3 {
			break
		}
		actions = append(actions, model.TenantNotificationCommandCenterAction{
			Title:      "Triage " + alert.Title,
			Detail:     firstNonEmpty(alert.NextAction, alert.Recommendation, alert.Detail),
			Severity:   alert.Severity,
			Channel:    notificationCommandPreferredChannel(alert),
			Status:     alert.Status,
			Owner:      alert.HostName,
			SLA:        notificationCommandSLA(alert.Severity),
			PaidTier:   notificationCommandAlertTier(alert),
			ObservedAt: alert.ObservedAt,
		})
	}
	for _, remedy := range remedies {
		if len(actions) >= 6 {
			break
		}
		if remedy.Status == constants.DeliveryRemediationStatusHealthy {
			continue
		}
		actions = append(actions, model.TenantNotificationCommandCenterAction{
			Title:      titleWord(remedy.Channel) + " delivery assurance",
			Detail:     firstNonEmpty(remedy.Plan, remedy.Problem, "Keep route proof current for paid notification SLAs."),
			Severity:   constants.SeverityMedium,
			Channel:    remedy.Channel,
			Status:     remedy.Status,
			Owner:      remedy.Owner,
			SLA:        remedy.SLATarget,
			PaidTier:   notificationCommandChannelTier(remedy.Channel),
			ObservedAt: remedy.CreatedAt,
		})
	}
	for _, signal := range conversion {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantNotificationCommandCenterAction{
			Title:      signal.Title,
			Detail:     signal.Detail,
			Severity:   signal.Severity,
			Channel:    signal.Channel,
			Status:     signal.Status,
			Owner:      signal.Owner,
			SLA:        "before paid demo",
			PaidTier:   constants.PlanFamilyPro,
			ObservedAt: signal.ObservedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantNotificationCommandCenterAction{
			Title:      "Notification command center ready",
			Detail:     "Anomaly alerts, email proof, push reach, dashboard inbox, and delivery assurance are visible.",
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			SLA:        "review weekly",
			PaidTier:   constants.PlanFamilyPro,
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func alertDeliveryProofStatus(proof []model.TenantAlertDeliveryProof, channel string) string {
	for _, item := range proof {
		if item.Channel == channel {
			return item.Status
		}
	}
	return constants.DeliveryStatusPending
}

func notificationCommandChannelTier(channel string) string {
	switch channel {
	case constants.DeliveryChannelEmail, constants.DeliveryChannelPush:
		return constants.PlanFamilyPro
	case constants.DeliveryChannelDashboard:
		return constants.PlanFree
	default:
		return constants.PlanBusiness
	}
}

func notificationCommandAlertTier(item model.TenantAlertInboxItem) string {
	switch item.Category {
	case constants.RiskCategoryRiskySoftware, constants.RiskCategoryAgentHealth, constants.RiskCategoryArchiveHealth:
		return constants.PlanBusiness
	case constants.RiskCategoryPolicyChange:
		return constants.PlanSchool
	default:
		return constants.PlanFamilyPro
	}
}

func notificationCommandPreferredChannel(item model.TenantAlertInboxItem) string {
	if alertDeliveryProofStatus(item.DeliveryProof, constants.DeliveryChannelEmail) != constants.DeliveryStatusPending {
		return constants.DeliveryChannelEmail
	}
	if alertDeliveryProofStatus(item.DeliveryProof, constants.DeliveryChannelPush) != constants.DeliveryStatusPending {
		return constants.DeliveryChannelPush
	}
	return constants.DeliveryChannelDashboard
}

func notificationCommandSLA(severity string) string {
	switch severity {
	case constants.SeverityCritical:
		return "15 minutes"
	case constants.SeverityHigh:
		return "1 hour"
	case constants.SeverityMedium:
		return "same day"
	default:
		return "weekly review"
	}
}

func statusFromCount(count int) string {
	if count > 0 {
		return constants.StatusHealthy
	}
	return constants.StatusWatch
}

func countStatus(count int) string {
	if count > 0 {
		return constants.StatusHealthy
	}
	return constants.StatusAttention
}

func deliveryValueStatus(delivered int) string {
	if delivered > 0 {
		return constants.StatusHealthy
	}
	return constants.StatusAttention
}

func archiveValueStatus(backlog int, cloudArchive bool) string {
	if !cloudArchive {
		return constants.StatusAttention
	}
	if backlog > 0 {
		return constants.StatusWatch
	}
	return constants.StatusHealthy
}

func scoreStatus(score int) string {
	switch {
	case score >= 75:
		return constants.StatusHealthy
	case score >= 50:
		return constants.StatusWatch
	default:
		return constants.StatusAttention
	}
}

func topTenantDeliveryProblem(deliveries []model.AlertDelivery) *model.AlertDelivery {
	var top *model.AlertDelivery
	for index := range deliveries {
		if deliveries[index].Status == constants.DeliveryStatusDelivered {
			continue
		}
		currentRank := deliveryProblemRank(deliveries[index].Status)
		topRank := 0
		if top != nil {
			topRank = deliveryProblemRank(top.Status)
		}
		if top == nil || currentRank > topRank || (currentRank == topRank && deliveries[index].LastAttemptAt.After(top.LastAttemptAt)) {
			current := deliveries[index]
			top = &current
		}
	}
	return top
}

func buildTenantDeliveryDrilldown(tenantID string, routes []model.NotificationRoute, deliveries []model.AlertDelivery, generatedAt time.Time, mode string) model.TenantDeliveryDrilldown {
	routes = append([]model.NotificationRoute(nil), routes...)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Channel != routes[j].Channel {
			return routes[i].Channel < routes[j].Channel
		}
		return routes[i].CreatedAt.Before(routes[j].CreatedAt)
	})

	items := make([]model.TenantDeliveryDrilldownRoute, 0, len(routes))
	summary := model.TenantDeliveryDrilldownSummary{
		RoutesTotal:   len(routes),
		RehearsalMode: mode,
	}
	actions := make([]model.TenantOperationsSignal, 0, len(routes))
	for _, route := range routes {
		latest := latestDeliveryForRoute(deliveries, route)
		item := deliveryDrilldownRoute(route, latest)
		items = append(items, item)

		if route.Enabled {
			summary.EnabledRoutes++
		}
		if route.Status == constants.StatusHealthy {
			summary.HealthyRoutes++
		}
		if item.ProofState != constants.DeliveryProofStateCustomer && item.ProofState != constants.DeliveryProofStateRehearsed {
			summary.RoutesNeedingProof++
			actions = append(actions, model.TenantOperationsSignal{
				Title:      titleWord(route.Channel) + " delivery proof needed",
				Detail:     item.NextAction,
				Severity:   constants.SeverityMedium,
				Channel:    route.Channel,
				Status:     item.RouteStatus,
				Owner:      route.RecipientLabel,
				ObservedAt: generatedAt,
			})
		}
		if route.LastVerifiedAt != nil {
			if summary.LastRehearsedAt == nil || route.LastVerifiedAt.After(*summary.LastRehearsedAt) {
				verifiedAt := *route.LastVerifiedAt
				summary.LastRehearsedAt = &verifiedAt
			}
		}
		switch route.Channel {
		case constants.DeliveryChannelEmail:
			summary.EmailReady = item.ProofState == constants.DeliveryProofStateCustomer || item.ProofState == constants.DeliveryProofStateRehearsed
		case constants.DeliveryChannelPush:
			summary.PushReady = item.ProofState == constants.DeliveryProofStateCustomer || item.ProofState == constants.DeliveryProofStateRehearsed
		case constants.DeliveryChannelDashboard:
			summary.DashboardReady = item.ProofState == constants.DeliveryProofStateCustomer || item.ProofState == constants.DeliveryProofStateRehearsed
		}
	}
	if summary.RoutesTotal > 0 {
		summary.DeliveryScore = (summary.HealthyRoutes * 100) / summary.RoutesTotal
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantOperationsSignal{
			Title:      "Delivery proof is rehearsal-ready",
			Detail:     "Email, push, and dashboard routes have content-safe proof for a paid demo.",
			Severity:   constants.SeverityInfo,
			Channel:    constants.DeliveryChannelDashboard,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			ObservedAt: generatedAt,
		})
	}
	return model.TenantDeliveryDrilldown{
		TenantID:        tenantID,
		GeneratedAt:     generatedAt,
		PrivacyBoundary: constants.DeliveryDrillPrivacyNote,
		Summary:         summary,
		Routes:          items,
		Actions:         actions,
	}
}

func deliveryDrilldownRoute(route model.NotificationRoute, delivery *model.AlertDelivery) model.TenantDeliveryDrilldownRoute {
	item := model.TenantDeliveryDrilldownRoute{
		RouteID:        route.ID,
		Channel:        route.Channel,
		Provider:       route.Provider,
		RecipientLabel: route.RecipientLabel,
		Enabled:        route.Enabled,
		RouteStatus:    route.Status,
		LastVerifiedAt: route.LastVerifiedAt,
		ProofState:     deliveryProofState(route, delivery),
		SLA:            deliveryDrilldownSLA(route.Channel),
		Evidence:       firstNonEmpty(route.LastSummary, "Route metadata is present without provider secrets."),
	}
	if delivery != nil {
		item.LatestDeliveryStatus = delivery.Status
		item.LatestDeliveryAt = &delivery.LastAttemptAt
		item.Attempts = delivery.Attempts
		item.Evidence = firstNonEmpty(delivery.Summary, route.LastSummary, "Latest delivery metadata is visible.")
	} else {
		item.LatestDeliveryStatus = constants.DeliveryStatusPending
	}
	item.RehearsalResult = deliveryRehearsalResult(route, delivery)
	item.NextAction = deliveryDrilldownNextAction(route, delivery, item.ProofState)
	return item
}

func latestDeliveryForRoute(deliveries []model.AlertDelivery, route model.NotificationRoute) *model.AlertDelivery {
	var latest *model.AlertDelivery
	for index := range deliveries {
		if deliveries[index].Channel != route.Channel {
			continue
		}
		if strings.TrimSpace(route.Provider) != "" && deliveries[index].Provider != route.Provider {
			continue
		}
		if latest == nil || deliveries[index].LastAttemptAt.After(latest.LastAttemptAt) {
			current := deliveries[index]
			latest = &current
		}
	}
	return latest
}

func deliveryProofState(route model.NotificationRoute, delivery *model.AlertDelivery) string {
	switch {
	case !route.Enabled:
		return constants.DeliveryProofStateDisabled
	case !deliveryProviderMatchesChannel(route.Provider, route.Channel):
		return constants.DeliveryProofStateMismatch
	case delivery != nil && delivery.Status == constants.DeliveryStatusDelivered:
		return constants.DeliveryProofStateCustomer
	case route.Status == constants.StatusHealthy && route.LastVerifiedAt != nil:
		return constants.DeliveryProofStateRehearsed
	case delivery != nil && (delivery.Status == constants.DeliveryStatusRetrying || delivery.Status == constants.DeliveryStatusFailed):
		return constants.DeliveryProofStateNeedsProvider
	default:
		return constants.DeliveryProofStateNeedsProof
	}
}

func deliveryRehearsalResult(route model.NotificationRoute, delivery *model.AlertDelivery) string {
	switch deliveryProofState(route, delivery) {
	case constants.DeliveryProofStateCustomer:
		return "latest metadata shows delivered route proof"
	case constants.DeliveryProofStateRehearsed:
		return "dry-run route rehearsal passed without provider payloads"
	case constants.DeliveryProofStateDisabled:
		return "route disabled; rehearsal skipped"
	case constants.DeliveryProofStateMismatch:
		return "provider/channel mismatch blocks rehearsal"
	case constants.DeliveryProofStateNeedsProvider:
		return "latest provider metadata needs retry or verification"
	default:
		return "dry-run rehearsal available"
	}
}

func deliveryDrilldownNextAction(route model.NotificationRoute, delivery *model.AlertDelivery, proofState string) string {
	switch proofState {
	case constants.DeliveryProofStateCustomer:
		return "Use latest delivery metadata as buyer proof."
	case constants.DeliveryProofStateRehearsed:
		return "Send a real proof notification when production provider credentials are configured."
	case constants.DeliveryProofStateDisabled:
		return "Enable this route before relying on it for anomaly alerts."
	case constants.DeliveryProofStateMismatch:
		return "Fix the provider/channel pairing before rehearsal."
	case constants.DeliveryProofStateNeedsProvider:
		if delivery != nil {
			return firstNonEmpty(delivery.LastError, delivery.Summary, "Review retry timing and provider status.")
		}
		return "Review retry timing and provider status."
	default:
		return "Run a dry-run delivery drilldown before a paid demo."
	}
}

func deliveryDrilldownSLA(channel string) string {
	switch channel {
	case constants.DeliveryChannelEmail:
		return "critical email proof within 5 minutes"
	case constants.DeliveryChannelPush:
		return "push proof within 60 seconds"
	case constants.DeliveryChannelDashboard:
		return "dashboard inbox proof immediately"
	default:
		return "route proof pending"
	}
}

func deliveryProviderMatchesChannel(provider string, channel string) bool {
	switch channel {
	case constants.DeliveryChannelEmail:
		return provider == constants.DeliveryProviderSMTP
	case constants.DeliveryChannelPush:
		return provider == constants.DeliveryProviderWebPush
	case constants.DeliveryChannelDashboard:
		return provider == constants.DeliveryProviderLocalFeed
	default:
		return false
	}
}

func buildTenantDeliveryRemediation(tenantID string, routes []model.NotificationRoute, deliveries []model.AlertDelivery, planned []model.TenantDeliveryRemediationAction, generatedAt time.Time) model.TenantDeliveryRemediation {
	routes = append([]model.NotificationRoute(nil), routes...)
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Channel != routes[j].Channel {
			return routes[i].Channel < routes[j].Channel
		}
		return routes[i].CreatedAt.Before(routes[j].CreatedAt)
	})

	actions := make([]model.TenantDeliveryRemediationAction, 0, len(routes))
	summary := model.TenantDeliveryRemediationSummary{RoutesTotal: len(routes)}
	for _, route := range routes {
		latest := latestDeliveryForRoute(deliveries, route)
		action := deliveryRemediationAction(route, latest, generatedAt)
		actions = append(actions, action)
		if action.Status != constants.DeliveryRemediationStatusHealthy {
			summary.ProblemsOpen++
		}
		if action.Action == constants.DeliveryRemediationActionSLAWatch {
			summary.SLAWatch++
		}
		if action.Status != constants.DeliveryRemediationStatusHealthy && action.NextRetryAt != nil && (summary.NextRetryAt == nil || action.NextRetryAt.Before(*summary.NextRetryAt)) {
			next := *action.NextRetryAt
			summary.NextRetryAt = &next
		}
		switch route.Channel {
		case constants.DeliveryChannelEmail:
			summary.EmailProtected = action.Status == constants.DeliveryRemediationStatusHealthy
		case constants.DeliveryChannelPush:
			summary.PushProtected = action.Status == constants.DeliveryRemediationStatusHealthy
		case constants.DeliveryChannelDashboard:
			summary.DashboardProtected = action.Status == constants.DeliveryRemediationStatusHealthy
		}
	}

	recent := cloneDeliveryRemediations(planned)
	sort.Slice(recent, func(i, j int) bool {
		return recent[i].CreatedAt.After(recent[j].CreatedAt)
	})
	for _, plan := range recent {
		summary.PlannedActions++
		if plan.Status == constants.DeliveryRemediationStatusAcked {
			summary.OwnerAcknowledged++
		}
		if summary.LastPlannedAt == nil || plan.CreatedAt.After(*summary.LastPlannedAt) {
			plannedAt := plan.CreatedAt
			summary.LastPlannedAt = &plannedAt
		}
	}
	if len(recent) > 6 {
		recent = recent[:6]
	}
	if summary.RoutesTotal > 0 {
		summary.RemediationScore = ((summary.RoutesTotal - summary.ProblemsOpen) * 100) / summary.RoutesTotal
	}

	return model.TenantDeliveryRemediation{
		TenantID:        tenantID,
		GeneratedAt:     generatedAt,
		PrivacyBoundary: constants.DeliveryRemediationPrivacyNote,
		Summary:         summary,
		Actions:         actions,
		RecentPlans:     recent,
	}
}

func deliveryRemediationAction(route model.NotificationRoute, delivery *model.AlertDelivery, generatedAt time.Time) model.TenantDeliveryRemediationAction {
	drilldown := deliveryDrilldownRoute(route, delivery)
	nextRetryAt := drilldown.LatestDeliveryAt
	if delivery != nil && delivery.NextRetryAt != nil {
		nextRetry := *delivery.NextRetryAt
		nextRetryAt = &nextRetry
	}
	action := model.TenantDeliveryRemediationAction{
		RouteID:              route.ID,
		Channel:              route.Channel,
		Provider:             route.Provider,
		RecipientLabel:       route.RecipientLabel,
		Action:               deliveryRemediationActionForProof(drilldown.ProofState),
		Status:               deliveryRemediationStatusForProof(drilldown.ProofState),
		Owner:                deliveryRemediationOwner(route),
		Problem:              drilldown.RehearsalResult,
		Plan:                 deliveryRemediationPlan(drilldown, delivery),
		SLATarget:            drilldown.SLA,
		LatestDeliveryStatus: drilldown.LatestDeliveryStatus,
		LatestDeliveryAt:     drilldown.LatestDeliveryAt,
		NextRetryAt:          nextRetryAt,
		AuditState:           constants.StatusPending,
		PrivacyBoundary:      constants.DeliveryRemediationPrivacyNote,
		CreatedAt:            generatedAt,
	}
	if action.NextRetryAt == nil && action.Status != constants.DeliveryRemediationStatusHealthy {
		nextRetry := deliveryRemediationNextRetry(route.Channel, generatedAt)
		action.NextRetryAt = &nextRetry
	}
	return action
}

func selectDeliveryRemediationRoute(routes []model.NotificationRoute, deliveries []model.AlertDelivery, routeID string, channel string) (model.NotificationRoute, *model.AlertDelivery) {
	var fallback *model.NotificationRoute
	for index := range routes {
		route := routes[index]
		if routeID != "" && route.ID != routeID {
			continue
		}
		if channel != "" && route.Channel != channel {
			continue
		}
		latest := latestDeliveryForRoute(deliveries, route)
		if deliveryProofState(route, latest) != constants.DeliveryProofStateCustomer && deliveryProofState(route, latest) != constants.DeliveryProofStateRehearsed {
			return route, latest
		}
		if fallback == nil {
			current := route
			fallback = &current
		}
	}
	if fallback != nil {
		return *fallback, latestDeliveryForRoute(deliveries, *fallback)
	}
	if len(routes) == 0 {
		return model.NotificationRoute{}, nil
	}
	return routes[0], latestDeliveryForRoute(deliveries, routes[0])
}

func deliveryRemediationActionForProof(proofState string) string {
	switch proofState {
	case constants.DeliveryProofStateCustomer, constants.DeliveryProofStateRehearsed:
		return constants.DeliveryRemediationActionMaintain
	case constants.DeliveryProofStateDisabled:
		return constants.DeliveryRemediationActionEnable
	case constants.DeliveryProofStateMismatch:
		return constants.DeliveryRemediationActionFix
	case constants.DeliveryProofStateNeedsProvider:
		return constants.DeliveryRemediationActionRetryPlan
	default:
		return constants.DeliveryRemediationActionRehearsal
	}
}

func deliveryRemediationStatusForProof(proofState string) string {
	switch proofState {
	case constants.DeliveryProofStateCustomer, constants.DeliveryProofStateRehearsed:
		return constants.DeliveryRemediationStatusHealthy
	default:
		return constants.DeliveryRemediationStatusOpen
	}
}

func deliveryRemediationStatusForAction(action string) string {
	switch action {
	case constants.DeliveryRemediationActionOwnerAck:
		return constants.DeliveryRemediationStatusAcked
	case constants.DeliveryRemediationActionMaintain:
		return constants.DeliveryRemediationStatusHealthy
	default:
		return constants.DeliveryRemediationStatusPlanned
	}
}

func deliveryRemediationOwner(route model.NotificationRoute) string {
	return firstNonEmpty(route.RecipientLabel, constants.RoleBusinessManager)
}

func deliveryRemediationPlan(route model.TenantDeliveryDrilldownRoute, delivery *model.AlertDelivery) string {
	if delivery != nil && strings.TrimSpace(delivery.LastError) != "" {
		return "Plan a provider-safe retry review for: " + delivery.LastError
	}
	return firstNonEmpty(route.NextAction, route.Evidence, "Plan dry-run verification before relying on this notification route.")
}

func deliveryRemediationNextRetry(channel string, generatedAt time.Time) time.Time {
	switch channel {
	case constants.DeliveryChannelPush:
		return generatedAt.Add(time.Minute)
	case constants.DeliveryChannelEmail:
		return generatedAt.Add(5 * time.Minute)
	case constants.DeliveryChannelDashboard:
		return generatedAt
	default:
		return generatedAt.Add(15 * time.Minute)
	}
}

func topTenantRiskEvents(events []model.RiskEvent, limit int) []model.RiskEvent {
	candidates := append([]model.RiskEvent(nil), events...)
	sort.Slice(candidates, func(i, j int) bool {
		statusDelta := riskStatusRank(candidates[j].Status) - riskStatusRank(candidates[i].Status)
		if statusDelta != 0 {
			return statusDelta < 0
		}
		severityDelta := severityRank(candidates[j].Severity) - severityRank(candidates[i].Severity)
		if severityDelta != 0 {
			return severityDelta < 0
		}
		return candidates[i].ObservedAt.After(candidates[j].ObservedAt)
	})
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func eventTitle(event model.RiskEvent) string {
	if strings.TrimSpace(event.AppName) != "" {
		return event.AppName
	}
	if strings.TrimSpace(event.Domain) != "" {
		return event.Domain
	}
	if strings.TrimSpace(event.ResourceLabel) != "" {
		return event.ResourceLabel
	}
	return event.Category
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func titleWord(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	return strings.ToUpper(value[:1]) + value[1:]
}

func deliveryProblemRank(status string) int {
	switch status {
	case constants.DeliveryStatusFailed:
		return 4
	case constants.DeliveryStatusRetrying:
		return 3
	case constants.DeliveryStatusPending:
		return 2
	case constants.DeliveryStatusSuppressed:
		return 1
	default:
		return 0
	}
}

func riskStatusRank(status string) int {
	switch status {
	case constants.RiskStatusOpen:
		return 3
	case constants.RiskStatusAcknowledged:
		return 2
	case constants.RiskStatusResolved:
		return 1
	default:
		return 0
	}
}

func severityRank(severity string) int {
	switch severity {
	case constants.SeverityCritical:
		return 5
	case constants.SeverityHigh:
		return 4
	case constants.SeverityMedium:
		return 3
	case constants.SeverityLow:
		return 2
	case constants.SeverityInfo:
		return 1
	default:
		return 0
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
