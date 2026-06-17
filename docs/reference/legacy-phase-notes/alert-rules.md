# Alert Rules Builder

Phase 19 adds the first no-code alert rule builder contract.

Rule templates are static product recipes:

- non-study YouTube over limit
- media playback after hours
- risky software detected
- archive backlog over limit

Saved rules are tenant-scoped and persisted in backend JSON state. They include
typed trigger, severity, channels, condition, enabled flag, and timestamps.

Endpoints:

```text
GET  /api/v1/alert-rule-templates
GET  /api/v1/tenants/{tenantId}/alert-rules
POST /api/v1/tenants/{tenantId}/alert-rules
```

Example create request:

```json
{
  "template_id": "risky_software_detected",
  "name": "Email when risky software appears",
  "trigger": "risky_software",
  "severity": "high",
  "channels": ["email", "dashboard"],
  "condition": {
    "subject": "category",
    "operator": "equals",
    "value": "torrent_client"
  },
  "enabled": true
}
```

The builder stores policy automation metadata only. It does not add new
collectors and does not collect passwords, keystrokes, cookies, tokens, private
messages, camera, microphone, screenshots, raw URLs, or raw page titles.
