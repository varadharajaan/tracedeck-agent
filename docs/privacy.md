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

Phase 44 delivery drilldown is metadata-only. Dry-run rehearsals update route
verification status, route summary, and audit metadata only. They do not send
live email or push payloads, and they do not collect or store provider secrets,
alert bodies, endpoint payloads, passwords, tokens, cookies, screenshots, or
private content.

Phase 46 delivery remediation is also metadata-only. It records typed recovery
plans, owner/SLA state, retry/check timing, and audit events for email, push,
and dashboard routes. It rejects live-send remediation modes and does not store
provider secrets, alert bodies, screenshots, passwords, cookies, tokens, raw
URLs, page titles, or private content.

Phase 47 premium notification command center is an aggregate view only. It
combines existing alert inbox, delivery proof, remediation, operations, and
monetisation metadata into a typed response for the dashboard. It does not add
collectors and does not store provider secrets, alert bodies, screenshots,
passwords, cookies, tokens, raw URLs, page titles, or private content.

Phase 49 notification preferences are typed policy metadata only. They store
channel choices, digest cadence, quiet hours, escalation owner labels,
study-safe suppression labels, paid tier, and retention evidence. They do not
store SMTP passwords, push endpoint secrets, provider credentials, alert bodies,
screenshots, raw URLs, page titles, cookies, tokens, or private content.

Phase 50 business dashboard is also metadata-only. It aggregates customer
health, anomaly categories, notification delivery proof, preference score,
archive/report readiness, paid package labels, and owner actions from existing
typed APIs. It does not add collectors and does not store provider secrets,
alert bodies, screenshots, passwords, cookies, tokens, raw URLs, page titles,
private content, or endpoint payloads.

Phase 51 delivery timeline is metadata-only. It exposes channel, provider
label, recipient label, status, attempts, retry timing, host label, event id,
safe summary, paid tier, and next action. It does not store SMTP passwords,
push endpoint secrets, provider credentials, alert bodies, screenshots, raw
URLs, page titles, cookies, tokens, private content, or endpoint payloads.

Phase 52 role experiences are metadata-only. They expose role labels,
dashboard scope, onboarding status, notification proof, archive/report
readiness, consent controls, paid-tier labels, and next actions. They do not
store SMTP passwords, push endpoint secrets, provider credentials, alert
bodies, screenshots, raw URLs, page titles, cookies, tokens, private content,
or endpoint payloads.
