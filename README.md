# TraceDeck Agent

TraceDeck Agent is a Go-first, privacy-aware endpoint activity, productivity,
and risk observability agent for Windows, macOS, and Linux laptops and managed
devices.

It tracks typed endpoint metadata such as application usage, browser
domain/category activity, software inventory changes, policy violations, S3
archive health, alert delivery health, and agent health using OpenTelemetry.

TraceDeck is not credential capture or covert surveillance. It does not collect
passwords, keystrokes, browser cookies, auth tokens, private messages, camera,
microphone, or hidden screen content. Browser monitoring is domain/category
based by default.

## Local Commands

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase0.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase7.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase8.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase15.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase17.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase35.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase44.ps1
go run ./agent/cmd/tracedeck-agent validate-config --config ./examples/policies/ai-btech-student.yaml
go run ./agent/cmd/tracedeck-agent schema --version v1alpha1 --out ./docs/schema/policy-v1alpha1.schema.json
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-policy-schema.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1
python ./devctl.py test theme
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-visual.ps1
python ./devctl.py cloud visual
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-autostart-assurance.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase5.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase6.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase9.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase11.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase12.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase13.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase14.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Action status -DryRun
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase16.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase17.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase18.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase19.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase20.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase22.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase23.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase24.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase34.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase36.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase37.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase38.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase39.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase40.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase41.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase42.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase43.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase44.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase44.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase45.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase45.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase45.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase46.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase47.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase48.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase48.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase48.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase51.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase51.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase51.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase52.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase52.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase52.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase53.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase53.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase53.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase54.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase54.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase54.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase55.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase55.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase55.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase56.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase56.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase56.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase57.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase57.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase57.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase58.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase58.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase58.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase49.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase50.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase50.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase50.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase83.ps1
python ./devctl.py test phase84
python ./devctl.py test phase83
```

All repeatable setup and verification work is kept under `scripts/`, and script
logs are written under `logs/local/`.

Phase 43 adds a buyer operations brief and screenshot-free dashboard layout
contract. The first-screen UI now makes anomaly alerting, mail proof, push
notification dispatch, weekly report delivery, archive retention, trust/audit,
delivery command, packaging snapshot, and next commercial action visible for a
monetisation demo. The Playwright check records layout metrics only under
`data/local/dashboard-layout/` and does not capture screenshots, video,
credentials, or page content.

Phase 44 adds provider-safe delivery drilldown. The backend exposes
`/api/v1/tenants/{tenantId}/delivery-drilldown` for current route proof and
dry-run rehearsals across email, push, and dashboard routes. The dashboard shows
route score, channel readiness, route evidence, and next actions without sending
live messages or storing provider secrets, alert bodies, endpoint payloads, or
sensitive content.

Phase 45 adds a monetisation command center as the first buyer-grade dashboard
surface. It rolls anomaly inbox urgency, push notification reach, mail delivery
proof, weekly report mail/PDF readiness, fleet coverage, S3 archive retention,
trust/audit posture, revenue package state, delivery proof, and owner actions
into one top panel backed by existing typed APIs. The phase adds presentation,
docs, smoke/Newman coverage, and layout guards only; it does not add password,
credential, raw URL, page title, private message, camera, microphone, or
covert screenshot collection.

Phase 66 adds a Deployment Readiness Center for reboot persistence and
cross-platform rollout proof. It exposes Windows Task Scheduler, macOS launchd,
Linux systemd, service manifest, live boot, background startup, offline replay,
archive backlog, and owner-action metadata through a typed API and dashboard
panel. Verification is under `scripts/verify/verify-phase66.ps1`,
`scripts/local/smoke-phase66.ps1`, and `scripts/local/newman-phase66.ps1`.

Phase 46 adds a delivery remediation center and typed
`/api/v1/tenants/{tenantId}/delivery-remediation` API. It turns mail/push route
problems into owned dry-run recovery plans with SLA state and audit proof, and
surfaces that proof in the monetisation command center. Live-send remediation
modes are rejected; provider secrets, alert bodies, screenshots, passwords,
tokens, cookies, raw URLs, and private content are not collected or stored.

Phase 47 adds a premium notification command center and typed
`/api/v1/tenants/{tenantId}/notification-command-center` API. It packages open
anomaly/policy/tamper alerts, email delivery proof, push reach, dashboard route
proof, remediation SLA state, paid-tier labels, and owner actions into a
buyer-facing dashboard surface. It remains metadata-only and does not collect
or store provider secrets, alert bodies, screenshots, passwords, tokens,
cookies, raw URLs, page titles, or private content.

Phase 48 adds a first-screen Growth Cockpit so the dashboard reads like a
monetisable endpoint productivity and risk observability product. It packages
revenue readiness, anomaly notification ops, mail delivery, push reach, weekly
report delivery, archive retention, trust/consent, and owner actions from
existing typed APIs without adding sensitive collectors.

Phase 49 adds a typed Notification Preference Center at
`/api/v1/tenants/{tenantId}/notification-preferences`. It models channel
policy, quiet hours, digest cadence, escalation, study-safe suppression, and
owner actions with metadata-only privacy boundaries, plus dashboard, smoke,
Newman, and verifier coverage.

Phase 50 adds a typed Business Dashboard at
`/api/v1/tenants/{tenantId}/business-dashboard`. It creates a monetisation-grade
first screen for customer health, anomaly notification, mail delivery, push
reach, route proof, archive/report value, paid package cards, and customer owner
actions. It aggregates existing metadata-only APIs and does not add password,
screenshot, raw URL, page title, alert body, provider secret, token, cookie, or
private-content collection.

Phase 53 adds a typed Executive Notification Console at
`/api/v1/tenants/{tenantId}/executive-console`. It promotes anomaly urgency,
mail delivery proof, push reach, weekly report readiness, archive posture, role
packaging, paid-plan value tiles, and owner actions into the dashboard first
screen so the product reads like a sellable endpoint productivity and risk
observability console. It aggregates metadata-only proof and does not add
password, screenshot, raw URL, page title, alert body, provider secret, token,
cookie, or private-content collection.

Phase 54 adds a typed Notification Revenue Cockpit at
`/api/v1/tenants/{tenantId}/notification-revenue-cockpit`. It gives the UI a
buyer-ready notification monetisation layer for anomaly SLA, mail proof, push
proof, dashboard delivery, weekly report readiness, escalation policy,
scenario templates, channel value, and upgrade action levers. It remains
metadata-only and does not add password, screenshot, raw URL, page title, alert
body, provider secret, token, cookie, or private-content collection.

Phase 55 adds a typed Provider Simulation Lab at
`/api/v1/tenants/{tenantId}/provider-simulation-lab`. It shows metadata-only
email, push, and dashboard dry-run proof, route SLA state, simulation
scenarios, provider action queue, audit proof, and privacy proof for paid
demos. It does not send live provider payloads or store SMTP passwords, push
endpoint payloads, alert bodies, screenshots, raw URLs, provider secrets,
tokens, cookies, or private content.

Phase 56 adds typed Package Billing Readiness at
`/api/v1/tenants/{tenantId}/package-billing-readiness`. It gives the dashboard
a buyer-facing package layer for plan fit, billing setup metadata, feature
gates, retention/archive value, weekly reports, notification proof, provider
simulation proof, trust/data-rights readiness, and upgrade actions. It is
metadata-only and does not collect payment card data, invoices, provider
secrets, passwords, screenshots, raw URLs, page titles, alert bodies, tokens,
cookies, private content, or endpoint payloads.

Phase 57 adds a typed Customer Control Room at
`/api/v1/tenants/{tenantId}/customer-control-room`. It becomes the first
dashboard surface for anomaly command, mail delivery, push notification
evidence, provider proof, report/archive readiness, package billing, customer
health, and owner monetisation actions. It is metadata-only and does not
collect passwords, screenshots, raw URLs, page titles, alert bodies, provider
secrets, endpoint payloads, private content, or payment card data.

Phase 58 adds a typed Customer Success Packet at
`/api/v1/tenants/{tenantId}/customer-success-packet`. It turns the same
metadata-only evidence into a buyer/admin review packet with anomaly proof,
mail delivery, push notification proof, report/archive readiness, package fit,
provider rehearsal, privacy assurances, objection answers, and owner actions.
It does not collect passwords, screenshots, raw URLs, page titles, alert
bodies, provider secrets, push endpoints, endpoint payloads, private content,
invoices, or payment card data.

Phase 59 adds a typed Push Activation Center at
`/api/v1/tenants/{tenantId}/push-activation-center`. It gives the dashboard a
monetisation-grade notification readiness panel for push delivered/retrying
proof, mail fallback, dashboard fallback, route proof, preference/escalation
coverage, provider-safe simulation, anomaly push/mail scenarios, owner
actions, and privacy guard. It is metadata-only and does not collect passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, private content, invoices, payment card data, or
raw provider payloads.

Phase 60 adds a typed Portfolio Center at
`/api/v1/tenants/{tenantId}/portfolio-center`. It gives parents, school admins,
and business managers a multi-host command view with portfolio score, alert
notification rows, mail delivery proof, push notification proof, dashboard
fallback, host rows, health/risk/alert counts, archive/sync posture, package
readiness, owner actions, and privacy guard. It is metadata-only and
does not collect passwords, screenshots, raw URLs, page titles, alert bodies,
provider secrets, push endpoints, endpoint payloads, private content,
invoices, payment card data, tokens, cookies, or raw provider payloads.

Phase 61 adds a typed Account Portfolio Index at
`/api/v1/account-portfolio-index`. It gives account owners and admins a
multi-tenant opening view with tenant rows, host coverage, anomaly pressure,
mail delivery proof, push notification proof, dashboard fallback, archive
posture, package readiness, owner actions, and privacy guard. Tenant-scoped
API sessions see only their allowed tenant. It is metadata-only and does not
collect passwords, screenshots, raw URLs, page titles, alert bodies, provider
secrets, push endpoints, endpoint payloads, private content, invoices, payment
card data, tokens, cookies, or raw provider payloads.

Phase 62 adds a Monetisation Overview as the first dashboard surface. It pulls
existing typed metadata into one buyer-grade opening view: account coverage,
host coverage, anomaly pressure, mail delivery, push reach, weekly report
readiness, archive posture, package fit, owner actions, and trust guardrails.
It does not add collectors; it remains metadata-only and avoids passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, private content, invoices, payment card data,
tokens, cookies, or raw provider payloads.

Phase 63 adds a typed Tenant Onboarding Center at
`/api/v1/tenants/{tenantId}/onboarding-center` and a dashboard panel for paid
activation. It shows setup checklist readiness, host reporting, reboot
persistence/autostart proof, anomaly notification policy, mail and push
delivery proof, archive retention, role dashboard handoff, package readiness,
privacy/data-rights guardrails, and owner actions. It remains metadata-only and
does not collect passwords, screenshots, raw URLs, page titles, alert bodies,
provider secrets, push endpoints, endpoint payloads, private content, invoices,
payment card data, tokens, cookies, or raw provider payloads.

Phase 64 adds a typed Customer Settings Center at
`/api/v1/tenants/{tenantId}/customer-settings-center` and a dashboard panel for
buyer/admin activation settings. It shows current and recommended plan,
retention, notification policy, mail route, push route, dashboard fallback,
archive, autostart, role dashboard, and privacy/data-rights settings with
owner actions. It remains metadata-only and does not collect passwords,
screenshots, raw URLs, page titles, alert bodies, provider secrets, push
endpoints, endpoint payloads, private content, invoices, payment card data,
tokens, cookies, or raw provider payloads.

Phase 65 adds a typed Revenue Operations Center at
`/api/v1/tenants/{tenantId}/revenue-operations-center` and a dashboard panel
that makes TraceDeck feel like a monetisable endpoint productivity and risk
observability product. It combines anomaly queue, mail delivery, push
notification, dashboard fallback, weekly report readiness, archive retention,
onboarding, customer settings, provider simulation, package fit, commercial
levers, and owner actions in one surface. It remains metadata-only and does
not collect passwords, screenshots, raw URLs, page titles, alert bodies,
provider secrets, push endpoints, endpoint payloads, private content,
invoices, payment card data, tokens, cookies, or raw provider payloads.

Phase 67 adds a typed Premium Operations Hub at
`/api/v1/tenants/{tenantId}/premium-operations-hub` and places it at the top of
the monetisation dashboard. It gives buyers and admins one polished first
screen for anomaly inbox, mail delivery proof, push notification route state,
dashboard fallback, weekly reports, archive retention, deployment readiness,
package value, commercial levers, and owner actions. It remains metadata-only
and does not collect passwords, screenshots, raw URLs, page titles, alert
bodies, provider secrets, push endpoints, endpoint payloads, private content,
invoices, payment card data, tokens, cookies, raw provider payloads,
keylogging, or hidden collection bypasses.

Phase 68 adds a typed Browser Activity Viewer at
`/api/v1/tenants/{tenantId}/browser-activity` plus a `/browser-activity` page
linked from the main dashboard toolbar. It gives buyers/admins Chrome, Edge,
and Brave domain activity by tenant and host, with category filters,
study-safe suppression, non-study YouTube review, notification proof, host and
browser breakdowns, Postman/Newman coverage, live smoke testing, and local
verification. It remains metadata-only and does not collect passwords,
screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint
payloads, provider secrets, push endpoints, alert bodies, keylogging, or hidden
collection bypasses.

Phase 69 adds the admin UI shell needed for monetisable daily use: a dark
theme toggle, explicit green/red server status light, dashboard page tabs to
avoid one long scroll, and the same controls on the Browser Activity Viewer.
It also adds a SAM-deployable `sam-app` frontend served by a public Lambda
Function URL without API Gateway. The Lambda frontend reads S3 archive objects
as source of truth, caches summaries in-memory to reduce S3 calls, displays
cache hit/miss percentages, and includes a configurable localhost `18080`
source switch for admin checks on the same machine. Use `python devctl.py
server status`, `python devctl.py server restart`, `python devctl.py test
phase69`, and `python devctl.py sam deploy`; stack outputs are saved under
`data/local/output/`.

Phase 74 adds a runtime doctor command for the operator question "is everything
actually reachable now?" Use `python devctl.py doctor` to check the local
backend, dashboard runtime controls, Browser Activity Viewer, typed browser
activity rows, alert-delivery provenance, Lambda Function URL health, S3
summary rows, and cache hit/miss state. It writes
`data/local/output/runtime-doctor.json` and
`data/local/output/runtime-doctor.txt`. Use `--skip-cloud` for isolated local
testing. The report is metadata-only and does not collect passwords,
screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint
payloads, keylogging, provider secrets, push endpoints, alert bodies, payment
data, or raw provider payloads.

Phase 75 adds a Delivery Assurance Center so notification proof cannot be
misread. `GET /api/v1/tenants/{tenantId}/delivery-assurance` separates
`provider_confirmed`, `dry_run_rehearsed`, `dashboard_visible`, `demo_only`,
`retrying`, `failed`, `route_disabled`, and `pending_provider` states. Seeded
dashboard data is labelled `demo_only`; retrying web push is labelled as not
screen-visible; buyer-ready proof requires provider-confirmed email, provider-
confirmed push, and dashboard fallback. Use `python devctl.py test phase75` to
run the full local verifier.

Phase 76 revamps the embedded dashboard and Browser Activity Viewer UI for a
more monetisable product surface. It removes pseudo-letter toolbar markers,
tightens light and dark theme palettes, improves card hierarchy, wraps chips
without clipped labels, contains wide tables, and hardens the screenshot-free
layout contract against horizontal overflow on desktop, tablet, and mobile.
Use `python devctl.py test phase76` or
`powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase76.ps1`
to rerun the full local verifier.

Phase 78 adds a Notification Provider Setup Center and refreshes the dashboard
shell as `TraceDeck Console` with a `Browser Viewer` drilldown. The provider
setup view separates configured email/push/dashboard routes from provider-
confirmed delivery, demo-only proof, retrying routes, buyer readiness, setup
checklist, owner actions, and metadata-only boundaries. The UI verifier now
also runs `scripts/local/test-dashboard-visual-quality.ps1` to block stale
debug labels, tiny chips/buttons, dark-mode regressions, and horizontal
overflow. Use `python devctl.py test phase78` for the full local gate.

Phase 80 brings the same product-quality guard to the public Lambda Cloud
Admin frontend. The Function URL page now uses a TraceDeck admin shell with a
symbolic brand mark, Workspace Source sidebar, full page labels, text-only
`Theme: Light/Dark` control, cache metrics, S3/localhost source switching, and
polished light/dark tables without pseudo-letter controls. Use
`python devctl.py cloud visual` for the local Lambda visual contract and
`python devctl.py test phase80` to compile, smoke, deploy SAM, run cloud
Newman, run runtime doctor, and re-check root hygiene.

Phase 81 tightens the embedded TraceDeck Console Workspace Navigator. The
navigator now uses full product labels such as `Deployment Readiness`,
`Customer Control Room`, `Provider Setup`, `Paid Operations`, and `Delivery
Assurance`, with separate live metadata rows, instead of terse shortcut text.
The visual-quality checker and Phase 81 smoke/Newman gates reject the old
shortcut markup. Use `python devctl.py test phase81` for the local verifier.

Phase 82 applies a modern admin UI polish pass across the local TraceDeck
Console, Browser Activity drilldown, and Lambda Cloud Admin frontend. The
visible text-logo mark is replaced with a symbolic product mark, light/dark
palettes are unified, command tiles and badges are larger and less debug-like,
and the static DOM, Playwright, Lambda visual, and Newman gates reject stale
`TD`, `Browser{}`, `Center{}`, bracket shortcut, and terse debug copy. Use
`python devctl.py test phase82` for the full local verifier.

Phase 83 adds an agent heartbeat telemetry event emitted on every collection
cycle. The event type is `agent.health.heartbeat` from
`collector.agent.heartbeat`; it records typed readiness metadata such as agent
version, collection mode, collection interval, archive enabled/due state,
backend sync enabled state, alerts enabled state, profile, and operating
system. It is metadata-only and does not collect passwords, screenshots, raw
URLs, page titles, cookies, tokens, private content, endpoint payloads,
provider secrets, alert bodies, keylogging, or hidden collection bypasses. Use
`python devctl.py test phase83` for the full local verifier.

Phase 85 adds the strict Go quality gate required by the engineering contract.
Use `python devctl.py test quality` to run gofmt, `go test ./...`,
`go test -race ./...`, `go vet ./...`, `golangci-lint run ./...`,
`govulncheck ./...`, and `gosec ./...`; reports are written under
`data/local/go-quality/` and logs under `logs/local/test/`. Use
`python devctl.py test phase85` to run the quality gate, Phase 85 Newman
runtime guard, runtime doctor, and root-clean check.
All PowerShell verification scripts set `GOCACHE` and `GOTMPDIR` under
`data/local/` so Go build/test artifacts stay inside the repo-local generated
artifact area instead of Windows user profile cache directories.

Phase 87 packages the current trust hardening bundle. Use
`python devctl.py test phase87` to rerun the Phase 86 premium UI/provenance
gate, the strict Go quality gate, a persistent local backend restart, the live
server provenance guard, and root-clean. Default host APIs suppress
`demo_seed` rows such as the seeded VLC/media-playback sample unless
`include_demo=true` is explicit, and weekly report email readiness requires
non-demo delivered email proof. The tenant activity feed uses the same typed
`include_demo` boundary, so host-scoped dashboard feeds cannot show the seeded
VLC row as live evidence. Dashboard HTML and JSON API responses are served with
no-store cache headers so stale UI bundles do not preserve old demo rows.
The screenshot-free dashboard visual/theme/layout checks wait for hydrated
product controls instead of browser network-idle to avoid false timeouts from
background API activity.

Phase 88 packages the cache/header and visual-contract follow-up. Use
`python devctl.py test phase88` to rerun Phase 87, Phase 88 smoke, Phase 88
Newman, and root-clean. This phase adds the shared PowerShell HTTP constants
helper required by live provenance checks and keeps the screenshot-free
Playwright contracts on DOM/hydrated-control readiness.

Phase 89 hardens tenant activity feed provenance. Use
`python devctl.py test phase89` to run the focused activity-feed guard, Go
quality checks, isolated live smoke, Newman, compatibility smoke checks, and
root-clean. Default tenant and host-scoped activity feed responses now suppress
`source_kind=demo_seed`, so seeded rows such as `VLC media player` cannot appear
as live host evidence unless `include_demo=true` is explicit.

Phase 90 hardens `devctl.py doctor` so it cannot turn demo notification rows
into false delivery proof. Use `python devctl.py test phase90` to run the
runtime-doctor provenance smoke, Phase 90 Newman collection, prior provenance
regression gate, and root-clean. The doctor report now records default alert
delivery count/source kinds separately from explicit `include_demo=true` demo
proof and marks buyer readiness false unless provider-confirmed mail/push proof
exists.

Phase 91 adds persistent local backend task controls for the admin console.
Use `python devctl.py server task-start`, `task-status`, `task-restart`, and
`task-stop` when you want `http://127.0.0.1:18080` to survive the devctl command
session. `task-status` reports Scheduler readback separately from runtime
health, so Windows `Access denied` becomes `task_state=inaccessible` while
`runtime_ok=true` still proves the backend pid and `/health` endpoint are alive.
Use `python devctl.py test phase91` to run the isolated scheduled-task smoke,
Phase 91 Newman collection, and root-clean check.

