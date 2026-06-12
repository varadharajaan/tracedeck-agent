package app

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/alert"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/archive"
	browsercollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/browser"
	healthcollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/health"
	processcollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/logging"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/storage/sqlite"
)

type RunOptions struct {
	ConfigPath            string
	DataDir               string
	LogDir                string
	LogLevel              string
	OutboxDir             string
	Once                  bool
	ProcessLimit          int
	ArchiveOnce           bool
	ArchiveDryRun         bool
	AlertOnce             bool
	AlertDryRun           bool
	CollectionInterval    string
	MaxCycles             int
	BrowserHistoryPath    []string
	BrowserHistoryLimit   int
	BrowserCacheDir       string
	DisableBrowserHistory bool
}

type RunResult struct {
	Cycles          int
	CollectedEvents int
	StoredEvents    int
	ArchiveBatch    string
	ArchiveUploaded bool
	AlertsRaised    int
	AlertOutboxPath string
	BrowserEvents   int
	HealthEvents    int
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

	runner := &cycleRunner{
		opts:   opts,
		policy: policy,
		logger: logger,
		store:  store,
	}

	if opts.Once {
		return runner.runCycle(ctx, opts.ArchiveOnce && policy.Archive.Enabled, opts.AlertOnce && policy.Alerts.Enabled)
	}

	return runner.runContinuous(ctx)
}

type cycleRunner struct {
	opts   RunOptions
	policy *config.Policy
	logger *slog.Logger
	store  *sqlite.Store
}

func (r *cycleRunner) runContinuous(ctx context.Context) (RunResult, error) {
	collectionInterval, err := parseDurationOrDefault(r.opts.CollectionInterval, constants.DefaultCollectionInterval)
	if err != nil {
		return RunResult{}, err
	}
	archiveInterval, err := parseDurationOrDefault(r.policy.Archive.UploadInterval, constants.DefaultUploadInterval)
	if err != nil {
		return RunResult{}, err
	}

	maxCycles := r.opts.MaxCycles
	if maxCycles < 0 {
		maxCycles = constants.DefaultMaxCycles
	}

	r.logger.Info("continuous agent loop started",
		"collection_interval", collectionInterval.String(),
		"archive_interval", archiveInterval.String(),
		"max_cycles", maxCycles,
	)

	var aggregate RunResult
	var lastArchiveAt time.Time

	for {
		if maxCycles > 0 && aggregate.Cycles >= maxCycles {
			return aggregate, nil
		}

		archiveDue := r.policy.Archive.Enabled && (lastArchiveAt.IsZero() || time.Since(lastArchiveAt) >= archiveInterval)
		cycle, err := r.runCycle(ctx, archiveDue, r.policy.Alerts.Enabled)
		if err != nil {
			return RunResult{}, err
		}
		aggregate.merge(cycle)
		if archiveDue && cycle.ArchiveBatch != "" {
			lastArchiveAt = time.Now()
		}

		if maxCycles > 0 && aggregate.Cycles >= maxCycles {
			return aggregate, nil
		}

		timer := time.NewTimer(collectionInterval)
		select {
		case <-ctx.Done():
			timer.Stop()
			r.logger.Info("continuous agent loop stopped", "reason", ctx.Err())
			return aggregate, nil
		case <-timer.C:
		}
	}
}

