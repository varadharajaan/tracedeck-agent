# Dashboard Demo Lifecycle

Phase 24 hardens the local dashboard demo launcher so live boot testing proves
the currently built backend is the process serving the dashboard.

## Problem

A stale TraceDeck demo listener can keep `127.0.0.1:18080` occupied. Without a
targeted listener check, a new demo process can exit on bind failure while the
health probe still succeeds against the old process.

## Behavior

`scripts/local/start-dashboard-demo.ps1` now:

- accepts `-DataPath` so smoke tests can use isolated backend state
- calls `scripts/local/stop-backend-dev.ps1 -Addr <addr>` before starting
- stops only TraceDeck-owned listeners for the requested address
- refuses to stop non-TraceDeck processes
- fails fast if the started backend exits during startup

`scripts/local/stop-backend-dev.ps1` now:

- stops the pid-file process when present
- stops a TraceDeck-owned listener on the requested address
- handles dashboard demo process names as well as backend process names
- keeps broad orphan cleanup available only when no address is supplied

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase24.ps1
```

The smoke starts the dashboard twice on the same port, verifies the first
process is stopped, verifies the second process owns the listener, checks the
current dashboard HTML, and verifies the seeded tenant operations API.

## Phase 90 Runtime Doctor Guard

Demo launchers seed buyer-demo rows, including notification proof, but those
rows are not live delivery proof. `devctl.py doctor` now verifies default
`/alert-deliveries` hides `source_kind=demo_seed` rows and records explicit
`include_demo=true` proof separately. This keeps local demos useful while
preventing seeded mail/push rows from being described as received alerts.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-devctl-runtime-doctor.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase90.ps1
```
