# Testing

Local verification replaces GitHub Actions for this project.

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/setup/install-go-tools.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase0.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase1.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase1b.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase2.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase2b.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase3.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase4.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase7.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase8.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase10.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase15.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase17.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase25.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase26.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase27.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase28.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase29.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase30.ps1
```

Verification logs are written under `logs/local/verify/`.
Root-level generated artifacts are rejected by `scripts/verify/check-root-clean.ps1`.
Cross-platform build outputs are written under `data/local/build/`.
Browser fixture smoke data is generated under `data/local/smoke-phase3/`, and
the smoke archive is checked to ensure raw URLs and page titles are not stored.
Earlier phase smokes pass `--disable-browser-history` so they do not collect
real local browser history while validating process/archive/alert behavior.

Phase 4 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-alert-engine.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase4.ps1
```

The Phase 4 smoke creates a local policy fixture under `data/local/`, boots the
agent once with a generated browser history fixture, stages a dry-run archive
and alert notification, and verifies both `non_study_youtube` and
`blocked_domain_opened` without leaking raw URL or page-title data.

Phase 5 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase5.ps1
```

The Phase 5 smoke builds a local backend executable under `data/local/`, boots
it on localhost, exercises health/version/device/template/archive/dashboard
routes, and stops the process. Newman runs the committed Postman collection and
writes its JSON report under `data/local/newman/phase5/`.

Phase 6 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase6.ps1
```

The Phase 6 smoke builds and boots the backend on localhost, exercises plan,
role, retention, tenant, audit, and dashboard routes, then stops the process.
Newman runs the committed Phase 6 Postman collection and writes its JSON report
under `data/local/newman/phase6/`.

Phase 7 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-platform-support.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1
```

The Phase 7 verifier checks platform support tests, renders macOS launchd and
Linux systemd manifests under `data/local/service-manifests/phase7/`, runs
cross-platform builds, and verifies root artifact hygiene.

Phase 8 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-windows-task-template.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1 -AllowMissing
```

The Phase 8 verifier renders and parses the Windows Task Scheduler XML,
confirms the logon trigger and continuous agent arguments, runs cross-platform
builds, and verifies root artifact hygiene. It does not register the task; the
registration script intentionally requests UAC elevation as an explicit local
operator action.

Phase 9 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase9.ps1
```

The Phase 9 smoke builds and boots the backend on localhost, creates a tenant,
enrolls a host, verifies host overview, policy violation, anomaly, tamper, alert
delivery, archive, and dashboard HTML behavior, then stops the process. Newman
runs `postman/tracedeck-backend-phase9.postman_collection.json` against a live
backend and writes its JSON report under `data/local/newman/phase9/`.

Phase 10 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1
```

The Phase 10 verifier starts the backend with seeded demo host data on
localhost, verifies the host overview and dashboard HTML, checks root hygiene,
and stops the demo backend.

Phase 11 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase11.ps1
```

The Phase 11 smoke builds the backend, boots it with an isolated JSON state
file and API key, verifies missing-key rejection, creates a tenant, enrolls a
host, confirms risk data exists, restarts the backend against the same state
file, and verifies the host and alert delivery rows survived restart. Newman
runs `postman/tracedeck-backend-phase11.postman_collection.json` against a live
API-key-protected backend.

Phase 14 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase14.ps1
```

The Phase 14 smoke builds and boots the backend on localhost, enrolls a host,
verifies generated weekly report JSON, verifies the PDF endpoint returns
`application/pdf`, confirms the dashboard weekly report panel still renders,
and stops the backend. Newman runs
`postman/tracedeck-backend-phase14.postman_collection.json` against a live
API-key-protected backend.

Phase 15 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-service-manager.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Action status -DryRun
```

The Phase 15 verifier dry-runs `install`, `start`, `stop`, `status`, and
`uninstall` for Windows, macOS, and Linux, confirms the generated action plans
include Task Scheduler, systemd, and launchd commands, renders service
manifests, verifies the Windows Task Scheduler template, runs cross-platform
builds, and checks root artifact hygiene. It does not register, start, stop, or
remove real host services during verification.

Phase 16 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase16.ps1
```

