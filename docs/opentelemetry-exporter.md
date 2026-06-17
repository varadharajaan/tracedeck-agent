# OpenTelemetry Exporter

TraceDeck includes optional metadata-only OTLP export.

Local collector manifests:

```text
deployments/otel/docker-compose.yaml
deployments/otel/otel-collector.yaml
```

Verification:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-otel-exporter.ps1
```

OTLP export must stay bounded to TraceDeck metadata events.
