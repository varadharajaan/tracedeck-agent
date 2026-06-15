param(
    [string]$Addr = "127.0.0.1:18247",
    [string]$TenantID = "family-varadha",
    [switch]$SkipServerStart
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-devctl-runtime-doctor" -LogRoot "logs/local/test" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRoot = "data/local/devctl-runtime-doctor/$timestamp"
$pidPath = "$runRoot/tracedeck-backend.pid"
$dataPath = "$runRoot/backend-state.json"
$doctorReportPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-doctor.json"

function Start-TraceDeckDoctorBackend {
    param([string]$ListenAddr, [string]$RelativePidPath, [string]$RelativeDataPath)

    Write-TraceDeckLog -Level "INFO" -Message "Starting isolated dashboard demo backend addr=$ListenAddr"
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
                    Write-TraceDeckLog -Level "INFO" -Message "Isolated dashboard demo backend ready addr=$ListenAddr helper_pid=$($helper.Id)"
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
    throw "Dashboard demo helper did not become ready at $baseUrl"
}

try {
    if (-not $SkipServerStart) {
        Start-TraceDeckDoctorBackend -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Run devctl runtime doctor" -Command {
        python ./devctl.py --addr $Addr doctor --tenant-id $TenantID --skip-cloud
    }

    if (-not (Test-Path $doctorReportPath)) {
        throw "Expected runtime doctor report was not created: $doctorReportPath"
    }

    $report = Get-Content -Path $doctorReportPath -Raw | ConvertFrom-Json
    if ($report.overall -ne "ok" -or $report.local.overall -ne "ok") {
        throw "Expected runtime doctor overall ok; got overall=$($report.overall) local=$($report.local.overall)"
    }
    if (-not $report.local.deliveries.ok) {
        throw "Expected delivery doctor check to pass."
    }
    if (-not $report.local.deliveries.default_demo_hidden) {
        throw "Expected default alert deliveries to hide demo evidence."
    }
    if ([int]$report.local.deliveries.default_count -ne 0) {
        throw "Expected isolated default alert delivery count to be zero; got $($report.local.deliveries.default_count)"
    }
    if (-not $report.local.deliveries.opt_in_demo_available -or [int]$report.local.deliveries.opt_in_demo_count -lt 1) {
        throw "Expected opt-in demo delivery evidence to remain available and labelled."
    }
    if ($report.local.delivery_assurance.buyer_ready -and [int]$report.local.delivery_assurance.provider_confirmed -eq 0) {
        throw "Runtime doctor must not mark buyer-ready without provider-confirmed delivery proof."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Devctl runtime doctor provenance passed report=$doctorReportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if (-not $SkipServerStart) {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
    }
}
