package browser

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
)

type fakeAdapter struct{}

func (fakeAdapter) Name() string {
	return constants.OperatingSystemWindows
}

func (fakeAdapter) Hostname(context.Context) (string, error) {
	return "test-host", nil
}

func (fakeAdapter) Capabilities() platform.Capabilities {
	return platform.Capabilities{
		OperatingSystem:   constants.OperatingSystemWindows,
		ProcessCollection: true,
		LocalStorage:      true,
	}
}

func (fakeAdapter) ForegroundApp(context.Context) (platform.ForegroundApp, error) {
	return platform.ForegroundApp{}, platform.ErrNoForegroundApp
}

func (fakeAdapter) SoftwareInventory(context.Context) ([]platform.InstalledSoftware, error) {
	return nil, platform.ErrUnsupportedCapability
}

func TestCollectPersistsDomainOnlyBrowserEvents(t *testing.T) {
	t.Parallel()

	historyPath := createHistoryFixture(t)
	policy := browserPolicy()

	events, err := New(
		[]string{historyPath},
		constants.DefaultBrowserLimit,
		filepath.Join(t.TempDir(), constants.BrowserCacheDirName),
		fakeAdapter{},
	).Collect(context.Background(), policy)
	if err != nil {
		t.Fatalf("collect browser history: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected browser events")
	}

	foundYouTube := false
	for _, evt := range events {
		if evt.Type != constants.EventTypeBrowserObserved {
			t.Fatalf("unexpected event type: %s", evt.Type)
		}
		if evt.Source != constants.EventSourceBrowserCollector {
			t.Fatalf("unexpected event source: %s", evt.Source)
		}
		if evt.Metadata[constants.EventMetadataStoredURLMode] != constants.URLModeDomainOnly {
			t.Fatalf("stored URL mode must remain domain_only: %+v", evt.Metadata)
		}
		for _, value := range evt.Metadata {
			if value == "https://www.youtube.com/watch?v=abc123" {
				t.Fatal("raw URL leaked into metadata")
			}
		}
		if evt.Metadata[constants.EventMetadataDomain] == constants.DomainYouTubeLong {
			foundYouTube = true
			if evt.Metadata[constants.EventMetadataCategory] != constants.CategoryVideoStreaming {
				t.Fatalf("expected YouTube video-streaming category, got %+v", evt.Metadata)
			}
			if evt.Metadata[constants.EventMetadataYouTubeVideoID] == "" {
				t.Fatalf("expected hashed YouTube video id: %+v", evt.Metadata)
			}
		}
	}
	if !foundYouTube {
		t.Fatal("expected youtube.com browser event")
	}
}

func TestNormalizeDomain(t *testing.T) {
	t.Parallel()

	domain := normalizeDomain("https://www.Learn.Microsoft.com/training")
	if domain != "learn.microsoft.com" {
		t.Fatalf("unexpected domain: %s", domain)
	}
}

func TestClassifyStudyYouTubeSearch(t *testing.T) {
	t.Parallel()

	category, study, _ := classify(browserPolicy(), constants.DomainYouTubeLong, "https://www.youtube.com/results?search_query=python+tutorial")
	if category != constants.CategoryStudy || !study {
		t.Fatalf("expected study youtube classification, got category=%s study=%t", category, study)
	}
}

func createHistoryFixture(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), constants.BrowserHistoryFileName)
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open fixture sqlite: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close fixture sqlite: %v", err)
		}
	}()

	if _, err := db.Exec(`
CREATE TABLE urls (
  id INTEGER PRIMARY KEY,
  url TEXT NOT NULL,
  title TEXT,
  visit_count INTEGER NOT NULL,
  last_visit_time INTEGER NOT NULL
)`); err != nil {
		t.Fatalf("create urls table: %v", err)
	}

	lastVisit := chromeMicroseconds(time.Now().UTC())
	entries := []struct {
		rawURL     string
		visitCount int
	}{
		{rawURL: "https://www.youtube.com/watch?v=abc123", visitCount: 2},
		{rawURL: "https://learn.microsoft.com/training", visitCount: 4},
	}
	for _, entry := range entries {
		if _, err := db.Exec(`INSERT INTO urls (url, title, visit_count, last_visit_time) VALUES (?, ?, ?, ?)`, entry.rawURL, "ignored title", entry.visitCount, lastVisit); err != nil {
			t.Fatalf("insert fixture row: %v", err)
		}
	}

	return path
}

func chromeMicroseconds(value time.Time) int64 {
	return value.UTC().Sub(time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)).Microseconds()
}

func browserPolicy() *config.Policy {
	return &config.Policy{
		TenantID:       constants.DefaultTenantID,
		DeviceID:       constants.DefaultDeviceID,
		Profile:        constants.DefaultProfile,
		AllowedDomains: []string{"learn.microsoft.com"},
		YouTubeKeywords: []string{
			"python",
		},
		Collection: config.CollectionPolicy{
			Browser: config.BrowserCollection{
				URLMode:            config.URLMode(constants.URLModeDomainOnly),
				YouTubeVideoIDMode: config.VideoIDMode(constants.VideoIDModeHashed),
			},
		},
	}
}
