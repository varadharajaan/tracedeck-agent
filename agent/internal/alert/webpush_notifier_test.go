package alert

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestWebPushNotifierSendsProviderPayload(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	subscriptionPath := filepath.Join(tempDir, "subscriptions.json")
	privateKeyPath := filepath.Join(tempDir, "vapid-private.key")
	if err := os.WriteFile(subscriptionPath, []byte(`{
  "subscriptions": [
    {
      "endpoint": "https://push.example.test/abc",
      "keys": {
        "p256dh": "client-public-key",
        "auth": "client-auth-secret"
      }
    }
  ]
}`), 0o600); err != nil {
		t.Fatalf("write subscription fixture: %v", err)
	}
	if err := os.WriteFile(privateKeyPath, []byte("private-key"), 0o600); err != nil {
		t.Fatalf("write key fixture: %v", err)
	}

	policy := testPushPolicy(subscriptionPath, privateKeyPath)
	notifier, err := NewWebPushNotifierFromPolicy(policy)
	if err != nil {
		t.Fatalf("new web push notifier: %v", err)
	}

	var capturedPayload string
	var capturedEndpoint string
	var capturedOptions *webpush.Options
	notifier.send = func(_ context.Context, payload []byte, subscription *webpush.Subscription, options *webpush.Options) (*http.Response, error) {
		capturedPayload = string(payload)
		capturedEndpoint = subscription.Endpoint
		capturedOptions = options
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil
	}

	ref, err := notifier.Notify(context.Background(), policy, []Alert{testAlert()})
	if err != nil {
		t.Fatalf("notify web push: %v", err)
	}
	if ref != "web_push://delivered/1/1" {
		t.Fatalf("unexpected delivery ref: %s", ref)
	}
	if capturedEndpoint != "https://push.example.test/abc" {
		t.Fatalf("unexpected endpoint: %s", capturedEndpoint)
	}
	if capturedOptions == nil || capturedOptions.VAPIDPublicKey != "public-key" || capturedOptions.VAPIDPrivateKey != "private-key" || capturedOptions.Subscriber != "mailto:alerts@example.com" {
		t.Fatalf("unexpected VAPID options: %+v", capturedOptions)
	}
	for _, expected := range []string{
		`"title":"TraceDeck HIGH alert"`,
		`"tenant_id":"family-varadha"`,
		`"device_id":"laptop-cousin-001"`,
		`"alert_count":1`,
	} {
		if !strings.Contains(capturedPayload, expected) {
			t.Fatalf("expected payload to contain %q, got %s", expected, capturedPayload)
		}
	}
}

func TestWebPushNotifierContinuesAfterFailedSubscription(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	subscriptionPath := filepath.Join(tempDir, "subscriptions.json")
	privateKeyPath := filepath.Join(tempDir, "vapid-private.key")
	if err := os.WriteFile(subscriptionPath, []byte(`{
  "subscriptions": [
    {
      "endpoint": "https://push.example.test/stale",
      "keys": {
        "p256dh": "stale-client-public-key",
        "auth": "stale-client-auth-secret"
      }
    },
    {
      "endpoint": "https://push.example.test/live",
      "keys": {
        "p256dh": "live-client-public-key",
        "auth": "live-client-auth-secret"
      }
    }
  ]
}`), 0o600); err != nil {
		t.Fatalf("write subscription fixture: %v", err)
	}
	if err := os.WriteFile(privateKeyPath, []byte("private-key"), 0o600); err != nil {
		t.Fatalf("write key fixture: %v", err)
	}

	policy := testPushPolicy(subscriptionPath, privateKeyPath)
	notifier, err := NewWebPushNotifierFromPolicy(policy)
	if err != nil {
		t.Fatalf("new web push notifier: %v", err)
	}

	var attempted []string
	notifier.send = func(_ context.Context, _ []byte, subscription *webpush.Subscription, _ *webpush.Options) (*http.Response, error) {
		attempted = append(attempted, subscription.Endpoint)
		if strings.Contains(subscription.Endpoint, "/stale") {
			return nil, io.ErrUnexpectedEOF
		}
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader(""))}, nil
	}

	ref, err := notifier.Notify(context.Background(), policy, []Alert{testAlert()})
	if err != nil {
		t.Fatalf("notify web push with partial subscription failure: %v", err)
	}
	if ref != "web_push://delivered/1/2" {
		t.Fatalf("unexpected delivery ref: %s", ref)
	}
	if len(attempted) != 2 {
		t.Fatalf("expected both subscriptions to be attempted, got %d", len(attempted))
	}
}

