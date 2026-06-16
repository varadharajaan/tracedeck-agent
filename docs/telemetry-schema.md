# Telemetry Schema

Initial metrics:

```text
tracedeck.device.active.duration_seconds
tracedeck.app.usage.duration_seconds
tracedeck.browser.domain.duration_seconds
tracedeck.browser.category.duration_seconds
tracedeck.browser.history.events
tracedeck.policy.violation.count
tracedeck.anomaly.count
tracedeck.software.risk.count
tracedeck.agent.export.failure.count
tracedeck.agent.buffer.queue_size
tracedeck.agent.s3.upload.count
tracedeck.alert.email.sent.count
```

## Event Types

| Event type | Source | Metadata | Privacy boundary |
| --- | --- | --- | --- |
| `agent.health.heartbeat` | `collector.agent.heartbeat` | `profile`, `operating_system`, `agent_healthy`, `agent_version`, `collection_mode`, `collection_interval`, `archive_enabled`, `archive_due`, `backend_sync_enabled`, `alerts_enabled` | Metadata-only agent readiness proof. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogs, or hidden collection bypass data. |

## OpenTelemetry Export

Phase 109 adds an optional OTLP/HTTP JSON log exporter. The exporter maps each
stored TraceDeck event to one OpenTelemetry log record.

Resource attributes:

```text
service.name
service.version
tracedeck.tenant_id
tracedeck.device_id
host.name
os.type
tracedeck.profile
tracedeck.privacy_boundary
```

Log attributes:

```text
tracedeck.event.id
event.name
event.source
tracedeck.tenant_id
tracedeck.device_id
host.name
process.executable.name
process.pid
tracedeck.path_hash
tracedeck.metadata.*
```

Sensitive metadata keys and raw URL-like values are filtered before export.
The exporter reports `otel_exported`, `otel_events`, `otel_dropped`,
`otel_attempts`, and `otel_backlog` in the agent run result.
