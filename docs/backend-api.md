# Backend API

Phase 5 adds a lightweight Go backend foundation using `net/http`.

The development server binds to localhost by default:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-backend-dev.ps1
```

Default base URL:

```text
http://127.0.0.1:18080
```

Backend state is durable by default in Phase 11:

```text
data/local/backend/backend-state.json
```

The backend can still run in memory only with:

```powershell
go run ./backend/cmd/tracedeck-backend --memory-store
```

Optional local API-key protection can be enabled with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-backend-dev.ps1 `
  -ApiKey "local-secret" `
  -ApiKeyTenantId "family-varadha"
```

When enabled, API routes require:

```text
X-TraceDeck-API-Key: local-secret
X-TraceDeck-Tenant-ID: family-varadha
```

`GET /health` and `GET /` remain available for local status and dashboard shell
loading. Tenant-scoped API keys can only create, list, and read resources for
their configured tenant.

Current endpoints:

```text
GET  /health
GET  /api/v1/version
GET  /api/v1/plans
GET  /api/v1/roles
GET  /api/v1/retention-tiers
GET  /api/v1/audit-events
GET  /api/v1/account-portfolio-index
GET  /api/v1/runtime-status-center
GET  /api/v1/verification-evidence-center
GET  /api/v1/operator-assurance-center
GET  /api/v1/promotion-readiness-center
POST /api/v1/tenants
GET  /api/v1/tenants
GET  /api/v1/tenants/{tenantId}
GET  /api/v1/tenants/{tenantId}/audit-events
GET  /api/v1/tenants/{tenantId}/alert-rules
POST /api/v1/tenants/{tenantId}/alert-rules
GET  /api/v1/tenants/{tenantId}/notification-routes
POST /api/v1/tenants/{tenantId}/notification-routes
GET  /api/v1/tenants/{tenantId}/notification-preferences
POST /api/v1/tenants/{tenantId}/notification-preferences
GET  /api/v1/tenants/{tenantId}/consent-center
GET  /api/v1/tenants/{tenantId}/operations-summary
GET  /api/v1/tenants/{tenantId}/monetization-summary
GET  /api/v1/tenants/{tenantId}/business-dashboard
GET  /api/v1/tenants/{tenantId}/role-experiences
GET  /api/v1/tenants/{tenantId}/customer-control-room
GET  /api/v1/tenants/{tenantId}/customer-success-packet
GET  /api/v1/tenants/{tenantId}/executive-console
GET  /api/v1/tenants/{tenantId}/notification-revenue-cockpit
GET  /api/v1/tenants/{tenantId}/provider-simulation-lab
POST /api/v1/tenants/{tenantId}/provider-simulation-lab
GET  /api/v1/tenants/{tenantId}/notification-provider-setup
GET  /api/v1/tenants/{tenantId}/package-billing-readiness
GET  /api/v1/tenants/{tenantId}/onboarding-center
GET  /api/v1/tenants/{tenantId}/customer-settings-center
GET  /api/v1/tenants/{tenantId}/revenue-operations-center
GET  /api/v1/tenants/{tenantId}/notification-command-center
GET  /api/v1/tenants/{tenantId}/delivery-timeline
GET  /api/v1/tenants/{tenantId}/sync-health
GET  /api/v1/tenants/{tenantId}/activity-feed
GET  /api/v1/tenants/{tenantId}/data-exports
POST /api/v1/tenants/{tenantId}/data-exports
GET  /api/v1/tenants/{tenantId}/delete-requests
POST /api/v1/tenants/{tenantId}/delete-requests
GET  /api/v1/tenants/{tenantId}/device-groups
POST /api/v1/tenants/{tenantId}/device-groups
GET  /api/v1/tenants/{tenantId}/policy-assignments
POST /api/v1/tenants/{tenantId}/policy-assignments
POST /api/v1/devices/enroll
GET  /api/v1/devices
GET  /api/v1/devices/{deviceId}
GET  /api/v1/devices/{deviceId}/overview
GET  /api/v1/devices/{deviceId}/health
GET  /api/v1/devices/{deviceId}/summary/daily
GET  /api/v1/devices/{deviceId}/reports/weekly
GET  /api/v1/devices/{deviceId}/reports/weekly/pdf
GET  /api/v1/devices/{deviceId}/policy-violations
GET  /api/v1/devices/{deviceId}/anomalies
GET  /api/v1/devices/{deviceId}/tamper-events
GET  /api/v1/devices/{deviceId}/alert-deliveries
POST /api/v1/devices/{deviceId}/telemetry-events
GET  /api/v1/devices/{deviceId}/telemetry-status
GET  /api/v1/policy-templates
GET  /api/v1/alert-rule-templates
GET  /api/v1/archive/status
GET  /
```

