# Windows Autostart

Phase 8 adds Windows Task Scheduler support for reboot persistence.

## What It Does

- Renders a Task Scheduler XML file from a committed template.
- Registers `\TraceDeck\TraceDeck Agent` with a user-logon trigger.
- Starts the compiled agent in continuous mode through a no-console launcher:

```text
wscript.exe run-agent-task-hidden.vbs -File run-agent-task.ps1 ...
tracedeck-agent run --config <policy> --data-dir <data> --log-dir <logs> --outbox-dir <outbox> --collection-interval 10m --max-cycles 0
```

- Uses `StartWhenAvailable` and restart-on-failure settings so missed starts
  can recover after reboot or temporary unavailability.
- Provides a status script that reports task state, last run time, last result,
  next run time, and missed run count.

## Register

Run from an interactive PowerShell session:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/register-windows-task.ps1 -BuildAgent
```

The script requests UAC elevation when needed because registering a reliable
logon task can require administrator approval. The `-BuildAgent` option writes
the scheduled-task executable under `data/local/install/windows/`. The elevated
PowerShell relaunch requests a hidden window after the UAC prompt. Normal
scheduled starts execute `wscript.exe`, which starts the PowerShell runner with
`-NonInteractive -WindowStyle Hidden`; the runner then starts the agent hidden
with stdout/stderr redirected under `logs/local/agent-live`.

To register and start immediately:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/register-windows-task.ps1 -BuildAgent -StartAfterRegister
```

Phase 15 also exposes the same flow through the cross-platform service manager:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action install -BuildAgent
```

## Query Status

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1 -OutputPath data/local/service-status/windows-task.json
```

Equivalent wrapper commands:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action status
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action start
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action stop
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action uninstall
```

The status script writes structured JSON to the console and logs under
`logs/local/service/`. With `-OutputPath`, it also writes the same JSON under
`data/local/`. With `-AllowMissing`, it returns a typed missing-task object
instead of failing, including `present=false`, `state=missing`,
`last_task_result`, `next_run_time`, and missed-run fields so future status
checks can distinguish "not installed yet" from a broken query.

## Live Repair And Proof

To repair this machine's live agent task and start it silently:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/repair-live-agent-autostart.ps1 -SkipBuild
```

To remove stale backend smoke tasks and verify S3 archive object metadata:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/cleanup-stale-backend-dev-tasks.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/check-live-s3-metrics.ps1
```

The backend dev task uses the same hidden launcher pattern when registered by
`scripts/local/start-backend-dev-task.ps1 -ForceRegister`, so reboot/logon
startup does not launch a visible PowerShell console for either live process.

## Assurance Test

Phase 39 adds a stronger local assurance script:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-autostart-assurance.ps1
```

The script renders the Task Scheduler XML under `data/local/`, parses it, and
verifies hidden startup, logon delay, continuous agent mode, restart-on-failure,
start-when-available, battery behavior, structured missing-task status JSON,
and Windows service-manager dry-run install/status plans.

## Transparency Boundary

The task runs the agent in the background to avoid a console-window flicker, but
this is not a covert-monitoring feature. TraceDeck policy still requires
transparent, consent-based monitoring and a visible indicator before expanding
interactive collection behavior.

## Deployment Readiness Center

Phase 66 surfaces Windows reboot-persistence proof in the
`deployment-readiness-center` API and dashboard panel. It shows the Task
Scheduler service manager label, XML template path, rendered output path,
registration script, status query script, UAC/admin-approved install mode,
background startup proof, and owner action. It is a readiness view only; actual
registration still requires running the explicit Windows scripts.
