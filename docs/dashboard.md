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
Phase 47 adds a Premium Notification Command Center backed by the typed
notification-command-center API. It packages anomaly/policy/tamper urgency,
email delivery proof, push reach, dashboard route proof, remediation SLA state,
paid-tier labels, and owner actions into a monetisable notification command
surface.
Phase 51 adds a Notification Evidence Timeline backed by the typed
delivery-timeline API. The panel shows selected-host delivery score, email
proof, push retry evidence, dashboard inbox proof, route proof gaps, next retry
timing, paid tier, metadata-only privacy boundary, and an audit trail of
provider-safe delivery rows.
Phase 52 adds a Role Experience Center backed by the typed role-experiences
API. The panel packages parent, student, school admin, and business manager
views with readiness scores, visible panels, notification promise,
archive/report promise, consent controls, paid tier, next action, and a paid
onboarding checklist.
Phase 53 adds an Executive Notification Console backed by the typed
executive-console API. It is the first monetisation-grade surface on the page
and shows sellable readiness, anomaly/open alert pressure, mail delivery proof,
push reach, weekly report readiness, archive posture, role packaging, value
tiles, provider-safe delivery proof, alert stream, and owner actions.
Phase 54 adds a Notification Revenue Cockpit backed by the typed
notification-revenue-cockpit API. It adds a buyer-ready notification
monetisation layer for anomaly SLA, mail proof, push proof, dashboard delivery,
weekly report readiness, escalation policy, scenario templates, channel proof
matrix, and upgrade action levers.
Phase 55 adds a Provider Simulation Lab backed by the typed
provider-simulation-lab API. It shows metadata-only email, push, and dashboard
dry-run proof, route SLA state, simulation scenarios, provider action queue,
privacy proof, and command navigation so paid demos can prove notification
readiness without provider secrets or alert payloads.
Phase 57 adds a Customer Control Room backed by the typed customer-control-room
API. It moves the monetisation story to the first dashboard surface: anomaly
command, mail delivery, push notification evidence, provider simulation proof,
report/archive readiness, package billing, customer health, and owner actions
are visible before the specialised drilldowns.
Phase 58 adds a Customer Success Packet backed by the typed
customer-success-packet API. It gives buyers and admins one review surface for
anomaly proof, mail delivery, push notification evidence, weekly report and S3
archive readiness, package fit, provider rehearsal, privacy assurances,
objection answers, and owner actions.
Phase 69/72 hardening adds visible source provenance on local and cloud
dashboard rows. Demo-seeded risk and delivery rows show `demo_seed`, live
browser telemetry shows `live_ingested`, and Lambda S3 rows show `s3_sample`.
This prevents demo proof from being mistaken for a real SMTP/web-push send.
Phase 76/77 UI polish gives the dashboard and Browser Activity viewer a cleaner
product shell: neutral light/dark palettes, stronger brand header, cleaner
toolbar actions, modern page tabs, calmer status chips, readable cards, and a
screenshot-free theme contract for both desktop and mobile.

Phase 78 tightens the paid-product shell again. The main page is labelled
`TraceDeck Console`, the browser drilldown is labelled `Browser Viewer`, and
the old command shortcut language is presented as a `Workspace Navigator` with
full product labels instead of abbreviations. A new visual-quality contract
checks rendered light and dark pages for stale brace labels, pseudo-letter
markers, tiny visible controls, dark-mode color posture, visible server lights,
and horizontal overflow without capturing screenshots.

Phase 81 makes those Workspace Navigator labels explicit in the markup. Each
navigation tile has a stable `command-label` row, such as `Deployment
Readiness`, `Customer Control Room`, `Provider Setup`, `Paid Operations`, and
`Delivery Assurance`, plus a separate `command-meta` row for live counts or
scores. The dashboard visual-quality contract now verifies the full product
labels and rejects the older terse shortcut labels.

Phase 82 adds the modern admin polish layer across the local dashboard,
Browser Activity drilldown, and Lambda Cloud Admin frontend. The visible `TD`
text mark is replaced by a symbolic mark, light/dark palettes share one visual
system, command tiles and status chips have larger readable treatment, and the
contracts reject stale `Browser{}`, `Center{}`, bracket shortcut, `TD`, and
internal abbreviation copy.

