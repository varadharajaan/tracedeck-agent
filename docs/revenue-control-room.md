# Revenue Control Room

Phase 27 makes the embedded dashboard read like a sellable TraceDeck command
center, not only a host monitoring page.

The new first-screen buyer layer includes:

- package fit from the selected tenant plan and audience
- paid proof from typed monetisation capabilities and readiness scores
- upgrade motion from tenant conversion stage and action queue
- renewal risk from customer health, host attention count, and average risk
- commercial lever from paid tier and value-panel evidence
- anomaly assurance from the highest-priority policy, anomaly, or tamper signal
- email delivery proof from alert-delivery routes
- push delivery proof from alert-delivery routes
- weekly report mail readiness, including PDF attachment readiness
- buyer outcome, last signal, and next action text

The dashboard does not add new collectors or sensitive collection behavior. It
only renders privacy-safe backend data already exposed through typed APIs:

- `/api/v1/tenants/{tenantId}/operations-summary`
- `/api/v1/tenants/{tenantId}/monetization-summary`
- `/api/v1/tenants/{tenantId}/notification-routes`
- `/api/v1/devices/{deviceId}/alert-deliveries`
- `/api/v1/devices/{deviceId}/reports/weekly`
- `/api/v1/devices/{deviceId}/overview`

Verification is scripted through:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase27.ps1
```

That verifier runs gofmt, backend API tests, live dashboard smoke, Newman, cross
platform builds, and the root artifact check.
