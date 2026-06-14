param(
    [string]$Addr = "127.0.0.1:18237"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase83" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase83/$timestamp"
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
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($health.status -eq "ok" -and $devices.count -ge 1) {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper ready addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch {
                Start-Sleep -Milliseconds 500
            }
        }
        elseif ($helper.HasExited -and $helper.ExitCode -ne 0) {
            throw "Dashboard demo helper failed with exit code $($helper.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo helper did not seed devices at $baseUrl"
}

function New-TraceDeckHeartbeatPolicy {
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
  request_timeout: 10s
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
    Invoke-TraceDeckLoggedCommand -Label "Run agent once with heartbeat telemetry sync" -Command {
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

function Get-TraceDeckProperty {
    param([object]$Object, [string]$Name)
    if ($null -eq $Object) {
        return $null
    }
    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $null
    }
    return $property.Value
}

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    New-TraceDeckHeartbeatPolicy -RelativePolicyPath $policyPath -BaseUrl $baseUrl

    $agentOutput = Invoke-TraceDeckAgentOnce
    $heartbeatEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "heartbeat_events")
    $telemetrySynced = Get-TraceDeckMetric -Text $agentOutput -Name "telemetry_synced"
    $telemetryEvents = [int](Get-TraceDeckMetric -Text $agentOutput -Name "telemetry_events")
    if ($heartbeatEvents -ne 1) {
        throw "Expected one heartbeat event from agent cycle. output=$agentOutput"
    }
    if ($telemetrySynced -ne "true" -or $telemetryEvents -lt 1) {
        throw "Expected backend sync to accept heartbeat telemetry. output=$agentOutput"
    }

    $status = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/laptop-cousin-001/telemetry-status"
    $heartbeatTypeCount = [int](Get-TraceDeckProperty -Object $status.counts_by_type -Name "agent.health.heartbeat")
    $heartbeatSourceCount = [int](Get-TraceDeckProperty -Object $status.counts_by_source -Name "collector.agent.heartbeat")
    if ($status.stored_events -lt 1 -or $heartbeatTypeCount -lt 1 -or $heartbeatSourceCount -lt 1) {
        throw "Expected telemetry status to count heartbeat events. status=$($status | ConvertTo-Json -Depth 8)"
    }
    if (-not $status.privacy_boundary -or $status.privacy_boundary -notmatch "metadata-only") {
        throw "Expected metadata-only privacy boundary in telemetry status."
    }

    $heartbeat = $status.recent_events | Where-Object { $_.type -eq "agent.health.heartbeat" } | Select-Object -First 1
    if ($null -eq $heartbeat) {
        throw "Expected recent telemetry events to include the heartbeat event."
    }
    $metadata = $heartbeat.metadata
    if ((Get-TraceDeckProperty -Object $metadata -Name "agent_healthy") -ne "true" -or
        (Get-TraceDeckProperty -Object $metadata -Name "backend_sync_enabled") -ne "true" -or
        (Get-TraceDeckProperty -Object $metadata -Name "collection_mode") -ne "once" -or
        (Get-TraceDeckProperty -Object $metadata -Name "collection_interval") -ne "10m") {
        throw "Expected typed heartbeat readiness metadata. metadata=$($metadata | ConvertTo-Json -Depth 8)"
    }
    foreach ($forbidden in @("password", "screenshot", "raw_url", "page_title", "cookie", "token", "keylog")) {
        if ($metadata.PSObject.Properties.Name -contains $forbidden) {
            throw "Heartbeat metadata contains forbidden key '$forbidden'."
        }
    }

    $syncHealth = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/sync-health"
    $syncDevice = $syncHealth.devices | Where-Object { $_.device_id -eq "laptop-cousin-001" } | Select-Object -First 1
    if (-not $syncHealth.backend_visible -or -not $syncHealth.offline_replay_ready -or $null -eq $syncDevice -or -not $syncDevice.backend_visible) {
        throw "Expected tenant sync health to show backend-visible heartbeat host. sync=$($syncHealth | ConvertTo-Json -Depth 8)"
    }
    if ($syncDevice.health_events -lt 1 -or -not ($syncDevice.recent_event_ids | Where-Object { $_ -like "local-event-*" })) {
        throw "Expected heartbeat to contribute to health/replay sync proof. device=$($syncDevice | ConvertTo-Json -Depth 8)"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 83 heartbeat telemetry smoke passed addr=$Addr heartbeat_type_count=$heartbeatTypeCount stored_events=$($status.stored_events)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
