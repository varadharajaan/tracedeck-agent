package api

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/store"
)

func TestHealthAndVersion(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, constants.RouteHealth, nil))
	if health.Code != http.StatusOK {
		t.Fatalf("expected health 200, got %d", health.Code)
	}
	if !strings.Contains(health.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected ok health response, got %s", health.Body.String())
	}

	version := httptest.NewRecorder()
	handler.ServeHTTP(version, httptest.NewRequest(http.MethodGet, constants.RouteVersion, nil))
	if version.Code != http.StatusOK {
		t.Fatalf("expected version 200, got %d", version.Code)
	}
	if !strings.Contains(version.Body.String(), constants.BackendName) {
		t.Fatalf("expected backend name in version response, got %s", version.Body.String())
	}
}

func TestDeviceEnrollmentAndLookup(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	body := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "laptop-cousin-001",
		"host_name": "study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)

	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(body)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, constants.RouteDevices, nil))
	if list.Code != http.StatusOK {
		t.Fatalf("expected list 200, got %d", list.Code)
	}
	var listed model.ListResponse[model.Device]
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listed.Count != 1 || listed.Items[0].DeviceID != "laptop-cousin-001" {
		t.Fatalf("unexpected list response: %+v", listed)
	}

	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("expected get 200, got %d", get.Code)
	}
}

func TestTelemetryIngestEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)
	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	body := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "agent-device-001",
		"host_name": "agent-host",
		"profile": "ai-btech-student",
		"os_name": "windows",
		"events": [
			{
				"type": "process_snapshot",
				"source": "process",
				"observed_at": "2026-06-12T08:00:00Z",
				"app_name": "Code.exe",
				"process_id": 123,
				"path_hash": "hash-only",
				"metadata": { "category": "coding" }
			}
		]
	}`)
	ingest := httptest.NewRecorder()
	handler.ServeHTTP(ingest, httptest.NewRequest(http.MethodPost, constants.RouteDevices+"/agent-device-001/"+constants.RouteSegmentTelemetry, bytes.NewReader(body)))
	if ingest.Code != http.StatusAccepted {
		t.Fatalf("expected telemetry ingest 202, got %d: %s", ingest.Code, ingest.Body.String())
	}
	var response model.IngestTelemetryResponse
	if err := json.Unmarshal(ingest.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode ingest response: %v", err)
	}
	if response.AcceptedEvents != 1 || response.PrivacyBoundary == "" || !response.BackendVisibleHost {
		t.Fatalf("unexpected ingest response: %+v", response)
	}

	status := httptest.NewRecorder()
	handler.ServeHTTP(status, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/agent-device-001/"+constants.RouteSegmentTelemetryStatus, nil))
	if status.Code != http.StatusOK {
		t.Fatalf("expected telemetry status 200, got %d: %s", status.Code, status.Body.String())
	}
	var telemetryStatus model.TelemetryIngestStatus
	if err := json.Unmarshal(status.Body.Bytes(), &telemetryStatus); err != nil {
		t.Fatalf("decode telemetry status: %v", err)
	}
	if telemetryStatus.StoredEvents != 1 || telemetryStatus.CountsBySource["process"] != 1 || telemetryStatus.RecentEvents[0].PathHash != "hash-only" {
		t.Fatalf("unexpected telemetry status: %+v", telemetryStatus)
	}
}

func TestHostDashboardRiskEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	body := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "laptop-cousin-001",
		"host_name": "study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)

	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(body)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	overview := httptest.NewRecorder()
	handler.ServeHTTP(overview, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentOverview, nil))
	if overview.Code != http.StatusOK {
		t.Fatalf("expected overview 200, got %d: %s", overview.Code, overview.Body.String())
	}
	var host model.HostOverview
	if err := json.Unmarshal(overview.Body.Bytes(), &host); err != nil {
		t.Fatalf("decode host overview: %v", err)
	}
	if host.Device.DeviceID != "laptop-cousin-001" || host.Summary.PolicyViolations == 0 {
		t.Fatalf("unexpected host overview: %+v", host)
	}
	if len(host.AlertDeliveries) == 0 || host.AlertDeliveries[0].Channel == "" {
		t.Fatalf("expected alert delivery visibility: %+v", host.AlertDeliveries)
	}
	if host.Health.Score == 0 || host.Health.Status == "" {
		t.Fatalf("expected host health score: %+v", host.Health)
	}

	health := httptest.NewRecorder()
	handler.ServeHTTP(health, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentHealth, nil))
	if health.Code != http.StatusOK {
		t.Fatalf("expected device health 200, got %d", health.Code)
	}
	var deviceHealth model.DeviceHealth
	if err := json.Unmarshal(health.Body.Bytes(), &deviceHealth); err != nil {
		t.Fatalf("decode device health: %v", err)
	}
	if deviceHealth.DeviceID != "laptop-cousin-001" || !deviceHealth.AgentHealthy {
		t.Fatalf("unexpected device health: %+v", deviceHealth)
	}

	policy := httptest.NewRecorder()
	handler.ServeHTTP(policy, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentPolicyEvents, nil))
	if policy.Code != http.StatusOK {
		t.Fatalf("expected policy violations 200, got %d", policy.Code)
	}
	var policyEvents model.ListResponse[model.RiskEvent]
	if err := json.Unmarshal(policy.Body.Bytes(), &policyEvents); err != nil {
		t.Fatalf("decode policy violations: %v", err)
	}
	if policyEvents.Count == 0 || policyEvents.Items[0].Type != constants.RiskTypePolicyViolation {
		t.Fatalf("unexpected policy events: %+v", policyEvents)
	}

	deliveries := httptest.NewRecorder()
	handler.ServeHTTP(deliveries, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentAlertDelivery, nil))
	if deliveries.Code != http.StatusOK {
		t.Fatalf("expected alert deliveries 200, got %d", deliveries.Code)
	}
	if !strings.Contains(deliveries.Body.String(), constants.DeliveryChannelEmail) {
		t.Fatalf("expected email delivery visibility, got %s", deliveries.Body.String())
	}

	weekly := httptest.NewRecorder()
	handler.ServeHTTP(weekly, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentReports+"/"+constants.RouteSegmentWeekly, nil))
	if weekly.Code != http.StatusOK {
		t.Fatalf("expected weekly report 200, got %d", weekly.Code)
	}
	var report model.WeeklyReport
	if err := json.Unmarshal(weekly.Body.Bytes(), &report); err != nil {
		t.Fatalf("decode weekly report: %v", err)
	}
	if !report.Generated || !report.EmailReady || !report.PDFReady {
		t.Fatalf("expected generated weekly report: %+v", report)
	}

	pdf := httptest.NewRecorder()
	handler.ServeHTTP(pdf, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentReports+"/"+constants.RouteSegmentWeekly+"/"+constants.RouteSegmentPDF, nil))
	if pdf.Code != http.StatusOK {
		t.Fatalf("expected weekly report pdf 200, got %d", pdf.Code)
	}
	if pdf.Header().Get("Content-Type") != constants.ContentTypePDF {
		t.Fatalf("expected pdf content type, got %s", pdf.Header().Get("Content-Type"))
	}
	if !bytes.HasPrefix(pdf.Body.Bytes(), []byte("%PDF-1.4")) {
		t.Fatalf("expected pdf body, got %q", pdf.Body.String())
	}
}

func TestHostDashboardRiskEndpointNotFound(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/missing-device/"+constants.RouteSegmentOverview, nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("expected overview 404 for missing device, got %d", recorder.Code)
	}
}

func TestAPIKeyMiddlewareAndTenantScope(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServerWithAuth(repo, slog.Default(), AuthConfig{
		APIKey:   "local-secret",
		TenantID: "tenant-a",
		ActorID:  "local-test",
		RoleID:   constants.RoleParent,
	}).Handler()

	noKey := httptest.NewRecorder()
	handler.ServeHTTP(noKey, httptest.NewRequest(http.MethodGet, constants.RouteDevices, nil))
	if noKey.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing api key 401, got %d", noKey.Code)
	}

	tenantABody := []byte(`{
		"tenant_id": "tenant-a",
		"name": "Tenant A",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)
	createTenant := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantABody))
	req.Header.Set(constants.HeaderAPIKey, "local-secret")
	handler.ServeHTTP(createTenant, req)
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	forbiddenTenantBody := []byte(`{
		"tenant_id": "tenant-b",
		"name": "Tenant B",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)
	forbidden := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(forbiddenTenantBody))
	req.Header.Set(constants.HeaderAPIKey, "local-secret")
	handler.ServeHTTP(forbidden, req)
	if forbidden.Code != http.StatusForbidden {
		t.Fatalf("expected tenant scope 403, got %d", forbidden.Code)
	}

	deviceBody := []byte(`{
		"tenant_id": "tenant-a",
		"device_id": "tenant-a-device",
		"host_name": "tenant-a-host",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody))
	req.Header.Set(constants.HeaderAPIKey, "local-secret")
	handler.ServeHTTP(enroll, req)
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected scoped device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	list := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, constants.RouteDevices, nil)
	req.Header.Set(constants.HeaderAPIKey, "local-secret")
	handler.ServeHTTP(list, req)
	if list.Code != http.StatusOK {
		t.Fatalf("expected scoped list 200, got %d", list.Code)
	}
	var devices model.ListResponse[model.Device]
	if err := json.Unmarshal(list.Body.Bytes(), &devices); err != nil {
		t.Fatalf("decode scoped device list: %v", err)
	}
	if devices.Count != 1 || devices.Items[0].TenantID != "tenant-a" {
		t.Fatalf("unexpected scoped devices: %+v", devices)
	}
}

