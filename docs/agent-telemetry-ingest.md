# Agent Telemetry Ingest

Phase 28 adds a privacy-safe bridge from the local Go agent to the local Go
backend.

The agent can now post the current collection cycle to:

```text
POST /api/v1/devices/{deviceId}/telemetry-events
```

The backend exposes current ingest proof at:

```text
GET /api/v1/devices/{deviceId}/telemetry-status
```

The ingest contract is metadata-only. It accepts typed event metadata already
produced by the agent:

- event type
- source collector
- observed timestamp
- tenant, device, host, profile, and OS metadata
- application name
- process id
- path hash
- metadata key/value map

It does not accept or require passwords, credentials, screenshots, tokens,
cookies, keylogs, private messages, raw URLs, page titles, provider secrets, or
raw file contents.

Agent policy example:

```yaml
backend_sync:
  enabled: true
  base_url: http://127.0.0.1:18080
  batch_limit: 100
  request_timeout: 10s
```

When enabled, the agent syncs after the local SQLite write succeeds. The local
store remains the first durability boundary.

Phase 30 adds durable backlog replay. The agent reads unsynced SQLite rows by
local event ID, posts up to `backend_sync.batch_limit`, and advances the
`backend_telemetry` cursor only after the backend accepts the payload. If the
backend is offline, the cycle still succeeds after local storage and the next
online cycle replays the backlog.

Stable replay IDs use the `local-event-{sqlite_row_id}` shape. The backend
acknowledges duplicate stable IDs without storing duplicate telemetry rows, so
cursor retry remains idempotent.

Dashboard support:

- Live Agent Telemetry
- Telemetry Privacy Boundary
- stored event count
- process and health source counts
- last ingest/observed timestamps
- recent metadata event proof

Verification:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase28.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase30.ps1
```

The verifier regenerates the policy schema, runs backend tests, runs agent
tests, starts a live backend, runs the real agent once with backend sync
enabled, checks backend telemetry status, runs Newman, cross-builds the agent
and backend, and checks root artifact hygiene.
