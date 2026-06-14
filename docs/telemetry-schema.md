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
