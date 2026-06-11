package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/alert"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/archive"
	processcollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/logging"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/storage/sqlite"
)

type RunOptions struct {
	ConfigPath    string
	DataDir       string
	LogDir        string
	LogLevel      string
	OutboxDir     string
	Once          bool
	ProcessLimit  int
	ArchiveOnce   bool
	ArchiveDryRun bool
	AlertOnce     bool
	AlertDryRun   bool
}

type RunResult struct {
	CollectedEvents int
	StoredEvents    int
	ArchiveBatch    string
	ArchiveUploaded bool
	AlertsRaised    int
	AlertOutboxPath string
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

	platformAdapter := platform.Current()
	collector := processcollector.New(opts.ProcessLimit, platformAdapter)
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

	result := RunResult{CollectedEvents: len(events), StoredEvents: total}

	if opts.ArchiveOnce && policy.Archive.Enabled {
		batch, err := archive.NewWriter(opts.OutboxDir).WriteBatch(ctx, policy, events)
		if err != nil {
			return RunResult{}, err
		}
		result.ArchiveBatch = batch.LocalPath
		logger.Info("archive batch staged",
			"bucket", policy.Archive.Bucket,
			"s3_key", batch.S3Key,
			"local_path", batch.LocalPath,
			"event_count", batch.Count,
			"dry_run", opts.ArchiveDryRun,
		)
		if !opts.ArchiveDryRun && batch.LocalPath != "" {
			uploader, err := archive.NewS3Uploader(ctx)
			if err != nil {
				return RunResult{}, err
			}
			if err := uploader.UploadFile(ctx, policy.Archive.Bucket, batch.S3Key, batch.LocalPath); err != nil {
				return RunResult{}, err
			}
			result.ArchiveUploaded = true
		}
	}

	if opts.AlertOnce && policy.Alerts.Enabled {
		alerts := alert.NewEvaluator().Evaluate(ctx, policy, events)
		result.AlertsRaised = len(alerts)
		if len(alerts) > 0 {
			if !opts.AlertDryRun {
				return RunResult{}, fmt.Errorf("email provider delivery is not enabled in this phase; rerun with alert dry-run")
			}
			outboxPath, err := alert.NewLocalNotifier(opts.OutboxDir).Notify(ctx, policy, alerts)
			if err != nil {
				return RunResult{}, err
			}
			result.AlertOutboxPath = outboxPath
			logger.Warn("alert notification staged",
				"alert_count", len(alerts),
				"outbox_path", outboxPath,
				"dry_run", opts.AlertDryRun,
			)
		}
	}

	logger.Info("process snapshot collected",
		"tenant_id", policy.TenantID,
		"device_id", policy.DeviceID,
		"operating_system", platformAdapter.Name(),
		"collected_events", len(events),
		"stored_events", total,
	)

	if !opts.Once {
		logger.Warn("continuous mode is not enabled in this phase; completed one local snapshot")
	}

	return result, nil
}

func FormatRunResult(result RunResult) string {
	return fmt.Sprintf(
		"TraceDeck run complete: collected_events=%d stored_events=%d archive_batch=%s archive_uploaded=%t alerts_raised=%d alert_outbox=%s",
		result.CollectedEvents,
		result.StoredEvents,
		result.ArchiveBatch,
		result.ArchiveUploaded,
		result.AlertsRaised,
		result.AlertOutboxPath,
	)
}
