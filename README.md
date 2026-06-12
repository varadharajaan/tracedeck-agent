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
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase8.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase15.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase17.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase35.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase44.ps1
go run ./agent/cmd/tracedeck-agent validate-config --config ./examples/policies/ai-btech-student.yaml
go run ./agent/cmd/tracedeck-agent schema --version v1alpha1 --out ./docs/schema/policy-v1alpha1.schema.json
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-policy-schema.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-autostart-assurance.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Action status -DryRun
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase17.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase44.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase44.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1
```

All repeatable setup and verification work is kept under `scripts/`, and script
logs are written under `logs/local/`.

Phase 43 adds a buyer operations brief and screenshot-free dashboard layout
contract. The first-screen UI now makes anomaly alerting, mail proof, push
notification dispatch, weekly report delivery, archive retention, trust/audit,
delivery command, packaging snapshot, and next commercial action visible for a
monetisation demo. The Playwright check records layout metrics only under
`data/local/dashboard-layout/` and does not capture screenshots, video,
credentials, or page content.

Phase 44 adds provider-safe delivery drilldown. The backend exposes
`/api/v1/tenants/{tenantId}/delivery-drilldown` for current route proof and
dry-run rehearsals across email, push, and dashboard routes. The dashboard shows
route score, channel readiness, route evidence, and next actions without sending
live messages or storing provider secrets, alert bodies, endpoint payloads, or
sensitive content.
