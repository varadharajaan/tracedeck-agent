package store

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
)

func TestPersistentStoreSurvivesRestart(t *testing.T) {
	t.Parallel()

	statePath := filepath.Join(t.TempDir(), "backend-state.json")
	first, err := NewPersistent(statePath)
	if err != nil {
		t.Fatalf("create first store: %v", err)
	}

	ctx := context.Background()
	tenant, err := first.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if tenant.TenantID != "family-varadha" {
		t.Fatalf("unexpected tenant: %+v", tenant)
	}

	device, err := first.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "persistent-device",
		HostName: "persistent-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}
	if _, err := first.HostOverview(ctx, device.DeviceID); err != nil {
		t.Fatalf("seed host overview: %v", err)
	}
	rule, err := first.CreateAlertRule(ctx, "family-varadha", model.CreateAlertRuleRequest{
		TemplateID: constants.AlertRuleTemplateRiskySoftware,
		Name:       "Persist risky software rule",
		Trigger:    constants.AlertTriggerRiskySoftware,
		Severity:   constants.SeverityHigh,
		Channels:   []string{constants.DeliveryChannelEmail, constants.DeliveryChannelDashboard},
		Condition: model.AlertRuleCondition{
			Subject:  constants.AlertConditionSubjectCategory,
			Operator: constants.AlertConditionOperatorEquals,
			Value:    "torrent_client",
		},
		Enabled: true,
	})
	if err != nil {
		t.Fatalf("create alert rule: %v", err)
	}
	if rule.ID == "" {
		t.Fatalf("expected alert rule id: %+v", rule)
	}
	route, err := first.CreateNotificationRoute(ctx, "family-varadha", model.CreateNotificationRouteRequest{
		Channel:        constants.DeliveryChannelPush,
		Provider:       constants.DeliveryProviderWebPush,
		RecipientLabel: "persistent parent mobile route",
		Status:         constants.StatusWatch,
		Enabled:        true,
		LastSummary:    "persistent route waiting for first delivered push proof",
	})
	if err != nil {
		t.Fatalf("create notification route: %v", err)
	}
	if route.ID == "" {
		t.Fatalf("expected notification route id: %+v", route)
	}
	group, err := first.CreateDeviceGroup(ctx, "family-varadha", model.CreateDeviceGroupRequest{
		Name:             "Persistent exam devices",
		Description:      "Devices assigned to exam mode",
		Profile:          "school-laptop",
		DeviceIDs:        []string{"persistent-device"},
		PolicyTemplateID: "school-laptop",
	})
	if err != nil {
		t.Fatalf("create device group: %v", err)
	}
	if group.ID == "" {
		t.Fatalf("expected device group id: %+v", group)
	}
	assignment, err := first.CreatePolicyAssignment(ctx, "family-varadha", model.CreatePolicyAssignmentRequest{
		Name:             "Persistent exam rollout",
		TargetType:       constants.PolicyAssignmentTargetDeviceGroup,
		TargetID:         group.ID,
		PolicyTemplateID: "school-laptop",
		AlertRuleIDs:     []string{rule.ID},
		Mode:             constants.PolicyAssignmentModeActive,
	})
	if err != nil {
		t.Fatalf("create policy assignment: %v", err)
	}
	if assignment.ID == "" {
		t.Fatalf("expected policy assignment id: %+v", assignment)
	}
	export, err := first.CreateTenantDataExport(ctx, "family-varadha", model.CreateTenantDataExportRequest{
		Format: constants.DataExportFormatJSON,
		Scope:  constants.DataExportScopeTenant,
	})
	if err != nil {
		t.Fatalf("create data export: %v", err)
	}
	if export.ID == "" || export.StorageKey == "" {
		t.Fatalf("expected data export id and key: %+v", export)
	}
	deleteRequest, err := first.CreateDeleteRequest(ctx, "family-varadha", model.CreateDeleteRequestRequest{
		Scope:  constants.DeleteRequestScopeTenant,
		Reason: "cleanup request",
	})
	if err != nil {
		t.Fatalf("create delete request: %v", err)
	}
	if deleteRequest.ID == "" || deleteRequest.Status != constants.DeleteRequestStatusQueued {
		t.Fatalf("expected queued delete request: %+v", deleteRequest)
	}
	summary, err := first.TenantOperationsSummary(ctx, "family-varadha")
	if err != nil {
		t.Fatalf("create tenant operations summary: %v", err)
	}
	if summary.HostsTotal != 1 || summary.LastEmail == nil || len(summary.PrioritySignals) == 0 {
		t.Fatalf("expected tenant operations summary: %+v", summary)
	}

	second, err := NewPersistent(statePath)
	if err != nil {
		t.Fatalf("create second store: %v", err)
	}
	loadedDevice, err := second.GetDevice(ctx, "persistent-device")
	if err != nil {
		t.Fatalf("load device after restart: %v", err)
	}
	if loadedDevice.HostName != "persistent-host" {
		t.Fatalf("unexpected loaded device: %+v", loadedDevice)
	}
	events, err := second.ListPolicyViolations(ctx, "persistent-device")
	if err != nil {
		t.Fatalf("load policy events after restart: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected persisted policy events after restart")
	}
	health, err := second.DeviceHealth(ctx, "persistent-device")
	if err != nil {
		t.Fatalf("load device health after restart: %v", err)
	}
	if health.Score == 0 || health.Status == "" {
		t.Fatalf("expected persisted device health: %+v", health)
	}
	rules := second.ListAlertRules(ctx, "family-varadha")
	if len(rules) < 3 {
		t.Fatalf("expected seeded and custom alert rules after restart: %+v", rules)
	}
	routes := second.ListNotificationRoutes(ctx, "family-varadha")
	if len(routes) < 4 {
		t.Fatalf("expected seeded and custom notification routes after restart: %+v", routes)
	}
	groups := second.ListDeviceGroups(ctx, "family-varadha")
	if len(groups) < 2 {
		t.Fatalf("expected seeded and custom device groups after restart: %+v", groups)
	}
	assignments := second.ListPolicyAssignments(ctx, "family-varadha")
	if len(assignments) < 2 {
		t.Fatalf("expected seeded and custom policy assignments after restart: %+v", assignments)
	}
	exports := second.ListTenantDataExports(ctx, "family-varadha")
	if len(exports) != 1 || exports[0].StorageKey == "" {
		t.Fatalf("expected persisted data export after restart: %+v", exports)
	}
	deleteRequests := second.ListDeleteRequests(ctx, "family-varadha")
	if len(deleteRequests) != 1 || deleteRequests[0].Status != constants.DeleteRequestStatusQueued {
		t.Fatalf("expected persisted delete request after restart: %+v", deleteRequests)
	}
	loadedSummary, err := second.TenantOperationsSummary(ctx, "family-varadha")
	if err != nil {
		t.Fatalf("load tenant operations summary after restart: %v", err)
	}
	if loadedSummary.HostsTotal != 1 || loadedSummary.DeliveryTotal == 0 || loadedSummary.MonetizationReadiness == 0 {
		t.Fatalf("expected loaded tenant operations summary: %+v", loadedSummary)
	}
	monetization, err := second.TenantMonetizationSummary(ctx, "family-varadha")
	if err != nil {
		t.Fatalf("load tenant monetization summary after restart: %v", err)
	}
	if monetization.ReadinessScore == 0 || monetization.NotificationScore == 0 || len(monetization.NotificationRoutes) != 3 {
		t.Fatalf("expected loaded tenant monetization summary: %+v", monetization)
	}
}
