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
	heartbeatcollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/heartbeat"
	processcollector "github.com/varadharajaan/tracedeck-agent/agent/internal/collector/process"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/exporter"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/logging"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/storage/sqlite"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/syncer"
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
	Cycles           int
	CollectedEvents  int
	StoredEvents     int
	ArchiveBatch     string
	ArchiveUploaded  bool
	AlertsRaised     int
	AlertOutboxPath  string
	AlertDelivered   bool
	BrowserEvents    int
	HealthEvents     int
	HeartbeatEvents  int
	TelemetrySynced  bool
	TelemetryEvents  int
	TelemetryBacklog int
	OTelExported     bool
	OTelEvents       int
	OTelDropped      int
	OTelAttempts     int
	OTelBacklog      int
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

	heartbeatEvents, err := heartbeatcollector.New(platformAdapter, heartbeatcollector.Options{
		ContinuousMode:     !r.opts.Once,
		CollectionInterval: r.opts.CollectionInterval,
		ArchiveEnabled:     r.policy.Archive.Enabled,
		ArchiveDue:         archiveEnabled,
		BackendSyncEnabled: r.policy.BackendSync.Enabled,
		AlertsEnabled:      alertEnabled,
	}).Collect(ctx, r.policy)
	if err != nil {
		return RunResult{}, err
	}
	events = append(events, heartbeatEvents...)

	for _, evt := range events {
		if err := r.store.SaveEvent(ctx, evt); err != nil {
			return RunResult{}, err
		}
	}

	total, err := r.store.CountEvents(ctx)
	if err != nil {
		return RunResult{}, err
	}

	result := RunResult{
		Cycles:          1,
		CollectedEvents: len(events),
		StoredEvents:    total,
		BrowserEvents:   len(browserEvents),
		HealthEvents:    len(healthEvents),
		HeartbeatEvents: len(heartbeatEvents),
	}

	if r.policy.BackendSync.Enabled {
		syncOutcome, err := r.syncBackendTelemetry(ctx, platformAdapter)
		result.TelemetrySynced = syncOutcome.Synced
		result.TelemetryEvents = syncOutcome.AcceptedEvents
		result.TelemetryBacklog = syncOutcome.PendingAfter
		if err != nil {
			r.logger.Warn("backend telemetry sync deferred",
				"error", err,
				"pending_events", syncOutcome.PendingBefore,
				"last_cursor", syncOutcome.LastCursor,
			)
		}
	}

	if r.policy.Observability.OpenTelemetry.Enabled {
		otelOutcome, err := r.exportOpenTelemetry(ctx, platformAdapter)
		result.OTelExported = otelOutcome.ExportedEvents > 0
		result.OTelEvents = otelOutcome.ExportedEvents
		result.OTelDropped = otelOutcome.DroppedEvents
		result.OTelAttempts = otelOutcome.Attempts
		result.OTelBacklog = otelOutcome.PendingAfter
		if err != nil {
			r.logger.Warn("opentelemetry export completed with drops",
				"error", err,
				"pending_before", otelOutcome.PendingBefore,
				"pending_after", otelOutcome.PendingAfter,
				"exported_events", otelOutcome.ExportedEvents,
				"dropped_events", otelOutcome.DroppedEvents,
				"attempts", otelOutcome.Attempts,
				"last_cursor", otelOutcome.LastCursor,
			)
		}
	}

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
			notifier, deliveryMode, err := r.alertNotifier()
			if err != nil {
				return RunResult{}, err
			}
			deliveryRef, err := notifier.Notify(ctx, r.policy, alerts)
			if err != nil {
				return RunResult{}, err
			}
			result.AlertOutboxPath = deliveryRef
			result.AlertDelivered = !r.opts.AlertDryRun
			r.logger.Warn("alert notification staged",
				"alert_count", len(alerts),
				"delivery_ref", deliveryRef,
				"delivery_mode", deliveryMode,
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
		"heartbeat_events", len(heartbeatEvents),
		"stored_events", total,
	)

	return result, nil
}

