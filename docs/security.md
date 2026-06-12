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
