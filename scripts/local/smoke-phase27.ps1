param(
    [string]$Addr = "127.0.0.1:18127"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase27" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase27/$timestamp"
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
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                if ($health.status -eq "ok") {
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
    throw "Dashboard demo helper did not become ready at $baseUrl"
}

function Wait-TraceDeckSeededDevices {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        try {
            $devices = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/api/v1/devices"
            if ($devices.count -ge 1) {
                return $devices
            }
        }
        catch { Start-Sleep -Milliseconds 500 }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo seed did not expose devices at $BaseUrl"
}

try {
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $baseUrl = "http://$Addr"
    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Revenue Control Room", "Buyer Notification Assurance", "revenue-package", "assurance-email", "assurance-push")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard HTML to include: $expected"
        }
    }

    $devices = Wait-TraceDeckSeededDevices -BaseUrl $baseUrl

    $deviceID = $devices.items[0].device_id
    $tenantID = $devices.items[0].tenant_id
    $overview = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/overview"
    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/alert-deliveries"
    $weekly = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/reports/weekly"
    $operations = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/operations-summary"
    $monetization = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/monetization-summary"

    if ($overview.device.device_id -ne $deviceID) {
        throw "Expected overview to match selected demo device."
    }
    if ($deliveries.count -lt 3) {
        throw "Expected email, push, and dashboard alert delivery proof."
    }
    if (-not $weekly.email_ready) {
        throw "Expected weekly report email readiness proof."
    }
    if ($operations.notification_score -le 0 -or $operations.monetization_readiness -le 0) {
        throw "Expected tenant operations notification and monetisation scores."
    }
    if ($monetization.readiness_score -le 0 -or $monetization.value_panels.Count -lt 1 -or $monetization.paid_capabilities.Count -lt 1) {
        throw "Expected monetisation summary value panels and paid capabilities."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 27 revenue control room smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
