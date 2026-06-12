package browser

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/varadharajaan/tracedeck-agent/agent/internal/config"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/event"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/domain/redaction"
	"github.com/varadharajaan/tracedeck-agent/agent/internal/platform"
)

type Collector struct {
	historyPaths    []string
	limit           int
	cacheDir        string
	platformAdapter platform.Adapter
}

type observation struct {
	browserName      string
	domain           string
	category         string
	visitCount       int
	observedAt       time.Time
	historyPath      string
	youtubeStudy     bool
	youtubeVideoHash string
}

type historyRow struct {
	rawURL        string
	lastVisitTime int64
	visitCount    int
}

func New(historyPaths []string, limit int, cacheDir string, platformAdapter platform.Adapter) *Collector {
	if limit <= 0 {
		limit = constants.DefaultBrowserLimit
	}
	if platformAdapter == nil {
		platformAdapter = platform.Current()
	}
	return &Collector{
		historyPaths:    historyPaths,
		limit:           limit,
		cacheDir:        cacheDir,
		platformAdapter: platformAdapter,
	}
}

func (c *Collector) Collect(ctx context.Context, policy *config.Policy) ([]event.Event, error) {
	historyPaths, explicit := c.resolveHistoryPaths()
	if len(historyPaths) == 0 {
		return nil, nil
	}

	hostName, err := c.platformAdapter.Hostname(ctx)
	if err != nil || hostName == "" {
		hostName = constants.UnknownHost
	}
	capabilities := c.platformAdapter.Capabilities()

	observations := make(map[string]observation)
	var collectionErrors []error

	for _, historyPath := range historyPaths {
		rows, err := c.readHistory(ctx, historyPath)
		if err != nil {
			if explicit {
				collectionErrors = append(collectionErrors, err)
			}
			continue
		}

		browserName := inferBrowserName(historyPath)
		for _, row := range rows {
			domain := normalizeDomain(row.rawURL)
			if domain == "" {
				continue
			}

			category, youtubeStudy, videoHash := classify(policy, domain, row.rawURL)
			observedAt := chromeTime(row.lastVisitTime)
			if observedAt.IsZero() {
				observedAt = time.Now().UTC()
			}
			key := browserName + "\x00" + domain + "\x00" + category
			current := observations[key]
			if observedAt.After(current.observedAt) {
				current.observedAt = observedAt
			}
			current.browserName = browserName
			current.domain = domain
			current.category = category
			current.visitCount += row.visitCount
			current.historyPath = historyPath
			current.youtubeStudy = current.youtubeStudy || youtubeStudy
			if current.youtubeVideoHash == "" {
				current.youtubeVideoHash = videoHash
			}
			observations[key] = current
		}
	}

	if len(observations) == 0 && len(collectionErrors) > 0 {
		return nil, errors.Join(collectionErrors...)
	}

	events := make([]event.Event, 0, len(observations))
	for _, item := range observations {
		metadata := map[string]string{
			constants.EventMetadataProfile:         policy.Profile,
			constants.EventMetadataOperatingSystem: capabilities.OperatingSystem,
			constants.EventMetadataBrowserName:     item.browserName,
			constants.EventMetadataDomain:          item.domain,
			constants.EventMetadataCategory:        item.category,
			constants.EventMetadataURLMode:         string(policy.Collection.Browser.URLMode),
			constants.EventMetadataStoredURLMode:   constants.URLModeDomainOnly,
			constants.EventMetadataVisitCount:      strconv.Itoa(item.visitCount),
		}
		if item.youtubeStudy {
			metadata[constants.EventMetadataYouTubeStudy] = strconv.FormatBool(item.youtubeStudy)
		}
		if item.youtubeVideoHash != "" {
			metadata[constants.EventMetadataYouTubeVideoID] = item.youtubeVideoHash
		}

		events = append(events, event.Event{
			Type:      constants.EventTypeBrowserObserved,
			Source:    constants.EventSourceBrowserCollector,
			Timestamp: item.observedAt.UTC(),
			TenantID:  policy.TenantID,
			DeviceID:  policy.DeviceID,
			HostName:  hostName,
			AppName:   item.browserName,
			PathHash:  redaction.HashPath(item.historyPath),
			Metadata:  metadata,
		})
	}

	return events, nil
}

