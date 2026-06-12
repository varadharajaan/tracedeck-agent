param(
    [string]$Addr = "127.0.0.1:18133"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase30" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase30/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$policyPath = "$smokeRoot/policy.yaml"
$agentDataDir = "$smokeRoot/agent-data"
$agentOutboxDir = "$smokeRoot/outbox"

function Start-TraceDeckDashboardDemo {
    param([string]$ListenAddr, [string]$RelativePidPath, [string]$RelativeDataPath)

    Write-TraceDeckLog -Level "INFO" -Message "Starting dashboard demo helper addr=$ListenAddr pid_path=$RelativePidPath"
    $helper = Start-Process -FilePath "powershell" -ArgumentList @(
        "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-File", "./scripts/local/start-dashboard-demo.ps1",
        "-Addr", $ListenAddr,
        "-PidPath", $RelativePidPath,
        "-DataPath", $RelativeDataPath
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -PassThru

    $baseUrl = "http://$ListenAddr"
    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $RelativePidPath
    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        if ((Test-Path $pidFullPath)) {
            try {
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($devices.count -ge 1) {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper completed readiness addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch { Start-Sleep -Milliseconds 500 }
        }
        elseif ($helper.HasExited -and $helper.ExitCode -ne 0) {
            throw "Dashboard demo helper failed with exit code $($helper.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo helper did not seed devices at $baseUrl"
}

function New-TraceDeckSyncPolicy {
    param([string]$RelativePolicyPath, [string]$BaseUrl)

    $fullPath = Join-Path $script:TraceDeckRepoRoot $RelativePolicyPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $fullPath) | Out-Null
    @"
tenant_id: family-varadha
device_id: laptop-cousin-001
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
  enabled: true
  base_url: $BaseUrl
  batch_limit: 100
  request_timeout: 2s
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
    param([string]$Label)

    Invoke-TraceDeckLoggedCommand -Label $Label -Command {
        $script:AgentRunOutput = @(
            go run ./agent/cmd/tracedeck-agent run `
                --config $policyPath `
                --data-dir $agentDataDir `
                --log-dir logs/local/agent `
                --outbox-dir $agentOutboxDir `
                --once `
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
    New-TraceDeckSyncPolicy -RelativePolicyPath $policyPath -BaseUrl "http://127.0.0.1:9"
    $offlineOutput = Invoke-TraceDeckAgentOnce -Label "Run agent once while backend is offline"
    $offlineSynced = Get-TraceDeckMetric -Text $offlineOutput -Name "telemetry_synced"
    $offlineBacklog = [int](Get-TraceDeckMetric -Text $offlineOutput -Name "telemetry_backlog")
    if ($offlineSynced -ne "false" -or $offlineBacklog -lt 1) {
        throw "Expected offline run to succeed with unsynced backlog. output=$offlineOutput"
    }

    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    New-TraceDeckSyncPolicy -RelativePolicyPath $policyPath -BaseUrl $baseUrl
    $onlineOutput = Invoke-TraceDeckAgentOnce -Label "Run agent once after backend returns online"
    $onlineSynced = Get-TraceDeckMetric -Text $onlineOutput -Name "telemetry_synced"
    $onlineEvents = [int](Get-TraceDeckMetric -Text $onlineOutput -Name "telemetry_events")
    if ($onlineSynced -ne "true" -or $onlineEvents -lt $offlineBacklog) {
        throw "Expected online run to replay offline backlog. output=$onlineOutput offline_backlog=$offlineBacklog"
    }

    $status = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/laptop-cousin-001/telemetry-status"
    if ($status.stored_events -lt $offlineBacklog) {
        throw "Expected backend telemetry status to include replayed offline events. stored=$($status.stored_events) offline_backlog=$offlineBacklog"
    }
    if (-not ($status.recent_events | Where-Object { $_.id -like "local-event-*" })) {
        throw "Expected backend telemetry status to include stable local event IDs."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 30 offline backend sync smoke passed addr=$Addr offline_backlog=$offlineBacklog synced_events=$onlineEvents stored_events=$($status.stored_events)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
