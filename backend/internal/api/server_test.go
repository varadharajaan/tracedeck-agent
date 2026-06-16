package api

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
	if health.Header().Get(constants.HeaderCache) != constants.CacheNoStore {
		t.Fatalf("expected no-store health cache header, got %s", health.Header().Get(constants.HeaderCache))
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

func TestRuntimeStatusCenterEndpoint(t *testing.T) {
	t.Parallel()

	summaryPath := filepath.Join(t.TempDir(), "runtime-summary.json")
	summaryJSON := []byte(`{
		"generated_at": "2026-06-16T10:38:06.4949867+05:30",
		"base_url": "http://127.0.0.1:18080",
		"output_json": "data/local/output/runtime-summary.json",
		"output_text": "data/local/output/runtime-summary.txt",
		"backend": {
			"task_name": "\\TraceDeck\\TraceDeck Backend Dev",
			"task_present": true,
			"task_state": "Running",
			"scheduler_readback": "verified",
			"launch_task_verified": true,
			"runtime_ok": true,
			"runtime_evidence": "pid_and_health",
			"health_ok": true,
			"pid": 146776,
			"pid_running": true,
			"ready_file_present": true,
			"ready_at": "2026-06-15T23:24:54.8276547+05:30",
			"advisory": {
				"severity": "ok",
				"code": "scheduler_verified_runtime_ready",
				"headline": "Backend runtime and Scheduler readback are verified.",
				"operator_action": "No action needed.",
				"can_continue": true
			}
		},
		"doctor": {
			"skipped": false,
			"overall": "ok",
			"local": "ok",
			"report_json": "data/local/output/runtime-doctor.json"
		},
		"frontend": {
			"url_present": true,
			"url": "https://example.lambda-url.ap-south-1.on.aws"
		},
		"git": {
			"branch": "main",
			"head": "dcd106e",
			"tracked_content_diff": false,
			"tracked_content_diff_count": 0,
			"tracked_content_diff_rows": [],
			"status_rows": ["## main...origin/main"]
		},
		"logs": {
			"summary_log": "logs/local/ops/get-runtime-summary.log",
			"backend_stdout": "logs/local/backend/backend-task.out.log",
			"backend_stderr": "logs/local/backend/backend-task.err.log"
		},
		"verdict": {
			"can_continue": true,
			"severity": "ok",
			"headline": "Runtime proof is healthy.",
			"next_actions": ["No action needed."]
		},
		"privacy": {
			"metadata_only": true,
			"sensitive_collection": "denied"
		}
	}`)
	summaryJSON = append([]byte{0xEF, 0xBB, 0xBF}, summaryJSON...)
	if err := os.WriteFile(summaryPath, summaryJSON, 0o600); err != nil {
		t.Fatalf("write runtime summary fixture: %v", err)
	}

	server := NewServer(store.NewMemory(), slog.Default())
	server.runtimeSummaryPath = summaryPath
	handler := server.Handler()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteRuntimeStatus, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected runtime status center 200, got %d: %s", response.Code, response.Body.String())
	}
	if response.Header().Get(constants.HeaderCache) != constants.CacheNoStore {
		t.Fatalf("expected no-store runtime status cache header, got %s", response.Header().Get(constants.HeaderCache))
	}
	var center model.RuntimeStatusCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode runtime status center: %v", err)
	}
	if !center.SummaryAvailable || center.Source != constants.RuntimeStatusSourcePhase97Summary {
		t.Fatalf("expected available Phase 97 runtime summary source: %+v", center)
	}
	if center.Summary.Status != constants.StatusOK || !center.Summary.CanContinue || !center.Summary.RuntimeOK || !center.Summary.HealthOK {
		t.Fatalf("expected healthy runtime status summary: %+v", center.Summary)
	}
	if center.Summary.SchedulerReadback != "verified" || center.Summary.DoctorOverall != constants.StatusOK || center.Summary.TrackedContentDiff {
		t.Fatalf("expected verified Scheduler, ok doctor, and clean diff: %+v", center.Summary)
	}
	if len(center.Proof) < 6 || len(center.Actions) == 0 {
		t.Fatalf("expected proof and action rows: %+v", center)
	}
	hasScheduler := false
	hasPrivacy := false
	for _, proof := range center.Proof {
		if proof.EvidenceScope != constants.EvidenceScopeMetadataOnly {
			t.Fatalf("expected metadata-only runtime proof: %+v", proof)
		}
		if proof.ID == constants.RuntimeStatusProofSchedulerID && proof.Value == "verified" {
			hasScheduler = true
		}
		if proof.ID == constants.RuntimeStatusProofPrivacyID && proof.Status == constants.StatusOK {
			hasPrivacy = true
		}
	}
	if !hasScheduler || !hasPrivacy {
		t.Fatalf("expected Scheduler and privacy proof rows: %+v", center.Proof)
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no passwords") || !strings.Contains(center.PrivacyBoundary, "no screenshots") || !strings.Contains(center.PrivacyBoundary, "keylogging") {
		t.Fatalf("expected strict runtime privacy boundary, got %q", center.PrivacyBoundary)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("runtime status center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestRuntimeStatusCenterMissingSummary(t *testing.T) {
	t.Parallel()

	server := NewServer(store.NewMemory(), slog.Default())
	server.runtimeSummaryPath = filepath.Join(t.TempDir(), "missing-runtime-summary.json")
	handler := server.Handler()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteRuntimeStatus, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected missing runtime summary to return operator action 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.RuntimeStatusCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode missing runtime status center: %v", err)
	}
	if center.SummaryAvailable || center.Summary.Status != constants.StatusAttention || center.Summary.CanContinue {
		t.Fatalf("expected missing runtime summary action state: %+v", center.Summary)
	}
	if len(center.Actions) != 1 || center.Actions[0].Command != constants.RuntimeSummaryCommand {
		t.Fatalf("expected runtime summary generation action: %+v", center.Actions)
	}
	if len(center.Proof) != 1 || center.Proof[0].EvidenceScope != constants.EvidenceScopeMetadataOnly {
		t.Fatalf("expected one metadata-only missing-summary proof: %+v", center.Proof)
	}
}

func TestVerificationEvidenceCenterEndpoint(t *testing.T) {
	t.Parallel()

	evidencePath := filepath.Join(t.TempDir(), "verification-evidence.json")
	evidenceJSON := []byte(`{
		"generated_at": "2026-06-16T11:40:00+05:30",
		"phase": "phase99",
		"base_url": "http://127.0.0.1:18080",
		"branch": "main",
		"head": "3dcb31f",
		"overall_status": "ok",
		"can_promote": true,
		"gates": [
			{
				"id": "gofmt",
				"label": "Go format check",
				"command": "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1",
				"status": "ok",
				"severity": "info",
				"log_path": "logs/local/verify/check-gofmt.log",
				"report_path": "",
				"detail": "gofmt check passed",
				"completed_at": "2026-06-16T11:39:00+05:30",
				"evidence_scope": "metadata_only"
			},
			{
				"id": "newman",
				"label": "Newman collection",
				"command": "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase99.ps1",
				"status": "ok",
				"severity": "info",
				"log_path": "logs/local/newman/newman-phase99.log",
				"report_path": "data/local/newman/phase99/newman-report.json",
				"detail": "Newman assertions passed",
				"completed_at": "2026-06-16T11:39:30+05:30",
				"evidence_scope": "metadata_only"
			}
		],
		"artifacts": [
			{
				"id": "runtime-summary",
				"label": "Runtime summary JSON",
				"path": "data/local/output/runtime-summary.json",
				"status": "ok",
				"evidence_scope": "metadata_only"
			},
			{
				"id": "newman-report",
				"label": "Newman JSON report",
				"path": "data/local/newman/phase99/newman-report.json",
				"status": "ok",
				"evidence_scope": "metadata_only"
			}
		],
		"actions": [
			{
				"id": "refresh-evidence",
				"title": "Refresh verification evidence",
				"detail": "Regenerate the local evidence file after any new gate run.",
				"command": "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1",
				"severity": "info",
				"status": "ok"
			}
		],
		"privacy": {
			"metadata_only": true,
			"sensitive_collection": "denied",
			"forbidden_categories": ["credentials", "provider secrets", "screenshots"]
		},
		"privacy_boundary": "metadata-only verification evidence; no passwords; no screenshots"
	}`)
	evidenceJSON = append([]byte{0xEF, 0xBB, 0xBF}, evidenceJSON...)
	if err := os.WriteFile(evidencePath, evidenceJSON, 0o600); err != nil {
		t.Fatalf("write verification evidence fixture: %v", err)
	}

	server := NewServer(store.NewMemory(), slog.Default())
	server.verificationEvidencePath = evidencePath
	handler := server.Handler()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteVerificationCenter, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected verification evidence center 200, got %d: %s", response.Code, response.Body.String())
	}
	if response.Header().Get(constants.HeaderCache) != constants.CacheNoStore {
		t.Fatalf("expected no-store verification evidence cache header, got %s", response.Header().Get(constants.HeaderCache))
	}
	var center model.VerificationEvidenceCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode verification evidence center: %v", err)
	}
	if !center.EvidenceAvailable || center.Source != constants.VerificationEvidenceSourcePhase99 {
		t.Fatalf("expected available Phase 99 verification evidence source: %+v", center)
	}
	if center.Summary.Status != constants.StatusOK || !center.Summary.CanPromote || center.Summary.GatesOK != 2 || center.Summary.GatesAttention != 0 {
		t.Fatalf("expected healthy verification evidence summary: %+v", center.Summary)
	}
	if len(center.Gates) != 2 || len(center.Proof) < 4 || len(center.Actions) == 0 || len(center.Artifacts) != 2 {
		t.Fatalf("expected gate, proof, action, and artifact rows: %+v", center)
	}
	for _, gate := range center.Gates {
		if gate.EvidenceScope != constants.EvidenceScopeMetadataOnly {
			t.Fatalf("expected metadata-only gate evidence: %+v", gate)
		}
	}
	hasPrivacy := false
	for _, proof := range center.Proof {
		if proof.EvidenceScope != constants.EvidenceScopeMetadataOnly {
			t.Fatalf("expected metadata-only verification proof: %+v", proof)
		}
		if proof.ID == constants.VerificationEvidenceProofPrivacyID && proof.Status == constants.StatusOK {
			hasPrivacy = true
		}
	}
	if !hasPrivacy {
		t.Fatalf("expected verification privacy proof row: %+v", center.Proof)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password_value", "provider_secret_value", "push_endpoint_value", "screenshot_bytes_value", "raw_url_value", "page_title_value", "alert_body_value", "card_number", "cvv", "payment_token", "keylogger"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("verification evidence center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestVerificationEvidenceCenterMissingArtifact(t *testing.T) {
	t.Parallel()

	server := NewServer(store.NewMemory(), slog.Default())
	server.verificationEvidencePath = filepath.Join(t.TempDir(), "missing-verification-evidence.json")
	handler := server.Handler()

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteVerificationCenter, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected missing verification evidence to return operator action 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.VerificationEvidenceCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode missing verification evidence center: %v", err)
	}
	if center.EvidenceAvailable || center.Summary.Status != constants.StatusAttention || center.Summary.CanPromote {
		t.Fatalf("expected missing verification evidence action state: %+v", center.Summary)
	}
	if len(center.Actions) != 1 || center.Actions[0].Command != constants.VerificationEvidenceCommand {
		t.Fatalf("expected verification evidence generation action: %+v", center.Actions)
	}
	if len(center.Proof) != 1 || center.Proof[0].EvidenceScope != constants.EvidenceScopeMetadataOnly {
		t.Fatalf("expected one metadata-only missing-evidence proof: %+v", center.Proof)
	}
}

