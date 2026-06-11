package archive

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

type Batch struct {
	LocalPath string
	S3Key     string
	Count     int
}

type Writer struct {
	outboxDir string
}

func NewWriter(outboxDir string) *Writer {
	if outboxDir == "" {
		outboxDir = constants.DefaultOutboxDir
	}
	return &Writer{outboxDir: outboxDir}
}

func (w *Writer) WriteBatch(_ context.Context, policy *config.Policy, events []event.Event) (Batch, error) {
	if len(events) == 0 {
		return Batch{}, nil
	}

	archiveDir := filepath.Join(w.outboxDir, constants.ArchiveOutboxDirName)
	if err := os.MkdirAll(archiveDir, 0o750); err != nil {
		return Batch{}, fmt.Errorf("create archive outbox: %w", err)
	}

	observedAt := events[0].Timestamp.UTC()
	if observedAt.IsZero() {
		observedAt = time.Now().UTC()
	}

	fileName := fmt.Sprintf("%s-%s%s",
		policy.DeviceID,
		observedAt.Format("20060102T150405Z"),
		constants.JSONLinesGzipExt,
	)
	localPath := filepath.Join(archiveDir, fileName)

	root, err := os.OpenRoot(archiveDir)
	if err != nil {
		return Batch{}, fmt.Errorf("open archive outbox root: %w", err)
	}
	defer func() {
		_ = root.Close()
	}()

	file, err := root.Create(fileName)
	if err != nil {
		return Batch{}, fmt.Errorf("create archive batch: %w", err)
	}
	defer func() {
		_ = file.Close()
	}()

	gzipWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzipWriter)
	for _, evt := range events {
		if err := encoder.Encode(evt); err != nil {
			_ = gzipWriter.Close()
			return Batch{}, fmt.Errorf("write archive event: %w", err)
		}
	}
	if err := gzipWriter.Close(); err != nil {
		return Batch{}, fmt.Errorf("close archive gzip stream: %w", err)
	}

	return Batch{
		LocalPath: localPath,
		S3Key:     renderS3Key(policy, events[0].HostName, observedAt, fileName),
		Count:     len(events),
	}, nil
}

func renderS3Key(policy *config.Policy, hostName string, observedAt time.Time, fileName string) string {
	if hostName == "" {
		hostName = constants.UnknownHost
	}

	prefix := policy.Archive.PrefixTemplate
	replacer := strings.NewReplacer(
		constants.TemplateTenantID, policy.TenantID,
		constants.TemplateDeviceID, policy.DeviceID,
		constants.TemplateHostName, hostName,
		constants.TemplateYear, observedAt.Format("2006"),
		constants.TemplateMonth, observedAt.Format("01"),
		constants.TemplateDay, observedAt.Format("02"),
		constants.TemplateHour, observedAt.Format("15"),
	)
	prefix = replacer.Replace(prefix)
	prefix = strings.TrimPrefix(prefix, "/")
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	return prefix + fileName
}
