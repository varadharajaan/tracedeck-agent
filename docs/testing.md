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