func TestDashboardLocalAuthPanel(t *testing.T) {
	t.Parallel()

	handler := NewServerWithAuth(store.NewMemory(), slog.Default(), AuthConfig{
		APIKey:   "local-secret",
		TenantID: "family-varadha",
		ActorID:  "dashboard-test",
		RoleID:   constants.RoleBusinessManager,
	}).Handler()

	dashboard := httptest.NewRecorder()
	handler.ServeHTTP(dashboard, httptest.NewRequest(http.MethodGet, constants.RouteDashboard, nil))
	if dashboard.Code != http.StatusOK {
		t.Fatalf("expected dashboard 200 without API key, got %d", dashboard.Code)
	}
	if dashboard.Header().Get(constants.HeaderCache) != constants.CacheNoStore {
		t.Fatalf("expected no-store dashboard cache header, got %s", dashboard.Header().Get(constants.HeaderCache))
	}
	body := dashboard.Body.String()
	for _, marker := range []string{
		"Local Dashboard Access",
		"X-TraceDeck-API-Key",
		"sessionStorage",
		"tracedeck.dashboard.apiKey",
	} {
		if !strings.Contains(body, marker) {
			t.Fatalf("expected dashboard auth marker %q, got %s", marker, body)
		}
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

func TestTenantSyncHealthEndpoint(t *testing.T) {
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
		"device_id": "sync-device-001",
		"host_name": "sync-host",
		"profile": "ai-btech-student",
		"os_name": "windows",
		"events": [
			{
				"id": "local-event-11",
				"type": "process.observed",
				"source": "collector.process",
				"observed_at": "2026-06-12T08:00:00Z",
				"app_name": "Code.exe",
				"path_hash": "hash-only",
				"metadata": { "category": "coding" }
			},
			{
				"id": "local-event-12",
				"type": "browser.summary",
				"source": "collector.browser.history",
				"observed_at": "2026-06-12T08:01:00Z",
				"metadata": { "domain": "youtube.com", "category": "study" }
			}
		]
	}`)
	ingest := httptest.NewRecorder()
	handler.ServeHTTP(ingest, httptest.NewRequest(http.MethodPost, constants.RouteDevices+"/sync-device-001/"+constants.RouteSegmentTelemetry, bytes.NewReader(body)))
	if ingest.Code != http.StatusAccepted {
		t.Fatalf("expected telemetry ingest 202, got %d: %s", ingest.Code, ingest.Body.String())
	}

	syncHealth := httptest.NewRecorder()
	handler.ServeHTTP(syncHealth, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentSyncHealth, nil))
	if syncHealth.Code != http.StatusOK {
		t.Fatalf("expected sync health 200, got %d: %s", syncHealth.Code, syncHealth.Body.String())
	}
	var summary model.TenantSyncHealth
	if err := json.Unmarshal(syncHealth.Body.Bytes(), &summary); err != nil {
		t.Fatalf("decode sync health: %v", err)
	}
	if summary.StoredEvents != 2 || summary.LastLocalEventID != 12 || summary.HostsReporting != 1 {
		t.Fatalf("unexpected tenant sync health: %+v", summary)
	}
	if len(summary.Devices) != 1 || summary.Devices[0].BrowserEvents != 1 || !summary.OfflineReplayReady {
		t.Fatalf("expected device sync and offline replay proof: %+v", summary)
	}
	if !strings.Contains(summary.PrivacyBoundary, "metadata-only") {
		t.Fatalf("expected metadata-only privacy boundary: %+v", summary)
	}
}

func TestTenantActivityFeedEndpoint(t *testing.T) {
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
		"device_id": "feed-device-001",
		"host_name": "feed-host",
		"profile": "ai-btech-student",
		"os_name": "windows",
		"events": [
			{
				"id": "local-event-41",
				"type": "process.observed",
				"source": "collector.process",
				"observed_at": "2026-06-12T08:00:00Z",
				"app_name": "Code.exe",
				"path_hash": "hash-only",
				"metadata": { "category": "coding" }
			}
		]
	}`)
	ingest := httptest.NewRecorder()
	handler.ServeHTTP(ingest, httptest.NewRequest(http.MethodPost, constants.RouteDevices+"/feed-device-001/"+constants.RouteSegmentTelemetry, bytes.NewReader(body)))
	if ingest.Code != http.StatusAccepted {
		t.Fatalf("expected telemetry ingest 202, got %d: %s", ingest.Code, ingest.Body.String())
	}

	feedResponse := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentActivityFeed+"?device_id=feed-device-001&kind=delivery&channel=email&limit=5", nil)
	handler.ServeHTTP(feedResponse, request)
	if feedResponse.Code != http.StatusOK {
		t.Fatalf("expected activity feed 200, got %d: %s", feedResponse.Code, feedResponse.Body.String())
	}
	var feed model.TenantActivityFeed
	if err := json.Unmarshal(feedResponse.Body.Bytes(), &feed); err != nil {
		t.Fatalf("decode activity feed: %v", err)
	}
	if feed.Filters.DeviceID != "feed-device-001" || feed.Filters.Channel != constants.DeliveryChannelEmail {
		t.Fatalf("expected normalized feed filters: %+v", feed.Filters)
	}
	if feed.Filters.IncludeDemo || feed.Summary.DeliveryItems != 0 || strings.Contains(feedResponse.Body.String(), constants.EvidenceSourceDemoSeed) {
		t.Fatalf("expected default email delivery feed to hide demo evidence: %+v", feed)
	}

	demoFeedResponse := httptest.NewRecorder()
	demoRequest := httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentActivityFeed+"?device_id=feed-device-001&kind=delivery&channel=email&limit=5&"+constants.QueryIncludeDemo+"="+constants.QueryValueTrue, nil)
	handler.ServeHTTP(demoFeedResponse, demoRequest)
	if demoFeedResponse.Code != http.StatusOK {
		t.Fatalf("expected demo activity feed 200, got %d: %s", demoFeedResponse.Code, demoFeedResponse.Body.String())
	}
	var demoFeed model.TenantActivityFeed
	if err := json.Unmarshal(demoFeedResponse.Body.Bytes(), &demoFeed); err != nil {
		t.Fatalf("decode demo activity feed: %v", err)
	}
	if !demoFeed.Filters.IncludeDemo || demoFeed.Summary.DeliveryItems == 0 || demoFeed.Items[0].Channel != constants.DeliveryChannelEmail {
		t.Fatalf("expected opt-in filtered email delivery items: %+v", demoFeed)
	}

	riskResponse := httptest.NewRecorder()
	handler.ServeHTTP(riskResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentActivityFeed+"?q=youtube&limit=3&"+constants.QueryIncludeDemo+"="+constants.QueryValueTrue, nil))
	if riskResponse.Code != http.StatusOK || !strings.Contains(riskResponse.Body.String(), "youtube") {
		t.Fatalf("expected query feed to include youtube risk item, got %d: %s", riskResponse.Code, riskResponse.Body.String())
	}
}

