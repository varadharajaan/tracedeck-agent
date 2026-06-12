param(
    [string]$Addr = "127.0.0.1:18115"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase25" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase25/$timestamp"
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
                if ($health.status -eq "ok") {
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
    throw "Dashboard demo helper did not become ready at $baseUrl"
}

function Wait-TraceDeckMonetizationSeed {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        try {
            $summary = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/api/v1/tenants/family-varadha/monetization-summary"
            if ($summary.readiness_score -gt 0 -and $summary.notification_routes.Count -eq 3) {
                return $summary
            }
        }
        catch { Start-Sleep -Milliseconds 500 }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo seed did not expose monetization summary at $BaseUrl"
}

try {
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $baseUrl = "http://$Addr"
    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Monetisation Command Center", "Notification Guarantee", "Paid Feature Proof", "Conversion Action Queue")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard HTML to include: $expected"
        }
    }

    $summary = Wait-TraceDeckMonetizationSeed -BaseUrl $baseUrl
    if ($summary.readiness_score -lt 1 -or $summary.notification_routes.Count -ne 3 -or -not $summary.notification_promise.email) {
        throw "Expected live monetization summary with route proof."
    }
    if ($summary.paid_capabilities.Count -lt 4 -or $summary.conversion_actions.Count -lt 1) {
        throw "Expected paid capability and conversion action proof."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 25 monetization dashboard smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
