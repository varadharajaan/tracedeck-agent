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
Phase 27 adds a buyer-facing revenue control layer above the existing technical
tables: package fit, paid proof, upgrade motion, renewal risk, commercial
lever, anomaly assurance, email delivery, push delivery, and report mail
readiness. The panels are populated from the existing typed tenant operations,
tenant monetisation summary, alert delivery, weekly report, and host risk APIs.
Phase 28 adds Live Agent Telemetry and Telemetry Privacy Boundary panels backed
by the agent-to-backend metadata ingest bridge.
Phase 29 moves the monetisation story into the first screen with a launch deck
for customer package readiness, anomaly push assurance, mail delivery
assurance, weekly report proof, host risk command, archive retention,
notification revenue stream, and buyer action prompts.
Phase 31 adds a Buyer Assurance Wall and Offline Replay Health panel so the
paid demo can show agent sync proof, anomaly pipeline status, mail delivery,
push notification reach, weekly report mail, archive trust, highest replay
cursor, source counts, and host recommendations from the tenant sync-health
API.
Phase 32 adds a Tenant Activity Feed and Filtered Command Feed backed by the
tenant activity-feed API. The selected host drives a feed filter so anomaly,
policy, tamper, mail, push, and telemetry sync proof can be reviewed in one
timeline.

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
- `/api/v1/devices/{deviceId}/telemetry-status`
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
- `/api/v1/tenants/{tenantId}/monetization-summary`
- `/api/v1/tenants/{tenantId}/sync-health`
- `/api/v1/tenants/{tenantId}/activity-feed`
- `/api/v1/tenants/{tenantId}/data-exports`
- `/api/v1/tenants/{tenantId}/delete-requests`
- `/api/v1/tenants/{tenantId}/device-groups`
- `/api/v1/tenants/{tenantId}/policy-assignments`

Current panels:

- local dashboard access panel for API-key protected backends; the key is kept
  in browser session storage and sent as `X-TraceDeck-API-Key`
- host filter and host identity
- monetisation launch deck for customer package, readiness, notification score,
  trust score, and conversion stage
- anomaly push assurance showing route status, recipient, provider, and proof
- mail delivery assurance showing critical alert email route status and proof
- weekly report proof showing email and PDF readiness
- host risk command for the highest-risk anomaly, policy, or tamper signal
- archive retention proof for S3-backed retention and backlog state
- notification revenue stream showing email, push, and dashboard delivery proof
- buyer action prompts for immediate action, route proof, and upgrade lever
- buyer assurance wall for agent sync, anomaly pipeline, mail delivery, push
  notification, weekly report mail, and archive trust in one monetisable view
- offline replay health for tenant reporting hosts, stored backend events,
  stable local-event cursor, last ingest, privacy boundary, and per-host source
  counts
- tenant activity feed for cross-host risk, alert delivery, and telemetry sync
  summary counts
- filtered command feed for the selected host across anomaly, policy, tamper,
  mail, push, and backend-visible metadata events
- monetisation command views for saved high-risk, mail proof, push retry, and
  sync/archive buyer workflows
- notification monetisation proof for saved filters, alert reach, buyer trust,
  and next pitch readiness
- priority action board for the highest-value intervention
- notification promise for email, push, and dashboard delivery status
- commercial readiness score for Family Pro, school, and business packaging
- trust coverage across agent, archive, delivery, and audit signals
- revenue control room for package fit, paid proof, upgrade motion, renewal
  risk, and commercial lever
- buyer notification assurance for anomaly alerting, email delivery, push
  delivery, weekly report mail, last signal, and next action
- live agent telemetry proof for backend-ingested process and health metadata
- telemetry privacy boundary showing metadata-only ingest limits
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
- revenue command center for paid-plan outcome, stage, plan, seats, and buyer
  readiness
- monetisation value stack for fleet coverage, anomaly queue, mail delivery,
  push reach, weekly report, archive plan, trust center, and upgrade lever
- notification proof rail for anomaly, email, push, dashboard, and weekly
  report delivery proof
- buyer demo checklist for anomaly, route, report, archive, consent/data
  rights, and saved-view readiness

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

Phase 34 adds a session-scoped local access panel so an API-key protected
dashboard can unlock API requests without placing the key in the URL, HTML,
logs, local backend state, or repo files.

Phase 36 makes the first screen more buyer-ready for monetisation demos. The
dashboard now groups anomaly notification, mail delivery, push delivery,
dashboard inbox proof, weekly report mail/PDF readiness, archive retention,
consent/audit, and data-rights readiness into revenue and buyer-checklist
panels backed by existing typed APIs.

A future frontend phase can move this surface to a richer application shell with
authentication, role-specific navigation, no-code alert rule editing, weekly
report drilldowns, durable event search, and paid customer onboarding.