func TestTenantBrowserActivityEndpoint(t *testing.T) {
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
		"device_id": "browser-device-001",
		"host_name": "browser-host",
		"profile": "ai-btech-student",
		"os_name": "windows",
		"events": [
			{
				"id": "browser-event-001",
				"type": "browser.domain.observed",
				"source": "collector.browser.history",
				"observed_at": "2026-06-12T08:00:00Z",
				"app_name": "chrome",
				"metadata": {
					"browser_name": "chrome",
					"domain": "docs.python.org",
					"category": "study",
					"visit_count": "8",
					"youtube_study_match": "false",
					"url_mode": "domain_only",
					"stored_url_mode": "domain_only"
				}
			},
			{
				"id": "browser-event-002",
				"type": "browser.domain.observed",
				"source": "collector.browser.history",
				"observed_at": "2026-06-12T08:02:00Z",
				"app_name": "edge",
				"metadata": {
					"browser_name": "edge",
					"domain": "youtube.com",
					"category": "video-streaming",
					"visit_count": "3",
					"youtube_study_match": "false",
					"url_mode": "domain_only",
					"stored_url_mode": "domain_only"
				}
			},
			{
				"id": "browser-event-003",
				"type": "browser.domain.observed",
				"source": "collector.browser.history",
				"observed_at": "2026-06-12T08:04:00Z",
				"app_name": "brave",
				"metadata": {
					"browser_name": "brave",
					"domain": "github.com",
					"category": "study",
					"visit_count": "5",
					"youtube_study_match": "false",
					"url_mode": "domain_only",
					"stored_url_mode": "domain_only"
				}
			}
		]
	}`)
	ingest := httptest.NewRecorder()
	handler.ServeHTTP(ingest, httptest.NewRequest(http.MethodPost, constants.RouteDevices+"/browser-device-001/"+constants.RouteSegmentTelemetry, bytes.NewReader(body)))
	if ingest.Code != http.StatusAccepted {
		t.Fatalf("expected telemetry ingest 202, got %d: %s", ingest.Code, ingest.Body.String())
	}

	viewResponse := httptest.NewRecorder()
	viewPath := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentBrowserActivity + "?device_id=browser-device-001&limit=10"
	handler.ServeHTTP(viewResponse, httptest.NewRequest(http.MethodGet, viewPath, nil))
	if viewResponse.Code != http.StatusOK {
		t.Fatalf("expected browser activity 200, got %d: %s", viewResponse.Code, viewResponse.Body.String())
	}
	var viewer model.TenantBrowserActivityViewer
	if err := json.Unmarshal(viewResponse.Body.Bytes(), &viewer); err != nil {
		t.Fatalf("decode browser activity: %v", err)
	}
	if viewer.Filters.DeviceID != "browser-device-001" || viewer.Summary.Total != 3 || len(viewer.Items) != 3 {
		t.Fatalf("expected browser activity rows: %+v", viewer)
	}
	if viewer.Summary.Chrome != 1 || viewer.Summary.Edge != 1 || viewer.Summary.Brave != 1 {
		t.Fatalf("expected Chrome, Edge, and Brave counts: %+v", viewer.Summary)
	}
	if viewer.Summary.StudySafe != 2 || viewer.Summary.NonStudyYouTube != 1 || viewer.Summary.NotificationProof == 0 {
		t.Fatalf("expected study-safe, YouTube review, and notification proof: %+v", viewer.Summary)
	}
	if viewer.Items[0].Domain == "" || viewer.Items[0].Recommendation == "" || len(viewer.Hosts) != 1 || len(viewer.Browsers) != 3 {
		t.Fatalf("expected typed browser activity detail: %+v", viewer)
	}
	if viewer.Items[0].SourceKind != constants.EvidenceSourceLiveIngest || viewer.Items[0].EvidenceScope != constants.EvidenceScopeLive || viewer.Items[0].EvidenceDetail == "" {
		t.Fatalf("expected live telemetry provenance on browser activity: %+v", viewer.Items[0])
	}
	if !strings.Contains(viewer.PrivacyBoundary, "metadata-only") || !strings.Contains(viewer.PrivacyBoundary, "no passwords") {
		t.Fatalf("expected browser privacy boundary: %q", viewer.PrivacyBoundary)
	}
	serialized := strings.ToLower(viewResponse.Body.String())
	for _, forbidden := range []string{"raw_url", "page_title", "cookie_value", "token_value", "password_value", "screenshot_bytes"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("browser activity leaked forbidden marker %q: %s", forbidden, viewResponse.Body.String())
		}
	}

	filteredResponse := httptest.NewRecorder()
	filteredPath := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentBrowserActivity + "?browser=edge&study_safe=false&limit=5"
	handler.ServeHTTP(filteredResponse, httptest.NewRequest(http.MethodGet, filteredPath, nil))
	if filteredResponse.Code != http.StatusOK {
		t.Fatalf("expected filtered browser activity 200, got %d: %s", filteredResponse.Code, filteredResponse.Body.String())
	}
	var filtered model.TenantBrowserActivityViewer
	if err := json.Unmarshal(filteredResponse.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("decode filtered browser activity: %v", err)
	}
	if filtered.Filters.Browser != constants.BrowserNameEdge || filtered.Filters.StudySafe == nil || *filtered.Filters.StudySafe || len(filtered.Items) != 1 || filtered.Items[0].Browser != constants.BrowserNameEdge {
		t.Fatalf("expected filtered Edge non-study row: %+v", filtered)
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentBrowserActivity+"?browser=opera", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid browser filter 400, got %d: %s", invalid.Code, invalid.Body.String())
	}
}

func TestTenantDeliveryTimelineEndpoint(t *testing.T) {
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

	deviceBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "timeline-device-001",
		"host_name": "timeline-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	timelineResponse := httptest.NewRecorder()
	timelinePath := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentDeliveryTimeline + "?device_id=timeline-device-001&limit=8"
	handler.ServeHTTP(timelineResponse, httptest.NewRequest(http.MethodGet, timelinePath, nil))
	if timelineResponse.Code != http.StatusOK {
		t.Fatalf("expected delivery timeline 200, got %d: %s", timelineResponse.Code, timelineResponse.Body.String())
	}
	var timeline model.TenantDeliveryTimeline
	if err := json.Unmarshal(timelineResponse.Body.Bytes(), &timeline); err != nil {
		t.Fatalf("decode delivery timeline: %v", err)
	}
	if timeline.Filters.DeviceID != "timeline-device-001" || timeline.Summary.Total < 3 || len(timeline.Items) < 3 {
		t.Fatalf("expected delivery timeline host proof: %+v", timeline)
	}
	if timeline.Summary.Email == 0 || timeline.Summary.Push == 0 || timeline.Summary.Dashboard == 0 {
		t.Fatalf("expected email, push, and dashboard delivery evidence: %+v", timeline.Summary)
	}
	if timeline.Summary.NotificationScore == 0 || timeline.Summary.RecommendedPaidTier == "" {
		t.Fatalf("expected monetisable notification score and tier: %+v", timeline.Summary)
	}
	if !strings.Contains(timeline.PrivacyBoundary, "metadata-only") || !strings.Contains(timeline.PrivacyBoundary, "no passwords") {
		t.Fatalf("expected strict delivery timeline privacy boundary: %q", timeline.PrivacyBoundary)
	}
	serialized := strings.ToLower(timelineResponse.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("delivery timeline leaked forbidden marker %q: %s", forbidden, timelineResponse.Body.String())
		}
	}

	filteredResponse := httptest.NewRecorder()
	filteredPath := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentDeliveryTimeline + "?device_id=timeline-device-001&channel=email&status=delivered&provider=smtp&limit=2"
	handler.ServeHTTP(filteredResponse, httptest.NewRequest(http.MethodGet, filteredPath, nil))
	if filteredResponse.Code != http.StatusOK {
		t.Fatalf("expected filtered delivery timeline 200, got %d: %s", filteredResponse.Code, filteredResponse.Body.String())
	}
	var filtered model.TenantDeliveryTimeline
	if err := json.Unmarshal(filteredResponse.Body.Bytes(), &filtered); err != nil {
		t.Fatalf("decode filtered delivery timeline: %v", err)
	}
	if filtered.Filters.Channel != constants.DeliveryChannelEmail || filtered.Filters.Status != constants.DeliveryStatusDelivered || filtered.Filters.Provider != constants.DeliveryProviderSMTP {
		t.Fatalf("expected normalized delivery filters: %+v", filtered.Filters)
	}
	if len(filtered.Items) == 0 || len(filtered.Items) > 2 || filtered.Items[0].Channel != constants.DeliveryChannelEmail || filtered.Items[0].Status != constants.DeliveryStatusDelivered {
		t.Fatalf("expected delivered email timeline items: %+v", filtered)
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryTimeline+"?status=open", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid delivery timeline status 400, got %d: %s", invalid.Code, invalid.Body.String())
	}
}

func TestTenantDeliveryAssuranceEndpoint(t *testing.T) {
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

	deviceBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "assurance-device-001",
		"host_name": "assurance-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	path := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentDeliveryAssure + "?device_id=assurance-device-001&limit=8"
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, path, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected delivery assurance 200, got %d: %s", response.Code, response.Body.String())
	}
	var assurance model.TenantDeliveryAssurance
	if err := json.Unmarshal(response.Body.Bytes(), &assurance); err != nil {
		t.Fatalf("decode delivery assurance: %v", err)
	}
	if assurance.Summary.RoutesTotal != 3 || assurance.Summary.ProviderConfirmed != 0 || assurance.Summary.DemoOnly == 0 || assurance.Summary.Retrying == 0 {
		t.Fatalf("expected route truth summary with demo-only and retrying proof: %+v", assurance.Summary)
	}
	if assurance.Summary.EmailProviderReady || assurance.Summary.PushProviderReady || assurance.Summary.BuyerReady {
		t.Fatalf("seeded demo delivery must not be marked provider ready: %+v", assurance.Summary)
	}
	if len(assurance.Routes) != 3 || len(assurance.Events) < 3 {
		t.Fatalf("expected route and event assurance rows: %+v", assurance)
	}
	if !strings.Contains(assurance.PrivacyBoundary, "metadata-only") || !strings.Contains(assurance.PrivacyBoundary, "no provider secrets") {
		t.Fatalf("expected strict delivery assurance privacy boundary: %q", assurance.PrivacyBoundary)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("delivery assurance leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}

	filtered := httptest.NewRecorder()
	filteredPath := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentDeliveryAssure + "?device_id=assurance-device-001&channel=email&assurance_state=demo_only"
	handler.ServeHTTP(filtered, httptest.NewRequest(http.MethodGet, filteredPath, nil))
	if filtered.Code != http.StatusOK {
		t.Fatalf("expected filtered delivery assurance 200, got %d: %s", filtered.Code, filtered.Body.String())
	}
	var emailDemo model.TenantDeliveryAssurance
	if err := json.Unmarshal(filtered.Body.Bytes(), &emailDemo); err != nil {
		t.Fatalf("decode filtered delivery assurance: %v", err)
	}
	if len(emailDemo.Routes) != 1 || emailDemo.Routes[0].Channel != constants.DeliveryChannelEmail || emailDemo.Routes[0].AssuranceState != constants.DeliveryAssuranceDemoOnly {
		t.Fatalf("expected filtered email demo-only route: %+v", emailDemo.Routes)
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryAssure+"?assurance_state=random", nil))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid delivery assurance state 400, got %d: %s", invalid.Code, invalid.Body.String())
	}
}

func TestTenantActivityViewsEndpoint(t *testing.T) {
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

	seededViews := httptest.NewRecorder()
	handler.ServeHTTP(seededViews, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentActivityViews, nil))
	if seededViews.Code != http.StatusOK {
		t.Fatalf("expected seeded activity views 200, got %d: %s", seededViews.Code, seededViews.Body.String())
	}
	var views model.ListResponse[model.TenantActivityView]
	if err := json.Unmarshal(seededViews.Body.Bytes(), &views); err != nil {
		t.Fatalf("decode activity views: %v", err)
	}
	if views.Count != 4 || views.Items[0].ID != constants.ActivityViewHighRiskOpen {
		t.Fatalf("expected seeded monetisation command views: %+v", views)
	}
	if views.Items[1].Filter.Channel != constants.DeliveryChannelEmail || views.Items[2].Filter.Channel != constants.DeliveryChannelPush {
		t.Fatalf("expected email and push saved filters: %+v", views.Items)
	}

	viewBody := []byte(`{
		"name": "Dashboard delivery misses",
		"description": "Watch dashboard delivery gaps before a paid demo",
		"paid_tier": "business",
		"sort_order": 9,
		"filter": {
			"kind": "delivery",
			"channel": "dashboard",
			"status": "failed",
			"limit": 10
		}
	}`)
	createView := httptest.NewRecorder()
	handler.ServeHTTP(createView, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentActivityViews, bytes.NewReader(viewBody)))
	if createView.Code != http.StatusCreated {
		t.Fatalf("expected activity view create 201, got %d: %s", createView.Code, createView.Body.String())
	}
	var created model.TenantActivityView
	if err := json.Unmarshal(createView.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode created activity view: %v", err)
	}
	if created.ID == "" || created.Filter.Channel != constants.DeliveryChannelDashboard || created.PaidTier != constants.PlanBusiness {
		t.Fatalf("unexpected created activity view: %+v", created)
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentActivityViews, strings.NewReader(`{"name":"bad","filter":{"kind":"random"}}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid activity view 400, got %d: %s", invalid.Code, invalid.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionActivityViewCreated) {
		t.Fatalf("expected activity view audit event, got %s", audit.Body.String())
	}
}

