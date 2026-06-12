param(
    [string]$Addr = "127.0.0.1:18113"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase24" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase24/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"

function Get-TraceDeckPid {
    param([string]$RelativePidPath)
    $fullPath = Join-Path $script:TraceDeckRepoRoot $RelativePidPath
    if (-not (Test-Path $fullPath)) { throw "Missing pid file: $fullPath" }
    return [int]((Get-Content -Path $fullPath -Raw).Trim())
}

function Start-TraceDeckDashboardDemo {
    param(
        [string]$ListenAddr,
        [string]$RelativePidPath,
        [string]$RelativeDataPath,
        [int]$PreviousPid = 0
    )

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
            $currentPid = [int]((Get-Content -Path $pidFullPath -Raw).Trim())
            if ($PreviousPid -ne 0 -and $currentPid -eq $PreviousPid) {
                Start-Sleep -Milliseconds 500
                continue
            }
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

try {
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    $firstPid = Get-TraceDeckPid -RelativePidPath $pidPath
    if (-not (Get-Process -Id $firstPid -ErrorAction SilentlyContinue)) {
        throw "Expected first dashboard demo process to be running: $firstPid"
    }

    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath -PreviousPid $firstPid
    $secondPid = Get-TraceDeckPid -RelativePidPath $pidPath
    if ($secondPid -eq $firstPid) {
        throw "Expected restart to replace first dashboard demo process."
    }
    if (Get-Process -Id $firstPid -ErrorAction SilentlyContinue) {
        throw "Expected stale first dashboard demo process to be stopped: $firstPid"
    }
    if (-not (Get-Process -Id $secondPid -ErrorAction SilentlyContinue)) {
        throw "Expected second dashboard demo process to be running: $secondPid"
    }

    $connection = Get-NetTCPConnection -LocalPort ([int](($Addr -split ":", 2)[1])) -State Listen -ErrorAction Stop |
        Where-Object { $_.OwningProcess -eq $secondPid } |
        Select-Object -First 1
    if (-not $connection) {
        throw "Expected second dashboard demo process to own listener on $Addr."
    }

    $baseUrl = "http://$Addr"
    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Customer Operations Cockpit", "Escalation Workbench", "Notification Delivery Board", "Upgrade Proof Pack")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard HTML to include: $expected"
        }
    }

    $summary = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/operations-summary"
    if ($summary.hosts_total -lt 1 -or $summary.delivery_total -lt 1 -or -not $summary.last_email) {
        throw "Expected live operations summary after lifecycle restart."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 24 dashboard lifecycle smoke passed first_pid=$firstPid second_pid=$secondPid addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
