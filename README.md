# TraceDeck Agent

TraceDeck Agent is a Go-first, privacy-aware endpoint activity, productivity,
and risk observability agent for Windows, macOS, and Linux laptops and managed
devices.

It tracks typed endpoint metadata such as application usage, browser
domain/category activity, software inventory changes, policy violations, S3
archive health, alert delivery health, and agent health using OpenTelemetry.

TraceDeck is not credential capture or covert surveillance. It does not collect
passwords, keystrokes, browser cookies, auth tokens, private messages, camera,
microphone, or hidden screen content. Browser monitoring is domain/category
based by default.

## Local Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase0.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase7.ps1
go run ./agent/cmd/tracedeck-agent validate-config --config ./examples/policies/ai-btech-student.yaml
go run ./agent/cmd/tracedeck-agent schema --out ./docs/schema/policy-v1alpha1.schema.json
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1
```

All repeatable setup and verification work is kept under `scripts/`, and script
logs are written under `logs/local/`.
