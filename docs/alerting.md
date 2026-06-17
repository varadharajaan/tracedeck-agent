# Alerting

TraceDeck separates alert detection from delivery proof.

## Alert Inputs

- Policy violations
- Risky software
- Non-study browser activity
- Archive backlog
- Agent health
- Tamper and service state

## Delivery Routes

- Email
- Web Push
- Dashboard fallback

## Delivery Truth

Delivery rows must say whether they are:

- provider-confirmed
- dry-run rehearsed
- dashboard-visible
- demo-only
- retrying
- failed
- disabled

Seeded demo proof must not be presented as provider-confirmed delivery.