Telemetry ingest accepts metadata-only agent events. Phase 30 makes ingest
idempotent for stable non-empty event IDs, so replaying the same
`local-event-*` payload is acknowledged without creating duplicate stored
telemetry rows.

Tenant creation request:

```json
{
  "tenant_id": "family-varadha",
  "name": "Family Varadha",
  "plan_id": "family_pro",
  "retention_tier_id": "family_cloud_90_365_archive",
  "primary_profile": "ai-btech-student"
}
```

Device enrollment request:

```json
{
  "tenant_id": "family-varadha",
  "device_id": "laptop-cousin-001",
  "host_name": "study-laptop",
  "profile": "ai-btech-student",
  "os_name": "windows"
}
```

Phase 5 storage is intentionally in-memory. Phase 6 keeps that boundary while
adding SaaS readiness catalogs for tenants, plans, roles, retention tiers, and
audit events.

Phase 9 adds typed host dashboard APIs:

- host overview with device, summary, risk score, risk level, archive health,
  device health, policy violations, anomalies, tamper events, and alert
  deliveries
- policy violation events
- anomaly events
- tamper and trust events
- alert delivery routes for email, push, and dashboard channels

Phase 9 introduced in-memory demo risk data for enrolled devices. Phase 11
persists backend tenants, devices, audit events, host risk events, tamper events,
and alert delivery rows to the local backend state file. Remote auth, durable
database migrations, and production push-provider delivery history remain later
SaaS phases.

Phase 12 adds `GET /api/v1/devices/{deviceId}/health` and includes the same
typed `DeviceHealth` payload inside host overview. The score is an aggregate
endpoint-health model for CPU, memory, disk, heartbeat, and operational
recommendations.

Phase 14 turns the weekly report route into a generated report from host
overview data and adds a PDF packaging endpoint at
`GET /api/v1/devices/{deviceId}/reports/weekly/pdf`.

Phase 19 adds no-code alert rule builder APIs:

- `GET /api/v1/alert-rule-templates`
- `GET /api/v1/tenants/{tenantId}/alert-rules`
- `POST /api/v1/tenants/{tenantId}/alert-rules`

Saved rules are tenant-scoped, persisted in local backend state, and include
typed trigger, severity, delivery channels, condition, enabled flag, and
timestamps.

Phase 20 adds `GET /api/v1/tenants/{tenantId}/consent-center` for a
tenant-scoped trust center. It returns visible monitoring status, pause-control
readiness, data export/delete readiness, alert recipients, collection
disclosures, and recent tenant/policy audit events. Passwords, credentials,
screenshots, private messages, cookies, tokens, camera, and microphone are
reported as denied collection categories.

Phase 21 adds tenant-scoped managed rollout APIs:

- `GET /api/v1/tenants/{tenantId}/device-groups`
- `POST /api/v1/tenants/{tenantId}/device-groups`
- `GET /api/v1/tenants/{tenantId}/policy-assignments`
- `POST /api/v1/tenants/{tenantId}/policy-assignments`

New tenants receive a seeded primary device group and policy assignment.
Creating groups or assignments records tenant audit events. Assignments use
typed target values (`tenant`, `device_group`, `device`) and typed modes
(`audit`, `active`).

Phase 22 adds data rights workflow APIs for tenant exports and delete requests.
Exports create ready manifests; delete requests are queued for review and audit
instead of silently deleting data.

Phase 23 adds `GET /api/v1/tenants/{tenantId}/operations-summary` for a
tenant-level customer operations view. It rolls up host count, hosts needing
attention, anomaly/policy/tamper pressure, archive backlog, mail/push/dashboard
delivery counts, latest mail and push delivery proof, notification score,
monetisation readiness, priority escalation signals, and upgrade proof signals.

