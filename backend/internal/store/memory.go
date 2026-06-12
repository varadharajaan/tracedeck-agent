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

func (m *Memory) TenantProviderSimulationLab(ctx context.Context, tenantID string) (model.TenantProviderSimulationLab, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantProviderSimulationLab{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantProviderSimulationLab{}, err
	}
	revenue, err := m.TenantNotificationRevenueCockpit(ctx, tenantID)
	if err != nil {
		return model.TenantProviderSimulationLab{}, err
	}
	drilldown, err := m.TenantDeliveryDrilldown(ctx, tenantID)
	if err != nil {
		return model.TenantProviderSimulationLab{}, err
	}
	remediation, err := m.TenantDeliveryRemediation(ctx, tenantID)
	if err != nil {
		return model.TenantProviderSimulationLab{}, err
	}

	return buildTenantProviderSimulationLab(operations, monetization, revenue, drilldown, remediation, generatedAt), nil
}

func (m *Memory) RunTenantProviderSimulation(ctx context.Context, tenantID string, req model.RunProviderSimulationRequest) (model.TenantProviderSimulationLab, error) {
	tenantID = strings.TrimSpace(tenantID)
	channel := strings.TrimSpace(req.Channel)
	_, err := m.RunTenantDeliveryDrilldown(ctx, tenantID, model.RunDeliveryDrilldownRequest{
		Mode:    constants.DeliveryDrillModeDryRun,
		Channel: channel,
		Reason:  firstNonEmpty(strings.TrimSpace(req.Reason), strings.TrimSpace(req.Scenario), "provider simulation lab dry run"),
	})
	if err != nil {
		return model.TenantProviderSimulationLab{}, err
	}

	now := time.Now().UTC()
	m.mu.Lock()
	if _, ok := m.tenants[tenantID]; !ok {
		m.mu.Unlock()
		return model.TenantProviderSimulationLab{}, ErrTenantNotFound
	}
	m.auditEvents = append(m.auditEvents, model.AuditEvent{
		ID:        auditID(tenantID, len(m.auditEvents)+1, now),
		TenantID:  tenantID,
		Category:  constants.AuditCategorySystem,
		Action:    constants.AuditActionProviderSimulation,
		Actor:     constants.AuditActorLocalAPI,
		ActorRole: constants.RoleBusinessManager,
		Summary:   fmt.Sprintf("provider simulation dry-run completed for %s routes", firstNonEmpty(channel, "all")),
		CreatedAt: now,
	})
	if err := m.persistLocked(); err != nil {
		m.mu.Unlock()
		return model.TenantProviderSimulationLab{}, err
	}
	m.mu.Unlock()

	return m.TenantProviderSimulationLab(ctx, tenantID)
}

func (m *Memory) TenantPackageBillingReadiness(ctx context.Context, tenantID string) (model.TenantPackageBillingReadiness, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantPackageBillingReadiness{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantPackageBillingReadiness{}, err
	}
	business, err := m.TenantBusinessDashboard(ctx, tenantID)
	if err != nil {
		return model.TenantPackageBillingReadiness{}, err
	}
	roles, err := m.TenantRoleExperiences(ctx, tenantID)
	if err != nil {
		return model.TenantPackageBillingReadiness{}, err
	}
	provider, err := m.TenantProviderSimulationLab(ctx, tenantID)
	if err != nil {
		return model.TenantPackageBillingReadiness{}, err
	}

	m.mu.RLock()
	tenant, ok := m.tenants[tenantID]
	m.mu.RUnlock()
	if !ok {
		return model.TenantPackageBillingReadiness{}, ErrTenantNotFound
	}

	return buildTenantPackageBillingReadiness(tenant, retentionTierByID(tenant.RetentionTierID), operations, monetization, business, roles, provider, generatedAt), nil
}

func (m *Memory) TenantCustomerControlRoom(ctx context.Context, tenantID string) (model.TenantCustomerControlRoom, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerControlRoom{}, err
	}
	business, err := m.TenantBusinessDashboard(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerControlRoom{}, err
	}
	executive, err := m.TenantExecutiveConsole(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerControlRoom{}, err
	}
	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerControlRoom{}, err
	}
	provider, err := m.TenantProviderSimulationLab(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerControlRoom{}, err
	}

	return buildTenantCustomerControlRoom(operations, business, executive, packageBilling, provider, generatedAt), nil
}

func (m *Memory) TenantCustomerSuccessPacket(ctx context.Context, tenantID string) (model.TenantCustomerSuccessPacket, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	controlRoom, err := m.TenantCustomerControlRoom(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSuccessPacket{}, err
	}
	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSuccessPacket{}, err
	}
	provider, err := m.TenantProviderSimulationLab(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSuccessPacket{}, err
	}
	roles, err := m.TenantRoleExperiences(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSuccessPacket{}, err
	}

	return buildTenantCustomerSuccessPacket(controlRoom, packageBilling, provider, roles, generatedAt), nil
}

func (m *Memory) TenantPushActivationCenter(ctx context.Context, tenantID string) (model.TenantPushActivationCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	preferences, err := m.TenantNotificationPreferences(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	drilldown, err := m.TenantDeliveryDrilldown(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	provider, err := m.TenantProviderSimulationLab(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	remediation, err := m.TenantDeliveryRemediation(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	inbox, err := m.TenantAlertInbox(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	timeline, err := m.TenantDeliveryTimeline(ctx, tenantID, model.TenantDeliveryTimelineFilter{
		Channel: constants.DeliveryChannelPush,
		Limit:   20,
	})
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}
	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantPushActivationCenter{}, err
	}

	return buildTenantPushActivationCenter(operations, preferences, drilldown, provider, remediation, inbox, timeline, packageBilling, generatedAt), nil
}

func (m *Memory) TenantPortfolioCenter(ctx context.Context, tenantID string) (model.TenantPortfolioCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantPortfolioCenter{}, err
	}
	business, err := m.TenantBusinessDashboard(ctx, tenantID)
	if err != nil {
		return model.TenantPortfolioCenter{}, err
	}
	syncHealth, err := m.TenantSyncHealth(ctx, tenantID)
	if err != nil {
		return model.TenantPortfolioCenter{}, err
	}
	inbox, err := m.TenantAlertInbox(ctx, tenantID)
	if err != nil {
		return model.TenantPortfolioCenter{}, err
	}
	timeline, err := m.TenantDeliveryTimeline(ctx, tenantID, model.TenantDeliveryTimelineFilter{Limit: 30})
	if err != nil {
		return model.TenantPortfolioCenter{}, err
	}
	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantPortfolioCenter{}, err
	}

	devices := m.ListDevices(ctx)
	hosts := make([]model.TenantPortfolioHost, 0)
	for _, device := range devices {
		if device.TenantID != tenantID {
			continue
		}
		overview, err := m.HostOverview(ctx, device.DeviceID)
		if err != nil {
			return model.TenantPortfolioCenter{}, err
		}
		hosts = append(hosts, tenantPortfolioHost(overview, timeline.Items))
	}
	sort.Slice(hosts, func(i, j int) bool {
		leftRank := statusRankValue(hosts[i].Status)
		rightRank := statusRankValue(hosts[j].Status)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if hosts[i].RiskScore != hosts[j].RiskScore {
			return hosts[i].RiskScore > hosts[j].RiskScore
		}
		return hosts[i].HostName < hosts[j].HostName
	})

	return buildTenantPortfolioCenter(operations, business, syncHealth, inbox, timeline, packageBilling, hosts, generatedAt), nil
}

func (m *Memory) AccountPortfolioIndex(ctx context.Context, tenantIDs []string) (model.AccountPortfolioIndex, error) {
	generatedAt := time.Now().UTC()
	allowed := make(map[string]struct{}, len(tenantIDs))
	for _, tenantID := range tenantIDs {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID != "" {
			allowed[tenantID] = struct{}{}
		}
	}
	filtered := len(allowed) > 0

	tenants := m.ListTenants(ctx)
	sort.Slice(tenants, func(i, j int) bool {
		return tenants[i].Name < tenants[j].Name
	})

	centers := make([]model.TenantPortfolioCenter, 0, len(tenants))
	for _, tenant := range tenants {
		if filtered {
			if _, ok := allowed[tenant.TenantID]; !ok {
				continue
			}
		}
		center, err := m.TenantPortfolioCenter(ctx, tenant.TenantID)
		if err != nil {
			return model.AccountPortfolioIndex{}, err
		}
		centers = append(centers, center)
	}

	return buildAccountPortfolioIndex(centers, generatedAt), nil
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

func (m *Memory) TenantRoleExperiences(ctx context.Context, tenantID string) (model.TenantRoleExperience, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantRoleExperience{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantRoleExperience{}, err
	}
	business, err := m.TenantBusinessDashboard(ctx, tenantID)
	if err != nil {
		return model.TenantRoleExperience{}, err
	}
	preferences, err := m.TenantNotificationPreferences(ctx, tenantID)
	if err != nil {
		return model.TenantRoleExperience{}, err
	}
	timeline, err := m.TenantDeliveryTimeline(ctx, tenantID, model.TenantDeliveryTimelineFilter{Limit: 10})
	if err != nil {
		return model.TenantRoleExperience{}, err
	}
	syncHealth, err := m.TenantSyncHealth(ctx, tenantID)
	if err != nil {
		return model.TenantRoleExperience{}, err
	}

	return buildTenantRoleExperience(operations, monetization, business, preferences, timeline, syncHealth, generatedAt), nil
}

func (m *Memory) TenantOnboardingCenter(ctx context.Context, tenantID string) (model.TenantOnboardingCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	roles, err := m.TenantRoleExperiences(ctx, tenantID)
	if err != nil {
		return model.TenantOnboardingCenter{}, err
	}
	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantOnboardingCenter{}, err
	}
	portfolio, err := m.TenantPortfolioCenter(ctx, tenantID)
	if err != nil {
		return model.TenantOnboardingCenter{}, err
	}
	push, err := m.TenantPushActivationCenter(ctx, tenantID)
	if err != nil {
		return model.TenantOnboardingCenter{}, err
	}
	preferences, err := m.TenantNotificationPreferences(ctx, tenantID)
	if err != nil {
		return model.TenantOnboardingCenter{}, err
	}
	syncHealth, err := m.TenantSyncHealth(ctx, tenantID)
	if err != nil {
		return model.TenantOnboardingCenter{}, err
	}

	return buildTenantOnboardingCenter(roles, packageBilling, portfolio, push, preferences, syncHealth, generatedAt), nil
}

func (m *Memory) TenantCustomerSettingsCenter(ctx context.Context, tenantID string) (model.TenantCustomerSettingsCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSettingsCenter{}, err
	}
	onboarding, err := m.TenantOnboardingCenter(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSettingsCenter{}, err
	}
	preferences, err := m.TenantNotificationPreferences(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSettingsCenter{}, err
	}
	roles, err := m.TenantRoleExperiences(ctx, tenantID)
	if err != nil {
		return model.TenantCustomerSettingsCenter{}, err
	}
	routes := m.ListNotificationRoutes(ctx, tenantID)

	m.mu.RLock()
	tenant, ok := m.tenants[tenantID]
	m.mu.RUnlock()
	if !ok {
		return model.TenantCustomerSettingsCenter{}, ErrTenantNotFound
	}

	return buildTenantCustomerSettingsCenter(tenant, packageBilling, onboarding, preferences, roles, routes, generatedAt), nil
}

