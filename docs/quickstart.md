# Quick Start

## Prerequisites

- Go
- Python 3
- PowerShell
- AWS CLI and SAM CLI for cloud commands

## Start The Local Backend

```powershell
python ./devctl.py server task-start
python ./devctl.py server task-status
```

Open:

```text
http://127.0.0.1:18080/
```

## Check The Runtime

```powershell
python ./devctl.py doctor --skip-cloud
```

Cloud-inclusive check:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-doctor.ps1 -IncludeCloud
```

## Start Or Repair The Live Agent On Windows

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/repair-live-agent-autostart.ps1 -SkipBuild
```

This registers the agent through a hidden `wscript.exe` launcher so normal
logon startup does not show a PowerShell console.

## Verify Live Collection

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-agent-live-health.ps1
```

## Verify S3 Archive

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/check-live-s3-metrics.ps1
```

## Deploy Cloud Admin

```powershell
python ./devctl.py sam build
python ./devctl.py sam deploy
python ./devctl.py sam outputs
python ./devctl.py doctor
```
