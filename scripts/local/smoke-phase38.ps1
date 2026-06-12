param(
    [string]$Addr = "127.0.0.1:18147"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase38" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase38/$timestamp"
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
    Invoke-TraceDeckLoggedCommand -Label "Dashboard contract guard" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
    }

    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Commercial Control Room",
        "Alert Delivery Evidence",
        "Customer Success Queue",
        "commercial-room-status",
        "commercial-delivery-list",
        "commercial-success-list",
        "commercial-sla"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected commercial dashboard marker '$expected'."
        }
    }

    $tenantID = "family-varadha"
    $operations = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/operations-summary"
    if ($operations.delivery_total -lt 1 -or $operations.email_delivered -lt 1 -or $operations.open_anomalies -lt 1) {
        throw "Expected tenant operations to back commercial alert, mail, and anomaly proof."
    }

    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/monetization-summary"
    if ($monetization.readiness_score -lt 1 -or $monetization.notification_score -lt 1 -or $monetization.conversion_actions.Count -lt 1) {
        throw "Expected monetisation summary to back commercial success queue."
    }

    $routes = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/notification-routes"
    if ($routes.count -lt 3 -or -not ($routes.items | Where-Object { $_.channel -eq "push" })) {
        throw "Expected email, push, and dashboard notification route proof."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 38 commercial dashboard smoke passed addr=$Addr readiness=$($monetization.readiness_score) notification=$($monetization.notification_score)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