Phase 84 adds a final customer-grade visual layer for the local dashboard and
Browser Viewer. The dashboard now reads as a calm endpoint observability command
center: a stronger app header, polished segmented page navigation, larger
status chips, aligned evidence cards, readable tables, and a consistent dark
theme. The Workspace Navigator metadata is stacked under full product labels
instead of appearing as tiny side labels. The Phase 84 smoke, Newman, visual,
theme, and layout contracts reject debug-looking brace labels, pseudo-letter
shortcuts, terse internal abbreviations, tiny visible controls, and horizontal
overflow.

To check the currently running local dashboard:

```powershell
python ./devctl.py test live
```

To rerun the full provenance hardening gate:

```powershell
python ./devctl.py test phase73
```

To write a one-file runtime assurance report for the local dashboard, Browser
Activity Viewer, Lambda Function URL, S3 summary, and cache state:

```powershell
python ./devctl.py doctor
python ./devctl.py doctor --skip-cloud
```

The reports are saved under:

```text
data/local/output/runtime-doctor.json
data/local/output/runtime-doctor.txt
```

To verify the dashboard and Browser Activity light/dark theme contract without
capturing screenshots:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 -BaseUrl http://127.0.0.1:18080
python ./devctl.py test theme
```

To verify the product visual-quality contract:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -BaseUrl http://127.0.0.1:18080
python ./devctl.py test visual
```

To rerun the full Phase 84 UI revamp gate:

```powershell
python ./devctl.py test phase84
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase84.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase84.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase84.ps1
```

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
- `/api/v1/tenants/{tenantId}/delivery-timeline?device_id={deviceId}&limit=8`
- `/api/v1/tenants/{tenantId}/role-experiences`
- `/api/v1/tenants/{tenantId}/customer-control-room`
- `/api/v1/tenants/{tenantId}/customer-success-packet`
- `/api/v1/tenants/{tenantId}/executive-console`
- `/api/v1/tenants/{tenantId}/notification-revenue-cockpit`
- `/api/v1/tenants/{tenantId}/provider-simulation-lab`
- `/api/v1/tenants/{tenantId}/package-billing-readiness`
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
- `/api/v1/tenants/{tenantId}/notification-command-center`
- `/api/v1/tenants/{tenantId}/sync-health`
- `/api/v1/tenants/{tenantId}/activity-feed`
- `/api/v1/tenants/{tenantId}/data-exports`
- `/api/v1/tenants/{tenantId}/delete-requests`
- `/api/v1/tenants/{tenantId}/device-groups`
- `/api/v1/tenants/{tenantId}/policy-assignments`

Current panels:

- local dashboard access panel for API-key protected backends; the key is kept
  in browser session storage and sent as `X-TraceDeck-API-Key`
- customer control room for customer-ready score, anomaly command, mail
  delivery, push reach, provider proof, package score, archive/report posture,
  paid value tiles, alert wall, delivery evidence, and owner monetisation
  actions
- executive notification console for sellable readiness, anomaly urgency, mail
  proof, push reach, weekly report readiness, archive posture, role packaging,
  paid value tiles, alert stream, delivery proof, and owner actions
- notification revenue cockpit for anomaly SLA, push notification proof, mail
  delivery proof, dashboard delivery, weekly report readiness, escalation
  policy, anomaly delivery scenarios, channel proof matrix, and upgrade action
  levers
- provider simulation lab for metadata-only email, push, and dashboard dry-run
  proof, route SLA state, simulation scenarios, action queue, and privacy proof
- host filter and host identity
- monetisation launch deck for customer package, readiness, notification score,
  trust score, and conversion stage
- anomaly push assurance showing route status, recipient, provider, and proof
- mail delivery assurance showing critical alert email route status and proof
- weekly report proof showing email and PDF readiness
- host risk command for the highest-risk anomaly, policy, or tamper signal
- archive retention proof for S3-backed retention and backlog state
- notification revenue stream showing email, push, and dashboard delivery proof
- delivery assurance center showing provider-confirmed, dry-run, dashboard-
  visible, demo-only, retrying, failed, disabled, and pending-provider route
  truth so demo rows never look like real email or screen notification proof
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
- commercial control room for host coverage, anomaly command, email proof,
  push proof, weekly report mail, upgrade trigger, delivery evidence, and
  customer success actions
- monetisation value stack for fleet coverage, anomaly queue, mail delivery,
  push reach, weekly report, archive plan, trust center, and upgrade lever
- notification proof rail for anomaly, email, push, dashboard, and weekly
  report delivery proof
- buyer demo checklist for anomaly, route, report, archive, consent/data
  rights, and saved-view readiness
