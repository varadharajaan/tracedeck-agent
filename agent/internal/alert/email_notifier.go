package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"time"

	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type Notifier interface {
	Notify(context.Context, *config.Policy, []Alert) (string, error)
}

type channelNotifier struct {
	name        string
	minSeverity config.Severity
	notifier    Notifier
}

type MultiNotifier struct {
	channels []channelNotifier
}

func NewMultiNotifier(channels []channelNotifier) *MultiNotifier {
	return &MultiNotifier{channels: channels}
}

func (n *MultiNotifier) Notify(ctx context.Context, policy *config.Policy, alerts []Alert) (string, error) {
	refs := make([]string, 0, len(n.channels))
	for _, channel := range n.channels {
		channelAlerts := filterBySeverity(append([]Alert(nil), alerts...), channel.minSeverity)
		if len(channelAlerts) == 0 {
			continue
		}
		ref, err := channel.notifier.Notify(ctx, policy, channelAlerts)
		if ref != "" {
			refs = append(refs, channel.name+"="+ref)
		}
		if err != nil {
			return strings.Join(refs, ";"), fmt.Errorf("%s provider delivery failed: %w", channel.name, err)
		}
	}
	return strings.Join(refs, ";"), nil
}

type SMTPNotifier struct {
	host      string
	port      string
	username  string
	password  string
	serverTLS string
	sendMail  func(string, smtp.Auth, string, []string, []byte) error
}

func NewSMTPNotifierFromEnv() (*SMTPNotifier, error) {
	host := strings.TrimSpace(os.Getenv(constants.EmailEnvSMTPHost))
	if host == "" {
		return nil, fmt.Errorf("%s is required for smtp alert delivery", constants.EmailEnvSMTPHost)
	}
	port := strings.TrimSpace(os.Getenv(constants.EmailEnvSMTPPort))
	if port == "" {
		port = constants.DefaultSMTPPort
	}
	serverTLS := strings.TrimSpace(os.Getenv(constants.EmailEnvSMTPServerTLS))
	if serverTLS == "" {
		serverTLS = constants.SMTPNoTLS
	}
	return &SMTPNotifier{
		host:      host,
		port:      port,
		username:  strings.TrimSpace(os.Getenv(constants.EmailEnvSMTPUsername)),
		password:  os.Getenv(constants.EmailEnvSMTPPassword),
		serverTLS: strings.ToLower(serverTLS),
		sendMail:  smtp.SendMail,
	}, nil
}

func (n *SMTPNotifier) Notify(_ context.Context, policy *config.Policy, alerts []Alert) (string, error) {
	if len(alerts) == 0 {
		return "", nil
	}
	if err := validateEmailPolicy(policy); err != nil {
		return "", err
	}
	if n.serverTLS == constants.SMTPServerTLS {
		return "", errors.New("smtp server tls delivery is not implemented; use an SMTP relay or SES provider")
	}

	addr := net.JoinHostPort(n.host, n.port)
	var auth smtp.Auth
	if n.username != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}
	msg, err := BuildEmailMessage(policy, alerts, time.Now().UTC())
	if err != nil {
		return "", err
	}
	if err := n.sendMail(addr, auth, policy.Alerts.Email.From, policy.Alerts.Email.To, msg); err != nil {
		return "", fmt.Errorf("send smtp alert email: %w", err)
	}
	return "smtp://" + addr, nil
}

type SESNotifier struct{}

func NewSESNotifier() *SESNotifier {
	return &SESNotifier{}
}

func (n *SESNotifier) Notify(ctx context.Context, policy *config.Policy, alerts []Alert) (string, error) {
	if len(alerts) == 0 {
		return "", nil
	}
	if err := validateEmailPolicy(policy); err != nil {
		return "", err
	}

	cfg, err := awscfg.LoadDefaultConfig(ctx)
	if err != nil {
		return "", fmt.Errorf("load aws config for ses: %w", err)
	}
	client := sesv2.NewFromConfig(cfg)
	subject := emailSubject(policy, alerts)
	body, err := emailBody(policy, alerts, time.Now().UTC())
	if err != nil {
		return "", err
	}
	output, err := client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: &policy.Alerts.Email.From,
		Destination: &types.Destination{
			ToAddresses: policy.Alerts.Email.To,
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: &subject},
				Body: &types.Body{
					Text: &types.Content{Data: &body},
				},
			},
		},
	})
	if err != nil {
		return "", fmt.Errorf("send ses alert email: %w", err)
	}
	if output.MessageId == nil {
		return "ses://message", nil
	}
	return "ses://" + *output.MessageId, nil
}

