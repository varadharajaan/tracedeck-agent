param(
    [string]$Addr = "127.0.0.1:18177"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase53" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase53/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase53-smoke/$timestamp"

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
        "Executive Notification Console",
        "executive-console-section",
        "executive-kpi-anomalies",
        "executive-kpi-mail",
        "executive-kpi-push",
        "Value Tiles",
        "Anomaly Alert Stream",
        "Mail And Push Proof",
        "Owner Action Board",
        "data-jump-target=`"executive-console-section`"",
        "Business Dashboard",
        "Role Experience Center"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 53 dashboard marker '$expected'."
        }
    }

    $console = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/executive-console"
    if ($console.summary.readiness_score -le 0 -or $console.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($console.summary.recommended_paid_package)) {
        throw "Expected executive console to expose readiness, notification score, and recommended package."
    }
    if ($console.summary.email_delivered -lt 1 -or $console.summary.dashboard_delivered -lt 1) {
        throw "Expected executive console mail and dashboard delivery proof."
    }
    if ($console.tiles.Count -lt 8 -or $console.alerts.Count -lt 1 -or $console.deliveries.Count -lt 3 -or $console.actions.Count -lt 1) {
        throw "Expected executive console tiles, alerts, deliveries, and actions."
    }
    $pushRoute = $console.deliveries | Where-Object { $_.channel -eq "push" -and -not [string]::IsNullOrWhiteSpace($_.status) -and -not [string]::IsNullOrWhiteSpace($_.paid_tier) } | Select-Object -First 1
    if ($null -eq $pushRoute) {
        throw "Expected executive console push notification proof."
    }
    if ($console.privacy_boundary -notmatch "metadata-only" -or $console.privacy_boundary -notmatch "no passwords" -or $console.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict executive console privacy boundary."
    }
    $serialized = ($console | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Executive console leaked forbidden field marker '$forbidden'."
        }
    }

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot

    Write-TraceDeckLog -Level "INFO" -Message "Phase 53 executive console smoke passed addr=$Addr score=$($console.summary.readiness_score) deliveries=$($console.deliveries.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
