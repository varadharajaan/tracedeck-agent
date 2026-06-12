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
risk, archive, and alert-delivery visibility. It remains a lightweight static
HTML/CSS/JavaScript asset embedded into the backend binary.

The dashboard reads the base backend endpoints:

- `/health`
- `/api/v1/devices`

For the selected host it reads:

- `/api/v1/devices/{deviceId}/overview`
- `/api/v1/devices/{deviceId}/policy-violations`
- `/api/v1/devices/{deviceId}/anomalies`
- `/api/v1/devices/{deviceId}/tamper-events`
- `/api/v1/devices/{deviceId}/alert-deliveries`

Current panels:

- host filter and host identity
- compliance score, risk score, policy, anomaly, tamper, and delivery metrics
- study/coding/entertainment activity mix
- S3 archive health and backlog
- risk timeline
- policy violation table
- anomaly table
- tamper and trust table
- email, push, and dashboard alert delivery table

API-provided text is escaped before rendering.

Phase 9 uses in-memory demo risk data for enrolled devices so the dashboard can
be smoke-tested before durable backend event storage exists. The slice does not
add new endpoint collectors and does not collect passwords, credentials,
keylogs, cookies, tokens, private messages, camera, microphone, raw URLs, page
titles, or covert screenshots.

A future frontend phase can move this surface to a richer application shell with
authentication, role-based views, saved filters, no-code alert rule editing,
weekly report drilldowns, and durable event search.
