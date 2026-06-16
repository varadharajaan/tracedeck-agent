param(
    [string]$Addr = "127.0.0.1:18264"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-refresh-backend-ready-proof" -LogRoot "logs/local/test" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$testRoot = "data/local/test-refresh-ready/$timestamp"
$pidPath = "$testRoot/tracedeck-backend.pid"
$dataPath = "$testRoot/backend-state.json"
$readyPath = "$testRoot/backend-task-ready.json"
$taskStatusPath = "$testRoot/backend-task-status.json"

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
        if (Test-Path $pidFullPath) {
            try {
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($health.status -eq "ok" -and $devices.count -ge 1) {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper ready addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch {
                Start-Sleep -Milliseconds 500
            }
        }
        elseif ($helper.HasExited -and $helper.ExitCode -ne 0) {
            throw "Dashboard demo helper failed with exit code $($helper.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo helper did not seed devices at $baseUrl"
}

function Write-TraceDeckStaleReadyFile {
    param(
        [string]$RelativePidPath,
        [string]$RelativeReadyPath,
        [string]$RelativeDataPath,
        [string]$ListenAddr
    )

    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $RelativePidPath
    $readyFullPath = Join-Path $script:TraceDeckRepoRoot $RelativeReadyPath
    $livePid = [int]((Get-Content -Path $pidFullPath -Raw).Trim())
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $readyFullPath) | Out-Null

    $ready = [ordered]@{
        addr = $ListenAddr
        base_url = "http://$ListenAddr"
        pid = $livePid + 100000
        ready_at = (Get-Date).AddMinutes(-10).ToString("o")
        pid_path = (Join-Path $script:TraceDeckRepoRoot $RelativePidPath)
        data_path = (Join-Path $script:TraceDeckRepoRoot $RelativeDataPath)
        stdout = "logs/local/backend/phase103-refresh-ready.out.log"
        stderr = "logs/local/backend/phase103-refresh-ready.err.log"
    }
    $ready | ConvertTo-Json -Depth 6 | Set-Content -Path $readyFullPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Wrote stale ready PID proof live_pid=$livePid ready_pid=$($ready.pid) ready_path=$RelativeReadyPath"
}

function Read-TraceDeckTaskStatus {
    param([string]$ListenAddr)

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
        -Addr $ListenAddr `
        -TaskName "\TraceDeck\TraceDeck Phase103 Refresh Missing" `
        -PidPath $pidPath `
        -ReadyPath $readyPath `
        -OutputPath $taskStatusPath | Out-Null
    return Get-Content -Path (Join-Path $script:TraceDeckRepoRoot $taskStatusPath) -Raw | ConvertFrom-Json
}

try {
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    Write-TraceDeckStaleReadyFile -RelativePidPath $pidPath -RelativeReadyPath $readyPath -RelativeDataPath $dataPath -ListenAddr $Addr

    $stale = Read-TraceDeckTaskStatus -ListenAddr $Addr
    if ($stale.ready_pid_status -ne "stale" -or $stale.ready_pid_matches_live) {
        throw "Expected stale ready proof before refresh, got status=$($stale.ready_pid_status) matches=$($stale.ready_pid_matches_live)."
    }

    Invoke-TraceDeckLoggedCommand -Label "Refresh isolated ready proof" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/refresh-backend-ready-proof.ps1 `
            -Addr $Addr `
            -PidPath $pidPath `
            -DataPath $dataPath `
            -ReadyPath $readyPath
    }

    $refreshed = Read-TraceDeckTaskStatus -ListenAddr $Addr
    if ($refreshed.ready_pid_status -ne "match" -or -not $refreshed.ready_pid_matches_live) {
        throw "Expected ready proof to match after refresh, got status=$($refreshed.ready_pid_status) matches=$($refreshed.ready_pid_matches_live)."
    }
    if ($refreshed.ready_pid -ne $refreshed.pid) {
        throw "Expected ready PID to equal live PID after refresh: ready=$($refreshed.ready_pid) live=$($refreshed.pid)."
    }
    if ($refreshed.ready.refreshed_by -ne "refresh-backend-ready-proof") {
        throw "Expected ready proof refreshed_by marker."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Ready proof refresh test passed addr=$Addr pid=$($refreshed.pid)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
