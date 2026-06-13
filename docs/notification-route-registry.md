# Notification Route Registry

Phase 26 adds a tenant-scoped route registry for email, push, and dashboard
delivery readiness.

## API

```text
GET  /api/v1/tenants/{tenantId}/notification-routes
POST /api/v1/tenants/{tenantId}/notification-routes
GET  /api/v1/tenants/{tenantId}/notification-preferences
POST /api/v1/tenants/{tenantId}/notification-preferences
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
- Notification Preference Center

These panels show configured routes, enabled count, verification count, and
routes needing attention before a paid customer demo.

Phase 44 adds provider-safe delivery drilldown on top of the registry:

```text
GET  /api/v1/tenants/{tenantId}/delivery-drilldown
POST /api/v1/tenants/{tenantId}/delivery-drilldown
```

The POST endpoint supports `mode: "dry_run"` only. It rehearses route metadata
for email, push, or dashboard delivery, updates route verification status, and
records an audit event. It does not send live messages and does not store SMTP
passwords, push endpoint secrets, alert bodies, tokens, or endpoint payloads.

Phase 46 adds delivery remediation on top of route registry and drilldown:

```text
GET  /api/v1/tenants/{tenantId}/delivery-remediation
POST /api/v1/tenants/{tenantId}/delivery-remediation
```

The remediation endpoint returns route recovery actions with owner, SLA target,
provider state, latest delivery status, next retry/check time, audit state, and
recent dry-run plans. POST supports `mode: "dry_run"` only and typed actions
such as `retry_plan`, `owner_ack`, `sla_watch`, `enable_route`, `fix_provider`,
`run_rehearsal`, and `maintain_proof`.

Remediation is planning and audit proof only. It does not send live mail or push
payloads, and it does not store provider secrets, SMTP passwords, push endpoint
secrets, alert bodies, screenshots, tokens, cookies, or raw URLs.

Phase 47 adds a premium command aggregate on top of the registry, drilldown,
remediation, and alert inbox:

```text
GET /api/v1/tenants/{tenantId}/notification-command-center
```

The response packages open alert counts, high-priority alert counts, per-channel
email/push/dashboard proof, route proof state, remediation SLA counts,
paid-tier labels, and owner actions. It is designed for the dashboard Notify Pro
view and remains metadata-only: no provider secrets, SMTP passwords, push
endpoint secrets, alert bodies, screenshots, tokens, cookies, raw URLs, page
titles, or private content are stored.

Phase 49 adds a notification preference center on top of routes and alert
rules:

```text
GET  /api/v1/tenants/{tenantId}/notification-preferences
POST /api/v1/tenants/{tenantId}/notification-preferences
```

The preference center stores digest cadence, quiet hours, escalation metadata,
and immediate/digest/silent rules. It can mark study-safe suppression rules so
learning-related activity is quieted while high-risk and tamper signals keep
their urgent channel policy. It remains metadata-only: no provider secrets,
SMTP passwords, push endpoint secrets, alert bodies, screenshots, raw URLs,
page titles, cookies, tokens, or private content are stored.

Phase 75 adds delivery assurance on top of routes, preferences, deliveries, and
dashboard fallback:

```text
GET /api/v1/tenants/{tenantId}/delivery-assurance
```

The endpoint separates `provider_confirmed`, `dry_run_rehearsed`,
`dashboard_visible`, `demo_only`, `retrying`, `failed`, `route_disabled`, and
`pending_provider` states. This prevents demo rows from appearing as real SMTP
or web-push delivery. A buyer-ready route requires provider-confirmed email,
provider-confirmed push, and dashboard fallback visibility. The response keeps
only route metadata, source labels, proof labels, timestamps, and recommended
next actions; it does not expose provider secrets, SMTP passwords, push
endpoints, alert bodies, tokens, cookies, screenshots, raw URLs, or page titles.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase26.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase26.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase26.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase75.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase75.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase75.ps1
```
