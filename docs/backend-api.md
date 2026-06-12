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
GET  /api/v1/tenants/{tenantId}/executive-console
GET  /api/v1/tenants/{tenantId}/notification-revenue-cockpit
GET  /api/v1/tenants/{tenantId}/provider-simulation-lab
POST /api/v1/tenants/{tenantId}/provider-simulation-lab
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
