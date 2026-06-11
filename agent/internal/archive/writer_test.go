package archive

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
)

func TestWriterCreatesCompressedBatchAndS3Key(t *testing.T) {
	t.Parallel()

	policy := &config.Policy{
		TenantID: constants.DefaultTenantID,
		DeviceID: constants.DefaultDeviceID,
		Archive: config.ArchivePolicy{
			PrefixTemplate: "tenants/{tenant_id}/devices/{device_id}/hosts/{host_name}/date={yyyy}-{mm}-{dd}/hour={hh}/",
		},
	}
	observedAt := time.Date(2026, time.June, 11, 17, 30, 0, 0, time.UTC)
	events := []event.Event{{
		Type:      constants.EventTypeProcessObserved,
		Source:    constants.EventSourceProcessCollector,
		Timestamp: observedAt,
		TenantID:  policy.TenantID,
		DeviceID:  policy.DeviceID,
		HostName:  constants.UnknownHost,
		AppName:   constants.AppName,
	}}

	batch, err := NewWriter(t.TempDir()).WriteBatch(context.Background(), policy, events)
	if err != nil {
		t.Fatalf("write batch: %v", err)
	}
	if batch.Count != len(events) {
		t.Fatalf("expected %d event, got %d", len(events), batch.Count)
	}
	if !strings.Contains(batch.S3Key, "date=2026-06-11/hour=17") {
		t.Fatalf("unexpected s3 key: %s", batch.S3Key)
	}

	file, err := os.Open(batch.LocalPath)
	if err != nil {
		t.Fatalf("open batch: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Fatalf("close batch: %v", err)
		}
	}()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		t.Fatalf("open gzip: %v", err)
	}
	defer func() {
		if err := gzipReader.Close(); err != nil {
			t.Fatalf("close gzip: %v", err)
		}
	}()

	var decoded event.Event
	if err := json.NewDecoder(gzipReader).Decode(&decoded); err != nil {
		t.Fatalf("decode event: %v", err)
	}
	if decoded.AppName != constants.AppName {
		t.Fatalf("unexpected app name: %s", decoded.AppName)
	}
}
