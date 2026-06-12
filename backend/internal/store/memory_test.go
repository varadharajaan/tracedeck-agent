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
}
