param(
    [string]$Addr = "127.0.0.1:18167"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase48" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase48/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase48-smoke/$timestamp"

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
    $tenantID = "family-varadha"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Growth Cockpit",
        "Paid product promise",
        "Anomaly Notification Ops",
        "Notification Delivery Proof",
        "Monetisation Owner Actions",
        "growth-cockpit-section",
        "growth-alert-ops-list",
        "growth-delivery-proof-list",
        "growth-owner-action-list"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 48 dashboard marker '$expected'."
        }
    }

    $commandCenter = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/notification-command-center"
    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/monetization-summary"
    $operations = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/operations-summary"
    if ($commandCenter.summary.open_alerts -lt 1 -or $commandCenter.channels.Count -lt 3 -or $commandCenter.actions.Count -lt 1) {
        throw "Expected notification command center to expose alerts, routes, and actions for Growth Cockpit."
    }
    if ($monetization.readiness_score -lt 1 -or $monetization.notification_score -lt 1) {
        throw "Expected monetization summary to expose readiness and notification score."
    }
    if ($operations.email_delivered -lt 1 -or $operations.delivery_total -lt 1) {
        throw "Expected operations summary to expose mail delivery proof."
    }
    if ($commandCenter.privacy_boundary -notmatch "no passwords" -or $commandCenter.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict notification command privacy boundary."
    }
    $serialized = ($commandCenter | ConvertTo-Json -Depth 20).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Notification command center leaked forbidden field marker '$forbidden'."
        }
    }

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot

    Write-TraceDeckLog -Level "INFO" -Message "Phase 48 growth cockpit smoke passed addr=$Addr readiness=$($monetization.readiness_score) notification_score=$($commandCenter.summary.notification_score) routes=$($commandCenter.channels.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
