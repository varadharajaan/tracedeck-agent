param(
    [string]$Addr = "127.0.0.1:18265"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase103" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase103/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$readyPath = "$smokeRoot/backend-task-ready.json"
$taskStatusPath = "$smokeRoot/backend-task-status.json"
$assurancePath = "$smokeRoot/operator-assurance.json"
$assuranceTextPath = "$smokeRoot/operator-assurance.txt"
$refreshCommand = "python ./devctl.py server task-refresh-ready"

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

function Update-TraceDeckRuntimeArtifacts {
    param([string]$ListenAddr, [string]$BaseUrl)

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
        -Addr $ListenAddr `
        -TaskName "\TraceDeck\TraceDeck Phase103 Smoke Missing" `
        -PidPath $pidPath `
        -ReadyPath $readyPath `
        -TaskStatusOutputPath $taskStatusPath | Out-Null
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 `
        -Phase "phase103" `
        -BaseUrl $BaseUrl | Out-Null
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-operator-assurance.ps1 `
        -BaseUrl $BaseUrl `
        -OutputPath $assurancePath `
        -TextOutputPath $assuranceTextPath | Out-Null
}

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    Write-TraceDeckStaleReadyFile -RelativePidPath $pidPath -RelativeReadyPath $readyPath -RelativeDataPath $dataPath -ListenAddr $Addr

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 103 stale runtime artifacts" -Command {
        Update-TraceDeckRuntimeArtifacts -ListenAddr $Addr -BaseUrl $baseUrl
    }

    $staleCenter = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/runtime-status-center"
    $staleAction = @($staleCenter.actions) | Where-Object { $_.id -eq "refresh-ready-pid-proof" } | Select-Object -First 1
    if ($staleCenter.summary.ready_pid_status -ne "stale" -or $null -eq $staleAction -or $staleAction.command -ne $refreshCommand) {
        throw "Expected stale ready PID action command '$refreshCommand'."
    }
    $staleAssurance = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/operator-assurance-center"
    $staleAssuranceAction = @($staleAssurance.actions) | Where-Object { $_.id -eq "refresh-ready-pid-proof" } | Select-Object -First 1
    if ($null -eq $staleAssuranceAction -or $staleAssuranceAction.command -ne $refreshCommand) {
        throw "Expected operator assurance stale ready PID command '$refreshCommand'."
    }

    Invoke-TraceDeckLoggedCommand -Label "Refresh Phase 103 ready PID proof" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/refresh-backend-ready-proof.ps1 `
            -Addr $Addr `
            -PidPath $pidPath `
            -DataPath $dataPath `
            -ReadyPath $readyPath
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 103 refreshed runtime artifacts" -Command {
        Update-TraceDeckRuntimeArtifacts -ListenAddr $Addr -BaseUrl $baseUrl
    }

    $refreshedCenter = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/runtime-status-center"
    if ($refreshedCenter.summary.ready_pid_status -ne "match" -or -not $refreshedCenter.summary.ready_pid_matches_live) {
        throw "Expected refreshed ready PID to match live PID."
    }
    $refreshedAction = @($refreshedCenter.actions) | Where-Object { $_.id -eq "refresh-ready-pid-proof" } | Select-Object -First 1
    if ($null -ne $refreshedAction) {
        throw "Did not expect refresh-ready-pid-proof action after ready PID matched live PID."
    }

    $serialized = (($staleCenter | ConvertTo-Json -Depth 24) + ($refreshedCenter | ConvertTo-Json -Depth 24)).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Phase 103 ready refresh contract leaked forbidden field marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 103 ready proof refresh smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