- command navigation for control, executive, paid ops, revenue, Notify Pro,
  notifications, reports, archive, trust, and hosts with KPI summaries backed
  by existing typed APIs
- buyer operations brief for monetisation demos with anomaly alerting, mail
  delivery proof, push notification dispatch, weekly report delivery, archive
  retention, trust/audit, delivery command, packaging snapshot, and next
  commercial action surfaced before the deeper panels
- provider-safe delivery drilldown for email, push, and dashboard dry-run
  rehearsal, route score, channel readiness, route evidence, and next actions
  without provider secrets or alert bodies
- delivery remediation center for SLA-aware route recovery, owner assignment,
  dry-run retry planning, audit proof, anomaly push recovery, and mail delivery
  assurance without live sends or provider payload storage
- premium notification command center for alert funnel, email proof, push
  reach, route assurance, paid-tier packaging, and customer action SLAs

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

Phase 37 adds a dashboard contract guard. The local test parses the embedded
dashboard, rejects duplicate DOM IDs, and verifies that JavaScript-rendered
targets referenced by `getElementById`, text/metric/bar helpers, and badge
replacement calls exist in the HTML. This catches missing panel IDs before a
live demo.

Phase 38 adds the Commercial Control Room ahead of the earlier launch deck so a
paid demo opens on buyer-ready proof instead of only host panels. It rolls up
tenant operations, monetisation summary, notification routes, weekly report
mail/PDF readiness, host risk, anomaly urgency, email delivery, push delivery,
and customer success actions into one first-screen surface. The phase adds
presentation and verification only; it does not add sensitive collection.

Phase 42 adds a sticky command navigation strip so paid demos can jump directly
to paid ops, revenue, notification proof, reports, archive, trust, and host
details. The strip summarizes alert, route, report, archive, readiness, and
trust state from the same typed APIs used by the rest of the dashboard.

Phase 43 adds a Buyer Operations Brief immediately after command navigation so
the paid demo opens with anomaly alerting, email proof, push notification
dispatch, weekly report readiness, S3 archive retention, trust/audit, delivery
routes, package fit, and next customer action in one surface. The phase also
adds a Playwright layout contract that checks desktop, tablet, and mobile
metrics only. It does not capture screenshots, video, credentials, or page
content.

Phase 44 adds Provider-Safe Delivery Drilldown panels. They render the tenant
delivery-drilldown API: route score, email/push/dashboard readiness, privacy
boundary, per-route rehearsal evidence, SLA promise, and next actions. Dry-run
rehearsal is metadata-only and does not send live provider messages, store
provider secrets, or persist alert body content.

Phase 45 adds the Monetisation Command Center immediately after command
navigation. It is the premium first-screen surface for a paid demo: anomaly
inbox, push notification route, mail delivery proof, weekly report mail/PDF,
fleet coverage, S3 archive retention, trust/audit, revenue package, delivery
proof, and owner action queue are visible before deeper host panels. The panel
uses existing tenant operations, monetisation summary, alert inbox, delivery
drilldown, sync health, consent, weekly report, and per-host delivery APIs.

Phase 46 adds a Delivery Remediation Center. It shows route recovery score,
open push/mail/dashboard problems, planned dry-run actions, owner
acknowledgement, SLA watch, next retry/check windows, and a remediation action
ledger. The Monetisation Command Center also includes remediation state in the
delivery proof list so paid demos answer not only "what failed?" but "who owns
the recovery and what proof exists?"

Phase 47 adds a Premium Notification Command Center and Notify Pro jump target.
It renders the tenant notification-command-center API: alert funnel,
high-priority alert count, email delivery proof, push reach, dashboard route
proof, remediation SLA state, paid-tier labels, and customer owner actions.
The view is an aggregate presentation layer and does not add sensitive
collectors or live provider sends.

Phase 48 adds a first-screen Growth Cockpit above command navigation. It turns
the existing typed API data into a monetisation-grade product view:
revenue readiness, anomaly notification ops, mail delivery, push delivery,
weekly report proof, archive retention, trust/consent, and owner actions are
visible before host-level details. The panel is backed by the notification
command center, monetisation summary, operations summary, alert inbox, weekly
report, delivery drilldown, remediation, archive, and consent metadata.

Phase 49 adds a Notification Preference Center. It shows preference score,
channel policy, quiet hours, escalation, digest cadence, immediate/digest/silent
rules, study-safe suppression, retention evidence, and owner actions. The
dashboard reads from the typed tenant notification-preferences API and keeps the
same metadata-only privacy boundary as the route registry and command center.

