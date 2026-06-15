param(
    [string]$Addr = "127.0.0.1:18137"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase32" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase32/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"

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

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @("Tenant Activity Feed", "Filtered Command Feed", "activity-feed-list")) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard marker '$expected'."
        }
    }

    $feed = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/activity-feed?limit=20"
    $feedJson = $feed | ConvertTo-Json -Depth 20
    if ($feed.filters.include_demo -or $feed.summary.risk_items -ne 0 -or $feed.summary.delivery_items -ne 0 -or $feedJson.Contains("demo_seed")) {
        throw "Expected default activity feed to hide demo risk and delivery items."
    }
    if ($feed.privacy_boundary -notmatch "metadata-only") {
        throw "Expected metadata-only privacy boundary on activity feed."
    }

    $demoFeed = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/activity-feed?limit=20&include_demo=true"
    if (-not $demoFeed.filters.include_demo -or $demoFeed.summary.risk_items -lt 1 -or $demoFeed.summary.delivery_items -lt 1 -or $demoFeed.items.Count -lt 1) {
        throw "Expected opt-in demo activity feed to include labelled risk and delivery items."
    }

    $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
    $selectedDeviceID = $devices.items[0].device_id
    $emailFeed = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/activity-feed?device_id=$selectedDeviceID&kind=delivery&channel=email&limit=5&include_demo=true"
    if ($emailFeed.summary.delivery_items -lt 1 -or -not ($emailFeed.items | Where-Object { $_.channel -eq "email" })) {
        throw "Expected selected-host email delivery feed item."
    }

    $payload = @{
        tenant_id = "family-varadha"
        device_id = "phase32-feed-device"
        host_name = "phase32-feed-host"
        profile = "ai-btech-student"
        os_name = "windows"
        events = @(@{
            id = "local-event-3201"
            type = "process.observed"
            source = "collector.process"
            observed_at = "2026-06-12T08:00:00Z"
            app_name = "Code.exe"
            process_id = 3201
            path_hash = "hash-only"
            metadata = @{ category = "coding" }
        })
    } | ConvertTo-Json -Depth 8
    $ingest = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/devices/phase32-feed-device/telemetry-events" -ContentType "application/json" -Body $payload
    if ($ingest.accepted_events -ne 1 -or -not $ingest.backend_visible_host) {
        throw "Expected telemetry ingest to feed the activity timeline."
    }
    $telemetryFeed = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/activity-feed?device_id=phase32-feed-device&kind=telemetry&limit=5"
    if ($telemetryFeed.summary.telemetry_items -ne 1 -or $telemetryFeed.items[0].id -ne "local-event-3201") {
        throw "Expected activity feed telemetry item with stable local event ID."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 32 smoke passed addr=$Addr feed_total=$($feed.summary.total) telemetry_items=$($telemetryFeed.summary.telemetry_items)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
