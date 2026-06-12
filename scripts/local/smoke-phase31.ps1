param(
    [string]$Addr = "127.0.0.1:18135"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase31" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase31/$timestamp"
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

    $payload = @{
        tenant_id = "family-varadha"
        device_id = "phase31-sync-device"
        host_name = "phase31-sync-host"
        profile = "ai-btech-student"
        os_name = "windows"
        events = @(
            @{
                id = "local-event-31"
                type = "process.observed"
                source = "collector.process"
                observed_at = "2026-06-12T08:00:00Z"
                app_name = "Code.exe"
                process_id = 301
                path_hash = "hash-only"
                metadata = @{ category = "coding" }
            },
            @{
                id = "local-event-32"
                type = "browser.summary"
                source = "collector.browser.history"
                observed_at = "2026-06-12T08:01:00Z"
                metadata = @{ domain = "youtube.com"; category = "study" }
            },
            @{
                id = "local-event-33"
                type = "device.health"
                source = "collector.device.health"
                observed_at = "2026-06-12T08:02:00Z"
                metadata = @{ battery = "ok" }
            }
        )
    } | ConvertTo-Json -Depth 8
    $ingest = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/devices/phase31-sync-device/telemetry-events" -ContentType "application/json" -Body $payload
    if ($ingest.accepted_events -ne 3 -or -not $ingest.backend_visible_host) {
        throw "Expected phase31 sync telemetry to be accepted and backend-visible."
    }

    $syncHealth = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/sync-health"
    if ($syncHealth.hosts_reporting -lt 1 -or $syncHealth.stored_events -lt 3 -or $syncHealth.last_local_event_id -lt 33) {
        throw "Expected tenant sync health to expose reporting hosts, stored events, and stable replay cursor."
    }
    if (-not $syncHealth.offline_replay_ready -or $syncHealth.privacy_boundary -notmatch "metadata-only") {
        throw "Expected offline replay readiness and metadata-only privacy boundary."
    }
    $phaseDevice = $syncHealth.devices | Where-Object { $_.device_id -eq "phase31-sync-device" } | Select-Object -First 1
    if (-not $phaseDevice -or $phaseDevice.process_events -ne 1 -or $phaseDevice.browser_events -ne 1 -or $phaseDevice.health_events -ne 1) {
        throw "Expected source counts for phase31 sync device."
    }

    $operations = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/operations-summary"
    if ($operations.email_delivered -lt 1 -or $operations.delivery_total -lt 1) {
        throw "Expected email delivery proof in tenant operations."
    }
    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/monetization-summary"
    if ([string]::IsNullOrWhiteSpace($monetization.notification_promise.email) -or [string]::IsNullOrWhiteSpace($monetization.notification_promise.push)) {
        throw "Expected monetisation notification promise to include email and push."
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Buyer Assurance Wall",
        "Offline Replay Health",
        "product-assurance-push",
        "product-assurance-mail",
        "sync-health-device-list",
        "Mail Delivery Assurance",
        "Push Notification Center"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard marker '$expected'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 31 smoke passed addr=$Addr stored_events=$($syncHealth.stored_events) cursor=$($syncHealth.last_local_event_id)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
