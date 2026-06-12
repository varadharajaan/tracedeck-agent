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

Phase 18 makes the dashboard feel more like a paid product cockpit than an
operator debug screen. The first screen now surfaces:

- priority action for the most urgent route or risk event
- notification promise across email, push, and dashboard channels
- commercial readiness score for demos and paid packaging
- trust coverage across agent health, S3 archive, dashboard feed, and audit
- executive briefing with top risk, study signal, alert outcome, and archive
  trust
- notification action queue for retrying push routes, mail route issues, and
  open anomaly/policy/tamper items

Phase 19 adds a paid no-code alert rules slice:

- rule recipe catalog for family, school, and business templates
- tenant-scoped saved rules
- typed conditions for app/category/time/window thresholds
- delivery channels for email, push, and dashboard
- dashboard visibility for saved automations and recipe packaging

Phase 20 adds buyer-facing trust and alert proof:

- consent and audit center for visible collection status, recipients, data
  export/delete readiness, and audit history
- alert revenue operations for anomaly coverage, mail delivery proof, push
  notification reach, and customer audit evidence
- push notification center for mobile recipient, anomaly, status, provider,
  retry timing, and last-send state
- static dashboard disclosure that passwords, credentials, screenshots, and
  private content are denied collection categories

This still uses the existing privacy-aware API data: typed risk categories,
delivery routes, report readiness, health, archive, role, retention, and policy
template metadata. It does not add credential, keylog, private-message, raw URL,
page-title, camera, microphone, or covert screen collection.