func TestHostDashboardRiskEndpoints(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	_, err := repo.CreateTenant(context.Background(), model.CreateTenantRequest{
		TenantID:        "family-varadha",
		Name:            "Family Varadha",
		PlanID:          constants.PlanFamilyPro,
		RetentionTierID: constants.RetentionFamilyCloud,
		PrimaryProfile:  "ai-btech-student",
	})
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	handler := NewServer(repo, slog.Default()).Handler()
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
	if host.Device.DeviceID != "laptop-cousin-001" {
		t.Fatalf("unexpected host overview: %+v", host)
	}
	if host.Summary.PolicyViolations != 0 || host.Summary.AlertsRaised != 0 || host.RiskScore != constants.RiskScoreNone {
		t.Fatalf("expected default overview to hide demo evidence: %+v", host)
	}
	if len(host.PolicyViolations) != 0 || len(host.Anomalies) != 0 || len(host.TamperEvents) != 0 || len(host.AlertDeliveries) != 0 {
		t.Fatalf("expected default overview lists to hide demo evidence: %+v", host)
	}
	if strings.Contains(overview.Body.String(), constants.DemoRiskMediaAppName) || strings.Contains(overview.Body.String(), constants.EvidenceSourceDemoSeed) {
		t.Fatalf("default overview leaked demo evidence: %s", overview.Body.String())
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
	if policyEvents.Count != 0 || len(policyEvents.Items) != 0 {
		t.Fatalf("expected default policy endpoint to hide demo evidence: %+v", policyEvents)
	}
	if strings.Contains(policy.Body.String(), constants.DemoRiskMediaAppName) || strings.Contains(policy.Body.String(), constants.EvidenceSourceDemoSeed) {
		t.Fatalf("default policy endpoint leaked demo evidence: %s", policy.Body.String())
	}

	activity := httptest.NewRecorder()
	activityPath := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentActivityFeed + "?device_id=laptop-cousin-001&limit=10"
	handler.ServeHTTP(activity, httptest.NewRequest(http.MethodGet, activityPath, nil))
	if activity.Code != http.StatusOK {
		t.Fatalf("expected tenant activity feed 200, got %d: %s", activity.Code, activity.Body.String())
	}
	var activityFeed model.TenantActivityFeed
	if err := json.Unmarshal(activity.Body.Bytes(), &activityFeed); err != nil {
		t.Fatalf("decode tenant activity feed: %v", err)
	}
	if activityFeed.Summary.RiskItems != 0 || activityFeed.Summary.DeliveryItems != 0 || activityFeed.Filters.IncludeDemo {
		t.Fatalf("expected default activity feed to hide demo evidence: %+v", activityFeed)
	}
	if strings.Contains(activity.Body.String(), constants.DemoRiskMediaAppName) || strings.Contains(activity.Body.String(), constants.EvidenceSourceDemoSeed) {
		t.Fatalf("default activity feed leaked demo evidence: %s", activity.Body.String())
	}

	demoPolicy := httptest.NewRecorder()
	demoPolicyPath := constants.RouteDevices + "/laptop-cousin-001/" + constants.RouteSegmentPolicyEvents + "?" + constants.QueryIncludeDemo + "=" + constants.QueryValueTrue
	handler.ServeHTTP(demoPolicy, httptest.NewRequest(http.MethodGet, demoPolicyPath, nil))
	if demoPolicy.Code != http.StatusOK {
		t.Fatalf("expected demo policy violations 200, got %d", demoPolicy.Code)
	}
	var demoPolicyEvents model.ListResponse[model.RiskEvent]
	if err := json.Unmarshal(demoPolicy.Body.Bytes(), &demoPolicyEvents); err != nil {
		t.Fatalf("decode demo policy violations: %v", err)
	}
	if demoPolicyEvents.Count == 0 || demoPolicyEvents.Items[0].Type != constants.RiskTypePolicyViolation {
		t.Fatalf("unexpected demo policy events: %+v", demoPolicyEvents)
	}
	if demoPolicyEvents.Items[0].SourceKind != constants.EvidenceSourceDemoSeed || demoPolicyEvents.Items[0].EvidenceScope != constants.EvidenceScopeDemo || demoPolicyEvents.Items[0].AppName != constants.DemoRiskMediaAppName {
		t.Fatalf("expected opt-in seeded risk provenance: %+v", demoPolicyEvents.Items[0])
	}

	demoActivity := httptest.NewRecorder()
	demoActivityPath := activityPath + "&" + constants.QueryIncludeDemo + "=" + constants.QueryValueTrue
	handler.ServeHTTP(demoActivity, httptest.NewRequest(http.MethodGet, demoActivityPath, nil))
	if demoActivity.Code != http.StatusOK {
		t.Fatalf("expected demo tenant activity feed 200, got %d: %s", demoActivity.Code, demoActivity.Body.String())
	}
	var demoActivityFeed model.TenantActivityFeed
	if err := json.Unmarshal(demoActivity.Body.Bytes(), &demoActivityFeed); err != nil {
		t.Fatalf("decode demo tenant activity feed: %v", err)
	}
	if !demoActivityFeed.Filters.IncludeDemo || !strings.Contains(demoActivity.Body.String(), constants.DemoRiskMediaAppName) {
		t.Fatalf("expected opt-in activity feed to expose demo evidence: %s", demoActivity.Body.String())
	}

	deliveries := httptest.NewRecorder()
	handler.ServeHTTP(deliveries, httptest.NewRequest(http.MethodGet, constants.RouteDevices+"/laptop-cousin-001/"+constants.RouteSegmentAlertDelivery, nil))
	if deliveries.Code != http.StatusOK {
		t.Fatalf("expected alert deliveries 200, got %d", deliveries.Code)
	}
	var deliveryEvents model.ListResponse[model.AlertDelivery]
	if err := json.Unmarshal(deliveries.Body.Bytes(), &deliveryEvents); err != nil {
		t.Fatalf("decode alert deliveries: %v", err)
	}
	if deliveryEvents.Count != 0 || len(deliveryEvents.Items) != 0 {
		t.Fatalf("expected default alert deliveries to hide demo evidence: %+v", deliveryEvents)
	}
	if strings.Contains(deliveries.Body.String(), constants.EvidenceSourceDemoSeed) {
		t.Fatalf("default deliveries leaked demo evidence: %s", deliveries.Body.String())
	}

	demoDeliveries := httptest.NewRecorder()
	demoDeliveriesPath := constants.RouteDevices + "/laptop-cousin-001/" + constants.RouteSegmentAlertDelivery + "?" + constants.QueryIncludeDemo + "=" + constants.QueryValueTrue
	handler.ServeHTTP(demoDeliveries, httptest.NewRequest(http.MethodGet, demoDeliveriesPath, nil))
	if demoDeliveries.Code != http.StatusOK {
		t.Fatalf("expected demo alert deliveries 200, got %d", demoDeliveries.Code)
	}
	var demoDeliveryEvents model.ListResponse[model.AlertDelivery]
	if err := json.Unmarshal(demoDeliveries.Body.Bytes(), &demoDeliveryEvents); err != nil {
		t.Fatalf("decode demo alert deliveries: %v", err)
	}
	if demoDeliveryEvents.Count == 0 || demoDeliveryEvents.Items[0].SourceKind != constants.EvidenceSourceDemoSeed || demoDeliveryEvents.Items[0].EvidenceScope != constants.EvidenceScopeDeliveryProof {
		t.Fatalf("expected opt-in seeded delivery provenance: %+v", demoDeliveryEvents)
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
	if !report.Generated || report.EmailReady || !report.PDFReady {
		t.Fatalf("expected generated weekly report without fake email proof: %+v", report)
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

func TestNotificationPreferenceEndpoints(t *testing.T) {
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

	preferenceURL := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentNotificationPref
	initialResponse := httptest.NewRecorder()
	handler.ServeHTTP(initialResponse, httptest.NewRequest(http.MethodGet, preferenceURL, nil))
	if initialResponse.Code != http.StatusOK {
		t.Fatalf("expected notification preferences 200, got %d: %s", initialResponse.Code, initialResponse.Body.String())
	}
	var initial model.NotificationPreferenceCenter
	if err := json.Unmarshal(initialResponse.Body.Bytes(), &initial); err != nil {
		t.Fatalf("decode notification preferences: %v", err)
	}
	if initial.Summary.RulesTotal < 3 || !initial.Summary.EmailEnabled || !initial.Summary.PushEnabled || !initial.Summary.DashboardEnabled {
		t.Fatalf("expected seeded preference channel coverage: %+v", initial.Summary)
	}
	if initial.Summary.StudySuppressionRules == 0 || !initial.QuietHours.Enabled || !initial.Escalation.Enabled {
		t.Fatalf("expected study suppression, quiet hours, and escalation: %+v", initial)
	}
	if !strings.Contains(initial.PrivacyBoundary, "no passwords") || !strings.Contains(initial.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict privacy boundary, got %q", initial.PrivacyBoundary)
	}

	updateBody := []byte(`{
		"digest_cadence": "daily",
		"quiet_hours": {
			"enabled": true,
			"start_local": "21:30",
			"end_local": "06:00",
			"timezone": "Asia/Calcutta"
		},
		"escalation": {
			"enabled": true,
			"after_minutes": 10,
			"repeat_every_minutes": 20,
			"max_repeats": 3,
			"channels": ["email", "push"],
			"owner": "parent escalation"
		},
		"rules": [
			{
				"name": "High-risk software immediate alert",
				"event_type": "risky_software",
				"severity": "high",
				"channels": ["email", "push", "dashboard"],
				"mode": "immediate",
				"recipient_group": "parent escalation",
				"quiet_hours_bypass": true,
				"paid_tier": "family_pro",
				"delivery_sla": "10 minutes",
				"next_action": "Verify delivery proof before relying on this rule.",
				"retention_evidence": "metadata-only alert and delivery proof"
			},
			{
				"name": "Study-safe digest",
				"event_type": "non_study_youtube",
				"severity": "low",
				"channels": ["dashboard"],
				"mode": "silent",
				"recipient_group": "dashboard archive",
				"suppression_label": "study topics suppressed",
				"study_safe": true,
				"paid_tier": "free",
				"delivery_sla": "dashboard only",
				"next_action": "Keep study-safe learning out of noisy alert paths.",
				"retention_evidence": "category metadata only"
			}
		]
	}`)
	updateResponse := httptest.NewRecorder()
	handler.ServeHTTP(updateResponse, httptest.NewRequest(http.MethodPost, preferenceURL, bytes.NewReader(updateBody)))
	if updateResponse.Code != http.StatusOK {
		t.Fatalf("expected preference update 200, got %d: %s", updateResponse.Code, updateResponse.Body.String())
	}
	var updated model.NotificationPreferenceCenter
	if err := json.Unmarshal(updateResponse.Body.Bytes(), &updated); err != nil {
		t.Fatalf("decode updated preferences: %v", err)
	}
	if updated.DigestCadence != constants.NotificationDigestCadenceDaily || updated.QuietHours.StartLocal != "21:30" {
		t.Fatalf("expected updated cadence and quiet hours: %+v", updated)
	}
	if updated.Summary.RulesTotal != 2 || updated.Summary.ImmediateRules != 1 || updated.Summary.SilentRules != 1 {
		t.Fatalf("expected updated rule counts: %+v", updated.Summary)
	}

	invalidResponse := httptest.NewRecorder()
	handler.ServeHTTP(invalidResponse, httptest.NewRequest(http.MethodPost, preferenceURL, strings.NewReader(`{"digest_cadence":"hourly"}`)))
	if invalidResponse.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid preference 400, got %d: %s", invalidResponse.Code, invalidResponse.Body.String())
	}

	auditResponse := httptest.NewRecorder()
	handler.ServeHTTP(auditResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if auditResponse.Code != http.StatusOK {
		t.Fatalf("expected audit events 200, got %d", auditResponse.Code)
	}
	if !strings.Contains(auditResponse.Body.String(), constants.AuditActionNotificationPref) {
		t.Fatalf("expected notification preference audit event, got %s", auditResponse.Body.String())
	}
}

func TestTenantDeliveryDrilldownEndpoint(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)
	deviceBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "demo-study-laptop",
		"host_name": "Demo Study Laptop",
		"profile": "ai-btech-student",
		"os_name": "Windows 11"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	drilldownResponse := httptest.NewRecorder()
	handler.ServeHTTP(drilldownResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryDrill, nil))
	if drilldownResponse.Code != http.StatusOK {
		t.Fatalf("expected drilldown 200, got %d: %s", drilldownResponse.Code, drilldownResponse.Body.String())
	}
	var drilldown model.TenantDeliveryDrilldown
	if err := json.Unmarshal(drilldownResponse.Body.Bytes(), &drilldown); err != nil {
		t.Fatalf("decode delivery drilldown: %v", err)
	}
	if drilldown.Summary.RoutesTotal != 3 || len(drilldown.Routes) != 3 {
		t.Fatalf("expected three drilldown routes: %+v", drilldown)
	}
	if !strings.Contains(drilldown.PrivacyBoundary, "no provider secrets") {
		t.Fatalf("expected privacy boundary to deny provider secrets: %q", drilldown.PrivacyBoundary)
	}
	if drilldown.Routes[0].Evidence == "" || strings.Contains(strings.ToLower(drilldownResponse.Body.String()), "smtp_password") {
		t.Fatalf("expected content-safe drilldown evidence, got %s", drilldownResponse.Body.String())
	}

	runBody := []byte(`{"mode":"dry_run","channel":"push","reason":"paid demo rehearsal"}`)
	runResponse := httptest.NewRecorder()
	handler.ServeHTTP(runResponse, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryDrill, bytes.NewReader(runBody)))
	if runResponse.Code != http.StatusAccepted {
		t.Fatalf("expected drilldown run 202, got %d: %s", runResponse.Code, runResponse.Body.String())
	}
	var rehearsed model.TenantDeliveryDrilldown
	if err := json.Unmarshal(runResponse.Body.Bytes(), &rehearsed); err != nil {
		t.Fatalf("decode rehearsed delivery drilldown: %v", err)
	}
	if !rehearsed.Summary.PushReady || rehearsed.Summary.LastRehearsedAt == nil {
		t.Fatalf("expected push route rehearsal proof: %+v", rehearsed.Summary)
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryDrill, strings.NewReader(`{"mode":"send_live"}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid drilldown mode 400, got %d: %s", invalid.Code, invalid.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionDeliveryDrillRun) {
		t.Fatalf("expected delivery drilldown audit event, got %s", audit.Body.String())
	}
}

func TestTenantDeliveryRemediationEndpoint(t *testing.T) {
	t.Parallel()

	handler := NewServer(store.NewMemory(), slog.Default()).Handler()
	tenantBody := []byte(`{
		"tenant_id": "family-varadha",
		"name": "Family Varadha",
		"plan_id": "family_pro",
		"retention_tier_id": "family_cloud_90_365_archive",
		"primary_profile": "ai-btech-student"
	}`)
	deviceBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "demo-study-laptop",
		"host_name": "Demo Study Laptop",
		"profile": "ai-btech-student",
		"os_name": "Windows 11"
	}`)

	createTenant := httptest.NewRecorder()
	handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(tenantBody)))
	if createTenant.Code != http.StatusCreated {
		t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
	}
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	remediationResponse := httptest.NewRecorder()
	handler.ServeHTTP(remediationResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryRemedy, nil))
	if remediationResponse.Code != http.StatusOK {
		t.Fatalf("expected remediation 200, got %d: %s", remediationResponse.Code, remediationResponse.Body.String())
	}
	var remediation model.TenantDeliveryRemediation
	if err := json.Unmarshal(remediationResponse.Body.Bytes(), &remediation); err != nil {
		t.Fatalf("decode delivery remediation: %v", err)
	}
	if remediation.Summary.RoutesTotal != 3 || len(remediation.Actions) != 3 {
		t.Fatalf("expected three remediation routes: %+v", remediation)
	}
	if !strings.Contains(remediation.PrivacyBoundary, "without live provider sends") {
		t.Fatalf("expected provider-safe privacy boundary: %q", remediation.PrivacyBoundary)
	}
	if strings.Contains(strings.ToLower(remediationResponse.Body.String()), "smtp_password") {
		t.Fatalf("expected remediation payload without provider secrets: %s", remediationResponse.Body.String())
	}

	runBody := []byte(`{"mode":"dry_run","channel":"push","action":"retry_plan","reason":"plan route recovery for paid demo","owner":"parent mobile push subscription"}`)
	runResponse := httptest.NewRecorder()
	handler.ServeHTTP(runResponse, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryRemedy, bytes.NewReader(runBody)))
	if runResponse.Code != http.StatusAccepted {
		t.Fatalf("expected remediation run 202, got %d: %s", runResponse.Code, runResponse.Body.String())
	}
	var planned model.TenantDeliveryRemediation
	if err := json.Unmarshal(runResponse.Body.Bytes(), &planned); err != nil {
		t.Fatalf("decode planned delivery remediation: %v", err)
	}
	if planned.Summary.PlannedActions < 1 || len(planned.RecentPlans) < 1 {
		t.Fatalf("expected planned remediation action: %+v", planned.Summary)
	}
	if planned.RecentPlans[0].Action != constants.DeliveryRemediationActionRetryPlan || planned.RecentPlans[0].Channel != constants.DeliveryChannelPush {
		t.Fatalf("expected push retry plan, got %+v", planned.RecentPlans[0])
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentDeliveryRemedy, strings.NewReader(`{"mode":"send_live","action":"retry_plan"}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid remediation mode 400, got %d: %s", invalid.Code, invalid.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionDeliveryRemediation) {
		t.Fatalf("expected delivery remediation audit event, got %s", audit.Body.String())
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

func TestTenantAlertInboxEndpoint(t *testing.T) {
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
		"device_id": "inbox-device-001",
		"host_name": "inbox-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	inboxResponse := httptest.NewRecorder()
	handler.ServeHTTP(inboxResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAlertInbox, nil))
	if inboxResponse.Code != http.StatusOK {
		t.Fatalf("expected alert inbox 200, got %d: %s", inboxResponse.Code, inboxResponse.Body.String())
	}
	var inbox model.TenantAlertInbox
	if err := json.Unmarshal(inboxResponse.Body.Bytes(), &inbox); err != nil {
		t.Fatalf("decode alert inbox: %v", err)
	}
	if inbox.Summary.Total == 0 || inbox.Summary.WithEmail == 0 || inbox.Summary.WithPush == 0 || inbox.Summary.WithDashboard == 0 {
		t.Fatalf("expected event-linked channel proof in alert inbox: %+v", inbox.Summary)
	}
	if len(inbox.Items) == 0 || inbox.Items[0].EventID == "" || inbox.Items[0].NextAction == "" {
		t.Fatalf("expected actionable alert inbox items: %+v", inbox.Items)
	}
	if !strings.Contains(inbox.PrivacyBoundary, "no passwords") || !strings.Contains(inbox.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected privacy boundary on alert inbox: %q", inbox.PrivacyBoundary)
	}
}

func TestTenantNotificationCommandCenterEndpoint(t *testing.T) {
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
		"device_id": "notification-command-device-001",
		"host_name": "notification-command-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	commandResponse := httptest.NewRecorder()
	handler.ServeHTTP(commandResponse, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentNotificationCmd, nil))
	if commandResponse.Code != http.StatusOK {
		t.Fatalf("expected notification command center 200, got %d: %s", commandResponse.Code, commandResponse.Body.String())
	}
	var commandCenter model.TenantNotificationCommandCenter
	if err := json.Unmarshal(commandResponse.Body.Bytes(), &commandCenter); err != nil {
		t.Fatalf("decode notification command center: %v", err)
	}
	if commandCenter.Summary.OpenAlerts == 0 || commandCenter.Summary.NotificationScore == 0 {
		t.Fatalf("expected alert and notification summary proof: %+v", commandCenter.Summary)
	}
	if len(commandCenter.Channels) < 3 || len(commandCenter.Alerts) == 0 || len(commandCenter.Actions) == 0 {
		t.Fatalf("expected channels, alerts, and actions: %+v", commandCenter)
	}
	if commandCenter.Channels[0].PaidTier == "" || commandCenter.Alerts[0].EmailStatus == "" || commandCenter.Alerts[0].PushStatus == "" {
		t.Fatalf("expected channel and alert delivery proof: %+v %+v", commandCenter.Channels, commandCenter.Alerts)
	}
	if !strings.Contains(commandCenter.PrivacyBoundary, "no passwords") || !strings.Contains(commandCenter.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict privacy boundary, got %q", commandCenter.PrivacyBoundary)
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

func TestTenantBusinessDashboardEndpoint(t *testing.T) {
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
		"device_id": "business-dashboard-device-001",
		"host_name": "business-dashboard-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentBusinessDash, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected business dashboard 200, got %d: %s", response.Code, response.Body.String())
	}
	var dashboard model.TenantBusinessDashboard
	if err := json.Unmarshal(response.Body.Bytes(), &dashboard); err != nil {
		t.Fatalf("decode business dashboard: %v", err)
	}
	if dashboard.Summary.ProductScore == 0 || dashboard.Summary.NotificationScore == 0 || dashboard.Summary.RecommendedPackage == "" {
		t.Fatalf("expected scored business dashboard summary: %+v", dashboard.Summary)
	}
	if dashboard.Summary.MailDelivered == 0 || dashboard.Summary.DashboardDelivered == 0 {
		t.Fatalf("expected mail and dashboard delivery proof: %+v", dashboard.Summary)
	}
	if len(dashboard.Metrics) < 8 || len(dashboard.Alerts) == 0 || len(dashboard.Channels) < 3 || len(dashboard.Packages) < 3 || len(dashboard.Actions) == 0 {
		t.Fatalf("expected monetisable dashboard surfaces: %+v", dashboard)
	}
	hasPush := false
	for _, channel := range dashboard.Channels {
		if channel.Channel == constants.DeliveryChannelPush && channel.Status != "" {
			hasPush = true
		}
	}
	if !hasPush {
		t.Fatalf("expected push notification route proof: %+v", dashboard.Channels)
	}
	if dashboard.Packages[0].Tier == "" || dashboard.Channels[0].PaidTier == "" || dashboard.Actions[0].Source == "" {
		t.Fatalf("expected typed package, channel, and action metadata: %+v %+v %+v", dashboard.Packages, dashboard.Channels, dashboard.Actions)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("business dashboard leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(dashboard.PrivacyBoundary, "no passwords") || !strings.Contains(dashboard.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict business dashboard privacy boundary, got %q", dashboard.PrivacyBoundary)
	}
}

func TestTenantRoleExperiencesEndpoint(t *testing.T) {
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
		"device_id": "role-experience-device-001",
		"host_name": "role-experience-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentRoleExperience, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected role experiences 200, got %d: %s", response.Code, response.Body.String())
	}
	var experience model.TenantRoleExperience
	if err := json.Unmarshal(response.Body.Bytes(), &experience); err != nil {
		t.Fatalf("decode role experiences: %v", err)
	}
	if experience.Summary.ReadinessScore == 0 || experience.Summary.RolesTotal != 4 || len(experience.Roles) != 4 || len(experience.Onboarding) < 4 {
		t.Fatalf("expected four scored role experiences and onboarding items: %+v", experience)
	}
	roleIDs := map[string]bool{}
	for _, role := range experience.Roles {
		roleIDs[role.RoleID] = true
		if role.ReadinessScore == 0 || role.PaidTier == "" || role.NextAction == "" || len(role.VisiblePanels) == 0 || len(role.Metrics) == 0 {
			t.Fatalf("expected typed role metadata for %+v", role)
		}
	}
	for _, roleID := range []string{constants.RoleParent, constants.RoleStudent, constants.RoleSchoolAdmin, constants.RoleBusinessManager} {
		if !roleIDs[roleID] {
			t.Fatalf("expected role %q in role experiences: %+v", roleID, experience.Roles)
		}
	}
	if !strings.Contains(experience.PrivacyBoundary, "metadata-only") || !strings.Contains(experience.PrivacyBoundary, "no passwords") {
		t.Fatalf("expected strict role experience privacy boundary: %q", experience.PrivacyBoundary)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("role experiences leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestTenantExecutiveConsoleEndpoint(t *testing.T) {
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
		"device_id": "executive-console-device-001",
		"host_name": "executive-console-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentExecutiveConsole, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected executive console 200, got %d: %s", response.Code, response.Body.String())
	}
	var console model.TenantExecutiveConsole
	if err := json.Unmarshal(response.Body.Bytes(), &console); err != nil {
		t.Fatalf("decode executive console: %v", err)
	}
	if console.Summary.ReadinessScore == 0 || console.Summary.NotificationScore == 0 || console.Summary.RecommendedPaidPackage == "" || console.Summary.NextBestAction == "" {
		t.Fatalf("expected monetisable executive summary: %+v", console.Summary)
	}
	if console.Summary.EmailDelivered == 0 || console.Summary.DashboardDelivered == 0 {
		t.Fatalf("expected mail and dashboard delivery proof: %+v", console.Summary)
	}
	if len(console.Tiles) < 8 || len(console.Alerts) == 0 || len(console.Deliveries) < 3 || len(console.Actions) == 0 {
		t.Fatalf("expected tiles, alerts, deliveries, and actions: %+v", console)
	}
	hasPush := false
	for _, delivery := range console.Deliveries {
		if delivery.Channel == constants.DeliveryChannelPush && delivery.Status != "" && delivery.PaidTier != "" {
			hasPush = true
		}
	}
	if !hasPush {
		t.Fatalf("expected push notification proof in executive console: %+v", console.Deliveries)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("executive console leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(console.PrivacyBoundary, "metadata-only") || !strings.Contains(console.PrivacyBoundary, "no passwords") || !strings.Contains(console.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict executive console privacy boundary, got %q", console.PrivacyBoundary)
	}
}

func TestTenantCustomerControlRoomEndpoint(t *testing.T) {
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
		"device_id": "customer-control-device-001",
		"host_name": "customer-control-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentCustomerControl, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected customer control room 200, got %d: %s", response.Code, response.Body.String())
	}
	var room model.TenantCustomerControlRoom
	if err := json.Unmarshal(response.Body.Bytes(), &room); err != nil {
		t.Fatalf("decode customer control room: %v", err)
	}
	if room.Summary.ProductScore == 0 || room.Summary.NotificationScore == 0 || room.Summary.PackageScore == 0 || room.Summary.NextBestAction == "" {
		t.Fatalf("expected scored customer control summary: %+v", room.Summary)
	}
	if room.Summary.MailDelivered == 0 || room.Summary.DashboardDelivered == 0 {
		t.Fatalf("expected mail and dashboard proof: %+v", room.Summary)
	}
	if len(room.Tiles) < 8 || len(room.Alerts) == 0 || len(room.Deliveries) < 3 || len(room.Actions) == 0 {
		t.Fatalf("expected tiles, alerts, deliveries, and actions: %+v", room)
	}
	hasPush := false
	hasPackage := false
	hasProvider := false
	for _, tile := range room.Tiles {
		if tile.ID == "push-reach" && tile.Channel == constants.DeliveryChannelPush && tile.PaidTier != "" {
			hasPush = true
		}
		if tile.ID == "package-billing" && tile.Value != "" && tile.PaidTier != "" {
			hasPackage = true
		}
		if tile.ID == "provider-simulation" && tile.Value != "" && tile.Status != "" {
			hasProvider = true
		}
	}
	if !hasPush || !hasPackage || !hasProvider {
		t.Fatalf("expected push, package, and provider tiles: %+v", room.Tiles)
	}
	hasPushDelivery := false
	for _, delivery := range room.Deliveries {
		if delivery.Channel == constants.DeliveryChannelPush && delivery.Status != "" && delivery.NextAction != "" {
			hasPushDelivery = true
		}
	}
	if !hasPushDelivery {
		t.Fatalf("expected push notification delivery evidence: %+v", room.Deliveries)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("customer control room leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(room.PrivacyBoundary, "metadata-only") || !strings.Contains(room.PrivacyBoundary, "no passwords") || !strings.Contains(room.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict customer control privacy boundary, got %q", room.PrivacyBoundary)
	}
}

func TestTenantCustomerSuccessPacketEndpoint(t *testing.T) {
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
		"device_id": "customer-success-device-001",
		"host_name": "customer-success-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentSuccessPacket, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected customer success packet 200, got %d: %s", response.Code, response.Body.String())
	}
	var packet model.TenantCustomerSuccessPacket
	if err := json.Unmarshal(response.Body.Bytes(), &packet); err != nil {
		t.Fatalf("decode customer success packet: %v", err)
	}
	if packet.Summary.ReadinessScore == 0 || packet.Summary.NotificationScore == 0 || packet.Summary.PackageScore == 0 || packet.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored customer success summary: %+v", packet.Summary)
	}
	if packet.Summary.MailDelivered == 0 || packet.Summary.HostsTotal == 0 {
		t.Fatalf("expected mail and host proof in customer success packet: %+v", packet.Summary)
	}
	if len(packet.Proofs) < 7 || len(packet.Objections) < 4 || len(packet.Actions) == 0 {
		t.Fatalf("expected proofs, objections, and actions: %+v", packet)
	}
	hasAnomalyProof := false
	hasMailProof := false
	hasPushProof := false
	hasPrivacyAnswer := false
	for _, proof := range packet.Proofs {
		if proof.ID == "anomaly-command" && proof.BuyerImpact != "" && proof.PaidTier != "" {
			hasAnomalyProof = true
		}
		if proof.ID == "mail-delivery" && proof.Status != "" {
			hasMailProof = true
		}
		if proof.ID == "push-notification" && proof.Status != "" {
			hasPushProof = true
		}
	}
	for _, objection := range packet.Objections {
		if objection.ID == "privacy-boundary" && strings.Contains(objection.Answer, "metadata-only") {
			hasPrivacyAnswer = true
		}
	}
	if !hasAnomalyProof || !hasMailProof || !hasPushProof || !hasPrivacyAnswer {
		t.Fatalf("expected anomaly, mail, push, and privacy proof: proofs=%+v objections=%+v", packet.Proofs, packet.Objections)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("customer success packet leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(packet.PrivacyBoundary, "metadata-only") || !strings.Contains(packet.PrivacyBoundary, "no passwords") || !strings.Contains(packet.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict customer success privacy boundary, got %q", packet.PrivacyBoundary)
	}
}

func TestTenantPushActivationCenterEndpoint(t *testing.T) {
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
		"device_id": "push-activation-device-001",
		"host_name": "push-activation-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentPushActivation, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected push activation center 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.TenantPushActivationCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode push activation center: %v", err)
	}
	if center.Summary.ActivationScore == 0 || center.Summary.NotificationScore == 0 || center.Summary.RecommendedPaidPackage == "" || center.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored push activation summary: %+v", center.Summary)
	}
	if center.Summary.MailDelivered == 0 || center.Summary.DashboardDelivered == 0 || center.Summary.PushRoutesTotal == 0 || center.Summary.AlertRulesUsingPush == 0 || center.Summary.AlertsWithPush == 0 {
		t.Fatalf("expected push, mail fallback, dashboard fallback, routes, and alert rule proof: %+v", center.Summary)
	}
	if len(center.Routes) == 0 || len(center.Scenarios) < 3 || len(center.Actions) == 0 {
		t.Fatalf("expected push routes, scenarios, and owner actions: %+v", center)
	}
	route := center.Routes[0]
	if route.Provider != constants.DeliveryProviderWebPush || route.SubscriptionLabel == "" || route.ProofState == "" || route.EndpointStorage == "" || route.NextAction == "" {
		t.Fatalf("expected provider-safe push route proof: %+v", route)
	}
	if !strings.Contains(route.EndpointStorage, "raw push endpoint is not stored") {
		t.Fatalf("expected route to deny raw endpoint storage, got %q", route.EndpointStorage)
	}
	if center.Scenarios[0].Trigger == "" || len(center.Scenarios[0].Channels) == 0 || center.Scenarios[0].BuyerValue == "" {
		t.Fatalf("expected typed anomaly notification scenario: %+v", center.Scenarios)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("push activation center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no push endpoints") || !strings.Contains(center.PrivacyBoundary, "no passwords") {
		t.Fatalf("expected strict push activation privacy boundary, got %q", center.PrivacyBoundary)
	}
}

func TestTenantPortfolioCenterEndpoint(t *testing.T) {
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

	for _, body := range [][]byte{
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "portfolio-device-001",
			"host_name": "portfolio-study-laptop",
			"profile": "ai-btech-student",
			"os_name": "windows"
		}`),
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "portfolio-device-002",
			"host_name": "portfolio-lab-laptop",
			"profile": "developer-workstation",
			"os_name": "linux"
		}`),
	} {
		enroll := httptest.NewRecorder()
		handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(body)))
		if enroll.Code != http.StatusCreated {
			t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentPortfolioCenter, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected portfolio center 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.TenantPortfolioCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode portfolio center: %v", err)
	}
	if center.Summary.PortfolioScore == 0 || center.Summary.NotificationScore == 0 || center.Summary.HostsTotal < 2 || center.Summary.RecommendedPaidPackage == "" || center.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored portfolio summary: %+v", center.Summary)
	}
	if center.Summary.MailDelivered == 0 || center.Summary.DashboardDelivered == 0 || center.Summary.PushRetrying == 0 || center.Summary.HostsPending == 0 {
		t.Fatalf("expected notification fallback, push retry, and sync proof: %+v", center.Summary)
	}
	if len(center.Hosts) < 2 || len(center.Segments) < 5 || len(center.AlertNotifications) == 0 || len(center.DeliveryProof) < 5 || len(center.Actions) == 0 {
		t.Fatalf("expected host rows, notification proof, portfolio segments, and actions: %+v", center)
	}
	hasMailProof := false
	hasPushProof := false
	for _, proof := range center.DeliveryProof {
		if proof.Channel == constants.DeliveryChannelEmail && proof.Status != "" && proof.NextAction != "" {
			hasMailProof = true
		}
		if proof.Channel == constants.DeliveryChannelPush && proof.Status != "" && proof.NextAction != "" {
			hasPushProof = true
		}
	}
	if !hasMailProof || !hasPushProof {
		t.Fatalf("expected mail and push delivery proof in portfolio center: %+v", center.DeliveryProof)
	}
	alert := center.AlertNotifications[0]
	if alert.EmailStatus == "" || alert.PushStatus == "" || alert.DashboardStatus == "" || alert.NextAction == "" {
		t.Fatalf("expected alert notification route proof: %+v", alert)
	}
	host := center.Hosts[0]
	if host.DeviceID == "" || host.HostName == "" || host.Profile == "" || host.Status == "" || host.MetadataProofSummary == "" || host.NextAction == "" {
		t.Fatalf("expected typed metadata-only host row: %+v", host)
	}
	if host.EmailStatus == "" || host.PushStatus == "" || host.DashboardStatus == "" {
		t.Fatalf("expected host delivery channel status: %+v", host)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("portfolio center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no passwords") || !strings.Contains(center.PrivacyBoundary, "no screenshots") || !strings.Contains(center.PrivacyBoundary, "push endpoints") {
		t.Fatalf("expected strict portfolio privacy boundary, got %q", center.PrivacyBoundary)
	}
}

func TestAccountPortfolioIndexEndpoint(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServer(repo, slog.Default()).Handler()
	for _, body := range [][]byte{
		[]byte(`{
			"tenant_id": "family-varadha",
			"name": "Family Varadha",
			"plan_id": "family_pro",
			"retention_tier_id": "family_cloud_90_365_archive",
			"primary_profile": "ai-btech-student"
		}`),
		[]byte(`{
			"tenant_id": "school-alpha",
			"name": "School Alpha",
			"plan_id": "school",
			"retention_tier_id": "school_year_archive",
			"primary_profile": "school-laptop"
		}`),
	} {
		createTenant := httptest.NewRecorder()
		handler.ServeHTTP(createTenant, httptest.NewRequest(http.MethodPost, constants.RouteTenants, bytes.NewReader(body)))
		if createTenant.Code != http.StatusCreated {
			t.Fatalf("expected tenant create 201, got %d: %s", createTenant.Code, createTenant.Body.String())
		}
	}

	for _, body := range [][]byte{
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "account-device-family",
			"host_name": "account-family-laptop",
			"profile": "ai-btech-student",
			"os_name": "windows"
		}`),
		[]byte(`{
			"tenant_id": "school-alpha",
			"device_id": "account-device-school",
			"host_name": "account-school-laptop",
			"profile": "school-laptop",
			"os_name": "linux"
		}`),
	} {
		enroll := httptest.NewRecorder()
		handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(body)))
		if enroll.Code != http.StatusCreated {
			t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
		}
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteAccountPortfolio, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected account portfolio 200, got %d: %s", response.Code, response.Body.String())
	}
	var index model.AccountPortfolioIndex
	if err := json.Unmarshal(response.Body.Bytes(), &index); err != nil {
		t.Fatalf("decode account portfolio index: %v", err)
	}
	if index.Summary.AccountScore == 0 || index.Summary.NotificationScore == 0 || index.Summary.TenantsTotal != 2 || index.Summary.HostsTotal < 2 || index.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored account portfolio summary: %+v", index.Summary)
	}
	if index.Summary.MailDelivered == 0 || index.Summary.DashboardDelivered == 0 || index.Summary.RoutesNeedingProof == 0 {
		t.Fatalf("expected account notification and route proof: %+v", index.Summary)
	}
	if len(index.Tenants) != 2 || len(index.Proof) < 5 || len(index.Actions) == 0 {
		t.Fatalf("expected tenant rows, proof cards, and actions: %+v", index)
	}
	if index.Tenants[0].TenantID == "" || index.Tenants[0].NextAction == "" || index.Tenants[0].PrivacyBoundary == "" {
		t.Fatalf("expected typed metadata-only tenant row: %+v", index.Tenants[0])
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("account portfolio leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(index.PrivacyBoundary, "metadata-only") || !strings.Contains(index.PrivacyBoundary, "no passwords") || !strings.Contains(index.PrivacyBoundary, "no screenshots") || !strings.Contains(index.PrivacyBoundary, "push endpoints") {
		t.Fatalf("expected strict account privacy boundary, got %q", index.PrivacyBoundary)
	}

	scopedHandler := NewServerWithAuth(repo, slog.Default(), AuthConfig{APIKey: "local-key", TenantID: "school-alpha"}).Handler()
	scopedRequest := httptest.NewRequest(http.MethodGet, constants.RouteAccountPortfolio, nil)
	scopedRequest.Header.Set(constants.HeaderAPIKey, "local-key")
	scopedResponse := httptest.NewRecorder()
	scopedHandler.ServeHTTP(scopedResponse, scopedRequest)
	if scopedResponse.Code != http.StatusOK {
		t.Fatalf("expected scoped account portfolio 200, got %d: %s", scopedResponse.Code, scopedResponse.Body.String())
	}
	var scoped model.AccountPortfolioIndex
	if err := json.Unmarshal(scopedResponse.Body.Bytes(), &scoped); err != nil {
		t.Fatalf("decode scoped account portfolio index: %v", err)
	}
	if scoped.Summary.TenantsTotal != 1 || len(scoped.Tenants) != 1 || scoped.Tenants[0].TenantID != "school-alpha" {
		t.Fatalf("expected tenant-scoped account portfolio, got %+v", scoped)
	}
}

