# Policy Assignment

Phase 21 adds tenant-scoped device groups and policy assignments. This moves
TraceDeck toward managed family, school, and small-business rollouts without
adding new endpoint collectors.

## APIs

```text
GET  /api/v1/tenants/{tenantId}/device-groups
POST /api/v1/tenants/{tenantId}/device-groups
GET  /api/v1/tenants/{tenantId}/policy-assignments
POST /api/v1/tenants/{tenantId}/policy-assignments
```

New tenants are seeded with:

- a primary study device group
- a primary policy assignment using the tenant primary profile

Custom device groups include name, description, profile, device IDs, and policy
template ID. Policy assignments include name, target type, target ID, policy
template ID, alert rule IDs, mode, status, and timestamps.

Allowed assignment targets:

- `tenant`
- `device_group`
- `device`

Allowed assignment modes:

- `audit`
- `active`

## Audit

Creating a group writes `device_group.created`. Creating an assignment writes
`policy_assignment.created`. Both events appear in tenant audit APIs and the
dashboard audit center.

## Dashboard

The embedded dashboard now includes:

- Device Groups
- Policy Assignments

These panels show cohort size, profile, policy template, assignment target,
rollout mode, status, and alert-rule coverage.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase21.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase21.ps1
```