Phase 25 adds `GET /api/v1/tenants/{tenantId}/monetization-summary` for a
customer-ready product cockpit. It returns plan packaging, conversion stage,
readiness/notification/trust scores, email/push/dashboard promise lines,
route-level delivery proof, paid capability proof, and conversion actions for
Family Pro, school, and business demos.

Phase 26 adds tenant-scoped notification route registry APIs:

- `GET /api/v1/tenants/{tenantId}/notification-routes`
- `POST /api/v1/tenants/{tenantId}/notification-routes`

Routes are typed records for email, push, and dashboard delivery readiness.
They store channel, provider, enabled state, status, recipient label, route
summary, and verification timestamp. They do not store SMTP passwords, API
keys, push endpoint secrets, auth tokens, or raw message content.

Phase 49 adds tenant-scoped notification preference APIs:

- `GET /api/v1/tenants/{tenantId}/notification-preferences`
- `POST /api/v1/tenants/{tenantId}/notification-preferences`

The `GET` route returns a typed preference center with digest cadence, quiet
hours, escalation policy, immediate/digest/silent rules, channel coverage,
study-safe suppression count, route proof gaps, paid tier, retention evidence,
and a privacy boundary. The `POST` route accepts typed metadata-only updates
for cadence, quiet hours, escalation, and rules. Validation rejects unknown
channels, severities, preference modes, and digest cadences. It never stores
passwords, provider secrets, alert bodies, screenshots, tokens, cookies, raw
URLs, page titles, or private content.

Phase 44 adds provider-safe delivery drilldown APIs:

- `GET /api/v1/tenants/{tenantId}/delivery-drilldown`
- `POST /api/v1/tenants/{tenantId}/delivery-drilldown`

The `GET` route returns a typed route score, channel readiness flags, per-route
proof state, latest delivery metadata, dry-run rehearsal result, SLA wording,
and next actions for email, push, and dashboard routes. The `POST` route accepts
`mode: "dry_run"` with an optional `channel` filter and updates route
verification metadata plus an audit event. It never sends live provider
messages and never stores SMTP passwords, push endpoint secrets, alert bodies,
tokens, cookies, screenshots, or endpoint payloads.

Phase 46 adds provider-safe delivery remediation APIs:

- `GET /api/v1/tenants/{tenantId}/delivery-remediation`
- `POST /api/v1/tenants/{tenantId}/delivery-remediation`

The `GET` route returns a typed remediation summary, route recovery actions,
owner labels, SLA targets, next retry windows, route/provider state, recent
dry-run remediation plans, audit state, and a privacy boundary. The `POST`
route accepts `mode: "dry_run"` only with typed actions such as `retry_plan`,
`owner_ack`, `sla_watch`, `enable_route`, `fix_provider`, `run_rehearsal`, and
`maintain_proof`. It records a provider-safe plan and audit event. It never
sends live provider messages and never stores provider secrets, alert bodies,
tokens, cookies, passwords, screenshots, raw URLs, or endpoint payloads.

Phase 47 adds `GET /api/v1/tenants/{tenantId}/notification-command-center`.
It aggregates the tenant operations summary, monetisation summary, alert inbox,
delivery drilldown, and delivery remediation into one typed buyer-facing
notification command contract. The response includes notification score,
monetisation readiness, open alert counts, high-priority counts, email/push/
dashboard delivery proof, route proof state, remediation SLA counts, paid-tier
labels, and owner action items. It is metadata-only and does not store provider
secrets, alert bodies, screenshots, passwords, tokens, cookies, raw URLs, page
titles, or private content.

Phase 50 adds `GET /api/v1/tenants/{tenantId}/business-dashboard`. It composes
operations, monetisation, alert inbox, notification command center,
notification preferences, delivery drilldown, and delivery remediation into one
typed product contract for the dashboard. The response includes product score,
customer health, revenue stage, recommended package, host attention, open and
high-priority alerts, mail/push/dashboard delivery proof, route proof gaps,
archive backlog, weekly report readiness, KPI tiles, alert rows, channel rows,
paid package cards, and customer owner actions. It is metadata-only and does
not store provider secrets, alert bodies, screenshots, passwords, tokens,
cookies, raw URLs, page titles, private content, or endpoint payloads.

