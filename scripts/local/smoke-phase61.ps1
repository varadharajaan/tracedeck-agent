param(
    [string]$Addr = "127.0.0.1:18193"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase61" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase61/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase61-smoke/$timestamp"

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
        "Account Portfolio Index",
        "Account Tenant Rows",
        "Account Proof Cards",
        "Account Owner Actions",
        "data-jump-target=`"account-portfolio-section`"",
        "Portfolio Center",
        "Push Activation Center",
        "Customer Success Packet"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 61 dashboard marker '$expected'."
        }
    }

    $index = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/account-portfolio-index"
    if ($index.summary.account_score -le 0 -or $index.summary.notification_score -le 0 -or [string]::IsNullOrWhiteSpace($index.summary.owner_next_step)) {
        throw "Expected account portfolio scores and owner next step."
    }
    if ($index.summary.tenants_total -lt 1 -or $index.summary.hosts_total -lt 1 -or $index.summary.mail_delivered -lt 1 -or $index.summary.dashboard_delivered -lt 1) {
        throw "Expected account tenant, host, mail, and dashboard proof."
    }
    if ($index.tenants.Count -lt 1 -or $index.proof.Count -lt 5 -or $index.actions.Count -lt 1) {
        throw "Expected account tenant rows, proof cards, and actions."
    }
    if ([string]::IsNullOrWhiteSpace($index.tenants[0].tenant_id) -or [string]::IsNullOrWhiteSpace($index.tenants[0].next_action) -or [string]::IsNullOrWhiteSpace($index.tenants[0].privacy_boundary)) {
        throw "Expected typed metadata-only account tenant row."
    }
    if ($index.privacy_boundary -notmatch "metadata-only" -or $index.privacy_boundary -notmatch "no passwords" -or $index.privacy_boundary -notmatch "no screenshots") {
        throw "Expected strict account portfolio privacy boundary."
    }

    $serialized = ($index | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "card_number", "cvv", "payment_token")) {
        if ($serialized.Contains($forbidden)) {
            throw "Account portfolio leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 61 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 61 account portfolio smoke passed addr=$Addr account_score=$($index.summary.account_score) tenants=$($index.tenants.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