func TestTenantOnboardingCenterEndpoint(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServer(repo, slog.Default()).Handler()
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
		"device_id": "onboarding-device-001",
		"host_name": "onboarding-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	route := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentOnboardingCenter
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected onboarding center 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.TenantOnboardingCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode onboarding center: %v", err)
	}
	if center.Summary.ReadinessScore == 0 || center.Summary.SetupStepsTotal < 8 || center.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored onboarding summary: %+v", center.Summary)
	}
	if center.Summary.HostsTotal < 1 || center.Summary.RolesTotal < 4 || len(center.Roles) < 4 {
		t.Fatalf("expected host and role onboarding proof: %+v", center)
	}
	if len(center.Steps) < 8 || len(center.Proof) < 6 || len(center.Actions) < 1 {
		t.Fatalf("expected setup steps, proof, and actions: %+v", center)
	}
	hasAutostart := false
	hasNotification := false
	hasPrivacy := false
	for _, step := range center.Steps {
		if step.ID == "autostart" && step.Owner != "" && step.Evidence != "" {
			hasAutostart = true
		}
		if step.ID == "mail-push-proof" && step.Blocking {
			hasNotification = true
		}
		if step.ID == "privacy-guard" && step.Blocking {
			hasPrivacy = true
		}
	}
	if !hasAutostart || !hasNotification || !hasPrivacy {
		t.Fatalf("expected autostart, notification, and privacy onboarding steps: %+v", center.Steps)
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no passwords") || !strings.Contains(center.PrivacyBoundary, "no screenshots") || !strings.Contains(center.PrivacyBoundary, "push endpoints") {
		t.Fatalf("expected strict onboarding privacy boundary, got %q", center.PrivacyBoundary)
	}

	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("onboarding center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}

	scopedHandler := NewServerWithAuth(repo, slog.Default(), AuthConfig{APIKey: "local-key", TenantID: "school-alpha"}).Handler()
	scopedRequest := httptest.NewRequest(http.MethodGet, route, nil)
	scopedRequest.Header.Set(constants.HeaderAPIKey, "local-key")
	scopedResponse := httptest.NewRecorder()
	scopedHandler.ServeHTTP(scopedResponse, scopedRequest)
	if scopedResponse.Code != http.StatusForbidden {
		t.Fatalf("expected tenant-scoped onboarding route to reject another tenant, got %d: %s", scopedResponse.Code, scopedResponse.Body.String())
	}
}