Phase 92 hardens that task status contract for normal, non-elevated Windows
shells. Use `python devctl.py test phase92` to verify the status classifier,
isolated scheduled-task smoke, Newman provenance checks, and root-clean. The
accepted healthy states are `task_present=true` with `runtime_ok=true`, or
`task_state=inaccessible` with `runtime_ok=true` when Windows denies Scheduler
metadata but the backend pid and `/health` endpoint are both alive. If the
current shell cannot create an isolated scheduled task, the Phase 92 smoke falls
back to the default `18080` task-status proof and still runs live provenance plus
runtime doctor checks.

Phase 93 adds an actionable advisory object to backend task status JSON and
prints it from `python devctl.py server task-status`. The advisory gives a
typed severity/code, whether local dashboard work can continue, whether elevated
Scheduler readback is recommended, and the next command to run. Use
`python devctl.py test phase93` to verify the advisory helper, live status
output, Newman provenance contract, and root-clean.

Phase 94 surfaces deployment service advisories in the Deployment Readiness
Center API and dashboard. Use `python devctl.py test phase94` to verify typed
metadata-only advisories for live boot, native autostart, background start,
offline replay, archive backlog, and ready states.

Phase 95 hardens scripted Go verification on Windows by routing `GOCACHE` and
`GOTMPDIR` to `data/local/go-build-cache/` and `data/local/go-tmp/` from the
shared PowerShell script bootstrap. Use `python devctl.py test phase95` to
rerun the Phase 94 gate and prove Go build/test artifacts stay repo-local.

