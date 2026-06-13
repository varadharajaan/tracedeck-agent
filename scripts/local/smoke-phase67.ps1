param(
    [string]$Addr = "127.0.0.1:18214"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase67" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase67/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase67-smoke/$timestamp"

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
        "Premium Operations Hub",
        "Premium Value Tiles",
        "Anomaly Notification Wall",
        "Mail And Push Delivery Ops",
        "Premium Owner Actions",
        "data-jump-target=`"premium-operations-section`"",
        "premium-kpi-push",
        "premium-delivery-list"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 67 dashboard marker '$expected'."
        }
    }

    $hub = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/premium-operations-hub"
    if ($hub.summary.premium_score -le 0 -or [string]::IsNullOrWhiteSpace($hub.summary.owner_next_step)) {
        throw "Expected premium operations score and owner next step."
    }
    if ($hub.summary.mail_delivered -lt 1 -or $hub.summary.dashboard_delivered -lt 1) {
        throw "Expected mail and dashboard delivery proof."
    }
    foreach ($tileID in @("anomaly-inbox", "mail-delivery", "push-notifications", "dashboard-fallback", "weekly-report", "archive-retention", "deployment-readiness", "package-value")) {
        if (@($hub.tiles | Where-Object { $_.id -eq $tileID }).Count -ne 1) {
            throw "Expected premium tile '$tileID'."
        }
    }
    if (@($hub.alerts).Count -lt 1 -or @($hub.deliveries).Count -lt 3 -or @($hub.actions).Count -lt 1) {
        throw "Expected premium alerts, deliveries, and actions."
    }
    if ($hub.privacy_boundary -notmatch "metadata-only" -or $hub.privacy_boundary -notmatch "no passwords" -or $hub.privacy_boundary -notmatch "no screenshots" -or $hub.privacy_boundary -notmatch "hidden collection bypasses") {
        throw "Expected strict premium operations privacy boundary."
    }

    $serialized = ($hub | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token", "keylogger")) {
        if ($serialized.Contains($forbidden)) {
            throw "Premium operations hub leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 67 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 67 premium operations smoke passed addr=$Addr score=$($hub.summary.premium_score)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
