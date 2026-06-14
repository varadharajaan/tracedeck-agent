package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"

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
	ingest, err := first.IngestTelemetryEvents(ctx, device.DeviceID, model.IngestTelemetryRequest{
		TenantID: "family-varadha",
		DeviceID: device.DeviceID,
		HostName: device.HostName,
		Profile:  device.Profile,
		OSName:   device.OSName,
		Events: []model.TelemetryEvent{{
			Type:       "process_snapshot",
			Source:     constants.RiskSourceProcess,
			ObservedAt: device.LastSeenAt,
			AppName:    "Code.exe",
			ProcessID:  123,
			PathHash:   "hash-only",
			Metadata:   map[string]string{"category": "coding"},
		}},
	})
	if err != nil {
		t.Fatalf("ingest telemetry: %v", err)
	}
	if ingest.AcceptedEvents != 1 || ingest.StoredEvents != 1 {
		t.Fatalf("expected telemetry ingest counts: %+v", ingest)
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
	telemetryStatus, err := second.TelemetryIngestStatus(ctx, "persistent-device")
	if err != nil {
		t.Fatalf("load telemetry status after restart: %v", err)
	}
	if telemetryStatus.StoredEvents != 1 || telemetryStatus.CountsByType["process_snapshot"] != 1 {
		t.Fatalf("expected persisted telemetry status: %+v", telemetryStatus)
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

func TestTelemetryIngestIsIdempotentForStableEventIDs(t *testing.T) {
	t.Parallel()

	repo := NewMemory()
	ctx := context.Background()
	_, err := repo.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	_, err = repo.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "replay-device",
		HostName: "replay-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}

	request := model.IngestTelemetryRequest{
		TenantID: "family-varadha",
		DeviceID: "replay-device",
		HostName: "replay-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
		Events: []model.TelemetryEvent{{
			ID:         "local-event-1",
			Type:       "process_snapshot",
			Source:     constants.RiskSourceProcess,
			ObservedAt: time.Date(2026, 6, 12, 8, 0, 0, 0, time.UTC),
			AppName:    "Code.exe",
			ProcessID:  123,
			PathHash:   "hash-only",
			Metadata:   map[string]string{"category": "coding"},
		}},
	}
	first, err := repo.IngestTelemetryEvents(ctx, "replay-device", request)
	if err != nil {
		t.Fatalf("first ingest: %v", err)
	}
	if first.AcceptedEvents != 1 || first.StoredEvents != 1 {
		t.Fatalf("expected first ingest to store event: %+v", first)
	}
	second, err := repo.IngestTelemetryEvents(ctx, "replay-device", request)
	if err != nil {
		t.Fatalf("second ingest: %v", err)
	}
	if second.AcceptedEvents != 1 || second.StoredEvents != 1 {
		t.Fatalf("expected duplicate ingest to be acknowledged without duplicate storage: %+v", second)
	}
	status, err := repo.TelemetryIngestStatus(ctx, "replay-device")
	if err != nil {
		t.Fatalf("telemetry status: %v", err)
	}
	if status.StoredEvents != 1 || status.RecentEvents[0].ID != "local-event-1" {
		t.Fatalf("expected one stored telemetry event: %+v", status)
	}
}

func TestTenantSyncHealthSummarizesOfflineReplayProof(t *testing.T) {
	t.Parallel()

	repo := NewMemory()
	ctx := context.Background()
	_, err := repo.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	_, err = repo.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "sync-device",
		HostName: "sync-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}

	_, err = repo.IngestTelemetryEvents(ctx, "sync-device", model.IngestTelemetryRequest{
		TenantID: "family-varadha",
		DeviceID: "sync-device",
		HostName: "sync-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
		Events: []model.TelemetryEvent{
			{
				ID:         "local-event-7",
				Type:       "process.observed",
				Source:     "collector.process",
				ObservedAt: time.Date(2026, 6, 12, 8, 0, 0, 0, time.UTC),
				AppName:    "Code.exe",
				PathHash:   "hash-only",
				Metadata:   map[string]string{"category": "coding"},
			},
			{
				ID:         "local-event-8",
				Type:       "device.health",
				Source:     "collector.device.health",
				ObservedAt: time.Date(2026, 6, 12, 8, 1, 0, 0, time.UTC),
				Metadata:   map[string]string{"battery": "ok"},
			},
			{
				ID:         "local-event-9",
				Type:       constants.TelemetryTypeAgentHeartbeat,
				Source:     constants.TelemetrySourceAgentHeartbeat,
				ObservedAt: time.Date(2026, 6, 12, 8, 2, 0, 0, time.UTC),
				Metadata:   map[string]string{"agent_healthy": "true", "agent_version": "0.1.0-dev"},
			},
		},
	})
	if err != nil {
		t.Fatalf("ingest telemetry: %v", err)
	}

	summary, err := repo.TenantSyncHealth(ctx, "family-varadha")
	if err != nil {
		t.Fatalf("tenant sync health: %v", err)
	}
	if summary.HostsReporting != 1 || summary.HostsPending != 0 || summary.StoredEvents != 3 {
		t.Fatalf("unexpected sync health counts: %+v", summary)
	}
	if summary.LastLocalEventID != 9 || !summary.OfflineReplayReady || !summary.BackendVisible {
		t.Fatalf("expected replay proof and backend visibility: %+v", summary)
	}
	if summary.PrivacyBoundary == "" || len(summary.Devices) != 1 {
		t.Fatalf("expected privacy boundary and one device summary: %+v", summary)
	}
	device := summary.Devices[0]
	if device.ProcessEvents != 1 || device.HealthEvents != 2 || device.LastLocalEventID != 9 {
		t.Fatalf("unexpected device source counts: %+v", device)
	}
	if len(device.RecentEventIDs) != 3 || device.RecentEventIDs[0] != "local-event-9" {
		t.Fatalf("expected recent stable event IDs newest first: %+v", device.RecentEventIDs)
	}
}

func TestTenantActivityFeedFiltersRiskDeliveryAndTelemetry(t *testing.T) {
	t.Parallel()

	repo := NewMemory()
	ctx := context.Background()
	_, err := repo.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	_, err = repo.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "feed-device",
		HostName: "feed-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}
	_, err = repo.IngestTelemetryEvents(ctx, "feed-device", model.IngestTelemetryRequest{
		TenantID: "family-varadha",
		DeviceID: "feed-device",
		HostName: "feed-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
		Events: []model.TelemetryEvent{{
			ID:         "local-event-21",
			Type:       "process.observed",
			Source:     "collector.process",
			ObservedAt: time.Date(2026, 6, 12, 8, 0, 0, 0, time.UTC),
			AppName:    "Code.exe",
			PathHash:   "hash-only",
			Metadata:   map[string]string{"category": "coding"},
		}},
	})
	if err != nil {
		t.Fatalf("ingest telemetry: %v", err)
	}

	feed, err := repo.TenantActivityFeed(ctx, "family-varadha", model.TenantActivityFeedFilter{DeviceID: "feed-device", Limit: 20})
	if err != nil {
		t.Fatalf("tenant activity feed: %v", err)
	}
	if feed.Summary.RiskItems == 0 || feed.Summary.DeliveryItems == 0 || feed.Summary.TelemetryItems != 1 {
		t.Fatalf("expected risk, delivery, and telemetry feed items: %+v", feed.Summary)
	}
	if feed.Summary.SourceHostCount != 1 || feed.Summary.ReportingHosts != 1 || feed.PrivacyBoundary == "" {
		t.Fatalf("expected host and privacy proof: %+v", feed)
	}

	emailFeed, err := repo.TenantActivityFeed(ctx, "family-varadha", model.TenantActivityFeedFilter{
		DeviceID: "feed-device",
		Kind:     constants.ActivityFeedKindDelivery,
		Channel:  constants.DeliveryChannelEmail,
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("tenant activity feed email filter: %v", err)
	}
	if emailFeed.Summary.Total == 0 || emailFeed.Items[0].Channel != constants.DeliveryChannelEmail {
		t.Fatalf("expected email delivery feed item: %+v", emailFeed)
	}

	queryFeed, err := repo.TenantActivityFeed(ctx, "family-varadha", model.TenantActivityFeedFilter{
		DeviceID: "feed-device",
		Query:    "youtube",
		Limit:    5,
	})
	if err != nil {
		t.Fatalf("tenant activity feed query filter: %v", err)
	}
	if queryFeed.Summary.Total == 0 || queryFeed.Items[0].DeviceID != "feed-device" {
		t.Fatalf("expected query feed to find selected host events: %+v", queryFeed)
	}
}

func TestTenantDeliveryTimelineFiltersDeliveryProof(t *testing.T) {
	t.Parallel()

	repo := NewMemory()
	ctx := context.Background()
	_, err := repo.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	_, err = repo.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "timeline-device",
		HostName: "timeline-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}

	timeline, err := repo.TenantDeliveryTimeline(ctx, "family-varadha", model.TenantDeliveryTimelineFilter{DeviceID: "timeline-device", Limit: 10})
	if err != nil {
		t.Fatalf("tenant delivery timeline: %v", err)
	}
	if timeline.Summary.Total < 3 || timeline.Summary.Email == 0 || timeline.Summary.Push == 0 || timeline.Summary.Dashboard == 0 {
		t.Fatalf("expected email, push, and dashboard timeline proof: %+v", timeline.Summary)
	}
	if timeline.Summary.RouteProofGaps == 0 || timeline.Summary.NotificationScore == 0 || timeline.Summary.RecommendedPaidTier == "" {
		t.Fatalf("expected route proof gaps, score, and paid tier: %+v", timeline.Summary)
	}
	if len(timeline.Items) == 0 || timeline.Items[0].PaidTier == "" || timeline.Items[0].NextAction == "" {
		t.Fatalf("expected typed timeline item action metadata: %+v", timeline.Items)
	}

	filtered, err := repo.TenantDeliveryTimeline(ctx, "family-varadha", model.TenantDeliveryTimelineFilter{
		DeviceID: "timeline-device",
		Channel:  constants.DeliveryChannelPush,
		Status:   constants.DeliveryStatusRetrying,
		Provider: constants.DeliveryProviderWebPush,
		Query:    "youtube",
		Limit:    1,
	})
	if err != nil {
		t.Fatalf("tenant delivery timeline filter: %v", err)
	}
	if len(filtered.Items) != 1 || filtered.Items[0].Channel != constants.DeliveryChannelPush || filtered.Items[0].Status != constants.DeliveryStatusRetrying {
		t.Fatalf("expected filtered push retry timeline item: %+v", filtered)
	}
}

func TestTenantDeliveryAssuranceSeparatesDemoFromProviderProof(t *testing.T) {
	t.Parallel()

	repo := NewMemory()
	ctx := context.Background()
	_, err := repo.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	_, err = repo.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "assurance-device",
		HostName: "assurance-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}

	assurance, err := repo.TenantDeliveryAssurance(ctx, "family-varadha", model.TenantDeliveryAssuranceFilter{DeviceID: "assurance-device", Limit: 10})
	if err != nil {
		t.Fatalf("tenant delivery assurance: %v", err)
	}
	if assurance.Summary.RoutesTotal != 3 || assurance.Summary.ProviderConfirmed != 0 || assurance.Summary.DemoOnly == 0 || assurance.Summary.Retrying == 0 || !assurance.Summary.DashboardRouteReady {
		t.Fatalf("expected demo-only email, retrying push, dashboard fallback, and no provider-confirmed proof: %+v", assurance.Summary)
	}
	if assurance.Summary.BuyerReady {
		t.Fatalf("seeded delivery rows must not be buyer-ready provider proof: %+v", assurance.Summary)
	}
	if len(assurance.Routes) != 3 || assurance.PrivacyBoundary == "" {
		t.Fatalf("expected route assurance rows and privacy boundary: %+v", assurance)
	}

	filtered, err := repo.TenantDeliveryAssurance(ctx, "family-varadha", model.TenantDeliveryAssuranceFilter{
		DeviceID:       "assurance-device",
		Channel:        constants.DeliveryChannelEmail,
		AssuranceState: constants.DeliveryAssuranceDemoOnly,
		Limit:          5,
	})
	if err != nil {
		t.Fatalf("tenant delivery assurance filter: %v", err)
	}
	if len(filtered.Routes) != 1 || filtered.Routes[0].AssuranceState != constants.DeliveryAssuranceDemoOnly || filtered.Routes[0].SourceKind != constants.EvidenceSourceDemoSeed {
		t.Fatalf("expected filtered email demo-only proof: %+v", filtered.Routes)
	}
	if filtered.Routes[0].OperatorTruth == "" || filtered.Routes[0].UserVisibleLabel == "" || filtered.Routes[0].NextAction == "" {
		t.Fatalf("expected operator truth labels: %+v", filtered.Routes[0])
	}
}

func TestTenantRoleExperiences(t *testing.T) {
	t.Parallel()

	repo := NewMemory()
	ctx := context.Background()
	_, err := repo.CreateTenant(ctx, model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	_, err = repo.EnrollDevice(ctx, model.EnrollDeviceRequest{
		TenantID: "family-varadha",
		DeviceID: "role-device",
		HostName: "role-host",
		Profile:  "ai-btech-student",
		OSName:   "windows",
	})
	if err != nil {
		t.Fatalf("enroll device: %v", err)
	}

	experience, err := repo.TenantRoleExperiences(ctx, "family-varadha")
	if err != nil {
		t.Fatalf("tenant role experiences: %v", err)
	}
	if experience.Summary.RolesTotal != 4 || experience.Summary.ReadinessScore == 0 || experience.Summary.RecommendedPackage == "" {
		t.Fatalf("expected scored role experience summary: %+v", experience.Summary)
	}
	if len(experience.Roles) != 4 || len(experience.Onboarding) < 4 {
		t.Fatalf("expected role cards and onboarding checklist: %+v", experience)
	}
	roleIDs := map[string]bool{}
	for _, role := range experience.Roles {
		roleIDs[role.RoleID] = true
		if role.PaidTier == "" || role.PrimaryGoal == "" || role.NotificationPromise == "" || len(role.Metrics) == 0 {
			t.Fatalf("expected role card metadata: %+v", role)
		}
	}
	for _, roleID := range []string{constants.RoleParent, constants.RoleStudent, constants.RoleSchoolAdmin, constants.RoleBusinessManager} {
		if !roleIDs[roleID] {
			t.Fatalf("expected role %q in role experiences: %+v", roleID, experience.Roles)
		}
	}
	if experience.PrivacyBoundary == "" {
		t.Fatalf("expected privacy boundary")
	}
}
