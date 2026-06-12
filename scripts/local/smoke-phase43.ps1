param(
    [string]$Addr = "127.0.0.1:18157"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase43" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase43/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase43-smoke/$timestamp"

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
        "Buyer Operations Brief",
        "buyer-ops-section",
        "buyer-delivery-list",
        "buyer-package-list",
        "buyer-action-list",
        "Push Notification Dispatch",
        "Mail Delivery Proof",
        "Action And SLA"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 43 buyer operations dashboard marker '$expected'."
        }
    }

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot

    $layoutFullRoot = Join-Path $script:TraceDeckRepoRoot $layoutRoot
    $report = Get-ChildItem -Path $layoutFullRoot -Recurse -Filter "dashboard-layout-report.json" |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    if (-not $report) {
        throw "Expected dashboard layout report under $layoutFullRoot"
    }
    $model = Get-Content -Path $report.FullName -Raw | ConvertFrom-Json
    if (-not $model.ok -or $model.privacy_boundary -notmatch "no screenshots") {
        throw "Expected screenshot-free dashboard layout report to pass."
    }
    if (@($model.viewports).Count -lt 3) {
        throw "Expected desktop, tablet, and mobile layout viewports."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 43 layout smoke passed addr=$Addr report=$($report.FullName)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
