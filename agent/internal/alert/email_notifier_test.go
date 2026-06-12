package alert

import (
	"context"
	"net/smtp"
	"strings"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

func TestBuildEmailMessageIncludesHeadersAndAlertBody(t *testing.T) {
	t.Parallel()

	policy := testEmailPolicy()
	msg, err := BuildEmailMessage(policy, []Alert{testAlert()}, time.Date(2026, 6, 12, 5, 30, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("build email message: %v", err)
	}
	text := string(msg)
	for _, expected := range []string{
		"From: alerts@example.com",
		"To: varathu09@example.com",
		"Subject: TraceDeck alert: 1 event(s) for laptop-cousin-001",
		constants.AlertRuleBlockedAppOpened,
		"vlc.exe",
	} {
		if !strings.Contains(text, expected) {
			t.Fatalf("expected email body to contain %q, got %s", expected, text)
		}
	}
}

func TestSMTPNotifierSendsMessage(t *testing.T) {
	t.Parallel()

	var capturedAddr string
	var capturedFrom string
	var capturedTo []string
	var capturedBody string
	notifier := &SMTPNotifier{
		host:      "127.0.0.1",
		port:      "2525",
		serverTLS: constants.SMTPNoTLS,
		sendMail: func(addr string, _ smtp.Auth, from string, to []string, msg []byte) error {
			capturedAddr = addr
			capturedFrom = from
			capturedTo = append([]string(nil), to...)
			capturedBody = string(msg)
			return nil
		},
	}

	ref, err := notifier.Notify(context.Background(), testEmailPolicy(), []Alert{testAlert()})
	if err != nil {
		t.Fatalf("notify smtp: %v", err)
	}
	if ref != "smtp://127.0.0.1:2525" {
		t.Fatalf("unexpected delivery ref: %s", ref)
	}
	if capturedAddr != "127.0.0.1:2525" || capturedFrom != "alerts@example.com" || len(capturedTo) != 1 || capturedTo[0] != "varathu09@example.com" {
		t.Fatalf("unexpected smtp send args addr=%s from=%s to=%v", capturedAddr, capturedFrom, capturedTo)
	}
	if !strings.Contains(capturedBody, constants.AlertRuleBlockedAppOpened) {
		t.Fatalf("expected alert body, got %s", capturedBody)
	}
}

func testEmailPolicy() *config.Policy {
	return &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Alerts: config.AlertPolicy{
			Email: config.EmailPolicy{
				Provider: config.EmailProvider(constants.EmailProviderSMTP),
				From:     "alerts@example.com",
				To:       []string{"varathu09@example.com"},
			},
		},
	}
}

func testAlert() Alert {
	return Alert{
		RuleName: constants.AlertRuleBlockedAppOpened,
		Severity: constants.SeverityHigh,
		Reason:   "blocked app opened",
		AppName:  "vlc.exe",
		Metadata: map[string]string{
			constants.AlertMetadataRuleName: constants.AlertRuleBlockedAppOpened,
		},
	}
}
