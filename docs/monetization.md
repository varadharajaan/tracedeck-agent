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

Phase 21 adds managed-policy packaging:

- device groups for family, school, coaching center, and business cohorts
- policy assignments by tenant, device group, or device
- seeded primary group and assignment for new tenants
- audit evidence for group creation and assignment rollout
- dashboard visibility for assignment mode, status, target, policy template,
  and alert-rule coverage

Phase 22 adds data-rights packaging for paid trust plans:

- tenant export manifests
- delete request queue
- audit events for export and delete workflows
- dashboard proof for compliance and family trust conversations

Phase 23 adds customer operations packaging for monetisation demos:

- tenant-level fleet coverage and hosts needing attention
- anomaly, policy, tamper, and archive pressure summary
- mail delivery proof for alerts and reports
- push notification reach and retry visibility
- escalation workbench for customer follow-up
- upgrade proof pack for Family Pro, school, and business buyers

Phase 29 adds a buyer-ready first-screen launch deck:

- customer package, readiness score, notification score, trust score, and
  conversion stage
- anomaly push assurance with route status, recipient, provider, and proof
- mail delivery assurance with critical alert email proof
- weekly report proof with email and PDF readiness
- host risk command with the top policy, anomaly, or tamper signal
- S3 archive retention proof and backlog status
- notification revenue stream across email, push, and dashboard channels
- immediate buyer action, route proof, and upgrade-lever prompts

This still uses the existing privacy-aware API data: typed risk categories,
delivery routes, report readiness, health, archive, role, retention, and policy
template metadata. It does not add credential, keylog, private-message, raw URL,
page-title, camera, microphone, or covert screen collection.

Phase 36 upgrades the embedded dashboard into a stronger revenue command
surface:

- revenue command center for paid-plan outcome, conversion stage, seats, and
  buyer readiness
- monetisation value stack for host coverage, anomaly queue, mail delivery,
  push reach, weekly report, S3 archive retention, consent/audit trust, and
  upgrade lever
- notification proof rail for anomaly alert, email delivery, push delivery,
  dashboard inbox, and report mail proof
- buyer demo checklist for route proof, report/PDF, archive, consent/data
  rights, data export/delete readiness, and saved buyer views

The phase adds presentation and verification only; it does not add sensitive
collectors.

Phase 38 adds a Commercial Control Room as the first buyer-facing layer:

- host coverage and hosts needing attention
- anomaly command with the top policy, anomaly, or tamper signal
- email proof for alert and report delivery
- push proof for immediate anomaly notification
- weekly report mail/PDF readiness
- upgrade trigger and conversion-stage visibility
- alert delivery evidence across anomaly, email, push, dashboard, and report
  proof
- provider-safe delivery drilldown for dry-run email, push, and dashboard route
  rehearsal without provider secrets or alert body storage
- customer success queue for parent, school, coaching center, and business
  buyers

This keeps TraceDeck positioned as endpoint productivity and risk observability
rather than a narrow monitoring tool, while continuing to use the existing
privacy-aware typed APIs.

Phase 45 turns the top of the dashboard into a monetisation command center:

- anomaly and notification inbox for policy, anomaly, and tamper urgency
- push notification reach and retry state
- mail delivery proof for critical alerts
- weekly report mail and PDF readiness
- fleet coverage and selected-host context
- S3 archive retention and backlog status
- trust center with visible monitoring, agent health, and audit posture
- revenue package and paid capability signal
- delivery and mail proof list for email, push, dashboard, and reports
- owner action queue for parent, school, coaching center, and business buyers

The phase is presentation and verification only. It improves the paid-product
story without adding sensitive collectors or live provider sends.

Phase 46 turns route failures into a monetisable trust workflow:

- delivery remediation center for anomaly push, critical mail, dashboard inbox,
  and weekly report delivery recovery
- route recovery score, open problem count, owner acknowledgement, and SLA watch
- dry-run retry plans that create audit proof without sending provider payloads
- command-center delivery proof that includes remediation state
- Newman and smoke coverage for summary, dry-run plan creation, live-send mode
  rejection, and audit evidence

