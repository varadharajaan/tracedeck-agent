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
