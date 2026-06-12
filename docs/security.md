# Security

TraceDeck binds local browser APIs to localhost, validates policy before
collectors start, rejects forbidden collection settings, and keeps secrets out
of repo files and logs.

S3 objects must be encrypted at rest. Network export must use TLS.

The Phase 5 backend also binds to localhost by default and rejects non-local
bind addresses. It is a foundation API without remote authentication, so it
must not be exposed on `0.0.0.0` or a LAN interface.
