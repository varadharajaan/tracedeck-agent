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
GET  /api/v1/tenants/{tenantId}/consent-center
GET  /api/v1/tenants/{tenantId}/operations-summary
GET  /api/v1/tenants/{tenantId}/monetization-summary
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

The backend rejects non-local bind addresses to avoid exposing an
unauthenticated remote API during the foundation phase.
