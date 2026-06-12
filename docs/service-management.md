# Service Management

Phase 15 adds one local wrapper for native agent lifecycle operations across
Windows, macOS, and Linux:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Action status
```

Supported actions:

- `install`
- `start`
- `stop`
- `status`
- `uninstall`

Supported platforms:

- `windows`: Task Scheduler through the Phase 8 XML/register/status scripts
- `darwin`: launchd with the Phase 7 plist template
- `linux`: systemd with the Phase 7 service template

Use `-DryRun` to generate a structured action plan without mutating the host:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action install -BuildAgent -DryRun
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform linux -Action status -DryRun
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform darwin -Action start -DryRun
```

Dry-run action plans are written under:

```text
data/local/service-actions/phase15/
```

The Windows install path delegates to:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/register-windows-task.ps1 -BuildAgent -StartAfterRegister
```

That registration path requests UAC elevation when required. The task is hidden
from console flicker and starts at user logon after reboot, but it remains a
normal Task Scheduler entry that can be queried, stopped, and removed.

Phase 39 adds `scripts/local/test-autostart-assurance.ps1` for a stronger
Windows lifecycle check. It verifies the rendered Task Scheduler XML, confirms
continuous no-flicker startup settings, writes typed missing-task status JSON,
and dry-runs Windows service install/status plans without mutating the host.

macOS and Linux install actions render launchd/systemd manifests under
`data/local/service-manifests/phase15/`, then copy them to the native service
location and enable/start the service when run without `-DryRun`.

The wrapper does not add new collectors or relax the privacy contract. It only
manages the lifecycle for the existing agent command.

## Phase 66 Deployment Proof

Phase 66 adds a typed Deployment Readiness Center for service rollout proof.
The dashboard and API show Windows Task Scheduler, macOS launchd, Linux systemd,
manifest templates, rendered output paths, live boot status, autostart status,
background start status, offline replay readiness, archive backlog, and owner
actions. The proof remains metadata-only and does not install, start, stop, or
remove native services unless an operator explicitly runs the service scripts.