func TestTenantCustomerSettingsCenterEndpoint(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServer(repo, slog.Default()).Handler()
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
		"device_id": "settings-device-001",
		"host_name": "settings-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	route := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentCustomerSettings
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected customer settings center 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.TenantCustomerSettingsCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode customer settings center: %v", err)
	}
	if center.Summary.SettingsScore == 0 || center.Summary.SettingsTotal < 8 || center.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored customer settings summary: %+v", center.Summary)
	}
	if len(center.Settings) < 8 || len(center.PlanOptions) < 4 || len(center.RetentionOptions) < 3 || len(center.Channels) < 3 || len(center.Actions) < 1 {
		t.Fatalf("expected settings, plan, retention, channel, and action rows: %+v", center)
	}
	hasPlan := false
	hasRetention := false
	hasPush := false
	hasPrivacy := false
	for _, setting := range center.Settings {
		switch setting.ID {
		case "plan":
			hasPlan = setting.Configurable && setting.CurrentValue != "" && setting.RecommendedValue != ""
		case "retention":
			hasRetention = setting.Configurable && setting.Evidence != ""
		case "push-route":
			hasPush = setting.Configurable && strings.Contains(setting.Evidence, "push")
		case "privacy-data-rights":
			hasPrivacy = !setting.Configurable && setting.NextAction != ""
		}
	}
	if !hasPlan || !hasRetention || !hasPush || !hasPrivacy {
		t.Fatalf("expected plan, retention, push, and privacy settings: %+v", center.Settings)
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no passwords") || !strings.Contains(center.PrivacyBoundary, "no screenshots") || !strings.Contains(center.PrivacyBoundary, "push endpoints") {
		t.Fatalf("expected strict customer settings privacy boundary, got %q", center.PrivacyBoundary)
	}

	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint_url", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("customer settings center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}

	scopedHandler := NewServerWithAuth(repo, slog.Default(), AuthConfig{APIKey: "local-key", TenantID: "school-alpha"}).Handler()
	scopedRequest := httptest.NewRequest(http.MethodGet, route, nil)
	scopedRequest.Header.Set(constants.HeaderAPIKey, "local-key")
	scopedResponse := httptest.NewRecorder()
	scopedHandler.ServeHTTP(scopedResponse, scopedRequest)
	if scopedResponse.Code != http.StatusForbidden {
		t.Fatalf("expected tenant-scoped settings route to reject another tenant, got %d: %s", scopedResponse.Code, scopedResponse.Body.String())
	}
}

