# Device Health Score

Phase 12 adds a privacy-safe device health score for endpoint productivity and
risk observability.

The agent emits a `device.health.observed` event with aggregate operational
metrics only:

- CPU usage percent
- memory usage percent
- disk usage percent
- boot time
- uptime
- score
- status

The backend exposes the current typed score through:

```text
GET /api/v1/devices/{deviceId}/health
```

The dashboard also includes the score in host overview and shows operational
panels for CPU, memory, disk, heartbeat, battery placeholder, startup app
placeholder, crash placeholder, and recommendation text.

## Scoring Model

The first implementation uses a lightweight weighted score:

- CPU pressure lowers the score.
- memory pressure lowers the score.
- disk pressure lowers the score.
- the minimum score is clamped so the dashboard can still display a meaningful
  watch state.

Status mapping:

```text
healthy   score >= 85
watch     score >= 65 and score < 85
attention score < 65
```

## Privacy Boundary

Device health does not collect passwords, keystrokes, cookies, tokens, private
messages, camera, microphone, screenshots, raw URLs, or page titles. It is
intended for endpoint reliability and paid readiness features such as school,
family, and business device health views.
