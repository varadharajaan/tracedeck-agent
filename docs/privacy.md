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

Phase 53 executive console is metadata-only. It exposes product readiness,
anomaly categories, host labels, email/push/dashboard delivery proof, weekly
report readiness, archive status, role packaging, paid-tier labels, and owner
actions. It does not add collectors and does not store SMTP passwords, push
endpoint secrets, provider credentials, alert bodies, screenshots, raw URLs,
page titles, cookies, tokens, private content, or endpoint payloads.

Phase 54 notification revenue cockpit is metadata-only. It exposes anomaly SLA
categories, mail/push/dashboard delivery proof, weekly report readiness,
escalation state, scenario labels, channel value, paid-package levers, and
owner actions. It does not add collectors and does not store SMTP passwords,
push endpoint secrets, provider credentials, alert bodies, screenshots, raw
URLs, page titles, cookies, tokens, private content, or endpoint payloads.

Phase 55 provider simulation lab is metadata-only. It exposes route labels,
channel, provider type, delivery status, SLA result, retry posture, scenario
templates, buyer value, and owner actions. Dry-run simulation records audit
proof only and does not send live provider payloads. It does not store SMTP
passwords, push endpoint secrets, provider credentials, alert bodies,
screenshots, raw URLs, page titles, cookies, tokens, private content, or
endpoint payloads.

Phase 56 package billing readiness is metadata-only. It exposes plan labels,
feature gates, seat counts, retention tier, billing setup status, report and
archive value, notification proof, provider simulation proof, trust/data-rights
readiness, and upgrade actions. It does not collect or store payment card data,
invoices, provider secrets, passwords, screenshots, raw URLs, page titles,
alert bodies, tokens, cookies, private content, or endpoint payloads.

Phase 57 customer control room is metadata-only. It exposes host labels,
anomaly categories, notification route status, mail and push proof, provider
simulation state, report readiness, archive posture, package fit, customer
health, and owner actions. It does not add collectors and does not store
passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets,
endpoint payloads, private content, tokens, cookies, or payment card data.

Phase 58 customer success packet is metadata-only. It exposes customer-ready
scores, host count, anomaly categories, mail and push proof, route proof gaps,
report/archive readiness, package fit, provider simulation state, role
readiness, privacy assurances, buyer objection answers, and owner actions. It
does not add collectors and does not store passwords, screenshots, raw URLs,
page titles, alert bodies, provider secrets, push endpoints, endpoint payloads,
private content, invoices, tokens, cookies, or payment card data.

Phase 59 push activation center is metadata-only. It exposes push route labels,
subscription labels, proof state, retry posture, push delivery status, mail
fallback count, dashboard fallback count, notification preference coverage,
escalation coverage, quiet-hours status, provider-safe simulation status,
scenario labels, paid-tier labels, and owner actions. It does not add
collectors and does not store passwords, screenshots, raw URLs, page titles,
alert bodies, provider secrets, push endpoints, endpoint payloads, private
content, invoices, tokens, cookies, payment card data, or raw provider
payloads.

Phase 60 portfolio center is metadata-only. It exposes host labels, profiles,
OS labels, risk and health scores, policy/anomaly/tamper counts, notification
status, alert notification route states, delivery proof labels, archive
backlog, sync posture, paid-tier labels, portfolio segments, and owner actions.
It does not add collectors and does not store passwords, screenshots, raw URLs,
page titles, alert bodies, provider secrets, push endpoints, endpoint payloads,
private content, invoices, tokens, cookies, payment card data, or raw provider
payloads.

Phase 61 account portfolio index is metadata-only. It exposes tenant labels,
plan labels, host counts, score summaries, alert counts, notification proof,
archive backlog, package readiness, tenant-scoped row filtering, and owner
actions. It does not add collectors and does not store passwords, screenshots,
raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint
payloads, private content, invoices, tokens, cookies, payment card data, or raw
provider payloads.

