# Alerting

Alerts are generated from policy violations and anomaly events. Delivery uses a
notifier interface with provider-specific adapters for email and browser push.

The family profile sends medium, high, and critical alerts to
`varathu09@gmail.com`.
Alert payloads may include app, domain, category, reason, severity, event id,
and media metadata only when policy allows it.

Phase 2 evaluates blocked app alerts from process snapshot events and writes
dry-run notification payloads under `data/local/outbox/alerts/`.

Phase 4 extends the evaluator into a small policy/anomaly engine:

- `blocked_app_opened` raises one alert per blocked process name observed in a
  snapshot.
- `blocked_domain_opened` raises one alert per blocked browser domain observed
  from domain-only browser activity.
- `non_study_youtube` raises an alert when YouTube activity is categorized as
  video streaming and not marked as study-related by policy keywords.

The evaluator applies the lowest enabled channel severity threshold after rule
evaluation. Each provider still filters its own delivery batch, so an email
route can stay `high` while a push route receives `medium` anomaly alerts.

By default, the agent uses `--alert-dry-run=true` and writes notification JSON
files under `data/local/outbox/alerts/`. Phase 17 adds provider-backed delivery
when the operator explicitly runs with `--alert-dry-run=false`:

- `alerts.email.provider: smtp` sends through an SMTP relay configured only by
  environment variables.
- `alerts.email.provider: ses` sends through AWS SESv2 using the default AWS SDK
  credential chain.
- `alerts.push.provider: web_push` sends browser notifications through VAPID and
  a local Push API subscription file.
- `alerts.email.from` is required when alerts are enabled.

SMTP environment variables:

```text
TRACEDECK_SMTP_HOST
TRACEDECK_SMTP_PORT
TRACEDECK_SMTP_USERNAME
TRACEDECK_SMTP_PASSWORD
TRACEDECK_SMTP_SERVER_TLS
```

SMTP credentials are never stored in policy YAML or alert payloads. The Phase
17 live smoke uses `scripts/tools/fake-smtp` to capture a local `.eml` under
`data/local/` and verify delivery without sending real email.

Web Push setup is scriptable:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/generate-webpush-vapid-keys.ps1
```

Then open the local dashboard, click `Enable Push`, and allow notifications.
The browser subscription is saved under `data/local/webpush/subscriptions.json`.
If the dashboard reports `Notifications are blocked`, reset the site
notification permission in the browser for `http://127.0.0.1:18080`, then click
`Enable Push` again. If VAPID keys were rotated, the dashboard detects the
stale browser subscription key, unsubscribes it, and creates a fresh
subscription with the current public key.
Start the agent with `-LivePush -LiveAlerts` only after that subscription file
exists:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-agent-live-local.ps1 -LiveArchive -LiveAlerts -LivePush
```

The subscription endpoint is stored locally so the provider can send browser
notifications. It is not returned by dashboard APIs, archived in S3 telemetry,
or included in alert payloads.

Phase 19 adds the no-code alert rules builder API and dashboard panels. Rule
templates and saved tenant rules describe triggers, severity, channels, and
conditions for policy automation. This is configuration metadata only; it does
not add new collection capabilities or store forbidden sensitive content.