func TestDeviceEnrollmentValidation(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, strings.NewReader(`{"device_id":"missing-tenant"}`)))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "tenant_id is required") {
		t.Fatalf("expected validation message, got %s", recorder.Body.String())
	}
}

func TestTenantReadinessEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	body := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	create := httptest.NewRecorder()
	handler.ServeHTTP(create, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(body)))
	if create.Code != http.StatusCreated {
		t.Fatalf("expected create tenant 201, got %d: %s", create.Code, create.Body.String())
	}

	list := httptest.NewRecorder()
	handler.ServeHTTP(list, httptest.NewRequest(http.MethodGet, constants.RouteTenants, nil))
	if list.Code != http.StatusOK {
		t.Fatalf("expected tenant list 200, got %d", list.Code)
	}
	var tenants model.ListResponse[model.Tenant]
	if err := json.Unmarshal(list.Body.Bytes(), &tenants); err != nil {
		t.Fatalf("decode tenant list: %v", err)
	}
	if tenants.Count != 1 || tenants.Items[0].PlanID != constants.PlanFamilyPro {
		t.Fatalf("unexpected tenant list: %+v", tenants)
	}

	get := httptest.NewRecorder()
	handler.ServeHTTP(get, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha", nil))
	if get.Code != http.StatusOK {
		t.Fatalf("expected tenant get 200, got %d", get.Code)
	}

	tenantAudit := httptest.NewRecorder()
	handler.ServeHTTP(tenantAudit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/audit-events", nil))
	if tenantAudit.Code != http.StatusOK {
		t.Fatalf("expected tenant audit 200, got %d", tenantAudit.Code)
	}
	if !strings.Contains(tenantAudit.Body.String(), constants.AuditActionTenantCreated) {
		t.Fatalf("expected tenant created audit event, got %s", tenantAudit.Body.String())
	}
}

func TestNoCodeAlertRuleEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	templates := httptest.NewRecorder()
	handler.ServeHTTP(templates, httptest.NewRequest(http.MethodGet, constants.RouteAlertRuleTemplates, nil))
	if templates.Code != http.StatusOK {
		t.Fatalf("expected alert rule templates 200, got %d", templates.Code)
	}
	if !strings.Contains(templates.Body.String(), constants.AlertRuleTemplateMediaAfterHours) {
		t.Fatalf("expected alert rule template catalog, got %s", templates.Body.String())
	}

	initialRules := httptest.NewRecorder()
	handler.ServeHTTP(initialRules, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAlertRules, nil))
	if initialRules.Code != http.StatusOK {
		t.Fatalf("expected seeded rules 200, got %d", initialRules.Code)
	}
	var seeded model.ListResponse[model.AlertRule]
	if err := json.Unmarshal(initialRules.Body.Bytes(), &seeded); err != nil {
		t.Fatalf("decode seeded rules: %v", err)
	}
	if seeded.Count < 2 {
		t.Fatalf("expected seeded alert rules, got %+v", seeded)
	}

	ruleBody := []byte(`{
		"template_id": "risky_software_detected",
		"name": "Email when risky software appears",
		"trigger": "risky_software",
		"severity": "high",
		"channels": ["email", "dashboard"],
		"condition": {
			"subject": "category",
			"operator": "equals",
			"value": "torrent_client"
		},
		"enabled": true
	}`)
	createRule := httptest.NewRecorder()
	handler.ServeHTTP(createRule, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAlertRules, bytes.NewReader(ruleBody)))
	if createRule.Code != http.StatusCreated {
		t.Fatalf("expected rule create 201, got %d: %s", createRule.Code, createRule.Body.String())
	}
	if !strings.Contains(createRule.Body.String(), "Email when risky software appears") {
		t.Fatalf("expected created rule body, got %s", createRule.Body.String())
	}
}

func TestNotificationRouteEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	seededRoutes := httptest.NewRecorder()
	handler.ServeHTTP(seededRoutes, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentNotifications, nil))
	if seededRoutes.Code != http.StatusOK {
		t.Fatalf("expected seeded routes 200, got %d", seededRoutes.Code)
	}
	var routes model.ListResponse[model.NotificationRoute]
	if err := json.Unmarshal(seededRoutes.Body.Bytes(), &routes); err != nil {
		t.Fatalf("decode notification routes: %v", err)
	}
	if routes.Count != 3 {
		t.Fatalf("expected three seeded notification routes: %+v", routes)
	}

	routeBody := []byte(`{
		"channel": "push",
		"provider": "web_push",
		"recipient_label": "parent secondary phone",
		"status": "watch",
		"enabled": true,
		"last_summary": "Waiting for first delivered push proof."
	}`)
	createRoute := httptest.NewRecorder()
	handler.ServeHTTP(createRoute, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentNotifications, bytes.NewReader(routeBody)))
	if createRoute.Code != http.StatusCreated {
		t.Fatalf("expected route create 201, got %d: %s", createRoute.Code, createRoute.Body.String())
	}
	if !strings.Contains(createRoute.Body.String(), "parent secondary phone") {
		t.Fatalf("expected created route body, got %s", createRoute.Body.String())
	}

	invalidRoute := httptest.NewRecorder()
	handler.ServeHTTP(invalidRoute, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentNotifications, strings.NewReader(`{"channel":"email","provider":"web_push","recipient_label":"bad","enabled":true}`)))
	if invalidRoute.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid route 400, got %d: %s", invalidRoute.Code, invalidRoute.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionNotificationRoute) {
		t.Fatalf("expected route audit event, got %s", audit.Body.String())
	}
}