Phase 96 adds a reusable post-merge verifier. Use
`python devctl.py test postmerge` or `python devctl.py test phase96` to run the
current phase gate, backend task-status, runtime doctor, live provenance,
root-clean, and git diff hygiene from one command.

Phase 97 adds a runtime summary command. Use `python devctl.py summary` to
write `data/local/output/runtime-summary.json` and
`data/local/output/runtime-summary.txt` with backend health, Scheduler
readback, task advisory, runtime doctor, frontend URL, git diff hygiene, and
operator next actions.

Phase 98 surfaces that summary through `GET /api/v1/runtime-status-center` and
the dashboard Rollout page as Runtime Status Center. Use
`python devctl.py test phase98` to live-boot an isolated backend, generate the
runtime summary, verify the typed API, run Newman, and run screenshot-free
layout/theme/visual checks.

Phase 99 adds a Verification Evidence Center. Use
`python devctl.py evidence` to write
`data/local/output/verification-evidence.json`, then
`python devctl.py test phase99` to live-boot an isolated backend, verify
`GET /api/v1/verification-evidence-center`, run Newman, and run the
screenshot-free layout/theme/visual contracts. The evidence center is
metadata-only: scripted gate labels, statuses, commands, timestamps, log paths,
report paths, git labels, and operator actions only.

