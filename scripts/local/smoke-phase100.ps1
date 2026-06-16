param(
    [string]$Addr = "127.0.0.1:18260"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase100" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase100/$timestamp"
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

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 100 runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Phase100 Smoke Missing" `
            -PidPath $pidPath `
            -ReadyPath $readyPath `
            -TaskStatusOutputPath $taskStatusPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 100 verification evidence" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 `
            -Phase "phase100" `
            -BaseUrl $baseUrl
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 100 operator assurance pack" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-operator-assurance.ps1 `
            -BaseUrl $baseUrl `
            -OutputPath $assurancePath `
            -TextOutputPath $assuranceTextPath
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Operator Assurance Center",
        "Assurance Cards",
        "Assurance Actions",
        "operator-assurance-status",
        "assurance-card-list",
        "assurance-action-list",
        "data-jump-target=`"operator-assurance-section`"",
        "Operator Assurance"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 100 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/operator-assurance-center"
    if ($center.source -ne "phase100_operator_assurance") {
        throw "Expected Phase 100 operator assurance source."
    }
    if (@($center.cards).Count -lt 6) {
        throw "Expected operator assurance cards."
    }
    if (@($center.actions).Count -lt 1) {
        throw "Expected operator assurance actions."
    }
    foreach ($card in @($center.cards)) {
        if ($card.evidence_scope -ne "metadata_only") {
            throw "Expected metadata_only assurance card evidence scope."
        }
    }
    if ([string]::IsNullOrWhiteSpace($center.privacy_boundary) -or $center.privacy_boundary -notmatch "metadata-only") {
        throw "Expected operator assurance privacy boundary."
    }
    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $assurancePath))) {
        throw "Expected operator assurance JSON export at $assurancePath"
    }
    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $assuranceTextPath))) {
        throw "Expected operator assurance text export at $assuranceTextPath"
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Operator assurance center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 100 dashboard layout" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 `
            -BaseUrl $baseUrl `
            -OutputRoot "data/local/dashboard-layout/phase100"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 100 dashboard theme" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 `
            -BaseUrl $baseUrl `
            -OutputRoot "data/local/dashboard-theme/phase100"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 100 dashboard visual quality" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 `
            -BaseUrl $baseUrl `
            -OutputRoot "data/local/dashboard-visual-quality/phase100"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 100 operator assurance smoke passed addr=$Addr cards=$(@($center.cards).Count) actions=$(@($center.actions).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