Phase 51 adds `GET /api/v1/tenants/{tenantId}/delivery-timeline`. Query
parameters are typed by the backend filter model: `device_id`, `channel`,
`status`, `provider`, `q`, and `limit`. The response rolls up provider-safe
email, push, and dashboard delivery evidence into notification score, route
proof gaps, retry timing, last delivered time, source host count, paid-tier
recommendation, and timestamped delivery rows. It is metadata-only and does not
store provider secrets, alert bodies, screenshots, passwords, tokens, cookies,
raw URLs, page titles, private content, or endpoint payloads.

Phase 52 adds `GET /api/v1/tenants/{tenantId}/role-experiences`. It packages
parent, student, school admin, and business manager dashboard experiences into
one typed onboarding contract. The response includes role readiness score,
ready role count, notification score, trust score, recommended package,
per-role visible panels, notification promise, archive/report promise, consent
controls, paid tier, next action, and role-specific onboarding items. It is
metadata-only and does not store provider secrets, alert bodies, screenshots,
passwords, tokens, cookies, raw URLs, page titles, private content, or endpoint
payloads.

Phase 53 adds `GET /api/v1/tenants/{tenantId}/executive-console`. It composes
operations, monetisation, business dashboard, role experience, notification
command center, and delivery timeline proof into a typed first-screen product
contract. The response includes sellable readiness, anomaly/open alert counts,
mail/push/dashboard delivery proof, route proof gaps, weekly report readiness,
archive backlog, role-view readiness, recommended paid package, next best
action, value tiles, alert stream rows, delivery proof rows, owner actions, and
a strict metadata-only privacy boundary. It does not store provider secrets,
alert bodies, screenshots, passwords, tokens, cookies, raw URLs, page titles,
private content, or endpoint payloads.

Phase 54 adds `GET /api/v1/tenants/{tenantId}/notification-revenue-cockpit`.
It composes operations, monetisation, notification command center,
notification preferences, and delivery timeline proof into a typed
buyer-facing notification revenue contract. The response includes revenue
readiness, notification score, anomaly SLA readiness, mail/push/dashboard
delivery proof, route proof gaps, weekly report readiness, escalation state,
recommended paid package, next best action, KPI proof rows, channel proof
matrix, anomaly delivery scenarios, upgrade action levers, and a strict
metadata-only privacy boundary. It does not store provider secrets, alert
bodies, screenshots, passwords, tokens, cookies, raw URLs, page titles,
private content, or endpoint payloads.

Phase 55 adds `GET /api/v1/tenants/{tenantId}/provider-simulation-lab` and
`POST /api/v1/tenants/{tenantId}/provider-simulation-lab`. The GET response
packages provider-safe email, push, and dashboard simulation readiness with
route proof, SLA state, scenario templates, action levers, paid package value,
and strict metadata-only privacy proof. The POST endpoint accepts typed
`dry_run` simulations by channel and records audit proof; live send modes are
rejected. It does not store provider secrets, SMTP passwords, push endpoint
payloads, alert bodies, screenshots, passwords, tokens, cookies, raw URLs, page
titles, private content, or endpoint payloads.

Phase 77 adds `GET /api/v1/tenants/{tenantId}/notification-provider-setup`.
It composes delivery assurance and provider simulation metadata into a typed
setup readiness contract for email, push, and dashboard routes. The response
separates configured routes, provider-confirmed proof, dry-run proof, demo-only
rows, retrying routes, checklist state, and owner actions so demo evidence is
not mistaken for actual email or push delivery. It stores metadata only and
does not collect provider secrets, SMTP passwords, push endpoints, raw provider
payloads, alert bodies, screenshots, raw URLs, page titles, tokens, cookies,
private content, endpoint payloads, or passwords.

Phase 56 adds `GET /api/v1/tenants/{tenantId}/package-billing-readiness`.
The response composes operations, monetisation, business dashboard, role
experience, provider simulation, plan, and retention metadata into a typed
buyer-facing package contract. It includes package score, billing status,
current and recommended package, seat usage, feature gate readiness, retention
and archive readiness, weekly report readiness, notification/provider proof,
plan fit rows, billing milestones, upgrade actions, and a strict metadata-only
privacy boundary. It does not collect or store payment card data, invoices,
provider secrets, passwords, screenshots, raw URLs, page titles, alert bodies,
tokens, cookies, private content, or endpoint payloads.