This supports Family Pro, school, and small-business packaging because buyers
pay for reliable alerting, owned recovery, and proof that mail/push routes are
working or being repaired.

Phase 47 adds a premium notification command center:

- typed `notification-command-center` API for a single buyer-facing alert and
  delivery contract
- alert funnel for anomaly, policy, and tamper urgency
- mail and push delivery proof with provider-safe route evidence
- route assurance state from delivery drilldown and remediation
- customer action SLAs with owner, severity, channel, paid tier, and next step
- dashboard jump target for Notify Pro so notification value is visible as a
  paid product surface, not a background technical table

This makes notification reliability easier to monetise for Family Pro, school,
coaching center, and business packaging while preserving the metadata-only
privacy boundary.

Phase 48 adds a Growth Cockpit as the first paid-product surface:

- revenue readiness and package stage
- anomaly notification operations for policy, anomaly, and tamper alerts
- mail delivery proof for critical alerts and weekly report delivery
- push reach for immediate anomaly notifications
- archive retention and weekly report proof in the same executive view
- trust and consent posture for legitimacy
- monetisation owner actions for parent, school, coaching center, and business
  buyers

This gives TraceDeck a stronger monetisable front door without adding new
collectors. The panel aggregates existing typed, privacy-aware backend
metadata and keeps passwords, screenshots, raw URLs, page titles, alert bodies,
provider secrets, and private content out of scope.

Phase 49 adds a Notification Preference Center:

- immediate, digest, and silent notification policies
- email, push, and dashboard channel coverage
- quiet hours with critical bypass metadata
- escalation owner and repeat policy
- study-safe suppression for learning activity
- paid-tier labels, retention evidence, and owner actions

This is monetisable because parents, schools, coaching centers, and small
businesses need confidence that alerting is configurable and explainable, not
just noisy. The feature stores typed policy metadata only and avoids provider
secrets, alert bodies, screenshots, raw URLs, page titles, and private content.

Phase 50 adds a Business Dashboard as the top paid-product surface:

- customer health and product readiness headline
- anomaly notification inbox with mail, push, and dashboard status
- push and mail route proof for paid alert reliability
- route proof and preference policy readiness
- archive and weekly report value in the same cockpit
- Family Pro, School, and Business package cards
- customer owner actions with owner, SLA, source, and paid tier

This makes TraceDeck easier to sell because a buyer can see value, proof, and
next action without digging through host-level tables. It composes existing
typed metadata-only APIs and does not add sensitive collectors.

Phase 51 adds a Notification Evidence Timeline:

- host-level delivery history for anomaly email, push, and dashboard routes
- notification score with delivered, retrying, failed, and suppressed counts
- mail delivery proof and push retry proof in a buyer-friendly audit trail
- route proof gaps and next retry timing for operational confidence
- paid-tier recommendation so notification reliability maps to Family Pro,
  school, and business packaging

This helps monetisation because customers pay for trusted alert delivery, not
only detection. The feature stays provider-safe and metadata-only.

Phase 52 adds a Role Experience Center:

- parent view for family alert proof, weekly report readiness, consent, and
  study-safe monitoring
- student self-view for transparency, study-safe suppression, sync proof, and
  data rights
- school admin view for cohorts, policy rollout, offline replay, route proof,
  and audit history
- business manager view for productivity/risk observability, delivery audit,
  archive trust, and compliance export value
- paid onboarding checklist that maps each role to Family Pro, School, or
  Business packaging

This turns the same endpoint metadata into multiple paid buying stories without
adding sensitive collectors.

Phase 53 adds an Executive Notification Console as the new first-screen buyer
surface:

- sellable readiness headline and recommended paid package
- anomaly alert stream with mail, push, and dashboard delivery state
- mail delivery proof for critical alerts and weekly report trust
- push reach for urgent anomaly notification
- weekly report readiness, archive backlog, role-view packaging, and route gaps
- paid value tiles and owner actions for Family Pro, School, and Business demos

This makes the dashboard feel closer to a monetisable operations product:
buyers can see what happened, who was notified, whether mail/push worked, what
is ready for reporting/archive, and what action remains. It composes existing
typed metadata-only proof and does not add sensitive collectors.