type backendSyncOutcome struct {
	Synced         bool
	AcceptedEvents int
	StoredEvents   int
	PendingBefore  int
	PendingAfter   int
	LastCursor     int64
}

type openTelemetryOutcome struct {
	ExportedEvents int
	DroppedEvents  int
	Attempts       int
	PendingBefore  int
	PendingAfter   int
	LastCursor     int64
}

func (r *cycleRunner) syncBackendTelemetry(ctx context.Context, platformAdapter platform.Adapter) (backendSyncOutcome, error) {
	timeout, err := parseDurationOrDefault(r.policy.BackendSync.RequestTimeout, constants.DefaultBackendSyncTimeout)
	if err != nil {
		return backendSyncOutcome{}, err
	}
	client, err := syncer.NewClient(r.policy.BackendSync.BaseURL, timeout)
	if err != nil {
		return backendSyncOutcome{}, err
	}
	cursor, err := r.store.BackendSyncCursor(ctx, constants.BackendSyncCursorName)
	if err != nil {
		return backendSyncOutcome{}, err
	}
	pending, err := r.store.PendingBackendSyncEvents(ctx, cursor, r.policy.BackendSync.BatchLimit)
	if err != nil {
		return backendSyncOutcome{LastCursor: cursor}, err
	}
	outcome := backendSyncOutcome{
		PendingBefore: len(pending),
		PendingAfter:  len(pending),
		LastCursor:    cursor,
	}
	if len(pending) == 0 {
		r.logger.Info("backend telemetry sync idle", "last_cursor", cursor)
		return outcome, nil
	}

	hostName, err := platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	syncEvents := make([]event.Event, 0, len(pending))
	for _, stored := range pending {
		syncEvents = append(syncEvents, stored.Event)
	}
	syncResult, err := client.IngestEvents(ctx, r.policy, hostName, platformAdapter.Name(), syncEvents)
	if err != nil {
		return outcome, err
	}
	acceptedEvents := syncResult.AcceptedEvents
	if acceptedEvents < 0 {
		acceptedEvents = 0
	}
	if acceptedEvents > len(pending) {
		acceptedEvents = len(pending)
	}
	if acceptedEvents > 0 {
		outcome.LastCursor = pending[acceptedEvents-1].LocalID
		if err := r.store.MarkBackendSyncCursor(ctx, constants.BackendSyncCursorName, outcome.LastCursor); err != nil {
			return outcome, err
		}
	}
	outcome.Synced = acceptedEvents > 0
	outcome.AcceptedEvents = acceptedEvents
	outcome.StoredEvents = syncResult.StoredEvents
	outcome.PendingAfter = len(pending) - acceptedEvents
	r.logger.Info("backend telemetry backlog synced",
		"tenant_id", syncResult.TenantID,
		"device_id", syncResult.DeviceID,
		"accepted_events", acceptedEvents,
		"stored_events", syncResult.StoredEvents,
		"pending_before", outcome.PendingBefore,
		"pending_after", outcome.PendingAfter,
		"last_cursor", outcome.LastCursor,
		"privacy_boundary", syncResult.PrivacyBoundary,
	)
	if acceptedEvents < len(pending) {
		r.logger.Warn("backend telemetry partial ingest",
			"accepted_events", acceptedEvents,
			"attempted_events", len(pending),
			"last_cursor", outcome.LastCursor,
		)
	}
	return outcome, nil
}

