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

`go test -race ./...` is run when the local Go toolchain supports it. On
Windows shells where CGO is disabled or no race-capable C toolchain is active,
the verification script logs a warning and continues.
