param(
    [string]$Addr = "127.0.0.1:18266"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "newman-phase103" -LogRoot "logs/local/newman" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRoot = "data/local/newman/phase103/$timestamp"
$pidPath = "$runRoot/tracedeck-backend.pid"
$dataPath = "$runRoot/backend-state.json"
$readyPath = "$runRoot/backend-task-ready.json"
$taskStatusPath = "$runRoot/backend-task-status.json"
$assurancePath = "$runRoot/operator-assurance.json"
$assuranceTextPath = "$runRoot/operator-assurance.txt"
$reportPath = "$runRoot/newman-report.json"

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
        stdout = "logs/local/backend/phase103-stale-ready.out.log"
        stderr = "logs/local/backend/phase103-stale-ready.err.log"
    }
    $ready | ConvertTo-Json -Depth 6 | Set-Content -Path $readyFullPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Wrote stale ready PID proof live_pid=$livePid ready_pid=$($ready.pid) ready_path=$RelativeReadyPath"
}

try {
    $newman = Get-Command newman -ErrorAction SilentlyContinue
    if (-not $newman) {
        throw "newman is not installed or not on PATH"
    }

    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    Write-TraceDeckStaleReadyFile -RelativePidPath $pidPath -RelativeReadyPath $readyPath -RelativeDataPath $dataPath -ListenAddr $Addr

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 103 Newman runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Phase103 Newman Missing" `
            -PidPath $pidPath `
            -ReadyPath $readyPath `
            -TaskStatusOutputPath $taskStatusPath
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 103 Newman verification evidence" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 `
            -Phase "phase103" `
            -BaseUrl $baseUrl
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 103 Newman operator assurance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-operator-assurance.ps1 `
            -BaseUrl $baseUrl `
            -OutputPath $assurancePath `
            -TextOutputPath $assuranceTextPath
    }

    $reportDir = Split-Path -Parent (Join-Path $script:TraceDeckRepoRoot $reportPath)
    New-Item -ItemType Directory -Force -Path $reportDir | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Run Newman Phase 103 collection" -Command {
        newman run ./postman/tracedeck-backend-phase103.postman_collection.json --env-var "baseUrl=$baseUrl" --reporters "cli,json" --reporter-json-export $reportPath
    }

    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $reportPath))) {
        throw "Expected Newman report was not created: $reportPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Newman Phase 103 report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
