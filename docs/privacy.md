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
