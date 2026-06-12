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
and retention plan panels. Phase 16 adds an explicit buyer-facing operations
layer for anomaly notification inbox, mail delivery center, push routing, alert
route SLA details, and paid packaging cues. It remains a lightweight static
HTML/CSS/JavaScript asset embedded into the backend binary.
Phase 18 upgrades the first screen into a product-grade command center with a
priority action board, notification promise, commercial readiness score, trust
coverage, executive briefing, and notification action queue before the deeper
technical tables.
Phase 19 adds no-code alert rule builder panels for saved tenant automations
and paid rule recipes.
Phase 20 adds consent/audit trust panels and a stronger paid alert operations
band: anomaly notification proof, mail delivery proof, push notification reach,
and customer audit evidence.
Phase 21 adds managed policy rollout panels for device groups and policy
assignments.
Phase 22 adds data rights workflow panels for tenant export manifests and
delete request queues.
Phase 23 adds tenant-level customer operations panels for monetisation demos:
fleet coverage, anomaly pipeline, mail delivery proof, push reach, escalation
signals, notification delivery score, and upgrade proof pack.
Phase 24 hardens dashboard demo lifecycle scripts so stale TraceDeck listeners
are stopped before live boot testing and the served dashboard is proven to come
from the current build.

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
- `/api/v1/alert-rule-templates`
- `/api/v1/tenants/{tenantId}/alert-rules`
- `/api/v1/tenants/{tenantId}/consent-center`
- `/api/v1/tenants/{tenantId}/operations-summary`
- `/api/v1/tenants/{tenantId}/data-exports`
- `/api/v1/tenants/{tenantId}/delete-requests`
- `/api/v1/tenants/{tenantId}/device-groups`
- `/api/v1/tenants/{tenantId}/policy-assignments`

Current panels:

- host filter and host identity
- priority action board for the highest-value intervention
- notification promise for email, push, and dashboard delivery status
- commercial readiness score for Family Pro, school, and business packaging
- trust coverage across agent, archive, delivery, and audit signals
- customer operations cockpit for fleet, anomaly, mail delivery, push reach,
  and paid value
- escalation workbench for tenant-level policy, anomaly, delivery, and archive
  follow-up
- notification delivery board for tenant-level email, push, dashboard, retry,
  and failure proof
- upgrade proof pack for family, school, and business packaging
- executive briefing for top risk, study signal, alert outcome, and archive
  trust
- notification action queue for delivery retries and open risk events
- alert revenue operations for anomaly coverage, mail delivery, push reach, and
  audit proof
- push notification center for mobile anomaly routing, retry timing, provider,
  recipient, and last-send state
- compliance score, risk score, device health, policy, anomaly, tamper, and
  delivery metrics
- study/coding/entertainment activity mix
- S3 archive health and backlog
- device health score, CPU, memory, disk, heartbeat, and recommendation
- plan readiness and tenant packaging
- anomaly notification inbox with email, push, and dashboard route badges
- mail delivery center with recipient, subject, preview, PDF readiness, and
  last-send status
- notification operations for email, push, dashboard feed, and retry queue
- route-level email SLA, push routing, dashboard feed provider, attempts,
  retry, and error visibility
- product packaging for weekly report, policy marketplace, roles, and audit
- paid trigger and upgrade-path cues for Family Pro, school, and business
  packaging
- weekly report email/PDF readiness
- risk timeline
- policy violation table
- anomaly table
- risky software watchlist for torrent, VPN/proxy, game launcher, non-standard
  browser, and downloads-installer signals
- tamper and trust table
- email, push, and dashboard alert delivery table
- policy template marketplace
- retention and archive plan catalog
- no-code alert rules for saved tenant automations
- rule builder recipes for paid alert templates
- device groups for managed family, school, and business cohorts
- policy assignments for tenant and group-level policy rollout status
- consent and audit center with visible monitoring, recipients, export/delete
  readiness, pause controls, and denied sensitive collection categories
- policy audit trail for recent tenant and policy changes
- data export center for auditable tenant export manifests
- delete request queue for non-destructive data deletion workflows

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