func (m *Memory) TenantRevenueOperationsCenter(ctx context.Context, tenantID string) (model.TenantRevenueOperationsCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	controlRoom, err := m.TenantCustomerControlRoom(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	successPacket, err := m.TenantCustomerSuccessPacket(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	pushActivation, err := m.TenantPushActivationCenter(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	portfolio, err := m.TenantPortfolioCenter(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	onboarding, err := m.TenantOnboardingCenter(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	settings, err := m.TenantCustomerSettingsCenter(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	packageBilling, err := m.TenantPackageBillingReadiness(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}
	provider, err := m.TenantProviderSimulationLab(ctx, tenantID)
	if err != nil {
		return model.TenantRevenueOperationsCenter{}, err
	}

	return buildTenantRevenueOperationsCenter(controlRoom, successPacket, pushActivation, portfolio, onboarding, settings, packageBilling, provider, generatedAt), nil
}

func (m *Memory) TenantDeploymentReadinessCenter(ctx context.Context, tenantID string) (model.TenantDeploymentReadinessCenter, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	onboarding, err := m.TenantOnboardingCenter(ctx, tenantID)
	if err != nil {
		return model.TenantDeploymentReadinessCenter{}, err
	}
	settings, err := m.TenantCustomerSettingsCenter(ctx, tenantID)
	if err != nil {
		return model.TenantDeploymentReadinessCenter{}, err
	}
	syncHealth, err := m.TenantSyncHealth(ctx, tenantID)
	if err != nil {
		return model.TenantDeploymentReadinessCenter{}, err
	}
	portfolio, err := m.TenantPortfolioCenter(ctx, tenantID)
	if err != nil {
		return model.TenantDeploymentReadinessCenter{}, err
	}
	revenueOps, err := m.TenantRevenueOperationsCenter(ctx, tenantID)
	if err != nil {
		return model.TenantDeploymentReadinessCenter{}, err
	}

	return buildTenantDeploymentReadinessCenter(onboarding, settings, syncHealth, portfolio, revenueOps, generatedAt), nil
}

func (m *Memory) TenantExecutiveConsole(ctx context.Context, tenantID string) (model.TenantExecutiveConsole, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantExecutiveConsole{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantExecutiveConsole{}, err
	}
	business, err := m.TenantBusinessDashboard(ctx, tenantID)
	if err != nil {
		return model.TenantExecutiveConsole{}, err
	}
	roles, err := m.TenantRoleExperiences(ctx, tenantID)
	if err != nil {
		return model.TenantExecutiveConsole{}, err
	}
	commandCenter, err := m.TenantNotificationCommandCenter(ctx, tenantID)
	if err != nil {
		return model.TenantExecutiveConsole{}, err
	}
	timeline, err := m.TenantDeliveryTimeline(ctx, tenantID, model.TenantDeliveryTimelineFilter{Limit: 8})
	if err != nil {
		return model.TenantExecutiveConsole{}, err
	}

	return buildTenantExecutiveConsole(operations, monetization, business, roles, commandCenter, timeline, generatedAt), nil
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

func (m *Memory) TenantNotificationRevenueCockpit(ctx context.Context, tenantID string) (model.TenantNotificationRevenueCockpit, error) {
	generatedAt := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)

	operations, err := m.TenantOperationsSummary(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationRevenueCockpit{}, err
	}
	monetization, err := m.TenantMonetizationSummary(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationRevenueCockpit{}, err
	}
	commandCenter, err := m.TenantNotificationCommandCenter(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationRevenueCockpit{}, err
	}
	preferences, err := m.TenantNotificationPreferences(ctx, tenantID)
	if err != nil {
		return model.TenantNotificationRevenueCockpit{}, err
	}
	timeline, err := m.TenantDeliveryTimeline(ctx, tenantID, model.TenantDeliveryTimelineFilter{Limit: 8})
	if err != nil {
		return model.TenantNotificationRevenueCockpit{}, err
	}

	return buildTenantNotificationRevenueCockpit(operations, monetization, commandCenter, preferences, timeline, generatedAt), nil
}

func (m *Memory) TenantDeliveryTimeline(_ context.Context, tenantID string, filter model.TenantDeliveryTimelineFilter) (model.TenantDeliveryTimeline, error) {
	now := time.Now().UTC()
	tenantID = strings.TrimSpace(tenantID)
	filter = normalizeDeliveryTimelineFilter(filter)

	m.mu.Lock()
	defer m.mu.Unlock()

	tenant, ok := m.tenants[tenantID]
	if !ok {
		return model.TenantDeliveryTimeline{}, ErrTenantNotFound
	}

	items := make([]model.TenantDeliveryTimelineItem, 0)
	sourceHosts := map[string]bool{}
	for _, device := range m.devices {
		if device.TenantID != tenantID {
			continue
		}
		if filter.DeviceID != "" && filter.DeviceID != device.DeviceID {
			continue
		}
		sourceHosts[device.DeviceID] = true
		m.seedDashboardForDeviceLocked(device)
		for _, delivery := range m.alertDeliveries[device.DeviceID] {
			item := deliveryTimelineItem(tenantID, device, delivery)
			if deliveryTimelineItemMatches(item, filter) {
				items = append(items, item)
			}
		}
	}
	if err := m.persistLocked(); err != nil {
		return model.TenantDeliveryTimeline{}, err
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].LastAttemptAt.Equal(items[j].LastAttemptAt) {
			return deliveryProblemRank(items[i].Status) > deliveryProblemRank(items[j].Status)
		}
		return items[i].LastAttemptAt.After(items[j].LastAttemptAt)
	})
	summary := deliveryTimelineSummary(items, len(sourceHosts))
	limited := items
	if len(limited) > filter.Limit {
		limited = limited[:filter.Limit]
	}

	return model.TenantDeliveryTimeline{
		TenantID:        tenant.TenantID,
		TenantName:      tenant.Name,
		Filters:         filter,
		Summary:         summary,
		Items:           append([]model.TenantDeliveryTimelineItem(nil), limited...),
		GeneratedAt:     now,
		PrivacyBoundary: constants.DeliveryTimelinePrivacyNote,
	}, nil
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

func normalizeDeliveryTimelineFilter(filter model.TenantDeliveryTimelineFilter) model.TenantDeliveryTimelineFilter {
	filter.DeviceID = strings.TrimSpace(filter.DeviceID)
	filter.Channel = strings.ToLower(strings.TrimSpace(filter.Channel))
	filter.Status = strings.ToLower(strings.TrimSpace(filter.Status))
	filter.Provider = strings.ToLower(strings.TrimSpace(filter.Provider))
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

func deliveryTimelineItem(tenantID string, device model.Device, delivery model.AlertDelivery) model.TenantDeliveryTimelineItem {
	return model.TenantDeliveryTimelineItem{
		ID:               delivery.ID,
		TenantID:         tenantID,
		DeviceID:         device.DeviceID,
		HostName:         device.HostName,
		EventID:          delivery.EventID,
		Channel:          delivery.Channel,
		Provider:         delivery.Provider,
		Recipient:        delivery.Recipient,
		Status:           delivery.Status,
		Attempts:         delivery.Attempts,
		Summary:          delivery.Summary,
		NextAction:       notificationNextAction(&delivery),
		PaidTier:         notificationCommandChannelTier(delivery.Channel),
		LastAttemptAt:    delivery.LastAttemptAt,
		NextRetryAt:      delivery.NextRetryAt,
		LastError:        delivery.LastError,
		SuppressedReason: delivery.SuppressedReason,
	}
}

func deliveryTimelineItemMatches(item model.TenantDeliveryTimelineItem, filter model.TenantDeliveryTimelineFilter) bool {
	if filter.Channel != "" && strings.ToLower(item.Channel) != filter.Channel {
		return false
	}
	if filter.Status != "" && strings.ToLower(item.Status) != filter.Status {
		return false
	}
	if filter.Provider != "" && strings.ToLower(item.Provider) != filter.Provider {
		return false
	}
	if filter.Query != "" && !strings.Contains(deliveryTimelineSearchText(item), filter.Query) {
		return false
	}
	return true
}

func deliveryTimelineSearchText(item model.TenantDeliveryTimelineItem) string {
	parts := []string{
		item.ID,
		item.TenantID,
		item.DeviceID,
		item.HostName,
		item.EventID,
		item.Channel,
		item.Provider,
		item.Recipient,
		item.Status,
		item.Summary,
		item.NextAction,
		item.PaidTier,
		item.LastError,
		item.SuppressedReason,
	}
	return strings.ToLower(strings.Join(parts, " "))
}

func deliveryTimelineSummary(items []model.TenantDeliveryTimelineItem, sourceHostCount int) model.TenantDeliveryTimelineSummary {
	summary := model.TenantDeliveryTimelineSummary{
		Total:               len(items),
		SourceHostCount:     sourceHostCount,
		RecommendedPaidTier: constants.PlanFamilyPro,
	}
	for _, item := range items {
		switch item.Status {
		case constants.DeliveryStatusDelivered:
			summary.Delivered++
			if summary.LastDeliveredAt == nil || item.LastAttemptAt.After(*summary.LastDeliveredAt) {
				deliveredAt := item.LastAttemptAt
				summary.LastDeliveredAt = &deliveredAt
			}
		case constants.DeliveryStatusRetrying:
			summary.Retrying++
			summary.RouteProofGaps++
		case constants.DeliveryStatusFailed:
			summary.Failed++
			summary.RouteProofGaps++
		case constants.DeliveryStatusSuppressed:
			summary.Suppressed++
		case constants.DeliveryStatusPending:
			summary.RouteProofGaps++
		}
		switch item.Channel {
		case constants.DeliveryChannelEmail:
			summary.Email++
		case constants.DeliveryChannelPush:
			summary.Push++
		case constants.DeliveryChannelDashboard:
			summary.Dashboard++
		}
		if item.NextRetryAt != nil && (summary.NextRetryAt == nil || item.NextRetryAt.Before(*summary.NextRetryAt)) {
			nextRetryAt := *item.NextRetryAt
			summary.NextRetryAt = &nextRetryAt
		}
		if item.PaidTier == constants.PlanBusiness {
			summary.RecommendedPaidTier = constants.PlanBusiness
		}
	}
	if len(items) > 0 {
		summary.NotificationScore = (summary.Delivered * 100) / len(items)
	}
	return summary
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

func retentionTierByID(tierID string) model.RetentionTier {
	for _, tier := range RetentionTiers() {
		if tier.ID == strings.TrimSpace(tierID) {
			return tier
		}
	}
	return model.RetentionTier{ID: strings.TrimSpace(tierID), Name: strings.TrimSpace(tierID)}
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

func buildTenantNotificationRevenueCockpit(
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	commandCenter model.TenantNotificationCommandCenter,
	preferences model.NotificationPreferenceCenter,
	timeline model.TenantDeliveryTimeline,
	generatedAt time.Time,
) model.TenantNotificationRevenueCockpit {
	routeScore := routeProofScore(commandCenter.Summary.RoutesTotal, commandCenter.Summary.RoutesNeedingProof)
	alertSLAReady := averageScore(routeScore, preferences.Summary.PreferenceScore, commandCenter.Summary.NotificationScore)
	revenueReadiness := averageScore(monetization.ReadinessScore, commandCenter.Summary.MonetizationReadiness, alertSLAReady, monetization.TrustScore)
	buyerDemoReady := notificationRevenueDemoReady(commandCenter.Summary, preferences.Summary, revenueReadiness)
	summary := model.TenantNotificationRevenueSummary{
		RevenueReadiness:       revenueReadiness,
		NotificationScore:      commandCenter.Summary.NotificationScore,
		AlertSLAReady:          alertSLAReady,
		OpenAnomalies:          commandCenter.Summary.Anomalies,
		HighPriorityAlerts:     commandCenter.Summary.HighPriorityAlerts,
		EmailDelivered:         commandCenter.Summary.EmailDelivered,
		PushDelivered:          commandCenter.Summary.PushDelivered,
		DashboardDelivered:     commandCenter.Summary.DashboardDelivered,
		DeliveryFailed:         commandCenter.Summary.DeliveryFailed,
		DeliveryRetrying:       commandCenter.Summary.DeliveryRetrying,
		RoutesNeedingProof:     commandCenter.Summary.RoutesNeedingProof,
		WeeklyReportReady:      commandCenter.Summary.WeeklyReportReady,
		EscalationReady:        preferences.Summary.EscalationEnabled,
		BuyerDemoReady:         buyerDemoReady,
		RecommendedPaidPackage: firstNonEmpty(commandCenter.Summary.RecommendedPaidPackage, monetization.PlanName, operations.PlanName, constants.PlanFamilyPro),
		NextBestAction:         notificationRevenueNextAction(commandCenter.Actions, monetization.ConversionActions),
	}
	summary.Status = notificationRevenueStatus(summary)
	summary.Headline, summary.Detail = notificationRevenueNarrative(summary, operations, monetization, timeline)

	return model.TenantNotificationRevenueCockpit{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        monetization.Audience,
		Summary:         summary,
		KPIs:            notificationRevenueKPIs(summary, monetization, preferences, timeline),
		Channels:        notificationRevenueChannels(commandCenter.Channels),
		Scenarios:       notificationRevenueScenarios(summary, preferences),
		Actions:         notificationRevenueActions(commandCenter.Actions, monetization.ConversionActions, timeline.Items, generatedAt),
		PrivacyBoundary: constants.NotificationRevenuePrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func routeProofScore(routesTotal int, routesNeedingProof int) int {
	if routesTotal <= 0 {
		return 0
	}
	return ((routesTotal - routesNeedingProof) * 100) / routesTotal
}

func notificationRevenueDemoReady(summary model.TenantNotificationCommandCenterSummary, preferences model.NotificationPreferenceCenterSummary, readiness int) bool {
	return readiness >= 70 &&
		summary.EmailDelivered > 0 &&
		summary.DashboardDelivered > 0 &&
		preferences.EmailEnabled &&
		preferences.PushEnabled &&
		preferences.DashboardEnabled &&
		preferences.EscalationEnabled &&
		summary.RoutesNeedingProof <= 1 &&
		summary.DeliveryFailed == 0
}

func notificationRevenueStatus(summary model.TenantNotificationRevenueSummary) string {
	switch {
	case summary.DeliveryFailed > 0 || summary.HighPriorityAlerts > 0:
		return constants.StatusAttention
	case summary.RoutesNeedingProof > 0 || summary.DeliveryRetrying > 0 || !summary.EscalationReady:
		return constants.StatusWatch
	case summary.BuyerDemoReady:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func notificationRevenueNarrative(summary model.TenantNotificationRevenueSummary, operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, timeline model.TenantDeliveryTimeline) (string, string) {
	if summary.HighPriorityAlerts > 0 {
		return fmt.Sprintf("%d high-priority anomalies need notification assurance", summary.HighPriorityAlerts),
			fmt.Sprintf("%s can show mail, push, dashboard, and owner-action proof without exposing alert content.", firstNonEmpty(operations.TenantName, monetization.TenantName, "TraceDeck"))
	}
	if summary.RoutesNeedingProof > 0 {
		return fmt.Sprintf("%d notification routes need buyer-proof evidence", summary.RoutesNeedingProof),
			"Finish provider-safe mail, push, and dashboard proof before pitching the paid notification package."
	}
	if summary.BuyerDemoReady {
		return fmt.Sprintf("%s notification revenue cockpit is demo-ready", summary.RecommendedPaidPackage),
			fmt.Sprintf("%d recent delivery rows support anomaly alert SLAs, mail proof, push reach, weekly reports, and upgrade packaging.", timeline.Summary.Total)
	}
	return "Notification revenue cockpit is building proof",
		"Keep anomaly classification, escalation policy, delivery route proof, and weekly report readiness current for paid demos."
}

func notificationRevenueKPIs(summary model.TenantNotificationRevenueSummary, monetization model.TenantMonetizationSummary, preferences model.NotificationPreferenceCenter, timeline model.TenantDeliveryTimeline) []model.TenantNotificationRevenueKPI {
	return []model.TenantNotificationRevenueKPI{
		{
			ID:       "revenue-readiness",
			Label:    "Revenue Readiness",
			Value:    fmt.Sprintf("%d%%", summary.RevenueReadiness),
			Detail:   fmt.Sprintf("%s stage for %s", monetization.ConversionStage, firstNonEmpty(monetization.PlanName, summary.RecommendedPaidPackage)),
			Status:   scoreStatus(summary.RevenueReadiness),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "anomaly-sla",
			Label:    "Anomaly SLA",
			Value:    fmt.Sprintf("%d%%", summary.AlertSLAReady),
			Detail:   fmt.Sprintf("%d high-priority alerts, %d open anomalies", summary.HighPriorityAlerts, summary.OpenAnomalies),
			Status:   scoreStatus(summary.AlertSLAReady),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "mail-proof",
			Label:    "Mail Proof",
			Value:    fmt.Sprintf("%d delivered", summary.EmailDelivered),
			Detail:   firstNonEmpty(monetization.NotificationPromise.Email, "critical anomaly email proof pending"),
			Status:   deliveryValueStatus(summary.EmailDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "push-proof",
			Label:    "Push Proof",
			Value:    fmt.Sprintf("%d delivered", summary.PushDelivered),
			Detail:   firstNonEmpty(monetization.NotificationPromise.Push, "push route proof pending"),
			Status:   deliveryValueStatus(summary.PushDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "weekly-report",
			Label:    "Weekly Report",
			Value:    readyLabel(summary.WeeklyReportReady),
			Detail:   fmt.Sprintf("%d delivery rows in provider-safe timeline", timeline.Summary.Total),
			Status:   boolStatus(summary.WeeklyReportReady),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "escalation-policy",
			Label:    "Escalation Policy",
			Value:    readyLabel(summary.EscalationReady),
			Detail:   fmt.Sprintf("%d rules, %d routes need proof", preferences.Summary.RulesTotal, summary.RoutesNeedingProof),
			Status:   boolStatus(summary.EscalationReady),
			PaidTier: firstNonEmpty(preferences.Summary.RecommendedPaidTier, constants.PlanFamilyPro),
		},
	}
}

func notificationRevenueChannels(channels []model.TenantNotificationCommandCenterChannel) []model.TenantNotificationRevenueChannel {
	items := make([]model.TenantNotificationRevenueChannel, 0, len(channels))
	for _, channel := range channels {
		items = append(items, model.TenantNotificationRevenueChannel{
			Channel:              channel.Channel,
			Provider:             channel.Provider,
			RecipientLabel:       channel.Recipient,
			Status:               firstNonEmpty(channel.LatestDeliveryStatus, channel.RouteStatus, constants.StatusPending),
			ProofState:           channel.ProofState,
			LatestDeliveryStatus: channel.LatestDeliveryStatus,
			Attempts:             channel.Attempts,
			LastDeliveryAt:       channel.LastDeliveryAt,
			SLA:                  channel.SLA,
			BusinessValue:        notificationRevenueChannelValue(channel.Channel),
			NextAction:           channel.NextAction,
			PaidTier:             channel.PaidTier,
		})
	}
	return items
}

func notificationRevenueChannelValue(channel string) string {
	switch channel {
	case constants.DeliveryChannelEmail:
		return "Critical anomaly and weekly report proof for parent, school, and business buyers."
	case constants.DeliveryChannelPush:
		return "Fast anomaly reach and escalation signal for mobile-first paid plans."
	case constants.DeliveryChannelDashboard:
		return "Always-on audit trail for admins who need visible proof before renewal."
	default:
		return "Provider-safe route proof for paid notification workflows."
	}
}

func notificationRevenueScenarios(summary model.TenantNotificationRevenueSummary, preferences model.NotificationPreferenceCenter) []model.TenantNotificationRevenueScenario {
	escalationStatus := boolStatus(preferences.Summary.EscalationEnabled)
	return []model.TenantNotificationRevenueScenario{
		{
			Title:          "Non-study video threshold",
			Detail:         "Classify study-safe content separately, then alert only when entertainment usage crosses policy.",
			Trigger:        constants.AlertTriggerNonStudyYouTube,
			Channels:       []string{constants.DeliveryChannelEmail, constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Severity:       constants.SeverityMedium,
			Status:         escalationStatus,
			ExampleOutcome: "Buyer sees ignored study video and escalated entertainment anomaly as separate outcomes.",
			PaidTier:       constants.PlanFamilyPro,
		},
		{
			Title:          "Media player after hours",
			Detail:         "Escalate media playback category and safe file-label metadata without collecting private content.",
			Trigger:        constants.AlertTriggerMediaPlayback,
			Channels:       []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Severity:       constants.SeverityHigh,
			Status:         countStatus(summary.HighPriorityAlerts),
			ExampleOutcome: "Parent or admin receives mail proof and sees dashboard action history.",
			PaidTier:       constants.PlanFamilyPro,
		},
		{
			Title:          "Risky software install",
			Detail:         "Surface torrent, proxy, unsigned installer, and unknown browser categories for risk review.",
			Trigger:        constants.AlertTriggerRiskySoftware,
			Channels:       []string{constants.DeliveryChannelEmail, constants.DeliveryChannelPush},
			Severity:       constants.SeverityHigh,
			Status:         summary.Status,
			ExampleOutcome: "Paid dashboard proves who was notified and which owner action is next.",
			PaidTier:       constants.PlanSchool,
		},
		{
			Title:          "Archive or agent trust gap",
			Detail:         "Convert upload backlog, stopped agent, and route proof gaps into renewal-risk actions.",
			Trigger:        constants.AlertTriggerArchiveBacklog,
			Channels:       []string{constants.DeliveryChannelDashboard, constants.DeliveryChannelEmail},
			Severity:       constants.SeverityMedium,
			Status:         proofGapStatus(summary.RoutesNeedingProof),
			ExampleOutcome: "Business buyer sees backlog, retry timing, and owner acknowledgement evidence.",
			PaidTier:       constants.PlanBusiness,
		},
	}
}

func notificationRevenueActions(commandActions []model.TenantNotificationCommandCenterAction, conversionActions []model.TenantOperationsSignal, timelineItems []model.TenantDeliveryTimelineItem, generatedAt time.Time) []model.TenantNotificationRevenueAction {
	actions := make([]model.TenantNotificationRevenueAction, 0, 8)
	for _, action := range commandActions {
		actions = append(actions, model.TenantNotificationRevenueAction{
			Title:           action.Title,
			Detail:          action.Detail,
			Owner:           action.Owner,
			Status:          action.Status,
			Severity:        action.Severity,
			SLA:             action.SLA,
			ConversionLever: firstNonEmpty(action.PaidTier, constants.PlanFamilyPro),
			Source:          "notification_command_center",
			ObservedAt:      action.ObservedAt,
		})
		if len(actions) >= 4 {
			break
		}
	}
	for _, action := range conversionActions {
		actions = append(actions, model.TenantNotificationRevenueAction{
			Title:           action.Title,
			Detail:          action.Detail,
			Owner:           action.Owner,
			Status:          action.Status,
			Severity:        action.Severity,
			SLA:             "commercial owner review",
			ConversionLever: constants.PlanFamilyPro,
			Source:          "monetization_summary",
			ObservedAt:      action.ObservedAt,
		})
		if len(actions) >= 7 {
			break
		}
	}
	if len(timelineItems) > 0 {
		item := timelineItems[0]
		actions = append(actions, model.TenantNotificationRevenueAction{
			Title:           titleWord(item.Channel) + " delivery evidence review",
			Detail:          firstNonEmpty(item.NextAction, item.Summary, "Review provider-safe delivery row before a paid demo."),
			Owner:           item.Recipient,
			Status:          item.Status,
			Severity:        constants.SeverityInfo,
			SLA:             firstNonEmpty(item.PaidTier, constants.PlanFamilyPro),
			ConversionLever: "buyer proof timeline",
			Source:          "delivery_timeline",
			ObservedAt:      item.LastAttemptAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantNotificationRevenueAction{
			Title:           "Package notification proof for a buyer demo",
			Detail:          "Verify email, push, dashboard, escalation, report, and archive proof before upgrade review.",
			Owner:           constants.RoleBusinessManager,
			Status:          constants.StatusPending,
			Severity:        constants.SeverityInfo,
			SLA:             "before next demo",
			ConversionLever: constants.PlanFamilyPro,
			Source:          "notification_revenue_cockpit",
			ObservedAt:      generatedAt,
		})
	}
	if len(actions) > 8 {
		return actions[:8]
	}
	return actions
}

func notificationRevenueNextAction(commandActions []model.TenantNotificationCommandCenterAction, conversionActions []model.TenantOperationsSignal) string {
	if len(commandActions) > 0 {
		return firstNonEmpty(commandActions[0].Detail, commandActions[0].Title)
	}
	if len(conversionActions) > 0 {
		return firstNonEmpty(conversionActions[0].Detail, conversionActions[0].Title)
	}
	return "Verify mail, push, dashboard, escalation, and weekly report proof before the paid demo."
}

func readyLabel(ready bool) string {
	if ready {
		return "ready"
	}
	return "pending"
}

func boolStatus(ready bool) string {
	if ready {
		return constants.StatusHealthy
	}
	return constants.StatusWatch
}

func proofGapStatus(count int) string {
	if count > 0 {
		return constants.StatusWatch
	}
	return constants.StatusHealthy
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

func buildTenantRoleExperience(
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	business model.TenantBusinessDashboard,
	preferences model.NotificationPreferenceCenter,
	timeline model.TenantDeliveryTimeline,
	syncHealth model.TenantSyncHealth,
	generatedAt time.Time,
) model.TenantRoleExperience {
	roles := []model.TenantRoleExperienceRole{
		roleExperienceParent(operations, monetization, business, preferences, timeline),
		roleExperienceStudent(operations, monetization, preferences, syncHealth),
		roleExperienceSchoolAdmin(operations, monetization, business, syncHealth),
		roleExperienceBusinessManager(operations, monetization, business, timeline),
	}
	ready := 0
	for _, role := range roles {
		if role.ReadinessScore >= 70 {
			ready++
		}
	}
	onboarding := roleOnboardingItems(operations, monetization, business, preferences, timeline, syncHealth)
	status := constants.StatusPending
	score := averageRoleReadiness(roles)
	switch {
	case score >= 80 && ready == len(roles):
		status = constants.StatusHealthy
	case score >= 60:
		status = constants.StatusWatch
	default:
		status = constants.StatusAttention
	}
	return model.TenantRoleExperience{
		TenantID:   operations.TenantID,
		TenantName: operations.TenantName,
		PlanID:     operations.PlanID,
		PlanName:   operations.PlanName,
		Audience:   monetization.Audience,
		Summary: model.TenantRoleExperienceSummary{
			Status:             status,
			Headline:           fmt.Sprintf("%d/%d role views ready for paid onboarding", ready, len(roles)),
			Detail:             "Parent, student, school admin, and business manager experiences are packaged from notification, report, archive, consent, and delivery proof.",
			ReadinessScore:     score,
			RolesTotal:         len(roles),
			RolesReady:         ready,
			OwnerActions:       len(onboarding),
			NotificationScore:  operations.NotificationScore,
			TrustScore:         monetization.TrustScore,
			PrivacyVisible:     true,
			RecommendedPackage: firstNonEmpty(business.Summary.RecommendedPackage, monetization.PlanName, operations.PlanName),
		},
		Roles:           roles,
		Onboarding:      onboarding,
		PrivacyBoundary: constants.RoleExperiencePrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func buildTenantOnboardingCenter(
	roles model.TenantRoleExperience,
	packageBilling model.TenantPackageBillingReadiness,
	portfolio model.TenantPortfolioCenter,
	push model.TenantPushActivationCenter,
	preferences model.NotificationPreferenceCenter,
	syncHealth model.TenantSyncHealth,
	generatedAt time.Time,
) model.TenantOnboardingCenter {
	installReady := syncHealth.HostsReporting > 0 && portfolio.Summary.HostsTotal > 0
	autostartReady := syncHealth.OfflineReplayReady && syncHealth.HostsReporting > 0
	notificationReady := push.Summary.NotificationScore >= 60 && push.Summary.MailDelivered > 0 && push.Summary.DashboardDelivered > 0
	archiveReady := packageBilling.Summary.ArchiveReady && portfolio.Summary.ArchiveBacklog == 0
	packageReady := packageBilling.Summary.PackageScore >= 60 && packageBilling.Summary.FeatureGatesReady > 0
	privacyReady := roles.Summary.PrivacyVisible && packageBilling.Summary.TrustScore >= 60

	steps := tenantOnboardingSteps(packageBilling, portfolio, push, preferences, syncHealth, installReady, autostartReady, notificationReady, archiveReady, packageReady, privacyReady)
	readySteps := 0
	for _, step := range steps {
		if step.Status == constants.StatusHealthy {
			readySteps++
		}
	}
	readiness := 0
	if len(steps) > 0 {
		readiness = (readySteps * 100) / len(steps)
	}
	readiness = averageScore(readiness, roles.Summary.ReadinessScore, packageBilling.Summary.PackageScore, push.Summary.ActivationScore)
	summary := model.TenantOnboardingSummary{
		ReadinessScore:     readiness,
		SetupStepsReady:    readySteps,
		SetupStepsTotal:    len(steps),
		HostsTotal:         portfolio.Summary.HostsTotal,
		HostsReporting:     syncHealth.HostsReporting,
		InstallReady:       installReady,
		AutostartReady:     autostartReady,
		NotificationReady:  notificationReady,
		ArchiveReady:       archiveReady,
		PackageReady:       packageReady,
		RolesReady:         roles.Summary.RolesReady,
		RolesTotal:         roles.Summary.RolesTotal,
		PrivacyReady:       privacyReady,
		RecommendedPackage: firstNonEmpty(packageBilling.Summary.RecommendedPackage, roles.Summary.RecommendedPackage, portfolio.Summary.RecommendedPaidPackage),
		OwnerNextStep:      tenantOnboardingNextStep(steps),
	}
	summary.Status = tenantOnboardingStatus(summary)
	summary.Headline, summary.Detail = tenantOnboardingNarrative(summary)

	return model.TenantOnboardingCenter{
		TenantID:        roles.TenantID,
		TenantName:      roles.TenantName,
		PlanID:          roles.PlanID,
		PlanName:        roles.PlanName,
		Audience:        roles.Audience,
		Summary:         summary,
		Steps:           steps,
		Roles:           tenantOnboardingRoles(roles.Roles),
		Proof:           tenantOnboardingProof(summary, packageBilling, portfolio, push, syncHealth),
		Actions:         tenantOnboardingActions(steps, roles.Onboarding, packageBilling.Actions),
		PrivacyBoundary: constants.OnboardingCenterPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func tenantOnboardingSteps(
	packageBilling model.TenantPackageBillingReadiness,
	portfolio model.TenantPortfolioCenter,
	push model.TenantPushActivationCenter,
	preferences model.NotificationPreferenceCenter,
	syncHealth model.TenantSyncHealth,
	installReady bool,
	autostartReady bool,
	notificationReady bool,
	archiveReady bool,
	packageReady bool,
	privacyReady bool,
) []model.TenantOnboardingStep {
	return []model.TenantOnboardingStep{
		tenantOnboardingStep("agent-install", "Install endpoint agent", installReady, constants.RoleBusinessManager, fmt.Sprintf("%d/%d hosts reporting", syncHealth.HostsReporting, portfolio.Summary.HostsTotal), "Use the signed agent build and local config profile before assigning production policy.", constants.PlanFamilyPro, true),
		tenantOnboardingStep("autostart", "Enable reboot persistence", autostartReady, constants.RoleBusinessManager, "Windows Task Scheduler, macOS launchd, and Linux systemd manifests are managed by scripts.", "Confirm the agent starts after restart and local outbox retry remains enabled.", constants.PlanFamilyPro, true),
		tenantOnboardingStep("notification-policy", "Configure notification policy", preferences.Summary.PreferenceScore >= 60, constants.RoleParent, fmt.Sprintf("%d rules across email=%t push=%t dashboard=%t", preferences.Summary.RulesTotal, preferences.Summary.EmailEnabled, preferences.Summary.PushEnabled, preferences.Summary.DashboardEnabled), "Immediate, digest, silent, quiet-hours, and escalation settings should match the buyer role.", constants.PlanFamilyPro, true),
		tenantOnboardingStep("mail-push-proof", "Prove mail and push delivery", notificationReady, constants.RoleParent, fmt.Sprintf("%d mail delivered, %d push delivered, %d dashboard delivered", push.Summary.MailDelivered, push.Summary.PushDelivered, push.Summary.DashboardDelivered), "Verify anomaly alerts reach the owner and dashboard fallback before production rollout.", constants.PlanFamilyPro, true),
		tenantOnboardingStep("archive-retention", "Confirm archive retention", archiveReady, constants.RoleSchoolAdmin, fmt.Sprintf("%d archive batches pending", portfolio.Summary.ArchiveBacklog), "S3 lifecycle and local TTL proof should be ready before selling retention value.", constants.PlanFamilyPro, false),
		tenantOnboardingStep("role-dashboards", "Assign role dashboards", packageBilling.Summary.FeatureGatesReady > 0 && packageBilling.Summary.NotificationReady, constants.RoleBusinessManager, fmt.Sprintf("%d/%d feature gates ready", packageBilling.Summary.FeatureGatesReady, packageBilling.Summary.FeatureGatesTotal), "Parent, student, school admin, and business manager views should be visible for onboarding.", constants.PlanSchool, false),
		tenantOnboardingStep("package-readiness", "Review package readiness", packageReady, constants.RoleBusinessManager, fmt.Sprintf("%d%% package score for %s", packageBilling.Summary.PackageScore, packageBilling.Summary.RecommendedPackage), "Use feature gates, plan fit, milestones, and owner actions for the paid handoff.", constants.PlanFamilyPro, false),
		tenantOnboardingStep("privacy-guard", "Review privacy and data rights", privacyReady, constants.RoleBusinessManager, fmt.Sprintf("%d%% trust score", packageBilling.Summary.TrustScore), "Confirm metadata-only collection, visible monitoring, export, delete request, and audit proof.", constants.PlanBusiness, true),
	}
}

func tenantOnboardingStep(id string, title string, ready bool, owner string, evidence string, detail string, paidTier string, blocking bool) model.TenantOnboardingStep {
	status := constants.StatusAttention
	if ready {
		status = constants.StatusHealthy
	} else if !blocking {
		status = constants.StatusWatch
	}
	return model.TenantOnboardingStep{
		ID:       id,
		Title:    title,
		Detail:   detail,
		Owner:    owner,
		Status:   status,
		Evidence: evidence,
		PaidTier: paidTier,
		Blocking: blocking,
	}
}

func tenantOnboardingStatus(summary model.TenantOnboardingSummary) string {
	switch {
	case summary.SetupStepsReady == summary.SetupStepsTotal && summary.ReadinessScore >= 80:
		return constants.StatusHealthy
	case !summary.InstallReady || !summary.AutostartReady || !summary.NotificationReady || !summary.PrivacyReady:
		return constants.StatusAttention
	case summary.ReadinessScore >= 60:
		return constants.StatusWatch
	default:
		return constants.StatusPending
	}
}

func tenantOnboardingNarrative(summary model.TenantOnboardingSummary) (string, string) {
	if !summary.InstallReady || !summary.AutostartReady {
		return "Endpoint deployment needs boot persistence proof",
			fmt.Sprintf("%d/%d setup steps ready with %d/%d hosts reporting.", summary.SetupStepsReady, summary.SetupStepsTotal, summary.HostsReporting, summary.HostsTotal)
	}
	if !summary.NotificationReady {
		return "Notification proof is blocking production onboarding",
			"Mail, push, dashboard fallback, and alert policy proof must be visible before owner handoff."
	}
	if !summary.PrivacyReady {
		return "Privacy and data-rights proof needs review",
			"Visible monitoring, audit, export, delete request, and metadata-only guardrails should be confirmed."
	}
	return fmt.Sprintf("%s onboarding is %d%% ready", summary.RecommendedPackage, summary.ReadinessScore),
		fmt.Sprintf("%d/%d setup steps, %d/%d role views, archive=%t, package=%t.", summary.SetupStepsReady, summary.SetupStepsTotal, summary.RolesReady, summary.RolesTotal, summary.ArchiveReady, summary.PackageReady)
}

func tenantOnboardingNextStep(steps []model.TenantOnboardingStep) string {
	for _, step := range steps {
		if step.Blocking && step.Status != constants.StatusHealthy {
			return step.Title
		}
	}
	for _, step := range steps {
		if step.Status != constants.StatusHealthy {
			return step.Title
		}
	}
	return "Run the customer onboarding review and keep evidence fresh."
}

func tenantOnboardingRoles(roles []model.TenantRoleExperienceRole) []model.TenantOnboardingRole {
	rows := make([]model.TenantOnboardingRole, 0, len(roles))
	for _, role := range roles {
		rows = append(rows, model.TenantOnboardingRole{
			RoleID:     role.RoleID,
			Name:       role.Name,
			Status:     role.Status,
			ViewMode:   role.ViewMode,
			PaidTier:   role.PaidTier,
			NextAction: role.NextAction,
		})
	}
	return rows
}

func tenantOnboardingProof(
	summary model.TenantOnboardingSummary,
	packageBilling model.TenantPackageBillingReadiness,
	portfolio model.TenantPortfolioCenter,
	push model.TenantPushActivationCenter,
	syncHealth model.TenantSyncHealth,
) []model.TenantOnboardingProof {
	return []model.TenantOnboardingProof{
		{ID: "setup", Label: "Setup steps", Value: fmt.Sprintf("%d/%d ready", summary.SetupStepsReady, summary.SetupStepsTotal), Detail: summary.OwnerNextStep, Status: summary.Status, PaidTier: constants.PlanFamilyPro},
		{ID: "hosts", Label: "Host reporting", Value: fmt.Sprintf("%d/%d hosts", syncHealth.HostsReporting, portfolio.Summary.HostsTotal), Detail: syncHealth.OfflineReplaySummary, Status: syncHealth.Status, PaidTier: constants.PlanFamilyPro},
		{ID: "notifications", Label: "Notification proof", Value: fmt.Sprintf("%d%%", push.Summary.NotificationScore), Detail: fmt.Sprintf("%d mail, %d push, %d dashboard delivered", push.Summary.MailDelivered, push.Summary.PushDelivered, push.Summary.DashboardDelivered), Status: push.Summary.Status, PaidTier: constants.PlanFamilyPro},
		{ID: "archive", Label: "Archive posture", Value: fmt.Sprintf("%d pending", portfolio.Summary.ArchiveBacklog), Detail: packageBilling.RetentionName, Status: boolStatus(summary.ArchiveReady), PaidTier: constants.PlanFamilyPro},
		{ID: "package", Label: "Package readiness", Value: fmt.Sprintf("%d%%", packageBilling.Summary.PackageScore), Detail: packageBilling.Summary.RecommendedPackage, Status: packageBilling.Summary.Status, PaidTier: constants.PlanFamilyPro},
		{ID: "privacy", Label: "Privacy guard", Value: boolReady(summary.PrivacyReady), Detail: constants.OnboardingCenterPrivacyNote, Status: boolStatus(summary.PrivacyReady), PaidTier: constants.PlanBusiness},
	}
}

func tenantOnboardingActions(steps []model.TenantOnboardingStep, roleActions []model.TenantRoleOnboardingItem, packageActions []model.TenantPackageBillingAction) []model.TenantOnboardingAction {
	actions := make([]model.TenantOnboardingAction, 0, 8)
	for _, step := range steps {
		if step.Status == constants.StatusHealthy {
			continue
		}
		actions = append(actions, model.TenantOnboardingAction{
			Title:    step.Title,
			Detail:   step.Detail,
			Owner:    step.Owner,
			Status:   step.Status,
			Severity: onboardingSeverity(step),
			PaidTier: step.PaidTier,
			Source:   "onboarding step",
		})
		if len(actions) >= 4 {
			break
		}
	}
	for _, action := range roleActions {
		if len(actions) >= 6 {
			break
		}
		actions = append(actions, model.TenantOnboardingAction{
			Title:    action.Title,
			Detail:   action.Detail,
			Owner:    action.Owner,
			Status:   action.Status,
			Severity: constants.SeverityMedium,
			PaidTier: action.PaidTier,
			Source:   "role onboarding",
		})
	}
	for _, action := range packageActions {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantOnboardingAction{
			Title:    action.Title,
			Detail:   firstNonEmpty(action.NextAction, action.Detail, action.ConversionLever),
			Owner:    action.Owner,
			Status:   action.Status,
			Severity: constants.SeverityMedium,
			PaidTier: action.PaidTier,
			Source:   "package billing",
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantOnboardingAction{
			Title:    "Run onboarding review",
			Detail:   "All core setup proof is ready. Review package, privacy, notification, archive, and role evidence with the owner.",
			Owner:    constants.RoleBusinessManager,
			Status:   constants.StatusHealthy,
			Severity: constants.SeverityInfo,
			PaidTier: constants.PlanFamilyPro,
			Source:   "onboarding center",
		})
	}
	return actions
}

func onboardingSeverity(step model.TenantOnboardingStep) string {
	if step.Blocking {
		return constants.SeverityHigh
	}
	return constants.SeverityMedium
}

func buildTenantCustomerSettingsCenter(
	tenant model.Tenant,
	packageBilling model.TenantPackageBillingReadiness,
	onboarding model.TenantOnboardingCenter,
	preferences model.NotificationPreferenceCenter,
	roles model.TenantRoleExperience,
	routes []model.NotificationRoute,
	generatedAt time.Time,
) model.TenantCustomerSettingsCenter {
	plan := planByID(tenant.PlanID)
	retention := retentionTierByID(tenant.RetentionTierID)
	recommendedPlan := customerSettingsRecommendedPlan(plan, packageBilling)
	settings := tenantCustomerSettings(tenant, plan, retention, recommendedPlan, packageBilling, onboarding, preferences, roles, routes)
	configured := 0
	for _, setting := range settings {
		if setting.Status == constants.StatusHealthy {
			configured++
		}
	}
	settingsScore := 0
	if len(settings) > 0 {
		settingsScore = (configured * 100) / len(settings)
	}
	settingsScore = averageScore(settingsScore, packageBilling.Summary.PackageScore, onboarding.Summary.ReadinessScore, preferences.Summary.PreferenceScore, roles.Summary.ReadinessScore)
	summary := model.TenantCustomerSettingsSummary{
		SettingsScore:      settingsScore,
		ConfiguredSettings: configured,
		SettingsTotal:      len(settings),
		CurrentPlan:        firstNonEmpty(plan.Name, tenant.PlanID),
		RecommendedPlan:    firstNonEmpty(recommendedPlan.Name, packageBilling.Summary.RecommendedPackage, packageBilling.PlanName),
		RetentionTier:      firstNonEmpty(retention.Name, tenant.RetentionTierID),
		NotificationReady:  onboarding.Summary.NotificationReady && preferences.Summary.PreferenceScore >= 60,
		ArchiveReady:       onboarding.Summary.ArchiveReady && packageBilling.Summary.RetentionReady,
		AutostartReady:     onboarding.Summary.AutostartReady,
		RoleViewsReady:     roles.Summary.RolesReady == roles.Summary.RolesTotal && roles.Summary.RolesTotal > 0,
		DataRightsReady:    packageBilling.Summary.TrustScore >= 60 && onboarding.Summary.PrivacyReady,
		PackageReady:       packageBilling.Summary.PackageScore >= 60,
		BillingReady:       packageBilling.Summary.BillingReady,
		OwnerNextStep:      tenantCustomerSettingsNextStep(settings),
	}
	summary.Status = tenantCustomerSettingsStatus(summary)
	summary.Headline, summary.Detail = tenantCustomerSettingsNarrative(summary)

	return model.TenantCustomerSettingsCenter{
		TenantID:         tenant.TenantID,
		TenantName:       tenant.Name,
		PlanID:           tenant.PlanID,
		PlanName:         firstNonEmpty(plan.Name, tenant.PlanID),
		RetentionTierID:  tenant.RetentionTierID,
		RetentionName:    firstNonEmpty(retention.Name, tenant.RetentionTierID),
		Audience:         firstNonEmpty(plan.Audience, onboarding.Audience, packageBilling.Audience),
		Summary:          summary,
		Settings:         settings,
		PlanOptions:      tenantCustomerSettingsPlanOptions(plan, recommendedPlan),
		RetentionOptions: tenantCustomerSettingsRetentionOptions(retention, recommendedPlan),
		Channels:         tenantCustomerSettingsChannels(preferences, routes),
		Actions:          tenantCustomerSettingsActions(settings, packageBilling.Actions, onboarding.Actions),
		PrivacyBoundary:  constants.CustomerSettingsPrivacyNote,
		GeneratedAt:      generatedAt,
	}
}

func customerSettingsRecommendedPlan(current model.Plan, packageBilling model.TenantPackageBillingReadiness) model.Plan {
	for _, plan := range packageBilling.Plans {
		if plan.Recommended {
			return planByID(plan.PlanID)
		}
	}
	if packageBilling.Summary.PackageScore >= 70 && current.ID == constants.PlanFree {
		return planByID(constants.PlanFamilyPro)
	}
	return current
}

func tenantCustomerSettings(
	tenant model.Tenant,
	plan model.Plan,
	retention model.RetentionTier,
	recommendedPlan model.Plan,
	packageBilling model.TenantPackageBillingReadiness,
	onboarding model.TenantOnboardingCenter,
	preferences model.NotificationPreferenceCenter,
	roles model.TenantRoleExperience,
	routes []model.NotificationRoute,
) []model.TenantCustomerSetting {
	return []model.TenantCustomerSetting{
		customerSetting("plan", "Plan Package", firstNonEmpty(plan.Name, tenant.PlanID), firstNonEmpty(recommendedPlan.Name, packageBilling.Summary.RecommendedPackage), packageBilling.Summary.BillingReady, constants.RoleBusinessManager, constants.PlanFamilyPro, true, fmt.Sprintf("%d%% package score", packageBilling.Summary.PackageScore), packageBilling.Summary.NextBestAction),
		customerSetting("retention", "Retention Tier", firstNonEmpty(retention.Name, tenant.RetentionTierID), customerSettingsRecommendedRetention(recommendedPlan), packageBilling.Summary.RetentionReady, constants.RoleBusinessManager, constants.PlanFamilyPro, true, fmt.Sprintf("%d local days, %d S3 standard days, archive after %d days", retention.LocalTTLDays, retention.S3StandardDays, retention.S3ArchiveAfterDays), "Confirm local TTL, S3 lifecycle, and archive policy before selling retention."),
		customerSetting("notification-policy", "Notification Policy", fmt.Sprintf("%d rules, immediate=%d", preferences.Summary.RulesTotal, preferences.Summary.ImmediateRules), "Immediate critical alerts with digest and study-safe suppression", preferences.Summary.PreferenceScore >= 60, constants.RoleParent, constants.PlanFamilyPro, true, fmt.Sprintf("%d%% preference score", preferences.Summary.PreferenceScore), "Tune quiet hours, escalation, and study-safe suppression before owner handoff."),
		customerSetting("mail-route", "Mail Delivery", customerSettingsRouteValue(routes, constants.DeliveryChannelEmail), "Verified email route for anomaly and weekly report proof", customerSettingsRouteReady(routes, constants.DeliveryChannelEmail), constants.RoleParent, constants.PlanFamilyPro, true, "email route proof is metadata-only", "Verify SMTP/SES route labels without storing provider secrets."),
		customerSetting("push-route", "Push Notification", customerSettingsRouteValue(routes, constants.DeliveryChannelPush), "Verified push route plus dashboard fallback", customerSettingsRouteReady(routes, constants.DeliveryChannelPush), constants.RoleParent, constants.PlanFamilyPro, true, "push route proof excludes raw endpoints", "Verify push route proof and dashboard fallback before promising urgent alerts."),
		customerSetting("archive", "Archive And Sync", boolReady(onboarding.Summary.ArchiveReady), "S3 standard, IA, archive lifecycle, and offline replay proof", onboarding.Summary.ArchiveReady, constants.RoleSchoolAdmin, constants.PlanFamilyPro, true, fmt.Sprintf("%d archive proof cards", len(packageBilling.FeatureGates)), "Clear archive backlog and keep lifecycle proof visible."),
		customerSetting("autostart", "Autostart", boolReady(onboarding.Summary.AutostartReady), "Windows Task Scheduler, macOS launchd, and Linux systemd proof", onboarding.Summary.AutostartReady, constants.RoleBusinessManager, constants.PlanFamilyPro, false, "managed by local scripts and manifests", "Verify reboot persistence during live boot testing."),
		customerSetting("role-views", "Role Dashboards", fmt.Sprintf("%d/%d roles ready", roles.Summary.RolesReady, roles.Summary.RolesTotal), "Parent, student, school admin, and business manager views", roles.Summary.RolesReady == roles.Summary.RolesTotal && roles.Summary.RolesTotal > 0, constants.RoleBusinessManager, constants.PlanSchool, true, roles.Summary.Headline, "Assign the right dashboard view before paid rollout."),
		customerSetting("privacy-data-rights", "Privacy And Data Rights", boolReady(onboarding.Summary.PrivacyReady), "Visible monitoring, audit, export, delete request, and pause policy proof", onboarding.Summary.PrivacyReady && packageBilling.Summary.TrustScore >= 60, constants.RoleBusinessManager, constants.PlanBusiness, false, fmt.Sprintf("%d%% trust score", packageBilling.Summary.TrustScore), "Review data-rights proof with the customer before activation."),
	}
}

func customerSetting(id string, label string, current string, recommended string, ready bool, owner string, paidTier string, configurable bool, evidence string, nextAction string) model.TenantCustomerSetting {
	return model.TenantCustomerSetting{
		ID:               id,
		Label:            label,
		CurrentValue:     current,
		RecommendedValue: recommended,
		Status:           boolStatus(ready),
		Owner:            owner,
		PaidTier:         paidTier,
		Configurable:     configurable,
		Evidence:         evidence,
		NextAction:       nextAction,
	}
}

func customerSettingsRecommendedRetention(plan model.Plan) string {
	switch plan.ID {
	case constants.PlanBusiness:
		return "Business Compliance"
	case constants.PlanSchool:
		return "School Year Archive"
	case constants.PlanFamilyPro:
		return "Family Cloud 90/365 Archive"
	default:
		return "Local Only 7 Days"
	}
}

func customerSettingsRouteValue(routes []model.NotificationRoute, channel string) string {
	for _, route := range routes {
		if route.Channel == channel {
			return fmt.Sprintf("%s via %s: %s", route.RecipientLabel, route.Provider, route.Status)
		}
	}
	return "route not configured"
}

func customerSettingsRouteReady(routes []model.NotificationRoute, channel string) bool {
	for _, route := range routes {
		if route.Channel == channel && route.Enabled && route.Status == constants.StatusHealthy {
			return true
		}
	}
	return false
}

func tenantCustomerSettingsStatus(summary model.TenantCustomerSettingsSummary) string {
	switch {
	case summary.SettingsScore >= 80 && summary.NotificationReady && summary.DataRightsReady:
		return constants.StatusHealthy
	case !summary.BillingReady || !summary.NotificationReady || !summary.DataRightsReady:
		return constants.StatusAttention
	case summary.SettingsScore >= 60:
		return constants.StatusWatch
	default:
		return constants.StatusPending
	}
}

func tenantCustomerSettingsNarrative(summary model.TenantCustomerSettingsSummary) (string, string) {
	switch {
	case !summary.BillingReady:
		return "Customer settings need package confirmation",
			"Plan, retention, trust, and billing-safe metadata must be ready before activation."
	case !summary.NotificationReady:
		return "Notification settings need proof",
			"Mail, push, dashboard fallback, preference policy, and escalation settings should be ready before rollout."
	case !summary.DataRightsReady:
		return "Privacy and data rights settings need review",
			"Visible monitoring, audit, export, delete request, and metadata-only guardrails must be confirmed."
	default:
		return fmt.Sprintf("%s settings are %d%% ready", summary.CurrentPlan, summary.SettingsScore),
			fmt.Sprintf("%d/%d settings configured with %s retention and %s recommended.", summary.ConfiguredSettings, summary.SettingsTotal, summary.RetentionTier, summary.RecommendedPlan)
	}
}

func tenantCustomerSettingsNextStep(settings []model.TenantCustomerSetting) string {
	for _, setting := range settings {
		if setting.Status != constants.StatusHealthy {
			return setting.NextAction
		}
	}
	return "Customer settings are ready for activation review."
}

func tenantCustomerSettingsPlanOptions(current model.Plan, recommended model.Plan) []model.TenantCustomerSettingsPlanOption {
	candidates := []model.Plan{planByID(constants.PlanFree), planByID(constants.PlanFamilyPro), planByID(constants.PlanSchool), planByID(constants.PlanBusiness)}
	options := make([]model.TenantCustomerSettingsPlanOption, 0, len(candidates))
	for _, plan := range candidates {
		isCurrent := plan.ID == current.ID
		isRecommended := plan.ID == recommended.ID
		status := constants.StatusWatch
		if isCurrent || isRecommended {
			status = constants.StatusHealthy
		}
		if plan.ID == constants.PlanFree && !isCurrent {
			status = constants.StatusPending
		}
		options = append(options, model.TenantCustomerSettingsPlanOption{
			PlanID:      plan.ID,
			Name:        plan.Name,
			Status:      status,
			Current:     isCurrent,
			Recommended: isRecommended,
			Audience:    plan.Audience,
			PriceModel:  plan.PriceModel,
			DeviceLimit: plan.DeviceLimit,
			BuyerValue:  businessPackageValue(plan.ID),
			NextAction:  packagePlanNextAction(plan, isCurrent, isRecommended),
		})
	}
	return options
}

func tenantCustomerSettingsRetentionOptions(current model.RetentionTier, recommendedPlan model.Plan) []model.TenantCustomerSettingsRetentionOption {
	recommendedID := customerSettingsRecommendedRetentionID(recommendedPlan)
	tiers := RetentionTiers()
	options := make([]model.TenantCustomerSettingsRetentionOption, 0, len(tiers))
	for _, tier := range tiers {
		isCurrent := tier.ID == current.ID
		isRecommended := tier.ID == recommendedID
		status := constants.StatusWatch
		if isCurrent || isRecommended {
			status = constants.StatusHealthy
		}
		if tier.ID == constants.RetentionLocalOnly && !isCurrent {
			status = constants.StatusPending
		}
		options = append(options, model.TenantCustomerSettingsRetentionOption{
			ID:                 tier.ID,
			Name:               tier.Name,
			Status:             status,
			Current:            isCurrent,
			Recommended:        isRecommended,
			LocalTTLDays:       tier.LocalTTLDays,
			S3StandardDays:     tier.S3StandardDays,
			S3StandardIAUntil:  tier.S3StandardIAUntil,
			S3ArchiveAfterDays: tier.S3ArchiveAfterDays,
			ComplianceExport:   tier.ComplianceExport,
			NextAction:         customerSettingsRetentionNextAction(tier, isCurrent, isRecommended),
		})
	}
	return options
}

func customerSettingsRecommendedRetentionID(plan model.Plan) string {
	switch plan.ID {
	case constants.PlanBusiness:
		return constants.RetentionBusiness
	case constants.PlanSchool:
		return constants.RetentionSchoolYear
	case constants.PlanFamilyPro:
		return constants.RetentionFamilyCloud
	default:
		return constants.RetentionLocalOnly
	}
}

func customerSettingsRetentionNextAction(tier model.RetentionTier, current bool, recommended bool) string {
	switch {
	case current:
		return "Use this retention tier as the current archive promise."
	case recommended:
		return "Prepare this retention tier as the recommended paid setting."
	case tier.ID == constants.RetentionLocalOnly:
		return "Keep local-only retention for starter trials."
	default:
		return "Keep this retention tier available for higher-trust buyers."
	}
}

func tenantCustomerSettingsChannels(preferences model.NotificationPreferenceCenter, routes []model.NotificationRoute) []model.TenantCustomerSettingsChannel {
	return []model.TenantCustomerSettingsChannel{
		customerSettingsChannel(constants.DeliveryChannelEmail, preferences.Summary.EmailEnabled, preferences, routes),
		customerSettingsChannel(constants.DeliveryChannelPush, preferences.Summary.PushEnabled, preferences, routes),
		customerSettingsChannel(constants.DeliveryChannelDashboard, preferences.Summary.DashboardEnabled, preferences, routes),
	}
}

func customerSettingsChannel(channel string, enabled bool, preferences model.NotificationPreferenceCenter, routes []model.NotificationRoute) model.TenantCustomerSettingsChannel {
	proof := customerSettingsRouteValue(routes, channel)
	ready := enabled && (channel == constants.DeliveryChannelDashboard || customerSettingsRouteReady(routes, channel))
	mode := "disabled"
	if enabled {
		mode = "enabled"
	}
	return model.TenantCustomerSettingsChannel{
		Channel:        channel,
		Enabled:        enabled,
		Status:         boolStatus(ready),
		PreferenceMode: mode,
		DeliveryProof:  proof,
		Evidence:       fmt.Sprintf("%d immediate, %d digest, %d silent rules", preferences.Summary.ImmediateRules, preferences.Summary.DigestRules, preferences.Summary.SilentRules),
		NextAction:     customerSettingsChannelNextAction(channel, ready),
	}
}

func customerSettingsChannelNextAction(channel string, ready bool) string {
	if ready {
		return "Keep route proof fresh for customer reviews."
	}
	return fmt.Sprintf("Enable and verify %s route proof before activation.", channel)
}

func tenantCustomerSettingsActions(settings []model.TenantCustomerSetting, packageActions []model.TenantPackageBillingAction, onboardingActions []model.TenantOnboardingAction) []model.TenantCustomerSettingsAction {
	actions := make([]model.TenantCustomerSettingsAction, 0, 8)
	for _, setting := range settings {
		if setting.Status == constants.StatusHealthy {
			continue
		}
		actions = append(actions, model.TenantCustomerSettingsAction{
			Title:    setting.Label,
			Detail:   setting.NextAction,
			Owner:    setting.Owner,
			Status:   setting.Status,
			Severity: constants.SeverityMedium,
			PaidTier: setting.PaidTier,
			Source:   "customer setting",
		})
		if len(actions) >= 4 {
			break
		}
	}
	for _, action := range packageActions {
		if len(actions) >= 6 {
			break
		}
		actions = append(actions, model.TenantCustomerSettingsAction{
			Title:    action.Title,
			Detail:   firstNonEmpty(action.NextAction, action.Detail),
			Owner:    action.Owner,
			Status:   action.Status,
			Severity: constants.SeverityMedium,
			PaidTier: action.PaidTier,
			Source:   "package billing",
		})
	}
	for _, action := range onboardingActions {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantCustomerSettingsAction{
			Title:    action.Title,
			Detail:   action.Detail,
			Owner:    action.Owner,
			Status:   action.Status,
			Severity: action.Severity,
			PaidTier: action.PaidTier,
			Source:   "onboarding",
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantCustomerSettingsAction{
			Title:    "Customer settings review",
			Detail:   "Plan, retention, notification, archive, role, and privacy settings are ready for activation review.",
			Owner:    constants.RoleBusinessManager,
			Status:   constants.StatusHealthy,
			Severity: constants.SeverityInfo,
			PaidTier: constants.PlanFamilyPro,
			Source:   "customer settings",
		})
	}
	return actions
}

func buildTenantRevenueOperationsCenter(
	controlRoom model.TenantCustomerControlRoom,
	successPacket model.TenantCustomerSuccessPacket,
	pushActivation model.TenantPushActivationCenter,
	portfolio model.TenantPortfolioCenter,
	onboarding model.TenantOnboardingCenter,
	settings model.TenantCustomerSettingsCenter,
	packageBilling model.TenantPackageBillingReadiness,
	provider model.TenantProviderSimulationLab,
	generatedAt time.Time,
) model.TenantRevenueOperationsCenter {
	summary := model.TenantRevenueOperationsSummary{
		ProductScore:           averageScore(controlRoom.Summary.ProductScore, successPacket.Summary.ReadinessScore, portfolio.Summary.PortfolioScore, onboarding.Summary.ReadinessScore),
		NotificationScore:      averageScore(controlRoom.Summary.NotificationScore, successPacket.Summary.NotificationScore, pushActivation.Summary.NotificationScore, portfolio.Summary.NotificationScore),
		TrustScore:             averageScore(controlRoom.Summary.TrustScore, successPacket.Summary.TrustScore, packageBilling.Summary.TrustScore, portfolio.Summary.TrustScore),
		PackageScore:           averageScore(controlRoom.Summary.PackageScore, successPacket.Summary.PackageScore, packageBilling.Summary.PackageScore),
		SettingsScore:          settings.Summary.SettingsScore,
		OnboardingScore:        onboarding.Summary.ReadinessScore,
		OpenAlerts:             controlRoom.Summary.OpenAlerts,
		HighPriorityAlerts:     controlRoom.Summary.HighPriorityAlerts,
		HostsTotal:             controlRoom.Summary.HostsTotal,
		HostsAttention:         controlRoom.Summary.HostsAttention,
		MailDelivered:          controlRoom.Summary.MailDelivered,
		PushDelivered:          controlRoom.Summary.PushDelivered,
		DashboardDelivered:     controlRoom.Summary.DashboardDelivered,
		RoutesNeedingProof:     maxInt(controlRoom.Summary.RoutesNeedingProof, successPacket.Summary.RoutesNeedingProof, pushActivation.Summary.PushRoutesNeedingProof, portfolio.Summary.RoutesNeedingProof, provider.Summary.RoutesNeedingProof),
		WeeklyReportReady:      successPacket.Summary.WeeklyReportReady || packageBilling.Summary.WeeklyReportReady || controlRoom.Summary.WeeklyReportReady,
		ArchiveBacklog:         maxInt(controlRoom.Summary.ArchiveBacklog, successPacket.Summary.ArchiveBacklog, portfolio.Summary.ArchiveBacklog),
		ProviderReady:          controlRoom.Summary.ProviderReady || successPacket.Summary.ProviderReady || provider.Summary.SLAReady,
		BillingReady:           controlRoom.Summary.BillingReady || successPacket.Summary.BillingReady || packageBilling.Summary.BillingReady,
		RecommendedPaidPackage: firstNonEmpty(successPacket.Summary.RecommendedPaidPackage, controlRoom.Summary.RecommendedPaidPackage, packageBilling.Summary.RecommendedPackage, settings.Summary.RecommendedPlan, controlRoom.PlanName),
	}
	summary.RevenueScore = averageScore(summary.ProductScore, summary.NotificationScore, summary.TrustScore, summary.PackageScore, summary.SettingsScore, summary.OnboardingScore, provider.Summary.ReadinessScore)
	summary.Status = revenueOperationsStatus(summary)
	summary.Headline = revenueOperationsHeadline(summary)
	summary.Detail = revenueOperationsDetail(summary)
	summary.OwnerNextStep = revenueOperationsNextStep(summary, controlRoom.Actions, settings.Actions, packageBilling.Actions, provider.Actions)

	return model.TenantRevenueOperationsCenter{
		TenantID:        controlRoom.TenantID,
		TenantName:      controlRoom.TenantName,
		PlanID:          controlRoom.PlanID,
		PlanName:        controlRoom.PlanName,
		Audience:        firstNonEmpty(controlRoom.Audience, settings.Audience, packageBilling.Audience, portfolio.Audience),
		Summary:         summary,
		Signals:         revenueOperationsSignals(summary, onboarding, settings, packageBilling, provider),
		Alerts:          revenueOperationsAlerts(controlRoom.Alerts, portfolio.AlertNotifications),
		Deliveries:      revenueOperationsDeliveries(controlRoom.Deliveries, provider.Routes),
		Levers:          revenueOperationsLevers(packageBilling, pushActivation, onboarding, settings, provider),
		Actions:         revenueOperationsActions(controlRoom.Actions, successPacket.Actions, pushActivation.Actions, settings.Actions, packageBilling.Actions, provider.Actions, generatedAt),
		PrivacyBoundary: constants.RevenueOperationsPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func revenueOperationsStatus(summary model.TenantRevenueOperationsSummary) string {
	switch {
	case summary.HighPriorityAlerts > 0 || summary.RoutesNeedingProof > 0:
		return constants.StatusAttention
	case !summary.ProviderReady || !summary.BillingReady || summary.ArchiveBacklog > 0:
		return constants.StatusWatch
	case summary.RevenueScore >= 80 && summary.NotificationScore >= 70:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func revenueOperationsHeadline(summary model.TenantRevenueOperationsSummary) string {
	if summary.HighPriorityAlerts > 0 {
		return fmt.Sprintf("%d urgent alert%s need revenue-ops attention", summary.HighPriorityAlerts, pluralSuffix(summary.HighPriorityAlerts))
	}
	if summary.RoutesNeedingProof > 0 {
		return fmt.Sprintf("%d notification route proof gap%s block paid confidence", summary.RoutesNeedingProof, pluralSuffix(summary.RoutesNeedingProof))
	}
	return fmt.Sprintf("%s is %d%% revenue-ready", firstNonEmpty(summary.RecommendedPaidPackage, "TraceDeck"), summary.RevenueScore)
}

func revenueOperationsDetail(summary model.TenantRevenueOperationsSummary) string {
	return fmt.Sprintf("%d hosts, %d open alerts, %d mail, %d push, %d dashboard deliveries, report %s, archive backlog %d, provider %s, billing %s.",
		summary.HostsTotal,
		summary.OpenAlerts,
		summary.MailDelivered,
		summary.PushDelivered,
		summary.DashboardDelivered,
		boolReady(summary.WeeklyReportReady),
		summary.ArchiveBacklog,
		boolReady(summary.ProviderReady),
		boolReady(summary.BillingReady),
	)
}

func revenueOperationsNextStep(summary model.TenantRevenueOperationsSummary, controlActions []model.TenantCustomerControlAction, settingsActions []model.TenantCustomerSettingsAction, packageActions []model.TenantPackageBillingAction, providerActions []model.TenantProviderSimulationAction) string {
	if summary.RoutesNeedingProof > 0 {
		return "Close mail, push, and dashboard route proof gaps before selling real-time anomaly assurance."
	}
	if summary.HighPriorityAlerts > 0 {
		return "Review high-priority anomaly rows and confirm owner notification proof."
	}
	if len(controlActions) > 0 {
		return firstNonEmpty(controlActions[0].Detail, controlActions[0].Title)
	}
	if len(settingsActions) > 0 {
		return firstNonEmpty(settingsActions[0].Detail, settingsActions[0].Title)
	}
	if len(packageActions) > 0 {
		return firstNonEmpty(packageActions[0].NextAction, packageActions[0].Detail, packageActions[0].Title)
	}
	if len(providerActions) > 0 {
		return firstNonEmpty(providerActions[0].Detail, providerActions[0].Title)
	}
	return "Use the Revenue Operations Center for the next paid customer review."
}

func revenueOperationsSignals(summary model.TenantRevenueOperationsSummary, onboarding model.TenantOnboardingCenter, settings model.TenantCustomerSettingsCenter, packageBilling model.TenantPackageBillingReadiness, provider model.TenantProviderSimulationLab) []model.TenantRevenueOperationsSignal {
	return []model.TenantRevenueOperationsSignal{
		{
			ID:       "anomaly-command",
			Label:    "Anomaly Command",
			Value:    fmt.Sprintf("%d open", summary.OpenAlerts),
			Detail:   fmt.Sprintf("%d high-priority alerts across %d hosts", summary.HighPriorityAlerts, summary.HostsTotal),
			Status:   countStatus(summary.OpenAlerts),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "mail-delivery",
			Label:    "Mail Delivery",
			Value:    fmt.Sprintf("%d delivered", summary.MailDelivered),
			Detail:   "Owner email proof for anomaly notifications and weekly report trust.",
			Status:   deliveryValueStatus(summary.MailDelivered),
			Channel:  constants.DeliveryChannelEmail,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "push-reach",
			Label:    "Push Reach",
			Value:    fmt.Sprintf("%d delivered", summary.PushDelivered),
			Detail:   fmt.Sprintf("%d route proof gaps, %d%% notification score", summary.RoutesNeedingProof, summary.NotificationScore),
			Status:   scoreStatus(summary.NotificationScore),
			Channel:  constants.DeliveryChannelPush,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "dashboard-fallback",
			Label:    "Dashboard Fallback",
			Value:    fmt.Sprintf("%d delivered", summary.DashboardDelivered),
			Detail:   "In-app anomaly and delivery state for admins when mail or push needs retry.",
			Status:   deliveryValueStatus(summary.DashboardDelivered),
			Channel:  constants.DeliveryChannelDashboard,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "weekly-report",
			Label:    "Weekly Report",
			Value:    boolReady(summary.WeeklyReportReady),
			Detail:   "PDF/email report readiness for paid family, school, and business reviews.",
			Status:   boolStatus(summary.WeeklyReportReady),
			Channel:  constants.DeliveryChannelEmail,
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "archive-retention",
			Label:    "Archive Retention",
			Value:    fmt.Sprintf("%d pending", summary.ArchiveBacklog),
			Detail:   "S3-backed retention, offline replay, and archive backlog proof.",
			Status:   archiveValueStatus(summary.ArchiveBacklog, true),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "setup-readiness",
			Label:    "Setup Readiness",
			Value:    fmt.Sprintf("%d%%", summary.OnboardingScore),
			Detail:   onboarding.Summary.Detail,
			Status:   onboarding.Summary.Status,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "customer-settings",
			Label:    "Customer Settings",
			Value:    fmt.Sprintf("%d%%", summary.SettingsScore),
			Detail:   settings.Summary.Detail,
			Status:   settings.Summary.Status,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "package-billing",
			Label:    "Package Billing",
			Value:    fmt.Sprintf("%d%%", summary.PackageScore),
			Detail:   fmt.Sprintf("%d/%d feature gates ready; billing %s", packageBilling.Summary.FeatureGatesReady, packageBilling.Summary.FeatureGatesTotal, boolReady(summary.BillingReady)),
			Status:   packageBilling.Summary.Status,
			PaidTier: constants.PlanBusiness,
		},
		{
			ID:       "provider-simulation",
			Label:    "Provider Simulation",
			Value:    fmt.Sprintf("%d routes", provider.Summary.SimulatedRoutes),
			Detail:   fmt.Sprintf("%d provider risks; %d%% readiness", provider.Summary.ProviderRisks, provider.Summary.ReadinessScore),
			Status:   provider.Summary.Status,
			PaidTier: constants.PlanBusiness,
		},
	}
}

func revenueOperationsAlerts(controlAlerts []model.TenantCustomerControlAlert, portfolioAlerts []model.TenantPortfolioAlertNotification) []model.TenantRevenueOperationsAlert {
	alerts := make([]model.TenantRevenueOperationsAlert, 0, 8)
	for _, item := range controlAlerts {
		alerts = append(alerts, model.TenantRevenueOperationsAlert{
			ID:              item.ID,
			Title:           item.Title,
			Detail:          item.Detail,
			Severity:        item.Severity,
			Status:          item.Status,
			HostName:        item.HostName,
			Category:        item.Category,
			EmailStatus:     item.EmailStatus,
			PushStatus:      item.PushStatus,
			DashboardStatus: item.DashboardStatus,
			NextAction:      item.NextAction,
			PaidTier:        item.PaidTier,
			ObservedAt:      item.ObservedAt,
		})
		if len(alerts) >= 6 {
			break
		}
	}
	for _, item := range portfolioAlerts {
		if len(alerts) >= 8 {
			break
		}
		if revenueOperationsHasAlert(alerts, item.Title, item.HostName) {
			continue
		}
		alerts = append(alerts, model.TenantRevenueOperationsAlert{
			ID:              fmt.Sprintf("portfolio-%d", len(alerts)+1),
			Title:           item.Title,
			Detail:          item.Detail,
			Severity:        item.Severity,
			Status:          item.Status,
			HostName:        item.HostName,
			Category:        item.Category,
			EmailStatus:     item.EmailStatus,
			PushStatus:      item.PushStatus,
			DashboardStatus: item.DashboardStatus,
			NextAction:      item.NextAction,
			PaidTier:        item.PaidTier,
			ObservedAt:      item.ObservedAt,
		})
	}
	return alerts
}

func revenueOperationsHasAlert(alerts []model.TenantRevenueOperationsAlert, title string, host string) bool {
	for _, alert := range alerts {
		if alert.Title == title && alert.HostName == host {
			return true
		}
	}
	return false
}

func revenueOperationsDeliveries(controlDeliveries []model.TenantCustomerControlDelivery, providerRoutes []model.TenantProviderSimulationRoute) []model.TenantRevenueOperationsDelivery {
	deliveries := make([]model.TenantRevenueOperationsDelivery, 0, 6)
	for _, item := range controlDeliveries {
		deliveries = append(deliveries, model.TenantRevenueOperationsDelivery{
			Channel:        item.Channel,
			Provider:       item.Provider,
			RecipientLabel: item.RecipientLabel,
			Status:         item.Status,
			ProofState:     item.ProofState,
			Attempts:       item.Attempts,
			LastDeliveryAt: item.LastDeliveryAt,
			SLA:            item.SLA,
			Evidence:       item.Evidence,
			NextAction:     item.NextAction,
			PaidTier:       item.PaidTier,
		})
		if len(deliveries) >= 3 {
			break
		}
	}
	for _, route := range providerRoutes {
		if len(deliveries) >= 6 {
			break
		}
		if revenueOperationsHasDelivery(deliveries, route.Channel) {
			continue
		}
		deliveries = append(deliveries, model.TenantRevenueOperationsDelivery{
			Channel:        route.Channel,
			Provider:       route.Provider,
			RecipientLabel: route.RecipientLabel,
			Status:         route.SimulationStatus,
			ProofState:     route.ProofState,
			LastDeliveryAt: route.LastSimulatedAt,
			SLA:            route.SLATarget,
			Evidence:       route.Evidence,
			NextAction:     route.NextAction,
			PaidTier:       route.PaidTier,
		})
	}
	return deliveries
}

func revenueOperationsHasDelivery(items []model.TenantRevenueOperationsDelivery, channel string) bool {
	for _, item := range items {
		if item.Channel == channel {
			return true
		}
	}
	return false
}

func revenueOperationsLevers(packageBilling model.TenantPackageBillingReadiness, pushActivation model.TenantPushActivationCenter, onboarding model.TenantOnboardingCenter, settings model.TenantCustomerSettingsCenter, provider model.TenantProviderSimulationLab) []model.TenantRevenueOperationsLever {
	levers := make([]model.TenantRevenueOperationsLever, 0, 8)
	for _, plan := range packageBilling.Plans {
		levers = append(levers, model.TenantRevenueOperationsLever{
			ID:           "plan-" + plan.PlanID,
			Name:         plan.Name,
			Tier:         plan.PlanID,
			Value:        fmt.Sprintf("%d%% fit", plan.FitScore),
			Status:       plan.Status,
			BuyerOutcome: plan.Value,
			NextAction:   plan.NextAction,
		})
		if len(levers) >= 4 {
			break
		}
	}
	levers = append(levers,
		model.TenantRevenueOperationsLever{
			ID:           "push-activation",
			Name:         "Push Notification Activation",
			Tier:         constants.PlanFamilyPro,
			Value:        fmt.Sprintf("%d%%", pushActivation.Summary.ActivationScore),
			Status:       pushActivation.Summary.Status,
			BuyerOutcome: "Urgent anomaly alerts feel immediate with mail and dashboard fallback.",
			NextAction:   pushActivation.Summary.OwnerNextStep,
		},
		model.TenantRevenueOperationsLever{
			ID:           "onboarding-readiness",
			Name:         "Managed Device Onboarding",
			Tier:         constants.PlanSchool,
			Value:        fmt.Sprintf("%d/%d steps", onboarding.Summary.SetupStepsReady, onboarding.Summary.SetupStepsTotal),
			Status:       onboarding.Summary.Status,
			BuyerOutcome: "Install, reboot persistence, archive, notification, and role handoff are packaged for admins.",
			NextAction:   onboarding.Summary.OwnerNextStep,
		},
		model.TenantRevenueOperationsLever{
			ID:           "settings-center",
			Name:         "Customer Settings",
			Tier:         constants.PlanBusiness,
			Value:        fmt.Sprintf("%d/%d settings", settings.Summary.ConfiguredSettings, settings.Summary.SettingsTotal),
			Status:       settings.Summary.Status,
			BuyerOutcome: "Customer admins can review plan, retention, notification, archive, role, and trust settings.",
			NextAction:   settings.Summary.OwnerNextStep,
		},
		model.TenantRevenueOperationsLever{
			ID:           "provider-simulation",
			Name:         "Provider Simulation",
			Tier:         constants.PlanBusiness,
			Value:        fmt.Sprintf("%d/%d routes", provider.Summary.SimulatedRoutes, provider.Summary.RoutesTotal),
			Status:       provider.Summary.Status,
			BuyerOutcome: "Delivery proof can be rehearsed without storing provider secrets or payloads.",
			NextAction:   provider.Summary.NextBestAction,
		},
	)
	return levers
}

func revenueOperationsActions(
	controlActions []model.TenantCustomerControlAction,
	successActions []model.TenantCustomerSuccessPacketAction,
	pushActions []model.TenantPushActivationAction,
	settingsActions []model.TenantCustomerSettingsAction,
	packageActions []model.TenantPackageBillingAction,
	providerActions []model.TenantProviderSimulationAction,
	generatedAt time.Time,
) []model.TenantRevenueOperationsAction {
	actions := make([]model.TenantRevenueOperationsAction, 0, 10)
	for _, action := range controlActions {
		if len(actions) >= 3 {
			break
		}
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   action.Severity,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range successActions {
		if len(actions) >= 5 {
			break
		}
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   action.Severity,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range pushActions {
		if len(actions) >= 7 {
			break
		}
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   action.Severity,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range settingsActions {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   action.Severity,
			SLA:        "before paid activation",
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: generatedAt,
		})
	}
	for _, action := range packageActions {
		if len(actions) >= 9 {
			break
		}
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      action.Title,
			Detail:     firstNonEmpty(action.NextAction, action.Detail, action.ConversionLever),
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   constants.SeverityInfo,
			SLA:        "before package review",
			PaidTier:   action.PaidTier,
			Source:     "package billing readiness",
			ObservedAt: generatedAt,
		})
	}
	for _, action := range providerActions {
		if len(actions) >= 10 {
			break
		}
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   constants.SeverityInfo,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     "provider simulation lab",
			ObservedAt: generatedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantRevenueOperationsAction{
			Title:      "Revenue operations review ready",
			Detail:     "Anomaly visibility, mail delivery, push reach, dashboard fallback, reports, archive, setup, settings, package, provider, and privacy proof are visible.",
			Owner:      constants.RoleBusinessManager,
			Status:     constants.StatusHealthy,
			Severity:   constants.SeverityInfo,
			SLA:        "weekly review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "revenue operations center",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func buildTenantDeploymentReadinessCenter(
	onboarding model.TenantOnboardingCenter,
	settings model.TenantCustomerSettingsCenter,
	syncHealth model.TenantSyncHealth,
	portfolio model.TenantPortfolioCenter,
	revenueOps model.TenantRevenueOperationsCenter,
	generatedAt time.Time,
) model.TenantDeploymentReadinessCenter {
	platforms := deploymentPlatforms(onboarding, settings)
	manifests := deploymentManifests(onboarding)
	platformsReady := countDeploymentPlatformReady(platforms)
	manifestsReady := countDeploymentManifestReady(manifests)
	liveBootReady := syncHealth.HostsReporting > 0 && syncHealth.BackendVisible
	autostartReady := onboarding.Summary.AutostartReady && settings.Summary.AutostartReady
	silentStartReady := autostartReady && platformsReady == len(platforms)
	offlineReplayReady := syncHealth.OfflineReplayReady
	archiveBacklog := maxInt(revenueOps.Summary.ArchiveBacklog, portfolio.Summary.ArchiveBacklog)
	readiness := averageScore(
		onboarding.Summary.ReadinessScore,
		settings.Summary.SettingsScore,
		boolScore(liveBootReady),
		boolScore(autostartReady),
		boolScore(silentStartReady),
		boolScore(offlineReplayReady),
		(manifestsReady*100)/maxInt(len(manifests), 1),
	)

	summary := model.TenantDeploymentReadinessSummary{
		ReadinessScore:     readiness,
		PlatformsReady:     platformsReady,
		PlatformsTotal:     len(platforms),
		ManifestsReady:     manifestsReady,
		ManifestsTotal:     len(manifests),
		LiveBootReady:      liveBootReady,
		AutostartReady:     autostartReady,
		SilentStartReady:   silentStartReady,
		OfflineReplayReady: offlineReplayReady,
		ArchiveBacklog:     archiveBacklog,
		HostsTotal:         maxInt(syncHealth.HostsTotal, onboarding.Summary.HostsTotal, portfolio.Summary.HostsTotal),
		HostsReporting:     maxInt(syncHealth.HostsReporting, onboarding.Summary.HostsReporting),
		RecommendedPackage: firstNonEmpty(revenueOps.Summary.RecommendedPaidPackage, onboarding.Summary.RecommendedPackage, settings.Summary.RecommendedPlan, revenueOps.PlanName),
	}
	summary.Status = deploymentReadinessStatus(summary)
	summary.Headline = deploymentReadinessHeadline(summary)
	summary.Detail = deploymentReadinessDetail(summary)
	summary.OwnerNextStep = deploymentReadinessNextStep(summary)

	return model.TenantDeploymentReadinessCenter{
		TenantID:        onboarding.TenantID,
		TenantName:      onboarding.TenantName,
		PlanID:          onboarding.PlanID,
		PlanName:        onboarding.PlanName,
		Audience:        firstNonEmpty(onboarding.Audience, settings.Audience, revenueOps.Audience),
		Summary:         summary,
		Platforms:       platforms,
		Manifests:       manifests,
		Proof:           deploymentProof(summary, syncHealth, onboarding, revenueOps),
		Actions:         deploymentActions(summary),
		PrivacyBoundary: constants.DeploymentReadinessPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func deploymentReadinessStatus(summary model.TenantDeploymentReadinessSummary) string {
	switch {
	case !summary.LiveBootReady || !summary.AutostartReady:
		return constants.StatusAttention
	case summary.ArchiveBacklog > 0 || !summary.OfflineReplayReady:
		return constants.StatusWatch
	case summary.ReadinessScore >= 85 && summary.PlatformsReady == summary.PlatformsTotal:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func deploymentReadinessHeadline(summary model.TenantDeploymentReadinessSummary) string {
	if !summary.LiveBootReady {
		return "Live boot proof is missing for deployed hosts"
	}
	if !summary.AutostartReady {
		return "Reboot persistence needs admin verification"
	}
	if summary.ArchiveBacklog > 0 {
		return fmt.Sprintf("%d offline/archive batch%s need replay proof", summary.ArchiveBacklog, pluralSuffix(summary.ArchiveBacklog))
	}
	return fmt.Sprintf("%s deployment is %d%% ready", firstNonEmpty(summary.RecommendedPackage, "TraceDeck"), summary.ReadinessScore)
}

func deploymentReadinessDetail(summary model.TenantDeploymentReadinessSummary) string {
	return fmt.Sprintf("%d/%d platforms ready, %d/%d manifests ready, %d/%d hosts reporting, live boot %s, autostart %s, background start %s, offline replay %s.",
		summary.PlatformsReady,
		summary.PlatformsTotal,
		summary.ManifestsReady,
		summary.ManifestsTotal,
		summary.HostsReporting,
		summary.HostsTotal,
		boolReady(summary.LiveBootReady),
		boolReady(summary.AutostartReady),
		boolReady(summary.SilentStartReady),
		boolReady(summary.OfflineReplayReady),
	)
}

func deploymentReadinessNextStep(summary model.TenantDeploymentReadinessSummary) string {
	switch {
	case !summary.LiveBootReady:
		return "Run the live boot smoke and verify the host reports to the local backend after restart."
	case !summary.AutostartReady:
		return "Register the native service and query Task Scheduler, launchd, or systemd status from scripts."
	case summary.ArchiveBacklog > 0:
		return "Let the agent replay local backlog and confirm S3 archive status returns to current."
	case !summary.OfflineReplayReady:
		return "Run offline replay verification before claiming laptop-off recovery."
	default:
		return "Keep native service manifests, status scripts, and live boot proof current before paid rollout."
	}
}

func deploymentPlatforms(onboarding model.TenantOnboardingCenter, settings model.TenantCustomerSettingsCenter) []model.TenantDeploymentPlatform {
	autostart := onboarding.Summary.AutostartReady && settings.Summary.AutostartReady
	status := boolStatus(autostart)
	return []model.TenantDeploymentPlatform{
		{
			Platform:       constants.PlatformWindows,
			ServiceManager: constants.ServiceManagerTaskScheduler,
			Manifest:       constants.WindowsTaskOutputPath,
			RegisterScript: "scripts/local/register-windows-task.ps1",
			StatusScript:   "scripts/local/get-windows-task-status.ps1",
			InstallMode:    "admin-approved Task Scheduler registration",
			Autostart:      "at logon through committed XML template",
			SilentStart:    "no foreground console after native task registration",
			Status:         status,
			Evidence:       "render-windows-task.ps1 and test-windows-task-template.ps1 verify XML before registration.",
			NextAction:     "Run register-windows-task.ps1 with UAC approval, then query task status.",
			PaidTier:       constants.PlanFamilyPro,
		},
		{
			Platform:       constants.PlatformDarwin,
			ServiceManager: constants.ServiceManagerLaunchd,
			Manifest:       constants.DarwinLaunchdOutput,
			RegisterScript: "scripts/local/manage-agent-service.ps1 -Platform darwin -Action install",
			StatusScript:   "scripts/local/manage-agent-service.ps1 -Platform darwin -Action status",
			InstallMode:    "launchd user service manifest",
			Autostart:      "launchctl bootstrap and enable",
			SilentStart:    "background launchd job with visible app consent/audit surfaces",
			Status:         status,
			Evidence:       "render-service-manifests.ps1 renders the launchd plist from a committed template.",
			NextAction:     "Dry-run the darwin install/status plan before native macOS rollout.",
			PaidTier:       constants.PlanSchool,
		},
		{
			Platform:       constants.PlatformLinux,
			ServiceManager: constants.ServiceManagerSystemd,
			Manifest:       constants.LinuxSystemdOutput,
			RegisterScript: "scripts/local/manage-agent-service.ps1 -Platform linux -Action install",
			StatusScript:   "scripts/local/manage-agent-service.ps1 -Platform linux -Action status",
			InstallMode:    "systemd unit with enable/start",
			Autostart:      "systemctl enable and start",
			SilentStart:    "background systemd unit with visible admin controls",
			Status:         status,
			Evidence:       "render-service-manifests.ps1 renders the systemd unit from a committed template.",
			NextAction:     "Dry-run the linux install/status plan before native Linux rollout.",
			PaidTier:       constants.PlanSchool,
		},
	}
}

func deploymentManifests(onboarding model.TenantOnboardingCenter) []model.TenantDeploymentManifest {
	status := boolStatus(onboarding.Summary.InstallReady)
	return []model.TenantDeploymentManifest{
		{
			ID:           "windows-task",
			Platform:     constants.PlatformWindows,
			TemplatePath: constants.WindowsTaskTemplatePath,
			OutputPath:   constants.WindowsTaskOutputPath,
			Manager:      constants.ServiceManagerTaskScheduler,
			Status:       status,
			Evidence:     "Windows scheduled task XML template is committed and parsed by local tests.",
			NextAction:   "Render the XML under data/local and register it through the scripted UAC flow.",
		},
		{
			ID:           "macos-launchd",
			Platform:     constants.PlatformDarwin,
			TemplatePath: constants.DarwinLaunchdTemplate,
			OutputPath:   constants.DarwinLaunchdOutput,
			Manager:      constants.ServiceManagerLaunchd,
			Status:       status,
			Evidence:     "macOS launchd plist template is committed and rendered by local scripts.",
			NextAction:   "Render and dry-run launchd install/status before native rollout.",
		},
		{
			ID:           "linux-systemd",
			Platform:     constants.PlatformLinux,
			TemplatePath: constants.LinuxSystemdTemplate,
			OutputPath:   constants.LinuxSystemdOutput,
			Manager:      constants.ServiceManagerSystemd,
			Status:       status,
			Evidence:     "Linux systemd unit template is committed and rendered by local scripts.",
			NextAction:   "Render and dry-run systemd install/status before native rollout.",
		},
	}
}

func deploymentProof(summary model.TenantDeploymentReadinessSummary, syncHealth model.TenantSyncHealth, onboarding model.TenantOnboardingCenter, revenueOps model.TenantRevenueOperationsCenter) []model.TenantDeploymentProof {
	return []model.TenantDeploymentProof{
		{
			ID:       "live-boot",
			Label:    "Live Boot Proof",
			Value:    boolReady(summary.LiveBootReady),
			Detail:   fmt.Sprintf("%d/%d hosts reporting after backend boot; %s", summary.HostsReporting, summary.HostsTotal, syncHealth.OfflineReplaySummary),
			Status:   boolStatus(summary.LiveBootReady),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "windows-task-scheduler",
			Label:    "Windows Task Scheduler",
			Value:    boolReady(summary.AutostartReady),
			Detail:   "Task XML, registration script, status query script, and service manager wrapper are available.",
			Status:   boolStatus(summary.AutostartReady),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "macos-linux-services",
			Label:    "macOS And Linux Services",
			Value:    fmt.Sprintf("%d/%d manifests", summary.ManifestsReady, summary.ManifestsTotal),
			Detail:   "launchd and systemd templates render from committed deployment assets.",
			Status:   scoreStatus((summary.ManifestsReady * 100) / maxInt(summary.ManifestsTotal, 1)),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "offline-replay",
			Label:    "Offline Replay",
			Value:    boolReady(summary.OfflineReplayReady),
			Detail:   firstNonEmpty(syncHealth.OfflineReplaySummary, "Offline laptop recovery replays local metadata batches when online."),
			Status:   boolStatus(summary.OfflineReplayReady),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "archive-backlog",
			Label:    "Archive Backlog",
			Value:    fmt.Sprintf("%d pending", summary.ArchiveBacklog),
			Detail:   fmt.Sprintf("Revenue operations archive posture: %d pending batch%s.", revenueOps.Summary.ArchiveBacklog, pluralSuffix(revenueOps.Summary.ArchiveBacklog)),
			Status:   archiveValueStatus(summary.ArchiveBacklog, true),
			PaidTier: constants.PlanBusiness,
		},
		{
			ID:       "setup-evidence",
			Label:    "Setup Evidence",
			Value:    fmt.Sprintf("%d/%d steps", onboarding.Summary.SetupStepsReady, onboarding.Summary.SetupStepsTotal),
			Detail:   onboarding.Summary.Detail,
			Status:   onboarding.Summary.Status,
			PaidTier: constants.PlanFamilyPro,
		},
	}
}

func deploymentActions(summary model.TenantDeploymentReadinessSummary) []model.TenantDeploymentAction {
	return []model.TenantDeploymentAction{
		{
			Title:    "Register Windows reboot persistence",
			Detail:   "Render Task Scheduler XML, approve the UAC registration, then query status and last run result through scripts.",
			Owner:    constants.RoleBusinessManager,
			Status:   boolStatus(summary.AutostartReady),
			Severity: constants.SeverityMedium,
			SLA:      "before Windows pilot rollout",
			PaidTier: constants.PlanFamilyPro,
			Source:   "windows task scheduler",
		},
		{
			Title:    "Dry-run macOS launchd rollout",
			Detail:   "Render the launchd plist and capture install/status dry-run proof before testing on a managed Mac.",
			Owner:    constants.RoleBusinessManager,
			Status:   scoreStatus((summary.ManifestsReady * 100) / maxInt(summary.ManifestsTotal, 1)),
			Severity: constants.SeverityInfo,
			SLA:      "before macOS beta",
			PaidTier: constants.PlanSchool,
			Source:   "launchd manifest",
		},
		{
			Title:    "Dry-run Linux systemd rollout",
			Detail:   "Render the systemd unit and capture enable/start/status dry-run proof before Linux rollout.",
			Owner:    constants.RoleBusinessManager,
			Status:   scoreStatus((summary.ManifestsReady * 100) / maxInt(summary.ManifestsTotal, 1)),
			Severity: constants.SeverityInfo,
			SLA:      "before Linux beta",
			PaidTier: constants.PlanSchool,
			Source:   "systemd manifest",
		},
		{
			Title:    "Verify restart recovery",
			Detail:   summary.OwnerNextStep,
			Owner:    constants.RoleBusinessManager,
			Status:   summary.Status,
			Severity: constants.SeverityHigh,
			SLA:      "before paid deployment",
			PaidTier: constants.PlanBusiness,
			Source:   "deployment readiness center",
		},
	}
}

func countDeploymentPlatformReady(items []model.TenantDeploymentPlatform) int {
	ready := 0
	for _, item := range items {
		if item.Status == constants.StatusHealthy {
			ready++
		}
	}
	return ready
}

func countDeploymentManifestReady(items []model.TenantDeploymentManifest) int {
	ready := 0
	for _, item := range items {
		if item.Status == constants.StatusHealthy {
			ready++
		}
	}
	return ready
}

func boolScore(ready bool) int {
	if ready {
		return 100
	}
	return 45
}

func maxInt(values ...int) int {
	maximum := 0
	for _, value := range values {
		if value > maximum {
			maximum = value
		}
	}
	return maximum
}

func buildTenantExecutiveConsole(
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	business model.TenantBusinessDashboard,
	roles model.TenantRoleExperience,
	commandCenter model.TenantNotificationCommandCenter,
	timeline model.TenantDeliveryTimeline,
	generatedAt time.Time,
) model.TenantExecutiveConsole {
	readiness := averageScore(business.Summary.ProductScore, monetization.ReadinessScore, roles.Summary.ReadinessScore)
	status := executiveConsoleStatus(business.Summary, commandCenter.Summary, readiness)
	nextAction := executiveNextAction(business.Actions, roles.Onboarding, commandCenter.Actions)
	summary := model.TenantExecutiveConsoleSummary{
		Status:                 status,
		Headline:               executiveHeadline(business.Summary, commandCenter.Summary, roles.Summary, readiness),
		Detail:                 executiveDetail(business.Summary, commandCenter.Summary, roles.Summary, timeline.Summary),
		ReadinessScore:         readiness,
		NotificationScore:      business.Summary.NotificationScore,
		TrustScore:             business.Summary.TrustScore,
		OpenAlerts:             business.Summary.OpenAlerts,
		HighPriorityAlerts:     business.Summary.HighPriorityAlerts,
		HostsTotal:             business.Summary.HostsTotal,
		HostsAttention:         business.Summary.HostsAttention,
		EmailDelivered:         business.Summary.MailDelivered,
		PushDelivered:          business.Summary.PushDelivered,
		DashboardDelivered:     business.Summary.DashboardDelivered,
		DeliveryFailed:         commandCenter.Summary.DeliveryFailed,
		DeliveryRetrying:       commandCenter.Summary.DeliveryRetrying,
		RoutesNeedingProof:     business.Summary.RoutesNeedingProof,
		WeeklyReportReady:      business.Summary.WeeklyReportReady,
		ArchiveBacklog:         business.Summary.ArchiveBacklog,
		RolesReady:             roles.Summary.RolesReady,
		RolesTotal:             roles.Summary.RolesTotal,
		RecommendedPaidPackage: firstNonEmpty(business.Summary.RecommendedPackage, commandCenter.Summary.RecommendedPaidPackage, monetization.PlanName, operations.PlanName),
		NextBestAction:         nextAction,
	}

	return model.TenantExecutiveConsole{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        monetization.Audience,
		Summary:         summary,
		Tiles:           executiveConsoleTiles(summary, business, roles, timeline),
		Alerts:          executiveConsoleAlerts(business.Alerts),
		Deliveries:      executiveConsoleDeliveries(commandCenter.Channels),
		Actions:         executiveConsoleActions(business.Actions, roles.Onboarding, commandCenter.Actions, generatedAt),
		PrivacyBoundary: constants.ExecutiveConsolePrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func executiveConsoleStatus(business model.TenantBusinessDashboardSummary, command model.TenantNotificationCommandCenterSummary, readiness int) string {
	switch {
	case command.DeliveryFailed > 0 || business.HighPriorityAlerts > 0:
		return constants.StatusAttention
	case business.RoutesNeedingProof > 0 || command.DeliveryRetrying > 0 || business.HostsAttention > 0:
		return constants.StatusWatch
	case readiness >= 75 && business.NotificationScore >= 65:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func executiveHeadline(business model.TenantBusinessDashboardSummary, command model.TenantNotificationCommandCenterSummary, roles model.TenantRoleExperienceSummary, readiness int) string {
	if business.HighPriorityAlerts > 0 {
		return fmt.Sprintf("%d high-priority alert%s need notification proof", business.HighPriorityAlerts, pluralSuffix(business.HighPriorityAlerts))
	}
	if command.DeliveryFailed+command.DeliveryRetrying > 0 {
		total := command.DeliveryFailed + command.DeliveryRetrying
		return fmt.Sprintf("%d delivery route%s need buyer-visible assurance", total, pluralSuffix(total))
	}
	if roles.RolesReady < roles.RolesTotal {
		return fmt.Sprintf("%d/%d role views ready for paid onboarding", roles.RolesReady, roles.RolesTotal)
	}
	return fmt.Sprintf("%s is %d%% ready for a paid walkthrough", firstNonEmpty(business.RecommendedPackage, command.RecommendedPaidPackage, "TraceDeck"), readiness)
}

func executiveDetail(business model.TenantBusinessDashboardSummary, command model.TenantNotificationCommandCenterSummary, roles model.TenantRoleExperienceSummary, timeline model.TenantDeliveryTimelineSummary) string {
	return fmt.Sprintf("%d open alerts, %d mail, %d push, %d dashboard deliveries, %d route proof gaps, %d/%d role views, %d timeline events, report %s, archive backlog %d.",
		business.OpenAlerts,
		business.MailDelivered,
		business.PushDelivered,
		business.DashboardDelivered,
		firstNonZero(business.RoutesNeedingProof, timeline.RouteProofGaps),
		roles.RolesReady,
		roles.RolesTotal,
		firstNonZero(timeline.Total, command.RoutesTotal),
		boolReady(business.WeeklyReportReady),
		business.ArchiveBacklog,
	)
}

func executiveNextAction(businessActions []model.TenantBusinessDashboardAction, roleOnboarding []model.TenantRoleOnboardingItem, commandActions []model.TenantNotificationCommandCenterAction) string {
	if len(businessActions) > 0 {
		return businessActions[0].Title
	}
	if len(commandActions) > 0 {
		return commandActions[0].Title
	}
	if len(roleOnboarding) > 0 {
		return roleOnboarding[0].Title
	}
	return "Use the executive console for the paid customer walkthrough."
}

func executiveConsoleTiles(summary model.TenantExecutiveConsoleSummary, business model.TenantBusinessDashboard, roles model.TenantRoleExperience, timeline model.TenantDeliveryTimeline) []model.TenantExecutiveConsoleTile {
	return []model.TenantExecutiveConsoleTile{
		{
			ID:       "commercial-readiness",
			Label:    "Commercial Readiness",
			Value:    fmt.Sprintf("%d%%", summary.ReadinessScore),
			Detail:   fmt.Sprintf("%s for %s", summary.RecommendedPaidPackage, business.Audience),
			Status:   scoreStatus(summary.ReadinessScore),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "anomaly-stream",
			Label:    "Anomaly Stream",
			Value:    fmt.Sprintf("%d open", summary.OpenAlerts),
			Detail:   fmt.Sprintf("%d high-priority, %d hosts need attention", summary.HighPriorityAlerts, summary.HostsAttention),
			Status:   countStatus(summary.OpenAlerts),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "mail-delivery",
			Label:    "Mail Delivery",
			Value:    fmt.Sprintf("%d delivered", summary.EmailDelivered),
			Detail:   "Email proof for alerts, weekly reports, and buyer trust.",
			Status:   deliveryValueStatus(summary.EmailDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "push-reach",
			Label:    "Push Reach",
			Value:    fmt.Sprintf("%d delivered", summary.PushDelivered),
			Detail:   "Push proof for urgent anomaly notification.",
			Status:   deliveryValueStatus(summary.PushDelivered),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "weekly-report",
			Label:    "Weekly Report",
			Value:    boolReady(summary.WeeklyReportReady),
			Detail:   "Study, coding, entertainment, anomaly, archive, and mail summary.",
			Status:   archiveValueStatus(summary.ArchiveBacklog, summary.WeeklyReportReady),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "archive-trust",
			Label:    "Archive Trust",
			Value:    fmt.Sprintf("%d pending", summary.ArchiveBacklog),
			Detail:   fmt.Sprintf("%d delivery timeline events with metadata-only proof", timeline.Summary.Total),
			Status:   gapStatus(summary.ArchiveBacklog),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "role-views",
			Label:    "Role Views",
			Value:    fmt.Sprintf("%d/%d ready", summary.RolesReady, summary.RolesTotal),
			Detail:   roles.Summary.Detail,
			Status:   roles.Summary.Status,
			PaidTier: constants.PlanBusiness,
		},
		{
			ID:       "route-proof",
			Label:    "Route Proof",
			Value:    fmt.Sprintf("%d gaps", summary.RoutesNeedingProof),
			Detail:   fmt.Sprintf("%d%% notification score, %d retrying, %d failed", summary.NotificationScore, summary.DeliveryRetrying, summary.DeliveryFailed),
			Status:   gapStatus(summary.RoutesNeedingProof),
			PaidTier: constants.PlanBusiness,
		},
	}
}

func executiveConsoleAlerts(alerts []model.TenantBusinessDashboardAlert) []model.TenantExecutiveConsoleAlert {
	items := make([]model.TenantExecutiveConsoleAlert, 0, len(alerts))
	for _, alert := range alerts {
		items = append(items, model.TenantExecutiveConsoleAlert{
			ID:              alert.ID,
			Title:           alert.Title,
			Detail:          alert.Detail,
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

func executiveConsoleDeliveries(channels []model.TenantNotificationCommandCenterChannel) []model.TenantExecutiveConsoleDelivery {
	items := make([]model.TenantExecutiveConsoleDelivery, 0, len(channels))
	for _, channel := range channels {
		items = append(items, model.TenantExecutiveConsoleDelivery{
			Channel:        channel.Channel,
			Provider:       channel.Provider,
			Status:         firstNonEmpty(channel.LatestDeliveryStatus, channel.RouteStatus),
			ProofState:     channel.ProofState,
			RecipientLabel: channel.Recipient,
			Attempts:       channel.Attempts,
			LastDeliveryAt: channel.LastDeliveryAt,
			SLA:            channel.SLA,
			Evidence:       channel.Evidence,
			NextAction:     channel.NextAction,
			PaidTier:       channel.PaidTier,
		})
	}
	return items
}

func executiveConsoleActions(
	businessActions []model.TenantBusinessDashboardAction,
	roleOnboarding []model.TenantRoleOnboardingItem,
	commandActions []model.TenantNotificationCommandCenterAction,
	generatedAt time.Time,
) []model.TenantExecutiveConsoleAction {
	actions := make([]model.TenantExecutiveConsoleAction, 0, 8)
	for _, action := range businessActions {
		if len(actions) >= 4 {
			break
		}
		actions = append(actions, model.TenantExecutiveConsoleAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Severity:   action.Severity,
			Status:     action.Status,
			Owner:      action.Owner,
			Channel:    action.Channel,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range commandActions {
		if len(actions) >= 6 {
			break
		}
		actions = append(actions, model.TenantExecutiveConsoleAction{
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
	for _, item := range roleOnboarding {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantExecutiveConsoleAction{
			Title:      item.Title,
			Detail:     item.Detail,
			Severity:   constants.SeverityInfo,
			Status:     item.Status,
			Owner:      item.Owner,
			Channel:    constants.DeliveryChannelDashboard,
			SLA:        "before onboarding",
			PaidTier:   item.PaidTier,
			Source:     "role experience center",
			ObservedAt: generatedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantExecutiveConsoleAction{
			Title:      "Executive console ready",
			Detail:     "Anomaly stream, push reach, mail proof, weekly reports, archive trust, and paid package evidence are visible.",
			Severity:   constants.SeverityInfo,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			Channel:    constants.DeliveryChannelDashboard,
			SLA:        "weekly review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "executive console",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func buildTenantCustomerControlRoom(
	operations model.TenantOperationsSummary,
	business model.TenantBusinessDashboard,
	executive model.TenantExecutiveConsole,
	packageBilling model.TenantPackageBillingReadiness,
	provider model.TenantProviderSimulationLab,
	generatedAt time.Time,
) model.TenantCustomerControlRoom {
	summary := model.TenantCustomerControlSummary{
		ProductScore:           averageScore(executive.Summary.ReadinessScore, business.Summary.ProductScore, packageBilling.Summary.PackageScore),
		NotificationScore:      operations.NotificationScore,
		PackageScore:           packageBilling.Summary.PackageScore,
		TrustScore:             business.Summary.TrustScore,
		CustomerHealth:         operations.CustomerHealth,
		OpenAlerts:             business.Summary.OpenAlerts,
		HighPriorityAlerts:     business.Summary.HighPriorityAlerts,
		HostsTotal:             operations.HostsTotal,
		HostsAttention:         operations.HostsAttention,
		MailDelivered:          operations.EmailDelivered,
		PushDelivered:          operations.PushDelivered,
		DashboardDelivered:     operations.DashboardDelivered,
		DeliveryFailed:         operations.DeliveryFailed,
		DeliveryRetrying:       operations.DeliveryRetrying,
		RoutesNeedingProof:     business.Summary.RoutesNeedingProof,
		WeeklyReportReady:      business.Summary.WeeklyReportReady,
		ArchiveBacklog:         operations.ArchiveBacklog,
		BillingReady:           packageBilling.Summary.BillingReady,
		ProviderReady:          packageBilling.Summary.ProviderReady && provider.Summary.ProviderRisks == 0,
		RecommendedPaidPackage: firstNonEmpty(packageBilling.Summary.RecommendedPackage, executive.Summary.RecommendedPaidPackage, business.Summary.RecommendedPackage, operations.PlanName),
	}
	summary.Status = customerControlStatus(summary)
	summary.Headline = customerControlHeadline(summary)
	summary.Detail = customerControlDetail(summary)
	summary.NextBestAction = customerControlNextAction(business.Actions, executive.Actions, packageBilling.Actions, summary)

	return model.TenantCustomerControlRoom{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        business.Audience,
		Summary:         summary,
		Tiles:           customerControlTiles(summary, business, packageBilling, provider),
		Alerts:          customerControlAlerts(executive.Alerts),
		Deliveries:      customerControlDeliveries(executive.Deliveries, provider.Routes),
		Actions:         customerControlActions(business.Actions, executive.Actions, packageBilling.Actions, generatedAt),
		PrivacyBoundary: constants.CustomerControlPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func customerControlStatus(summary model.TenantCustomerControlSummary) string {
	switch {
	case summary.DeliveryFailed > 0 || summary.HighPriorityAlerts > 0:
		return constants.StatusAttention
	case summary.DeliveryRetrying > 0 || summary.RoutesNeedingProof > 0 || summary.HostsAttention > 0 || !summary.ProviderReady:
		return constants.StatusWatch
	case summary.ProductScore >= 80 && summary.NotificationScore >= 70 && summary.BillingReady:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func customerControlHeadline(summary model.TenantCustomerControlSummary) string {
	if summary.HighPriorityAlerts > 0 {
		return fmt.Sprintf("%d high-priority anomalies need owner-visible notification proof", summary.HighPriorityAlerts)
	}
	if summary.DeliveryFailed+summary.DeliveryRetrying > 0 {
		return fmt.Sprintf("%d notification deliveries need recovery before a paid demo", summary.DeliveryFailed+summary.DeliveryRetrying)
	}
	if !summary.BillingReady {
		return fmt.Sprintf("%s needs billing and package proof before monetisation", firstNonEmpty(summary.RecommendedPaidPackage, "TraceDeck"))
	}
	return fmt.Sprintf("%s is %d%% ready for a customer demo", firstNonEmpty(summary.RecommendedPaidPackage, "TraceDeck"), summary.ProductScore)
}

func customerControlDetail(summary model.TenantCustomerControlSummary) string {
	return fmt.Sprintf("%d open alerts, %d mail, %d push, %d dashboard deliveries, %d route proof gaps, report %s, archive backlog %d, package score %d%%.",
		summary.OpenAlerts,
		summary.MailDelivered,
		summary.PushDelivered,
		summary.DashboardDelivered,
		summary.RoutesNeedingProof,
		boolReady(summary.WeeklyReportReady),
		summary.ArchiveBacklog,
		summary.PackageScore,
	)
}

func customerControlNextAction(
	businessActions []model.TenantBusinessDashboardAction,
	executiveActions []model.TenantExecutiveConsoleAction,
	packageActions []model.TenantPackageBillingAction,
	summary model.TenantCustomerControlSummary,
) string {
	if len(businessActions) > 0 {
		return firstNonEmpty(businessActions[0].Detail, businessActions[0].Title)
	}
	if len(executiveActions) > 0 {
		return firstNonEmpty(executiveActions[0].Detail, executiveActions[0].Title)
	}
	if len(packageActions) > 0 {
		return firstNonEmpty(packageActions[0].Detail, packageActions[0].Title)
	}
	if !summary.ProviderReady {
		return "Run provider-safe mail and push rehearsal before promising anomaly notification SLA."
	}
	return "Use Customer Control Room as the first screen for buyer demos and owner reviews."
}

func customerControlTiles(
	summary model.TenantCustomerControlSummary,
	business model.TenantBusinessDashboard,
	packageBilling model.TenantPackageBillingReadiness,
	provider model.TenantProviderSimulationLab,
) []model.TenantCustomerControlTile {
	return []model.TenantCustomerControlTile{
		{
			ID:       "anomaly-command",
			Label:    "Anomaly Command",
			Value:    fmt.Sprintf("%d open", summary.OpenAlerts),
			Detail:   fmt.Sprintf("%d high-priority, %d hosts need attention", summary.HighPriorityAlerts, summary.HostsAttention),
			Status:   countStatus(summary.OpenAlerts),
			Channel:  constants.DeliveryChannelDashboard,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "mail-delivery",
			Label:    "Mail Delivery",
			Value:    fmt.Sprintf("%d delivered", summary.MailDelivered),
			Detail:   "Email proof for anomaly alerts, weekly reports, and owner trust.",
			Status:   deliveryValueStatus(summary.MailDelivered),
			Channel:  constants.DeliveryChannelEmail,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "push-reach",
			Label:    "Push Reach",
			Value:    fmt.Sprintf("%d delivered", summary.PushDelivered),
			Detail:   "Push proof for urgent non-study, media, tamper, and risky software alerts.",
			Status:   deliveryValueStatus(summary.PushDelivered),
			Channel:  constants.DeliveryChannelPush,
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "provider-simulation",
			Label:    "Provider Simulation",
			Value:    fmt.Sprintf("%d routes", provider.Summary.SimulatedRoutes),
			Detail:   fmt.Sprintf("%d provider risks, %d%% readiness", provider.Summary.ProviderRisks, provider.Summary.ReadinessScore),
			Status:   provider.Summary.Status,
			Channel:  constants.DeliveryChannelDashboard,
			PaidTier: constants.PlanBusiness,
		},
		{
			ID:       "weekly-report",
			Label:    "Weekly Report",
			Value:    boolReady(summary.WeeklyReportReady),
			Detail:   "Email/PDF readiness for study, coding, anomaly, app, and archive summary.",
			Status:   boolStatus(summary.WeeklyReportReady),
			Channel:  constants.DeliveryChannelEmail,
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "archive-retention",
			Label:    "Archive Retention",
			Value:    fmt.Sprintf("%d pending", summary.ArchiveBacklog),
			Detail:   "S3-backed retention proof for Family Pro, school, and business archive plans.",
			Status:   gapStatus(summary.ArchiveBacklog),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "package-billing",
			Label:    "Package Billing",
			Value:    fmt.Sprintf("%d%%", packageBilling.Summary.PackageScore),
			Detail:   fmt.Sprintf("%d/%d gates ready, billing %s", packageBilling.Summary.FeatureGatesReady, packageBilling.Summary.FeatureGatesTotal, boolReady(packageBilling.Summary.BillingReady)),
			Status:   packageBilling.Summary.Status,
			PaidTier: constants.PlanBusiness,
		},
		{
			ID:       "customer-health",
			Label:    "Customer Health",
			Value:    titleWord(summary.CustomerHealth),
			Detail:   business.Summary.Detail,
			Status:   summary.CustomerHealth,
			PaidTier: constants.PlanBusiness,
		},
	}
}

func customerControlAlerts(alerts []model.TenantExecutiveConsoleAlert) []model.TenantCustomerControlAlert {
	items := make([]model.TenantCustomerControlAlert, 0, len(alerts))
	for _, alert := range alerts {
		items = append(items, model.TenantCustomerControlAlert{
			ID:              alert.ID,
			Title:           alert.Title,
			Detail:          alert.Detail,
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

func customerControlDeliveries(executiveDeliveries []model.TenantExecutiveConsoleDelivery, providerRoutes []model.TenantProviderSimulationRoute) []model.TenantCustomerControlDelivery {
	items := make([]model.TenantCustomerControlDelivery, 0, len(executiveDeliveries))
	for _, delivery := range executiveDeliveries {
		items = append(items, model.TenantCustomerControlDelivery{
			Channel:        delivery.Channel,
			Provider:       delivery.Provider,
			RecipientLabel: delivery.RecipientLabel,
			Status:         delivery.Status,
			ProofState:     delivery.ProofState,
			Attempts:       delivery.Attempts,
			LastDeliveryAt: delivery.LastDeliveryAt,
			SLA:            delivery.SLA,
			Evidence:       delivery.Evidence,
			NextAction:     delivery.NextAction,
			PaidTier:       delivery.PaidTier,
		})
	}
	for _, route := range providerRoutes {
		if customerControlHasDelivery(items, route.Channel) {
			continue
		}
		items = append(items, model.TenantCustomerControlDelivery{
			Channel:        route.Channel,
			Provider:       route.Provider,
			RecipientLabel: route.RecipientLabel,
			Status:         route.SimulationStatus,
			ProofState:     route.ProofState,
			Attempts:       0,
			LastDeliveryAt: route.LastSimulatedAt,
			SLA:            route.SLATarget,
			Evidence:       route.Evidence,
			NextAction:     route.NextAction,
			PaidTier:       route.PaidTier,
		})
	}
	if len(items) > 6 {
		return items[:6]
	}
	return items
}

func customerControlHasDelivery(items []model.TenantCustomerControlDelivery, channel string) bool {
	for _, item := range items {
		if item.Channel == channel {
			return true
		}
	}
	return false
}

func customerControlActions(
	businessActions []model.TenantBusinessDashboardAction,
	executiveActions []model.TenantExecutiveConsoleAction,
	packageActions []model.TenantPackageBillingAction,
	generatedAt time.Time,
) []model.TenantCustomerControlAction {
	actions := make([]model.TenantCustomerControlAction, 0, 8)
	for _, action := range businessActions {
		if len(actions) >= 3 {
			break
		}
		actions = append(actions, model.TenantCustomerControlAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Severity:   action.Severity,
			Status:     action.Status,
			Owner:      action.Owner,
			Channel:    action.Channel,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range executiveActions {
		if len(actions) >= 5 {
			break
		}
		actions = append(actions, model.TenantCustomerControlAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Severity:   action.Severity,
			Status:     action.Status,
			Owner:      action.Owner,
			Channel:    action.Channel,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range packageActions {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantCustomerControlAction{
			Title:      action.Title,
			Detail:     firstNonEmpty(action.Detail, action.NextAction, action.ConversionLever),
			Severity:   constants.SeverityInfo,
			Status:     action.Status,
			Owner:      action.Owner,
			Channel:    constants.DeliveryChannelDashboard,
			SLA:        "before paid review",
			PaidTier:   action.PaidTier,
			Source:     "package billing readiness",
			ObservedAt: generatedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantCustomerControlAction{
			Title:      "Customer control room ready",
			Detail:     "Anomaly stream, mail proof, push reach, provider simulation, reports, archive, and package billing are visible.",
			Severity:   constants.SeverityInfo,
			Status:     constants.StatusHealthy,
			Owner:      constants.RoleBusinessManager,
			Channel:    constants.DeliveryChannelDashboard,
			SLA:        "weekly review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "customer control room",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func buildTenantCustomerSuccessPacket(
	controlRoom model.TenantCustomerControlRoom,
	packageBilling model.TenantPackageBillingReadiness,
	provider model.TenantProviderSimulationLab,
	roles model.TenantRoleExperience,
	generatedAt time.Time,
) model.TenantCustomerSuccessPacket {
	summary := model.TenantCustomerSuccessPacketSummary{
		ReadinessScore:         averageScore(controlRoom.Summary.ProductScore, packageBilling.Summary.PackageScore, roles.Summary.ReadinessScore),
		NotificationScore:      controlRoom.Summary.NotificationScore,
		PackageScore:           packageBilling.Summary.PackageScore,
		TrustScore:             controlRoom.Summary.TrustScore,
		OpenAlerts:             controlRoom.Summary.OpenAlerts,
		HighPriorityAlerts:     controlRoom.Summary.HighPriorityAlerts,
		HostsTotal:             controlRoom.Summary.HostsTotal,
		MailDelivered:          controlRoom.Summary.MailDelivered,
		PushDelivered:          controlRoom.Summary.PushDelivered,
		RoutesNeedingProof:     controlRoom.Summary.RoutesNeedingProof,
		WeeklyReportReady:      controlRoom.Summary.WeeklyReportReady,
		ArchiveBacklog:         controlRoom.Summary.ArchiveBacklog,
		ProviderReady:          controlRoom.Summary.ProviderReady,
		BillingReady:           packageBilling.Summary.BillingReady,
		RolesReady:             roles.Summary.RolesReady,
		RolesTotal:             roles.Summary.RolesTotal,
		RecommendedPaidPackage: firstNonEmpty(controlRoom.Summary.RecommendedPaidPackage, packageBilling.Summary.RecommendedPackage, roles.Summary.RecommendedPackage, controlRoom.PlanName),
	}
	summary.Status = customerSuccessPacketStatus(summary)
	summary.Headline = customerSuccessPacketHeadline(summary)
	summary.Detail = customerSuccessPacketDetail(summary)
	summary.OwnerNextStep = customerSuccessPacketNextStep(controlRoom.Actions, packageBilling.Actions, summary)

	return model.TenantCustomerSuccessPacket{
		TenantID:        controlRoom.TenantID,
		TenantName:      controlRoom.TenantName,
		PlanID:          controlRoom.PlanID,
		PlanName:        controlRoom.PlanName,
		Audience:        controlRoom.Audience,
		Summary:         summary,
		Proofs:          customerSuccessPacketProofs(summary, controlRoom, packageBilling, provider, roles),
		Objections:      customerSuccessPacketObjections(summary, packageBilling, provider, roles),
		Actions:         customerSuccessPacketActions(controlRoom.Actions, packageBilling.Actions, generatedAt),
		PrivacyBoundary: constants.CustomerSuccessPacketPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func customerSuccessPacketStatus(summary model.TenantCustomerSuccessPacketSummary) string {
	switch {
	case summary.HighPriorityAlerts > 0 || summary.RoutesNeedingProof > 0:
		return constants.StatusAttention
	case !summary.ProviderReady || !summary.BillingReady || summary.ArchiveBacklog > 0:
		return constants.StatusWatch
	case summary.ReadinessScore >= 80 && summary.NotificationScore >= 70:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func customerSuccessPacketHeadline(summary model.TenantCustomerSuccessPacketSummary) string {
	if summary.HighPriorityAlerts > 0 {
		return fmt.Sprintf("%d urgent alert%s anchor the customer success packet", summary.HighPriorityAlerts, pluralSuffix(summary.HighPriorityAlerts))
	}
	if summary.RoutesNeedingProof > 0 {
		return fmt.Sprintf("%d route proof gap%s need closure before renewal review", summary.RoutesNeedingProof, pluralSuffix(summary.RoutesNeedingProof))
	}
	return fmt.Sprintf("%s success packet is %d%% ready", firstNonEmpty(summary.RecommendedPaidPackage, "TraceDeck"), summary.ReadinessScore)
}

func customerSuccessPacketDetail(summary model.TenantCustomerSuccessPacketSummary) string {
	return fmt.Sprintf("%d open alerts, %d mail delivered, %d push delivered, report %s, archive backlog %d, billing %s, provider %s, role views %d/%d.",
		summary.OpenAlerts,
		summary.MailDelivered,
		summary.PushDelivered,
		boolReady(summary.WeeklyReportReady),
		summary.ArchiveBacklog,
		boolReady(summary.BillingReady),
		boolReady(summary.ProviderReady),
		summary.RolesReady,
		summary.RolesTotal,
	)
}

func customerSuccessPacketNextStep(controlActions []model.TenantCustomerControlAction, packageActions []model.TenantPackageBillingAction, summary model.TenantCustomerSuccessPacketSummary) string {
	if len(controlActions) > 0 {
		return firstNonEmpty(controlActions[0].Detail, controlActions[0].Title)
	}
	if len(packageActions) > 0 {
		return firstNonEmpty(packageActions[0].NextAction, packageActions[0].Detail, packageActions[0].Title)
	}
	if !summary.ProviderReady {
		return "Run provider-safe delivery proof before sending the customer packet."
	}
	return "Share the success packet during the next paid customer review."
}

func customerSuccessPacketProofs(
	summary model.TenantCustomerSuccessPacketSummary,
	controlRoom model.TenantCustomerControlRoom,
	packageBilling model.TenantPackageBillingReadiness,
	provider model.TenantProviderSimulationLab,
	roles model.TenantRoleExperience,
) []model.TenantCustomerSuccessPacketProof {
	return []model.TenantCustomerSuccessPacketProof{
		{
			ID:          "anomaly-command",
			Label:       "Anomaly command",
			Value:       fmt.Sprintf("%d open", summary.OpenAlerts),
			Detail:      fmt.Sprintf("%d high-priority signals across %d hosts", summary.HighPriorityAlerts, summary.HostsTotal),
			Status:      countStatus(summary.OpenAlerts),
			Evidence:    firstNonEmpty(controlRoom.Summary.Headline, "Customer Control Room alert wall"),
			PaidTier:    constants.PlanFamilyPro,
			BuyerImpact: "Shows the customer what needs attention without exposing private content.",
		},
		{
			ID:          "mail-delivery",
			Label:       "Mail delivery",
			Value:       fmt.Sprintf("%d delivered", summary.MailDelivered),
			Detail:      "Email proof for critical alerts and weekly report trust.",
			Status:      deliveryValueStatus(summary.MailDelivered),
			Evidence:    "Provider-safe mail route metadata",
			PaidTier:    constants.PlanFamilyPro,
			BuyerImpact: "Proves alerts can reach the owner by email.",
		},
		{
			ID:          "push-notification",
			Label:       "Push notification",
			Value:       fmt.Sprintf("%d delivered", summary.PushDelivered),
			Detail:      fmt.Sprintf("%d route proof gaps, %d%% notification score", summary.RoutesNeedingProof, summary.NotificationScore),
			Status:      scoreStatus(summary.NotificationScore),
			Evidence:    "Push route status and retry metadata",
			PaidTier:    constants.PlanFamilyPro,
			BuyerImpact: "Makes urgent anomaly awareness feel immediate.",
		},
		{
			ID:          "report-archive",
			Label:       "Report and archive",
			Value:       boolReady(summary.WeeklyReportReady),
			Detail:      fmt.Sprintf("%d archive batches pending", summary.ArchiveBacklog),
			Status:      archiveValueStatus(summary.ArchiveBacklog, summary.WeeklyReportReady),
			Evidence:    "Weekly report readiness and S3 archive posture",
			PaidTier:    constants.PlanSchool,
			BuyerImpact: "Supports retention, review meetings, and school/business reporting.",
		},
		{
			ID:          "package-billing",
			Label:       "Package fit",
			Value:       fmt.Sprintf("%d%%", packageBilling.Summary.PackageScore),
			Detail:      fmt.Sprintf("%d/%d gates ready", packageBilling.Summary.FeatureGatesReady, packageBilling.Summary.FeatureGatesTotal),
			Status:      packageBilling.Summary.Status,
			Evidence:    packageBilling.Summary.Headline,
			PaidTier:    constants.PlanBusiness,
			BuyerImpact: "Connects feature proof to Family Pro, School, and Business packages.",
		},
		{
			ID:          "provider-simulation",
			Label:       "Provider simulation",
			Value:       fmt.Sprintf("%d routes", provider.Summary.SimulatedRoutes),
			Detail:      fmt.Sprintf("%d provider risks, %d%% readiness", provider.Summary.ProviderRisks, provider.Summary.ReadinessScore),
			Status:      provider.Summary.Status,
			Evidence:    provider.Summary.Headline,
			PaidTier:    constants.PlanBusiness,
			BuyerImpact: "Proves notification routes can be rehearsed without storing secrets or payloads.",
		},
		{
			ID:          "role-onboarding",
			Label:       "Role onboarding",
			Value:       fmt.Sprintf("%d/%d ready", roles.Summary.RolesReady, roles.Summary.RolesTotal),
			Detail:      roles.Summary.Detail,
			Status:      roles.Summary.Status,
			Evidence:    roles.Summary.Headline,
			PaidTier:    constants.PlanBusiness,
			BuyerImpact: "Lets the same product support parent, student, school, and business views.",
		},
	}
}

func customerSuccessPacketObjections(
	summary model.TenantCustomerSuccessPacketSummary,
	packageBilling model.TenantPackageBillingReadiness,
	provider model.TenantProviderSimulationLab,
	roles model.TenantRoleExperience,
) []model.TenantCustomerSuccessPacketObjection {
	return []model.TenantCustomerSuccessPacketObjection{
		{
			ID:       "privacy-boundary",
			Concern:  "Will this expose private content?",
			Answer:   "The packet uses metadata-only proof and excludes passwords, screenshots, raw URLs, alert bodies, provider secrets, and payment data.",
			Status:   constants.StatusHealthy,
			Evidence: constants.CustomerSuccessPacketPrivacyNote,
			Owner:    constants.RoleBusinessManager,
		},
		{
			ID:       "notification-reliability",
			Concern:  "Will anomaly notifications actually reach someone?",
			Answer:   fmt.Sprintf("%d mail deliveries, %d push deliveries, and %d route proof gaps are visible before renewal.", summary.MailDelivered, summary.PushDelivered, summary.RoutesNeedingProof),
			Status:   proofGapStatus(summary.RoutesNeedingProof),
			Evidence: provider.Summary.Headline,
			Owner:    constants.RoleParent,
		},
		{
			ID:       "billing-readiness",
			Concern:  "Which package should this customer buy?",
			Answer:   fmt.Sprintf("%s is recommended with %d%% package score and %d/%d feature gates ready.", firstNonEmpty(summary.RecommendedPaidPackage, packageBilling.Summary.RecommendedPackage), summary.PackageScore, packageBilling.Summary.FeatureGatesReady, packageBilling.Summary.FeatureGatesTotal),
			Status:   packageBilling.Summary.Status,
			Evidence: packageBilling.Summary.NextBestAction,
			Owner:    constants.RoleBusinessManager,
		},
		{
			ID:       "role-fit",
			Concern:  "Can this work for family, school, and business buyers?",
			Answer:   fmt.Sprintf("%d/%d role experiences are ready with role-specific onboarding actions.", roles.Summary.RolesReady, roles.Summary.RolesTotal),
			Status:   roles.Summary.Status,
			Evidence: roles.Summary.Headline,
			Owner:    constants.RoleSchoolAdmin,
		},
	}
}

func customerSuccessPacketActions(controlActions []model.TenantCustomerControlAction, packageActions []model.TenantPackageBillingAction, generatedAt time.Time) []model.TenantCustomerSuccessPacketAction {
	actions := make([]model.TenantCustomerSuccessPacketAction, 0, 8)
	for _, action := range controlActions {
		if len(actions) >= 5 {
			break
		}
		actions = append(actions, model.TenantCustomerSuccessPacketAction{
			Title:      action.Title,
			Detail:     action.Detail,
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   action.Severity,
			SLA:        action.SLA,
			PaidTier:   action.PaidTier,
			Source:     action.Source,
			ObservedAt: action.ObservedAt,
		})
	}
	for _, action := range packageActions {
		if len(actions) >= 8 {
			break
		}
		actions = append(actions, model.TenantCustomerSuccessPacketAction{
			Title:      action.Title,
			Detail:     firstNonEmpty(action.NextAction, action.Detail, action.ConversionLever),
			Owner:      action.Owner,
			Status:     action.Status,
			Severity:   constants.SeverityInfo,
			SLA:        "before customer review",
			PaidTier:   action.PaidTier,
			Source:     "package billing readiness",
			ObservedAt: generatedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantCustomerSuccessPacketAction{
			Title:      "Share customer success packet",
			Detail:     "Use the packet in the next customer review with anomaly, delivery, archive, package, and privacy proof.",
			Owner:      constants.RoleBusinessManager,
			Status:     constants.StatusHealthy,
			Severity:   constants.SeverityInfo,
			SLA:        "weekly review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "customer success packet",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func buildTenantPushActivationCenter(
	operations model.TenantOperationsSummary,
	preferences model.NotificationPreferenceCenter,
	drilldown model.TenantDeliveryDrilldown,
	provider model.TenantProviderSimulationLab,
	remediation model.TenantDeliveryRemediation,
	inbox model.TenantAlertInbox,
	timeline model.TenantDeliveryTimeline,
	packageBilling model.TenantPackageBillingReadiness,
	generatedAt time.Time,
) model.TenantPushActivationCenter {
	routes := pushActivationRoutes(drilldown.Routes, provider.Routes, remediation.Actions)
	summary := pushActivationSummary(operations, preferences, drilldown, provider, remediation, inbox, timeline, packageBilling, routes)
	summary.Status = pushActivationStatus(summary)
	summary.Headline = pushActivationHeadline(summary)
	summary.Detail = pushActivationDetail(summary)
	summary.OwnerNextStep = pushActivationNextStep(summary)

	return model.TenantPushActivationCenter{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        firstNonEmpty(packageBilling.Audience, preferences.Audience, constants.PlanFamilyPro),
		Summary:         summary,
		Routes:          routes,
		Scenarios:       pushActivationScenarios(summary),
		Actions:         pushActivationActions(summary, routes, generatedAt),
		PrivacyBoundary: constants.PushActivationPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func pushActivationSummary(
	operations model.TenantOperationsSummary,
	preferences model.NotificationPreferenceCenter,
	drilldown model.TenantDeliveryDrilldown,
	provider model.TenantProviderSimulationLab,
	remediation model.TenantDeliveryRemediation,
	inbox model.TenantAlertInbox,
	timeline model.TenantDeliveryTimeline,
	packageBilling model.TenantPackageBillingReadiness,
	routes []model.TenantPushActivationRoute,
) model.TenantPushActivationSummary {
	summary := model.TenantPushActivationSummary{
		NotificationScore:      operations.NotificationScore,
		MailDelivered:          operations.EmailDelivered,
		DashboardDelivered:     operations.DashboardDelivered,
		PushPreferenceEnabled:  preferences.Summary.PushEnabled,
		PushEscalationEnabled:  stringSliceHas(preferences.Escalation.Channels, constants.DeliveryChannelPush),
		QuietHoursProtected:    preferences.Summary.QuietHoursEnabled,
		PushSimulationReady:    provider.Summary.PushReady,
		AlertRulesUsingPush:    pushPreferenceRules(preferences.Rules),
		AlertsWithPush:         inbox.Summary.WithPush,
		RecommendedPaidPackage: firstNonEmpty(packageBilling.Summary.RecommendedPackage, provider.Summary.RecommendedPaidPackage, operations.PlanName, constants.PlanFamilyPro),
	}
	for _, item := range timeline.Items {
		switch item.Status {
		case constants.DeliveryStatusDelivered:
			summary.PushDelivered++
		case constants.DeliveryStatusRetrying:
			summary.PushRetrying++
		case constants.DeliveryStatusFailed:
			summary.PushFailed++
		case constants.DeliveryStatusPending:
			summary.PushPending++
		}
	}
	for _, route := range routes {
		summary.PushRoutesTotal++
		if route.ProofState == constants.DeliveryProofStateCustomer || route.ProofState == constants.DeliveryProofStateRehearsed || route.SimulationStatus == constants.StatusHealthy {
			summary.PushRoutesReady++
		} else {
			summary.PushRoutesNeedingProof++
		}
	}
	if summary.PushRoutesTotal == 0 {
		for _, route := range drilldown.Routes {
			if route.Channel == constants.DeliveryChannelPush {
				summary.PushRoutesTotal++
				summary.PushRoutesNeedingProof++
			}
		}
	}
	routeScore := 0
	if summary.PushRoutesTotal > 0 {
		routeScore = (summary.PushRoutesReady * 100) / summary.PushRoutesTotal
	}
	deliveryScore := deliveryValueScore(summary.PushDelivered)
	if summary.PushDelivered == 0 && summary.PushRetrying > 0 {
		deliveryScore = 55
	}
	summary.ActivationScore = averageScore(
		operations.NotificationScore,
		routeScore,
		deliveryScore,
		scoreFromBool(summary.PushPreferenceEnabled),
		scoreFromBool(summary.PushEscalationEnabled),
		scoreFromBool(summary.PushSimulationReady),
		scoreFromBool(remediation.Summary.PushProtected),
	)
	return summary
}

func pushActivationRoutes(drillRoutes []model.TenantDeliveryDrilldownRoute, providerRoutes []model.TenantProviderSimulationRoute, remediationActions []model.TenantDeliveryRemediationAction) []model.TenantPushActivationRoute {
	providerByRoute := make(map[string]model.TenantProviderSimulationRoute)
	for _, route := range providerRoutes {
		if route.Channel == constants.DeliveryChannelPush {
			providerByRoute[route.RouteID] = route
		}
	}
	remediationByRoute := make(map[string]model.TenantDeliveryRemediationAction)
	for _, action := range remediationActions {
		if action.Channel == constants.DeliveryChannelPush {
			remediationByRoute[action.RouteID] = action
		}
	}
	routes := make([]model.TenantPushActivationRoute, 0)
	for _, route := range drillRoutes {
		if route.Channel != constants.DeliveryChannelPush {
			continue
		}
		providerRoute := providerByRoute[route.RouteID]
		remediation := remediationByRoute[route.RouteID]
		routes = append(routes, model.TenantPushActivationRoute{
			RouteID:              route.RouteID,
			Provider:             route.Provider,
			SubscriptionLabel:    route.RecipientLabel,
			Status:               route.RouteStatus,
			ProofState:           route.ProofState,
			LatestDeliveryStatus: route.LatestDeliveryStatus,
			LatestDeliveryAt:     route.LatestDeliveryAt,
			Attempts:             route.Attempts,
			NextRetryAt:          remediation.NextRetryAt,
			SimulationStatus:     firstNonEmpty(providerRoute.SimulationStatus, route.RouteStatus),
			SLATarget:            firstNonEmpty(providerRoute.SLATarget, route.SLA, "push proof within 60 seconds"),
			EndpointStorage:      "subscription label only; raw push endpoint is not stored in TraceDeck evidence",
			Evidence:             firstNonEmpty(providerRoute.Evidence, route.Evidence, route.RehearsalResult),
			NextAction:           firstNonEmpty(remediation.Plan, providerRoute.NextAction, route.NextAction),
			PaidTier:             firstNonEmpty(providerRoute.PaidTier, constants.PlanFamilyPro),
		})
	}
	return routes
}

func pushActivationStatus(summary model.TenantPushActivationSummary) string {
	switch {
	case summary.PushFailed > 0:
		return constants.StatusAttention
	case summary.PushRetrying > 0 || summary.PushRoutesNeedingProof > 0 || !summary.PushPreferenceEnabled || !summary.PushEscalationEnabled:
		return constants.StatusWatch
	case summary.ActivationScore >= 80:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func pushActivationHeadline(summary model.TenantPushActivationSummary) string {
	switch {
	case summary.PushFailed > 0:
		return fmt.Sprintf("%d push delivery failure%s need provider-safe remediation", summary.PushFailed, pluralSuffix(summary.PushFailed))
	case summary.PushRetrying > 0:
		return fmt.Sprintf("%d push delivery%s are retrying with setup evidence visible", summary.PushRetrying, pluralSuffix(summary.PushRetrying))
	case summary.PushRoutesNeedingProof > 0:
		return fmt.Sprintf("%d push route%s need proof before the buyer demo", summary.PushRoutesNeedingProof, pluralSuffix(summary.PushRoutesNeedingProof))
	case summary.PushDelivered > 0:
		return fmt.Sprintf("%d push delivery proof row%s support immediate anomaly notifications", summary.PushDelivered, pluralSuffix(summary.PushDelivered))
	default:
		return "Push activation is configured and ready for rehearsal"
	}
}

func pushActivationDetail(summary model.TenantPushActivationSummary) string {
	return fmt.Sprintf("%d delivered, %d retrying, %d failed, %d/%d routes ready, %d push rules, %d alerts with push, preference %s, escalation %s, simulation %s.",
		summary.PushDelivered,
		summary.PushRetrying,
		summary.PushFailed,
		summary.PushRoutesReady,
		summary.PushRoutesTotal,
		summary.AlertRulesUsingPush,
		summary.AlertsWithPush,
		boolReady(summary.PushPreferenceEnabled),
		boolReady(summary.PushEscalationEnabled),
		boolReady(summary.PushSimulationReady),
	)
}

func pushActivationNextStep(summary model.TenantPushActivationSummary) string {
	switch {
	case summary.PushRetrying > 0:
		return "Run provider-safe push retry rehearsal and keep dashboard fallback visible."
	case summary.PushFailed > 0:
		return "Plan push remediation before promising immediate anomaly notification."
	case summary.PushRoutesNeedingProof > 0:
		return "Run push dry-run simulation and attach the metadata-only proof to the customer packet."
	case !summary.PushPreferenceEnabled:
		return "Enable push in the notification preference center for high and critical anomaly rules."
	case !summary.PushEscalationEnabled:
		return "Add push to escalation channels so urgent anomalies do not wait for weekly review."
	default:
		return "Use push activation proof in the Family Pro and school onboarding demo."
	}
}

func pushActivationScenarios(summary model.TenantPushActivationSummary) []model.TenantPushActivationScenario {
	status := constants.StatusHealthy
	if summary.PushRetrying > 0 || summary.PushRoutesNeedingProof > 0 {
		status = constants.StatusWatch
	}
	return []model.TenantPushActivationScenario{
		{
			ID:         "non-study-youtube-push",
			Title:      "Non-study YouTube push",
			Trigger:    "non-study YouTube crosses the configured threshold",
			Channels:   []string{constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Status:     status,
			BuyerValue: "Shows parents immediate awareness without alert bodies, raw URLs, or video titles.",
			StudySafe:  true,
			PaidTier:   constants.PlanFamilyPro,
		},
		{
			ID:         "media-playback-push",
			Title:      "Media playback push",
			Trigger:    "VLC or media playback appears during protected study hours",
			Channels:   []string{constants.DeliveryChannelPush, constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Status:     status,
			BuyerValue: "Turns movie-player anomalies into visible owner action without collecting screenshots.",
			StudySafe:  true,
			PaidTier:   constants.PlanFamilyPro,
		},
		{
			ID:         "tamper-fallback-push",
			Title:      "Tamper fallback push",
			Trigger:    "agent, archive, route, or sync trust signal needs attention",
			Channels:   []string{constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Status:     status,
			BuyerValue: "Keeps trust and archive posture visible even when a provider route retries.",
			StudySafe:  true,
			PaidTier:   constants.PlanSchool,
		},
	}
}

func pushActivationActions(summary model.TenantPushActivationSummary, routes []model.TenantPushActivationRoute, generatedAt time.Time) []model.TenantPushActivationAction {
	actions := make([]model.TenantPushActivationAction, 0, 6)
	for _, route := range routes {
		if len(actions) >= 3 {
			break
		}
		if route.ProofState == constants.DeliveryProofStateCustomer && route.LatestDeliveryStatus == constants.DeliveryStatusDelivered {
			continue
		}
		actions = append(actions, model.TenantPushActivationAction{
			Title:      "Close push route proof",
			Detail:     firstNonEmpty(route.NextAction, "Run provider-safe push simulation and verify retry timing."),
			Owner:      firstNonEmpty(route.SubscriptionLabel, constants.RoleParent),
			Status:     route.Status,
			Severity:   constants.SeverityMedium,
			SLA:        firstNonEmpty(route.SLATarget, "push proof within 60 seconds"),
			PaidTier:   firstNonEmpty(route.PaidTier, constants.PlanFamilyPro),
			Source:     "push activation route",
			ObservedAt: generatedAt,
		})
	}
	if !summary.PushPreferenceEnabled {
		actions = append(actions, model.TenantPushActivationAction{
			Title:      "Enable push preferences",
			Detail:     "High and critical anomaly rules should include push plus dashboard fallback.",
			Owner:      constants.RoleBusinessManager,
			Status:     constants.StatusWatch,
			Severity:   constants.SeverityMedium,
			SLA:        "before paid demo",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "notification preference center",
			ObservedAt: generatedAt,
		})
	}
	if !summary.PushEscalationEnabled {
		actions = append(actions, model.TenantPushActivationAction{
			Title:      "Add push escalation",
			Detail:     "Escalation should include push for urgent anomaly awareness with dashboard fallback.",
			Owner:      constants.RoleParent,
			Status:     constants.StatusWatch,
			Severity:   constants.SeverityMedium,
			SLA:        "before onboarding",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "notification escalation policy",
			ObservedAt: generatedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantPushActivationAction{
			Title:      "Use push activation proof",
			Detail:     "Show route, preference, escalation, simulation, and retry metadata in the next customer review.",
			Owner:      constants.RoleBusinessManager,
			Status:     constants.StatusHealthy,
			Severity:   constants.SeverityInfo,
			SLA:        "weekly review",
			PaidTier:   firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
			Source:     "push activation center",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func pushPreferenceRules(rules []model.NotificationPreferenceRule) int {
	count := 0
	for _, rule := range rules {
		if stringSliceHas(rule.Channels, constants.DeliveryChannelPush) {
			count++
		}
	}
	return count
}

func stringSliceHas(values []string, expected string) bool {
	for _, value := range values {
		if strings.EqualFold(strings.TrimSpace(value), expected) {
			return true
		}
	}
	return false
}

func buildTenantPortfolioCenter(
	operations model.TenantOperationsSummary,
	business model.TenantBusinessDashboard,
	syncHealth model.TenantSyncHealth,
	inbox model.TenantAlertInbox,
	timeline model.TenantDeliveryTimeline,
	packageBilling model.TenantPackageBillingReadiness,
	hosts []model.TenantPortfolioHost,
	generatedAt time.Time,
) model.TenantPortfolioCenter {
	summary := tenantPortfolioSummary(operations, business, syncHealth, inbox, timeline, packageBilling, hosts)
	summary.Status = tenantPortfolioStatus(summary)
	summary.Headline = tenantPortfolioHeadline(summary)
	summary.Detail = tenantPortfolioDetail(summary)
	summary.OwnerNextStep = tenantPortfolioNextStep(summary)

	return model.TenantPortfolioCenter{
		TenantID:           operations.TenantID,
		TenantName:         operations.TenantName,
		PlanID:             operations.PlanID,
		PlanName:           operations.PlanName,
		Audience:           firstNonEmpty(packageBilling.Audience, business.Audience, constants.PlanFamilyPro),
		Summary:            summary,
		Hosts:              hosts,
		Segments:           tenantPortfolioSegments(summary, syncHealth),
		AlertNotifications: tenantPortfolioAlertNotifications(business.Alerts, hosts, generatedAt),
		DeliveryProof:      tenantPortfolioDeliveryProof(summary, business.Channels, business.Summary, generatedAt),
		Actions:            tenantPortfolioActions(summary, hosts, generatedAt),
		PrivacyBoundary:    constants.PortfolioCenterPrivacyNote,
		GeneratedAt:        generatedAt,
	}
}

func tenantPortfolioSummary(
	operations model.TenantOperationsSummary,
	business model.TenantBusinessDashboard,
	syncHealth model.TenantSyncHealth,
	inbox model.TenantAlertInbox,
	timeline model.TenantDeliveryTimeline,
	packageBilling model.TenantPackageBillingReadiness,
	hosts []model.TenantPortfolioHost,
) model.TenantPortfolioSummary {
	hostsAttention := operations.HostsAttention
	for _, host := range hosts {
		if host.Status == constants.StatusAttention {
			hostsAttention++
		}
	}
	if hostsAttention > len(hosts) {
		hostsAttention = len(hosts)
	}

	summary := model.TenantPortfolioSummary{
		NotificationScore:      firstPositive(operations.NotificationScore, business.Summary.NotificationScore, timeline.Summary.NotificationScore),
		TrustScore:             business.Summary.TrustScore,
		RiskScore:              operations.RiskScore,
		HostsTotal:             firstPositive(operations.HostsTotal, len(hosts), syncHealth.HostsTotal),
		HostsAttention:         hostsAttention,
		HostsReporting:         syncHealth.HostsReporting,
		HostsPending:           syncHealth.HostsPending,
		OpenAlerts:             inbox.Summary.Open,
		HighPriorityAlerts:     inbox.Summary.HighOrCritical,
		MailDelivered:          operations.EmailDelivered,
		PushDelivered:          operations.PushDelivered,
		DashboardDelivered:     operations.DashboardDelivered,
		ArchiveBacklog:         operations.ArchiveBacklog,
		StoredTelemetryEvents:  syncHealth.StoredEvents,
		RoutesNeedingProof:     firstPositive(business.Summary.RoutesNeedingProof, timeline.Summary.RouteProofGaps),
		RecommendedPaidPackage: firstNonEmpty(packageBilling.Summary.RecommendedPackage, business.Summary.RecommendedPackage, operations.PlanName, constants.PlanFamilyPro),
	}
	for _, item := range timeline.Items {
		if item.Channel == constants.DeliveryChannelPush && item.Status == constants.DeliveryStatusRetrying {
			summary.PushRetrying++
		}
	}
	summary.PortfolioScore = averageScore(
		operations.MonetizationReadiness,
		business.Summary.ProductScore,
		summary.NotificationScore,
		summary.TrustScore,
		ratioScore(summary.HostsReporting, summary.HostsTotal),
		100-summary.RiskScore,
	)
	return summary
}

func tenantPortfolioHost(overview model.HostOverview, timelineItems []model.TenantDeliveryTimelineItem) model.TenantPortfolioHost {
	emailStatus, emailAt := hostDeliveryStatus(overview.AlertDeliveries, constants.DeliveryChannelEmail)
	pushStatus, pushAt := hostDeliveryStatus(overview.AlertDeliveries, constants.DeliveryChannelPush)
	dashboardStatus, dashboardAt := hostDeliveryStatus(overview.AlertDeliveries, constants.DeliveryChannelDashboard)
	latestDeliveryAt := latestTime(emailAt, pushAt, dashboardAt)

	status := constants.StatusHealthy
	if overview.RiskScore >= 75 || len(overview.TamperEvents) > 0 {
		status = constants.StatusAttention
	} else if len(overview.Anomalies) > 0 || overview.Summary.ArchiveBacklog > 0 || pushStatus == constants.DeliveryStatusRetrying || overview.Health.Score < 80 {
		status = constants.StatusWatch
	}
	if hostHasFailedDelivery(overview.AlertDeliveries) {
		status = constants.StatusAttention
	}

	return model.TenantPortfolioHost{
		DeviceID:             overview.Device.DeviceID,
		HostName:             overview.Device.HostName,
		Profile:              overview.Device.Profile,
		OSName:               overview.Device.OSName,
		Status:               status,
		RiskLevel:            overview.RiskLevel,
		RiskScore:            overview.RiskScore,
		HealthScore:          overview.Health.Score,
		ComplianceScore:      overview.Summary.ComplianceScore,
		PolicyViolations:     len(overview.PolicyViolations),
		Anomalies:            len(overview.Anomalies),
		TamperSignals:        len(overview.TamperEvents),
		ArchiveBacklog:       overview.Summary.ArchiveBacklog,
		DataCompletenessPct:  overview.Summary.DataCompletenessPct,
		EmailStatus:          firstNonEmpty(emailStatus, constants.StatusPending),
		PushStatus:           firstNonEmpty(pushStatus, constants.StatusPending),
		DashboardStatus:      firstNonEmpty(dashboardStatus, constants.StatusPending),
		LastSeenAt:           overview.Device.LastSeenAt,
		LastDeliveryAt:       latestDeliveryAt,
		NextAction:           tenantPortfolioHostNextAction(overview, pushStatus),
		PaidTier:             constants.PlanFamilyPro,
		MetadataProofSummary: hostMetadataProofSummary(overview, timelineItems),
	}
}

func tenantPortfolioStatus(summary model.TenantPortfolioSummary) string {
	switch {
	case summary.HighPriorityAlerts > 0 || summary.HostsAttention > 0:
		return constants.StatusAttention
	case summary.PushRetrying > 0 || summary.RoutesNeedingProof > 0 || summary.HostsPending > 0 || summary.ArchiveBacklog > 0:
		return constants.StatusWatch
	case summary.PortfolioScore >= 80:
		return constants.StatusHealthy
	default:
		return constants.StatusPending
	}
}

func tenantPortfolioHeadline(summary model.TenantPortfolioSummary) string {
	switch {
	case summary.HighPriorityAlerts > 0:
		return fmt.Sprintf("%d high-priority alert%s need owner review across the portfolio", summary.HighPriorityAlerts, pluralSuffix(summary.HighPriorityAlerts))
	case summary.HostsAttention > 0:
		return fmt.Sprintf("%d host%s need attention before the next customer review", summary.HostsAttention, pluralSuffix(summary.HostsAttention))
	case summary.PushRetrying > 0:
		return fmt.Sprintf("%d push notification%s are retrying with mail/dashboard fallback visible", summary.PushRetrying, pluralSuffix(summary.PushRetrying))
	case summary.HostsTotal > 0:
		return fmt.Sprintf("%d host%s are visible in the portfolio command view", summary.HostsTotal, pluralSuffix(summary.HostsTotal))
	default:
		return "Portfolio center is waiting for enrolled hosts"
	}
}

func tenantPortfolioDetail(summary model.TenantPortfolioSummary) string {
	return fmt.Sprintf("%d/%d hosts reporting, %d open alerts, %d mail delivered, %d push delivered, %d push retrying, %d dashboard delivered, %d archive batches pending, %d route proof gaps.",
		summary.HostsReporting,
		summary.HostsTotal,
		summary.OpenAlerts,
		summary.MailDelivered,
		summary.PushDelivered,
		summary.PushRetrying,
		summary.DashboardDelivered,
		summary.ArchiveBacklog,
		summary.RoutesNeedingProof,
	)
}

func tenantPortfolioNextStep(summary model.TenantPortfolioSummary) string {
	switch {
	case summary.HighPriorityAlerts > 0:
		return "Review high-priority host rows and confirm owner notification proof."
	case summary.PushRetrying > 0:
		return "Close push retry proof while keeping mail and dashboard fallback visible."
	case summary.RoutesNeedingProof > 0:
		return "Run provider-safe delivery rehearsal for routes that still need proof."
	case summary.HostsPending > 0:
		return "Confirm offline replay and startup status for hosts that have not reported."
	case summary.ArchiveBacklog > 0:
		return "Flush local archive backlog before the next weekly review."
	default:
		return "Use the portfolio center as the parent, school, or business admin opening view."
	}
}

func tenantPortfolioSegments(summary model.TenantPortfolioSummary, syncHealth model.TenantSyncHealth) []model.TenantPortfolioSegment {
	return []model.TenantPortfolioSegment{
		{
			ID:       "fleet-coverage",
			Label:    "Fleet coverage",
			Value:    fmt.Sprintf("%d/%d hosts", summary.HostsReporting, summary.HostsTotal),
			Detail:   syncHealth.OfflineReplaySummary,
			Status:   gapStatus(summary.HostsPending),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "risk-queue",
			Label:    "Risk queue",
			Value:    fmt.Sprintf("%d open", summary.OpenAlerts),
			Detail:   fmt.Sprintf("%d high priority and %d hosts needing attention", summary.HighPriorityAlerts, summary.HostsAttention),
			Status:   tenantPortfolioStatus(summary),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "notification-proof",
			Label:    "Notification proof",
			Value:    fmt.Sprintf("%d%%", summary.NotificationScore),
			Detail:   fmt.Sprintf("%d mail, %d push, %d dashboard delivered; %d push retrying", summary.MailDelivered, summary.PushDelivered, summary.DashboardDelivered, summary.PushRetrying),
			Status:   gapStatus(summary.PushRetrying + summary.RoutesNeedingProof),
			PaidTier: constants.PlanFamilyPro,
		},
		{
			ID:       "archive-sync",
			Label:    "Archive and sync",
			Value:    fmt.Sprintf("%d events", summary.StoredTelemetryEvents),
			Detail:   fmt.Sprintf("%d archive batches pending", summary.ArchiveBacklog),
			Status:   gapStatus(summary.ArchiveBacklog),
			PaidTier: constants.PlanSchool,
		},
		{
			ID:       "package-readiness",
			Label:    "Package readiness",
			Value:    fmt.Sprintf("%d%%", summary.PortfolioScore),
			Detail:   fmt.Sprintf("%s supports portfolio packaging", summary.RecommendedPaidPackage),
			Status:   tenantPortfolioStatus(summary),
			PaidTier: firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
		},
	}
}

func tenantPortfolioAlertNotifications(alerts []model.TenantBusinessDashboardAlert, hosts []model.TenantPortfolioHost, generatedAt time.Time) []model.TenantPortfolioAlertNotification {
	items := make([]model.TenantPortfolioAlertNotification, 0, len(alerts)+1)
	for _, alert := range alerts {
		items = append(items, model.TenantPortfolioAlertNotification{
			Title:           alert.Title,
			Detail:          alert.Detail,
			Severity:        alert.Severity,
			Status:          alert.Status,
			HostName:        alert.HostName,
			Category:        alert.Category,
			EmailStatus:     firstNonEmpty(alert.EmailStatus, constants.StatusPending),
			PushStatus:      firstNonEmpty(alert.PushStatus, constants.StatusPending),
			DashboardStatus: firstNonEmpty(alert.DashboardStatus, constants.StatusPending),
			NextAction:      firstNonEmpty(alert.NextAction, "Review this anomaly and confirm owner notification proof."),
			PaidTier:        firstNonEmpty(alert.PaidTier, constants.PlanFamilyPro),
			ObservedAt:      alert.ObservedAt,
		})
	}
	if len(items) > 0 || len(hosts) == 0 {
		return items
	}
	attentionHost := hosts[0]
	for _, host := range hosts {
		if statusRankValue(host.Status) > statusRankValue(attentionHost.Status) {
			attentionHost = host
		}
	}
	return append(items, model.TenantPortfolioAlertNotification{
		Title:           "Portfolio anomaly notifications ready",
		Detail:          fmt.Sprintf("%s has host-level policy, anomaly, tamper, mail, push, and dashboard proof available.", attentionHost.HostName),
		Severity:        constants.SeverityInfo,
		Status:          constants.StatusHealthy,
		HostName:        attentionHost.HostName,
		Category:        constants.PortfolioProofAlertInbox,
		EmailStatus:     attentionHost.EmailStatus,
		PushStatus:      attentionHost.PushStatus,
		DashboardStatus: attentionHost.DashboardStatus,
		NextAction:      "Use this notification queue as the paid owner review surface.",
		PaidTier:        attentionHost.PaidTier,
		ObservedAt:      generatedAt,
	})
}

func tenantPortfolioDeliveryProof(summary model.TenantPortfolioSummary, channels []model.TenantBusinessDashboardChannel, businessSummary model.TenantBusinessDashboardSummary, generatedAt time.Time) []model.TenantPortfolioDeliveryProof {
	proofs := make([]model.TenantPortfolioDeliveryProof, 0, 6)
	channelByName := make(map[string]model.TenantBusinessDashboardChannel, len(channels))
	for _, channel := range channels {
		channelByName[channel.Channel] = channel
	}
	proofs = append(proofs,
		tenantPortfolioChannelProof(
			constants.PortfolioProofMailDelivery,
			fmt.Sprintf("%d delivered", summary.MailDelivered),
			"Critical anomaly mail proof for owners, parents, school admins, and business managers.",
			constants.DeliveryChannelEmail,
			channelByName[constants.DeliveryChannelEmail],
			summary.RoutesNeedingProof,
			generatedAt,
		),
		tenantPortfolioChannelProof(
			constants.PortfolioProofPushNotifications,
			fmt.Sprintf("%d delivered, %d retrying", summary.PushDelivered, summary.PushRetrying),
			"Push notification reach with mail and dashboard fallback visible when provider delivery retries.",
			constants.DeliveryChannelPush,
			channelByName[constants.DeliveryChannelPush],
			summary.PushRetrying,
			generatedAt,
		),
		tenantPortfolioChannelProof(
			constants.PortfolioProofDashboardFallback,
			fmt.Sprintf("%d delivered", summary.DashboardDelivered),
			"Dashboard fallback keeps anomaly evidence visible even when external providers need proof.",
			constants.DeliveryChannelDashboard,
			channelByName[constants.DeliveryChannelDashboard],
			0,
			generatedAt,
		),
	)
	proofs = append(proofs,
		model.TenantPortfolioDeliveryProof{
			Label:      constants.PortfolioProofAlertInbox,
			Value:      fmt.Sprintf("%d open, %d high", summary.OpenAlerts, summary.HighPriorityAlerts),
			Detail:     "Owner-facing anomaly, policy, and tamper queue packaged for paid reviews.",
			Channel:    constants.DeliveryChannelDashboard,
			Status:     tenantPortfolioStatus(summary),
			ProofState: gapStatus(summary.HighPriorityAlerts),
			PaidTier:   constants.PlanFamilyPro,
			NextAction: tenantPortfolioNextStep(summary),
			ObservedAt: &generatedAt,
		},
		model.TenantPortfolioDeliveryProof{
			Label:      constants.PortfolioProofWeeklyArchive,
			Value:      boolReady(businessSummary.WeeklyReportReady),
			Detail:     fmt.Sprintf("%d archive batches pending and %d stored telemetry events for weekly reports.", summary.ArchiveBacklog, summary.StoredTelemetryEvents),
			Channel:    constants.DeliveryChannelEmail,
			Status:     gapStatus(summary.ArchiveBacklog),
			ProofState: boolProofState(businessSummary.WeeklyReportReady),
			PaidTier:   constants.PlanFamilyPro,
			NextAction: "Keep weekly PDF/email report and archive retention proof ready for renewals.",
			ObservedAt: &generatedAt,
		},
		model.TenantPortfolioDeliveryProof{
			Label:      constants.PortfolioProofHostCoverage,
			Value:      fmt.Sprintf("%d/%d reporting", summary.HostsReporting, summary.HostsTotal),
			Detail:     fmt.Sprintf("%d hosts need attention and %d hosts are pending replay.", summary.HostsAttention, summary.HostsPending),
			Channel:    constants.DeliveryChannelDashboard,
			Status:     gapStatus(summary.HostsPending + summary.HostsAttention),
			ProofState: gapStatus(summary.HostsPending),
			PaidTier:   constants.PlanSchool,
			NextAction: "Use host coverage proof for family, school, and business expansion.",
			ObservedAt: &generatedAt,
		},
	)
	return proofs
}

func tenantPortfolioChannelProof(label string, value string, detail string, channelName string, channel model.TenantBusinessDashboardChannel, gapCount int, generatedAt time.Time) model.TenantPortfolioDeliveryProof {
	observedAt := channel.LastDeliveryAt
	if observedAt == nil {
		observedAt = &generatedAt
	}
	return model.TenantPortfolioDeliveryProof{
		Label:      label,
		Value:      value,
		Detail:     detail,
		Channel:    channelName,
		Status:     firstNonEmpty(channel.Status, gapStatus(gapCount)),
		ProofState: firstNonEmpty(channel.ProofState, gapStatus(gapCount)),
		PaidTier:   firstNonEmpty(channel.PaidTier, constants.PlanFamilyPro),
		NextAction: firstNonEmpty(channel.NextAction, "Keep provider-safe delivery proof current for paid demos."),
		ObservedAt: observedAt,
	}
}

func boolProofState(value bool) string {
	if value {
		return constants.StatusHealthy
	}
	return constants.StatusPending
}

func buildAccountPortfolioIndex(centers []model.TenantPortfolioCenter, generatedAt time.Time) model.AccountPortfolioIndex {
	tenants := accountPortfolioTenantRows(centers)
	summary := accountPortfolioSummary(centers)
	summary.Status = accountPortfolioStatus(summary)
	summary.Headline = accountPortfolioHeadline(summary)
	summary.Detail = accountPortfolioDetail(summary)
	summary.OwnerNextStep = accountPortfolioNextStep(summary)
	return model.AccountPortfolioIndex{
		Summary:         summary,
		Tenants:         tenants,
		Proof:           accountPortfolioProof(summary),
		Actions:         accountPortfolioActions(summary, tenants),
		PrivacyBoundary: constants.AccountPortfolioPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func accountPortfolioTenantRows(centers []model.TenantPortfolioCenter) []model.AccountPortfolioTenant {
	rows := make([]model.AccountPortfolioTenant, 0, len(centers))
	for _, center := range centers {
		summary := center.Summary
		rows = append(rows, model.AccountPortfolioTenant{
			TenantID:               center.TenantID,
			TenantName:             center.TenantName,
			PlanID:                 center.PlanID,
			PlanName:               center.PlanName,
			Audience:               center.Audience,
			Status:                 summary.Status,
			PortfolioScore:         summary.PortfolioScore,
			NotificationScore:      summary.NotificationScore,
			TrustScore:             summary.TrustScore,
			RiskScore:              summary.RiskScore,
			HostsTotal:             summary.HostsTotal,
			HostsAttention:         summary.HostsAttention,
			OpenAlerts:             summary.OpenAlerts,
			HighPriorityAlerts:     summary.HighPriorityAlerts,
			MailDelivered:          summary.MailDelivered,
			PushDelivered:          summary.PushDelivered,
			DashboardDelivered:     summary.DashboardDelivered,
			ArchiveBacklog:         summary.ArchiveBacklog,
			RoutesNeedingProof:     summary.RoutesNeedingProof,
			RecommendedPaidPackage: summary.RecommendedPaidPackage,
			NextAction:             summary.OwnerNextStep,
			PrivacyBoundary:        center.PrivacyBoundary,
		})
	}
	sort.Slice(rows, func(i, j int) bool {
		leftRank := statusRankValue(rows[i].Status)
		rightRank := statusRankValue(rows[j].Status)
		if leftRank != rightRank {
			return leftRank > rightRank
		}
		if rows[i].PortfolioScore != rows[j].PortfolioScore {
			return rows[i].PortfolioScore < rows[j].PortfolioScore
		}
		return rows[i].TenantName < rows[j].TenantName
	})
	return rows
}

func accountPortfolioSummary(centers []model.TenantPortfolioCenter) model.AccountPortfolioSummary {
	summary := model.AccountPortfolioSummary{
		TenantsTotal:           len(centers),
		RecommendedPaidPackage: constants.PlanFamilyPro,
	}
	var accountScores []int
	var notificationScores []int
	var trustScores []int
	packageRank := 0
	for _, center := range centers {
		item := center.Summary
		if item.Status == constants.StatusAttention || item.Status == constants.StatusWatch {
			summary.TenantsAttention++
		}
		summary.HostsTotal += item.HostsTotal
		summary.HostsAttention += item.HostsAttention
		summary.OpenAlerts += item.OpenAlerts
		summary.HighPriorityAlerts += item.HighPriorityAlerts
		summary.MailDelivered += item.MailDelivered
		summary.PushDelivered += item.PushDelivered
		summary.DashboardDelivered += item.DashboardDelivered
		summary.ArchiveBacklog += item.ArchiveBacklog
		summary.RoutesNeedingProof += item.RoutesNeedingProof
		accountScores = append(accountScores, item.PortfolioScore)
		notificationScores = append(notificationScores, item.NotificationScore)
		trustScores = append(trustScores, item.TrustScore)
		if rank := paidPackageRank(item.RecommendedPaidPackage); rank > packageRank {
			packageRank = rank
			summary.RecommendedPaidPackage = item.RecommendedPaidPackage
		}
	}
	summary.AccountScore = averageScore(accountScores...)
	summary.NotificationScore = averageScore(notificationScores...)
	summary.TrustScore = averageScore(trustScores...)
	return summary
}

func accountPortfolioStatus(summary model.AccountPortfolioSummary) string {
	switch {
	case summary.HighPriorityAlerts > 0 || summary.TenantsAttention > 0:
		return constants.StatusAttention
	case summary.RoutesNeedingProof > 0 || summary.ArchiveBacklog > 0:
		return constants.StatusWatch
	case summary.AccountScore >= 80:
		return constants.StatusHealthy
	case summary.TenantsTotal == 0:
		return constants.StatusPending
	default:
		return constants.StatusWatch
	}
}

func accountPortfolioHeadline(summary model.AccountPortfolioSummary) string {
	switch {
	case summary.TenantsTotal == 0:
		return "Account portfolio is waiting for tenants"
	case summary.HighPriorityAlerts > 0:
		return fmt.Sprintf("%d high-priority alert%s need account-level review", summary.HighPriorityAlerts, pluralSuffix(summary.HighPriorityAlerts))
	case summary.TenantsAttention > 0:
		return fmt.Sprintf("%d tenant%s need owner attention across the account", summary.TenantsAttention, pluralSuffix(summary.TenantsAttention))
	default:
		return fmt.Sprintf("%d tenant%s are ready for account portfolio review", summary.TenantsTotal, pluralSuffix(summary.TenantsTotal))
	}
}

func accountPortfolioDetail(summary model.AccountPortfolioSummary) string {
	return fmt.Sprintf("%d tenants, %d hosts, %d open alerts, %d mail delivered, %d push delivered, %d dashboard delivered, %d route proof gaps, %d archive batches pending.",
		summary.TenantsTotal,
		summary.HostsTotal,
		summary.OpenAlerts,
		summary.MailDelivered,
		summary.PushDelivered,
		summary.DashboardDelivered,
		summary.RoutesNeedingProof,
		summary.ArchiveBacklog,
	)
}

func accountPortfolioNextStep(summary model.AccountPortfolioSummary) string {
	switch {
	case summary.TenantsTotal == 0:
		return "Create or sync tenants before account review."
	case summary.HighPriorityAlerts > 0:
		return "Open tenant rows with high-priority alerts and confirm mail/push proof."
	case summary.RoutesNeedingProof > 0:
		return "Run delivery proof rehearsals for tenants with route gaps."
	case summary.ArchiveBacklog > 0:
		return "Clear archive backlog before renewal or school review."
	case summary.TenantsAttention > 0:
		return "Review attention tenants and assign owner actions."
	default:
		return "Use this account portfolio as the admin opening view."
	}
}

func accountPortfolioProof(summary model.AccountPortfolioSummary) []model.AccountPortfolioProof {
	return []model.AccountPortfolioProof{
		{
			Label:      "Tenant coverage",
			Value:      fmt.Sprintf("%d tenants / %d hosts", summary.TenantsTotal, summary.HostsTotal),
			Detail:     fmt.Sprintf("%d tenants and %d hosts need attention.", summary.TenantsAttention, summary.HostsAttention),
			Status:     gapStatus(summary.TenantsAttention),
			Channel:    constants.DeliveryChannelDashboard,
			PaidTier:   firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
			NextAction: "Use tenant coverage as the account admin opening proof.",
		},
		{
			Label:      "Notification delivery proof",
			Value:      fmt.Sprintf("%d mail / %d push / %d dashboard", summary.MailDelivered, summary.PushDelivered, summary.DashboardDelivered),
			Detail:     fmt.Sprintf("%d route proof gaps remain across visible tenants.", summary.RoutesNeedingProof),
			Status:     gapStatus(summary.RoutesNeedingProof),
			Channel:    constants.DeliveryChannelEmail,
			PaidTier:   constants.PlanFamilyPro,
			NextAction: "Close route proof gaps before a paid customer review.",
		},
		{
			Label:      "Alert queue proof",
			Value:      fmt.Sprintf("%d open / %d high", summary.OpenAlerts, summary.HighPriorityAlerts),
			Detail:     "Account-level anomaly, policy, and tamper pressure without exposing raw content.",
			Status:     gapStatus(summary.HighPriorityAlerts),
			Channel:    constants.DeliveryChannelDashboard,
			PaidTier:   constants.PlanFamilyPro,
			NextAction: accountPortfolioNextStep(summary),
		},
		{
			Label:      "Package readiness",
			Value:      fmt.Sprintf("%d%% account score", summary.AccountScore),
			Detail:     fmt.Sprintf("%s is the recommended account package posture.", firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro)),
			Status:     accountPortfolioStatus(summary),
			Channel:    constants.DeliveryChannelDashboard,
			PaidTier:   firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
			NextAction: "Use score and package proof for expansion and renewal motions.",
		},
		{
			Label:      "Archive and sync proof",
			Value:      fmt.Sprintf("%d pending", summary.ArchiveBacklog),
			Detail:     "Archive backlog is summarized as metadata only for account review.",
			Status:     gapStatus(summary.ArchiveBacklog),
			Channel:    constants.DeliveryChannelDashboard,
			PaidTier:   constants.PlanSchool,
			NextAction: "Clear backlog before compliance export or weekly report review.",
		},
	}
}

func accountPortfolioActions(summary model.AccountPortfolioSummary, tenants []model.AccountPortfolioTenant) []model.AccountPortfolioAction {
	actions := make([]model.AccountPortfolioAction, 0, 6)
	for _, tenant := range tenants {
		if len(actions) >= 3 {
			break
		}
		if tenant.Status == constants.StatusHealthy {
			continue
		}
		actions = append(actions, model.AccountPortfolioAction{
			Title:    "Review tenant account row",
			Detail:   fmt.Sprintf("%s: %s", tenant.TenantName, tenant.NextAction),
			Owner:    tenant.TenantName,
			Status:   tenant.Status,
			Severity: constants.SeverityMedium,
			PaidTier: firstNonEmpty(tenant.RecommendedPaidPackage, constants.PlanFamilyPro),
			Source:   "account portfolio tenant row",
		})
	}
	if summary.RoutesNeedingProof > 0 {
		actions = append(actions, model.AccountPortfolioAction{
			Title:    "Close account route proof gaps",
			Detail:   fmt.Sprintf("%d provider-safe proof gaps remain across visible tenants.", summary.RoutesNeedingProof),
			Owner:    constants.RoleBusinessManager,
			Status:   constants.StatusWatch,
			Severity: constants.SeverityMedium,
			PaidTier: constants.PlanFamilyPro,
			Source:   "account notification proof",
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.AccountPortfolioAction{
			Title:    "Use account portfolio for renewal",
			Detail:   "Open with tenant coverage, notification proof, archive posture, and paid package readiness.",
			Owner:    constants.RoleBusinessManager,
			Status:   constants.StatusHealthy,
			Severity: constants.SeverityInfo,
			PaidTier: firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
			Source:   "account portfolio index",
		})
	}
	return actions
}

func paidPackageRank(packageID string) int {
	switch strings.TrimSpace(packageID) {
	case constants.PlanEnterprise:
		return 5
	case constants.PlanBusiness:
		return 4
	case constants.PlanSchool:
		return 3
	case constants.PlanFamilyPro:
		return 2
	case constants.PlanFree:
		return 1
	default:
		return 0
	}
}

func tenantPortfolioActions(summary model.TenantPortfolioSummary, hosts []model.TenantPortfolioHost, generatedAt time.Time) []model.TenantPortfolioAction {
	actions := make([]model.TenantPortfolioAction, 0, 6)
	for _, host := range hosts {
		if len(actions) >= 3 {
			break
		}
		if host.Status == constants.StatusHealthy {
			continue
		}
		actions = append(actions, model.TenantPortfolioAction{
			Title:      "Review host portfolio row",
			Detail:     fmt.Sprintf("%s: %s", host.HostName, host.NextAction),
			Owner:      host.HostName,
			Status:     host.Status,
			Severity:   constants.SeverityMedium,
			SLA:        "before next review",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "portfolio host row",
			ObservedAt: generatedAt,
		})
	}
	if summary.PushRetrying > 0 {
		actions = append(actions, model.TenantPortfolioAction{
			Title:      "Close push retry proof",
			Detail:     "Run provider-safe push rehearsal and keep mail/dashboard fallback visible in the portfolio.",
			Owner:      constants.RoleParent,
			Status:     constants.StatusWatch,
			Severity:   constants.SeverityMedium,
			SLA:        "same day",
			PaidTier:   constants.PlanFamilyPro,
			Source:     "portfolio notification proof",
			ObservedAt: generatedAt,
		})
	}
	if summary.HostsPending > 0 {
		actions = append(actions, model.TenantPortfolioAction{
			Title:      "Confirm reporting hosts",
			Detail:     "Check startup/autostart and offline replay for hosts that are not reporting.",
			Owner:      constants.RoleBusinessManager,
			Status:     constants.StatusWatch,
			Severity:   constants.SeverityMedium,
			SLA:        "before weekly report",
			PaidTier:   constants.PlanSchool,
			Source:     "portfolio sync coverage",
			ObservedAt: generatedAt,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantPortfolioAction{
			Title:      "Use portfolio proof in onboarding",
			Detail:     "Open with host coverage, notification proof, archive posture, and paid package readiness.",
			Owner:      constants.RoleBusinessManager,
			Status:     constants.StatusHealthy,
			Severity:   constants.SeverityInfo,
			SLA:        "weekly review",
			PaidTier:   firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
			Source:     "portfolio center",
			ObservedAt: generatedAt,
		})
	}
	return actions
}

func hostDeliveryStatus(deliveries []model.AlertDelivery, channel string) (string, time.Time) {
	var latest model.AlertDelivery
	found := false
	for _, delivery := range deliveries {
		if delivery.Channel != channel {
			continue
		}
		if !found || delivery.LastAttemptAt.After(latest.LastAttemptAt) {
			latest = delivery
			found = true
		}
	}
	if !found {
		return "", time.Time{}
	}
	return latest.Status, latest.LastAttemptAt
}

func latestTime(values ...time.Time) time.Time {
	var latest time.Time
	for _, value := range values {
		if value.After(latest) {
			latest = value
		}
	}
	return latest
}

func hostHasFailedDelivery(deliveries []model.AlertDelivery) bool {
	for _, delivery := range deliveries {
		if delivery.Status == constants.DeliveryStatusFailed {
			return true
		}
	}
	return false
}

func tenantPortfolioHostNextAction(overview model.HostOverview, pushStatus string) string {
	switch {
	case len(overview.TamperEvents) > 0:
		return "Review tamper trust and archive posture for this host."
	case overview.RiskScore >= 75 || len(overview.Anomalies) > 0:
		return "Review anomaly queue and notification proof for this host."
	case pushStatus == constants.DeliveryStatusRetrying:
		return "Verify push retry proof while mail and dashboard fallback remain visible."
	case overview.Summary.ArchiveBacklog > 0:
		return "Let the host sync archive backlog when it is online."
	case overview.Health.Score < 80:
		return "Review device health before relying on continuous monitoring."
	default:
		return "Use this host as clean portfolio proof."
	}
}

func hostMetadataProofSummary(overview model.HostOverview, timelineItems []model.TenantDeliveryTimelineItem) string {
	hostDeliveries := 0
	for _, item := range timelineItems {
		if item.DeviceID == overview.Device.DeviceID {
			hostDeliveries++
		}
	}
	return fmt.Sprintf("%d policy, %d anomaly, %d tamper, %d delivery, %d timeline proof rows",
		len(overview.PolicyViolations),
		len(overview.Anomalies),
		len(overview.TamperEvents),
		len(overview.AlertDeliveries),
		hostDeliveries,
	)
}

func firstPositive(values ...int) int {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}

func ratioScore(value int, total int) int {
	if total <= 0 {
		return 0
	}
	score := (value * 100) / total
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

func statusRankValue(status string) int {
	switch status {
	case constants.StatusAttention:
		return 4
	case constants.StatusWatch:
		return 3
	case constants.StatusPending:
		return 2
	case constants.StatusHealthy:
		return 1
	default:
		return 0
	}
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func firstNonZero(values ...int) int {
	for _, value := range values {
		if value != 0 {
			return value
		}
	}
	return 0
}

func roleExperienceParent(operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, business model.TenantBusinessDashboard, preferences model.NotificationPreferenceCenter, timeline model.TenantDeliveryTimeline) model.TenantRoleExperienceRole {
	checks := []bool{
		operations.EmailDelivered > 0,
		operations.DashboardDelivered > 0,
		preferences.Summary.PreferenceScore >= 60,
		timeline.Summary.Total > 0,
		business.Summary.WeeklyReportReady,
	}
	score := readinessFromChecks(checks)
	return model.TenantRoleExperienceRole{
		RoleID:               constants.RoleParent,
		Name:                 "Parent",
		Audience:             "family buyer",
		ViewMode:             "family_pro_guardian",
		Status:               scoreStatus(score),
		ReadinessScore:       score,
		PrimaryGoal:          "See study-safe productivity, anomaly alerts, mail proof, push reach, and weekly reports without private content.",
		VisiblePanels:        []string{"Business Dashboard", "Anomaly Notification Inbox", "Notification Evidence Timeline", "Weekly Report", "Consent Center"},
		NotificationPromise:  monetization.NotificationPromise.Summary,
		ArchiveReportPromise: fmt.Sprintf("%s report readiness with %d archive batches waiting", boolReady(business.Summary.WeeklyReportReady), operations.ArchiveBacklog),
		ConsentControls:      "Visible monitoring disclosure, recipients, quiet hours, data export, and delete request metadata.",
		PaidTier:             constants.PlanFamilyPro,
		NextAction:           firstNonEmpty(parentNextAction(operations, preferences, timeline), "Use the parent view for the Family Pro onboarding demo."),
		Metrics: []model.TenantRoleExperienceMetric{
			roleMetric("Mail proof", fmt.Sprintf("%d delivered", operations.EmailDelivered), monetization.NotificationPromise.Email, deliveryValueStatus(operations.EmailDelivered)),
			roleMetric("Push reach", fmt.Sprintf("%d delivered", operations.PushDelivered), monetization.NotificationPromise.Push, deliveryValueStatus(operations.PushDelivered)),
			roleMetric("Preference policy", fmt.Sprintf("%d%%", preferences.Summary.PreferenceScore), fmt.Sprintf("%d typed rules", preferences.Summary.RulesTotal), scoreStatus(preferences.Summary.PreferenceScore)),
		},
	}
}

func roleExperienceStudent(_ model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, preferences model.NotificationPreferenceCenter, syncHealth model.TenantSyncHealth) model.TenantRoleExperienceRole {
	checks := []bool{
		preferences.Summary.StudySuppressionRules > 0,
		syncHealth.BackendVisible,
		syncHealth.OfflineReplayReady,
		monetization.TrustScore >= 60,
	}
	score := readinessFromChecks(checks)
	return model.TenantRoleExperienceRole{
		RoleID:               constants.RoleStudent,
		Name:                 "Student",
		Audience:             "transparent self-view",
		ViewMode:             "study_self_view",
		Status:               scoreStatus(score),
		ReadinessScore:       score,
		PrimaryGoal:          "Show what is collected, why study sessions are suppressed, and how productivity is summarized.",
		VisiblePanels:        []string{"Telemetry Privacy Boundary", "Study-Safe Suppression", "Activity Mix", "Consent Center"},
		NotificationPromise:  "Student view receives transparency, not private alert bodies.",
		ArchiveReportPromise: syncHealth.OfflineReplaySummary,
		ConsentControls:      "Monitoring status, collection categories, pause/export/delete metadata, and policy history remain visible.",
		PaidTier:             constants.PlanFamilyPro,
		NextAction:           "Keep student transparency visible before enabling stricter school or family policies.",
		Metrics: []model.TenantRoleExperienceMetric{
			roleMetric("Study-safe rules", fmt.Sprintf("%d", preferences.Summary.StudySuppressionRules), "Learning sessions can be quieted without hiding risk.", countStatus(preferences.Summary.StudySuppressionRules)),
			roleMetric("Sync proof", fmt.Sprintf("%d/%d hosts", syncHealth.HostsReporting, syncHealth.HostsTotal), syncHealth.OfflineReplaySummary, countStatus(syncHealth.HostsReporting)),
			roleMetric("Trust score", fmt.Sprintf("%d%%", monetization.TrustScore), "Consent, audit, export, and delete workflows support legitimacy.", scoreStatus(monetization.TrustScore)),
		},
	}
}

func roleExperienceSchoolAdmin(operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, business model.TenantBusinessDashboard, syncHealth model.TenantSyncHealth) model.TenantRoleExperienceRole {
	managedRollout := paidCapabilityStatus(monetization.PaidCapabilities, "Managed rollout")
	checks := []bool{
		operations.HostsTotal > 0,
		managedRollout == constants.StatusHealthy,
		syncHealth.HostsReporting > 0,
		business.Summary.RoutesNeedingProof <= 1,
	}
	score := readinessFromChecks(checks)
	return model.TenantRoleExperienceRole{
		RoleID:               constants.RoleSchoolAdmin,
		Name:                 "School Admin",
		Audience:             "school or coaching center",
		ViewMode:             "managed_cohort_admin",
		Status:               scoreStatus(score),
		ReadinessScore:       score,
		PrimaryGoal:          "Manage cohorts, policy rollout, notification proof, archive retention, and audit history across devices.",
		VisiblePanels:        []string{"Device Groups", "Policy Assignments", "Notification Evidence Timeline", "Offline Replay Health", "Audit Center"},
		NotificationPromise:  fmt.Sprintf("%d/%d notification routes delivered", operations.DeliveryDelivered, operations.DeliveryTotal),
		ArchiveReportPromise: fmt.Sprintf("%d reporting hosts and %d stored metadata events", syncHealth.HostsReporting, syncHealth.StoredEvents),
		ConsentControls:      "Admin view stays metadata-only with role-scoped audit and data rights evidence.",
		PaidTier:             constants.PlanSchool,
		NextAction:           schoolNextAction(managedRollout, business),
		Metrics: []model.TenantRoleExperienceMetric{
			roleMetric("Fleet coverage", fmt.Sprintf("%d hosts", operations.HostsTotal), fmt.Sprintf("%d hosts need attention", operations.HostsAttention), countStatus(operations.HostsTotal)),
			roleMetric("Managed rollout", managedRollout, "Device groups and policy assignments prove school packaging.", managedRollout),
			roleMetric("Route proof", fmt.Sprintf("%d gaps", business.Summary.RoutesNeedingProof), "Notification evidence should be ready before rollout.", gapStatus(business.Summary.RoutesNeedingProof)),
		},
	}
}

func roleExperienceBusinessManager(operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, business model.TenantBusinessDashboard, timeline model.TenantDeliveryTimeline) model.TenantRoleExperienceRole {
	compliance := paidCapabilityStatus(monetization.PaidCapabilities, "Compliance export")
	checks := []bool{
		monetization.TrustScore >= 70,
		compliance == constants.StatusHealthy,
		operations.NotificationScore >= 60,
		timeline.Summary.RouteProofGaps <= 2,
	}
	score := readinessFromChecks(checks)
	return model.TenantRoleExperienceRole{
		RoleID:               constants.RoleBusinessManager,
		Name:                 "Business Manager",
		Audience:             "small business buyer",
		ViewMode:             "business_risk_observability",
		Status:               scoreStatus(score),
		ReadinessScore:       score,
		PrimaryGoal:          "Show productivity, risky software, delivery assurance, archive retention, and compliance export proof.",
		VisiblePanels:        []string{"Revenue Command Center", "Business Dashboard", "Delivery Audit Trail", "Risky Software", "Data Rights"},
		NotificationPromise:  fmt.Sprintf("%d%% notification score with %d route gaps", operations.NotificationScore, timeline.Summary.RouteProofGaps),
		ArchiveReportPromise: fmt.Sprintf("%s package with %d%% trust score", firstNonEmpty(business.Summary.RecommendedPackage, monetization.PlanName), monetization.TrustScore),
		ConsentControls:      "Business view exposes audit and export metadata without raw content or provider secrets.",
		PaidTier:             constants.PlanBusiness,
		NextAction:           businessNextAction(compliance, timeline, monetization),
		Metrics: []model.TenantRoleExperienceMetric{
			roleMetric("Trust score", fmt.Sprintf("%d%%", monetization.TrustScore), "Audit, export, delete, and consent proof.", scoreStatus(monetization.TrustScore)),
			roleMetric("Compliance export", compliance, "Export and audit records support business packaging.", compliance),
			roleMetric("Delivery audit", fmt.Sprintf("%d events", timeline.Summary.Total), fmt.Sprintf("%d route proof gaps", timeline.Summary.RouteProofGaps), gapStatus(timeline.Summary.RouteProofGaps)),
		},
	}
}

func roleOnboardingItems(operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, business model.TenantBusinessDashboard, preferences model.NotificationPreferenceCenter, timeline model.TenantDeliveryTimeline, syncHealth model.TenantSyncHealth) []model.TenantRoleOnboardingItem {
	items := []model.TenantRoleOnboardingItem{
		{
			ID:       "parent-alert-proof",
			Title:    "Parent alert proof",
			Detail:   "Parent view should show delivered mail, push reach, dashboard inbox proof, and weekly report readiness.",
			Owner:    constants.RoleParent,
			Status:   scoreStatus(averageScore(deliveryValueScore(operations.EmailDelivered), deliveryValueScore(operations.PushDelivered), deliveryValueScore(operations.DashboardDelivered))),
			PaidTier: constants.PlanFamilyPro,
			Evidence: fmt.Sprintf("%d mail, %d push, %d dashboard deliveries", operations.EmailDelivered, operations.PushDelivered, operations.DashboardDelivered),
		},
		{
			ID:       "student-transparency",
			Title:    "Student transparency",
			Detail:   "Student self-view should explain collection scope, study-safe suppression, sync proof, and data rights.",
			Owner:    constants.RoleStudent,
			Status:   scoreStatus(averageScore(scoreFromBool(preferences.Summary.StudySuppressionRules > 0), scoreFromBool(syncHealth.BackendVisible), monetization.TrustScore)),
			PaidTier: constants.PlanFamilyPro,
			Evidence: fmt.Sprintf("%d study-safe suppressions and %d reporting hosts", preferences.Summary.StudySuppressionRules, syncHealth.HostsReporting),
		},
		{
			ID:       "school-rollout",
			Title:    "School rollout",
			Detail:   "School admin view should package cohorts, policy assignments, offline replay, and route proof.",
			Owner:    constants.RoleSchoolAdmin,
			Status:   paidCapabilityStatus(monetization.PaidCapabilities, "Managed rollout"),
			PaidTier: constants.PlanSchool,
			Evidence: fmt.Sprintf("%d hosts, %d route proof gaps", operations.HostsTotal, business.Summary.RoutesNeedingProof),
		},
		{
			ID:       "business-compliance",
			Title:    "Business compliance proof",
			Detail:   "Business manager view should show productivity/risk evidence, delivery audit trail, archive trust, and export readiness.",
			Owner:    constants.RoleBusinessManager,
			Status:   paidCapabilityStatus(monetization.PaidCapabilities, "Compliance export"),
			PaidTier: constants.PlanBusiness,
			Evidence: fmt.Sprintf("%d%% trust and %d delivery timeline events", monetization.TrustScore, timeline.Summary.Total),
		},
	}
	return items
}

func roleMetric(label string, value string, detail string, status string) model.TenantRoleExperienceMetric {
	return model.TenantRoleExperienceMetric{
		Label:  label,
		Value:  value,
		Detail: detail,
		Status: status,
	}
}

func readinessFromChecks(checks []bool) int {
	if len(checks) == 0 {
		return 0
	}
	passed := 0
	for _, check := range checks {
		if check {
			passed++
		}
	}
	return (passed * 100) / len(checks)
}

func averageRoleReadiness(roles []model.TenantRoleExperienceRole) int {
	if len(roles) == 0 {
		return 0
	}
	total := 0
	for _, role := range roles {
		total += role.ReadinessScore
	}
	return total / len(roles)
}

func parentNextAction(operations model.TenantOperationsSummary, preferences model.NotificationPreferenceCenter, timeline model.TenantDeliveryTimeline) string {
	switch {
	case operations.PushDelivered == 0:
		return "Finish push route proof before pitching immediate anomaly notifications."
	case preferences.Summary.PreferenceScore < 70:
		return "Tune quiet hours, digest cadence, escalation, and study-safe suppression."
	case timeline.Summary.RouteProofGaps > 0:
		return "Review delivery timeline gaps with the parent before the weekly report."
	default:
		return "Use delivered mail, dashboard, report, and preference proof for Family Pro onboarding."
	}
}

func schoolNextAction(managedRollout string, business model.TenantBusinessDashboard) string {
	if managedRollout != constants.StatusHealthy {
		return "Create or verify device groups and policy assignments before school onboarding."
	}
	if business.Summary.RoutesNeedingProof > 0 {
		return "Resolve route proof gaps before promising school notification SLAs."
	}
	return "Use cohort rollout, offline replay, and notification evidence for the school package."
}

func businessNextAction(compliance string, timeline model.TenantDeliveryTimeline, monetization model.TenantMonetizationSummary) string {
	if compliance != constants.StatusHealthy {
		return "Prepare export and audit evidence before selling business compliance."
	}
	if timeline.Summary.RouteProofGaps > 0 {
		return "Clear delivery audit gaps before a business review."
	}
	if monetization.TrustScore < 80 {
		return "Strengthen consent, audit, export, and delete proof for business buyers."
	}
	return "Use delivery audit, trust score, and archive proof for Business packaging."
}

func paidCapabilityStatus(capabilities []model.TenantPaidCapability, name string) string {
	for _, item := range capabilities {
		if strings.EqualFold(strings.TrimSpace(item.Name), strings.TrimSpace(name)) {
			return item.Status
		}
	}
	return constants.StatusPending
}

func gapStatus(gaps int) string {
	if gaps <= 0 {
		return constants.StatusHealthy
	}
	if gaps <= 2 {
		return constants.StatusWatch
	}
	return constants.StatusAttention
}

func scoreFromBool(value bool) int {
	if value {
		return 100
	}
	return 0
}

func deliveryValueScore(count int) int {
	if count > 0 {
		return 100
	}
	return 0
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

func buildTenantProviderSimulationLab(operations model.TenantOperationsSummary, monetization model.TenantMonetizationSummary, revenue model.TenantNotificationRevenueCockpit, drilldown model.TenantDeliveryDrilldown, remediation model.TenantDeliveryRemediation, generatedAt time.Time) model.TenantProviderSimulationLab {
	routes := providerSimulationRoutes(drilldown.Routes)
	scenarios := providerSimulationScenarios()
	simulated := 0
	providerRisks := 0
	for _, route := range routes {
		if route.ProofState == constants.DeliveryProofStateCustomer || route.ProofState == constants.DeliveryProofStateRehearsed {
			simulated++
		}
		if route.ProofState == constants.DeliveryProofStateNeedsProvider || route.ProofState == constants.DeliveryProofStateMismatch || route.ProofState == constants.DeliveryProofStateDisabled {
			providerRisks++
		}
	}
	readiness := averageScore(revenue.Summary.RevenueReadiness, drilldown.Summary.DeliveryScore, remediation.Summary.RemediationScore, operations.NotificationScore)
	summary := model.TenantProviderSimulationSummary{
		ReadinessScore:         readiness,
		SimulationScore:        drilldown.Summary.DeliveryScore,
		RoutesTotal:            drilldown.Summary.RoutesTotal,
		SimulatedRoutes:        simulated,
		RoutesNeedingProof:     drilldown.Summary.RoutesNeedingProof,
		ProviderRisks:          providerRisks + remediation.Summary.ProblemsOpen,
		EmailReady:             drilldown.Summary.EmailReady,
		PushReady:              drilldown.Summary.PushReady,
		DashboardReady:         drilldown.Summary.DashboardReady,
		SLAReady:               drilldown.Summary.EmailReady && drilldown.Summary.PushReady && drilldown.Summary.DashboardReady,
		RecommendedPaidPackage: firstNonEmpty(revenue.Summary.RecommendedPaidPackage, monetization.PlanName),
	}
	summary.Status = providerSimulationStatus(summary)
	summary.Headline, summary.Detail = providerSimulationNarrative(summary, revenue)
	summary.NextBestAction = providerSimulationNextAction(summary, remediation)

	return model.TenantProviderSimulationLab{
		TenantID:        operations.TenantID,
		TenantName:      operations.TenantName,
		PlanID:          operations.PlanID,
		PlanName:        operations.PlanName,
		Audience:        monetization.Audience,
		Summary:         summary,
		Routes:          routes,
		Scenarios:       scenarios,
		Actions:         providerSimulationActions(summary, routes, generatedAt),
		PrivacyBoundary: constants.ProviderSimulationPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func providerSimulationStatus(summary model.TenantProviderSimulationSummary) string {
	switch {
	case summary.ProviderRisks > 0 || summary.RoutesNeedingProof > 2:
		return constants.StatusAttention
	case summary.ReadinessScore >= 80 && summary.SLAReady:
		return constants.StatusHealthy
	case summary.ReadinessScore >= 55 || summary.SimulatedRoutes > 0:
		return constants.StatusWatch
	default:
		return constants.StatusPending
	}
}

func providerSimulationNarrative(summary model.TenantProviderSimulationSummary, revenue model.TenantNotificationRevenueCockpit) (string, string) {
	switch {
	case summary.SLAReady:
		return "Provider simulation is buyer-ready", fmt.Sprintf("%d/%d routes have metadata-safe simulation or customer proof for %s.", summary.SimulatedRoutes, summary.RoutesTotal, firstNonEmpty(summary.RecommendedPaidPackage, revenue.PlanName))
	case summary.PushReady && summary.EmailReady:
		return "Push and mail simulation proof is ready", "Dashboard delivery still needs simulation proof before a full buyer SLA demo."
	case summary.ProviderRisks > 0:
		return "Provider simulation needs attention", fmt.Sprintf("%d provider risk signals need route proof, retry planning, or owner acknowledgement.", summary.ProviderRisks)
	default:
		return "Provider simulation lab is ready to rehearse", "Run a dry-run simulation for email, push, and dashboard routes before the paid demo."
	}
}

func providerSimulationNextAction(summary model.TenantProviderSimulationSummary, remediation model.TenantDeliveryRemediation) string {
	switch {
	case !summary.PushReady:
		return "Run push dry-run simulation and attach the provider-safe result to the buyer demo."
	case !summary.EmailReady:
		return "Run email dry-run simulation before promising critical alert SLA."
	case !summary.DashboardReady:
		return "Run dashboard inbox simulation so every anomaly has visible fallback proof."
	case remediation.Summary.ProblemsOpen > 0:
		return "Close remediation actions and keep provider simulation proof current."
	default:
		return "Use provider simulation proof in the Family Pro, school, and business upgrade narrative."
	}
}

func providerSimulationRoutes(routes []model.TenantDeliveryDrilldownRoute) []model.TenantProviderSimulationRoute {
	items := make([]model.TenantProviderSimulationRoute, 0, len(routes))
	for _, route := range routes {
		items = append(items, model.TenantProviderSimulationRoute{
			RouteID:              route.RouteID,
			Channel:              route.Channel,
			Provider:             route.Provider,
			RecipientLabel:       route.RecipientLabel,
			SimulationStatus:     providerSimulationRouteStatus(route),
			ProofState:           route.ProofState,
			Scenario:             providerSimulationScenarioForChannel(route.Channel),
			SLATarget:            route.SLA,
			SimulatedLatency:     providerSimulationLatency(route.Channel),
			LatestDeliveryStatus: route.LatestDeliveryStatus,
			LastSimulatedAt:      route.LastVerifiedAt,
			BusinessValue:        providerSimulationBusinessValue(route.Channel),
			Evidence:             firstNonEmpty(route.Evidence, route.RehearsalResult, "Provider-safe simulation proof pending."),
			NextAction:           providerSimulationRouteNextAction(route),
			PaidTier:             notificationCommandChannelTier(route.Channel),
		})
	}
	return items
}

func providerSimulationRouteStatus(route model.TenantDeliveryDrilldownRoute) string {
	switch route.ProofState {
	case constants.DeliveryProofStateCustomer, constants.DeliveryProofStateRehearsed:
		return constants.StatusHealthy
	case constants.DeliveryProofStateNeedsProvider, constants.DeliveryProofStateMismatch:
		return constants.StatusAttention
	default:
		return constants.StatusWatch
	}
}

func providerSimulationScenarioForChannel(channel string) string {
	switch channel {
	case constants.DeliveryChannelEmail:
		return "critical-alert-mail"
	case constants.DeliveryChannelPush:
		return "urgent-anomaly-push"
	case constants.DeliveryChannelDashboard:
		return "dashboard-fallback-proof"
	default:
		return "provider-route-proof"
	}
}

func providerSimulationLatency(channel string) string {
	switch channel {
	case constants.DeliveryChannelPush:
		return "under 60 seconds"
	case constants.DeliveryChannelEmail:
		return "under 5 minutes"
	case constants.DeliveryChannelDashboard:
		return "immediate local dashboard"
	default:
		return "SLA pending"
	}
}

func providerSimulationBusinessValue(channel string) string {
	switch channel {
	case constants.DeliveryChannelEmail:
		return "Proves critical anomaly mail delivery without exposing SMTP passwords or alert bodies."
	case constants.DeliveryChannelPush:
		return "Proves urgent push readiness for non-study video, media, tamper, and risky software alerts."
	case constants.DeliveryChannelDashboard:
		return "Proves every alert still lands in the dashboard when provider routes retry."
	default:
		return "Proves notification route readiness with metadata-only evidence."
	}
}

func providerSimulationRouteNextAction(route model.TenantDeliveryDrilldownRoute) string {
	switch route.ProofState {
	case constants.DeliveryProofStateCustomer:
		return "Use delivered metadata as customer proof and keep simulation current."
	case constants.DeliveryProofStateRehearsed:
		return "Keep dry-run proof current until production provider credentials are configured."
	case constants.DeliveryProofStateNeedsProvider:
		return "Plan retry or provider verification before promising live notification SLA."
	default:
		return firstNonEmpty(route.NextAction, "Run provider-safe simulation before a buyer demo.")
	}
}

func providerSimulationScenarios() []model.TenantProviderSimulationScenario {
	return []model.TenantProviderSimulationScenario{
		{
			ID:         "urgent-anomaly-push",
			Name:       "Urgent anomaly push",
			Trigger:    "non-study YouTube, VLC, media player, tamper, or risky software alert",
			Channels:   []string{constants.DeliveryChannelPush, constants.DeliveryChannelDashboard},
			Severity:   constants.SeverityHigh,
			Outcome:    "Parent or manager sees push readiness and dashboard fallback proof.",
			BuyerValue: "Immediate anomaly notification is packaged as Family Pro value.",
			PaidTier:   constants.PlanFamilyPro,
			StudySafe:  true,
		},
		{
			ID:         "critical-alert-mail",
			Name:       "Critical alert mail",
			Trigger:    "high-severity policy, tamper, archive, or software risk",
			Channels:   []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Severity:   constants.SeverityHigh,
			Outcome:    "Mail route has provider-safe proof without SMTP secrets or message bodies.",
			BuyerValue: "Email evidence supports school and business audit reviews.",
			PaidTier:   constants.PlanSchool,
			StudySafe:  true,
		},
		{
			ID:         "weekly-report-delivery",
			Name:       "Weekly report delivery",
			Trigger:    "weekly report generated with anomaly and study-hour summary",
			Channels:   []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
			Severity:   constants.SeverityMedium,
			Outcome:    "Report mail readiness and dashboard fallback are visible to the owner.",
			BuyerValue: "Report proof increases retention for family, school, and coaching buyers.",
			PaidTier:   constants.PlanFamilyPro,
			StudySafe:  true,
		},
	}
}

func providerSimulationActions(summary model.TenantProviderSimulationSummary, routes []model.TenantProviderSimulationRoute, generatedAt time.Time) []model.TenantProviderSimulationAction {
	_ = generatedAt
	actions := make([]model.TenantProviderSimulationAction, 0, len(routes)+1)
	for _, route := range routes {
		if route.SimulationStatus == constants.StatusHealthy {
			continue
		}
		actions = append(actions, model.TenantProviderSimulationAction{
			Title:           titleWord(route.Channel) + " simulation proof",
			Detail:          route.NextAction,
			Owner:           firstNonEmpty(route.RecipientLabel, constants.RoleBusinessManager),
			Channel:         route.Channel,
			Status:          route.SimulationStatus,
			SLA:             route.SLATarget,
			ConversionLever: route.BusinessValue,
			PaidTier:        route.PaidTier,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.TenantProviderSimulationAction{
			Title:           "Provider simulation buyer proof",
			Detail:          firstNonEmpty(summary.NextBestAction, "Use simulation proof in the next paid demo."),
			Owner:           constants.RoleBusinessManager,
			Channel:         constants.DeliveryChannelDashboard,
			Status:          constants.StatusHealthy,
			SLA:             "all simulation routes ready",
			ConversionLever: "Provider-safe notification proof supports Family Pro, school, and business packaging.",
			PaidTier:        firstNonEmpty(summary.RecommendedPaidPackage, constants.PlanFamilyPro),
		})
	}
	return actions
}

func buildTenantPackageBillingReadiness(
	tenant model.Tenant,
	retention model.RetentionTier,
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	business model.TenantBusinessDashboard,
	roles model.TenantRoleExperience,
	provider model.TenantProviderSimulationLab,
	generatedAt time.Time,
) model.TenantPackageBillingReadiness {
	plan := planByID(tenant.PlanID)
	retentionReady := plan.CloudArchive && retention.S3StandardDays > 0 && retention.S3ArchiveAfterDays >= 365
	archiveReady := retentionReady && operations.ArchiveBacklog == 0
	weeklyReady := plan.WeeklyReports && business.Summary.WeeklyReportReady
	notificationReady := operations.NotificationScore >= 60 && operations.DeliveryDelivered > 0
	providerReady := provider.Summary.SimulatedRoutes > 0 && provider.Summary.ProviderRisks == 0
	billingReady := strings.TrimSpace(tenant.PlanID) != "" &&
		tenant.PlanID != constants.PlanFree &&
		strings.TrimSpace(tenant.RetentionTierID) != "" &&
		monetization.TrustScore >= 60

	featureGates := packageBillingFeatureGates(plan, retention, operations, monetization, business, roles, provider, retentionReady, archiveReady, weeklyReady, notificationReady, providerReady, billingReady)
	readyGates, totalGates := packageFeatureGateCounts(featureGates)
	featureScore := 0
	if totalGates > 0 {
		featureScore = (readyGates * 100) / totalGates
	}
	packageScore := averageScore(monetization.ReadinessScore, business.Summary.ProductScore, provider.Summary.ReadinessScore, featureScore, monetization.TrustScore)
	recommended := firstNonEmpty(business.Summary.RecommendedPackage, provider.Summary.RecommendedPaidPackage, monetization.PlanName, plan.Name)
	summary := model.TenantPackageBillingSummary{
		PackageScore:       packageScore,
		BillingStatus:      packageBillingStatus(billingReady, packageScore, featureScore),
		RevenueStage:       monetization.ConversionStage,
		CurrentPlan:        firstNonEmpty(plan.Name, tenant.PlanID),
		RecommendedPackage: recommended,
		SeatsUsed:          monetization.SeatsUsed,
		SeatsIncluded:      monetization.SeatsIncluded,
		SeatUtilization:    packageSeatUtilization(monetization.SeatsUsed, monetization.SeatsIncluded),
		FeatureGatesReady:  readyGates,
		FeatureGatesTotal:  totalGates,
		UpgradeReady:       packageScore >= 80 && billingReady,
		BillingReady:       billingReady,
		RetentionReady:     retentionReady,
		ArchiveReady:       archiveReady,
		WeeklyReportReady:  weeklyReady,
		NotificationReady:  notificationReady,
		ProviderReady:      providerReady,
		TrustScore:         monetization.TrustScore,
	}
	summary.Status = packageBillingStatus(summary.BillingReady, summary.PackageScore, featureScore)
	summary.Headline, summary.Detail = packageBillingNarrative(summary, retention)
	summary.NextBestAction = packageBillingNextAction(summary)

	return model.TenantPackageBillingReadiness{
		TenantID:        tenant.TenantID,
		TenantName:      tenant.Name,
		PlanID:          tenant.PlanID,
		PlanName:        firstNonEmpty(plan.Name, tenant.PlanID),
		Audience:        firstNonEmpty(plan.Audience, monetization.Audience),
		RetentionTierID: tenant.RetentionTierID,
		RetentionName:   firstNonEmpty(retention.Name, tenant.RetentionTierID),
		Summary:         summary,
		Plans:           packageBillingPlans(plan, recommended, packageScore),
		FeatureGates:    featureGates,
		Milestones:      packageBillingMilestones(summary, retention, provider),
		Actions:         packageBillingActions(summary, monetization.ConversionActions, business.Actions),
		PrivacyBoundary: constants.PackageBillingPrivacyNote,
		GeneratedAt:     generatedAt,
	}
}

func packageBillingStatus(billingReady bool, packageScore int, featureScore int) string {
	switch {
	case !billingReady || featureScore < 60:
		return constants.StatusAttention
	case packageScore >= 80:
		return constants.StatusHealthy
	case packageScore >= 60:
		return constants.StatusWatch
	default:
		return constants.StatusPending
	}
}

func packageBillingNarrative(summary model.TenantPackageBillingSummary, retention model.RetentionTier) (string, string) {
	switch {
	case summary.UpgradeReady:
		return fmt.Sprintf("%s is ready for paid package review", summary.RecommendedPackage),
			fmt.Sprintf("%d/%d feature gates are ready with %s retention proof and %d%% trust.", summary.FeatureGatesReady, summary.FeatureGatesTotal, firstNonEmpty(retention.Name, "configured"), summary.TrustScore)
	case !summary.BillingReady:
		return "Billing readiness needs trust and package proof",
			"Current plan, retention tier, notification proof, and data-rights evidence must be visible before a buyer handoff."
	case !summary.NotificationReady:
		return "Notification proof is the package gap",
			"Mail, push, dashboard, and provider-safe delivery evidence make the paid plan easier to sell."
	case !summary.RetentionReady:
		return "Retention proof is the package gap",
			"S3 lifecycle, local TTL, archive, and compliance export evidence should be visible before selling archive plans."
	default:
		return fmt.Sprintf("%s is package-ready", summary.CurrentPlan),
			"Feature gates, package fit, billing setup metadata, and owner actions are ready for review."
	}
}

func packageBillingNextAction(summary model.TenantPackageBillingSummary) string {
	switch {
	case !summary.BillingReady:
		return "Confirm plan, retention tier, trust center, and billing setup metadata before buyer handoff."
	case !summary.ProviderReady:
		return "Run provider simulation so push, mail, and dashboard proof can support the package pitch."
	case !summary.WeeklyReportReady:
		return "Generate weekly report proof and include it in the package review."
	case !summary.ArchiveReady:
		return "Clear archive backlog or show lifecycle retry proof before selling retention value."
	default:
		return "Use package billing readiness in the Family Pro, school, or business upgrade conversation."
	}
}

func packageBillingPlans(current model.Plan, recommended string, packageScore int) []model.TenantPackageBillingPlan {
	candidates := []model.Plan{planByID(constants.PlanFree), planByID(constants.PlanFamilyPro), planByID(constants.PlanSchool), planByID(constants.PlanBusiness)}
	plans := make([]model.TenantPackageBillingPlan, 0, len(candidates))
	for _, plan := range candidates {
		isCurrent := plan.ID == current.ID
		isRecommended := strings.EqualFold(plan.Name, recommended) || plan.ID == recommended
		status := constants.StatusWatch
		if isCurrent || isRecommended {
			status = constants.StatusHealthy
		}
		if plan.ID == constants.PlanFree && current.ID != constants.PlanFree {
			status = constants.StatusPending
		}
		plans = append(plans, model.TenantPackageBillingPlan{
			PlanID:      plan.ID,
			Name:        plan.Name,
			Audience:    plan.Audience,
			PriceModel:  plan.PriceModel,
			Status:      status,
			Current:     isCurrent,
			Recommended: isRecommended,
			FitScore:    packagePlanFitScore(plan, current, packageScore),
			Features:    append([]string(nil), plan.Features...),
			Value:       businessPackageValue(plan.ID),
			NextAction:  packagePlanNextAction(plan, isCurrent, isRecommended),
		})
	}
	return plans
}

func packagePlanFitScore(plan model.Plan, current model.Plan, packageScore int) int {
	score := packageScore
	if plan.ID == current.ID {
		score += 10
	}
	if plan.ID == constants.PlanFree {
		score -= 25
	}
	if plan.ID == constants.PlanSchool || plan.ID == constants.PlanBusiness {
		score -= 5
	}
	if score > 100 {
		return 100
	}
	if score < 0 {
		return 0
	}
	return score
}

func packagePlanNextAction(plan model.Plan, current bool, recommended bool) string {
	switch {
	case current:
		return "Use this plan as the current package proof."
	case recommended:
		return "Prepare this package as the next upgrade offer."
	case plan.ID == constants.PlanFree:
		return "Keep free as a trial path with local-only limits."
	default:
		return "Keep collecting feature proof before positioning this package."
	}
}

func packageBillingFeatureGates(
	plan model.Plan,
	retention model.RetentionTier,
	operations model.TenantOperationsSummary,
	monetization model.TenantMonetizationSummary,
	business model.TenantBusinessDashboard,
	roles model.TenantRoleExperience,
	provider model.TenantProviderSimulationLab,
	retentionReady bool,
	archiveReady bool,
	weeklyReady bool,
	notificationReady bool,
	providerReady bool,
	billingReady bool,
) []model.TenantPackageBillingFeatureGate {
	return []model.TenantPackageBillingFeatureGate{
		packageFeatureGate("seat-capacity", "Seat Capacity", monetization.SeatsIncluded == 0 || monetization.SeatsUsed <= monetization.SeatsIncluded, fmt.Sprintf("%d/%d seats used", monetization.SeatsUsed, monetization.SeatsIncluded), "Seat limits make packaging clear for family, school, and business buyers.", constants.PlanFamilyPro),
		packageFeatureGate("billing-setup", "Billing Setup", billingReady, "Plan and retention metadata are configured; payment data is not collected.", "Buyer handoff can discuss package readiness without storing payment card data.", constants.PlanFamilyPro),
		packageFeatureGate("archive-retention", "Archive Retention", retentionReady, fmt.Sprintf("%s: %d local days, %d S3 standard days, archive after %d days", firstNonEmpty(retention.Name, retention.ID), retention.LocalTTLDays, retention.S3StandardDays, retention.S3ArchiveAfterDays), "Cloud archive and lifecycle proof monetise retention plans.", constants.PlanFamilyPro),
		packageFeatureGate("archive-health", "Archive Health", archiveReady, fmt.Sprintf("%d archive batches pending", operations.ArchiveBacklog), "A clean archive queue supports renewal and compliance trust.", constants.PlanSchool),
		packageFeatureGate("weekly-report", "Weekly Report", weeklyReady, boolReady(business.Summary.WeeklyReportReady), "Weekly AI report delivery is a Family Pro retention feature.", constants.PlanFamilyPro),
		packageFeatureGate("notification-proof", "Notification Proof", notificationReady, fmt.Sprintf("%d%% notification score with %d delivered events", operations.NotificationScore, operations.DeliveryDelivered), "Mail, push, and dashboard delivery proof makes anomaly alerts monetisable.", constants.PlanFamilyPro),
		packageFeatureGate("provider-simulation", "Provider Simulation", providerReady, fmt.Sprintf("%d/%d routes simulated", provider.Summary.SimulatedRoutes, provider.Summary.RoutesTotal), "Provider-safe proof supports paid demos without storing secrets or payloads.", constants.PlanFamilyPro),
		packageFeatureGate("role-dashboards", "Role Dashboards", plan.RoleBasedDashboard && roles.Summary.RolesReady > 0, fmt.Sprintf("%d/%d roles ready", roles.Summary.RolesReady, roles.Summary.RolesTotal), "Parent, student, school, and manager views unlock higher-tier packaging.", constants.PlanSchool),
		packageFeatureGate("trust-data-rights", "Trust And Data Rights", business.Summary.ConsentVisible && business.Summary.DataRightsReady, fmt.Sprintf("%d%% trust score", monetization.TrustScore), "Consent, export, and delete readiness reduces buyer friction.", constants.PlanBusiness),
		packageFeatureGate("package-evidence", "Package Evidence", len(business.Packages) > 0, fmt.Sprintf("%d package options visible", len(business.Packages)), "Visible package options help convert monitoring into paid observability.", constants.PlanFamilyPro),
	}
}

func packageFeatureGate(id string, label string, enabled bool, evidence string, buyerValue string, paidTier string) model.TenantPackageBillingFeatureGate {
	status := constants.StatusWatch
	if enabled {
		status = constants.StatusHealthy
	}
	return model.TenantPackageBillingFeatureGate{
		ID:         id,
		Label:      label,
		Status:     status,
		Enabled:    enabled,
		Evidence:   evidence,
		BuyerValue: buyerValue,
		PaidTier:   paidTier,
	}
}

func packageFeatureGateCounts(gates []model.TenantPackageBillingFeatureGate) (int, int) {
	ready := 0
	for _, gate := range gates {
		if gate.Enabled {
			ready++
		}
	}
	return ready, len(gates)
}

func packageBillingMilestones(summary model.TenantPackageBillingSummary, retention model.RetentionTier, provider model.TenantProviderSimulationLab) []model.TenantPackageBillingMilestone {
	return []model.TenantPackageBillingMilestone{
		packageMilestone("plan-fit", "Plan fit", summary.BillingReady, summary.CurrentPlan, "Current plan and tenant metadata are configured.", "Keep package evidence current.", "before buyer review", constants.PlanFamilyPro),
		packageMilestone("retention-lifecycle", "Retention lifecycle", summary.RetentionReady, firstNonEmpty(retention.Name, retention.ID), "Local TTL, S3 standard, IA, and archive lifecycle are visible.", "Confirm lifecycle proof before selling archive value.", "before archive pitch", constants.PlanFamilyPro),
		packageMilestone("report-proof", "Report proof", summary.WeeklyReportReady, boolReady(summary.WeeklyReportReady), "Weekly report readiness is visible in dashboard and API.", "Generate report proof before renewal review.", "weekly", constants.PlanFamilyPro),
		packageMilestone("notification-proof", "Notification proof", summary.NotificationReady, fmt.Sprintf("%d/%d feature gates ready", summary.FeatureGatesReady, summary.FeatureGatesTotal), "Alert delivery proof supports mail, push, and dashboard claims.", "Close route gaps before promising SLA.", "same day", constants.PlanFamilyPro),
		packageMilestone("provider-proof", "Provider proof", summary.ProviderReady, fmt.Sprintf("%d simulated routes", provider.Summary.SimulatedRoutes), "Provider simulation is metadata-only and avoids provider secrets.", "Run simulation before paid demo.", "before demo", constants.PlanFamilyPro),
		packageMilestone("trust-review", "Trust review", summary.TrustScore >= 65, fmt.Sprintf("%d%% trust score", summary.TrustScore), "Consent center, audit trail, data exports, and delete requests support trust.", "Review data-rights proof with customer.", "before billing", constants.PlanBusiness),
	}
}

func packageMilestone(id string, title string, ready bool, detail string, evidence string, nextAction string, sla string, paidTier string) model.TenantPackageBillingMilestone {
	status := constants.StatusWatch
	if ready {
		status = constants.StatusHealthy
	}
	return model.TenantPackageBillingMilestone{
		ID:         id,
		Title:      title,
		Detail:     detail,
		Status:     status,
		Owner:      constants.RoleBusinessManager,
		Evidence:   evidence,
		NextAction: nextAction,
		SLA:        sla,
		PaidTier:   paidTier,
	}
}

func packageBillingActions(summary model.TenantPackageBillingSummary, conversion []model.TenantOperationsSignal, businessActions []model.TenantBusinessDashboardAction) []model.TenantPackageBillingAction {
	actions := make([]model.TenantPackageBillingAction, 0, 6)
	for _, signal := range conversion {
		if len(actions) >= 3 {
			break
		}
		actions = append(actions, model.TenantPackageBillingAction{
			Title:           signal.Title,
			Detail:          signal.Detail,
			Owner:           signal.Owner,
			Status:          signal.Status,
			PaidTier:        constants.PlanFamilyPro,
			ConversionLever: firstNonEmpty(summary.RecommendedPackage, summary.CurrentPlan),
			NextAction:      "Use this signal in the package review.",
		})
	}
	for _, action := range businessActions {
		if len(actions) >= 5 {
			break
		}
		actions = append(actions, model.TenantPackageBillingAction{
			Title:           action.Title,
			Detail:          action.Detail,
			Owner:           action.Owner,
			Status:          action.Status,
			PaidTier:        action.PaidTier,
			ConversionLever: firstNonEmpty(action.Source, "business dashboard"),
			NextAction:      "Resolve this before billing or renewal discussion.",
		})
	}
	actions = append(actions, model.TenantPackageBillingAction{
		Title:           "Package billing readiness",
		Detail:          summary.NextBestAction,
		Owner:           constants.RoleBusinessManager,
		Status:          summary.Status,
		PaidTier:        firstNonEmpty(summary.RecommendedPackage, constants.PlanFamilyPro),
		ConversionLever: "Feature gates, billing setup metadata, notification proof, reports, and archive trust are in one buyer view.",
		NextAction:      summary.NextBestAction,
	})
	return actions
}

func packageSeatUtilization(used int, included int) int {
	if included <= 0 {
		if used > 0 {
			return 100
		}
		return 0
	}
	utilization := (used * 100) / included
	if utilization > 100 {
		return 100
	}
	return utilization
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
