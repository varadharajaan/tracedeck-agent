param(
    [string]$Addr = "127.0.0.1:18200"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "newman-phase64" -LogRoot "logs/local/newman" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRoot = "data/local/newman/phase64/$timestamp"
$pidPath = "$runRoot/tracedeck-backend.pid"
$dataPath = "$runRoot/backend-state.json"
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
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    Invoke-TraceDeckLoggedCommand -Label "Run Newman Phase 64 collection" -Command {
        newman run ./postman/tracedeck-backend-phase64.postman_collection.json --env-var "baseUrl=$baseUrl" --reporters "cli,json" --reporter-json-export $reportPath
    }

    if (!(Test-Path $reportPath)) {
        throw "Expected Newman report at $reportPath"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Newman Phase 64 report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