func TestWebPushNotifierFailsOnlyWhenAllSubscriptionsFail(t *testing.T) {
	t.Parallel()

	tempDir := t.TempDir()
	subscriptionPath := filepath.Join(tempDir, "subscriptions.json")
	privateKeyPath := filepath.Join(tempDir, "vapid-private.key")
	if err := os.WriteFile(subscriptionPath, []byte(`{
  "subscriptions": [
    {
      "endpoint": "https://push.example.test/stale-a",
      "keys": {
        "p256dh": "stale-a-client-public-key",
        "auth": "stale-a-client-auth-secret"
      }
    },
    {
      "endpoint": "https://push.example.test/stale-b",
      "keys": {
        "p256dh": "stale-b-client-public-key",
        "auth": "stale-b-client-auth-secret"
      }
    }
  ]
}`), 0o600); err != nil {
		t.Fatalf("write subscription fixture: %v", err)
	}
	if err := os.WriteFile(privateKeyPath, []byte("private-key"), 0o600); err != nil {
		t.Fatalf("write key fixture: %v", err)
	}

	policy := testPushPolicy(subscriptionPath, privateKeyPath)
	notifier, err := NewWebPushNotifierFromPolicy(policy)
	if err != nil {
		t.Fatalf("new web push notifier: %v", err)
	}
	notifier.send = func(_ context.Context, _ []byte, _ *webpush.Subscription, _ *webpush.Options) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusGone, Body: io.NopCloser(strings.NewReader(""))}, nil
	}

	ref, err := notifier.Notify(context.Background(), policy, []Alert{testAlert()})
	if err == nil {
		t.Fatalf("expected all-failed web push send to return error")
	}
	if ref != "web_push://delivered/0/2" {
		t.Fatalf("unexpected delivery ref: %s", ref)
	}
}

func TestBuildWebPushPayloadAvoidsRawAlertBody(t *testing.T) {
	t.Parallel()

	payload, err := BuildWebPushPayload(testPushPolicy("", ""), []Alert{{
		Severity:   constants.SeverityCritical,
		Reason:     constants.AlertReasonNonStudyYouTubeObserved,
		ObservedAt: time.Now().UTC(),
		HostName:   constants.UnknownHost,
		Metadata: map[string]string{
			constants.AlertMetadataDomain: "youtube.com",
		},
	}}, time.Date(2026, 6, 16, 4, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("build web push payload: %v", err)
	}
	text := string(payload)
	if strings.Contains(text, "youtube.com") {
		t.Fatalf("web push payload must not include raw alert metadata: %s", text)
	}
	if !strings.Contains(text, constants.AlertReasonNonStudyYouTubeObserved) {
		t.Fatalf("expected high-level alert reason in payload: %s", text)
	}
}

func testPushPolicy(subscriptionPath string, privateKeyPath string) *config.Policy {
	return &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Alerts: config.AlertPolicy{
			Push: config.PushPolicy{
				Provider:            config.PushProvider(constants.PushProviderWebPush),
				SubscriptionFile:    subscriptionPath,
				VAPIDPublicKey:      "public-key",
				VAPIDPrivateKeyFile: privateKeyPath,
				VAPIDSubject:        "mailto:alerts@example.com",
				TTLSeconds:          constants.DefaultWebPushTTLSeconds,
				MinSeverity:         config.Severity(constants.SeverityMedium),
				CooldownMinutes:     constants.DefaultAlertCooldownMins,
			},
		},
	}
}