func TestConsentCenterEndpoint(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	consent := httptest.NewRecorder()
	handler.ServeHTTP(consent, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentConsentCenter, nil))
	if consent.Code != http.StatusOK {
		t.Fatalf("expected consent center 200, got %d: %s", consent.Code, consent.Body.String())
	}
	var center model.ConsentCenter
	if err := json.Unmarshal(consent.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode consent center: %v", err)
	}
	if !center.MonitoringVisible || !center.DataExportReady || !center.DeleteRequestReady {
		t.Fatalf("expected consent readiness flags: %+v", center)
	}
	if len(center.AuditEvents) == 0 || len(center.AlertRecipients) == 0 {
		t.Fatalf("expected audit and recipient visibility: %+v", center)
	}
	if !strings.Contains(consent.Body.String(), constants.ConsentCollectionPasswords) || !strings.Contains(consent.Body.String(), constants.ConsentStatusDenied) {
		t.Fatalf("expected denied sensitive collection disclosure, got %s", consent.Body.String())
	}
}

func TestTenantOperationsSummaryEndpoint(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	deviceBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "ops-device-001",
		"host_name": "ops-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	summaryResponse := httptest.NewRecorder()
	handler.ServeHTTP(summaryResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentOperations, nil))
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("expected operations summary 200, got %d: %s", summaryResponse.Code, summaryResponse.Body.String())
	}
	var summary model.TenantOperationsSummary
	if err := json.Unmarshal(summaryResponse.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode operations summary: %v", err)
	}
	if summary.HostsTotal != 1 || summary.OpenPolicyViolations == 0 || summary.DeliveryTotal == 0 {
		t.Fatalf("unexpected operations summary: %+v", summary)
	}
	if summary.LastEmail == nil || summary.LastEmail.Channel != constants.DeliveryChannelEmail {
		t.Fatalf("expected latest email delivery proof: %+v", summary.LastEmail)
	}
	if len(summary.PrioritySignals) == 0 || len(summary.UpgradeSignals) == 0 {
		t.Fatalf("expected priority and upgrade signals: %+v", summary)
	}
}

func TestTenantMonetizationSummaryEndpoint(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	deviceBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "monetization-device-001",
		"host_name": "monetization-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	summaryResponse := httptest.NewRecorder()
	handler.ServeHTTP(summaryResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentMonetization, nil))
	if summaryResponse.Code != http.StatusOK {
		t.Fatalf("expected monetization summary 200, got %d: %s", summaryResponse.Code, summaryResponse.Body.String())
	}
	var summary model.TenantMonetizationSummary
	if err := json.Unmarshal(summaryResponse.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode monetization summary: %v", err)
	}
	if summary.PlanID != constants.PlanFamilyPro || summary.ReadinessScore == 0 || summary.NotificationScore == 0 {
		t.Fatalf("unexpected monetization scores: %+v", summary)
	}
	if len(summary.NotificationRoutes) != 3 || summary.NotificationRoutes[0].Channel == "" {
		t.Fatalf("expected email, push, and dashboard route proof: %+v", summary.NotificationRoutes)
	}
	if len(summary.ValuePanels) < 4 || len(summary.PaidCapabilities) < 4 || len(summary.ConversionActions) == 0 {
		t.Fatalf("expected product value surfaces: %+v", summary)
	}
	if summary.NotificationPromise.Email == "" || summary.NotificationPromise.Push == "" {
		t.Fatalf("expected notification promise lines: %+v", summary.NotificationPromise)
	}
}

