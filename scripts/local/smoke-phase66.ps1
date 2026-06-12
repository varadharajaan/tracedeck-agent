param(
    [string]$Addr = "127.0.0.1:18211"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase66" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase66/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase66-smoke/$timestamp"

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
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Deployment Readiness Center",
        "Platform Service Proof",
        "Service Manifest Proof",
        "Boot And Replay Proof",
        "Deployment Owner Actions",
        "data-jump-target=`"deployment-readiness-section`"",
        "deployment-kpi-liveboot",
        "deployment-kpi-autostart"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 66 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/deployment-readiness-center"
    if ($center.summary.readiness_score -le 0 -or [string]::IsNullOrWhiteSpace($center.summary.owner_next_step)) {
        throw "Expected deployment readiness score and owner next step."
    }
    if ($center.summary.platforms_total -ne 3 -or $center.summary.manifests_total -ne 3 -or $center.summary.hosts_total -lt 1) {
        throw "Expected three platforms, three manifests, and host proof."
    }
    if ($null -eq $center.summary.live_boot_ready -or $null -eq $center.summary.offline_replay_ready) {
        throw "Expected live boot and offline replay readiness fields."
    }
    foreach ($proofID in @("live-boot", "offline-replay", "windows-task-scheduler", "macos-linux-services")) {
        if (@($center.proof | Where-Object { $_.id -eq $proofID }).Count -ne 1) {
            throw "Expected deployment proof '$proofID'."
        }
    }
    if (@($center.platforms).Count -ne 3 -or @($center.manifests).Count -ne 3 -or @($center.proof).Count -lt 5 -or @($center.actions).Count -lt 4) {
        throw "Expected deployment platforms, manifests, proof, and actions."
    }
    foreach ($platform in @("windows", "darwin", "linux")) {
        if (@($center.platforms | Where-Object { $_.platform -eq $platform }).Count -ne 1) {
            throw "Expected deployment platform '$platform'."
        }
    }
    foreach ($manager in @("task_scheduler", "launchd", "systemd")) {
        if (@($center.platforms | Where-Object { $_.service_manager -eq $manager }).Count -ne 1) {
            throw "Expected deployment service manager '$manager'."
        }
    }
    foreach ($manifestID in @("windows-task", "macos-launchd", "linux-systemd")) {
        if (@($center.manifests | Where-Object { $_.id -eq $manifestID }).Count -ne 1) {
            throw "Expected deployment manifest '$manifestID'."
        }
    }
    if ($center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no passwords" -or $center.privacy_boundary -notmatch "no screenshots" -or $center.privacy_boundary -notmatch "hidden collection bypasses") {
        throw "Expected strict deployment readiness privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Deployment readiness center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 66 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 66 deployment readiness smoke passed addr=$Addr score=$($center.summary.readiness_score) platforms=$(@($center.platforms).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
