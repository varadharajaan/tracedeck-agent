param(
    [string]$Addr = "127.0.0.1:18262"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase102" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase102/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$readyPath = "$smokeRoot/backend-task-ready.json"
$taskStatusPath = "$smokeRoot/backend-task-status.json"
$assurancePath = "$smokeRoot/operator-assurance.json"
$assuranceTextPath = "$smokeRoot/operator-assurance.txt"

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
    $readyDir = Split-Path -Parent $readyFullPath
    New-Item -ItemType Directory -Force -Path $readyDir | Out-Null

    $ready = [ordered]@{
        addr = $ListenAddr
        base_url = "http://$ListenAddr"
        pid = $livePid + 100000
        ready_at = (Get-Date).AddMinutes(-10).ToString("o")
        pid_path = (Join-Path $script:TraceDeckRepoRoot $RelativePidPath)
        data_path = (Join-Path $script:TraceDeckRepoRoot $RelativeDataPath)
        stdout = "logs/local/backend/phase102-stale-ready.out.log"
        stderr = "logs/local/backend/phase102-stale-ready.err.log"
    }
    $ready | ConvertTo-Json -Depth 6 | Set-Content -Path $readyFullPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Wrote stale ready PID proof live_pid=$livePid ready_pid=$($ready.pid) ready_path=$RelativeReadyPath"
}

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath
    Write-TraceDeckStaleReadyFile -RelativePidPath $pidPath -RelativeReadyPath $readyPath -RelativeDataPath $dataPath -ListenAddr $Addr

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 102 runtime summary with stale ready PID" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Phase102 Smoke Missing" `
            -PidPath $pidPath `
            -ReadyPath $readyPath `
            -TaskStatusOutputPath $taskStatusPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 102 verification evidence" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 `
            -Phase "phase102" `
            -BaseUrl $baseUrl
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 102 operator assurance pack" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-operator-assurance.ps1 `
            -BaseUrl $baseUrl `
            -OutputPath $assurancePath `
            -TextOutputPath $assuranceTextPath
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/runtime-status-center"
    if (-not $center.summary_available) {
        throw "Expected runtime status summary to be available."
    }
    if (-not $center.summary.runtime_ok -or -not $center.summary.health_ok) {
        throw "Expected live backend PID and health proof to stay healthy."
    }
    if ($center.summary.ready_pid_status -ne "stale" -or $center.summary.ready_pid_matches_live) {
        throw "Expected stale ready PID summary, got status=$($center.summary.ready_pid_status) matches=$($center.summary.ready_pid_matches_live)."
    }
    if ($center.summary.status -ne "watch") {
        throw "Expected stale ready PID to produce watch summary status."
    }

    $pidProof = @($center.proof) | Where-Object { $_.id -eq "pid-reconciliation" } | Select-Object -First 1
    if ($null -eq $pidProof -or $pidProof.value -ne "stale" -or $pidProof.status -ne "watch") {
        throw "Expected PID reconciliation proof row to show stale/watch."
    }
    $refreshAction = @($center.actions) | Where-Object { $_.id -eq "refresh-ready-pid-proof" } | Select-Object -First 1
    if ($null -eq $refreshAction -or $refreshAction.status -ne "watch") {
        throw "Expected refresh-ready-pid-proof watch action."
    }
    foreach ($proof in @($center.proof)) {
        if ($proof.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only runtime proof scope."
        }
    }
    if ([string]::IsNullOrWhiteSpace($center.privacy_boundary) -or $center.privacy_boundary -notmatch "metadata-only") {
        throw "Expected runtime status privacy boundary."
    }

    $assurance = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/operator-assurance-center"
    $assuranceRefreshAction = @($assurance.actions) | Where-Object { $_.id -eq "refresh-ready-pid-proof" } | Select-Object -First 1
    if ($null -eq $assuranceRefreshAction -or $assuranceRefreshAction.status -ne "watch") {
        throw "Expected operator assurance refresh-ready-pid-proof action."
    }
    if (-not (($assurance.cards | ConvertTo-Json -Depth 12) -match "ready_pid=stale")) {
        throw "Expected operator assurance runtime card to include ready_pid=stale."
    }
    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $assurancePath))) {
        throw "Expected operator assurance JSON export at $assurancePath"
    }
    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $assuranceTextPath))) {
        throw "Expected operator assurance text export at $assuranceTextPath"
    }

    $serialized = (($center | ConvertTo-Json -Depth 24) + ($assurance | ConvertTo-Json -Depth 24)).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Phase 102 stale PID contract leaked forbidden field marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 102 stale ready PID smoke passed addr=$Addr live_pid=$($center.proof[0].detail) proof=$(@($center.proof).Count) actions=$(@($center.actions).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
