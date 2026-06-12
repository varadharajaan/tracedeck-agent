# Alerting

Alerts are generated from policy violations and anomaly events. Email delivery
uses a notifier interface with provider-specific adapters.

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

The evaluator applies `alerts.email.min_severity` after rule evaluation, so a
rule below the configured minimum is intentionally suppressed. Dry-run alert
notifications are JSON files under `data/local/outbox/alerts/`; provider-backed
email delivery remains behind the notifier adapter.
