param(
    [string]$Addr = "127.0.0.1:18267"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase105" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase105/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$readyPath = "$smokeRoot/backend-task-ready.json"
$taskStatusPath = "$smokeRoot/backend-task-status.json"
$assurancePath = "$smokeRoot/operator-assurance.json"
$assuranceTextPath = "$smokeRoot/operator-assurance.txt"
$promotionPath = "$smokeRoot/promotion-readiness.json"
$promotionTextPath = "$smokeRoot/promotion-readiness.txt"

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

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 105 runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Phase105 Smoke Missing" `
            -PidPath $pidPath `
            -ReadyPath $readyPath `
            -TaskStatusOutputPath $taskStatusPath
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 105 verification evidence" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 `
            -Phase "phase105" `
            -BaseUrl $baseUrl
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 105 operator assurance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-operator-assurance.ps1 `
            -BaseUrl $baseUrl `
            -OutputPath $assurancePath `
            -TextOutputPath $assuranceTextPath
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 105 promotion readiness" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-promotion-readiness.ps1 `
            -BaseUrl $baseUrl `
            -OutputPath $promotionPath `
            -TextOutputPath $promotionTextPath
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Promotion Readiness Center",
        "Promotion Proof",
        "Promotion Actions",
        "promotion-readiness-status",
        "promotion-proof-list",
        "promotion-action-list",
        "data-jump-target=`"promotion-readiness-section`"",
        "Promotion Readiness"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 105 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/promotion-readiness-center"
    if ($center.source -ne "phase105_promotion_readiness") {
        throw "Expected Phase 105 promotion readiness source."
    }
    if ([string]::IsNullOrWhiteSpace($center.summary.status)) {
        throw "Expected promotion readiness summary status."
    }
    if (@($center.proof).Count -lt 6) {
        throw "Expected promotion readiness proof rows."
    }
    if (@($center.actions).Count -lt 1) {
        throw "Expected promotion readiness actions."
    }
    foreach ($proof in @($center.proof)) {
        if ($proof.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only promotion proof evidence scope."
        }
    }
    foreach ($action in @($center.actions)) {
        if ($action.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only promotion action evidence scope."
        }
    }
    if ([string]::IsNullOrWhiteSpace($center.privacy_boundary) -or $center.privacy_boundary -notmatch "metadata-only") {
        throw "Expected promotion readiness privacy boundary."
    }
    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $promotionPath))) {
        throw "Expected promotion readiness JSON export at $promotionPath"
    }
    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $promotionTextPath))) {
        throw "Expected promotion readiness text export at $promotionTextPath"
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Promotion readiness center leaked forbidden field marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 105 promotion readiness smoke passed addr=$Addr status=$($center.summary.status) proof=$(@($center.proof).Count) actions=$(@($center.actions).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
