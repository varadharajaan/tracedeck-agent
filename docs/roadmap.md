# Roadmap

1. Phase 0: governance and repo foundation.
2. Phase 1: local Windows agent.
3. Phase 1B: platform adapter contracts and cross-platform builds.
4. Phase 2: S3 archive and email alerting.
5. Phase 2B: continuous runner and bounded local service smoke.
6. Phase 3: browser activity. Domain-only Chromium history collection is added.
7. Phase 4: policy and anomaly engine. Blocked app, blocked domain, and
   non-study YouTube alert evaluation with smoke-tested privacy boundaries.
8. Phase 5: backend and dashboard. Localhost Go backend foundation, device
   enrollment APIs, policy template catalog, archive status, embedded dashboard
   shell, and Newman-backed API verification.
9. Phase 6: SaaS readiness. Localhost backend APIs for tenant readiness,
   subscription plan catalog, role catalog, retention tier catalog, audit
   events, dashboard visibility, and Newman-backed verification.
10. Phase 7: macOS and Linux endpoint support hardening. Typed platform
    capability metadata, unsupported-capability errors, launchd/systemd service
    templates, render script, docs, tests, and cross-platform verification.
11. Phase 8: Windows scheduled-task autostart. Task Scheduler XML, render,
    register, status/query scripts, reboot persistence docs, and local
    verification.
12. Phase 9: host dashboard risk views. Embedded command-center dashboard with
    host filtering, typed host overview, policy violation, anomaly, tamper,
    alert delivery, archive health panels, Postman/Newman coverage, and local
    verification.
13. Phase 10: dashboard demo launcher. Local script starts the backend, seeds a
    demo host, verifies dashboard data, and leaves the UI ready to view.
14. Phase 11: durable backend storage and auth groundwork. JSON-backed backend
    state, optional local API-key middleware, tenant-scoped access checks,
    restart persistence smoke, Postman/Newman coverage, and local verification.
15. Phase 12: device health score and monetisation dashboard. Privacy-safe
    endpoint health events, typed backend health API, persisted host health,
    richer dashboard panels for notification operations, product packaging,
    policy marketplace, retention plans, docs, Postman/Newman coverage, and
    local verification.
16. Phase 13: risky software detection. Process telemetry classification for
    torrent clients, VPN/proxy tools, game launchers, non-standard browsers,
    downloads installers, alert evaluator rule, dashboard watchlist, docs,
    Postman/Newman coverage, and local verification.
17. Phase 14: weekly report generation and packaging. Generated weekly report
    JSON from host overview, email subject/preview readiness, lightweight PDF
    endpoint, dashboard readiness visibility, docs, Postman/Newman coverage,
    and local verification.
18. Phase 15: native service management wrapper. One scripted command surface
    for Windows Task Scheduler, macOS launchd, and Linux systemd
    install/start/stop/status/uninstall dry-runs, docs, and local verification.
19. Phase 16: monetisation dashboard upgrade. Embedded dashboard panels for
    anomaly notification inbox, mail delivery center, push routing, alert route
    SLA visibility, paid trigger, upgrade path, docs, Postman/Newman coverage,
    and local verification.
20. Phase 17: provider-backed email alerts. SMTP and AWS SESv2 notifier
    adapters, required typed sender policy, env-only provider credentials,
    fake-SMTP live smoke, schema/docs updates, cross-platform builds, and local
    verification.
21. Phase 18: product-grade dashboard command center. First-screen priority
    action, notification promise, commercial readiness, trust coverage,
    executive briefing, notification action queue, Postman/Newman coverage, and
    local verification.
22. Phase 19: no-code alert rules builder. Rule recipe catalog, tenant-scoped
    saved rules, persisted local backend state, dashboard builder panels,
    Postman/Newman coverage, and local verification.
23. Phase 20: consent, audit, and paid alert operations. Tenant-scoped consent
    center API, dashboard consent/audit panels, alert revenue operations, push
    notification center, denied sensitive collection disclosures,
    Postman/Newman coverage, and local verification.
24. Phase 21: device groups and policy assignments. Tenant-scoped managed
    rollout APIs, seeded primary group/assignment, assignment audit events,
    dashboard rollout panels, Postman/Newman coverage, and local verification.
25. Phase 22: data rights workflows. Tenant-scoped export manifests,
    non-destructive delete request queue, access audit events, dashboard
    export/delete panels, Postman/Newman coverage, and local verification.
