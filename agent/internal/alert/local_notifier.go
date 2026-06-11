package alert

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
)

type Notification struct {
	To        []string  `json:"to"`
	Subject   string    `json:"subject"`
	CreatedAt time.Time `json:"created_at"`
	Alerts    []Alert   `json:"alerts"`
}

type LocalNotifier struct {
	outboxDir string
}

func NewLocalNotifier(outboxDir string) *LocalNotifier {
	if outboxDir == "" {
		outboxDir = constants.DefaultOutboxDir
	}
	return &LocalNotifier{outboxDir: outboxDir}
}

func (n *LocalNotifier) Notify(_ context.Context, policy *config.Policy, alerts []Alert) (string, error) {
	if len(alerts) == 0 {
		return "", nil
	}

	alertDir := filepath.Join(n.outboxDir, constants.AlertOutboxDirName)
	if err := os.MkdirAll(alertDir, 0o750); err != nil {
		return "", fmt.Errorf("create alert outbox: %w", err)
	}

	createdAt := time.Now().UTC()
	notification := Notification{
		To:        policy.Alerts.Email.To,
		Subject:   fmt.Sprintf("TraceDeck alert: %d event(s) for %s", len(alerts), policy.DeviceID),
		CreatedAt: createdAt,
		Alerts:    alerts,
	}

	path := filepath.Join(alertDir, fmt.Sprintf("%s-%s%s",
		policy.DeviceID,
		createdAt.Format("20060102T150405Z"),
		constants.JSONExt,
	))

	root, err := os.OpenRoot(alertDir)
	if err != nil {
		return "", fmt.Errorf("open alert outbox root: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	file, err := root.Create(filepath.Base(path))
	if err != nil {
		return "", fmt.Errorf("create alert notification: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(notification); err != nil {
		return "", fmt.Errorf("write alert notification: %w", err)
	}

	return path, nil
}
