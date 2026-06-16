# Platform Support

Phase 7 hardens the platform contract for Windows, macOS, and Linux.

## Capability Metadata

Each platform adapter reports:

- operating system id
- service manager id
- process collection support
- local storage support
- typed support rows for service manager, foreground app, software inventory,
  media metadata, browser history, and local indicator

Support status values are centralized constants:

- `supported`
- `requires_permission`
- `partial`
- `planned`
- `unsupported`

Callers can use `Capabilities.Require(capability_id)` and branch on
`platform.ErrUnsupportedCapability` without parsing strings.

## macOS

Service manager: `launchd`.

Phase 7 adds a launchd template:

```text
deployments/service/darwin/io.tracedeck.agent.plist.tmpl
```

Foreground app collection is marked `requires_permission` because macOS
requires Accessibility permission for active app/window observation. Phase 110
does not request permissions automatically; unsupported or permission-required
states are surfaced through typed capability errors and the collector skips
without failing the agent run.

## Linux

Service manager: `systemd`.

Phase 7 adds a systemd template:

```text
deployments/service/linux/tracedeck-agent.service.tmpl
```

Foreground app collection is marked `partial` because X11 and Wayland expose
active-window information differently. Wayland support depends on compositor and
desktop portal behavior.

## Foreground App Collection

Phase 110 implements the collector package and Windows adapter path. The
Windows adapter uses desktop foreground-window APIs only to resolve the active
process id, process name, and executable path hash. It does not call window
title APIs and does not capture screenshots.

The emitted event type is:

```text
foreground_app.observed
```

The source is:

```text
collector.foreground_app
```

macOS and Linux remain explicit adapter states until native permission and
desktop-environment handling are hardened.

## Software Inventory Collection

Phase 111 implements metadata-only software inventory behind the same platform
adapter contract. Windows reads uninstall registry display metadata without
install locations. macOS reads application bundle display names without bundle
contents. Linux reads dpkg package metadata when available and otherwise
reports the capability as unavailable for that host.

The emitted event types are:

```text
software.installed
software.uninstalled
```

The source is:

```text
collector.software.inventory
```

The collector compares local hashed snapshots and emits changes only after a
baseline exists.

## Manifest Rendering

Render local review copies with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1
```

Generated manifests are written under:

```text
data/local/service-manifests/phase7/
```

They are not committed. Phase 15 adds
`scripts/local/manage-agent-service.ps1` as the native install/start/stop/status
wrapper for Windows Task Scheduler, macOS launchd, and Linux systemd. Dry-run
plans are written under `data/local/service-actions/phase15/`, and generated
service manifests for install flows are written under
`data/local/service-manifests/phase15/`.

## Windows Task Scheduler

Phase 8 adds a Windows Task Scheduler template:

```text
deployments/service/windows/tracedeck-agent-task.xml.tmpl
```

Render and validate the XML locally with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1
```

Register the task with UAC elevation when needed:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/register-windows-task.ps1 -BuildAgent
```

Query the registered task:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1
```

The task starts the continuous agent at user logon after reboot. It launches the
agent executable directly and logs through the agent's normal file logger.

The cross-platform service wrapper can call the same Windows path:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action install -BuildAgent
```

## Deployment Readiness Center

Phase 66 exposes the existing service-management proof through
`GET /api/v1/tenants/{tenantId}/deployment-readiness-center` and the dashboard
Deploy panel. The center summarizes Windows Task Scheduler, macOS launchd, and
Linux systemd readiness, committed service templates, local rendered manifest
paths under `data/local/service-manifests/phase66/`, live boot status, autostart
status, background startup status, offline replay readiness, archive backlog,
and owner actions. The verifier renders manifests, checks the Windows task
template, runs autostart assurance, and dry-runs service manager flows without
mutating the host.
