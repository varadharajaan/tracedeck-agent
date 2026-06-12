param(
    [string]$Addr = "127.0.0.1:18183"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase56" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase56/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase56-smoke/$timestamp"

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
        "Package Billing Readiness",
        "package-billing-section",
        "Plan Fit Matrix",
        "Feature Gate Proof",
        "Billing Milestones",
        "Upgrade Actions",
        "data-jump-target=`"package-billing-section`"",
        "Provider Simulation Lab",
        "Notification Revenue Cockpit",
        "Executive Notification Console"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 56 dashboard marker '$expected'."
        }
    }

    $readiness = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/package-billing-readiness"
    if ($readiness.summary.package_score -le 0 -or [string]::IsNullOrWhiteSpace($readiness.summary.billing_status) -or [string]::IsNullOrWhiteSpace($readiness.summary.next_best_action)) {
        throw "Expected package billing score, status, and next action."
    }
    if ($readiness.plans.Count -lt 4 -or $readiness.feature_gates.Count -lt 8 -or $readiness.milestones.Count -lt 5 -or $readiness.actions.Count -lt 1) {
        throw "Expected package billing plans, feature gates, milestones, and actions."
    }
    if ($readiness.privacy_boundary -notmatch "metadata-only" -or $readiness.privacy_boundary -notmatch "no payment card data" -or $readiness.privacy_boundary -notmatch "no passwords") {
        throw "Expected strict package billing privacy boundary."
    }

    $serialized = ($readiness | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("card_number", "cvv", "payment_token", "provider_secret", "screenshot_bytes", "raw_url", "page_title", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Package billing readiness leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 56 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 56 package billing smoke passed addr=$Addr package_score=$($readiness.summary.package_score) gates=$($readiness.feature_gates.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}

