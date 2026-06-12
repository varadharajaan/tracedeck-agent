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
