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
	EnrollDevice(context.Context, model.EnrollDeviceRequest) (model.Device, error)
	ListDevices(context.Context) []model.Device
	GetDevice(context.Context, string) (model.Device, error)
	DailySummary(context.Context, string, string) (model.DeviceSummary, error)
	HostOverview(context.Context, string) (model.HostOverview, error)
	ListPolicyViolations(context.Context, string) ([]model.RiskEvent, error)
	ListAnomalies(context.Context, string) ([]model.RiskEvent, error)
	ListTamperEvents(context.Context, string) ([]model.RiskEvent, error)
	ListAlertDeliveries(context.Context, string) ([]model.AlertDelivery, error)
}