Phase 62 monetisation overview is metadata-only. It composes existing dashboard
and account evidence into a first-screen commercial view: account labels, host
counts, scores, anomaly counts, mail/push/dashboard route status, weekly report
readiness, archive posture, package labels, owner actions, and privacy
guardrails. It does not add collectors and does not store passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, private content, invoices, tokens, cookies,
payment card data, or raw provider payloads.

Phase 63 tenant onboarding center is metadata-only. It exposes setup step
labels, host counts, autostart readiness, notification policy status,
mail/push/dashboard delivery counts, archive posture, role view labels, package
readiness, proof rows, and owner actions. It does not add collectors and does
not store passwords, screenshots, raw URLs, page titles, alert bodies, provider
secrets, push endpoints, endpoint payloads, private content, invoices, tokens,
cookies, payment card data, or raw provider payloads.

Phase 64 customer settings center is metadata-only. It exposes plan labels,
retention tier labels, notification preference status, route proof state, role
settings, archive posture, autostart readiness, data-rights state, and owner
actions. It does not add collectors and does not store passwords, screenshots,
raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint
payloads, private content, invoices, tokens, cookies, payment card data, or raw
provider payloads.

Phase 65 revenue operations center is metadata-only. It exposes tenant labels,
host counts, anomaly categories, notification delivery proof, mail/push/
dashboard route state, weekly report readiness, archive backlog, setup
readiness, settings readiness, package labels, commercial levers, owner
actions, and trust guardrails. It does not add collectors and does not store
passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets,
push endpoints, endpoint payloads, private content, invoices, tokens, cookies,
payment card data, or raw provider payloads.

Phase 66 deployment readiness center is metadata-only. It exposes tenant
labels, host counts, platform names, service manager labels, manifest paths,
dry-run command labels, autostart status, live boot labels, offline replay and
archive backlog counts, owner actions, and setup evidence. It does not add
collectors and does not store passwords, screenshots, raw URLs, page titles,
alert bodies, provider secrets, push endpoints, endpoint payloads, private
content, invoices, tokens, cookies, payment card data, raw provider payloads,
keylogging, or hidden collection bypasses.

Phase 67 premium operations hub is metadata-only. It exposes product scores,
host counts, anomaly categories, mail/push/dashboard delivery proof, weekly
report readiness, archive backlog, deployment readiness, package labels,
commercial levers, owner actions, and buyer-facing trust proof. It does not add
collectors and does not store passwords, screenshots, raw URLs, page titles,
alert bodies, provider secrets, push endpoints, endpoint payloads, private
content, invoices, tokens, cookies, payment card data, raw provider payloads,
keylogging, or hidden collection bypasses.

Phase 68 browser activity viewer is metadata-only. It exposes browser names,
host labels, domains, categories, study-safe flags, visit counts, timestamps,
and notification proof derived from existing browser-domain telemetry. It does
not store raw URLs, page titles, browser cookies, tokens, passwords,
screenshots, private content, endpoint payloads, provider secrets, push
endpoints, alert bodies, keylogging, or hidden collection bypasses.

Phase 75 delivery assurance is metadata-only. It exposes notification route
truth labels, delivery state counts, source labels, proof labels, retry timing,
and next actions. It explicitly marks seeded rows as `demo_only` and retrying
push rows as not screen-visible until provider proof exists. It does not store
provider secrets, SMTP passwords, push endpoints, raw provider payloads, alert
bodies, raw URLs, page titles, screenshots, private content, cookies, tokens,
passwords, keylogging, or hidden collection bypasses.

Phase 110 foreground app collection is metadata-only. It stores active app
name, process id, hashed executable path, foreground state, operating system,
profile, and `window_title_mode=none`. It does not collect screenshots, window
titles, raw URLs, page titles, cookies, tokens, passwords, keylogging, private
content, provider secrets, alert bodies, payment data, hidden collection
bypasses, or raw provider payloads.

Phase 111 software inventory collection is metadata-only. It stores software
display name, optional version/publisher display metadata, source label, and
hashed snapshot identity. It does not store install paths, file contents,
installer payloads, screenshots, passwords, raw URLs, page titles, cookies,
tokens, private content, provider secrets, alert bodies, keylogging data,
hidden collection bypasses, payment data, or raw provider payloads.
