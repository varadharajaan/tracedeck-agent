param(
    [string]$Addr = "127.0.0.1:18097",
    [string]$ApiKey = "phase16-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase16" -LogRoot "logs/local/smoke" | Out-Null

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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase16/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 16 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase16-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 16 backend pid=$($backend.Id) addr=$Addr"
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
        device_id = "phase16-dashboard-device"
        host_name = "phase16-dashboard-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody -Headers $authHeaders | Out-Null

    $overview = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase16-dashboard-device/overview" -Headers $authHeaders
    if ($overview.alert_deliveries.Count -lt 3 -or $overview.policy_violations.Count -lt 1 -or $overview.anomalies.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected host overview to include policy, anomaly, and delivery signals."
        exit 1
    }

    $deliveries = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase16-dashboard-device/alert-deliveries" -Headers $authHeaders
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "email" -and $_.status -eq "delivered" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected delivered email alert route."
        exit 1
    }
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "push" -and $_.status -eq "retrying" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected retrying push notification route."
        exit 1
    }
    if (-not ($deliveries.items | Where-Object { $_.channel -eq "dashboard" -and $_.status -eq "delivered" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected delivered dashboard feed route."
        exit 1
    }

    $report = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase16-dashboard-device/reports/weekly" -Headers $authHeaders
    if (-not $report.email_ready -or -not $report.pdf_ready -or [string]::IsNullOrWhiteSpace($report.email_subject)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected weekly report email and PDF readiness."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    $requiredDashboardText = @(
        "Anomaly Notification Inbox",
        "Mail Delivery Center",
        "Push Routing",
        "Email SLA",
        "Paid Trigger",
        "Endpoint productivity and risk observability",
        "Policy Template Marketplace",
        "Retention And Archive Plans"
    )
    foreach ($expected in $requiredDashboardText) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected dashboard HTML to include: $expected"
            exit 1
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 16 monetisation dashboard smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 16 backend pid=$($backend.Id)"
    }
}
