package config

import (
	"os"
	"strings"
	"testing"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestLoadValidSamplePolicy(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("../../../examples/policies/ai-btech-student.yaml")
	if err != nil {
		t.Fatalf("read sample policy: %v", err)
	}

	policy, err := Load(data)
	if err != nil {
		t.Fatalf("expected sample policy to validate: %v", err)
	}

	if policy.TenantID != "family-varadha" {
		t.Fatalf("unexpected tenant id: %s", policy.TenantID)
	}
	if policy.Collection.SensitiveCapabilities.Screenshots != SensitiveCapabilityMode(constants.SensitiveCapabilityDeny) {
		t.Fatalf("screenshots must remain deny-only")
	}
	if !policy.Collection.ForegroundApp.Enabled ||
		policy.Collection.ForegroundApp.WindowTitleMode != WindowTitleMode(constants.WindowTitleModeNone) {
		t.Fatalf("foreground app collection should be enabled without window titles: %+v", policy.Collection.ForegroundApp)
	}
	if !policy.Collection.Software.Enabled ||
		policy.Collection.Software.InventoryMode != SoftwareInventoryMode(constants.SoftwareInventoryModeMetadataOnly) {
		t.Fatalf("software inventory should be enabled in metadata-only mode: %+v", policy.Collection.Software)
	}
}

func TestLoadRejectsUnknownFields(t *testing.T) {
	t.Parallel()

	_, err := Load([]byte(`
tenant_id: family-varadha
device_id: laptop-cousin-001
profile: ai-btech-student
unknown_field: nope
`))
	if err == nil {
		t.Fatal("expected unknown field to fail validation")
	}
	if !strings.Contains(err.Error(), "field unknown_field not found") {
		t.Fatalf("expected unknown field error, got: %v", err)
	}
}

func TestLoadRejectsSensitiveCapabilityEnablement(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "screenshots: deny", "screenshots: enabled", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected screenshot enablement to be rejected")
	}
	if !strings.Contains(err.Error(), "screenshots") {
		t.Fatalf("expected screenshots in error, got: %v", err)
	}
}

func TestLoadRejectsBadEnum(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "url_mode: domain_only", "url_mode: free_text", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected invalid URL mode to be rejected")
	}
	if !strings.Contains(err.Error(), "collection.browser.url_mode") {
		t.Fatalf("expected url_mode in error, got: %v", err)
	}
}

func TestLoadRejectsForegroundWindowTitleCollection(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "window_title_mode: none", "window_title_mode: raw_title", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected foreground window title mode to be rejected")
	}
	if !strings.Contains(err.Error(), constants.ConfigFieldForegroundWindowTitle) {
		t.Fatalf("expected foreground window title field in error, got: %v", err)
	}
}

func TestLoadRejectsRawSoftwareInventoryMode(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "inventory_mode: metadata_only", "inventory_mode: raw_paths", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected raw software inventory mode to be rejected")
	}
	if !strings.Contains(err.Error(), constants.ConfigFieldSoftwareInventoryMode) {
		t.Fatalf("expected software inventory field in error, got: %v", err)
	}
}

func TestLoadRejectsBadArchiveUploadInterval(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "enabled: false", "enabled: true", 1)
	data = strings.Replace(data, "provider: none", "provider: s3", 1)
	data = strings.Replace(data, `bucket: ""`, "bucket: test-bucket", 1)
	data = strings.Replace(data, `prefix_template: ""`, "prefix_template: tenants/{tenant_id}/", 1)
	data = strings.Replace(data, `upload_interval: ""`, "upload_interval: soon", 1)

	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected invalid archive upload interval to be rejected")
	}
	if !strings.Contains(err.Error(), "archive.upload_interval") {
		t.Fatalf("expected archive upload interval in error, got: %v", err)
	}
}

func TestLoadRejectsBadBackendSyncURL(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "enabled: false\n  base_url: http://127.0.0.1:18080", "enabled: true\n  base_url: not-a-url", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected invalid backend sync URL to be rejected")
	}
	if !strings.Contains(err.Error(), constants.ConfigFieldBackendSyncBaseURL) {
		t.Fatalf("expected backend sync base URL in error, got: %v", err)
	}
}

