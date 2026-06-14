param(
    [string]$Addr = "127.0.0.1:18155"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase42" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase42/$timestamp"
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
        "Workspace Navigator",
        "command-navigation",
        "command-nav-status",
        "command-nav-title",
        "command-nav-alerts",
        "command-nav-routes",
        "command-nav-report",
        "command-nav-archive",
        "data-jump-target=""paid-ops-section""",
        "data-jump-target=""revenue-section""",
        "data-jump-target=""notification-proof-section""",
        "data-jump-target=""mail-report-section""",
        "data-jump-target=""archive-proof-section""",
        "data-jump-target=""trust-proof-section""",
        "data-jump-target=""host-detail-section""",
        "nav-paid-ops-meta",
        "nav-revenue-meta",
        "nav-notification-meta",
        "nav-report-meta",
        "nav-archive-meta",
        "nav-trust-meta",
        "nav-host-meta"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 42 command navigation marker '$expected'."
        }
    }

    $tenantID = "family-varadha"
    $deviceID = "demo-study-laptop"
    $inbox = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/alert-inbox"
    $routes = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/notification-routes"
    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/monetization-summary"
    $weekly = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/reports/weekly"

    if ($inbox.summary.open -lt 1) {
        throw "Expected command navigation alert KPI backing data."
    }
    if ($routes.count -lt 3) {
        throw "Expected command navigation route KPI backing data."
    }
    if ($monetization.readiness_score -lt 1 -or $monetization.notification_score -lt 1) {
        throw "Expected command navigation readiness and notification backing data."
    }
    if (-not $weekly.email_ready -or -not $weekly.pdf_ready) {
        throw "Expected command navigation report KPI backing data."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 42 command navigation smoke passed addr=$Addr alerts=$($inbox.summary.open) routes=$($routes.count) readiness=$($monetization.readiness_score)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
