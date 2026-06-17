# Testing

This is the short testing guide. The old phase-by-phase testing log lives under
`docs/reference/legacy-phase-notes/testing.md`.

## Runtime Smoke

```powershell
python ./devctl.py doctor --skip-cloud
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-doctor.ps1 -IncludeCloud
```

## Agent And Backend

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-agent-live-health.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1
```

## Dashboard

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl http://127.0.0.1:18080
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 -BaseUrl http://127.0.0.1:18080
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -BaseUrl http://127.0.0.1:18080
```

## Cloud

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/check-live-s3-metrics.ps1
python ./devctl.py cloud smoke
python ./devctl.py doctor
```

## Quality

```powershell
python ./devctl.py test quality
```

## Hygiene

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
```
