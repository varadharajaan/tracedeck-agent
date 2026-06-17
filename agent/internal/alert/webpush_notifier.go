package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type WebPushNotification struct {
	Title      string    `json:"title"`
	Body       string    `json:"body"`
	CreatedAt  time.Time `json:"created_at"`
	TenantID   string    `json:"tenant_id"`
	DeviceID   string    `json:"device_id"`
	HostName   string    `json:"host_name,omitempty"`
	Severity   string    `json:"severity"`
	AlertCount int       `json:"alert_count"`
}

type WebPushSubscriptionSet struct {
	Subscriptions []webpush.Subscription `json:"subscriptions"`
}

type webPushSendFunc func(context.Context, []byte, *webpush.Subscription, *webpush.Options) (*http.Response, error)

type WebPushNotifier struct {
	subscriptionFile string
	vapidPublicKey   string
	vapidPrivateKey  string
	vapidSubject     string
	ttlSeconds       int
	send             webPushSendFunc
}

func NewWebPushNotifierFromPolicy(policy *config.Policy) (*WebPushNotifier, error) {
	if policy == nil {
		return nil, errors.New("policy is required")
	}
	push := policy.Alerts.Push
	subscriptionFile := firstNonEmpty(os.Getenv(constants.WebPushEnvSubscriptionFile), push.SubscriptionFile)
	publicKey := firstNonEmpty(os.Getenv(constants.WebPushEnvVAPIDPublicKey), push.VAPIDPublicKey)
	privateKey, err := webPushPrivateKey(push.VAPIDPrivateKeyFile)
	if err != nil {
		return nil, err
	}
	subject := firstNonEmpty(os.Getenv(constants.WebPushEnvVAPIDSubject), push.VAPIDSubject)
	ttlSeconds := push.TTLSeconds
	if ttlSeconds <= 0 {
		ttlSeconds = constants.DefaultWebPushTTLSeconds
	}
	if strings.TrimSpace(subscriptionFile) == "" {
		return nil, errors.New("web push subscription file is required")
	}
	if strings.TrimSpace(publicKey) == "" {
		return nil, errors.New("web push VAPID public key is required")
	}
	if strings.TrimSpace(privateKey) == "" {
		return nil, errors.New("web push VAPID private key is required")
	}
	if strings.TrimSpace(subject) == "" {
		return nil, errors.New("web push VAPID subject is required")
	}
	return &WebPushNotifier{
		subscriptionFile: subscriptionFile,
		vapidPublicKey:   publicKey,
		vapidPrivateKey:  privateKey,
		vapidSubject:     subject,
		ttlSeconds:       ttlSeconds,
		send:             webpush.SendNotificationWithContext,
	}, nil
}

func (n *WebPushNotifier) Notify(ctx context.Context, policy *config.Policy, alerts []Alert) (string, error) {
	if len(alerts) == 0 {
		return "", nil
	}
	if policy == nil {
		return "", errors.New("policy is required")
	}
	subscriptions, err := loadWebPushSubscriptions(n.subscriptionFile)
	if err != nil {
		return "", err
	}
	if len(subscriptions) == 0 {
		return "", errors.New("web push subscription file did not contain any subscriptions")
	}
	payload, err := BuildWebPushPayload(policy, alerts, time.Now().UTC())
	if err != nil {
		return "", err
	}
	options := &webpush.Options{
		Subscriber:      n.vapidSubject,
		VAPIDPublicKey:  n.vapidPublicKey,
		VAPIDPrivateKey: n.vapidPrivateKey,
		TTL:             n.ttlSeconds,
		Urgency:         webpush.UrgencyHigh,
	}

	delivered := 0
	failures := 0
	for index := range subscriptions {
		response, sendErr := n.send(ctx, payload, &subscriptions[index], options)
		if response != nil {
			_ = response.Body.Close()
		}
		if sendErr != nil {
			failures++
			continue
		}
		if response == nil {
			failures++
			continue
		}
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			failures++
			continue
		}
		delivered++
	}
	if delivered == 0 {
		return fmt.Sprintf("web_push://delivered/%d/%d", delivered, len(subscriptions)), fmt.Errorf("send web push notification failed for all subscriptions: %d failed", failures)
	}
	return fmt.Sprintf("web_push://delivered/%d/%d", delivered, len(subscriptions)), nil
}

func BuildWebPushPayload(policy *config.Policy, alerts []Alert, createdAt time.Time) ([]byte, error) {
	if policy == nil {
		return nil, errors.New("policy is required")
	}
	highest := highestSeverity(alerts)
	first := Alert{}
	if len(alerts) > 0 {
		first = alerts[0]
	}
	body := fmt.Sprintf("%d TraceDeck alert(s) need attention.", len(alerts))
	if first.Reason != "" {
		body = first.Reason
	}
	notification := WebPushNotification{
		Title:      fmt.Sprintf("TraceDeck %s alert", strings.ToUpper(highest)),
		Body:       body,
		CreatedAt:  createdAt,
		TenantID:   policy.TenantID,
		DeviceID:   policy.DeviceID,
		HostName:   first.HostName,
		Severity:   highest,
		AlertCount: len(alerts),
	}
	return json.Marshal(notification)
}

func loadWebPushSubscriptions(path string) ([]webpush.Subscription, error) {
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return nil, fmt.Errorf("read web push subscription file: %w", err)
	}
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})

	var set WebPushSubscriptionSet
	if err := json.Unmarshal(data, &set); err == nil && len(set.Subscriptions) > 0 {
		return set.Subscriptions, nil
	}

	var one webpush.Subscription
	if err := json.Unmarshal(data, &one); err != nil {
		return nil, fmt.Errorf("parse web push subscription file: %w", err)
	}
	if strings.TrimSpace(one.Endpoint) == "" {
		return nil, errors.New("web push subscription endpoint is required")
	}
	return []webpush.Subscription{one}, nil
}

func webPushPrivateKey(path string) (string, error) {
	value := strings.TrimSpace(os.Getenv(constants.WebPushEnvVAPIDPrivateKey))
	if value != "" {
		return value, nil
	}
	if strings.TrimSpace(path) == "" {
		return "", nil
	}
	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		return "", fmt.Errorf("read web push VAPID private key file: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

func highestSeverity(alerts []Alert) string {
	highest := constants.SeverityLow
	highestRank := 0
	for _, candidate := range alerts {
		rank := severityRank(candidate.Severity)
		if rank > highestRank {
			highestRank = rank
			highest = candidate.Severity
		}
	}
	return highest
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