func (c *Collector) resolveHistoryPaths() ([]string, bool) {
	if len(c.historyPaths) > 0 {
		return c.historyPaths, true
	}
	return discoverHistoryPaths(), false
}

func (c *Collector) readHistory(ctx context.Context, historyPath string) ([]historyRow, error) {
	if _, err := os.Stat(historyPath); err != nil {
		return nil, fmt.Errorf("stat browser history %s: %w", historyPath, err)
	}

	cacheDir := c.cacheDir
	if cacheDir == "" {
		cacheDir = filepath.Join(constants.DefaultDataDir, constants.BrowserCacheDirName)
	}
	if err := os.MkdirAll(cacheDir, 0o750); err != nil {
		return nil, fmt.Errorf("create browser cache dir: %w", err)
	}

	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%s-%d.sqlite", constants.BrowserHistoryFileName, time.Now().UnixNano()))
	cacheName := filepath.Base(cachePath)
	if err := copyFile(historyPath, cacheDir, cacheName); err != nil {
		return nil, err
	}
	defer func() {
		_ = os.Remove(cachePath)
	}()

	db, err := sql.Open("sqlite", cachePath)
	if err != nil {
		return nil, fmt.Errorf("open browser history copy: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	rows, err := db.QueryContext(ctx, constants.ChromeHistoryQuery, c.limit)
	if err != nil {
		return nil, fmt.Errorf("query browser history: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	out := make([]historyRow, 0, c.limit)
	for rows.Next() {
		var row historyRow
		if err := rows.Scan(&row.rawURL, &row.lastVisitTime, &row.visitCount); err != nil {
			return nil, fmt.Errorf("scan browser history row: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate browser history rows: %w", err)
	}
	return out, nil
}

func discoverHistoryPaths() []string {
	home, _ := os.UserHomeDir()
	localAppData := os.Getenv("LOCALAPPDATA")

	patterns := []string{}
	switch runtime.GOOS {
	case constants.OperatingSystemWindows:
		patterns = append(patterns,
			filepath.Join(localAppData, "Google", "Chrome", "User Data", "*", constants.BrowserHistoryFileName),
			filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "*", constants.BrowserHistoryFileName),
			filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data", "*", constants.BrowserHistoryFileName),
		)
	case constants.OperatingSystemMacOS:
		patterns = append(patterns,
			filepath.Join(home, "Library", "Application Support", "Google", "Chrome", "*", constants.BrowserHistoryFileName),
			filepath.Join(home, "Library", "Application Support", "Microsoft Edge", "*", constants.BrowserHistoryFileName),
			filepath.Join(home, "Library", "Application Support", "BraveSoftware", "Brave-Browser", "*", constants.BrowserHistoryFileName),
		)
	default:
		patterns = append(patterns,
			filepath.Join(home, ".config", "google-chrome", "*", constants.BrowserHistoryFileName),
			filepath.Join(home, ".config", "microsoft-edge", "*", constants.BrowserHistoryFileName),
			filepath.Join(home, ".config", "BraveSoftware", "Brave-Browser", "*", constants.BrowserHistoryFileName),
		)
	}

	seen := map[string]struct{}{}
	out := []string{}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if _, ok := seen[match]; ok {
				continue
			}
			seen[match] = struct{}{}
			out = append(out, match)
		}
	}
	return out
}

func inferBrowserName(historyPath string) string {
	lower := strings.ToLower(historyPath)
	switch {
	case strings.Contains(lower, "brave"):
		return constants.BrowserNameBrave
	case strings.Contains(lower, "edge"):
		return constants.BrowserNameEdge
	case strings.Contains(lower, "chrome"):
		return constants.BrowserNameChrome
	default:
		return constants.BrowserNameUnknown
	}
}

func normalizeDomain(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := strings.ToLower(strings.TrimSpace(parsed.Hostname()))
	host = strings.TrimPrefix(host, "www.")
	return host
}

func classify(policy *config.Policy, domain string, rawURL string) (string, bool, string) {
	if domainMatchesAny(domain, policy.BlockedDomains) {
		return constants.CategoryBlocked, false, ""
	}
	if domainMatchesAny(domain, policy.AllowedDomains) {
		return constants.CategoryStudy, false, ""
	}

	if isYouTubeDomain(domain) {
		study := containsAnyFold(rawURL, policy.YouTubeKeywords)
		videoHash := ""
		if policy.Collection.Browser.YouTubeVideoIDMode == config.VideoIDMode(constants.VideoIDModeHashed) {
			videoHash = redaction.HashValue(extractYouTubeVideoID(rawURL))
		}
		if study {
			return constants.CategoryStudy, true, videoHash
		}
		return constants.CategoryVideoStreaming, false, videoHash
	}

	switch {
	case domainMatchesAny(domain, []string{"netflix.com", "primevideo.com", "hotstar.com", "disneyplus.com"}):
		return constants.CategoryVideoStreaming, false, ""
	case domainMatchesAny(domain, []string{"facebook.com", "instagram.com", "x.com", "twitter.com", "reddit.com", "snapchat.com"}):
		return constants.CategorySocialMedia, false, ""
	case domainMatchesAny(domain, []string{"steampowered.com", "epicgames.com", "roblox.com"}):
		return constants.CategoryGaming, false, ""
	case domainMatchesAny(domain, []string{"amazon.com", "flipkart.com", "myntra.com"}):
		return constants.CategoryShopping, false, ""
	default:
		return constants.CategoryUnknown, false, ""
	}
}

func isYouTubeDomain(domain string) bool {
	return domain == constants.DomainYouTubeLong ||
		strings.HasSuffix(domain, "."+constants.DomainYouTubeLong) ||
		domain == constants.DomainYouTubeShort
}

func domainMatchesAny(domain string, candidates []string) bool {
	for _, candidate := range candidates {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		candidate = strings.TrimPrefix(candidate, "www.")
		if candidate == "" {
			continue
		}
		if domain == candidate || strings.HasSuffix(domain, "."+candidate) {
			return true
		}
	}
	return false
}

func containsAnyFold(value string, candidates []string) bool {
	lower := strings.ToLower(value)
	for _, candidate := range candidates {
		candidate = strings.ToLower(strings.TrimSpace(candidate))
		if candidate != "" && strings.Contains(lower, candidate) {
			return true
		}
	}
	return false
}

func extractYouTubeVideoID(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	host := normalizeDomain(rawURL)
	if host == constants.DomainYouTubeShort {
		return strings.Trim(strings.TrimPrefix(parsed.Path, "/"), "/")
	}
	if value := parsed.Query().Get("v"); value != "" {
		return value
	}
	if strings.HasPrefix(parsed.Path, "/shorts/") {
		return strings.Trim(strings.TrimPrefix(parsed.Path, "/shorts/"), "/")
	}
	return ""
}

func chromeTime(value int64) time.Time {
	if value <= 0 {
		return time.Time{}
	}
	return time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC).Add(time.Duration(value) * time.Microsecond)
}

func copyFile(sourcePath string, targetDir string, targetName string) error {
	sourceRoot, err := os.OpenRoot(filepath.Dir(sourcePath))
	if err != nil {
		return fmt.Errorf("open browser history source root: %w", err)
	}
	defer func() {
		_ = sourceRoot.Close()
	}()

	source, err := sourceRoot.Open(filepath.Base(sourcePath))
	if err != nil {
		return fmt.Errorf("open browser history source: %w", err)
	}
	defer func() {
		_ = source.Close()
	}()

	targetRoot, err := os.OpenRoot(targetDir)
	if err != nil {
		return fmt.Errorf("open browser history cache root: %w", err)
	}
	defer func() {
		_ = targetRoot.Close()
	}()

	target, err := targetRoot.OpenFile(targetName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create browser history cache: %w", err)
	}
	defer func() {
		_ = target.Close()
	}()

	if _, err := io.Copy(target, source); err != nil {
		return fmt.Errorf("copy browser history: %w", err)
	}
	return nil
}
