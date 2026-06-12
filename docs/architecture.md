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

Phase 2B turns the one-shot runner into a continuous loop. `run --once` remains
the single-snapshot path, while `run` without `--once` repeats collection using
`--collection-interval` until interrupted or `--max-cycles` is reached.

Phase 3 adds a browser history collector for Chromium-style history databases.
It copies browser history into a bounded local cache, reads recent rows through
SQLite, and persists domain/category events only. Raw URLs, query strings, and
page titles are used only transiently for classification and are not stored.

Phase 4 adds the first policy/anomaly evaluator layer. The app pipeline collects
events, stores them locally, optionally archives the batch, then evaluates alert
rules against the in-memory event batch. Evaluators are split by rule family so
new rules can be added without turning the runner into rule-specific code.
Current rule families cover blocked process names, blocked browser domains, and
non-study YouTube domain activity.

Phase 5 starts the backend/dashboard foundation with a separate Go command under
`backend/cmd/tracedeck-backend`. It uses the standard library HTTP server,
binds to localhost by default, exposes health/version/device/template/archive
routes, and serves an embedded dashboard shell. Backend storage is in-memory for
this foundation slice; later SaaS phases can replace the repository with durable
multi-tenant storage without changing handler contracts.
