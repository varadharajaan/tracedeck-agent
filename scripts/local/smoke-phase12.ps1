param(
    [string]$Addr = "127.0.0.1:18092",
    [string]$ApiKey = "phase12-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase12" -LogRoot "logs/local/smoke" | Out-Null

$backend = $null

function Invoke-TraceDeckJson {
    param(
        [string]$Method,
        [string]$Uri,
        [string]$Body = "",
        [hashtable]$Headers = @{}
    )

    $requestHeaders = @{ "Content-Type" = "application/json" }
    foreach ($key in $Headers.Keys) {
        $requestHeaders[$key] = $Headers[$key]
    }
    if ($Body) {
        return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $requestHeaders -Body $Body
    }
    return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $requestHeaders
}

function Wait-TraceDeckBackend {
    param(
        [string]$BaseUrl
    )

    $deadline = (Get-Date).AddSeconds(45)
    while ((Get-Date) -lt $deadline) {
        try {
            $health = Invoke-TraceDeckJson -Method "GET" -Uri "$BaseUrl/health"
            if ($health.status -eq "ok") {
                return
            }
        }
        catch {
            Start-Sleep -Milliseconds 500
        }
    }
    throw "Backend did not become healthy at $BaseUrl"
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase12/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 12 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase12-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 12 backend pid=$($backend.Id) addr=$Addr"
    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody -Headers $authHeaders | Out-Null

    $deviceBody = @{
        tenant_id = "family-varadha"
        device_id = "phase12-health-device"
        host_name = "phase12-health-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody -Headers $authHeaders | Out-Null

    $overview = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase12-health-device/overview" -Headers $authHeaders
    if ($overview.health.score -lt 1 -or -not $overview.health.status) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected health score in host overview."
        exit 1
    }

    $deviceHealth = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase12-health-device/health" -Headers $authHeaders
    if ($deviceHealth.device_id -ne "phase12-health-device" -or $deviceHealth.score -lt 1 -or -not $deviceHealth.agent_healthy) {
        Write-TraceDeckLog -Level "ERROR" -Message "Unexpected standalone device health payload."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch "Device Health" -or $dashboard.Content -notmatch "Notification Operations" -or $dashboard.Content -notmatch "Product Packaging") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected monetisation-ready dashboard panels in HTML."
        exit 1
    }

    $report = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase12-health-device/reports/weekly" -Headers $authHeaders
    if ($report.device_id -ne "phase12-health-device") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected weekly report endpoint for selected host."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 12 device health and dashboard smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 12 backend pid=$($backend.Id)"
    }
}
