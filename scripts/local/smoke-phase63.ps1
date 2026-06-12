param(
    [string]$Addr = "127.0.0.1:18197"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase63" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase63/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase63-smoke/$timestamp"

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
        "Tenant Onboarding Center",
        "Setup Checklist",
        "Role Handoff",
        "Onboarding Proof",
        "Onboarding Owner Actions",
        "data-jump-target=`"onboarding-center-section`"",
        "onboarding-kpi-notifications",
        "onboarding-kpi-autostart",
        "onboarding-kpi-archive"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 63 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/onboarding-center"
    if ($center.summary.readiness_score -le 0 -or $center.summary.setup_steps_total -lt 8) {
        throw "Expected onboarding readiness score and setup checklist."
    }
    if ($center.summary.hosts_total -lt 1 -or $center.summary.roles_total -lt 4) {
        throw "Expected host and role onboarding proof."
    }
    if (@($center.proof).Count -lt 6 -or @($center.actions).Count -lt 1) {
        throw "Expected onboarding proof cards and owner actions."
    }
    foreach ($stepID in @("agent-install", "autostart", "notification-policy", "mail-push-proof", "archive-retention", "role-dashboards", "package-readiness", "privacy-guard")) {
        if (@($center.steps | Where-Object { $_.id -eq $stepID }).Count -ne 1) {
            throw "Expected onboarding step '$stepID'."
        }
    }
    if ($center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no passwords" -or $center.privacy_boundary -notmatch "no screenshots" -or $center.privacy_boundary -notmatch "push endpoints") {
        throw "Expected strict onboarding privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint_url", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Onboarding center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 63 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 63 onboarding center smoke passed addr=$Addr readiness=$($center.summary.readiness_score) steps=$(@($center.steps).Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
