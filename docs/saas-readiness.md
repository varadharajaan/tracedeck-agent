# SaaS Readiness

Phase 6 prepares TraceDeck for monetizable SaaS packaging without exposing a
remote unauthenticated backend.

## Current Slice

- Tenant readiness profile with tenant id, display name, selected plan,
  retention tier, primary profile, status, and device limit.
- Plan catalog for Free, Family Pro, School, Business, and Enterprise.
- Role catalog for parent, student, school admin, and business manager.
- Retention tier catalog for local-only, family cloud archive, school year
  archive, and business compliance packaging.
- Audit event list for administrative backend actions.
- Dashboard visibility for tenant, plan, role, and audit catalog signals.

## Boundaries

- Storage remains in-memory.
- Backend remains localhost-only.
- No billing provider is called.
- No SSO, remote auth, or tenant authorization is claimed yet.
- No endpoint collector behavior changes.
- No password, keylogging, cookie/token, private-message, camera, microphone,
  screenshot, or full-URL collection is added.

## Next SaaS Steps

- Durable tenant store.
- Authentication and tenant authorization.
- Billing provider integration.
- Policy assignment by tenant, role, and device group.
- Consent and audit center UI.
- Report export and retention enforcement.
