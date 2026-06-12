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
