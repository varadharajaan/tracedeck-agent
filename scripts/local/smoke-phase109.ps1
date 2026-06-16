param(
    [string]$Addr = "127.0.0.1:18431"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase109" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase109/$timestamp"
$policyPath = "$smokeRoot/policy.yaml"
$agentDataDir = "$smokeRoot/agent-data"
$agentOutboxDir = "$smokeRoot/outbox"
$receiverReady = "$smokeRoot/fake-otlp-ready.json"
$receiverCapture = "$smokeRoot/fake-otlp-capture.json"
$receiverStdout = "logs/local/otel/fake-otlp-$timestamp.out.log"
$receiverStderr = "logs/local/otel/fake-otlp-$timestamp.err.log"
$script:ReceiverProcess = $null

function Start-TraceDeckFakeOTLP {
    $stdoutFullPath = Join-Path $script:TraceDeckRepoRoot $receiverStdout
    $stderrFullPath = Join-Path $script:TraceDeckRepoRoot $receiverStderr
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $stdoutFullPath), (Split-Path -Parent $stderrFullPath) | Out-Null

    Write-TraceDeckLog -Level "INFO" -Message "Starting fake OTLP receiver addr=$Addr"
    $script:ReceiverProcess = Start-Process -FilePath "go" -ArgumentList @(
        "run",
        "./scripts/tools/fake-otlp",
        "--addr", $Addr,
        "--ready", $receiverReady,
        "--output", $receiverCapture
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutFullPath -RedirectStandardError $stderrFullPath -PassThru

    $readyFullPath = Join-Path $script:TraceDeckRepoRoot $receiverReady
    $deadline = (Get-Date).AddSeconds(45)
    while ((Get-Date) -lt $deadline) {
        if (Test-Path $readyFullPath) {
            try {
                $health = Invoke-RestMethod -Method "GET" -Uri "http://$Addr/health"
                if ($health.status -eq "ok") {
                    Write-TraceDeckLog -Level "INFO" -Message "Fake OTLP receiver ready pid=$($script:ReceiverProcess.Id)"
                    return
                }
            }
            catch {
                Start-Sleep -Milliseconds 300
            }
        }
        elseif ($script:ReceiverProcess.HasExited -and $script:ReceiverProcess.ExitCode -ne 0) {
            throw "Fake OTLP receiver failed with exit code $($script:ReceiverProcess.ExitCode). stderr=$receiverStderr"
        }
        Start-Sleep -Milliseconds 300
    }
    throw "Fake OTLP receiver did not become ready at http://$Addr"
}

function New-TraceDeckOTelPolicy {
    $fullPath = Join-Path $script:TraceDeckRepoRoot $policyPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullPath) | Out-Null
    @"
tenant_id: family-varadha
device_id: phase109-otel-device
profile: ai-btech-student
collection:
  transparency_mode: visible_indicator_required
  browser:
    url_mode: domain_only
    collect_page_title: false
    youtube_classification: enabled
    youtube_video_id_mode: hashed
  media:
    collect_file_name: true
    collect_file_path: true
    path_mode: full_path
  sensitive_capabilities:
    credentials: deny
    keystrokes: deny
    cookies: deny
    tokens: deny
    private_messages: deny
    screenshots: deny
retention:
  local_ttl_days: 90
  max_local_storage_mb: 512
archive:
  enabled: false
  provider: none
  bucket: ""
  prefix_template: ""
  upload_interval: ""
  retry_when_online: true
  storage_class_days:
    standard: 90
    standard_ia_until: 365
    archive_after: 365
backend_sync:
  enabled: false
  base_url: http://127.0.0.1:18080
  batch_limit: 100
  request_timeout: 10s
observability:
  opentelemetry:
    enabled: true
    protocol: otlp_http_json
    endpoint: http://$Addr/v1/logs
    batch_limit: 100
    request_timeout: 5s
    retry:
      max_attempts: 2
alerts:
  enabled: false
  email:
    provider: none
    from: ""
    to: []
    min_severity: high
    cooldown_minutes: 30
thresholds:
  max_video_minutes_per_day: 60
  max_social_minutes_per_day: 30
  max_unknown_app_minutes_per_day: 45
  late_night_usage_start: "23:30"
  late_night_usage_end: "05:00"
alert_rules: {}
"@ | Set-Content -Path $fullPath -Encoding UTF8
}

function Invoke-TraceDeckAgentOnce {
    Invoke-TraceDeckLoggedCommand -Label "Run agent once with OpenTelemetry export" -Command {
        $script:AgentRunOutput = @(
            go run ./agent/cmd/tracedeck-agent run `
                --config $policyPath `
                --data-dir $agentDataDir `
                --log-dir logs/local/agent `
                --outbox-dir $agentOutboxDir `
                --once `
                --collection-interval 10m `
                --disable-browser-history `
                --process-limit 8 2>&1
        )
        $script:AgentRunOutput | ForEach-Object { $_ }
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
    return ($script:AgentRunOutput -join "`n")
}

function Get-TraceDeckMetric {
    param([string]$Text, [string]$Name)
    $match = [regex]::Match($Text, "$Name=([^ ]+)")
    if (-not $match.Success) {
        throw "Expected metric $Name in agent output: $Text"
    }
    return $match.Groups[1].Value
}

try {
    Start-TraceDeckFakeOTLP
    New-TraceDeckOTelPolicy

    $agentOutput = Invoke-TraceDeckAgentOnce
    $otelExported = Get-TraceDeckMetric -Text $agentOutput -Name "otel_exported"
    $otelEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "otel_events")
    $otelDropped = [int](Get-TraceDeckMetric -Text $agentOutput -Name "otel_dropped")
    $otelAttempts = [int](Get-TraceDeckMetric -Text $agentOutput -Name "otel_attempts")
    if ($otelExported -ne "true" -or $otelEvents -lt 1 -or $otelDropped -ne 0 -or $otelAttempts -ne 1) {
        throw "Expected successful OpenTelemetry export. output=$agentOutput"
    }

    $captureFullPath = Join-Path $script:TraceDeckRepoRoot $receiverCapture
    $deadline = (Get-Date).AddSeconds(30)
    while ((Get-Date) -lt $deadline -and -not (Test-Path $captureFullPath)) {
        Start-Sleep -Milliseconds 300
    }
    if (-not (Test-Path $captureFullPath)) {
        throw "Expected fake OTLP capture file: $receiverCapture"
    }
    $capture = Get-Content -Raw -Path $captureFullPath | ConvertFrom-Json
    if ($capture.path -ne "/v1/logs" -or $capture.resource_logs -lt 1 -or $capture.log_records -lt 1) {
        throw "Expected OTLP log capture with records. capture=$($capture | ConvertTo-Json -Depth 8)"
    }
    if (-not $capture.privacy_safe) {
        throw "OTLP capture included forbidden privacy markers: $($capture.forbidden_hits -join ', ')"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 109 OTLP smoke passed addr=$Addr otel_events=$otelEvents capture=$receiverCapture"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($null -ne $script:ReceiverProcess -and -not $script:ReceiverProcess.HasExited) {
        Stop-Process -Id $script:ReceiverProcess.Id -Force -ErrorAction SilentlyContinue
    }
}
