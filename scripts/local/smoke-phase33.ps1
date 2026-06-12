param(
    [string]$Addr = "127.0.0.1:18139"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase33" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase33/$timestamp"
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
    foreach ($expected in @("Monetisation Command Views", "Notification Monetisation Proof", "activity-view-list", "proof-alert-reach")) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard marker '$expected'."
        }
    }

    $views = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/activity-views"
    if ($views.count -ne 4) {
        throw "Expected four seeded monetisation command views."
    }
    $emailView = $views.items | Where-Object { $_.id -eq "email-proof" } | Select-Object -First 1
    $pushView = $views.items | Where-Object { $_.id -eq "push-retry" } | Select-Object -First 1
    if (-not $emailView -or $emailView.filter.channel -ne "email" -or -not $pushView -or $pushView.filter.channel -ne "push") {
        throw "Expected seeded email and push command filters."
    }

    $payload = @{
        name = "Business dashboard misses"
        description = "Paid demo view for failed dashboard delivery proof."
        paid_tier = "business"
        sort_order = 9
        filter = @{
            kind = "delivery"
            channel = "dashboard"
            status = "failed"
            limit = 10
        }
    } | ConvertTo-Json -Depth 6
    $created = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/activity-views" -ContentType "application/json" -Body $payload
    if ($created.paid_tier -ne "business" -or $created.filter.channel -ne "dashboard") {
        throw "Expected created business dashboard activity view."
    }

    $audit = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/audit-events"
    if (-not ($audit.items | Where-Object { $_.action -eq "activity_view.created" })) {
        throw "Expected activity view audit event."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 33 smoke passed addr=$Addr views=$($views.count) created=$($created.id)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
