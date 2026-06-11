# Architecture

TraceDeck uses ports and adapters around a Go endpoint agent. Domain packages
hold policy, event, scoring, alert, anomaly, and categorization rules. Platform,
storage, archive, notifier, and telemetry packages implement interfaces outside
the domain boundary.

Windows is the first implementation target. macOS and Linux support is added
through platform adapters with build tags, keeping the domain and pipeline
portable.

The Phase 1 local slice runs one process snapshot, hashes executable paths,
stores process metadata events in SQLite, and writes structured JSON logs with
rotation under `logs/local/agent/`.
SQLite schema changes are applied from ordered migration files under
`agent/internal/storage/sqlite/migrations/`.

Phase 1B introduces platform adapters behind Go build tags for Windows, macOS,
Linux, and fallback operating systems. The app and collectors depend on the
adapter contract instead of calling host OS APIs directly.
