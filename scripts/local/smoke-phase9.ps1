param(
    [string]$Addr = "127.0.0.1:18084"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase9" -LogRoot "logs/local/smoke" | Out-Null

$backend = $null

function Invoke-TraceDeckJson {
    param(
        [string]$Method,
        [string]$Uri,
        [string]$Body = ""
    )

    $headers = @{ "Content-Type" = "application/json" }
    if ($Body) {
        return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $headers -Body $Body
    }
    return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $headers
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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase9/$timestamp"
    $stdoutPath = Join-Path $smokeRoot "backend.out.log"
    $stderrPath = Join-Path $smokeRoot "backend.err.log"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 9 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started backend Phase 9 smoke process: $($backend.Id)"

    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody | Out-Null

    $deviceBody = @{
        tenant_id = "family-varadha"
        device_id = "phase9-dashboard-device"
        host_name = "phase9-study-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody | Out-Null

    $overview = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase9-dashboard-device/overview"
    if ($overview.device.device_id -ne "phase9-dashboard-device" -or $overview.risk_score -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Host overview smoke failed."
        exit 1
    }
    if ($overview.policy_violations.Count -lt 1 -or $overview.alert_deliveries.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected policy and alert delivery data in host overview."
        exit 1
    }

    $policy = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase9-dashboard-device/policy-violations"
    if ($policy.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected policy violations."
        exit 1
    }

    $anomalies = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase9-dashboard-device/anomalies"
    if ($anomalies.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected anomalies."
        exit 1
    }

    $tamper = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase9-dashboard-device/tamper-events"
    if ($tamper.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected tamper events."
        exit 1
    }

    $deliveries = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase9-dashboard-device/alert-deliveries"
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "email" -and $_.status -eq "delivered" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected delivered email alert route."
        exit 1
    }
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "push" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected push alert route visibility."
        exit 1
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Method "GET" -Uri "$baseUrl/"
    if (
        $dashboard.StatusCode -ne 200 -or
        $dashboard.Content -notmatch "TraceDeck Console" -or
        $dashboard.Content -notmatch "Alert Delivery" -or
        $dashboard.Content -notmatch "Tamper And Trust"
    ) {
        Write-TraceDeckLog -Level "ERROR" -Message "Phase 9 dashboard smoke failed."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 9 dashboard smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped backend Phase 9 smoke process: $($backend.Id)"
    }
}
