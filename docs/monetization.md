# Monetization

TraceDeck is positioned as privacy-aware endpoint productivity and risk
observability.

Paid features include weekly AI reports, policy templates, compliance scoring,
risky software detection, role-based dashboards, no-code alert rules, consent
and audit center, archive retention plans, AI classification, and tamper
detection.

Phase 6 adds the first backend representation for monetizable packaging:

- Plans: Free, Family Pro, School, Business, Enterprise.
- Retention tiers: local-only starter, family cloud archive, school year
  archive, and business compliance.
- Roles: parent, student, school admin, and business manager.
- Tenant readiness profile: selected plan, retention tier, primary profile,
  status, and device limit.
- Audit events: backend administrative actions that can power a consent and
  audit center later.

The initial storage is in-memory and does not perform billing. It creates the
typed API contract needed before Stripe, Paddle, SSO, durable tenant storage, or
remote admin dashboards are added.

Phase 16 makes the embedded dashboard itself more sellable. The host view now
surfaces:

- anomaly notification inbox with routed delivery badges
- mail delivery center for weekly report subject, preview, PDF readiness, and
  last-send status
- push notification routing with provider, retry, attempts, and last error
- email SLA and local dashboard feed route health
- paid trigger and upgrade-path cues for Family Pro, school, and business
  packaging

This still uses the existing privacy-aware API data: typed risk categories,
delivery routes, report readiness, health, archive, role, retention, and policy
template metadata. It does not add credential, keylog, private-message, raw URL,
page-title, camera, microphone, or covert screen collection.
