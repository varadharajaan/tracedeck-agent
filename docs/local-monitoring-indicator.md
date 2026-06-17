# Local Monitoring Indicator

The local monitoring indicator gives an operator-readable status page for what
TraceDeck is doing on the machine.

Generate it with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-local-monitoring-indicator.ps1
```

Outputs are written under:

```text
data/local/output/
```

The indicator is metadata-only and should not expose passwords, screenshots,
raw URLs, page titles, cookies, tokens, or private content.