26. Phase 23: customer operations dashboard. Tenant-level operations summary
    API, monetisation cockpit, escalation workbench, notification delivery
    board, upgrade proof pack, Postman/Newman coverage, and local verification.
27. Phase 24: dashboard demo lifecycle hardening. Targeted stale listener
    cleanup, isolated demo data paths, startup exit checks, lifecycle smoke,
    Newman coverage, and local verification.
28. Phase 25: monetization command center. Typed tenant monetisation summary,
    notification guarantee, paid feature proof, conversion actions, route-level
    proof, Postman/Newman coverage, and local verification.
29. Phase 26: notification route registry. Tenant-scoped email, push, and
    dashboard route records, provider/channel validation, route readiness
    dashboard panels, Postman/Newman coverage, and local verification.
30. Phase 27: revenue control room. Buyer-facing dashboard layer for package
    fit, paid proof, upgrade motion, renewal risk, commercial lever, anomaly
    assurance, email delivery, push delivery, report mail readiness,
    Postman/Newman coverage, and local verification.
31. Phase 28: agent telemetry ingest bridge. Typed metadata-only backend ingest
    endpoint, telemetry status endpoint, agent backend sync policy/client,
    dashboard live telemetry panels, regenerated schema, Postman/Newman
    coverage, real agent live smoke, and local verification.
32. Phase 29: monetisation launch deck. Buyer-ready first-screen dashboard for
    customer package readiness, anomaly push assurance, mail delivery
    assurance, weekly report proof, host risk command, S3 archive retention,
    notification revenue stream, Postman/Newman coverage, and local
    verification.
33. Phase 30: offline backend sync. Durable SQLite sync cursor/backlog replay,
    offline-tolerant agent sync behavior, idempotent backend telemetry ingest by
    stable event ID, Postman/Newman coverage, offline-to-online live smoke, and
    local verification.
34. Phase 31: buyer assurance sync health. Tenant sync-health API, dashboard
    Buyer Assurance Wall, Offline Replay Health panel, source-count replay
    proof, email/push monetisation checks, Postman/Newman coverage, live
    dashboard smoke, and local verification.
35. Phase 32: tenant activity feed. Tenant-level risk/delivery/telemetry feed
    API with host, kind, channel, status, query, and limit filters; dashboard
    Tenant Activity Feed and Filtered Command Feed panels; Postman/Newman
    coverage, live dashboard smoke, and local verification.
36. Phase 33: monetisation command views. Tenant-scoped saved activity views
    for high-risk anomalies, mail proof, push retry, and sync/archive evidence;
    dashboard command-view and notification-proof panels; typed validation,
    audit events, Postman/Newman coverage, live dashboard smoke, and local
    verification.
37. Phase 34: dashboard session auth. Session-scoped API-key unlock for the
    embedded localhost dashboard, protected API smoke/Newman coverage, security
    docs for session storage and header-based access, and local verification.
38. Phase 35: versioned policy schema checks. Typed policy-schema version
    registry, `schema --version` CLI support, checked-in schema drift tests,
    local schema verification script, docs, and local verification.
39. Phase 36: monetisation-grade revenue dashboard. First-screen Revenue
    Command Center, Monetisation Value Stack, Notification Proof Rail, Buyer
    Demo Checklist, live smoke/Newman coverage, docs, and local verification.
40. Phase 37: dashboard contract guard. Embedded dashboard DOM/JavaScript ID
    contract test, local script, smoke/Newman coverage, docs, and full local
    verification.
41. Phase 38: commercial control room. First-screen buyer-ready dashboard for
    host coverage, anomaly urgency, email proof, push proof, weekly report
    mail, delivery evidence, customer success actions, Postman/Newman coverage,
    docs, and full local verification.
42. Phase 39: autostart assurance. Stronger Windows Task Scheduler hidden
    startup/status JSON checks, service-manager dry-run proof, service trust
    smoke/Newman coverage, docs, and full local verification.
43. Phase 40: paid ops console. First-screen monetisation dashboard for anomaly
    response, push notifications, mail delivery, weekly reports, archive
    retention, tamper trust, escalation actions, live smoke/Newman coverage,
    docs, and full local verification.
44. Phase 41: backend alert inbox. Typed tenant alert inbox API that links
    policy/anomaly/tamper events to email, push, and dashboard delivery proof,
    dashboard panel, monetisation UI markers for paid ops, commercial control,
    revenue command, notification proof, mail delivery, push notification,
    archive retention, tamper trust, smoke/Newman coverage, docs, and full local
    verification.
