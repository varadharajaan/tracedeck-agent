# Security

TraceDeck binds local browser APIs to localhost, validates policy before
collectors start, rejects forbidden collection settings, and keeps secrets out
of repo files and logs.

S3 objects must be encrypted at rest. Network export must use TLS.

The Phase 5/6 backend also binds to localhost by default and rejects non-local
bind addresses. It is a foundation API without remote authentication, so it
must not be exposed on `0.0.0.0` or a LAN interface.

Phase 6 tenant, plan, role, retention, and audit endpoints are readiness APIs.
They do not grant remote access, do not implement billing side effects, and do
not change endpoint collection behavior. Remote multi-tenant use requires an
explicit authentication and authorization design first.

Phase 7 service templates do not install or start services automatically. macOS
foreground app support is marked as requiring Accessibility permission, and
Linux foreground support is marked partial because X11 and Wayland differ. No
new collector is enabled without typed policy and platform support work.

Phase 8 Windows scheduled-task registration is explicit and may request UAC
approval. The task avoids console-window flicker by launching the agent
executable in the background, but it is not a covert monitoring mechanism.
TraceDeck remains transparent and consent-based.

Phase 17 provider-backed email keeps provider secrets out of policy files.
SMTP credentials are read from `TRACEDECK_SMTP_*` environment variables, and
SES credentials come from the AWS SDK default credential chain. Alert payloads
must not include SMTP passwords, AWS credentials, raw browser URLs, cookies,
tokens, passwords, keystrokes, private messages, camera, microphone, or hidden
screen content.

Phase 20 consent and alert-operations panels expose trust evidence from existing
typed backend data: alert delivery rows, audit events, collection disclosures,
and tenant metadata. They do not add password, screenshot, private-message,
camera, microphone, cookie, token, keylog, raw URL, or page-title collection.

Phase 21 device groups and policy assignments are administrative metadata for
managed rollout. They do not change collector permissions, do not grant remote
access, and do not add sensitive data collection. Hosted multi-tenant rollout
still requires explicit authentication and authorization work.

Phase 22 delete requests are non-destructive workflow records. They do not
delete tenant data automatically; hosted deletion execution requires stronger
authorization, approval, and retention enforcement.

Phase 34 dashboard API-key unlock stores the local key only in browser
`sessionStorage` and attaches it as `X-TraceDeck-API-Key` for API calls. The key
is not embedded in the served dashboard HTML, not written to backend state, and
not logged by TraceDeck scripts. This is still a localhost development/admin
mechanism, not hosted SSO or internet-exposed authentication.
