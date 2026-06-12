# Dashboard

Phase 5 serves a small embedded dashboard shell from the Go backend at `/`.

The dashboard reads:

- `/health`
- `/api/v1/devices`
- `/api/v1/tenants`
- `/api/v1/plans`
- `/api/v1/roles`
- `/api/v1/policy-templates`
- `/api/v1/archive/status`
- `/api/v1/audit-events`

It shows device count, tenant count, plan count, policy template count, audit
event count, archive backlog, enrolled devices, starter policy templates,
tenant readiness rows, plans, and roles. API-provided text is escaped before
rendering.

This is the dashboard foundation. A full React dashboard, role-based UI,
policy builder, consent center, weekly reports, and compliance scoring remain
planned Phase 5/6 expansion work.