func TestLoadRejectsBadOpenTelemetryEndpoint(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "enabled: false\n    protocol: otlp_http_json", "enabled: true\n    protocol: otlp_http_json", 1)
	data = strings.Replace(data, "endpoint: http://127.0.0.1:4318/v1/logs", "endpoint: not-a-url", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected invalid OpenTelemetry endpoint to be rejected")
	}
	if !strings.Contains(err.Error(), constants.ConfigFieldOpenTelemetryEndpoint) {
		t.Fatalf("expected OpenTelemetry endpoint in error, got: %v", err)
	}
}

func TestLoadAcceptsWebPushOnlyAlertProvider(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "enabled: false\n  email:", "enabled: true\n  email:", 1)
	data = strings.Replace(data, "provider: none\n    to: []", "provider: none\n    to: []", 1)
	data = strings.Replace(data, `provider: none
    subscription_file: ""`, `provider: web_push
    subscription_file: "data/local/webpush/subscriptions.json"`, 1)
	data = strings.Replace(data, `vapid_public_key: ""`, `vapid_public_key: "public-key"`, 1)
	data = strings.Replace(data, `vapid_private_key_file: ""`, `vapid_private_key_file: "data/local/webpush/vapid-private.key"`, 1)
	data = strings.Replace(data, `vapid_subject: ""`, `vapid_subject: "mailto:alerts@example.com"`, 1)

	if _, err := Load([]byte(data)); err != nil {
		t.Fatalf("expected web push only policy to validate: %v", err)
	}
}

func TestLoadRejectsAlertsWithoutProvider(t *testing.T) {
	t.Parallel()

	data := strings.Replace(validMinimalPolicy(), "enabled: false\n  email:", "enabled: true\n  email:", 1)
	_, err := Load([]byte(data))
	if err == nil {
		t.Fatal("expected alerts without provider to fail validation")
	}
	if !strings.Contains(err.Error(), constants.ConfigErrorAlertProviderRequired) {
		t.Fatalf("expected alert provider error, got: %v", err)
	}
}

func validMinimalPolicy() string {
	return `
tenant_id: family-varadha
device_id: laptop-cousin-001
profile: ai-btech-student
collection:
  transparency_mode: visible_indicator_required
  browser:
    url_mode: domain_only
    collect_page_title: false
    youtube_classification: enabled
    youtube_video_id_mode: hashed
  foreground_app:
    enabled: true
    window_title_mode: none
  software:
    enabled: true
    inventory_mode: metadata_only
  media:
    collect_file_name: true
    collect_file_path: true
    path_mode: full_path
  sensitive_capabilities:
    credentials: deny
    keystrokes: deny
    cookies: deny
    tokens: deny
    private_messages: deny
    screenshots: deny
retention:
  local_ttl_days: 90
  max_local_storage_mb: 2048
archive:
  enabled: false
  provider: none
  bucket: ""
  prefix_template: ""
  upload_interval: ""
  retry_when_online: true
  storage_class_days:
    standard: 90
    standard_ia_until: 365
    archive_after: 365
backend_sync:
  enabled: false
  base_url: http://127.0.0.1:18080
  batch_limit: 100
  request_timeout: 10s
observability:
  opentelemetry:
    enabled: false
    protocol: otlp_http_json
    endpoint: http://127.0.0.1:4318/v1/logs
    batch_limit: 100
    request_timeout: 5s
    retry:
      max_attempts: 2
alerts:
  enabled: false
  email:
    provider: none
    to: []
    min_severity: high
    cooldown_minutes: 30
  push:
    provider: none
    subscription_file: ""
    vapid_public_key: ""
    vapid_private_key_file: ""
    vapid_subject: ""
    ttl_seconds: 3600
    min_severity: high
    cooldown_minutes: 30
thresholds:
  max_video_minutes_per_day: 60
  max_social_minutes_per_day: 30
  max_unknown_app_minutes_per_day: 45
  late_night_usage_start: "23:30"
  late_night_usage_end: "05:00"
alert_rules: {}
`
}
