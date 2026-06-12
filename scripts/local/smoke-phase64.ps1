param(
    [string]$Addr = "127.0.0.1:18199"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase64" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase64/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase64-smoke/$timestamp"

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
        "Customer Settings Center",
        "Settings Matrix",
        "Plan And Retention Options",
        "Notification Channel Settings",
        "Settings Owner Actions",
        "data-jump-target=`"customer-settings-section`"",
        "settings-kpi-notifications",
        "settings-kpi-data-rights"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 64 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/customer-settings-center"
    if ($center.summary.settings_score -le 0 -or $center.summary.settings_total -lt 8 -or [string]::IsNullOrWhiteSpace($center.summary.owner_next_step)) {
        throw "Expected customer settings score, settings total, and owner next step."
    }
    if (@($center.settings).Count -lt 8 -or @($center.plan_options).Count -lt 4 -or @($center.retention_options).Count -lt 3 -or @($center.channels).Count -lt 3 -or @($center.actions).Count -lt 1) {
        throw "Expected settings, plan options, retention options, channels, and actions."
    }
    foreach ($settingID in @("plan", "retention", "notification-policy", "mail-route", "push-route", "archive", "autostart", "role-views", "privacy-data-rights")) {
        if (@($center.settings | Where-Object { $_.id -eq $settingID }).Count -ne 1) {
            throw "Expected customer setting '$settingID'."
        }
    }
    if ($center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no passwords" -or $center.privacy_boundary -notmatch "no screenshots" -or $center.privacy_boundary -notmatch "push endpoints") {
        throw "Expected strict customer settings privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint_url", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Customer settings center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 64 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 64 customer settings smoke passed addr=$Addr score=$($center.summary.settings_score) settings=$(@($center.settings).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
