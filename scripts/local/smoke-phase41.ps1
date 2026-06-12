param(
    [string]$Addr = "127.0.0.1:18153"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase41" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase41/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"

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
        "Backend Alert Inbox",
        "tenant-alert-inbox-status",
        "tenant-alert-inbox-list",
        "tenant-alert-inbox-ready",
        "Paid Ops Console",
        "Commercial Control Room",
        "Revenue Command Center",
        "Notification Proof Rail",
        "Buyer Demo Checklist",
        "Mail Delivery Center",
        "Push Notification Center",
        "Archive Retention",
        "Tamper Trust"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 41 monetisation dashboard marker '$expected'."
        }
    }

    $tenantID = "family-varadha"
    $inbox = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/alert-inbox"
    if ($inbox.summary.total -lt 1 -or $inbox.summary.with_email -lt 1 -or $inbox.summary.with_push -lt 1 -or $inbox.summary.with_dashboard -lt 1) {
        throw "Expected alert inbox summary to include email, push, and dashboard proof."
    }
    if (-not ($inbox.privacy_boundary -match "no passwords") -or -not ($inbox.privacy_boundary -match "screenshots")) {
        throw "Expected alert inbox privacy boundary to deny sensitive collection."
    }
    $linked = @($inbox.items | Where-Object { $_.event_id -and $_.delivery_proof.Count -gt 0 })
    if ($linked.Count -lt 1) {
        throw "Expected alert inbox items linked to delivery proof."
    }

    $feed = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/activity-feed?kind=delivery&limit=5"
    if ($feed.summary.delivery_items -lt 1) {
        throw "Expected activity feed delivery proof to remain available."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 41 alert inbox smoke passed addr=$Addr alerts=$($inbox.summary.total) ready=$($inbox.summary.notification_ready)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
