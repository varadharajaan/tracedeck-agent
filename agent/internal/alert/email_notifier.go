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
	switch policy.Alerts.Email.Provider {
	case config.EmailProvider(constants.EmailProviderSMTP):
		return NewSMTPNotifierFromEnv()
	case config.EmailProvider(constants.EmailProviderSES):
		return NewSESNotifier(), nil
	default:
		return nil, fmt.Errorf("unsupported email provider %q", policy.Alerts.Email.Provider)
	}
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
