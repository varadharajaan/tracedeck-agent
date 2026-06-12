param(
    [string]$Addr = "127.0.0.1:18088",
    [string]$ApiKey = "phase11-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase11" -LogRoot "logs/local/smoke" | Out-Null

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

function Start-Phase11Backend {
    param(
        [string]$ExePath,
        [string]$StatePath,
        [string]$StdoutPath,
        [string]$StderrPath
    )

    $script:backend = Start-Process -FilePath $ExePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$StatePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase11-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $StdoutPath -RedirectStandardError $StderrPath -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 11 backend pid=$($script:backend.Id) addr=$Addr"
}

function Stop-Phase11Backend {
    if ($script:backend -and -not $script:backend.HasExited) {
        Stop-Process -Id $script:backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 11 backend pid=$($script:backend.Id)"
    }
    $script:backend = $null
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase11/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 11 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    Start-Phase11Backend -ExePath $exePath -StatePath $statePath -StdoutPath (Join-Path $smokeRoot "backend-first.out.log") -StderrPath (Join-Path $smokeRoot "backend-first.err.log")
    Wait-TraceDeckBackend -BaseUrl $baseUrl

    try {
        Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices" | Out-Null
        Write-TraceDeckLog -Level "ERROR" -Message "Expected unauthorized device list without API key."
        exit 1
    }
    catch {
        if ($_.Exception.Response.StatusCode.value__ -ne 401) {
            throw
        }
    }

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
        device_id = "phase11-persistent-device"
        host_name = "phase11-persistent-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody -Headers $authHeaders | Out-Null

    $overview = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase11-persistent-device/overview" -Headers $authHeaders
    if ($overview.device.device_id -ne "phase11-persistent-device" -or $overview.policy_violations.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected seeded risk data before restart."
        exit 1
    }

    if (-not (Test-Path -LiteralPath $statePath)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected persistent state file: $statePath"
        exit 1
    }

    Stop-Phase11Backend

    Start-Phase11Backend -ExePath $exePath -StatePath $statePath -StdoutPath (Join-Path $smokeRoot "backend-second.out.log") -StderrPath (Join-Path $smokeRoot "backend-second.err.log")
    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $devices = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices" -Headers $authHeaders
    if (-not ($devices.items | Where-Object { $_.device_id -eq "phase11-persistent-device" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Persistent device was not loaded after restart."
        exit 1
    }

    $deliveries = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase11-persistent-device/alert-deliveries" -Headers $authHeaders
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "email" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Persistent alert delivery data was not loaded after restart."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 11 persistence/auth smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    Stop-Phase11Backend
}
