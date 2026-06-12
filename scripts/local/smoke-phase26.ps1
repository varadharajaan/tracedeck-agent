param(
    [string]$Addr = "127.0.0.1:18117"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase26" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase26/$timestamp"
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

function Wait-TraceDeckNotificationRoutes {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        try {
            $routes = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/api/v1/tenants/family-varadha/notification-routes"
            if ($routes.count -ge 3) {
                return $routes
            }
        }
        catch { Start-Sleep -Milliseconds 500 }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo seed did not expose notification routes at $BaseUrl"
}

try {
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $baseUrl = "http://$Addr"
    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Notification Route Registry", "Route Readiness Proof")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected dashboard HTML to include: $expected"
        }
    }

    $routes = Wait-TraceDeckNotificationRoutes -BaseUrl $baseUrl
    if ($routes.count -lt 3) {
        throw "Expected seeded notification routes."
    }

    $body = @{
        channel = "push"
        provider = "web_push"
        recipient_label = "parent secondary phone"
        status = "watch"
        enabled = $true
        last_summary = "Secondary push route waiting for delivered proof."
    } | ConvertTo-Json -Compress
    $created = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/notification-routes" -ContentType "application/json" -Body $body
    if ($created.channel -ne "push" -or $created.provider -ne "web_push") {
        throw "Expected created push notification route."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 26 notification route registry smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