func NewProviderNotifier(policy *config.Policy) (Notifier, error) {
	channels := make([]channelNotifier, 0, 2)
	switch policy.Alerts.Email.Provider {
	case config.EmailProvider(constants.EmailProviderSMTP):
		notifier, err := NewSMTPNotifierFromEnv()
		if err != nil {
			return nil, err
		}
		channels = append(channels, channelNotifier{name: constants.EmailProviderSMTP, minSeverity: policy.Alerts.Email.MinSeverity, notifier: notifier})
	case config.EmailProvider(constants.EmailProviderSES):
		channels = append(channels, channelNotifier{name: constants.EmailProviderSES, minSeverity: policy.Alerts.Email.MinSeverity, notifier: NewSESNotifier()})
	case "", config.EmailProvider(constants.EmailProviderNone):
	default:
		return nil, fmt.Errorf("unsupported email provider %q", policy.Alerts.Email.Provider)
	}

	switch policy.Alerts.Push.Provider {
	case config.PushProvider(constants.PushProviderWebPush):
		notifier, err := NewWebPushNotifierFromPolicy(policy)
		if err != nil {
			return nil, err
		}
		channels = append(channels, channelNotifier{name: constants.PushProviderWebPush, minSeverity: policy.Alerts.Push.MinSeverity, notifier: notifier})
	case "", config.PushProvider(constants.PushProviderNone):
	default:
		return nil, fmt.Errorf("unsupported push provider %q", policy.Alerts.Push.Provider)
	}
	if len(channels) == 0 {
		return nil, errors.New("no alert provider is configured")
	}
	if len(channels) == 1 {
		return channels[0].notifier, nil
	}
	return NewMultiNotifier(channels), nil
}

func BuildEmailMessage(policy *config.Policy, alerts []Alert, createdAt time.Time) ([]byte, error) {
	if err := validateEmailPolicy(policy); err != nil {
		return nil, err
	}
	subject := emailSubject(policy, alerts)
	body, err := emailBody(policy, alerts, createdAt)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	headers := [][2]string{
		{constants.EmailHeaderFrom, policy.Alerts.Email.From},
		{constants.EmailHeaderTo, strings.Join(policy.Alerts.Email.To, ", ")},
		{constants.EmailHeaderSubject, subject},
		{constants.EmailHeaderMIMEVersion, constants.EmailMIMEVersion},
		{constants.EmailHeaderContentType, constants.EmailContentTypeTextPlain},
	}
	for _, header := range headers {
		_, _ = fmt.Fprintf(&buf, "%s: %s\r\n", header[0], header[1])
	}
	_, _ = fmt.Fprintf(&buf, "\r\n%s", body)
	return buf.Bytes(), nil
}

func emailSubject(policy *config.Policy, alerts []Alert) string {
	return fmt.Sprintf("TraceDeck alert: %d event(s) for %s", len(alerts), policy.DeviceID)
}

func emailBody(policy *config.Policy, alerts []Alert, createdAt time.Time) (string, error) {
	notification := Notification{
		To:        policy.Alerts.Email.To,
		Subject:   emailSubject(policy, alerts),
		CreatedAt: createdAt,
		Alerts:    alerts,
	}
	data, err := json.MarshalIndent(notification, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode alert email body: %w", err)
	}
	return string(data) + "\n", nil
}

func validateEmailPolicy(policy *config.Policy) error {
	if policy == nil {
		return errors.New("policy is required")
	}
	if _, err := mail.ParseAddress(policy.Alerts.Email.From); err != nil {
		return fmt.Errorf("invalid alert sender: %w", err)
	}
	if len(policy.Alerts.Email.To) == 0 {
		return errors.New("at least one alert recipient is required")
	}
	for _, recipient := range policy.Alerts.Email.To {
		if _, err := mail.ParseAddress(recipient); err != nil {
			return fmt.Errorf("invalid alert recipient %q: %w", recipient, err)
		}
	}
	return nil
}
