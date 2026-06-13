param(
    [string]$Addr = "127.0.0.1:18227"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase75" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase75/$timestamp"
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

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 75 dashboard markers" -Command {
        $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
        foreach ($expected in @(
            "Delivery Assurance Center",
            "Route Truth Matrix",
            "Delivery Truth Events",
            "Provider Proof Readiness",
            "data-jump-target=""delivery-assurance-section"""
        )) {
            if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 75 dashboard marker '$expected'."
            }
        }
    }

    $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
    if ($devices.count -lt 1) {
        throw "Expected at least one seeded device."
    }
    $deviceID = $devices.items[0].device_id

    $assurance = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/delivery-assurance?device_id=$deviceID&limit=8"
    if ($assurance.summary.routes_total -ne 3 -or $assurance.summary.provider_confirmed -ne 0 -or $assurance.summary.demo_only -lt 1 -or $assurance.summary.retrying -lt 1) {
        throw "Expected demo-only, retrying, and zero provider-confirmed delivery assurance. Summary: $($assurance.summary | ConvertTo-Json -Compress)"
    }
    if ($assurance.summary.email_provider_ready -eq $true -or $assurance.summary.push_provider_ready -eq $true -or $assurance.summary.buyer_ready -eq $true) {
        throw "Seeded demo delivery must not be provider-ready or buyer-ready."
    }
    if ($assurance.summary.dashboard_route_ready -ne $true) {
        throw "Expected local dashboard fallback to be ready."
    }
    if ($assurance.privacy_boundary -notmatch "metadata-only" -or $assurance.privacy_boundary -notmatch "no provider secrets") {
        throw "Expected strict delivery assurance privacy boundary."
    }

    $emailRoute = $assurance.routes | Where-Object { $_.channel -eq "email" } | Select-Object -First 1
    $pushRoute = $assurance.routes | Where-Object { $_.channel -eq "push" } | Select-Object -First 1
    if (-not $emailRoute -or $emailRoute.assurance_state -ne "demo_only" -or $emailRoute.source_kind -ne "demo_seed") {
        throw "Expected email route to be demo_only with demo_seed source."
    }
    if (-not $pushRoute -or $pushRoute.assurance_state -ne "retrying") {
        throw "Expected push route to be retrying."
    }

    $filtered = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/delivery-assurance?device_id=$deviceID&channel=email&assurance_state=demo_only"
    if (@($filtered.routes).Count -ne 1 -or $filtered.routes[0].assurance_state -ne "demo_only") {
        throw "Expected filtered email demo-only route."
    }

    $serialized = ($assurance | ConvertTo-Json -Depth 20).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "push_endpoint", "screenshot_bytes", "raw_url", "page_title", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Delivery assurance response exposed forbidden marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 75 runtime doctor assurance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-doctor.ps1 -Addr $Addr
    }

    $doctor = Get-Content -Path (Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-doctor.json") -Raw | ConvertFrom-Json
    if ($doctor.local.base_url -ne $baseUrl) {
        throw "Expected runtime doctor to check $baseUrl."
    }
    if ($doctor.local.delivery_assurance.ok -ne $true -or $doctor.local.delivery_assurance.demo_only -lt 1 -or $doctor.local.delivery_assurance.retrying -lt 1) {
        throw "Expected runtime doctor delivery assurance proof. Delivery assurance: $($doctor.local.delivery_assurance | ConvertTo-Json -Compress)"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 75 delivery assurance smoke passed addr=$Addr device=$deviceID score=$($assurance.summary.assurance_score) demo=$($assurance.summary.demo_only) retrying=$($assurance.summary.retrying)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
