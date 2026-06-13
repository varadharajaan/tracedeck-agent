param(
    [string]$Addr = "127.0.0.1:18216"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase68" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase68/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase68-smoke/$timestamp"

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
        "browser-activity-button",
        "/browser-activity",
        "Open Chrome, Brave, and Edge browser activity"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 68 dashboard marker '$expected'."
        }
    }

    $browserPage = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/browser-activity"
    foreach ($expected in @(
        "TraceDeck Browser Activity",
        "Browser Activity Viewer",
        "Non-Study YouTube",
        "Notification Proof",
        "Host Breakdown",
        "Browser Domain Activity"
    )) {
        if ($browserPage.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 68 browser page marker '$expected'."
        }
    }

    $viewer = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/browser-activity?limit=25"
    if ($viewer.summary.total -lt 3 -or $viewer.summary.chrome -lt 1 -or $viewer.summary.edge -lt 1 -or $viewer.summary.brave -lt 1) {
        throw "Expected Chrome, Edge, and Brave browser activity rows."
    }
    if ($viewer.summary.study_safe -lt 1 -or $viewer.summary.non_study_youtube -lt 1 -or $viewer.summary.notification_proof -lt 1) {
        throw "Expected study-safe, non-study YouTube, and notification proof counts."
    }
    if (@($viewer.hosts).Count -lt 1 -or @($viewer.browsers).Count -lt 3 -or @($viewer.items).Count -lt 3) {
        throw "Expected browser activity hosts, browsers, and item rows."
    }
    if ($viewer.privacy_boundary -notmatch "metadata-only" -or $viewer.privacy_boundary -notmatch "no passwords" -or $viewer.privacy_boundary -notmatch "raw URLs" -or $viewer.privacy_boundary -notmatch "page titles") {
        throw "Expected strict browser activity privacy boundary."
    }

    $filtered = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/browser-activity?browser=edge&study_safe=false&limit=10"
    if ($filtered.filters.browser -ne "edge" -or @($filtered.items).Count -lt 1) {
        throw "Expected filtered Edge browser activity."
    }

    $serialized = ($viewer | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "cookie_value", "token_value", "password_value", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Browser activity viewer leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 68 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 68 browser activity smoke passed addr=$Addr rows=$($viewer.summary.total)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