Phase 100 adds an Operator Assurance Center. Use
`python devctl.py assurance` to write
`data/local/output/operator-assurance.json` and
`data/local/output/operator-assurance.txt`, then
`python devctl.py test phase100` to verify
`GET /api/v1/operator-assurance-center`, the dashboard assurance panel, Newman,
and screenshot-free layout/theme/visual contracts. The assurance center
combines runtime status and verification evidence into compact metadata-only
cards, including a clear Scheduler-denied-but-runtime-healthy explanation. The
Phase 100 verifier refreshes the persistent local backend through
`scripts/local/wait-backend-health.ps1`, keeping the `18080` proof bounded and
logged.

Phase 101 hardens `scripts/verify/verify-postmerge.ps1` so the current phase
verification runs outside the output-capturing logger. This keeps strict
post-merge checks from hanging when a phase intentionally refreshes the
persistent local backend.

Phase 102 reconciles live backend PID proof with the scheduled ready-file PID.
`python devctl.py summary` now writes `ready_pid`, `ready_pid_matches_live`,
and `ready_pid_status` under `backend`; `stale` is shown as a watch item while
healthy live `pid_and_health` proof remains usable. Use
`python devctl.py test phase102`, `smoke102`, or `newman102` to verify the
stale-ready-PID Runtime Status and Operator Assurance contracts.

