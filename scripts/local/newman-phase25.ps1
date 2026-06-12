param(
    [string]$Addr = "127.0.0.1:18116"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "newman-phase25" -LogRoot "logs/local/newman" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRoot = "data/local/newman/phase25/$timestamp"
$pidPath = "$runRoot/tracedeck-backend.pid"
$dataPath = "$runRoot/backend-state.json"

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

function Wait-TraceDeckMonetizationSeed {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        try {
            $summary = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/api/v1/tenants/family-varadha/monetization-summary"
            if ($summary.readiness_score -gt 0 -and $summary.notification_routes.Count -eq 3) {
                return
            }
        }
        catch { Start-Sleep -Milliseconds 500 }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo seed did not expose monetization summary at $BaseUrl"
}

try {
    $newman = Get-Command newman -ErrorAction SilentlyContinue
    if (-not $newman) { throw "newman is not installed or not on PATH" }

    $reportRoot = Join-Path $script:TraceDeckRepoRoot $runRoot
    New-Item -ItemType Directory -Force -Path $reportRoot | Out-Null
    $reportPath = Join-Path $reportRoot "newman-report.json"
    $baseUrl = "http://$Addr"

    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    Wait-TraceDeckMonetizationSeed -BaseUrl $baseUrl

    Invoke-TraceDeckLoggedCommand -Label "Run Newman Phase 25 collection" -Command {
        newman run ./postman/tracedeck-backend-phase25.postman_collection.json --env-var "baseUrl=$baseUrl" --reporters "cli,json" --reporter-json-export $reportPath
    }

    if (-not (Test-Path $reportPath)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected Newman report was not created: $reportPath"
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Newman Phase 25 report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
