package exporter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

type Exporter interface {
	Export(ctx context.Context, request Request) (Result, error)
}

type Request struct {
	Policy   *config.Policy
	HostName string
	OSName   string
	Events   []event.Event
}

type Result struct {
	AttemptedEvents int
	ExportedEvents  int
	DroppedEvents   int
	Attempts        int
	LastStatus      string
	LastError       string
}

type OTLPHTTPLogExporter struct {
	endpoint    *url.URL
	httpClient  *http.Client
	maxAttempts int
}

func NewOTLPHTTPLogExporter(policy config.OpenTelemetryPolicy) (*OTLPHTTPLogExporter, error) {
	endpointValue := strings.TrimSpace(policy.Endpoint)
	if endpointValue == "" {
		endpointValue = constants.DefaultOpenTelemetryEndpoint
	}
	endpoint, err := url.Parse(endpointValue)
	if err != nil || endpoint == nil || endpoint.Host == "" || (endpoint.Scheme != "http" && endpoint.Scheme != "https") {
		return nil, fmt.Errorf("invalid opentelemetry endpoint %q", policy.Endpoint)
	}
	timeoutValue := strings.TrimSpace(policy.RequestTimeout)
	if timeoutValue == "" {
		timeoutValue = constants.DefaultOpenTelemetryTimeout
	}
	timeout, err := time.ParseDuration(timeoutValue)
	if err != nil || timeout <= 0 {
		return nil, fmt.Errorf("invalid opentelemetry request timeout %q", policy.RequestTimeout)
	}
	maxAttempts := policy.Retry.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = constants.DefaultOpenTelemetryMaxAttempts
	}
	return &OTLPHTTPLogExporter{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		maxAttempts: maxAttempts,
	}, nil
}

func (e *OTLPHTTPLogExporter) Export(ctx context.Context, request Request) (Result, error) {
	if e == nil {
		return Result{}, fmt.Errorf("opentelemetry exporter is nil")
	}
	if request.Policy == nil {
		return Result{}, fmt.Errorf("policy is required")
	}
	result := Result{
		AttemptedEvents: len(request.Events),
	}
	if len(request.Events) == 0 {
		return result, nil
	}
	body, err := buildOTLPLogsPayload(request)
	if err != nil {
		result.DroppedEvents = len(request.Events)
		result.LastError = err.Error()
		return result, err
	}

	var lastErr error
	for attempt := 1; attempt <= e.maxAttempts; attempt++ {
		result.Attempts = attempt
		status, err := e.post(ctx, body)
		result.LastStatus = status
		if err == nil {
			result.ExportedEvents = len(request.Events)
			return result, nil
		}
		lastErr = err
		result.LastError = err.Error()
	}
	result.DroppedEvents = len(request.Events)
	return result, fmt.Errorf("opentelemetry export failed after %d attempt(s): %w", e.maxAttempts, lastErr)
}

func (e *OTLPHTTPLogExporter) post(ctx context.Context, body []byte) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create opentelemetry request: %w", err)
	}
	req.Header.Set("Content-Type", constants.OpenTelemetryContentTypeJSON)

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("post opentelemetry logs: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return resp.Status, fmt.Errorf("opentelemetry collector returned %s", resp.Status)
	}
	return resp.Status, nil
}

func buildOTLPLogsPayload(request Request) ([]byte, error) {
	payload := otlpLogsPayload{
		ResourceLogs: []resourceLogs{
			{
				Resource: resource{
					Attributes: []attribute{
						stringAttribute(constants.OpenTelemetryServiceNameKey, constants.AppName),
						stringAttribute(constants.OpenTelemetryServiceVersionKey, constants.AppVersion),
						stringAttribute(constants.OpenTelemetryAttrTenantID, strings.TrimSpace(request.Policy.TenantID)),
						stringAttribute(constants.OpenTelemetryAttrDeviceID, strings.TrimSpace(request.Policy.DeviceID)),
						stringAttribute(constants.OpenTelemetryAttrHostName, strings.TrimSpace(request.HostName)),
						stringAttribute(constants.OpenTelemetryAttrOSName, strings.TrimSpace(request.OSName)),
						stringAttribute(constants.OpenTelemetryAttrProfile, strings.TrimSpace(request.Policy.Profile)),
						stringAttribute(constants.OpenTelemetryAttrPrivacyBoundary, constants.OpenTelemetryPrivacyBoundary),
					},
				},
				ScopeLogs: []scopeLogs{
					{
						Scope: instrumentationScope{
							Name:    constants.OpenTelemetryScopeName,
							Version: constants.AppVersion,
						},
						LogRecords: logRecords(request.Events),
					},
				},
			},
		},
	}
	return json.Marshal(payload)
}

