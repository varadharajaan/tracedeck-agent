param(
    [string]$Addr = "127.0.0.1:18185"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase57" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase57/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase57-smoke/$timestamp"

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
    $tenantID = "family-varadha"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Customer Control Room",
        "Customer Value Tiles",
        "Anomaly Command Wall",
        "Mail And Push Delivery",
        "Owner Monetisation Actions",
        "data-jump-target=`"customer-control-section`"",
        "Notification Revenue Cockpit",
        "Package Billing Readiness",
        "Provider Simulation Lab"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 57 dashboard marker '$expected'."
        }
    }

    $room = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/customer-control-room"
    if ($room.summary.product_score -le 0 -or $room.summary.notification_score -le 0 -or $room.summary.package_score -le 0 -or [string]::IsNullOrWhiteSpace($room.summary.next_best_action)) {
        throw "Expected customer control scores and next action."
    }
    if ($room.summary.mail_delivered -le 0 -or $room.summary.dashboard_delivered -le 0) {
        throw "Expected mail and dashboard delivery proof."
    }
    if ($room.tiles.Count -lt 8 -or $room.alerts.Count -lt 1 -or $room.deliveries.Count -lt 3 -or $room.actions.Count -lt 1) {
        throw "Expected customer control tiles, alerts, deliveries, and actions."
    }
    $pushDelivery = $room.deliveries | Where-Object { $_.channel -eq "push" -and -not [string]::IsNullOrWhiteSpace($_.status) } | Select-Object -First 1
    if (-not $pushDelivery) {
        throw "Expected push notification delivery evidence."
    }
    if ($room.privacy_boundary -notmatch "metadata-only" -or $room.privacy_boundary -notmatch "no passwords" -or $room.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict customer control privacy boundary."
    }

    $serialized = ($room | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Customer control room leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 57 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 57 customer control smoke passed addr=$Addr product_score=$($room.summary.product_score) deliveries=$($room.deliveries.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
