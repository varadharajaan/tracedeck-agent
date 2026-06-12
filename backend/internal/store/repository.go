package store

import (
	"context"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
)

type Repository interface {
	CreateTenant(context.Context, model.CreateTenantRequest) (model.Tenant, error)
	ListTenants(context.Context) []model.Tenant
	GetTenant(context.Context, string) (model.Tenant, error)
	ListAuditEvents(context.Context, string) []model.AuditEvent
	CreateAlertRule(context.Context, string, model.CreateAlertRuleRequest) (model.AlertRule, error)
	ListAlertRules(context.Context, string) []model.AlertRule
	CreateNotificationRoute(context.Context, string, model.CreateNotificationRouteRequest) (model.NotificationRoute, error)
	ListNotificationRoutes(context.Context, string) []model.NotificationRoute
	TenantNotificationPreferences(context.Context, string) (model.NotificationPreferenceCenter, error)
	UpdateTenantNotificationPreferences(context.Context, string, model.UpdateNotificationPreferencesRequest) (model.NotificationPreferenceCenter, error)
	TenantDeliveryDrilldown(context.Context, string) (model.TenantDeliveryDrilldown, error)
	RunTenantDeliveryDrilldown(context.Context, string, model.RunDeliveryDrilldownRequest) (model.TenantDeliveryDrilldown, error)
	TenantDeliveryRemediation(context.Context, string) (model.TenantDeliveryRemediation, error)
	RunTenantDeliveryRemediation(context.Context, string, model.RunDeliveryRemediationRequest) (model.TenantDeliveryRemediation, error)
	TenantProviderSimulationLab(context.Context, string) (model.TenantProviderSimulationLab, error)
	RunTenantProviderSimulation(context.Context, string, model.RunProviderSimulationRequest) (model.TenantProviderSimulationLab, error)
	TenantPackageBillingReadiness(context.Context, string) (model.TenantPackageBillingReadiness, error)
	TenantOperationsSummary(context.Context, string) (model.TenantOperationsSummary, error)
	TenantMonetizationSummary(context.Context, string) (model.TenantMonetizationSummary, error)
	TenantBusinessDashboard(context.Context, string) (model.TenantBusinessDashboard, error)
	TenantRoleExperiences(context.Context, string) (model.TenantRoleExperience, error)
	TenantCustomerControlRoom(context.Context, string) (model.TenantCustomerControlRoom, error)
	TenantCustomerSuccessPacket(context.Context, string) (model.TenantCustomerSuccessPacket, error)
	TenantExecutiveConsole(context.Context, string) (model.TenantExecutiveConsole, error)
	TenantNotificationRevenueCockpit(context.Context, string) (model.TenantNotificationRevenueCockpit, error)
	TenantAlertInbox(context.Context, string) (model.TenantAlertInbox, error)
	TenantNotificationCommandCenter(context.Context, string) (model.TenantNotificationCommandCenter, error)
	TenantDeliveryTimeline(context.Context, string, model.TenantDeliveryTimelineFilter) (model.TenantDeliveryTimeline, error)
	TenantSyncHealth(context.Context, string) (model.TenantSyncHealth, error)
	TenantActivityFeed(context.Context, string, model.TenantActivityFeedFilter) (model.TenantActivityFeed, error)
	CreateTenantActivityView(context.Context, string, model.CreateTenantActivityViewRequest) (model.TenantActivityView, error)
	ListTenantActivityViews(context.Context, string) []model.TenantActivityView
	CreateTenantDataExport(context.Context, string, model.CreateTenantDataExportRequest) (model.TenantDataExport, error)
	ListTenantDataExports(context.Context, string) []model.TenantDataExport
	CreateDeleteRequest(context.Context, string, model.CreateDeleteRequestRequest) (model.DeleteRequest, error)
	ListDeleteRequests(context.Context, string) []model.DeleteRequest
	CreateDeviceGroup(context.Context, string, model.CreateDeviceGroupRequest) (model.DeviceGroup, error)
	ListDeviceGroups(context.Context, string) []model.DeviceGroup
	CreatePolicyAssignment(context.Context, string, model.CreatePolicyAssignmentRequest) (model.PolicyAssignment, error)
	ListPolicyAssignments(context.Context, string) []model.PolicyAssignment
	EnrollDevice(context.Context, model.EnrollDeviceRequest) (model.Device, error)
	ListDevices(context.Context) []model.Device
	GetDevice(context.Context, string) (model.Device, error)
	DailySummary(context.Context, string, string) (model.DeviceSummary, error)
	HostOverview(context.Context, string) (model.HostOverview, error)
	DeviceHealth(context.Context, string) (model.DeviceHealth, error)
	ListPolicyViolations(context.Context, string) ([]model.RiskEvent, error)
	ListAnomalies(context.Context, string) ([]model.RiskEvent, error)
	ListTamperEvents(context.Context, string) ([]model.RiskEvent, error)
	ListAlertDeliveries(context.Context, string) ([]model.AlertDelivery, error)
	IngestTelemetryEvents(context.Context, string, model.IngestTelemetryRequest) (model.IngestTelemetryResponse, error)
	TelemetryIngestStatus(context.Context, string) (model.TelemetryIngestStatus, error)
}