The Phase 16 smoke builds and boots the backend on localhost, creates a tenant,
enrolls a host, verifies policy/anomaly/tamper signals, confirms email, push,
and dashboard alert delivery routes, checks weekly report email/PDF readiness,
and asserts the embedded dashboard contains the anomaly notification inbox, mail
delivery center, push routing, route SLA, paid trigger, marketplace, and
retention panels. Newman runs
`postman/tracedeck-backend-phase16.postman_collection.json` against a live
API-key-protected backend.

Phase 17 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-email-notifier.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase17.ps1
```

The Phase 17 smoke builds a local fake SMTP helper under `data/local/`, starts
it hidden on localhost, runs the agent once with `--alert-dry-run=false`,
captures the delivered `.eml`, verifies alert content, and confirms forbidden
URL and SMTP secret markers are not leaked. The verifier also regenerates the
policy schema, runs focused mail tests, runs cross-platform builds, and checks
root artifact hygiene.

Phase 18 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase18.ps1
```

The Phase 18 smoke builds and boots the backend on localhost, creates a Family
Pro tenant, enrolls a host, verifies seeded policy/anomaly/delivery signals,
and asserts the embedded dashboard includes the priority action board,
notification promise, commercial readiness, trust coverage, executive briefing,
and notification action queue. Newman runs
`postman/tracedeck-backend-phase18.postman_collection.json` against a live
API-key-protected backend and checks the same dashboard cockpit plus API
signals.

Phase 19 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase19.ps1
```

The Phase 19 smoke builds and boots the backend on localhost, creates a Family
Pro tenant, verifies alert rule templates, verifies seeded tenant rules, creates
a custom alert rule, and asserts the dashboard contains no-code alert rules and
rule builder recipes. Newman runs
`postman/tracedeck-backend-phase19.postman_collection.json` against a live
API-key-protected backend.

Phase 20 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase20.ps1
```

The Phase 20 smoke builds and boots the backend on localhost, creates a Family
Pro tenant, verifies the consent center readiness flags, checks denied password
and screenshot collection disclosures, confirms audit/recipient visibility, and
asserts the dashboard contains alert revenue operations, push notification
center, consent, and audit panels. Newman runs
`postman/tracedeck-backend-phase20.postman_collection.json` against a live
API-key-protected backend, enrolls a host, checks anomaly plus email, push, and
dashboard delivery routes, checks the consent center, and checks the dashboard
shell.

Phase 21 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase21.ps1
```

The Phase 21 smoke builds and boots the backend on localhost, creates a Family
Pro tenant, enrolls a host, verifies seeded device groups and policy
assignments, creates an Exam Mode device group, creates an active assignment,
checks audit events, and asserts the dashboard contains Device Groups and Policy
Assignments panels. Newman runs
`postman/tracedeck-backend-phase21.postman_collection.json` against a live
API-key-protected backend and checks the same managed rollout contract.

Phase 22 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase22.ps1
```

The Phase 22 smoke builds and boots the backend on localhost, creates a Family
Pro tenant, creates a ready tenant export manifest, queues a non-destructive
delete request, verifies list APIs and audit events, and asserts the dashboard
contains Data Export Center and Delete Request Queue panels. Newman runs
`postman/tracedeck-backend-phase22.postman_collection.json` against a live
API-key-protected backend.

Phase 23 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase23.ps1
```

The Phase 23 smoke builds and boots the backend on localhost, creates a Family
Pro tenant, enrolls a host, verifies the tenant operations summary, checks mail
delivery proof, priority signals, and upgrade proof signals, and asserts the
dashboard contains Customer Operations Cockpit, Escalation Workbench,
Notification Delivery Board, and Upgrade Proof Pack panels. Newman runs
`postman/tracedeck-backend-phase23.postman_collection.json` against a live
API-key-protected backend.

Phase 24 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase24.ps1
```

The Phase 24 smoke starts the dashboard demo twice on the same localhost port,
verifies the stale first listener is stopped, verifies the second process owns
the port, checks current dashboard customer operations panels, and verifies the
seeded tenant operations API. Newman starts the dashboard through the same
launcher and verifies health, dashboard HTML, and operations summary.

