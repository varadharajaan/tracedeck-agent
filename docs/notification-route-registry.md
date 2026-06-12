# Notification Route Registry

Phase 26 adds a tenant-scoped route registry for email, push, and dashboard
delivery readiness.

## API

```text
GET  /api/v1/tenants/{tenantId}/notification-routes
POST /api/v1/tenants/{tenantId}/notification-routes
```

Routes are typed records containing:

- channel: `email`, `push`, or `dashboard`
- provider: `smtp`, `web_push`, or `local_dashboard`
- enabled state
- route status
- recipient label
- last verification timestamp
- last route summary

Provider and channel pairs are validated. For example, `web_push` is valid only
for `push`, and `smtp` is valid only for `email`.

## Privacy Boundary

The registry stores route readiness metadata only. It does not store SMTP
passwords, AWS keys, push endpoint secrets, cookies, tokens, raw alert message
content, screenshots, credentials, or keystrokes.

## Dashboard

The embedded dashboard now includes:

- Notification Route Registry
- Route Readiness Proof

These panels show configured routes, enabled count, verification count, and
routes needing attention before a paid customer demo.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase26.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase26.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase26.ps1
```
