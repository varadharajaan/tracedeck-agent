param(
    [string]$Addr = "127.0.0.1:18189"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase59" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase59/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase59-smoke/$timestamp"

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
        "Push Activation Center",
        "Push Route Proof",
        "Anomaly Push And Mail Scenarios",
        "Push Owner Actions",
        "Push Privacy Guard",
        "data-jump-target=`"push-activation-section`"",
        "Customer Success Packet",
        "Notification Revenue Cockpit"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 59 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/push-activation-center"
    if ($center.summary.activation_score -le 0 -or $center.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($center.summary.owner_next_step)) {
        throw "Expected push activation scores and owner next step."
    }
    if ($center.summary.mail_delivered -le 0 -or $center.summary.dashboard_delivered -le 0 -or $center.summary.push_routes_total -le 0) {
        throw "Expected mail fallback, dashboard fallback, and push route proof."
    }
    if ($center.routes.Count -lt 1 -or $center.scenarios.Count -lt 3 -or $center.actions.Count -lt 1) {
        throw "Expected push routes, scenarios, and actions."
    }
    if ($center.routes[0].provider -ne "web_push" -or [string]::IsNullOrWhiteSpace($center.routes[0].subscription_label) -or [string]::IsNullOrWhiteSpace($center.routes[0].endpoint_storage)) {
        throw "Expected provider-safe web push route proof."
    }
    if ($center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no push endpoints" -or $center.privacy_boundary -notmatch "no passwords") {
        throw "Expected strict push activation privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Push activation center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 59 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 59 push activation smoke passed addr=$Addr activation_score=$($center.summary.activation_score) routes=$($center.routes.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
