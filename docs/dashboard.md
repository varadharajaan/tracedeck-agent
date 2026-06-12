# Dashboard

Phase 5 serves a small embedded dashboard shell from the Go backend at `/`.

The dashboard reads:

- `/health`
- `/api/v1/devices`
- `/api/v1/policy-templates`
- `/api/v1/archive/status`

It shows device count, policy template count, archive backlog, enrolled devices,
and starter policy templates. API-provided text is escaped before rendering.

This is the dashboard foundation. A full React dashboard, role-based UI,
policy builder, consent center, weekly reports, and compliance scoring remain
planned Phase 5/6 expansion work.
