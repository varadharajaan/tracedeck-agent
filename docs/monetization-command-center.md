# Monetization Command Center

Phase 25 adds a typed monetization summary contract and a stronger first-screen
dashboard for paid-plan demos.

## API

```text
GET /api/v1/tenants/{tenantId}/monetization-summary
```

The response includes:

- customer plan, audience, seat usage, conversion stage, and revenue health
- readiness, notification, and trust scores
- email, push, and dashboard notification promise lines
- route-level proof with provider, recipient, status, attempts, retry, and next
  action
- paid capability proof for weekly reports, alert rules, role dashboards,
  managed rollout, notification proof, and compliance export readiness
- conversion action queue for the next customer-facing steps

The endpoint aggregates existing privacy-aware metadata. It does not add
password, credential, screenshot, keystroke, raw URL, private message, camera,
or microphone collection.

## Dashboard

The embedded dashboard now adds:

- Monetisation Command Center
- Notification Guarantee
- Paid Feature Proof
- Conversion Action Queue

These panels appear near the top of the dashboard before deeper operational
tables. They make TraceDeck feel like a sellable endpoint productivity and risk
observability product rather than only a local monitoring utility.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase25.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase25.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase25.ps1
```