func TestDeviceGroupAndPolicyAssignmentEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	seededGroups := httptest.NewRecorder()
	handler.ServeHTTP(seededGroups, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeviceGroups, nil))
	if seededGroups.Code != http.StatusOK {
		t.Fatalf("expected seeded groups 200, got %d", seededGroups.Code)
	}
	var groups model.ListResponse[model.DeviceGroup]
	if err := json.Unmarshal(seededGroups.Body.Bytes(), &groups); err != nil {
		t.Fatalf("decode groups: %v", err)
	}
	if groups.Count < 1 || groups.Items[0].PolicyTemplateID != "ai-btech-student" {
		t.Fatalf("expected seeded device group: %+v", groups)
	}

	groupBody := []byte(`{
		"name": "Exam Mode Devices",
		"description": "Managed exam preparation laptops",
		"profile": "school-laptop",
		"device_ids": ["phase21-device-a", "phase21-device-b"],
		"policy_template_id": "school-laptop"
	}`)
	createGroup := httptest.NewRecorder()
	handler.ServeHTTP(createGroup, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeviceGroups, bytes.NewReader(groupBody)))
	if createGroup.Code != http.StatusCreated {
		t.Fatalf("expected group create 201, got %d: %s", createGroup.Code, createGroup.Body.String())
	}
	var group model.DeviceGroup
	if err := json.Unmarshal(createGroup.Body.Bytes(), &group); err != nil {
		t.Fatalf("decode created group: %v", err)
	}
	if group.ID == "" || group.PolicyTemplateID != "school-laptop" || len(group.DeviceIDs) != 2 {
		t.Fatalf("unexpected created group: %+v", group)
	}

	seededAssignments := httptest.NewRecorder()
	handler.ServeHTTP(seededAssignments, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentPolicyAssign, nil))
	if seededAssignments.Code != http.StatusOK {
		t.Fatalf("expected seeded assignments 200, got %d", seededAssignments.Code)
	}
	var assignments model.ListResponse[model.PolicyAssignment]
	if err := json.Unmarshal(seededAssignments.Body.Bytes(), &assignments); err != nil {
		t.Fatalf("decode assignments: %v", err)
	}
	if assignments.Count < 1 || assignments.Items[0].TargetType != constants.PolicyAssignmentTargetDeviceGroup {
		t.Fatalf("expected seeded policy assignment: %+v", assignments)
	}

	assignmentBody := []byte(`{
		"name": "Exam mode rollout",
		"target_type": "device_group",
		"target_id": "` + group.ID + `",
		"policy_template_id": "school-laptop",
		"alert_rule_ids": ["manual-rule-001"],
		"mode": "active"
	}`)
	createAssignment := httptest.NewRecorder()
	handler.ServeHTTP(createAssignment, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentPolicyAssign, bytes.NewReader(assignmentBody)))
	if createAssignment.Code != http.StatusCreated {
		t.Fatalf("expected assignment create 201, got %d: %s", createAssignment.Code, createAssignment.Body.String())
	}
	if !strings.Contains(createAssignment.Body.String(), "Exam mode rollout") || !strings.Contains(createAssignment.Body.String(), constants.PolicyAssignmentStatusActive) {
		t.Fatalf("expected created assignment body, got %s", createAssignment.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionDeviceGroupCreated) || !strings.Contains(audit.Body.String(), constants.AuditActionPolicyAssigned) {
		t.Fatalf("expected group and assignment audit events, got %s", audit.Body.String())
	}
}

