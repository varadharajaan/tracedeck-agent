# Dashboard

The Go backend serves an embedded dashboard from `/`.

To start a local dashboard with seeded host risk data:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1
```

Then open:

```text
http://127.0.0.1:18080/
```

Stop the demo backend with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1
```

Phase 9 expands the dashboard into a host-level command center for productivity,
risk, archive, and alert-delivery visibility. Phase 12 upgrades the same
embedded surface into a richer monetisation-ready operations dashboard with
device health, notification operations, product packaging, policy marketplace,
and retention plan panels. It remains a lightweight static HTML/CSS/JavaScript
asset embedded into the backend binary.

The dashboard reads the base backend endpoints:

- `/health`
- `/api/v1/devices`

For the selected host it reads:

- `/api/v1/devices/{deviceId}/overview`
- `/api/v1/devices/{deviceId}/health`
- `/api/v1/devices/{deviceId}/policy-violations`
- `/api/v1/devices/{deviceId}/anomalies`
- `/api/v1/devices/{deviceId}/tamper-events`
- `/api/v1/devices/{deviceId}/alert-deliveries`
- `/api/v1/devices/{deviceId}/reports/weekly`
- `/api/v1/tenants/{tenantId}`
- `/api/v1/plans`
- `/api/v1/roles`
- `/api/v1/retention-tiers`
- `/api/v1/audit-events`
- `/api/v1/policy-templates`

Current panels:

- host filter and host identity
- compliance score, risk score, device health, policy, anomaly, tamper, and
  delivery metrics
- study/coding/entertainment activity mix
- S3 archive health and backlog
- device health score, CPU, memory, disk, heartbeat, and recommendation
- plan readiness and tenant packaging
- notification operations for email, push, dashboard feed, and retry queue
- product packaging for weekly report, policy marketplace, roles, and audit
- risk timeline
- policy violation table
- anomaly table
- risky software watchlist for torrent, VPN/proxy, game launcher, non-standard
  browser, and downloads-installer signals
- tamper and trust table
- email, push, and dashboard alert delivery table
- policy template marketplace
- retention and archive plan catalog

API-provided text is escaped before rendering.

Phase 9 uses in-memory demo risk data for enrolled devices so the dashboard can
be smoke-tested before durable backend event storage exists. The slice does not
add new endpoint collectors and does not collect passwords, credentials,
keylogs, cookies, tokens, private messages, camera, microphone, raw URLs, page
titles, or covert screenshots.

Phase 11 persists the same backend dashboard state to
`data/local/backend/backend-state.json` by default. If the backend is started
with an API key, the dashboard shell still loads, but API requests require the
configured `X-TraceDeck-API-Key` and tenant scope headers.

A future frontend phase can move this surface to a richer application shell with
authentication, role-based views, saved filters, no-code alert rule editing,
weekly report drilldowns, and durable event search.
