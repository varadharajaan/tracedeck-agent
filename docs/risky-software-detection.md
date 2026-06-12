# Risky Software Detection

Phase 13 adds privacy-safe risky software classification to process telemetry.

The agent uses the process name and transient executable path to classify:

- torrent clients
- VPN/proxy tools
- game launchers
- non-standard browsers
- installers launched from a downloads location

Only the process name, path hash, risk category, and reason are stored. Raw
executable paths are not persisted or archived by this classifier.

## Event Metadata

Risk metadata is attached to `process.observed` events:

```text
software_risk_category
software_risk_reason
```

Supported categories:

```text
torrent_client
vpn_proxy
game_launcher
unknown_browser
downloads_installer
```

## Alert Rule

Enable the rule in policy YAML:

```yaml
alert_rules:
  risky_software_detected:
    enabled: true
    severity: high
```

The alert evaluator emits one alert per risky app per cycle and includes the
software risk category and reason in metadata.

## Privacy Boundary

This phase does not inspect file contents, collect raw installer paths, read
browser storage, collect cookies or tokens, capture credentials, record
keystrokes, capture screenshots, use camera or microphone, or collect private
messages.
