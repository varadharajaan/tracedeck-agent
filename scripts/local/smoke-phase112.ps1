param(
    [string]$Addr = "127.0.0.1:18276"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase112" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase112/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$readyPath = "$smokeRoot/backend-task-ready.json"
$taskStatusPath = "$smokeRoot/backend-task-status.json"
$indicatorPath = "$smokeRoot/local-monitoring-indicator.json"
$indicatorTextPath = "$smokeRoot/local-monitoring-indicator.txt"
$indicatorHtmlPath = "$smokeRoot/local-monitoring-indicator.html"

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

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 112 runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Phase112 Smoke" `
            -PidPath $pidPath `
            -ReadyPath $readyPath `
            -TaskStatusOutputPath $taskStatusPath `
            -SkipDoctor
    }
    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 112 local monitoring indicator" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-local-monitoring-indicator.ps1 `
            -BaseUrl $baseUrl `
            -OutputPath $indicatorPath `
            -TextOutputPath $indicatorTextPath `
            -HtmlOutputPath $indicatorHtmlPath
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Local Monitoring Indicator",
        "Indicator Proof",
        "Indicator Actions",
        "local-indicator-status",
        "indicator-proof-list",
        "indicator-action-list",
        "data-jump-target=`"local-indicator-section`"",
        "Local Indicator"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 112 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/local-monitoring-indicator"
    if ($center.source -ne "phase112_local_monitoring_indicator") {
        throw "Expected Phase 112 local indicator source."
    }
    if ($center.summary.status -notin @("ok", "watch")) {
        throw "Expected local indicator status ok/watch, got $($center.summary.status)."
    }
    if (-not $center.summary.visible_indicator_ready -or -not $center.summary.local_status_page_ready) {
        throw "Expected visible local status page readiness."
    }
    if (-not $center.summary.consent_visible -or -not $center.summary.sensitive_collection_denied) {
        throw "Expected consent visibility and sensitive collection deny proof."
    }
    if (@($center.proof).Count -lt 6) {
        throw "Expected local indicator proof rows."
    }
    if (@($center.actions).Count -lt 3) {
        throw "Expected local indicator actions."
    }
    foreach ($proof in @($center.proof)) {
        if ($proof.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only indicator proof evidence scope."
        }
    }
    foreach ($action in @($center.actions)) {
        if ($action.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only indicator action evidence scope."
        }
    }
    if ([string]::IsNullOrWhiteSpace($center.privacy_boundary) -or $center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no passwords" -or $center.privacy_boundary -notmatch "screenshots") {
        throw "Expected local indicator privacy boundary."
    }
    foreach ($path in @($indicatorPath, $indicatorTextPath, $indicatorHtmlPath)) {
        if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $path))) {
            throw "Expected local indicator artifact at $path"
        }
    }
    $html = Get-Content -Path (Join-Path $script:TraceDeckRepoRoot $indicatorHtmlPath) -Raw
    foreach ($expected in @("TraceDeck Local Monitoring Indicator", "Indicator Proof", "Privacy Boundary")) {
        if ($html -notmatch [regex]::Escape($expected)) {
            throw "Expected local indicator HTML marker '$expected'."
        }
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Local monitoring indicator leaked forbidden field marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 112 local monitoring indicator smoke passed addr=$Addr status=$($center.summary.status) proof=$(@($center.proof).Count) actions=$(@($center.actions).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
