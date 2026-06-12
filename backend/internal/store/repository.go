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
	TenantOperationsSummary(context.Context, string) (model.TenantOperationsSummary, error)
	TenantMonetizationSummary(context.Context, string) (model.TenantMonetizationSummary, error)
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
}