func logRecords(events []event.Event) []logRecord {
	records := make([]logRecord, 0, len(events))
	observedAt := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	for _, evt := range events {
		eventAt := evt.Timestamp.UTC()
		if eventAt.IsZero() {
			eventAt = time.Now().UTC()
		}
		attributes := []attribute{
			stringAttribute(constants.OpenTelemetryAttrEventID, strings.TrimSpace(evt.ID)),
			stringAttribute(constants.OpenTelemetryAttrEventType, strings.TrimSpace(evt.Type)),
			stringAttribute(constants.OpenTelemetryAttrEventSource, strings.TrimSpace(evt.Source)),
			stringAttribute(constants.OpenTelemetryAttrTenantID, strings.TrimSpace(evt.TenantID)),
			stringAttribute(constants.OpenTelemetryAttrDeviceID, strings.TrimSpace(evt.DeviceID)),
			stringAttribute(constants.OpenTelemetryAttrHostName, strings.TrimSpace(evt.HostName)),
		}
		if strings.TrimSpace(evt.AppName) != "" {
			attributes = append(attributes, stringAttribute(constants.OpenTelemetryAttrAppName, strings.TrimSpace(evt.AppName)))
		}
		if evt.ProcessID > 0 {
			attributes = append(attributes, intAttribute(constants.OpenTelemetryAttrProcessID, int64(evt.ProcessID)))
		}
		if strings.TrimSpace(evt.PathHash) != "" {
			attributes = append(attributes, stringAttribute(constants.OpenTelemetryAttrPathHash, strings.TrimSpace(evt.PathHash)))
		}
		attributes = append(attributes, metadataAttributes(evt.Metadata)...)
		records = append(records, logRecord{
			TimeUnixNano:         strconv.FormatInt(eventAt.UnixNano(), 10),
			ObservedTimeUnixNano: observedAt,
			SeverityText:         "INFO",
			Body:                 anyValue{StringValue: constants.OpenTelemetryLogBody},
			Attributes:           compactAttributes(attributes),
		})
	}
	return records
}

func metadataAttributes(metadata map[string]string) []attribute {
	attributes := make([]attribute, 0, len(metadata))
	for key, value := range metadata {
		cleanKey := strings.TrimSpace(key)
		cleanValue := strings.TrimSpace(value)
		if cleanKey == "" || cleanValue == "" || isSensitiveMetadata(cleanKey, cleanValue) {
			continue
		}
		attributes = append(attributes, stringAttribute(constants.OpenTelemetryAttrMetadataPrefix+cleanKey, cleanValue))
	}
	return attributes
}

func compactAttributes(attributes []attribute) []attribute {
	output := make([]attribute, 0, len(attributes))
	for _, attr := range attributes {
		if strings.TrimSpace(attr.Key) == "" || attr.Value.empty() {
			continue
		}
		output = append(output, attr)
	}
	return output
}

func isSensitiveMetadata(key string, value string) bool {
	normalizedKey := strings.ToLower(strings.TrimSpace(key))
	normalizedValue := strings.ToLower(strings.TrimSpace(value))
	for _, forbidden := range []string{
		constants.OpenTelemetrySensitiveKeyPassword,
		constants.OpenTelemetrySensitiveKeyCredential,
		constants.OpenTelemetrySensitiveKeyCookie,
		constants.OpenTelemetrySensitiveKeyToken,
		constants.OpenTelemetrySensitiveKeyScreenshot,
		constants.OpenTelemetrySensitiveKeyKeystroke,
		constants.OpenTelemetrySensitiveKeyPrivateMessage,
		constants.OpenTelemetrySensitiveKeyPageTitle,
		constants.OpenTelemetrySensitiveKeyRawURL,
		constants.OpenTelemetrySensitiveKeyFullURL,
		constants.OpenTelemetrySensitiveKeyProviderSecret,
		constants.OpenTelemetrySensitiveKeyPayment,
		constants.OpenTelemetrySensitiveKeyCard,
	} {
		if strings.Contains(normalizedKey, forbidden) {
			return true
		}
	}
	return strings.HasPrefix(normalizedValue, "http://") || strings.HasPrefix(normalizedValue, "https://")
}

func stringAttribute(key string, value string) attribute {
	return attribute{Key: key, Value: anyValue{StringValue: value}}
}

func intAttribute(key string, value int64) attribute {
	return attribute{Key: key, Value: anyValue{IntValue: strconv.FormatInt(value, 10)}}
}

type otlpLogsPayload struct {
	ResourceLogs []resourceLogs `json:"resourceLogs"`
}

type resourceLogs struct {
	Resource  resource    `json:"resource"`
	ScopeLogs []scopeLogs `json:"scopeLogs"`
}

type resource struct {
	Attributes []attribute `json:"attributes"`
}

type scopeLogs struct {
	Scope      instrumentationScope `json:"scope"`
	LogRecords []logRecord          `json:"logRecords"`
}

type instrumentationScope struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type logRecord struct {
	TimeUnixNano         string      `json:"timeUnixNano"`
	ObservedTimeUnixNano string      `json:"observedTimeUnixNano"`
	SeverityText         string      `json:"severityText"`
	Body                 anyValue    `json:"body"`
	Attributes           []attribute `json:"attributes"`
}

type attribute struct {
	Key   string   `json:"key"`
	Value anyValue `json:"value"`
}

type anyValue struct {
	StringValue string `json:"stringValue,omitempty"`
	IntValue    string `json:"intValue,omitempty"`
}

func (v anyValue) empty() bool {
	return v.StringValue == "" && v.IntValue == ""
}
