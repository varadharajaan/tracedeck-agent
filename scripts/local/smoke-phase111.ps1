param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase111" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase111/$timestamp"
$policyPath = "$smokeRoot/policy.yaml"
$agentDataDir = "$smokeRoot/agent-data"
$agentOutboxDir = "$smokeRoot/outbox"
$softwareCacheDir = "$smokeRoot/software-cache"
$snapshotPath = "$softwareCacheDir/software-inventory-snapshot.json"

function New-TraceDeckSoftwarePolicy {
    $fullPath = Join-Path $script:TraceDeckRepoRoot $policyPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullPath) | Out-Null
    @"
tenant_id: family-varadha
device_id: phase111-software-device
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
  software:
    enabled: true
    inventory_mode: metadata_only
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
  unknown_software_installed:
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

function Add-TraceDeckRemovedSoftwareFixture {
    $fullSnapshotPath = Join-Path $script:TraceDeckRepoRoot $snapshotPath
    if (-not (Test-Path $fullSnapshotPath)) {
        throw "Expected software inventory baseline snapshot at $snapshotPath"
    }
    $snapshot = Get-Content -Path $fullSnapshotPath -Raw | ConvertFrom-Json
    if ($null -eq $snapshot.entries) {
        $snapshot | Add-Member -NotePropertyName "entries" -NotePropertyValue ([pscustomobject]@{}) -Force
    }
    $fixture = [pscustomobject]@{
        id_hash = "phase111-removed-fixture-id"
        name = "TraceDeck Removed Fixture"
        name_hash = "phase111-removed-fixture-name"
        version = "0.0.1"
        publisher = "TraceDeck Test"
        source = "phase111_test_fixture"
    }
    $snapshot.entries | Add-Member -NotePropertyName "phase111_removed_fixture" -NotePropertyValue $fixture -Force
    $json = $snapshot | ConvertTo-Json -Depth 10
    $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
    [System.IO.File]::WriteAllText($fullSnapshotPath, $json + [Environment]::NewLine, $utf8NoBom)
}

try {
    New-TraceDeckSoftwarePolicy

    Invoke-TraceDeckLoggedCommand -Label "Run agent once to baseline software inventory" -Command {
        $script:BaselineOutput = @(
            go run ./agent/cmd/tracedeck-agent run `
                --config $policyPath `
                --data-dir $agentDataDir `
                --log-dir logs/local/agent `
                --outbox-dir $agentOutboxDir `
                --software-cache-dir $softwareCacheDir `
                --once `
                --collection-interval 10m `
                --disable-browser-history `
                --process-limit 6 2>&1
        )
        $script:BaselineOutput | ForEach-Object { $_ }
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    Add-TraceDeckRemovedSoftwareFixture

    Invoke-TraceDeckLoggedCommand -Label "Run agent once to detect software inventory change" -Command {
        $script:AgentRunOutput = @(
            go run ./agent/cmd/tracedeck-agent run `
                --config $policyPath `
                --data-dir $agentDataDir `
                --log-dir logs/local/agent `
                --outbox-dir $agentOutboxDir `
                --software-cache-dir $softwareCacheDir `
                --once `
                --collection-interval 10m `
                --disable-browser-history `
                --process-limit 6 2>&1
        )
        $script:AgentRunOutput | ForEach-Object { $_ }
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    $agentOutput = ($script:AgentRunOutput -join "`n")
    $softwareEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "software_events")
    $collectedEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "collected_events")
    if ($softwareEvents -lt 1) {
        throw "Expected at least one software inventory change event. output=$agentOutput"
    }
    if ($collectedEvents -lt 3) {
        throw "Expected agent run to collect baseline metadata events. output=$agentOutput"
    }
    foreach ($forbidden in @("password", "screenshot", "raw_url", "page_title", "window_title`"", "cookie", "token", "provider_secret", "alert_body")) {
        if ($agentOutput -match $forbidden) {
            throw "Agent output contains forbidden marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 111 software inventory smoke passed software_events=$softwareEvents collected_events=$collectedEvents"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
