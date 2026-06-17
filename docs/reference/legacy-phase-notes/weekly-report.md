# Weekly Report

Phase 14 generates a weekly report from the current typed host overview.

Endpoints:

```text
GET /api/v1/devices/{deviceId}/reports/weekly
GET /api/v1/devices/{deviceId}/reports/weekly/pdf
```

The JSON report includes:

- summary
- study, coding, compliance, health, and risk highlights
- policy/anomaly/tamper risks
- email subject and preview
- email readiness flag
- PDF readiness flag
- generated timestamp

The PDF endpoint returns a small generated `application/pdf` payload suitable
for local packaging and future email attachment delivery.

## Privacy Boundary

The weekly report is generated from already typed backend overview data. It does
not collect raw URLs, page titles, cookies, tokens, credentials, keylogs,
private messages, camera, microphone, screenshots, or raw executable paths.
