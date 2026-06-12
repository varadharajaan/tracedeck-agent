param(
    [string]$Addr = "127.0.0.1:18199"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase65" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase65/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase65-smoke/$timestamp"

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
        "Revenue Operations Center",
        "Revenue Signals",
        "Anomaly And Delivery Wall",
        "Mail, Push, Dashboard Proof",
        "Commercial Levers",
        "Revenue Owner Actions",
        "data-jump-target=`"revenue-operations-section`"",
        "revenue-kpi-mail",
        "revenue-kpi-push"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 65 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/revenue-operations-center"
    if ($center.summary.revenue_score -le 0 -or $center.summary.product_score -le 0 -or $center.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($center.summary.owner_next_step)) {
        throw "Expected revenue operations score, product score, notification score, and owner next step."
    }
    if ($center.summary.hosts_total -lt 1 -or $center.summary.mail_delivered -lt 1 -or $center.summary.dashboard_delivered -lt 1 -or [string]::IsNullOrWhiteSpace($center.summary.recommended_paid_package)) {
        throw "Expected host, mail, dashboard, and paid package proof."
    }
    if (@($center.signals).Count -lt 9 -or @($center.alerts).Count -lt 1 -or @($center.deliveries).Count -lt 3 -or @($center.levers).Count -lt 6 -or @($center.actions).Count -lt 1) {
        throw "Expected revenue signals, alerts, deliveries, levers, and actions."
    }
    foreach ($signalID in @("anomaly-command", "mail-delivery", "push-reach", "archive-retention", "customer-settings", "provider-simulation")) {
        if (@($center.signals | Where-Object { $_.id -eq $signalID }).Count -ne 1) {
            throw "Expected revenue signal '$signalID'."
        }
    }
    foreach ($channel in @("email", "push", "dashboard")) {
        if (@($center.deliveries | Where-Object { $_.channel -eq $channel }).Count -ne 1) {
            throw "Expected revenue delivery proof for '$channel'."
        }
    }
    if ($center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no passwords" -or $center.privacy_boundary -notmatch "no screenshots" -or $center.privacy_boundary -notmatch "push endpoints") {
        throw "Expected strict revenue operations privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Revenue operations center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 65 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 65 revenue operations smoke passed addr=$Addr score=$($center.summary.revenue_score) signals=$(@($center.signals).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
