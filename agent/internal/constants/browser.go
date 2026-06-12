package constants

const (
	DefaultBrowserHistoryLimit = 128
	BrowserCacheDirName        = "browser-cache"
)

const (
	BrowserNameChrome  = "chrome"
	BrowserNameEdge    = "edge"
	BrowserNameBrave   = "brave"
	BrowserNameUnknown = "unknown-browser"
)

const (
	BrowserHistoryFileName = "History"
)

const (
	CategoryStudy          = "study"
	CategoryVideoStreaming = "video-streaming"
	CategorySocialMedia    = "social-media"
	CategoryGaming         = "gaming"
	CategoryShopping       = "shopping"
	CategoryBlocked        = "blocked"
	CategoryUnknown        = "unknown"
)

const (
	DomainYouTubeLong  = "youtube.com"
	DomainYouTubeShort = "youtu.be"
)

const (
	ChromeHistoryQuery = `
SELECT url, last_visit_time, visit_count
FROM urls
WHERE url != ''
ORDER BY last_visit_time DESC
LIMIT ?`
)
