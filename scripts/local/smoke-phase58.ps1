param(
    [string]$Addr = "127.0.0.1:18187"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase58" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase58/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase58-smoke/$timestamp"

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
        "Customer Success Packet",
        "Success Proof Stack",
        "Buyer Objection Answers",
        "Success Packet Actions",
        "Delivery And Trust Promise",
        "data-jump-target=`"customer-success-section`"",
        "Customer Control Room",
        "Mail And Push Delivery",
        "Package Billing Readiness",
        "Provider Simulation Lab"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 58 dashboard marker '$expected'."
        }
    }

    $packet = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/customer-success-packet"
    if ($packet.summary.readiness_score -le 0 -or $packet.summary.notification_score -le 0 -or $packet.summary.package_score -le 0 -or [string]::IsNullOrWhiteSpace($packet.summary.owner_next_step)) {
        throw "Expected customer success scores and owner next step."
    }
    if ($packet.summary.mail_delivered -le 0 -or $packet.summary.hosts_total -le 0) {
        throw "Expected customer success mail and host proof."
    }
    if ($packet.proofs.Count -lt 7 -or $packet.objections.Count -lt 4 -or $packet.actions.Count -lt 1) {
        throw "Expected customer success proofs, objections, and actions."
    }
    $pushProof = $packet.proofs | Where-Object { $_.id -eq "push-notification" -and -not [string]::IsNullOrWhiteSpace($_.status) } | Select-Object -First 1
    $mailProof = $packet.proofs | Where-Object { $_.id -eq "mail-delivery" -and -not [string]::IsNullOrWhiteSpace($_.status) } | Select-Object -First 1
    if (-not $pushProof -or -not $mailProof) {
        throw "Expected push notification and mail delivery proof."
    }
    if ($packet.privacy_boundary -notmatch "metadata-only" -or $packet.privacy_boundary -notmatch "no passwords" -or $packet.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict customer success privacy boundary."
    }

    $serialized = ($packet | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Customer success packet leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 58 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 58 customer success smoke passed addr=$Addr readiness_score=$($packet.summary.readiness_score) proofs=$($packet.proofs.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