Phase 103 adds a direct remediation for that stale proof. Use
`python devctl.py server task-refresh-ready` to rewrite
`data/local/backend/backend-task-ready.json` from the healthy live PID and
`/health` proof without restarting the backend. Runtime Status Center and
Operator Assurance now point `refresh-ready-pid-proof` actions at that command.
Use `python devctl.py test phase103`, `smoke103`, or `newman103` to verify the
refresh command and metadata-only API actions.

Phase 104 seals that metadata-only action contract in the typed Go response
models. Use `python devctl.py test phase104` or `verify104` to rerun the API,
Newman, gofmt, and root-clean checks that prove action rows expose
`evidence_scope=metadata_only`.

Phase 105 adds a Promotion Readiness Center. Use `python devctl.py promote` to
write `data/local/output/promotion-readiness.json` and
`data/local/output/promotion-readiness.txt`, then
`python devctl.py test phase105` to verify
`GET /api/v1/promotion-readiness-center`, the dashboard promotion panel,
Newman, runtime summary, operator assurance, and root-clean. The promotion
bundle composes runtime status, verification evidence, operator assurance, git
hygiene, ready PID reconciliation, local export paths, and next actions as
metadata only.

Phase 106 adds a first-class phase ledger. Use `python devctl.py ledger` to
write `data/local/output/phase-ledger.json` and
`data/local/output/phase-ledger.txt`; the tracked human-readable ledger lives
at `docs/phase-ledger.md`. The current ledger answer is that `0` currently
defined numbered phases remain. Future work must be promoted into the planned
phase table before it becomes counted remaining phase work.

Phase 107 adds a metadata-only contract completion audit. Use
`python devctl.py audit` to write
`data/local/output/contract-completion-audit.json` and
`data/local/output/contract-completion-audit.txt`; the human-readable audit
contract lives at `docs/contract-completion-audit.md`. This audit intentionally
does not claim TraceDeck is end-to-end complete: it lists implemented,
partial, and missing deliverables so the phase ledger answer and the product
completion state stay separate and inspectable.

Phase 108 adds the Chrome, Edge, and Brave Browser Metadata Bridge skeleton.
The extension lives under `browser-extension/`, posts domain/category-only
`browser.domain.observed` telemetry to the existing localhost
`/api/v1/devices/{device_id}/telemetry-events` route, and is covered by
`python devctl.py test phase108`, `scripts/local/smoke-phase108.ps1`, and
`scripts/local/newman-phase108.ps1`. It does not store or transmit raw URLs,
page titles, cookies, tokens, passwords, screenshots, page content, provider
secrets, alert bodies, payment data, or raw provider payloads.
