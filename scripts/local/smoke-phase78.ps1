param(
    [string]$Addr = "127.0.0.1:18231"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase78" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase78/$timestamp"
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
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($health.status -eq "ok" -and $devices.count -ge 1) {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper ready addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch {
                Start-Sleep -Milliseconds 500
            }
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

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 78 provider setup dashboard markers" -Command {
        $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
        foreach ($expected in @(
            "Notification Provider Setup Center",
            "Provider Channel Setup",
            "Provider Setup Checklist",
            "Provider Setup Actions",
            "data-jump-target=""notification-provider-setup-section""",
            "nav-provider-setup-meta"
        )) {
            if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 78 dashboard marker '$expected'."
            }
        }
    }

    $setup = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/notification-provider-setup"
    if ($setup.summary.routes_total -ne 3 -or $setup.summary.channels_total -ne 3 -or @($setup.channels).Count -ne 3) {
        throw "Expected three notification provider setup channels. Summary: $($setup.summary | ConvertTo-Json -Compress)"
    }
    if ($setup.summary.demo_only -lt 1 -or $setup.summary.retrying -lt 1) {
        throw "Expected demo-only and retrying provider setup truth labels. Summary: $($setup.summary | ConvertTo-Json -Compress)"
    }
    if ($setup.summary.email_provider_confirmed -eq $true -or $setup.summary.push_provider_confirmed -eq $true -or $setup.summary.buyer_ready -eq $true) {
        throw "Seeded setup must not claim provider-confirmed email/push or buyer readiness."
    }
    if (@($setup.checklist).Count -lt 3 -or @($setup.actions).Count -lt 1) {
        throw "Expected setup checklist and owner actions."
    }
    if ($setup.privacy_boundary -notmatch "metadata-only" -or $setup.privacy_boundary -notmatch "no provider secrets") {
        throw "Expected strict notification provider setup privacy boundary."
    }

    $serialized = ($setup | ConvertTo-Json -Depth 20).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body", "raw_provider_payload", "password_value")) {
        if ($serialized.Contains($forbidden)) {
            throw "Notification provider setup response exposed forbidden marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 78 screenshot-free layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-layout/phase78-smoke"
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 78 light and dark theme contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-theme/phase78-smoke"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 78 product visual quality contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-visual-quality/phase78-smoke"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 78 notification provider setup smoke passed addr=$Addr setup_score=$($setup.summary.setup_score) demo=$($setup.summary.demo_only) retrying=$($setup.summary.retrying)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
