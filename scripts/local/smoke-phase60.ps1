param(
    [string]$Addr = "127.0.0.1:18191"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase60" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase60/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase60-smoke/$timestamp"

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
        "Portfolio Center",
        "Portfolio Alert Notifications",
        "Portfolio Delivery Proof",
        "Host Portfolio Rows",
        "Portfolio Segments",
        "Portfolio Owner Actions",
        "Portfolio Privacy Guard",
        "data-jump-target=`"portfolio-center-section`"",
        "Push Activation Center",
        "Customer Success Packet"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 60 dashboard marker '$expected'."
        }
    }

    $center = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/portfolio-center"
    if ($center.summary.portfolio_score -le 0 -or $center.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($center.summary.owner_next_step)) {
        throw "Expected portfolio scores and owner next step."
    }
    if ($center.summary.hosts_total -lt 1 -or $center.summary.mail_delivered -lt 1 -or $center.summary.dashboard_delivered -lt 1) {
        throw "Expected host, mail, and dashboard portfolio proof."
    }
    if ($center.hosts.Count -lt 1 -or $center.segments.Count -lt 5 -or $center.actions.Count -lt 1) {
        throw "Expected portfolio hosts, segments, and actions."
    }
    if ($center.alert_notifications.Count -lt 1 -or $center.delivery_proof.Count -lt 5) {
        throw "Expected alert notification rows and delivery proof cards."
    }
    $mailProof = @($center.delivery_proof | Where-Object {
        $_.channel -eq "email" -and -not [string]::IsNullOrWhiteSpace($_.status) -and -not [string]::IsNullOrWhiteSpace($_.next_action)
    })
    $pushProof = @($center.delivery_proof | Where-Object {
        $_.channel -eq "push" -and -not [string]::IsNullOrWhiteSpace($_.status) -and -not [string]::IsNullOrWhiteSpace($_.next_action)
    })
    if ($mailProof.Count -lt 1 -or $pushProof.Count -lt 1) {
        throw "Expected mail and push delivery proof cards with status and next action."
    }
    $alertProof = $center.alert_notifications[0]
    if ([string]::IsNullOrWhiteSpace($alertProof.email_status) -or [string]::IsNullOrWhiteSpace($alertProof.push_status) -or [string]::IsNullOrWhiteSpace($alertProof.dashboard_status) -or [string]::IsNullOrWhiteSpace($alertProof.next_action)) {
        throw "Expected portfolio alert notification route proof."
    }
    if ([string]::IsNullOrWhiteSpace($center.hosts[0].host_name) -or [string]::IsNullOrWhiteSpace($center.hosts[0].metadata_proof_summary)) {
        throw "Expected typed metadata-only host portfolio row."
    }
    if ($center.privacy_boundary -notmatch "metadata-only" -or $center.privacy_boundary -notmatch "no passwords" -or $center.privacy_boundary -notmatch "no screenshots") {
        throw "Expected strict portfolio privacy boundary."
    }

    $serialized = ($center | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Portfolio center leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 60 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 60 portfolio center smoke passed addr=$Addr portfolio_score=$($center.summary.portfolio_score) hosts=$($center.hosts.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
