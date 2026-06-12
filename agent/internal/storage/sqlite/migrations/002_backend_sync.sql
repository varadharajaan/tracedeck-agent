CREATE TABLE IF NOT EXISTS backend_sync_state (
  sync_name TEXT PRIMARY KEY,
  last_event_id INTEGER NOT NULL DEFAULT 0,
  updated_at TEXT NOT NULL
);
