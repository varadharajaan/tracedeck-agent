# Offline Backend Sync

Phase 30 makes backend telemetry sync durable across offline periods.

The local SQLite database remains the first durability boundary. Every collected
metadata event is written locally before backend sync runs. When
`backend_sync.enabled` is true, the agent now reads unsynced rows from SQLite in
ascending local event ID order and posts a bounded batch to the backend.

Cursor behavior:

- cursor table: `backend_sync_state`
- cursor name: `backend_telemetry`
- event ID sent to the backend: `local-event-{sqlite_row_id}`
- batch size: `backend_sync.batch_limit`
- cursor advances only after the backend accepts the payload
- backend network errors are logged as warnings and leave the local cursor
  unchanged
- the collection cycle still succeeds after local storage when the backend is
  temporarily offline
- the next online cycle replays the unsynced backlog before newer rows move
  past the cursor

Backend ingest is idempotent for non-empty event IDs. Replaying the same
`local-event-*` ID is acknowledged so the agent can advance its cursor, but the
backend does not store a duplicate telemetry row.

Privacy boundary:

- no passwords or credentials
- no screenshots
- no keylogs
- no cookies, tokens, browser storage, or provider secrets
- no private messages
- no raw URLs or page titles
- no raw file contents

Verification:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase30.ps1
```

The Phase 30 smoke points the agent at an offline localhost port, confirms the
run succeeds with a local telemetry backlog, starts the backend, reruns the
agent with the same SQLite data directory, and verifies the backend receives
the replayed metadata events.
