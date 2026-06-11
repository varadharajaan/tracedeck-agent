# Testing

Local verification replaces GitHub Actions for this project.

Run:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/setup/install-go-tools.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase0.ps1
```

Verification logs are written under `logs/local/verify/`.

`go test -race ./...` is run when the local Go toolchain supports it. On
Windows shells where CGO is disabled or no race-capable C toolchain is active,
the verification script logs a warning and continues.
