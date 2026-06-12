param(
    [string]$Addr = "127.0.0.1:18181"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase55" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase55/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase55-smoke/$timestamp"

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
        "Provider Simulation Lab",
        "provider-simulation-section",
        "Simulation Route Proof",
        "Simulation Scenarios",
        "Simulation Action Queue",
        "Provider Privacy Proof",
        "data-jump-target=`"provider-simulation-section`"",
        "Notification Revenue Cockpit",
        "Executive Notification Console"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 55 dashboard marker '$expected'."
        }
    }

    $lab = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/provider-simulation-lab"
    if ($lab.summary.readiness_score -le 0 -or $lab.summary.routes_total -lt 3 -or [string]::IsNullOrWhiteSpace($lab.summary.recommended_paid_package)) {
        throw "Expected provider simulation readiness, routes, and package proof."
    }
    if ($lab.routes.Count -lt 3 -or $lab.scenarios.Count -lt 3 -or $lab.actions.Count -lt 1) {
        throw "Expected provider simulation route, scenario, and action surfaces."
    }
    if ($lab.privacy_boundary -notmatch "metadata-only" -or $lab.privacy_boundary -notmatch "no provider secrets" -or $lab.privacy_boundary -notmatch "alert bodies") {
        throw "Expected strict provider simulation privacy boundary."
    }

    $runBody = @{ mode = "dry_run"; channel = "push"; scenario = "urgent-anomaly-push"; reason = "phase55 smoke push simulation" } | ConvertTo-Json -Compress
    $simulated = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants/$tenantID/provider-simulation-lab" -ContentType "application/json" -Body $runBody
    if (-not $simulated.summary.push_ready -or $simulated.summary.simulated_routes -lt 1) {
        throw "Expected provider simulation push route proof after dry run."
    }
    $pushProof = $simulated.routes | Where-Object { $_.channel -eq "push" -and $_.simulation_status -eq "healthy" -and -not [string]::IsNullOrWhiteSpace($_.business_value) } | Select-Object -First 1
    if ($null -eq $pushProof) {
        throw "Expected healthy push provider simulation proof with business value."
    }
    $audit = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/audit-events"
    $auditProof = $audit.items | Where-Object { $_.action -eq "provider_simulation.rehearsed" } | Select-Object -First 1
    if ($null -eq $auditProof) {
        throw "Expected provider simulation audit proof."
    }

    $serialized = ($simulated | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Provider simulation lab leaked forbidden field marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 55 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 55 provider simulation smoke passed addr=$Addr readiness=$($simulated.summary.readiness_score) routes=$($simulated.routes.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