Phase 29 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase29"
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase29.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase29.ps1
```

The Phase 29 verifier syntax-checks the embedded dashboard JavaScript, then the
smoke live-boots the seeded dashboard, asserts the embedded UI contains the
monetisation launch deck, anomaly push assurance, mail delivery assurance,
weekly report proof, and notification revenue stream, then verifies the backing
overview, delivery, weekly report, operations, and monetisation summary APIs.
Newman runs
`postman/tracedeck-backend-phase29.postman_collection.json` against a live
dashboard demo and checks the same buyer-ready notification and paid packaging
contract.

Phase 30 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase30.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase30.ps1
```

The Phase 30 smoke runs the real agent once with backend sync pointed at an
offline localhost port, verifies the run succeeds with an unsynced telemetry
backlog, starts the backend, reruns the agent with the same SQLite data
directory, and verifies the backlog replay reaches
`/api/v1/devices/{deviceId}/telemetry-status`. Newman runs
`postman/tracedeck-backend-phase30.postman_collection.json` against a live
dashboard demo and verifies duplicate stable telemetry event IDs are
acknowledged without duplicate backend storage.

Phase 31 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase31.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase31.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase31.ps1
```

The Phase 31 smoke live-boots the seeded dashboard, ingests stable
process/browser/health metadata events, verifies
`/api/v1/tenants/{tenantId}/sync-health`, checks email and push monetisation
proof, and asserts the dashboard contains the Buyer Assurance Wall and Offline
Replay Health panels. Newman runs
`postman/tracedeck-backend-phase31.postman_collection.json` against the live
dashboard demo and checks the same API/UI contract.

Phase 32 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase32.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase32.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase32.ps1
```

The Phase 32 smoke live-boots the seeded dashboard, verifies the Tenant
Activity Feed and Filtered Command Feed markers, checks the default tenant feed,
checks selected-host email delivery filtering, posts one stable metadata event,
and verifies telemetry sync proof appears in the same feed. Newman runs
`postman/tracedeck-backend-phase32.postman_collection.json` against the live
dashboard demo and checks the same API/UI contract.

Phase 33 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase33.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase33.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase33.ps1
```

The Phase 33 smoke live-boots the seeded dashboard, verifies Monetisation
Command Views and Notification Monetisation Proof markers, checks seeded
activity views, creates a custom business dashboard view, and confirms audit
proof. Newman runs
`postman/tracedeck-backend-phase33.postman_collection.json` against the live
dashboard demo and checks the same API/UI contract.

Phase 34 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase34.ps1
```

The Phase 34 smoke live-boots an API-key protected backend, verifies the
dashboard local access panel and session-storage wiring, confirms the served
HTML does not embed the configured API key, checks protected APIs reject
missing keys, and confirms authenticated tenant/device calls with dashboard
headers. Newman runs
`postman/tracedeck-backend-phase34.postman_collection.json` against the same
protected live backend.

Phase 35 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-policy-schema.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase35.ps1
```

The Phase 35 schema check generates the versioned `v1alpha1` policy schema
under `data/local/schema-check/phase35/`, SHA-256 compares normalized schema
text with the checked-in schema file, and fails on drift. The verifier also
runs gofmt, agent tests, backend API tests, dashboard JavaScript syntax checks,
the Phase 34 authenticated dashboard Newman guard, cross-platform builds, and
root artifact hygiene.

Phase 36 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase36.ps1
```

The Phase 36 smoke live-boots the seeded dashboard, verifies the Revenue
Command Center, Monetisation Value Stack, Notification Proof Rail, and Buyer
Demo Checklist markers, then checks monetisation summary, operations summary,
notification routes, and consent/data-rights APIs. Newman runs
`postman/tracedeck-backend-phase36.postman_collection.json` against the live
dashboard demo and checks the same revenue-dashboard contract.

