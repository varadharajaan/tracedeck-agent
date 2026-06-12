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