Phase 50 adds a Business Dashboard above the growth cockpit. It is the
monetisation-first surface for customer health, anomaly notification inbox,
push and mail proof, route proof, archive/report value, paid package cards, and
customer owner actions. The dashboard reads from the typed tenant
business-dashboard API so the first screen shows whether anomalies are present,
whether push/mail/dashboard delivery worked, what package value is proven, and
what an owner should do next.

Phase 56 adds Package Billing Readiness near the notification revenue and
provider simulation surfaces. It renders the typed
`package-billing-readiness` API as a package score, billing setup status,
feature gate proof, seat usage, plan fit matrix, billing milestones, and
upgrade action queue. The panel makes paid packaging visible for demos without
adding payment collection or exposing private content.

Phase 57 adds the Customer Control Room above the executive console. It renders
the typed `customer-control-room` API as the opening product surface with
customer-ready score, anomaly command wall, mail and push delivery proof,
provider simulation state, package billing score, report/archive readiness,
customer value tiles, and owner monetisation actions. The surface aggregates
existing metadata-only proof and does not add collectors or provider sends.

Phase 58 adds the Customer Success Packet after the Customer Control Room. It
renders the typed `customer-success-packet` API as a buyer-facing packet with a
success proof stack, buyer objection answers, success packet actions, and a
delivery/trust promise. The dashboard shows anomaly proof, mail delivery, push
notification reach, report/archive posture, package fit, provider readiness,
role readiness, and privacy proof without exposing sensitive payloads.

Phase 59 adds the Push Activation Center after the Customer Success Packet. It
renders the typed `push-activation-center` API as a paid notification
reliability surface: push delivered/retrying state, mail fallback, dashboard
fallback, push route proof, anomaly push/mail scenarios, owner actions,
preference and escalation readiness, provider-safe simulation readiness, and a
privacy guard. The command navigation includes a Push jump target so a demo can
answer "will I actually get notified?" without exposing raw endpoints, alert
bodies, screenshots, page titles, URLs, passwords, or provider secrets.

Phase 60 adds the Portfolio Center after Push Activation. It renders the typed
`portfolio-center` API as the multi-host owner/admin view: portfolio score,
host coverage, open/high-priority alerts, notification score, mail and push
proof, dashboard fallback, archive/sync posture, trust score, alert
notification rows, delivery proof cards, host portfolio rows, portfolio
segments, owner actions, and privacy guard. The command
navigation includes a Portfolio jump target so parents, school admins, and
business managers can compare hosts without opening every host panel.

Phase 61 adds the Account Portfolio Index before the command navigation. It
renders the typed `account-portfolio-index` API as a multi-tenant account/admin
opening view: account score, tenant count, host coverage, alert pressure,
notification score, mail and push delivery proof, dashboard fallback, archive
posture, tenant rows, proof cards, owner actions, and privacy guard. The
command navigation includes an Account jump target so admins can move between
account-level and tenant-level proof without losing context.

Phase 62 adds a Monetisation Overview above the existing drilldowns. It makes
the first screen read like a paid product cockpit: account and host proof,
anomaly notification proof, mail delivery, push reach, weekly report readiness,
archive posture, package/revenue fit, owner actions, and trust guardrails are
visible immediately. The section reuses existing typed APIs rather than adding
new collectors.

Phase 63 adds a Tenant Onboarding Center near the dashboard front door. It
renders the typed `onboarding-center` API as a paid activation checklist with
setup readiness, host reporting, autostart proof, anomaly notification policy,
mail/push/dashboard delivery proof, archive posture, role handoff, package
readiness, privacy guard, and owner actions. The command navigation includes an
Onboard target so a buyer/admin can answer whether deployment, alerts, archive,
reports, and roles are ready before moving into host-level drilldowns.

Phase 64 adds a Customer Settings Center next to onboarding. It renders the
typed `customer-settings-center` API as configurable activation proof: plan and
retention recommendations, notification policy state, mail route proof, push
route proof, dashboard fallback, archive and autostart readiness, role view
settings, data-rights settings, and owner actions. The command navigation adds
a Settings target so a buyer/admin can move from readiness proof to settings
review without exposing payment data, provider secrets, alert bodies, raw URLs,
or push endpoints.

