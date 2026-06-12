param(
    [string]$Addr = "127.0.0.1:18173"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase51" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase51/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase51-smoke/$timestamp"

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
        "Business Dashboard",
        "Notification Evidence Timeline",
        "Delivery Audit Trail",
        "Anomaly Notification Inbox",
        "Push And Mail Proof",
        "Paid Package Value",
        "Customer Owner Actions",
        "business-dashboard-section",
        "delivery-timeline-section",
        "delivery-timeline-list",
        "business-alert-list",
        "business-channel-list",
        "business-package-list",
        "business-action-list"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 50 dashboard marker '$expected'."
        }
    }

    $business = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/business-dashboard"
    if ($business.summary.product_score -le 0 -or $business.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($business.summary.recommended_package)) {
        throw "Expected business dashboard to expose product score, notification score, and recommended package."
    }
    if ($business.summary.mail_delivered -lt 1 -or $business.summary.dashboard_delivered -lt 1) {
        throw "Expected business dashboard mail and dashboard delivery proof."
    }
    if ($business.metrics.Count -lt 8 -or $business.alerts.Count -lt 1 -or $business.channels.Count -lt 3 -or $business.packages.Count -lt 3 -or $business.actions.Count -lt 1) {
        throw "Expected business dashboard metrics, alerts, channels, packages, and actions."
    }
    $pushRoute = $business.channels | Where-Object { $_.channel -eq "push" -and -not [string]::IsNullOrWhiteSpace($_.status) } | Select-Object -First 1
    if ($null -eq $pushRoute) {
        throw "Expected business dashboard push route proof."
    }

    $timeline = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/delivery-timeline?limit=8"
    if ($timeline.summary.total -lt 3 -or $timeline.summary.email -lt 1 -or $timeline.summary.push -lt 1 -or $timeline.summary.dashboard -lt 1) {
        throw "Expected delivery timeline to expose email, push, and dashboard evidence."
    }
    if ($timeline.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($timeline.summary.recommended_paid_tier)) {
        throw "Expected delivery timeline notification score and paid tier."
    }
    $filtered = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/delivery-timeline?channel=email&status=delivered&provider=smtp&limit=2"
    if ($filtered.items.Count -lt 1 -or $filtered.items[0].channel -ne "email" -or $filtered.items[0].status -ne "delivered") {
        throw "Expected filtered delivery timeline to expose delivered email evidence."
    }
    if ($timeline.privacy_boundary -notmatch "metadata-only" -or $timeline.privacy_boundary -notmatch "no passwords") {
        throw "Expected strict delivery timeline privacy boundary."
    }
    $timelineSerialized = ($timeline | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body")) {
        if ($timelineSerialized.Contains($forbidden)) {
            throw "Delivery timeline leaked forbidden field marker '$forbidden'."
        }
    }

    if ($business.privacy_boundary -notmatch "no passwords" -or $business.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict business dashboard privacy boundary."
    }
    $serialized = ($business | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Business dashboard leaked forbidden field marker '$forbidden'."
        }
    }

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot

    Write-TraceDeckLog -Level "INFO" -Message "Phase 51 delivery timeline smoke passed addr=$Addr score=$($timeline.summary.notification_score) events=$($timeline.summary.total)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
