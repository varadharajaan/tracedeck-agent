# Windows Autostart

TraceDeck uses Task Scheduler for reboot persistence on Windows.

Current live startup chain:

```text
Task Scheduler
  -> C:\Windows\System32\wscript.exe
  -> scripts/local/run-agent-task-hidden.vbs
  -> powershell.exe -NoProfile -NonInteractive -WindowStyle Hidden
  -> TraceDeck process
```

Repair the live agent task:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/repair-live-agent-autostart.ps1 -SkipBuild
```

Check the registered task:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1
```

The backend dev task is also registered through the same hidden launcher when
started with:

```powershell
python ./devctl.py server task-start
```