Phase 57 adds `GET /api/v1/tenants/{tenantId}/customer-control-room`. It
composes operations, business dashboard, executive console, package billing,
and provider simulation metadata into a single typed first-screen product
contract. The response includes customer-ready score, notification score,
package score, trust score, anomaly/open alert counts, mail/push/dashboard
delivery proof, route proof gaps, report/archive readiness, billing/provider
readiness, value tiles, alert rows, delivery rows, owner actions, and a strict
metadata-only privacy boundary. It does not collect or store passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint
payloads, private content, or payment card data.

Phase 58 adds `GET /api/v1/tenants/{tenantId}/customer-success-packet`. It
composes the customer control room, package billing readiness, provider
simulation, and role experience metadata into a typed buyer/admin review
packet. The response includes readiness, notification, package, and trust
scores; open and high-priority alert counts; host count; mail and push proof;
route proof gaps; report/archive readiness; provider and billing readiness;
role readiness; proof rows; buyer objection answers; owner actions; and a
strict metadata-only privacy boundary. It does not collect or store passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, invoices, private content, or payment card data.

Phase 59 adds `GET /api/v1/tenants/{tenantId}/push-activation-center`. It
composes operations, notification preferences, delivery drilldown, provider
simulation, delivery remediation, alert inbox, push delivery timeline, and
package billing metadata into a typed push activation contract. The response
includes activation and notification scores; mail fallback and dashboard
fallback counts; delivered/retrying/failed/pending push counts; route proof
readiness; rules and alert counts using push; preference, escalation, quiet
hours, and simulation readiness; push route rows; anomaly notification
scenarios; owner actions; and a strict metadata-only privacy boundary. It does
not collect or store passwords, screenshots, raw URLs, page titles, alert
bodies, provider secrets, push endpoints, endpoint payloads, private content,
invoices, tokens, cookies, payment card data, or raw provider payloads.

Phase 60 adds `GET /api/v1/tenants/{tenantId}/portfolio-center`. It composes
operations, business dashboard, sync health, alert inbox, delivery timeline,
package billing, and host overview metadata into a typed multi-host portfolio
contract. The response includes portfolio, notification, trust, and risk
scores; host reporting and pending counts; open/high-priority alerts; typed
alert notification rows with mail, push, and dashboard statuses; delivery proof
cards for mail, push, dashboard fallback, weekly report/archive, and host
coverage; archive backlog; route proof gaps; host rows with
health/risk/count/delivery status; portfolio segments; owner actions; and a
strict metadata-only privacy boundary. It does not collect or store
passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets,
push endpoints, endpoint payloads, private content, invoices, tokens, cookies,
payment card data, or raw provider payloads.

Phase 61 adds `GET /api/v1/account-portfolio-index`. It composes visible
tenant Portfolio Center contracts into a typed account-level index for
multi-tenant admins and customer owners. The response includes account,
notification, and trust scores; tenant and host coverage; open/high-priority
alerts; mail, push, and dashboard delivery proof; archive backlog; route proof
gaps; tenant rows with plan, audience, scores, alert counts, delivery counts,
and next action; account proof cards; owner actions; and a strict metadata-only
privacy boundary. Tenant-scoped API keys only see rows for their configured
tenant. It does not collect or store passwords, screenshots, raw URLs, page
titles, alert bodies, provider secrets, push endpoints, endpoint payloads,
private content, invoices, tokens, cookies, payment card data, or raw provider
payloads.

Phase 31 adds `GET /api/v1/tenants/{tenantId}/sync-health` for buyer and admin
proof that backend-visible telemetry is arriving per host. It returns reporting
host counts, stored metadata event totals, the highest stable `local-event-*`
cursor received, last ingest time, source counts for process/browser/health
metadata, offline replay readiness, per-host recommendations, and the
metadata-only privacy boundary. It cannot see local rows that have not synced
yet; it proves what the backend has safely received.

Phase 32 adds `GET /api/v1/tenants/{tenantId}/activity-feed` for cross-host
command triage. Query parameters are typed by the backend filter model:
`device_id`, `kind`, `severity`, `channel`, `status`, `q`, and `limit`.
The feed rolls up policy/anomaly/tamper risk items, alert delivery items, and
backend-visible telemetry metadata items into one time-ordered response with
summary counts for risk, delivery, sync, email proof, push retry, reporting
hosts, and source hosts.

