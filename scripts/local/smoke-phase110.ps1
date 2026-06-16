param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase110" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase110/$timestamp"
$policyPath = "$smokeRoot/policy.yaml"
$agentDataDir = "$smokeRoot/agent-data"
$agentOutboxDir = "$smokeRoot/outbox"

function New-TraceDeckForegroundPolicy {
    $fullPath = Join-Path $script:TraceDeckRepoRoot $policyPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullPath) | Out-Null
    @"
tenant_id: family-varadha
device_id: phase110-active-window-device
profile: ai-btech-student
collection:
  transparency_mode: visible_indicator_required
  browser:
    url_mode: domain_only
    collect_page_title: false
    youtube_classification: enabled
    youtube_video_id_mode: hashed
  foreground_app:
    enabled: true
    window_title_mode: none
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
    enabled: false
    protocol: otlp_http_json
    endpoint: http://127.0.0.1:4318/v1/logs
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
alert_rules:
  blocked_app_opened:
    enabled: true
    severity: high
  risky_software_detected:
    enabled: true
    severity: high
"@ | Set-Content -Path $fullPath -Encoding UTF8
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
    New-TraceDeckForegroundPolicy

    Invoke-TraceDeckLoggedCommand -Label "Run agent once with foreground app collection enabled" -Command {
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

    $agentOutput = ($script:AgentRunOutput -join "`n")
    $foregroundEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "foreground_events")
    $collectedEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "collected_events")
    if ($foregroundEvents -lt 0 -or $foregroundEvents -gt 1) {
        throw "Expected foreground event count to be bounded to 0 or 1. output=$agentOutput"
    }
    if ($collectedEvents -lt 2) {
        throw "Expected agent run to collect baseline metadata events. output=$agentOutput"
    }
    foreach ($forbidden in @("password", "screenshot", "raw_url", "page_title", "window_title", "cookie", "token", "keylog")) {
        if ($agentOutput -match $forbidden) {
            throw "Agent output contains forbidden marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 110 foreground app smoke passed foreground_events=$foregroundEvents collected_events=$collectedEvents"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