func TestTenantRevenueOperationsCenterEndpoint(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServer(repo, slog.Default()).Handler()
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

	for _, body := range [][]byte{
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "revenue-ops-device-001",
			"host_name": "revenue-ops-study-laptop",
			"profile": "ai-btech-student",
			"os_name": "windows"
		}`),
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "revenue-ops-device-002",
			"host_name": "revenue-ops-lab-laptop",
			"profile": "developer-workstation",
			"os_name": "linux"
		}`),
	} {
		enroll := httptest.NewRecorder()
		handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(body)))
		if enroll.Code != http.StatusCreated {
			t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
		}
	}

	response := httptest.NewRecorder()
	route := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentRevenueOps
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected revenue operations center 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.TenantRevenueOperationsCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode revenue operations center: %v", err)
	}
	if center.Summary.RevenueScore == 0 || center.Summary.ProductScore == 0 || center.Summary.NotificationScore == 0 || center.Summary.OwnerNextStep == "" {
		t.Fatalf("expected scored revenue operations summary: %+v", center.Summary)
	}
	if center.Summary.HostsTotal < 2 || center.Summary.MailDelivered == 0 || center.Summary.DashboardDelivered == 0 || center.Summary.RecommendedPaidPackage == "" {
		t.Fatalf("expected host, mail, dashboard, and paid package proof: %+v", center.Summary)
	}
	if len(center.Signals) < 9 || len(center.Alerts) == 0 || len(center.Deliveries) < 3 || len(center.Levers) < 6 || len(center.Actions) == 0 {
		t.Fatalf("expected revenue signals, alerts, deliveries, levers, and actions: %+v", center)
	}
	hasAnomaly := false
	hasMail := false
	hasPush := false
	hasArchive := false
	hasSettings := false
	hasProvider := false
	for _, signal := range center.Signals {
		switch signal.ID {
		case "anomaly-command":
			hasAnomaly = signal.Value != "" && signal.PaidTier != ""
		case "mail-delivery":
			hasMail = signal.Channel == constants.DeliveryChannelEmail && signal.Status != ""
		case "push-reach":
			hasPush = signal.Channel == constants.DeliveryChannelPush && signal.Status != ""
		case "archive-retention":
			hasArchive = signal.PaidTier != ""
		case "customer-settings":
			hasSettings = signal.Value != "" && signal.Detail != ""
		case "provider-simulation":
			hasProvider = signal.Value != "" && signal.Status != ""
		}
	}
	if !hasAnomaly || !hasMail || !hasPush || !hasArchive || !hasSettings || !hasProvider {
		t.Fatalf("expected anomaly, mail, push, archive, settings, and provider signals: %+v", center.Signals)
	}
	hasMailRoute := false
	hasPushRoute := false
	hasDashboardRoute := false
	for _, delivery := range center.Deliveries {
		if delivery.Channel == constants.DeliveryChannelEmail && delivery.NextAction != "" {
			hasMailRoute = true
		}
		if delivery.Channel == constants.DeliveryChannelPush && delivery.NextAction != "" {
			hasPushRoute = true
		}
		if delivery.Channel == constants.DeliveryChannelDashboard && delivery.NextAction != "" {
			hasDashboardRoute = true
		}
	}
	if !hasMailRoute || !hasPushRoute || !hasDashboardRoute {
		t.Fatalf("expected mail, push, and dashboard route proof: %+v", center.Deliveries)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("revenue operations center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no passwords") || !strings.Contains(center.PrivacyBoundary, "no screenshots") || !strings.Contains(center.PrivacyBoundary, "push endpoints") {
		t.Fatalf("expected strict revenue operations privacy boundary, got %q", center.PrivacyBoundary)
	}

	scopedHandler := NewServerWithAuth(repo, slog.Default(), AuthConfig{APIKey: "local-key", TenantID: "school-alpha"}).Handler()
	scopedRequest := httptest.NewRequest(http.MethodGet, route, nil)
	scopedRequest.Header.Set(constants.HeaderAPIKey, "local-key")
	scopedResponse := httptest.NewRecorder()
	scopedHandler.ServeHTTP(scopedResponse, scopedRequest)
	if scopedResponse.Code != http.StatusForbidden {
		t.Fatalf("expected tenant-scoped revenue operations route to reject another tenant, got %d: %s", scopedResponse.Code, scopedResponse.Body.String())
	}
}