Phase 33 adds tenant-scoped saved command views:

- `GET /api/v1/tenants/{tenantId}/activity-views`
- `POST /api/v1/tenants/{tenantId}/activity-views`

New tenants receive seeded monetisation views for open high-risk anomalies,
mail delivery proof, push retry watch, and sync/archive proof. Created views are
persisted, audited, and validated against typed feed kind, severity, channel,
status, limit, and paid-tier values.

Phase 34 improves the embedded dashboard's local API-key workflow. The
dashboard route remains loadable from localhost, while protected API calls keep
using `X-TraceDeck-API-Key`, `X-TraceDeck-Tenant-ID`, and
`X-TraceDeck-Actor-ID`. The browser shell stores the entered API key only in
`sessionStorage` and sends it as a header; it is not embedded in dashboard HTML
or backend state.

The backend rejects non-local bind addresses to avoid exposing an
unauthenticated remote API during the foundation phase.

Phase 63 adds `GET /api/v1/tenants/{tenantId}/onboarding-center`. It composes
role experience, package billing, portfolio, push activation, notification
preferences, and sync-health metadata into a typed tenant activation contract.
The response includes setup checklist status, install proof, autostart/reboot
persistence proof, notification policy proof, mail/push/dashboard delivery
proof, archive retention posture, role dashboard handoff rows, package
readiness, proof cards, owner actions, and a strict metadata-only privacy
boundary. It does not collect or store passwords, screenshots, raw URLs, page
titles, alert bodies, provider secrets, push endpoints, endpoint payloads,
private content, invoices, tokens, cookies, payment card data, or raw provider
payloads.

Phase 64 adds `GET /api/v1/tenants/{tenantId}/customer-settings-center`. It
composes plan, retention, package billing, onboarding, notification
preferences, notification routes, and role experience metadata into a typed
customer-facing settings contract. The response includes settings score,
current and recommended plan/retention settings, notification policy state,
mail/push/dashboard channel settings, archive/autostart/role/data-rights
settings, owner actions, plan options, retention options, and a strict
metadata-only privacy boundary. It does not collect or store passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, private content, invoices, tokens, cookies,
payment card data, or raw provider payloads.

Phase 65 adds `GET /api/v1/tenants/{tenantId}/revenue-operations-center`. It
composes customer control, customer success, push activation, portfolio,
onboarding, customer settings, package billing, and provider simulation
metadata into a typed monetisation operations contract. The response includes
revenue, product, notification, trust, package, settings, and onboarding
scores; open/high-priority alert counts; host coverage; mail, push, and
dashboard delivery proof; report/archive readiness; provider and billing
readiness; revenue signal rows; anomaly/delivery wall rows; delivery route
proof rows; commercial levers; owner actions; and a strict metadata-only
privacy boundary. It does not collect or store passwords, screenshots, raw
URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint
payloads, private content, invoices, tokens, cookies, payment card data, or raw
provider payloads.

Phase 66 adds `GET /api/v1/tenants/{tenantId}/deployment-readiness-center`. It
composes onboarding, customer settings, sync-health, portfolio, and revenue
operations metadata into a typed deployment readiness contract. The response
includes deployment readiness score, Windows/macOS/Linux platform rows, Task
Scheduler/launchd/systemd manifest rows, live boot and autostart status,
background start readiness, offline replay readiness, archive backlog, owner
actions, and a strict metadata-only privacy boundary. It does not collect or
store passwords, screenshots, raw URLs, page titles, alert bodies, provider
secrets, push endpoints, endpoint payloads, private content, invoices, tokens,
cookies, payment card data, raw provider payloads, keylogging, or hidden
collection bypasses.

Phase 98 adds `GET /api/v1/runtime-status-center`. It reads the local
`data/local/output/runtime-summary.json` artifact generated by
`python ./devctl.py summary` and returns typed status, proof rows, operator
actions, and the metadata-only privacy boundary. If the summary file is
missing, the endpoint returns an action state telling the operator to generate
the summary instead of failing the dashboard.

Phase 99 adds `GET /api/v1/verification-evidence-center`. It reads
`data/local/output/verification-evidence.json` generated by
`python ./devctl.py evidence` or
`scripts/local/get-verification-evidence.ps1` and returns typed scripted gate
rows, report/log path references, git branch/head labels, proof rows, operator
actions, and a strict metadata-only privacy boundary. If the artifact is
missing, the endpoint returns an action state that tells the operator to
generate evidence instead of pretending verification proof exists.

