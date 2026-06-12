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
GET  /api/v1/devices/{deviceId}/summary/daily
GET  /api/v1/devices/{deviceId}/reports/weekly
GET  /api/v1/devices/{deviceId}/policy-violations
GET  /api/v1/devices/{deviceId}/anomalies
GET  /api/v1/devices/{deviceId}/tamper-events
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
audit events. Durable backend storage remains a later SaaS phase.

The backend rejects non-local bind addresses to avoid exposing an
unauthenticated remote API during the foundation phase.