Phase 37 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase37.ps1
```

The Phase 37 dashboard contract guard parses the embedded dashboard HTML,
verifies there are no duplicate DOM IDs, extracts JavaScript-rendered ID
references from `getElementById`, `setText`, metric/bar helpers, and badge
replacement calls, and fails if any referenced DOM target is missing. The
smoke and Newman scripts keep live dashboard and monetisation API coverage in
the same verification path.

Phase 38 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase38.ps1
```

The Phase 38 smoke live-boots the seeded dashboard and verifies the Commercial
Control Room, Alert Delivery Evidence, Customer Success Queue, tenant
operations summary, monetisation summary, and notification route registry.
Newman runs `postman/tracedeck-backend-phase38.postman_collection.json` against
the same live demo so the monetisation-grade first screen remains covered by a
repeatable API/dashboard contract.

Phase 39 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-autostart-assurance.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase39.ps1
```

The Phase 39 autostart assurance test renders and parses the Windows Task
Scheduler XML, verifies hidden startup, logon delay, continuous agent mode,
restart-on-failure, start-when-available, battery settings, typed missing-task
status JSON, and Windows service-manager dry-run install/status plans. The
smoke then live-boots the seeded dashboard to keep service trust markers covered
after lifecycle-script changes. Newman runs
`postman/tracedeck-backend-phase39.postman_collection.json` against that live
demo.

Phase 40 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase40.ps1
```

The Phase 40 smoke live-boots the seeded dashboard and verifies the paid ops
console markers, notification route proof, monetisation action queue, tenant
monetisation scores, email/push/dashboard alert delivery channels, event-linked
delivery records, and weekly report email/PDF readiness. Newman runs
`postman/tracedeck-backend-phase40.postman_collection.json` against the same
live demo so the monetisation-grade UI contract includes anomaly, push, mail,
report, archive, and trust proof.

Phase 41 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase41.ps1
```

The Phase 41 smoke live-boots the seeded dashboard and verifies the typed tenant
alert inbox endpoint, backend alert inbox dashboard markers, monetisation UI
markers for paid ops, commercial control, revenue command, notification proof,
buyer checklist, mail delivery, push notification, archive retention, and tamper
trust, event-linked delivery proof for email/push/dashboard, activity feed
delivery continuity, and the metadata-only privacy boundary. Newman runs
`postman/tracedeck-backend-phase41.postman_collection.json` against the same
live demo.

Phase 42 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase42.ps1
```

The Phase 42 smoke live-boots the seeded dashboard and verifies the command
navigation strip, stable jump targets for paid ops, revenue, notifications,
reports, archive, trust, and hosts, and the typed API data that backs those
navigation KPIs: alert inbox, notification routes, monetisation readiness, and
weekly report email/PDF readiness. Newman runs
`postman/tracedeck-backend-phase42.postman_collection.json` against the same
live demo.

Phase 43 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/setup/install-playwright-python.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase43.ps1
```

The Phase 43 smoke live-boots the seeded dashboard, verifies the Buyer
Operations Brief markers for anomaly alerting, mail delivery proof, push
notification dispatch, weekly report delivery, archive retention, trust/audit,
delivery command, package snapshot, and action/SLA state, then runs the
screenshot-free Playwright layout contract across desktop, tablet, and mobile
viewports. The layout report is metrics only and is written under
`data/local/dashboard-layout/`; it does not capture screenshots, video,
credentials, or page content. Newman runs
`postman/tracedeck-backend-phase43.postman_collection.json` against the same
live demo.

Phase 44 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase44.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase44.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase44.ps1
```

The Phase 44 smoke live-boots the seeded dashboard, verifies Provider-Safe
Delivery Drilldown and Delivery Rehearsal Actions markers, checks the tenant
delivery-drilldown API, runs a push dry-run rehearsal, verifies the rehearsal
audit event, and reruns the screenshot-free dashboard layout contract. Newman
runs `postman/tracedeck-backend-phase44.postman_collection.json` against the
same live demo and covers current drilldown proof, dry-run rehearsal,
invalid live-send rejection, and audit evidence.