func TestTenantDeploymentReadinessCenterEndpoint(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServer(repo, slog.Default()).Handler()
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

	for _, body := range [][]byte{
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "deploy-ready-device-001",
			"host_name": "deploy-ready-windows-laptop",
			"profile": "ai-btech-student",
			"os_name": "windows"
		}`),
		[]byte(`{
			"tenant_id": "family-varadha",
			"device_id": "deploy-ready-device-002",
			"host_name": "deploy-ready-linux-laptop",
			"profile": "developer-workstation",
			"os_name": "linux"
		}`),
	} {
		enroll := httptest.NewRecorder()
		handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(body)))
		if enroll.Code != http.StatusCreated {
			t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
		}
	}

	telemetryBody := []byte(`{
		"tenant_id": "family-varadha",
		"device_id": "deploy-ready-device-001",
		"host_name": "deploy-ready-windows-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows",
		"events": [
			{
				"id": "local-event-66",
				"type": "health.snapshot",
				"source": "collector.health",
				"observed_at": "2026-06-12T08:00:00Z",
				"metadata": { "agent_healthy": "true" }
			}
		]
	}`)
	ingest := httptest.NewRecorder()
	handler.ServeHTTP(ingest, httptest.NewRequest(http.MethodPost, constants.RouteDevices+"/deploy-ready-device-001/"+constants.RouteSegmentTelemetry, bytes.NewReader(telemetryBody)))
	if ingest.Code != http.StatusAccepted {
		t.Fatalf("expected telemetry ingest 202, got %d: %s", ingest.Code, ingest.Body.String())
	}

	response := httptest.NewRecorder()
	route := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentDeploymentReady
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected deployment readiness center 200, got %d: %s", response.Code, response.Body.String())
	}
	var center model.TenantDeploymentReadinessCenter
	if err := json.Unmarshal(response.Body.Bytes(), &center); err != nil {
		t.Fatalf("decode deployment readiness center: %v", err)
	}
	if center.Summary.ReadinessScore == 0 || center.Summary.OwnerNextStep == "" || center.Summary.PlatformsTotal != 3 || center.Summary.ManifestsTotal != 3 {
		t.Fatalf("expected deployment readiness summary with platform and manifest proof: %+v", center.Summary)
	}
	if center.Summary.HostsTotal < 2 || !center.Summary.LiveBootReady || !center.Summary.OfflineReplayReady || center.Summary.RecommendedPackage == "" {
		t.Fatalf("expected live boot, offline replay, host, and package proof: %+v", center.Summary)
	}
	if len(center.Platforms) != 3 || len(center.Manifests) != 3 || len(center.Proof) < 5 || len(center.Actions) < 4 || len(center.Advisories) == 0 {
		t.Fatalf("expected platforms, manifests, proof, actions, and advisories: %+v", center)
	}
	hasServiceAdvisory := false
	for _, advisory := range center.Advisories {
		if advisory.Code == "" || advisory.Headline == "" || advisory.OperatorAction == "" || advisory.EvidenceScope != "metadata_only" {
			t.Fatalf("expected typed metadata-only deployment advisory: %+v", advisory)
		}
		if advisory.ServiceManager != "" {
			hasServiceAdvisory = true
		}
	}
	if !hasServiceAdvisory {
		t.Fatalf("expected service manager advisory proof: %+v", center.Advisories)
	}
	hasWindows := false
	hasDarwin := false
	hasLinux := false
	for _, platform := range center.Platforms {
		switch platform.Platform {
		case constants.PlatformWindows:
			hasWindows = platform.ServiceManager == constants.ServiceManagerTaskScheduler && platform.RegisterScript != "" && platform.StatusScript != ""
		case constants.PlatformDarwin:
			hasDarwin = platform.ServiceManager == constants.ServiceManagerLaunchd && platform.Manifest != ""
		case constants.PlatformLinux:
			hasLinux = platform.ServiceManager == constants.ServiceManagerSystemd && platform.Manifest != ""
		}
	}
	if !hasWindows || !hasDarwin || !hasLinux {
		t.Fatalf("expected Windows, macOS, and Linux deployment platform proof: %+v", center.Platforms)
	}
	hasTaskManifest := false
	hasLaunchdManifest := false
	hasSystemdManifest := false
	for _, manifest := range center.Manifests {
		if manifest.ID == "windows-task" && manifest.TemplatePath == constants.WindowsTaskTemplatePath {
			hasTaskManifest = true
		}
		if manifest.ID == "macos-launchd" && manifest.TemplatePath == constants.DarwinLaunchdTemplate {
			hasLaunchdManifest = true
		}
		if manifest.ID == "linux-systemd" && manifest.TemplatePath == constants.LinuxSystemdTemplate {
			hasSystemdManifest = true
		}
	}
	if !hasTaskManifest || !hasLaunchdManifest || !hasSystemdManifest {
		t.Fatalf("expected Windows task, launchd, and systemd manifest proof: %+v", center.Manifests)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("deployment readiness center leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(center.PrivacyBoundary, "metadata-only") || !strings.Contains(center.PrivacyBoundary, "no passwords") || !strings.Contains(center.PrivacyBoundary, "no screenshots") || !strings.Contains(center.PrivacyBoundary, "hidden collection bypasses") {
		t.Fatalf("expected strict deployment readiness privacy boundary, got %q", center.PrivacyBoundary)
	}

	scopedHandler := NewServerWithAuth(repo, slog.Default(), AuthConfig{APIKey: "local-key", TenantID: "school-alpha"}).Handler()
	scopedRequest := httptest.NewRequest(http.MethodGet, route, nil)
	scopedRequest.Header.Set(constants.HeaderAPIKey, "local-key")
	scopedResponse := httptest.NewRecorder()
	scopedHandler.ServeHTTP(scopedResponse, scopedRequest)
	if scopedResponse.Code != http.StatusForbidden {
		t.Fatalf("expected tenant-scoped deployment readiness route to reject another tenant, got %d: %s", scopedResponse.Code, scopedResponse.Body.String())
	}
}

func TestTenantPremiumOperationsHubEndpoint(t *testing.T) {
	t.Parallel()

	repo := store.NewMemory()
	handler := NewServer(repo, slog.Default()).Handler()
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
		"device_id": "premium-hub-device-001",
		"host_name": "premium-hub-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	route := constants.RouteTenants + "/family-varadha/" + constants.RouteSegmentPremiumOps
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, route, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected premium operations hub 200, got %d: %s", response.Code, response.Body.String())
	}
	var hub model.TenantPremiumOperationsHub
	if err := json.Unmarshal(response.Body.Bytes(), &hub); err != nil {
		t.Fatalf("decode premium operations hub: %v", err)
	}
	if hub.Summary.PremiumScore == 0 || hub.Summary.NotificationScore == 0 || hub.Summary.RecommendedPaidPackage == "" || hub.Summary.OwnerNextStep == "" {
		t.Fatalf("expected premium score, notification score, package, and owner action: %+v", hub.Summary)
	}
	if hub.Summary.MailDelivered == 0 || hub.Summary.DashboardDelivered == 0 || hub.Summary.HostsTotal == 0 {
		t.Fatalf("expected mail, dashboard, and host proof: %+v", hub.Summary)
	}
	if len(hub.Tiles) < 8 || len(hub.Alerts) == 0 || len(hub.Deliveries) < 3 || len(hub.Actions) == 0 {
		t.Fatalf("expected premium tiles, alerts, deliveries, and actions: %+v", hub)
	}
	hasMail := false
	hasPush := false
	hasDashboard := false
	for _, tile := range hub.Tiles {
		switch tile.ID {
		case "mail-delivery":
			hasMail = tile.Channel == constants.DeliveryChannelEmail && tile.NextAction != ""
		case "push-notifications":
			hasPush = tile.Channel == constants.DeliveryChannelPush && tile.NextAction != ""
		case "dashboard-fallback":
			hasDashboard = tile.Channel == constants.DeliveryChannelDashboard && tile.NextAction != ""
		}
	}
	if !hasMail || !hasPush || !hasDashboard {
		t.Fatalf("expected mail, push, and dashboard tile proof: %+v", hub.Tiles)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("premium operations hub leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(hub.PrivacyBoundary, "metadata-only") || !strings.Contains(hub.PrivacyBoundary, "no passwords") || !strings.Contains(hub.PrivacyBoundary, "no screenshots") || !strings.Contains(hub.PrivacyBoundary, "hidden collection bypasses") {
		t.Fatalf("expected strict premium operations privacy boundary, got %q", hub.PrivacyBoundary)
	}

	scopedHandler := NewServerWithAuth(repo, slog.Default(), AuthConfig{APIKey: "local-key", TenantID: "school-alpha"}).Handler()
	scopedRequest := httptest.NewRequest(http.MethodGet, route, nil)
	scopedRequest.Header.Set(constants.HeaderAPIKey, "local-key")
	scopedResponse := httptest.NewRecorder()
	scopedHandler.ServeHTTP(scopedResponse, scopedRequest)
	if scopedResponse.Code != http.StatusForbidden {
		t.Fatalf("expected tenant-scoped premium operations route to reject another tenant, got %d: %s", scopedResponse.Code, scopedResponse.Body.String())
	}
}

func TestTenantNotificationRevenueCockpitEndpoint(t *testing.T) {
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
		"device_id": "notification-revenue-device-001",
		"host_name": "notification-revenue-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentNotificationRev, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected notification revenue cockpit 200, got %d: %s", response.Code, response.Body.String())
	}
	var cockpit model.TenantNotificationRevenueCockpit
	if err := json.Unmarshal(response.Body.Bytes(), &cockpit); err != nil {
		t.Fatalf("decode notification revenue cockpit: %v", err)
	}
	if cockpit.Summary.RevenueReadiness == 0 || cockpit.Summary.NotificationScore == 0 || cockpit.Summary.AlertSLAReady == 0 || cockpit.Summary.RecommendedPaidPackage == "" {
		t.Fatalf("expected scored notification revenue summary: %+v", cockpit.Summary)
	}
	if cockpit.Summary.EmailDelivered == 0 || cockpit.Summary.DashboardDelivered == 0 || cockpit.Summary.NextBestAction == "" {
		t.Fatalf("expected mail/dashboard proof and next action: %+v", cockpit.Summary)
	}
	if len(cockpit.KPIs) < 6 || len(cockpit.Channels) < 3 || len(cockpit.Scenarios) < 4 || len(cockpit.Actions) == 0 {
		t.Fatalf("expected KPI, channel, scenario, and action surfaces: %+v", cockpit)
	}
	hasPush := false
	for _, channel := range cockpit.Channels {
		if channel.Channel == constants.DeliveryChannelPush && channel.PaidTier != "" && channel.BusinessValue != "" {
			hasPush = true
		}
	}
	if !hasPush {
		t.Fatalf("expected push notification business-value proof: %+v", cockpit.Channels)
	}
	if cockpit.Scenarios[0].Trigger == "" || len(cockpit.Scenarios[0].Channels) == 0 || cockpit.Actions[0].ConversionLever == "" {
		t.Fatalf("expected typed scenario and conversion action metadata: %+v %+v", cockpit.Scenarios, cockpit.Actions)
	}
	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("notification revenue cockpit leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
	if !strings.Contains(cockpit.PrivacyBoundary, "metadata-only") || !strings.Contains(cockpit.PrivacyBoundary, "no passwords") || !strings.Contains(cockpit.PrivacyBoundary, "screenshots") {
		t.Fatalf("expected strict notification revenue privacy boundary, got %q", cockpit.PrivacyBoundary)
	}
}

func TestTenantProviderSimulationLabEndpoint(t *testing.T) {
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
		"device_id": "provider-simulation-device-001",
		"host_name": "provider-simulation-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentProviderSim, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected provider simulation lab 200, got %d: %s", response.Code, response.Body.String())
	}
	var lab model.TenantProviderSimulationLab
	if err := json.Unmarshal(response.Body.Bytes(), &lab); err != nil {
		t.Fatalf("decode provider simulation lab: %v", err)
	}
	if lab.Summary.RoutesTotal != 3 || len(lab.Routes) != 3 || len(lab.Scenarios) < 3 || len(lab.Actions) == 0 {
		t.Fatalf("expected provider simulation routes, scenarios, and actions: %+v", lab)
	}
	if lab.Summary.ReadinessScore == 0 || lab.Summary.RecommendedPaidPackage == "" || lab.Summary.NextBestAction == "" {
		t.Fatalf("expected scored provider simulation summary: %+v", lab.Summary)
	}
	if !strings.Contains(lab.PrivacyBoundary, "metadata-only") || !strings.Contains(lab.PrivacyBoundary, "no provider secrets") {
		t.Fatalf("expected strict provider simulation privacy boundary, got %q", lab.PrivacyBoundary)
	}

	run := httptest.NewRecorder()
	runBody := []byte(`{"mode":"dry_run","channel":"push","scenario":"urgent-anomaly-push","reason":"paid buyer push simulation"}`)
	handler.ServeHTTP(run, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentProviderSim, bytes.NewReader(runBody)))
	if run.Code != http.StatusAccepted {
		t.Fatalf("expected provider simulation 202, got %d: %s", run.Code, run.Body.String())
	}
	var simulated model.TenantProviderSimulationLab
	if err := json.Unmarshal(run.Body.Bytes(), &simulated); err != nil {
		t.Fatalf("decode simulated provider lab: %v", err)
	}
	if !simulated.Summary.PushReady || simulated.Summary.SimulatedRoutes == 0 {
		t.Fatalf("expected push simulation proof: %+v", simulated.Summary)
	}
	hasPush := false
	for _, route := range simulated.Routes {
		if route.Channel == constants.DeliveryChannelPush && route.SimulationStatus == constants.StatusHealthy && route.LastSimulatedAt != nil && route.BusinessValue != "" {
			hasPush = true
		}
	}
	if !hasPush {
		t.Fatalf("expected provider-safe push route simulation proof: %+v", simulated.Routes)
	}

	invalid := httptest.NewRecorder()
	handler.ServeHTTP(invalid, httptest.NewRequest(http.MethodPost, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentProviderSim, strings.NewReader(`{"mode":"send_live"}`)))
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid provider simulation mode 400, got %d: %s", invalid.Code, invalid.Body.String())
	}

	audit := httptest.NewRecorder()
	handler.ServeHTTP(audit, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentAuditEvents, nil))
	if audit.Code != http.StatusOK {
		t.Fatalf("expected audit events 200, got %d", audit.Code)
	}
	if !strings.Contains(audit.Body.String(), constants.AuditActionProviderSimulation) {
		t.Fatalf("expected provider simulation audit event, got %s", audit.Body.String())
	}

	serialized := strings.ToLower(run.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("provider simulation lab leaked forbidden marker %q: %s", forbidden, run.Body.String())
		}
	}
}

func TestTenantNotificationProviderSetupEndpoint(t *testing.T) {
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
		"device_id": "provider-setup-device-001",
		"host_name": "provider-setup-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentProviderSetup, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected notification provider setup 200, got %d: %s", response.Code, response.Body.String())
	}
	var setup model.TenantNotificationProviderSetup
	if err := json.Unmarshal(response.Body.Bytes(), &setup); err != nil {
		t.Fatalf("decode notification provider setup: %v", err)
	}
	if setup.Summary.RoutesTotal != 3 || setup.Summary.ChannelsTotal != 3 || len(setup.Channels) != 3 {
		t.Fatalf("expected email, push, and dashboard setup channels: %+v", setup.Summary)
	}
	if setup.Summary.DemoOnly == 0 || setup.Summary.Retrying == 0 {
		t.Fatalf("expected demo-only and retrying truth labels: %+v", setup.Summary)
	}
	if setup.Summary.EmailProviderConfirmed || setup.Summary.PushProviderConfirmed || setup.Summary.BuyerReady {
		t.Fatalf("expected setup to avoid false provider-confirmed claims: %+v", setup.Summary)
	}
	if len(setup.Checklist) == 0 || len(setup.Actions) == 0 {
		t.Fatalf("expected provider setup checklist and actions: %+v", setup)
	}
	if !strings.Contains(setup.PrivacyBoundary, "metadata-only") || !strings.Contains(setup.PrivacyBoundary, "no provider secrets") {
		t.Fatalf("expected provider setup privacy boundary: %q", setup.PrivacyBoundary)
	}

	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "raw_provider_payload"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("provider setup leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
	}
}

func TestTenantPackageBillingReadinessEndpoint(t *testing.T) {
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
		"device_id": "package-billing-device-001",
		"host_name": "package-billing-study-laptop",
		"profile": "ai-btech-student",
		"os_name": "windows"
	}`)
	enroll := httptest.NewRecorder()
	handler.ServeHTTP(enroll, httptest.NewRequest(http.MethodPost, constants.RouteDeviceEnroll, bytes.NewReader(deviceBody)))
	if enroll.Code != http.StatusCreated {
		t.Fatalf("expected device enroll 201, got %d: %s", enroll.Code, enroll.Body.String())
	}

	response := httptest.NewRecorder()
	handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, constants.RouteTenants+"/family-varadha/"+constants.RouteSegmentPackageBilling, nil))
	if response.Code != http.StatusOK {
		t.Fatalf("expected package billing readiness 200, got %d: %s", response.Code, response.Body.String())
	}
	var readiness model.TenantPackageBillingReadiness
	if err := json.Unmarshal(response.Body.Bytes(), &readiness); err != nil {
		t.Fatalf("decode package billing readiness: %v", err)
	}
	if readiness.Summary.PackageScore == 0 || readiness.Summary.CurrentPlan == "" || readiness.Summary.BillingStatus == "" || readiness.Summary.NextBestAction == "" {
		t.Fatalf("expected scored package billing summary: %+v", readiness.Summary)
	}
	if readiness.PlanID != constants.PlanFamilyPro || readiness.RetentionTierID != constants.RetentionFamilyCloud || readiness.RetentionName == "" {
		t.Fatalf("expected typed plan and retention proof: %+v", readiness)
	}
	if len(readiness.Plans) < 4 || len(readiness.FeatureGates) < 8 || len(readiness.Milestones) < 5 || len(readiness.Actions) == 0 {
		t.Fatalf("expected plans, feature gates, milestones, and actions: %+v", readiness)
	}
	hasBillingGate := false
	hasArchiveGate := false
	for _, gate := range readiness.FeatureGates {
		if gate.ID == "billing-setup" && gate.BuyerValue != "" && gate.PaidTier != "" {
			hasBillingGate = true
		}
		if gate.ID == "archive-retention" && gate.Enabled {
			hasArchiveGate = true
		}
	}
	if !hasBillingGate || !hasArchiveGate {
		t.Fatalf("expected billing and archive feature gates: %+v", readiness.FeatureGates)
	}
	if !strings.Contains(readiness.PrivacyBoundary, "metadata-only") || !strings.Contains(readiness.PrivacyBoundary, "no payment card data") || !strings.Contains(readiness.PrivacyBoundary, "no passwords") {
		t.Fatalf("expected strict package billing privacy boundary, got %q", readiness.PrivacyBoundary)
	}

	serialized := strings.ToLower(response.Body.String())
	for _, forbidden := range []string{"card_number", "cvv", "payment_token", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("package billing readiness leaked forbidden marker %q: %s", forbidden, response.Body.String())
		}
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
