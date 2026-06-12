# Windows Autostart

Phase 8 adds Windows Task Scheduler support for reboot persistence.

## What It Does

- Renders a Task Scheduler XML file from a committed template.
- Registers `\TraceDeck\TraceDeck Agent` with a user-logon trigger.
- Starts the compiled agent in continuous mode:

```text
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
the scheduled-task executable under `data/local/install/windows/`.

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
```

Equivalent wrapper commands:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action status
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action start
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action stop
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action uninstall
```

The status script writes structured output to the console and logs under
`logs/local/service/`.

## Transparency Boundary

The task runs the agent in the background to avoid a console-window flicker, but
this is not a covert-monitoring feature. TraceDeck policy still requires
transparent, consent-based monitoring and a visible indicator before expanding
interactive collection behavior.
