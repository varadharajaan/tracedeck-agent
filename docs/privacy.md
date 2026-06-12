# Privacy

TraceDeck collects typed endpoint metadata only. It does not collect passwords,
keystrokes, browser cookies, auth tokens, private messages, camera, microphone,
or covert screenshots.

Sensitive capabilities are deny-only in typed policy. Media file name/path and
browser video metadata require explicit policy enablement. Browser activity is
domain/category based by default, not full URL based.

The process collector stores app/process names and executable path hashes. It
does not persist raw executable paths.

The browser collector copies history databases to a local cache only long enough
to read recent rows. It persists domains, categories, visit counts, and hashed
YouTube video IDs when policy requests video metadata. It does not persist raw
URLs, query strings, page titles, cookies, tokens, or browser credentials.
Local verification scripts before Phase 3 disable browser history collection so
they do not accidentally archive live operator browsing domains.

Phase 20 exposes these boundaries in the tenant consent center. The consent API
and dashboard mark application usage metadata, browser domain/category activity,
device health, and archive health as collected or derived metadata. Passwords,
credentials, screenshots, private messages, cookies, tokens, camera, and
microphone are shown as denied collection categories.

Phase 43 dashboard layout verification uses browser layout metrics only. The
Playwright helper checks viewport overflow, required panel visibility, command
navigation targets, and text fit, then writes a JSON report under
`data/local/dashboard-layout/`. It does not capture screenshots, videos,
credentials, raw page content, browser history, or personal data.
