# Windows Autostart

TraceDeck uses Task Scheduler for reboot persistence on Windows.

Current live startup chain:

```text
Task Scheduler
  -> tracedeck-agent.exe run ...
```

Repair the live agent task:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/repair-live-agent-autostart.ps1 -SkipBuild
```

Check the registered task:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1
```

The backend dev task is separate from the production agent task. Start it only
when a local dashboard server is needed:

```powershell
python ./devctl.py server task-start
```
