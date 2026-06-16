# OpenTelemetry Exporter

TraceDeck Agent can export collected metadata events as OTLP/HTTP JSON log
records to a local OpenTelemetry Collector.

The exporter is optional and disabled by default in the sample policy:

```yaml
observability:
  opentelemetry:
    enabled: false
    protocol: otlp_http_json
    endpoint: http://127.0.0.1:4318/v1/logs
    batch_limit: 100
    request_timeout: 5s
    retry:
      max_attempts: 2
```

## Local Collector

The local collector stack lives under:

```text
deployments/otel/docker-compose.yaml
deployments/otel/otel-collector.yaml
```

Run it manually with:

```powershell
docker compose -f ./deployments/otel/docker-compose.yaml up
```

Validate the stack contract with:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-otel-exporter.ps1
```

## Verification

Phase 109 proof commands:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-otel-exporter.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase109.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase109.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase109.ps1
```

The smoke starts a repo-local fake OTLP receiver, runs the agent once with
OpenTelemetry enabled, then verifies exported event count, attempts, dropped
count, and privacy-safe captured OTLP logs.

## Privacy

OTLP export is metadata-only. The exporter filters sensitive metadata keys and
values before export. It does not export passwords, screenshots, raw URLs, page
titles, cookies, tokens, private content, provider secrets, alert bodies,
keylogging data, hidden collection bypasses, payment data, or raw provider
payloads.
