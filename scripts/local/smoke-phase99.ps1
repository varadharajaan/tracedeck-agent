param(
    [string]$Addr = "127.0.0.1:18258"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase99" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase99/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$readyPath = "$smokeRoot/backend-task-ready.json"
$taskStatusPath = "$smokeRoot/backend-task-status.json"

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

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 99 runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Phase99 Smoke Missing" `
            -PidPath $pidPath `
            -ReadyPath $readyPath `
            -TaskStatusOutputPath $taskStatusPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 99 verification evidence" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 `
            -Phase "phase99" `
            -BaseUrl $baseUrl
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Verification Evidence Center",
        "Scripted Gate Evidence",
        "Artifact Evidence",
        "Verification Proof",
        "Verification Actions",
        "verification-evidence-status",
        "verification-gate-list",
        "verification-action-list",
        "data-jump-target=`"verification-evidence-section`"",
        "Verification Evidence"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 99 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/verification-evidence-center"
    if (-not $center.evidence_available) {
        throw "Expected verification evidence artifact to be available."
    }
    if (@($center.gates).Count -lt 6) {
        throw "Expected verification evidence gate rows."
    }
    if (@($center.proof).Count -lt 4) {
        throw "Expected verification evidence proof rows."
    }
    if (@($center.actions).Count -lt 1) {
        throw "Expected verification evidence action rows."
    }
    foreach ($gate in @($center.gates)) {
        if ($gate.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only gate evidence scope."
        }
    }
    if ([string]::IsNullOrWhiteSpace($center.privacy_boundary) -or $center.privacy_boundary -notmatch "metadata-only") {
        throw "Expected verification evidence privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Verification evidence center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 99 dashboard layout" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 `
            -BaseUrl $baseUrl `
            -OutputRoot "data/local/dashboard-layout/phase99"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 99 dashboard theme" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 `
            -BaseUrl $baseUrl `
            -OutputRoot "data/local/dashboard-theme/phase99"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 99 dashboard visual quality" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 `
            -BaseUrl $baseUrl `
            -OutputRoot "data/local/dashboard-visual-quality/phase99"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 99 verification evidence smoke passed addr=$Addr gates=$(@($center.gates).Count) proof=$(@($center.proof).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
