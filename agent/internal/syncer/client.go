package syncer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

type Client struct {
	baseURL    *url.URL
	httpClient *http.Client
}

type IngestResult struct {
	TenantID           string    `json:"tenant_id"`
	DeviceID           string    `json:"device_id"`
	AcceptedEvents     int       `json:"accepted_events"`
	StoredEvents       int       `json:"stored_events"`
	LastObservedAt     time.Time `json:"last_observed_at"`
	LastIngestedAt     time.Time `json:"last_ingested_at"`
	PrivacyBoundary    string    `json:"privacy_boundary"`
	BackendVisibleHost bool      `json:"backend_visible_host"`
}

type ingestRequest struct {
	TenantID string        `json:"tenant_id"`
	DeviceID string        `json:"device_id"`
	HostName string        `json:"host_name"`
	Profile  string        `json:"profile"`
	OSName   string        `json:"os_name"`
	Events   []ingestEvent `json:"events"`
}

type ingestEvent struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Source     string            `json:"source"`
	ObservedAt time.Time         `json:"observed_at"`
	TenantID   string            `json:"tenant_id"`
	DeviceID   string            `json:"device_id"`
	HostName   string            `json:"host_name"`
	AppName    string            `json:"app_name"`
	ProcessID  int32             `json:"process_id"`
	PathHash   string            `json:"path_hash"`
	Metadata   map[string]string `json:"metadata"`
}

func NewClient(baseURL string, timeout time.Duration) (*Client, error) {
	parsed, err := url.Parse(strings.TrimRight(strings.TrimSpace(baseURL), "/"))
	if err != nil || parsed == nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid backend sync base_url %q", baseURL)
	}
	if timeout <= 0 {
		return nil, fmt.Errorf("invalid backend sync timeout %s", timeout)
	}
	return &Client{
		baseURL: parsed,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

func (c *Client) IngestEvents(ctx context.Context, policy *config.Policy, hostName string, osName string, events []event.Event) (IngestResult, error) {
	if policy == nil {
		return IngestResult{}, fmt.Errorf("policy is required")
	}
	payload := ingestRequest{
		TenantID: strings.TrimSpace(policy.TenantID),
		DeviceID: strings.TrimSpace(policy.DeviceID),
		HostName: strings.TrimSpace(hostName),
		Profile:  strings.TrimSpace(policy.Profile),
		OSName:   strings.TrimSpace(osName),
		Events:   make([]ingestEvent, 0, len(events)),
	}
	for _, evt := range events {
		payload.Events = append(payload.Events, ingestEvent{
			ID:         strings.TrimSpace(evt.ID),
			Type:       strings.TrimSpace(evt.Type),
			Source:     strings.TrimSpace(evt.Source),
			ObservedAt: evt.Timestamp.UTC(),
			TenantID:   payload.TenantID,
			DeviceID:   payload.DeviceID,
			HostName:   payload.HostName,
			AppName:    strings.TrimSpace(evt.AppName),
			ProcessID:  evt.ProcessID,
			PathHash:   strings.TrimSpace(evt.PathHash),
			Metadata:   cloneMetadata(evt.Metadata),
		})
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return IngestResult{}, fmt.Errorf("encode backend sync payload: %w", err)
	}
	endpoint := c.baseURL.ResolveReference(&url.URL{
		Path: "/api/v1/devices/" + url.PathEscape(payload.DeviceID) + "/telemetry-events",
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return IngestResult{}, fmt.Errorf("create backend sync request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return IngestResult{}, fmt.Errorf("post backend telemetry: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	if resp.StatusCode != http.StatusAccepted {
		return IngestResult{}, fmt.Errorf("backend telemetry ingest returned %s", resp.Status)
	}

	var result IngestResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return IngestResult{}, fmt.Errorf("decode backend sync response: %w", err)
	}
	return result, nil
}

func cloneMetadata(input map[string]string) map[string]string {
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}
	return output
}
