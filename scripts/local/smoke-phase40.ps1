param(
    [string]$Addr = "127.0.0.1:18151"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase40" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase40/$timestamp"
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
    foreach ($expected in @(
        "Paid Ops Console",
        "Notification Route Proof",
        "Monetisation Action Queue",
        "paid-ops-status",
        "paid-push",
        "paid-mail",
        "paid-route-proof-list"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 40 paid dashboard marker '$expected'."
        }
    }

    $tenantID = "family-varadha"
    $operations = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/operations-summary"
    if ($operations.delivery_total -lt 3 -or $operations.email_delivered -lt 1 -or $operations.push_delivered -lt 0) {
        throw "Expected operations summary to include monetisation delivery proof."
    }

    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/monetization-summary"
    if ($monetization.notification_routes.Count -lt 3 -or $monetization.readiness_score -lt 1 -or $monetization.notification_score -lt 1) {
        throw "Expected monetization summary to include notification route proof and paid readiness."
    }

    $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
    $deviceID = $devices.items[0].device_id
    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/alert-deliveries"
    $channels = @($deliveries.items | ForEach-Object { $_.channel })
    foreach ($channel in @("email", "push", "dashboard")) {
        if ($channels -notcontains $channel) {
            throw "Expected alert delivery channel '$channel' in paid ops proof."
        }
    }
    if (@($deliveries.items | Where-Object { $_.event_id }).Count -lt 1) {
        throw "Expected alert deliveries to be linked to anomaly or policy event IDs."
    }

    $weekly = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/reports/weekly"
    if (-not $weekly.email_ready -or -not $weekly.pdf_ready) {
        throw "Expected weekly report email and PDF proof to be ready."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 40 paid ops smoke passed addr=$Addr readiness=$($monetization.readiness_score) notification=$($monetization.notification_score) routes=$($monetization.notification_routes.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
