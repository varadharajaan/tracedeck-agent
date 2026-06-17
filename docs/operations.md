# Operations

## Local Backend

```powershell
python ./devctl.py server task-start
python ./devctl.py server task-status
python ./devctl.py server task-restart
python ./devctl.py server task-stop
```

The local dashboard runs at:

```text
http://127.0.0.1:18080/
```

## Live Agent

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/repair-live-agent-autostart.ps1 -SkipBuild
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-agent-live-health.ps1
```

## Windows Autostart

TraceDeck uses Task Scheduler. The current live startup path is:

```text
Task Scheduler -> wscript.exe -> hidden PowerShell runner -> TraceDeck process
```

Both the agent and backend task use `wscript.exe` and the hidden launcher
`scripts/local/run-agent-task-hidden.vbs`.

## S3 Archive

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-s3-archive.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/check-live-s3-metrics.ps1
```

## Cloud Admin

```powershell
python ./devctl.py sam deploy
python ./devctl.py sam outputs
python ./devctl.py cloud smoke
python ./devctl.py doctor
```

## Logs And Outputs

Runtime logs:

```text
logs/local/
```

Generated reports and stack outputs:

```text
data/local/output/
```

Root-level generated files should not be created. Use:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
```