Phase 45 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase45.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase45.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase45.ps1
```

The Phase 45 smoke live-boots the seeded dashboard, verifies Monetisation
Command Center, Anomaly And Notification Inbox, Delivery And Mail Proof, and
Owner Action Queue markers, checks tenant operations, monetisation summary,
alert inbox, and delivery drilldown APIs, and reruns the screenshot-free layout
contract. Newman runs
`postman/tracedeck-backend-phase45.postman_collection.json` against the same
live demo and covers the command-center dashboard markers plus the typed APIs
that feed anomaly, notification, mail, archive, and revenue proof.

Phase 46 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase46.ps1
```

The Phase 46 smoke live-boots the seeded dashboard, verifies Delivery
Remediation Center, Remediation Action Ledger, and Remediation SLA markers,
checks the tenant delivery-remediation API, records a dry-run push retry plan,
rejects live-send remediation mode, verifies the audit event, and reruns the
screenshot-free dashboard layout contract. Newman runs
`postman/tracedeck-backend-phase46.postman_collection.json` against the same
live demo and covers remediation summary, plan creation, mode rejection, and
audit evidence.

Phase 47 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase48.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase48.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase48.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase50.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase50.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase50.ps1
```

The Phase 47 smoke live-boots the seeded dashboard, verifies Premium
Notification Command Center, Notification Assurance Funnel, Mail And Push
Delivery Proof, and Customer Action SLAs markers, checks the tenant
notification-command-center API, verifies strict privacy/forbidden markers, and
reruns the screenshot-free dashboard layout contract. Newman runs
`postman/tracedeck-backend-phase47.postman_collection.json` against the same
live demo and covers dashboard markers, typed alert/delivery/action proof, and
the compatibility of the existing alert inbox.

The Phase 48 smoke live-boots the seeded dashboard, verifies Growth Cockpit,
Paid product promise, Anomaly Notification Ops, Notification Delivery Proof,
and Monetisation Owner Actions markers, checks the notification-command-center,
monetisation-summary, and operations-summary APIs, verifies strict
privacy/forbidden markers, and reruns the screenshot-free dashboard layout
contract. Newman runs
`postman/tracedeck-backend-phase48.postman_collection.json` against the same
live demo and covers dashboard markers plus the typed notification,
monetisation, and operations data that powers the paid UI.

The Phase 49 smoke live-boots the seeded dashboard, verifies Notification
Preference Center, Preference Rule Matrix, Study-Safe Suppression, and
Preference Owner Actions markers, checks seeded GET
`notification-preferences`, posts a typed metadata-only preference update,
verifies strict privacy/forbidden markers, and reruns the screenshot-free
dashboard layout contract. Newman runs
`postman/tracedeck-backend-phase49.postman_collection.json` against the same
live demo and covers dashboard markers, seeded preferences, update validation,
invalid cadence rejection, and audit proof.

The Phase 50 smoke live-boots the seeded dashboard, verifies Business
Dashboard, Anomaly Notification Inbox, Push And Mail Proof, Paid Package Value,
and Customer Owner Actions markers, checks the tenant `business-dashboard` API,
verifies product score, mail/push/dashboard proof, paid packages, actions, and
strict privacy/forbidden markers, and reruns the screenshot-free dashboard
layout contract. Newman runs
`postman/tracedeck-backend-phase50.postman_collection.json` against the same
live demo and covers dashboard markers plus the typed business dashboard
contract.

Phase 13 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-risky-software.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase13.ps1
```

The Phase 13 smoke builds and boots the backend on localhost, enrolls a host,
verifies the seeded risky software anomaly, and confirms the dashboard includes
the risky software watchlist. Newman runs
`postman/tracedeck-backend-phase13.postman_collection.json` against a live
API-key-protected backend.

Phase 12 adds:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-device-health.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase12.ps1
```

The Phase 12 smoke builds and boots the backend on localhost, creates a tenant,
enrolls a host, verifies the standalone device-health API, confirms the health
payload is included in host overview, checks monetisation-ready dashboard panels,
and verifies weekly report readiness. Newman runs
`postman/tracedeck-backend-phase12.postman_collection.json` against a live
API-key-protected backend.

`go test -race ./...` is run when the local Go toolchain supports it. On
Windows shells where CGO is disabled or no race-capable C toolchain is active,
the verification script logs a warning and continues.
