package app

import (
	"context"
	"fmt"
	"log/slog"

	processcollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/logging"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/storage/sqlite"
)

type RunOptions struct {
	ConfigPath   string
	DataDir      string
	LogDir       string
	LogLevel     string
	Once         bool
	ProcessLimit int
}

type RunResult struct {
	CollectedEvents int
	StoredEvents    int
}

func Run(ctx context.Context, opts RunOptions) (RunResult, error) {
	if opts.ConfigPath == "" {
		opts.ConfigPath = constants.DefaultConfig
	}

	policy, err := config.LoadFile(opts.ConfigPath)
	if err != nil {
		return RunResult{}, err
	}

	logger, closeLogger, err := logging.New(opts.LogDir, opts.LogLevel)
	if err != nil {
		return RunResult{}, err
	}
	defer func() {
		if err := closeLogger(); err != nil {
			slog.Default().Warn("close logger", "error", err)
		}
	}()

	store, err := sqlite.Open(opts.DataDir)
	if err != nil {
		return RunResult{}, err
	}
	defer func() {
		if err := store.Close(); err != nil {
			logger.Warn("close sqlite store", "error", err)
		}
	}()

	if err := store.EnforceRetention(ctx, policy.Retention.LocalTTLDays); err != nil {
		return RunResult{}, err
	}

	collector := processcollector.New(opts.ProcessLimit)
	events, err := collector.Collect(ctx, policy)
	if err != nil {
		return RunResult{}, err
	}

	for _, evt := range events {
		if err := store.SaveEvent(ctx, evt); err != nil {
			return RunResult{}, err
		}
	}

	total, err := store.CountEvents(ctx)
	if err != nil {
		return RunResult{}, err
	}

	logger.Info("process snapshot collected",
		"tenant_id", policy.TenantID,
		"device_id", policy.DeviceID,
		"collected_events", len(events),
		"stored_events", total,
	)

	if !opts.Once {
		logger.Warn("continuous mode is not enabled in this phase; completed one local snapshot")
	}

	return RunResult{CollectedEvents: len(events), StoredEvents: total}, nil
}

func FormatRunResult(result RunResult) string {
	return fmt.Sprintf("TraceDeck run complete: collected_events=%d stored_events=%d", result.CollectedEvents, result.StoredEvents)
}