func (r *cycleRunner) runCycle(ctx context.Context, archiveEnabled bool, alertEnabled bool) (RunResult, error) {
	if err := r.store.EnforceRetention(ctx, r.policy.Retention.LocalTTLDays); err != nil {
		return RunResult{}, err
	}

	platformAdapter := platform.Current()
	collector := processcollector.New(r.opts.ProcessLimit, platformAdapter)
	processEvents, err := collector.Collect(ctx, r.policy)
	if err != nil {
		return RunResult{}, err
	}
	events := processEvents

	healthEvents, err := healthcollector.New(platformAdapter).Collect(ctx, r.policy)
	if err != nil {
		return RunResult{}, err
	}
	events = append(events, healthEvents...)

	browserEvents := []event.Event{}
	if !r.opts.DisableBrowserHistory {
		browserCacheDir := r.opts.BrowserCacheDir
		if browserCacheDir == "" {
			dataDir := r.opts.DataDir
			if dataDir == "" {
				dataDir = constants.DefaultDataDir
			}
			browserCacheDir = filepath.Join(dataDir, constants.BrowserCacheDirName)
		}
		browserEvents, err = browsercollector.New(
			r.opts.BrowserHistoryPath,
			r.opts.BrowserHistoryLimit,
			browserCacheDir,
			platformAdapter,
		).Collect(ctx, r.policy)
		if err != nil {
			return RunResult{}, err
		}
	}
	events = append(events, browserEvents...)

	for _, evt := range events {
		if err := r.store.SaveEvent(ctx, evt); err != nil {
			return RunResult{}, err
		}
	}

	total, err := r.store.CountEvents(ctx)
	if err != nil {
		return RunResult{}, err
	}

	result := RunResult{Cycles: 1, CollectedEvents: len(events), StoredEvents: total, BrowserEvents: len(browserEvents), HealthEvents: len(healthEvents)}

	if archiveEnabled {
		batch, err := archive.NewWriter(r.opts.OutboxDir).WriteBatch(ctx, r.policy, events)
		if err != nil {
			return RunResult{}, err
		}
		result.ArchiveBatch = batch.LocalPath
		r.logger.Info("archive batch staged",
			"bucket", r.policy.Archive.Bucket,
			"s3_key", batch.S3Key,
			"local_path", batch.LocalPath,
			"event_count", batch.Count,
			"dry_run", r.opts.ArchiveDryRun,
		)
		if !r.opts.ArchiveDryRun && batch.LocalPath != "" {
			uploader, err := archive.NewS3Uploader(ctx)
			if err != nil {
				return RunResult{}, err
			}
			if err := uploader.UploadFile(ctx, r.policy.Archive.Bucket, batch.S3Key, batch.LocalPath); err != nil {
				return RunResult{}, err
			}
			result.ArchiveUploaded = true
		}
	}

	if alertEnabled {
		alerts := alert.NewEvaluator().Evaluate(ctx, r.policy, events)
		result.AlertsRaised = len(alerts)
		if len(alerts) > 0 {
			if !r.opts.AlertDryRun {
				return RunResult{}, fmt.Errorf("email provider delivery is not enabled in this phase; rerun with alert dry-run")
			}
			outboxPath, err := alert.NewLocalNotifier(r.opts.OutboxDir).Notify(ctx, r.policy, alerts)
			if err != nil {
				return RunResult{}, err
			}
			result.AlertOutboxPath = outboxPath
			r.logger.Warn("alert notification staged",
				"alert_count", len(alerts),
				"outbox_path", outboxPath,
				"dry_run", r.opts.AlertDryRun,
			)
		}
	}

	r.logger.Info("process snapshot collected",
		"tenant_id", r.policy.TenantID,
		"device_id", r.policy.DeviceID,
		"operating_system", platformAdapter.Name(),
		"collected_events", len(events),
		"process_events", len(processEvents),
		"health_events", len(healthEvents),
		"browser_events", len(browserEvents),
		"stored_events", total,
	)

	return result, nil
}

func (r *RunResult) merge(next RunResult) {
	r.Cycles += next.Cycles
	r.CollectedEvents += next.CollectedEvents
	r.StoredEvents = next.StoredEvents
	if next.ArchiveBatch != "" {
		r.ArchiveBatch = next.ArchiveBatch
	}
	r.ArchiveUploaded = r.ArchiveUploaded || next.ArchiveUploaded
	r.AlertsRaised += next.AlertsRaised
	r.BrowserEvents += next.BrowserEvents
	r.HealthEvents += next.HealthEvents
	if next.AlertOutboxPath != "" {
		r.AlertOutboxPath = next.AlertOutboxPath
	}
}

func parseDurationOrDefault(value string, fallback string) (time.Duration, error) {
	if value == "" {
		value = fallback
	}
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("invalid duration %q", value)
	}
	return duration, nil
}

func FormatRunResult(result RunResult) string {
	return fmt.Sprintf(
		"TraceDeck run complete: cycles=%d collected_events=%d stored_events=%d browser_events=%d health_events=%d archive_batch=%s archive_uploaded=%t alerts_raised=%d alert_outbox=%s",
		result.Cycles,
		result.CollectedEvents,
		result.StoredEvents,
		result.BrowserEvents,
		result.HealthEvents,
		result.ArchiveBatch,
		result.ArchiveUploaded,
		result.AlertsRaised,
		result.AlertOutboxPath,
	)
}