func TestDataExportAndDeleteRequestEndpoints(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}

	exportBody := []byte(`{"format":"json","scope":"tenant"}`)
	createExport := httptest.NewRecorder()
	handler.ServeHTTP(createExport, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDataExports, bytes.NewReader(exportBody)))
	if createExport.Code != http.StatusCreated {
		t.Fatalf("expected export create 201, got %d: %s", createExport.Code, createExport.Body.String())
	}
	var export model.TenantDataExport
	if err := json.Unmarshal(createExport.Body.Bytes(), &export); err != nil {
		t.Fatalf("decode export: %v", err)
	}
	if export.Status != constants.DataExportStatusReady || export.ResourceCount == 0 || export.StorageKey == "" {
		t.Fatalf("unexpected data export: %+v", export)
	}

	deleteBody := []byte(`{"scope":"tenant","reason":"family account data cleanup"}`)
	createDelete := httptest.NewRecorder()
	handler.ServeHTTP(createDelete, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeleteRequests, bytes.NewReader(deleteBody)))
	if createDelete.Code != http.StatusCreated {
		t.Fatalf("expected delete request create 201, got %d: %s", createDelete.Code, createDelete.Body.String())
	}
	var deleteRequest model.DeleteRequest
	if err := json.Unmarshal(createDelete.Body.Bytes(), &deleteRequest); err != nil {
		t.Fatalf("decode delete request: %v", err)
	}
	if deleteRequest.Status != constants.DeleteRequestStatusQueued || deleteRequest.DueAt.IsZero() {
		t.Fatalf("unexpected delete request: %+v", deleteRequest)
	}

	exports := httptest.NewRecorder()
	handler.ServeHTTP(exports, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDataExports, nil))
	if exports.Code != http.StatusOK || !strings.Contains(exports.Body.String(), constants.DataExportStatusReady) {
		t.Fatalf("expected export list, got %d: %s", exports.Code, exports.Body.String())
	}

	deletes := httptest.NewRecorder()
	handler.ServeHTTP(deletes, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeleteRequests, nil))
	if deletes.Code != http.StatusOK || !strings.Contains(deletes.Body.String(), constants.DeleteRequestStatusQueued) {
		t.Fatalf("expected delete request list, got %d: %s", deletes.Code, deletes.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionDataExportCreated) || !strings.Contains(audit.Body.String(), constants.AuditActionDeleteRequestCreated) {
		t.Fatalf("expected data export and delete request audit events, got %s", audit.Body.String())
	}
}

func TestTenantValidation(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, constants.RouteTenants, strings.NewReader(`{"tenant_id":"family-varadha","name":"Family","plan_id":"unknown","retention_tier_id":"family_cloud_90_365_archive","primary_profile":"ai-btech-student"}`)))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %d", recorder.Code)
	}
	if !strings.Contains(recorder.Body.String(), "plan_id is unknown") {
		t.Fatalf("expected validation message, got %s", recorder.Body.String())
	}
}

func TestPolicyTemplatesAndArchiveStatus(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()

	templates := httptest.NewRecorder()
	handler.ServeHTTP(templates, httptest.NewRequest(http.MethodGet, constants.RoutePolicyTemplates, nil))
	if templates.Code != http.StatusOK {
		t.Fatalf("expected templates 200, got %d", templates.Code)
	}
	if !strings.Contains(templates.Body.String(), "AI BTech Student") {
		t.Fatalf("expected template catalog response, got %s", templates.Body.String())
	}

	archive := httptest.NewRecorder()
	handler.ServeHTTP(archive, httptest.NewRequest(http.MethodGet, constants.RouteArchiveStatus, nil))
	if archive.Code != http.StatusOK {
		t.Fatalf("expected archive status 200, got %d", archive.Code)
	}
	if !strings.Contains(archive.Body.String(), `"provider":"s3"`) {
		t.Fatalf("expected archive provider response, got %s", archive.Body.String())
	}
}

func TestSaaSReadinessCatalogs(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()

	for _, route := range []string{
		constants.RoutePlans,
		constants.RouteRoles,
		constants.RouteRetentionTiers,
		constants.RouteAuditEvents,
	} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, route, nil))
		if recorder.Code != http.StatusOK {
			t.Fatalf("expected %s 200, got %d", route, recorder.Code)
		}
		if !strings.Contains(recorder.Body.String(), `"count"`) {
			t.Fatalf("expected list response for %s, got %s", route, recorder.Body.String())
		}
	}
}

func TestLocalAddressValidation(t *testing.T) {
	t.Parallel()

	for _, addr := range []string{"127.0.0.1:18080", "localhost:18080", "[::1]:18080"} {
		if err := validateLocalAddress(addr); err != nil {
			t.Fatalf("expected %s to be allowed: %v", addr, err)
		}
	}

	if err := validateLocalAddress("0.0.0.0:18080"); err == nil {
		t.Fatal("expected non-local bind to be rejected")
	}
}