45. Phase 42: command navigation. Sticky dashboard command navigation for paid
    ops, revenue, notifications, reports, archive, trust, and hosts with typed
    KPI backing data, dashboard contract guard, smoke/Newman coverage, docs, and
    full local verification.
46. Phase 43: buyer operations layout contract. First-screen buyer operations
    brief for anomaly alerting, mail delivery proof, push notification
    dispatch, weekly report delivery, archive retention, trust/audit, delivery
    command, package snapshot, and next commercial action, plus screenshot-free
    Playwright layout metrics across desktop, tablet, and mobile viewports,
    smoke/Newman coverage, docs, and full local verification.
47. Phase 44: provider-safe delivery drilldown. Tenant API and dashboard panels
    for email, push, and dashboard dry-run rehearsal, route score, channel
    readiness, route evidence, next action, audit trail, Postman/Newman
    coverage, screenshot-free layout guard, docs, and full local verification.
48. Phase 45: monetisation command center. Premium first-screen dashboard for
    anomaly inbox, push notification reach, mail delivery proof, weekly report
    mail/PDF readiness, fleet coverage, S3 archive retention, trust/audit,
    revenue package state, delivery proof, owner action queue, smoke/Newman
    coverage, screenshot-free layout guard, docs, and full local verification.
49. Phase 46: delivery remediation center. Typed tenant API and dashboard
    panels for provider-safe route recovery planning, owner/SLA state, dry-run
    push/mail remediation plans, live-send rejection, audit proof,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
50. Phase 47: premium notification command center. Typed tenant aggregate API
    and dashboard Notify Pro surface for anomaly/policy/tamper alert funnel,
    email delivery proof, push reach, route assurance, remediation SLA state,
    paid-tier labels, owner action SLAs, Postman/Newman coverage,
    screenshot-free layout guard, docs, and full local verification.
51. Phase 48: growth cockpit dashboard. First-screen monetisation surface for
    revenue readiness, anomaly notification operations, mail delivery, push
    reach, weekly report delivery, archive retention, trust/consent, owner
    action workflow, Postman/Newman coverage, screenshot-free layout guard,
    docs, and full local verification.
52. Phase 49: notification preference center. Typed tenant preference API and
    dashboard panel for immediate/digest/silent rules, quiet hours, escalation,
    study-safe suppression, channel coverage, route proof gaps, audit proof,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
53. Phase 50: business dashboard. Typed tenant product API and first-screen UI
    for customer health, anomaly notification inbox, mail delivery, push reach,
    route proof, preference readiness, archive/report value, paid package cards,
    customer owner actions, Postman/Newman coverage, screenshot-free layout
    guard, docs, and full local verification.
54. Phase 51: delivery timeline. Typed tenant delivery history API and
    dashboard Notification Evidence Timeline for host-level anomaly email,
    push notification, dashboard inbox, retry timing, route proof gaps, paid
    tier recommendation, Postman/Newman coverage, screenshot-free layout guard,
    docs, and full local verification.
55. Phase 52: role experiences. Typed tenant role-experiences API and
    dashboard Role Experience Center for parent, student, school admin, and
    business manager readiness, paid onboarding checklist, notification proof,
    archive/report promise, consent controls, Postman/Newman coverage,
    screenshot-free layout guard, docs, and full local verification.
56. Phase 53: executive console. Typed tenant executive-console API and
    first-screen Executive Notification Console for sellable readiness, anomaly
    alert stream, mail delivery proof, push reach, weekly report readiness,
    archive posture, role packaging, paid value tiles, owner actions,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
57. Phase 54: notification revenue cockpit. Typed tenant
    notification-revenue-cockpit API and first-screen monetisation UI for
    anomaly SLA, mail proof, push proof, dashboard delivery, weekly report
    readiness, escalation policy, scenario templates, channel proof matrix,
    upgrade action levers, Postman/Newman coverage, screenshot-free layout
    guard, docs, and full local verification.
58. Phase 55: provider simulation lab. Typed tenant provider-simulation-lab
    API and first-screen dashboard panel for metadata-only email, push, and
    dashboard dry-run proof, route SLA state, provider scenarios, action queue,
    audit proof, Postman/Newman coverage, screenshot-free layout guard, docs,
    and full local verification.
59. Phase 56: package billing readiness. Typed tenant package-billing-readiness
    API and dashboard panel for plan fit, billing setup metadata, feature
    gates, seat usage, retention/archive value, weekly reports, notification
    proof, provider simulation proof, trust/data-rights readiness, upgrade
    actions, Postman/Newman coverage, screenshot-free layout guard, docs, and
    full local verification.
