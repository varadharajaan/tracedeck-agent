# Data Rights Workflows

Phase 22 adds auditable tenant data export and delete-request workflows.

## APIs

```text
GET  /api/v1/tenants/{tenantId}/data-exports
POST /api/v1/tenants/{tenantId}/data-exports
GET  /api/v1/tenants/{tenantId}/delete-requests
POST /api/v1/tenants/{tenantId}/delete-requests
```

Data exports create a ready manifest with format, scope, resource count,
storage key, requested-by, completion time, and expiry. Delete requests are
queued for review with reason, scope, status, due date, and audit history.

This phase does not silently delete tenant data. It creates a tracked request
that future hosted workflows can approve and execute with stronger auth.

## Dashboard

The dashboard adds:

- Data Export Center
- Delete Request Queue

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase22.ps1
```
