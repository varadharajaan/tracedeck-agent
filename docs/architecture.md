# Architecture

TraceDeck uses ports and adapters around a Go endpoint agent. Domain packages
hold policy, event, scoring, alert, anomaly, and categorization rules. Platform,
storage, archive, notifier, and telemetry packages implement interfaces outside
the domain boundary.

Windows is the first implementation target. macOS and Linux support is added
through platform adapters with build tags, keeping the domain and pipeline
portable.
