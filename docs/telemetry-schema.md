# Telemetry Schema

Telemetry events are metadata-only JSON records with stable event types and
source labels.

Common event types:

- `agent.health.heartbeat`
- `process.observed`
- `foreground_app.observed`
- `browser.domain.observed`
- `device.health.observed`
- `software.installed`
- `software.uninstalled`

Common fields:

- `tenant_id`
- `device_id`
- `host_name`
- `event_type`
- `source`
- `observed_at`
- `source_kind`

Schema and constants live under:

```text
agent/internal/constants/
backend/internal/constants/
docs/schema/policy-v1alpha1.schema.json
```
