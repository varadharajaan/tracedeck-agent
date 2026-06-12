# Policy Config

Policies are YAML files decoded into strongly typed Go structs. Unknown fields
fail validation. Collection modes, archive providers, email providers,
severities, and sensitive capabilities are enums backed by centralized
constants.

Generate the policy schema with:

```powershell
go run ./agent/cmd/tracedeck-agent schema --out ./docs/schema/policy-v1alpha1.schema.json
```

Risky software alerting is controlled by a typed alert rule:

```yaml
alert_rules:
  risky_software_detected:
    enabled: true
    severity: high
```

The classifier stores risk category and reason metadata on process events. It
does not store raw executable paths.

Email alert policy is also strongly typed. When `alerts.enabled` is true, the
policy must include a sender and at least one recipient:

```yaml
alerts:
  enabled: true
  email:
    provider: smtp
    from: alerts@example.com
    to:
      - varathu09@gmail.com
    min_severity: medium
```

Provider credentials are supplied by environment variables, not YAML. SMTP uses
`TRACEDECK_SMTP_*` variables, and SES uses the AWS SDK default credential chain.

## Backend Sync

`backend_sync` controls metadata-only sync from the local agent to the local
backend.

```yaml
backend_sync:
  enabled: false
  base_url: http://127.0.0.1:18080
  batch_limit: 100
  request_timeout: 10s
```

When enabled, the agent posts the current collection cycle after SQLite storage
succeeds. The payload contains typed event metadata only and does not include
passwords, credentials, screenshots, tokens, cookies, keylogs, private messages,
raw URLs, page titles, provider secrets, or raw file contents.
