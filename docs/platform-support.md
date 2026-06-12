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
requires Accessibility permission for active app/window observation. This phase
does not request permissions automatically.

## Linux

Service manager: `systemd`.

Phase 7 adds a systemd template:

```text
deployments/service/linux/tracedeck-agent.service.tmpl
```

Foreground app collection is marked `partial` because X11 and Wayland expose
active-window information differently. Wayland support depends on compositor and
desktop portal behavior.

## Manifest Rendering

Render local review copies with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1
```

Generated manifests are written under:

```text
data/local/service-manifests/phase7/
```

They are not committed. Installer packaging and native service install/start
commands remain future work.
