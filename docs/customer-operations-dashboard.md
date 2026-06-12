# Customer Operations Dashboard

Phase 23 adds a tenant-level operations summary for the monetisation dashboard.
It is designed for Family Pro, school, coaching center, and small-business
buyers who need to know whether TraceDeck noticed the right signals and
delivered the right notifications.

## API

```text
GET /api/v1/tenants/{tenantId}/operations-summary
```

The response is typed and includes:

- host count and hosts needing attention
- average tenant risk score
- open policy violations, anomalies, tamper signals, and archive backlog
- email, push, and dashboard delivery counts
- latest email and push delivery proof
- notification delivery score
- monetisation readiness score
- priority escalation signals
- upgrade proof signals for family, school, and business packaging

The endpoint aggregates existing privacy-aware metadata. It does not add new
collectors and does not collect passwords, credentials, screenshots, raw URLs,
private messages, cookies, tokens, camera, or microphone content.

## Dashboard

The embedded dashboard now includes:

- Customer Operations Cockpit
- Escalation Workbench
- Notification Delivery Board
- Upgrade Proof Pack

These panels sit near the top of the dashboard so the product story is visible
before deeper technical tables. They show host-level filtering context, anomaly
pressure, mail delivery proof, push notification reach, and customer-ready paid
plan value.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase23.ps1
```