60. Phase 57: customer control room. Typed tenant customer-control-room API and
    first-screen dashboard surface for anomaly command, mail delivery, push
    notification evidence, provider proof, report/archive readiness, package
    billing, customer health, owner monetisation actions, Postman/Newman
    coverage, screenshot-free layout guard, docs, and full local verification.
61. Phase 58: customer success packet. Typed tenant customer-success-packet API
    and dashboard surface for buyer/admin review packets, anomaly proof, mail
    delivery, push notification proof, report/archive readiness, package fit,
    provider rehearsal, privacy assurances, buyer objection answers, owner
    actions, Postman/Newman coverage, screenshot-free layout guard, docs, and
    full local verification.
62. Phase 59: push activation center. Typed tenant push-activation-center API
    and dashboard surface for monetisable push notification readiness, mail
    fallback, dashboard fallback, retry posture, route proof, preference and
    escalation coverage, anomaly push/mail scenarios, owner actions,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
63. Phase 60: portfolio center. Typed tenant portfolio-center API and dashboard
    surface for multi-host portfolio rows, health/risk/alert counts, anomaly
    alert notification rows, mail delivery proof, push notification proof,
    dashboard fallback proof, archive/sync posture, package readiness, owner
    actions, privacy guard, Postman/Newman coverage, screenshot-free layout
    guard, docs, and full local verification.
64. Phase 61: account portfolio index. Typed account portfolio API and dashboard
    surface for multi-tenant tenant rows, host coverage, alert pressure, mail
    delivery proof, push notification proof, dashboard fallback proof,
    archive/sync posture, package readiness, owner actions, tenant-scoped
    filtering, Postman/Newman coverage, screenshot-free layout guard, docs, and
    full local verification.
65. Phase 62: monetisation overview dashboard. First-screen buyer-grade UI that
    composes existing typed metadata into account coverage, host coverage,
    anomaly notification proof, mail delivery proof, push reach, weekly report
    readiness, archive posture, package fit, owner actions, and trust guardrails
    with Postman/Newman coverage, screenshot-free layout guard, docs, and full
    local verification.
66. Phase 63: tenant onboarding center. Typed tenant onboarding-center API and
    dashboard surface for endpoint install proof, reboot persistence/autostart,
    notification policy, anomaly mail and push proof, archive retention, role
    handoff, package readiness, privacy/data-rights guardrails, owner actions,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
67. Phase 64: customer settings center. Typed tenant customer-settings-center
    API and dashboard surface for plan settings, retention settings,
    notification policy, mail route proof, push route proof, dashboard fallback,
    archive, autostart, role dashboard settings, privacy/data-rights settings,
    owner actions, Postman/Newman coverage, screenshot-free layout guard, docs,
    and full local verification.
68. Phase 65: revenue operations center. Typed tenant
    revenue-operations-center API and dashboard surface for revenue readiness,
    anomaly queue, mail delivery proof, push notification proof, dashboard
    fallback, weekly report readiness, archive retention, onboarding, customer
    settings, provider proof, package fit, commercial levers, owner actions,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
69. Phase 66: deployment readiness center. Typed tenant
    deployment-readiness-center API and dashboard surface for Windows Task
    Scheduler, macOS launchd, Linux systemd, service manifests, reboot
    persistence, background startup, live boot, offline replay, archive backlog,
    owner actions, Postman/Newman coverage, screenshot-free layout guard,
    service/autostart verification, docs, and full local verification.
70. Phase 67: premium operations hub. Typed tenant premium-operations-hub API
    and polished dashboard first screen for anomaly inbox, mail delivery proof,
    push notification route state, dashboard fallback, weekly reports, archive
    retention, deployment readiness, package value, commercial levers, owner
    actions, Postman/Newman coverage, screenshot-free layout guard, docs, and
    full local verification.
71. Phase 68: browser activity viewer. Typed tenant browser-activity API and a
    dedicated `/browser-activity` page linked from the dashboard toolbar for
    Chrome, Edge, and Brave domain activity, host/category/study-safe filters,
    non-study YouTube review, notification proof, host and browser breakdowns,
    Postman/Newman coverage, screenshot-free layout guard, docs, and full local
    verification.
72. Phase 69: admin UI and Lambda frontend. Multipage dashboard tabs, dark
    theme toggle, green/red server connectivity indicators, Browser Activity
    Viewer parity, `devctl.py` for server/test/SAM/log operations, and a SAM
    Lambda Function URL admin frontend that reads S3 as source of truth,
    displays cache hit/miss percentages, supports a localhost `18080` source
    switch, saves stack outputs under `data/local/output/`, and is covered by
    smoke, Newman, contract, docs, and full local verification.
