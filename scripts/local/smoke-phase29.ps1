param(
    [string]$Addr = "127.0.0.1:18131"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase29" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase29/$timestamp"
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
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $baseUrl = "http://$Addr"
    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @(
        "Monetisation Launch Deck",
        "Anomaly Push Assurance",
        "Mail Delivery Assurance",
        "Weekly Report Proof",
        "Notification Revenue Stream",
        "launch-action-upgrade"
    )) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard HTML to include: $expected"
        }
    }

    $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
    if ($devices.count -lt 1) {
        throw "Expected seeded demo devices."
    }

    $deviceID = $devices.items[0].device_id
    $tenantID = $devices.items[0].tenant_id
    $overview = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/overview"
    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/alert-deliveries"
    $weekly = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/reports/weekly"
    $operations = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/operations-summary"
    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/monetization-summary"

    if ($overview.risk_score -le 0 -or $overview.anomalies.Count -lt 1) {
        throw "Expected host overview risk and anomaly proof."
    }
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "email" -and $_.status -eq "delivered" })) {
        throw "Expected delivered email proof for launch deck."
    }
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "push" })) {
        throw "Expected push route proof for launch deck."
    }
    if (-not $weekly.email_ready -or -not $weekly.pdf_ready) {
        throw "Expected weekly report email and PDF proof."
    }
    if ($operations.notification_score -le 0 -or $operations.monetization_readiness -le 0) {
        throw "Expected tenant operations notification and monetisation readiness proof."
    }
    if ($monetization.readiness_score -le 0 -or -not $monetization.notification_promise -or $monetization.notification_routes.Count -lt 3) {
        throw "Expected tenant monetisation summary with notification promise and routes."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 29 monetisation launch deck smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
