CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  event_type TEXT NOT NULL,
  source TEXT NOT NULL,
  observed_at TEXT NOT NULL,
  tenant_id TEXT NOT NULL,
  device_id TEXT NOT NULL,
  host_name TEXT NOT NULL,
  app_name TEXT NOT NULL,
  process_id INTEGER,
  path_hash TEXT,
  metadata_json TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_events_observed_at ON events(observed_at);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
