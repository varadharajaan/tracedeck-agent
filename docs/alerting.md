# Alerting

Alerts are generated from policy violations and anomaly events. Email delivery
uses a notifier interface with provider-specific adapters.

The family profile sends high and critical alerts to `varathu09@gmail.com`.
Alert payloads may include app, domain, category, reason, severity, event id,
and media metadata only when policy allows it.
