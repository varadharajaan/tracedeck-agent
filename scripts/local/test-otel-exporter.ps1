param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-otel-exporter" -LogRoot "logs/local/tests" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Run OpenTelemetry exporter Go tests" -Command {
        go test ./agent/internal/exporter ./agent/internal/app ./agent/internal/config ./agent/internal/schema
    }

    $composePath = Join-Path $script:TraceDeckRepoRoot "deployments/otel/docker-compose.yaml"
    $collectorPath = Join-Path $script:TraceDeckRepoRoot "deployments/otel/otel-collector.yaml"
    foreach ($path in @($composePath, $collectorPath)) {
        if (-not (Test-Path $path)) {
            throw "Expected OpenTelemetry deployment file: $path"
        }
    }
    $compose = Get-Content -Raw -Path $composePath
    $collector = Get-Content -Raw -Path $collectorPath
    foreach ($expected in @("otel/opentelemetry-collector-contrib", "4318:4318", "otel-collector.yaml")) {
        if ($compose -notmatch [regex]::Escape($expected)) {
            throw "Docker Compose file missing expected content: $expected"
        }
    }
    foreach ($expected in @("receivers:", "otlp:", "endpoint: 0.0.0.0:4318", "pipelines:", "logs:", "metrics:")) {
        if ($collector -notmatch [regex]::Escape($expected)) {
            throw "Collector config missing expected content: $expected"
        }
    }
    foreach ($forbidden in @("AWS_ACCESS_KEY", "AWS_SECRET", "password", "token:", "screenshot")) {
        if ($compose -match $forbidden -or $collector -match $forbidden) {
            throw "OpenTelemetry deployment config contains forbidden marker: $forbidden"
        }
    }

    $docker = Get-Command docker -ErrorAction SilentlyContinue
    if ($docker) {
        $dockerConfig = Join-Path $script:TraceDeckRepoRoot "data/local/docker-config"
        New-Item -ItemType Directory -Force -Path $dockerConfig | Out-Null
        $env:DOCKER_CONFIG = $dockerConfig
        Invoke-TraceDeckLoggedCommand -Label "Docker Compose OpenTelemetry config render" -Command {
            docker compose -f ./deployments/otel/docker-compose.yaml config
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "Docker CLI not found; static OpenTelemetry stack contract checks passed."
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