func (r *cycleRunner) exportOpenTelemetry(ctx context.Context, platformAdapter platform.Adapter) (openTelemetryOutcome, error) {
	batchLimit := r.policy.Observability.OpenTelemetry.BatchLimit
	if batchLimit <= 0 {
		batchLimit = constants.DefaultOpenTelemetryBatchLimit
	}
	cursor, err := r.store.BackendSyncCursor(ctx, constants.OpenTelemetryCursorName)
	if err != nil {
		return openTelemetryOutcome{}, err
	}
	pending, err := r.store.PendingBackendSyncEvents(ctx, cursor, batchLimit)
	if err != nil {
		return openTelemetryOutcome{LastCursor: cursor}, err
	}
	outcome := openTelemetryOutcome{
		PendingBefore: len(pending),
		PendingAfter:  len(pending),
		LastCursor:    cursor,
	}
	if len(pending) == 0 {
		r.logger.Info("opentelemetry export idle", "last_cursor", cursor)
		return outcome, nil
	}
	hostName, err := platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	events := make([]event.Event, 0, len(pending))
	for _, stored := range pending {
		events = append(events, stored.Event)
	}
	otelExporter, err := exporter.NewOTLPHTTPLogExporter(r.policy.Observability.OpenTelemetry)
	if err != nil {
		return outcome, err
	}
	result, err := otelExporter.Export(ctx, exporter.Request{
		Policy:   r.policy,
		HostName: hostName,
		OSName:   platformAdapter.Name(),
		Events:   events,
	})
	outcome.ExportedEvents = result.ExportedEvents
	outcome.DroppedEvents = result.DroppedEvents
	outcome.Attempts = result.Attempts
	if result.ExportedEvents > 0 || result.DroppedEvents > 0 {
		outcome.LastCursor = pending[len(pending)-1].LocalID
		if markErr := r.store.MarkBackendSyncCursor(ctx, constants.OpenTelemetryCursorName, outcome.LastCursor); markErr != nil {
			return outcome, markErr
		}
	}
	outcome.PendingAfter = len(pending) - result.ExportedEvents - result.DroppedEvents
	if outcome.PendingAfter < 0 {
		outcome.PendingAfter = 0
	}
	r.logger.Info("opentelemetry metadata batch processed",
		"exported_events", result.ExportedEvents,
		"dropped_events", result.DroppedEvents,
		"attempts", result.Attempts,
		"pending_before", outcome.PendingBefore,
		"pending_after", outcome.PendingAfter,
		"last_cursor", outcome.LastCursor,
		"last_status", result.LastStatus,
		"privacy_boundary", constants.OpenTelemetryPrivacyBoundary,
	)
	if err != nil {
		return outcome, err
	}
	return outcome, nil
}

func (r *cycleRunner) alertNotifier() (alert.Notifier, string, error) {
	if r.opts.AlertDryRun {
		return alert.NewLocalNotifier(r.opts.OutboxDir), "local_outbox", nil
	}
	notifier, err := alert.NewProviderNotifier(r.policy)
	if err != nil {
		return nil, "", err
	}
	return notifier, string(r.policy.Alerts.Email.Provider), nil
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
	r.HeartbeatEvents += next.HeartbeatEvents
	if next.AlertOutboxPath != "" {
		r.AlertOutboxPath = next.AlertOutboxPath
	}
	r.AlertDelivered = r.AlertDelivered || next.AlertDelivered
	r.TelemetrySynced = r.TelemetrySynced || next.TelemetrySynced
	r.TelemetryEvents += next.TelemetryEvents
	r.TelemetryBacklog = next.TelemetryBacklog
	r.OTelExported = r.OTelExported || next.OTelExported
	r.OTelEvents += next.OTelEvents
	r.OTelDropped += next.OTelDropped
	r.OTelAttempts += next.OTelAttempts
	r.OTelBacklog = next.OTelBacklog
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
		"TraceDeck run complete: cycles=%d collected_events=%d stored_events=%d browser_events=%d health_events=%d heartbeat_events=%d telemetry_synced=%t telemetry_events=%d telemetry_backlog=%d otel_exported=%t otel_events=%d otel_dropped=%d otel_attempts=%d otel_backlog=%d archive_batch=%s archive_uploaded=%t alerts_raised=%d alert_delivery=%s alert_delivered=%t",
		result.Cycles,
		result.CollectedEvents,
		result.StoredEvents,
		result.BrowserEvents,
		result.HealthEvents,
		result.HeartbeatEvents,
		result.TelemetrySynced,
		result.TelemetryEvents,
		result.TelemetryBacklog,
		result.OTelExported,
		result.OTelEvents,
		result.OTelDropped,
		result.OTelAttempts,
		result.OTelBacklog,
		result.ArchiveBatch,
		result.ArchiveUploaded,
		result.AlertsRaised,
		result.AlertOutboxPath,
		result.AlertDelivered,
	)
}