Phase 67 adds `GET /api/v1/tenants/{tenantId}/premium-operations-hub`. It
composes revenue operations, deployment readiness, and portfolio metadata into
a typed premium command contract. The response includes premium, revenue,
deployment, portfolio, notification, and trust scores; anomaly/open-alert
counts; host coverage; mail delivery proof; push notification route state;
dashboard fallback proof; weekly report readiness; archive backlog; deployment
readiness; package recommendation; premium value tiles; anomaly notification
rows; delivery route rows; commercial levers; owner actions; and a strict
metadata-only privacy boundary. It does not collect or store passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, private content, invoices, tokens, cookies,
payment card data, raw provider payloads, keylogging, or hidden collection
bypasses.

Phase 68 adds `GET /api/v1/tenants/{tenantId}/browser-activity`. It exposes a
typed browser activity viewer contract with filters for `device_id`, `browser`,
`category`, `domain`, `study_safe`, `q`, and `limit`. The response includes
Chrome, Edge, Brave, study-safe, non-study YouTube, notification-proof, host,
browser, and activity-row summaries for domain-level telemetry emitted by the
agent browser collector. Activity rows include `source_kind`, `evidence_scope`,
and `evidence_detail` so `demo_seed`, `live_ingested`, and cloud sampled rows
can be separated in dashboards and tests. It remains metadata-only and does not collect or store
passwords, screenshots, raw URLs, page titles, cookies, tokens, private
content, endpoint payloads, keylogging, hidden collection bypasses, provider
secrets, push endpoints, or alert bodies.

Phase 73 adds first-class verification for those provenance fields across local
API rows, alert delivery rows, the browser activity page, the main dashboard,
and Lambda S3 summary rows.

Phase 74 does not add a new backend route. It adds `python ./devctl.py doctor`,
which verifies existing backend and cloud contracts from the operator machine:
local `/health`, dashboard HTML, `/browser-activity`, tenant browser activity
rows, `/api/v1/devices`, device alert-delivery provenance, Lambda `/api/health`,
Lambda `/api/s3-summary`, and Lambda cache-hit behavior. The report is written
to `data/local/output/runtime-doctor.json` and
`data/local/output/runtime-doctor.txt`.

Phase 75 adds `GET /api/v1/tenants/{tenantId}/delivery-assurance`. It returns a
typed route and event truth contract for notification delivery proof. Query
filters are `device_id`, `channel`, `assurance_state`, and `limit`.

Assurance states are:

- `provider_confirmed`
- `dry_run_rehearsed`
- `dashboard_visible`
- `demo_only`
- `retrying`
- `failed`
- `route_disabled`
- `pending_provider`

The summary includes provider-confirmed email/push readiness, dashboard
fallback readiness, buyer-ready status, counts by assurance state, next action,
next retry time, last provider proof time, and a strict metadata-only privacy
boundary. Demo seed rows are never counted as provider proof.

Default host APIs are live-truth views. `GET /api/v1/devices/{deviceId}/overview`,
`/summary/daily`, `/policy-violations`, `/anomalies`, `/tamper-events`,
`/alert-deliveries`, `/reports/weekly`, and `/reports/weekly/pdf` suppress rows
with `source_kind=demo_seed`. This includes old persisted seeded rows such as
the VLC/media-playback sample and demo email/push proof rows. Demo evidence is
available only for explicit demos by passing `?include_demo=true`; callers must
not use that mode for live host reporting or delivery proof. Weekly report
`email_ready` requires a non-demo delivered email route; generated PDF/report
packaging alone is not a delivered email claim. Dashboard HTML, browser activity
HTML, and JSON API responses include no-store cache headers so stale clients do
not preserve old demo evidence in live views.

`GET /api/v1/tenants/{tenantId}/activity-feed` follows the same contract through
the typed `include_demo` filter. The default tenant and host-scoped activity
feed excludes `source_kind=demo_seed`, including the seeded VLC/media-playback
row. `?include_demo=true` is reserved for explicit demo scripts and labels the
response filter with `"include_demo": true`.

