# Consent And Audit Center

Phase 20 adds a tenant-scoped consent and audit center API plus dashboard
panels for legitimacy, trust reviews, and paid packaging.

## API

```text
GET /api/v1/tenants/{tenantId}/consent-center
```

The response includes:

- visible monitoring status
- pause-control readiness
- data export readiness
- delete-request readiness
- alert recipients for email, push, and dashboard delivery
- typed collection disclosures
- recent tenant and policy audit events

Sensitive categories are disclosed as denied collection capabilities:

- passwords and credentials
- screenshots
- private messages, cookies, tokens, camera, and microphone

## Dashboard

The dashboard adds:

- Consent + Audit Center
- Policy Audit Trail
- Alert Revenue Operations
- Push Notification Center

The alert operations panels make the paid value visible: anomaly coverage,
email delivery proof, push notification reach, and audit evidence. They use the
existing typed risk events and alert delivery rows; they do not add sensitive
collection.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase20.ps1
```
