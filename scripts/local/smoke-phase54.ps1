param(
    [string]$Addr = "127.0.0.1:18179"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase54" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase54/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase54-smoke/$timestamp"

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
        "Notification Revenue Cockpit",
        "notification-revenue-section",
        "Revenue KPI Proof",
        "Anomaly Delivery Scenarios",
        "Channel Proof Matrix",
        "Upgrade Action Levers",
        "data-jump-target=`"notification-revenue-section`"",
        "Executive Notification Console",
        "Business Dashboard"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 54 dashboard marker '$expected'."
        }
    }

    $cockpit = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/notification-revenue-cockpit"
    if ($cockpit.summary.revenue_readiness -le 0 -or $cockpit.summary.notification_score -le 0 -or $cockpit.summary.alert_sla_ready -le 0) {
        throw "Expected notification revenue cockpit readiness, notification, and alert SLA scores."
    }
    if ($cockpit.summary.email_delivered -lt 1 -or $cockpit.summary.dashboard_delivered -lt 1 -or [string]::IsNullOrWhiteSpace($cockpit.summary.recommended_paid_package)) {
        throw "Expected notification revenue cockpit mail/dashboard proof and recommended package."
    }
    if ($cockpit.kpis.Count -lt 6 -or $cockpit.channels.Count -lt 3 -or $cockpit.scenarios.Count -lt 4 -or $cockpit.actions.Count -lt 1) {
        throw "Expected notification revenue KPI, channel, scenario, and action surfaces."
    }
    $pushProof = $cockpit.channels | Where-Object { $_.channel -eq "push" -and -not [string]::IsNullOrWhiteSpace($_.business_value) -and -not [string]::IsNullOrWhiteSpace($_.paid_tier) } | Select-Object -First 1
    if ($null -eq $pushProof) {
        throw "Expected notification revenue push proof with business value."
    }
    if ($cockpit.privacy_boundary -notmatch "metadata-only" -or $cockpit.privacy_boundary -notmatch "no passwords" -or $cockpit.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict notification revenue privacy boundary."
    }
    $serialized = ($cockpit | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Notification revenue cockpit leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 54 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 54 notification revenue cockpit smoke passed addr=$Addr readiness=$($cockpit.summary.revenue_readiness) channels=$($cockpit.channels.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
