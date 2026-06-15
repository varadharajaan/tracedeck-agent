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

## Phase 91 Persistent Local Backend

`devctl.py server start` starts a backend process for the command session, which
can be reaped by the harness after the command exits. Phase 91 adds a hidden
Windows scheduled-task launcher for local admin testing:

```powershell
python ./devctl.py server task-start
python ./devctl.py server task-status
python ./devctl.py server task-restart
python ./devctl.py server task-stop
```

The task runner builds the backend under `data/local/backend/`, writes the pid
and ready files under `data/local/backend/`, redirects stdout/stderr under
`logs/local/backend/`, seeds the demo tenant/device, and waits on the backend
process so it stays alive after the original devctl command returns.

Task status intentionally separates two proofs:

- `runtime_ok=true` means the pid is running and `/health` returned `ok`.
- `launch_task_verified=true` means the Windows scheduled task was readable.
- `task_state=inaccessible` means Windows denied Scheduler readback; it is not
  reported as a missing task.
- `scheduler_readback=denied` with `runtime_evidence=pid_and_health` is a
  healthy non-elevated runtime proof, not a claim that reboot persistence was
  fully re-read from Scheduler metadata.

Use `powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-dev-task.ps1 -LeaveRunning`
when you want to leave `http://127.0.0.1:18080` available for manual browser
checks after the smoke passes.

Phase 92 adds a focused resilience guard:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-task-status-resilience.ps1
python ./devctl.py test phase92
```

That guard prevents the local tooling from turning Windows `Access denied`
Scheduler readback into a false missing-task failure while still failing if the
task is truly missing or the backend runtime proof is unhealthy.
If Windows denies isolated scheduled-task creation in the current shell, the
Phase 92 smoke uses the default `http://127.0.0.1:18080` task-status output as
the runtime proof and still runs live provenance plus runtime doctor checks.

Phase 93 adds a typed advisory to the same task-status JSON:

- `advisory.severity` is `ok`, `watch`, or `action_required`.
- `advisory.code` explains the exact condition, such as
  `scheduler_verified_runtime_ready` or
  `runtime_ready_scheduler_readback_denied`.
- `advisory.can_continue=true` means local dashboard testing can proceed.
- `advisory.admin_readback_recommended=true` means an elevated PowerShell
  session is useful for full Task Scheduler metadata, but the script does not
  bypass UAC or hide the task.
- `advisory.operator_action` gives the next command or operator step.

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase93.ps1
python ./devctl.py server task-status
python ./devctl.py test phase93
```
