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
POST /api/v1/devices/enroll
GET  /api/v1/devices
GET  /api/v1/devices/{deviceId}
GET  /api/v1/devices/{deviceId}/overview
GET  /api/v1/devices/{deviceId}/health
GET  /api/v1/devices/{deviceId}/summary/daily
GET  /api/v1/devices/{deviceId}/reports/weekly
GET  /api/v1/devices/{deviceId}/policy-violations
GET  /api/v1/devices/{deviceId}/anomalies
GET  /api/v1/devices/{deviceId}/tamper-events
GET  /api/v1/devices/{deviceId}/alert-deliveries
GET  /api/v1/policy-templates
GET  /api/v1/archive/status
GET  /
```

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

The backend rejects non-local bind addresses to avoid exposing an
unauthenticated remote API during the foundation phase.