73. Phase 72: cloud S3 sample and cache proof. Lambda S3 summary parser reads
    real agent archive `Metadata` maps, a logged local script creates and
    uploads a metadata-only JSONL gzip browser sample to S3, smoke/Newman prove
    Chrome/Edge/Brave rows, study-safe inference, non-study YouTube, forbidden
    privacy markers, and cache hit/miss metrics against the deployed Function
    URL, with `devctl.py cloud` helpers and full local verification.
74. Phase 74: runtime doctor assurance. `python ./devctl.py doctor` writes
    metadata-only JSON/text reports under `data/local/output/` proving local
    backend health, dashboard controls, Browser Activity Viewer reachability,
    browser activity provenance, alert-delivery provenance, Lambda Function URL
    health, S3 summary rows, and cache hit/miss behavior, with smoke/Newman
    coverage, docs, and full local verification.
75. Phase 75: delivery assurance truth center. Typed tenant delivery-assurance
    API, dashboard Delivery Assurance Center, route/event truth labels for
    provider-confirmed, dry-run, dashboard-visible, demo-only, retrying, failed,
    disabled, and pending-provider states, runtime doctor assertions,
    Postman/Newman coverage, docs, and full local verification.
76. Phase 76: dashboard UI revamp. Product-grade visual refresh for the
    embedded dashboard and Browser Activity Viewer with clearer light/dark
    palettes, no pseudo-letter toolbar markers, wrapped chips, contained
    tables, improved card hierarchy, screenshot-free desktop/tablet/mobile
    layout guard, Postman/Newman coverage, docs, `devctl.py test phase76`, and
    full local verification.
77. Phase 78: notification provider setup center and UI quality guard. Typed
    tenant provider setup API and dashboard section that separates configured
    routes from provider-confirmed delivery, demo-only proof, retrying routes,
    checklist rows, buyer readiness, and owner actions. The shell is refreshed
    as TraceDeck Console with Browser Viewer, full Workspace Navigator labels,
    and a screenshot-free visual-quality contract for stale debug copy, tiny
    controls, dark-mode posture, server lights, and overflow, with Postman,
    Newman, docs, `devctl.py test phase78`, and full local verification.
78. Phase 80: Lambda Cloud Admin visual parity. Refresh the public Function URL
    frontend with the TraceDeck admin shell, symbolic brand mark, Workspace
    Source sidebar, full page labels, text-only theme control, product-grade
    light/dark cards/chips/tables, cache metric rendering, localhost source
    switch, screenshot-free Lambda visual-quality checks, SAM deployment, cloud
    Newman, runtime doctor, docs, and `devctl.py test phase80`.
79. Phase 81: dashboard navigator clarity. Replace terse embedded dashboard
    Workspace Navigator shortcuts with full product labels, split each command
    tile into `command-label` and `command-meta` rows, harden the
    screenshot-free visual-quality contract against stale shortcut labels, add
    smoke/Newman/verify/publish scripts, docs, and `devctl.py test phase81`.
80. Phase 82: modern admin UI polish. Apply a unified buyer-facing light/dark
    visual system across the local TraceDeck Console, Browser Activity
    drilldown, and Lambda Cloud Admin frontend; replace visible text-logo marks
    with symbolic marks; enlarge command tiles, status chips, panels, KPI
    cards, and tables; harden Go DOM, Playwright, Lambda visual, and Newman
    checks against stale `TD`, `Browser{}`, `Center{}`, bracket shortcut, and
    debug-abbreviation copy; add smoke/Newman/verify/publish scripts, docs, and
    `devctl.py test phase82`.
81. Phase 83: agent heartbeat telemetry. Emit one metadata-only
    `agent.health.heartbeat` event from `collector.agent.heartbeat` per agent
    collection cycle; expose agent version, collection mode/interval, archive
    due state, backend sync state, alert state, profile, and operating system
    as typed metadata; verify backend telemetry status counts, tenant
    sync-health/replay proof, dashboard/Lambda visual contracts, Newman, docs,
    and `devctl.py test phase83`.
82. Phase 85: strict Go quality gates. Add a reusable script and verifier for
    gofmt, `go test ./...`, `go test -race ./...`, `go vet ./...`,
    `golangci-lint run ./...`, `govulncheck ./...`, and `gosec ./...`; keep
    reports under `data/local/go-quality/`, add Newman runtime guard,
    Postman/docs/devctl hooks, and verify root hygiene.