Phase 65 adds a Revenue Operations Center after Customer Settings. It renders
the typed `revenue-operations-center` API as the paid-product overview:
revenue readiness, anomaly queue, mail delivery, push notification, dashboard
fallback, weekly report, archive retention, setup readiness, customer settings,
package fit, provider simulation, commercial levers, and owner actions. The
command navigation adds a Rev Ops target so demos and admin reviews can answer
"is the product working and sellable today?" before drilling into individual
host or route panels.

Phase 66 adds a Deployment Readiness Center after Revenue Operations. It
renders the typed `deployment-readiness-center` API as the rollout proof view:
Windows Task Scheduler, macOS launchd, Linux systemd, service manifests, live
boot, reboot persistence, background startup, offline replay, archive backlog,
and deployment owner actions. The command navigation adds a Deploy target so
admins can answer "will the agent come back after restart?" without leaving the
monetisation dashboard.

Phase 67 adds a Premium Operations Hub above the existing monetisation
drilldowns. It renders the typed `premium-operations-hub` API as the
buyer/admin first screen: premium readiness, anomaly inbox, mail delivery,
push notification route proof, dashboard fallback, weekly report readiness,
archive retention, deployment readiness, package value, commercial levers, and
owner actions. The command navigation adds a Premium target so the user can
start from the sellable operating view and drill into Rev Ops, Deploy, Push,
Portfolio, or host details only when needed.

Phase 68 adds a dedicated Browser Activity Viewer at `/browser-activity` and a
Browser button in the main dashboard toolbar. The page renders the typed
`browser-activity` API with tenant, host, browser, category, study-safe, and
search filters; Chrome/Edge/Brave KPIs; host and browser breakdowns; non-study
YouTube visibility; notification proof; and a domain activity table. The page
uses the same local session header pattern as the main dashboard and stays
metadata-only: no raw URLs, page titles, cookies, tokens, passwords,
screenshots, provider secrets, alert bodies, private content, or endpoint
payloads.

Phase 69 adds dashboard navigation and operator polish. The main dashboard now
has page tabs for Overview, Operations, Notifications, Deployment, Portfolio,
Hosts, and Admin so admins do not have to scroll through every monetisation
panel in one pass. `?layout=all` remains available for the screenshot-free
layout contract. The main dashboard and `/browser-activity` both include a
dark theme toggle plus an explicit server status light: green for connected,
red for disconnected, and amber for locked local access.

Phase 69 also adds the optional cloud admin frontend under `sam-app/`. It is
served by a Lambda Function URL, reads TraceDeck archive metadata from S3,
shows cache hit/miss percentages, and can switch the browser-side source to a
configured local backend such as `http://127.0.0.1:18080` when an admin is on
the machine.

Phase 76 revamps the embedded dashboard and Browser Activity Viewer visual
system. The dashboard keeps the existing typed data and multi-page shell but
uses a cleaner product palette, stronger hierarchy for focus panels, quieter
KPI cards, wrapped chips, contained tables, polished light/dark modes, and no
visible pseudo-letter toolbar markers. The screenshot-free layout contract now
guards this shell against horizontal overflow on desktop, tablet, and mobile.

Phase 78 adds the Notification Provider Setup Center after the Provider
Simulation Lab. It shows configured versus provider-confirmed email, push, and
dashboard fallback setup; demo-only and retrying truth labels; buyer readiness;
setup checklist; owner actions; and metadata-only provider boundaries. It also
renames the shell to TraceDeck Console, uses Browser Viewer as the browser
drilldown label, removes internal abbreviations from the navigator, and adds
the screenshot-free visual-quality gate.

Phase 82 keeps the same functionality but raises the buyer-facing UI baseline:
symbolic brand mark, unified local/cloud light and dark themes, larger status
chips, cleaner panels and KPI cards, and product-grade command navigation.

Phase 83 adds backend-visible agent heartbeat proof. The agent emits
`agent.health.heartbeat` from `collector.agent.heartbeat` once per collection
cycle with readiness metadata such as version, collection mode, archive due
state, backend sync state, alert state, profile, and operating system. The
dashboard and sync-health views can use that event as host-level agent health
and replay proof without showing screenshots, passwords, raw URLs, page titles,
cookies, tokens, private content, endpoint payloads, provider secrets, alert
bodies, keylogs, or hidden collection bypass data.

Future frontend phases can move this surface to a richer application shell with
no-code alert rule editing, weekly report drilldowns, durable event search, and
paid customer onboarding workflows.