Phase 90 aligns `python ./devctl.py doctor` with that same truth boundary. The
doctor no longer requires a default alert-delivery row from seeded demo data.
Instead, it records `default_count`, `default_source_kinds`,
`default_demo_hidden`, `opt_in_demo_count`, `opt_in_demo_source_kinds`, and
`opt_in_demo_available` under `local.deliveries`. Isolated demo backends should
show `default_count=0`, `default_demo_hidden=true`, and
`opt_in_demo_available=true`; buyer-ready notification proof still requires
provider-confirmed delivery assurance, not demo-only rows.

Phase 78 adds `GET /api/v1/tenants/{tenantId}/notification-provider-setup`.
It returns a typed notification provider setup contract for dashboard rendering
and buyer readiness. The response includes configured versus provider-confirmed
email, push, and dashboard fallback setup; demo-only, retrying, pending-provider,
and dry-run counts; setup score; recommended paid package; next best action;
provider setup channels; checklist rows; owner actions; and a strict
metadata-only privacy boundary. It does not include provider secrets, SMTP
passwords, push endpoints, alert bodies, raw provider payloads, screenshots,
raw URLs, page titles, cookies, tokens, private content, or payment data.

Phase 69 does not add a new local backend route. It adds a SAM Lambda frontend
contract at `/api/health`, `/api/s3-summary`, and `/` inside `sam-app`. The
Lambda frontend uses S3 object listing and safe JSON/JSONL/GZip sampling to
summarize archive object counts, byte totals, host/browser/domain metadata,
source provenance, and cache hit/miss percentages. It is deployed with a Lambda Function URL and
does not define API Gateway resources.

Phase 83 extends the existing telemetry ingest/status and tenant sync-health
contracts with `agent.health.heartbeat` events from
`collector.agent.heartbeat`. The heartbeat event stores only typed readiness
metadata: profile, operating system, agent health, agent version, collection
mode, collection interval, archive enabled/due state, backend sync enabled
state, and alerts enabled state. `/api/v1/devices/{deviceId}/telemetry-status`
counts the event by type and source, and `/api/v1/tenants/{tenantId}/sync-health`
uses it as backend-visible agent health/replay proof. It does not include
passwords, screenshots, raw URLs, page titles, cookies, tokens, private
content, endpoint payloads, provider secrets, alert bodies, keylogs, or hidden
collection bypass data.

Phase 100 adds `GET /api/v1/operator-assurance-center`. It composes the local
runtime summary and verification evidence artifacts into typed assurance
summary fields, cards, and actions. It reports runtime readiness, Scheduler
readback explanation, verification gate counts, frontend artifact-cache state,
git hygiene, output paths, and privacy proof only. It does not collect or
return passwords, screenshots, raw browser URLs, page titles, cookies, tokens,
private content, endpoint payloads, provider secrets, alert bodies, keylogs,
payment data, or raw provider payloads.

Phase 102 extends the runtime summary artifact and
`GET /api/v1/runtime-status-center` with ready-file PID reconciliation fields:
`ready_pid`, `ready_pid_matches_live`, and `ready_pid_status`. Valid statuses
are `match`, `stale`, `absent`, and `unknown`. A stale ready-file PID adds the
`pid-reconciliation` proof row and `refresh-ready-pid-proof` operator action
with `watch` status, but does not override healthy live `pid_and_health`
runtime proof.

Phase 103 points that `refresh-ready-pid-proof` action at the runnable command
`python ./devctl.py server task-refresh-ready`. The command refreshes local
ready proof from live PID and `/health` metadata only; it does not restart the
backend and does not collect browser content, screenshots, credentials, alert
bodies, provider payloads, or private content. Runtime Status and Operator
Assurance action rows expose typed `evidence_scope=metadata_only` so clients can
separate metadata-only remediation from sensitive collection.

Phase 105 adds `GET /api/v1/promotion-readiness-center`. It composes Runtime
Status, Verification Evidence, and Operator Assurance into one publish/handoff
surface with status, readiness booleans, gate counts, Scheduler readback, ready
PID status, git hygiene, local export paths, proof rows, and actions. It reads
only existing local metadata proof artifacts and does not collect or return
passwords, screenshots, raw URLs, page titles, cookies, tokens, private
content, endpoint payloads, provider secrets, alert bodies, payment data,
keylogs, hidden collection bypasses, or raw provider payloads.