83. Phase 86: premium UI and provenance recovery. Apply the final local
    TraceDeck Console and Browser Viewer visual layer, remove debug-looking
    labels, add smoke/Newman/visual/theme/layout gates, and prove default host
    policy/delivery APIs do not expose seeded demo evidence.
84. Phase 87: trust, quality, and UI hardening wrapper. Package the Phase 85
    quality gates, Phase 86 UI checks, demo-provenance API fix, weekly report
    email-proof correction, persistent live-server restart/provenance check,
    docs, Postman, publish script, and `devctl.py test phase87`.
85. Phase 88: cache and visual contract hardening. Package no-store cache
    headers, live provenance guards, screenshot-free layout/theme/visual
    checks, Phase 88 smoke/Newman, docs, and `devctl.py test phase88`.
86. Phase 89: activity-feed provenance. Hide seeded demo activity rows from
    default tenant feeds, keep explicit demo opt-in, and verify live
    provenance, Newman, docs, and root hygiene.
87. Phase 90: runtime doctor delivery truth. Separate default delivery proof
    from explicit demo delivery evidence in `python devctl.py doctor` and
    verify the runtime doctor cannot claim demo mail/push as real delivery.
88. Phase 91: persistent local backend task controls. Add hidden Windows
    scheduled-task backed backend start/stop/status controls and verify the
    local dashboard can survive managed shell lifetime boundaries.
89. Phase 92: backend task status resilience. Treat non-elevated Scheduler
    readback denial as a watch advisory when PID plus `/health` prove runtime
    health.
90. Phase 93: backend task status advisory. Add typed operator action metadata
    for verified, denied, missing, error, and unhealthy runtime task states.
91. Phase 94: deployment service advisory center. Surface deployment service
    advisories in the Deployment Readiness Center and dashboard.
92. Phase 95: repo-local Go cache. Route scripted `GOCACHE` and `GOTMPDIR`
    under `data/local` to avoid Windows/OneDrive cache ACL failures.
93. Phase 96: reusable post-merge verifier. Add a repeatable strict
    post-merge checklist for current phase gate, runtime, root hygiene, diff
    hygiene, and optional GitHub issue/PR checks.
94. Phase 97: runtime summary command. Add `python devctl.py summary` and
    JSON/text operator summary exports under `data/local/output`.
95. Phase 98: runtime status center. Expose the runtime summary through typed
    backend and dashboard surfaces.
96. Phase 99: verification evidence center. Expose scripted verification
    evidence as metadata-only API/dashboard proof.
97. Phase 100: operator assurance center. Compose runtime and verification
    evidence into a single operator assurance bundle.
98. Phase 101: post-merge verifier hardening. Avoid persistent backend output
    capture hangs in the reusable post-merge wrapper.
99. Phase 102: runtime PID reconciliation. Compare live backend PID proof with
    ready-file PID proof and report match, stale, absent, or unknown status.
100. Phase 103: ready PID proof refresh. Add a direct no-restart remediation
     command for stale ready proof.
101. Phase 104: action schema seal. Commit typed metadata-only action evidence
     scope fields after the Phase 103 schema gap.
102. Phase 105: promotion readiness center. Compose runtime status,
     verification evidence, operator assurance, git hygiene, ready PID proof,
     exports, and next actions into `python devctl.py promote`.
103. Phase 106: phase ledger. Add `docs/phase-ledger.md`,
     `python devctl.py ledger`, JSON/text ledger exports, and a direct
     remaining-phase count. Current planned numbered phases remaining: `0`.
104. Phase 107: contract completion audit. Add
     `docs/contract-completion-audit.md`, `python devctl.py audit`, JSON/text
     audit exports, and a focused verifier that separates planned phase count
     from actual end-to-end completion gaps.
105. Phase 108: browser extension skeleton. Add the Chrome, Edge, and Brave
     Browser Metadata Bridge under `browser-extension/`, local-only
     domain/category telemetry posting, privacy contract tests, live smoke,
     Newman coverage, docs, and `devctl.py test phase108`.
106. Phase 109: OpenTelemetry exporter and local collector stack. Add typed
     `observability.opentelemetry` policy config, metadata-only OTLP/HTTP JSON
     export with bounded attempt/drop metrics, fake receiver smoke coverage,
     Docker Compose/OpenTelemetry Collector config, Newman coverage, docs, and
     `devctl.py test phase109`.
