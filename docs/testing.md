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

`go test -race ./...` is run when the local Go toolchain supports it. On
Windows shells where CGO is disabled or no race-capable C toolchain is active,
the verification script logs a warning and continues.
