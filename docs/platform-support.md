# Platform Support

TraceDeck is Go-first and structured for Windows, macOS, and Linux.

| Area | Windows | macOS | Linux |
| --- | --- | --- | --- |
| Agent build | Supported | Supported by Go target | Supported by Go target |
| Service manifest | Task Scheduler | launchd | systemd |
| Local backend | Supported | Supported by Go target | Supported by Go target |
| Browser domain metadata | Supported | Planned adapter expansion | Planned adapter expansion |
| Foreground app metadata | Adapter-backed | Adapter-backed | Adapter-backed |
| Software inventory metadata | Adapter-backed | Adapter-backed | Adapter-backed |

Manifests:

```text
deployments/service/windows/
deployments/service/darwin/
deployments/service/linux/
```

Service manager:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Action status
```
