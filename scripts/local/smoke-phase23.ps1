param(
    [string]$Addr = "127.0.0.1:18111",
    [string]$ApiKey = "phase23-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase23" -LogRoot "logs/local/smoke" | Out-Null

$backend = $null

function Invoke-TraceDeckJson {
    param([string]$Method, [string]$Uri, [string]$Body = "", [hashtable]$Headers = @{})
    $requestHeaders = @{ "Content-Type" = "application/json" }
    foreach ($key in $Headers.Keys) { $requestHeaders[$key] = $Headers[$key] }
    if ($Body) { return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $requestHeaders -Body $Body }
    return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $requestHeaders
}

function Wait-TraceDeckBackend {
    param([string]$BaseUrl)
    $deadline = (Get-Date).AddSeconds(45)
    while ((Get-Date) -lt $deadline) {
        try {
            $health = Invoke-TraceDeckJson -Method "GET" -Uri "$BaseUrl/health"
            if ($health.status -eq "ok") { return }
        }
        catch { Start-Sleep -Milliseconds 500 }
    }
    throw "Backend did not become healthy at $BaseUrl"
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase23/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 23 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase23-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 23 backend pid=$($backend.Id) addr=$Addr"
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
        device_id = "phase23-device-001"
        host_name = "phase23-study-laptop"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody -Headers $authHeaders | Out-Null

    $summary = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/operations-summary" -Headers $authHeaders
    if ($summary.hosts_total -lt 1 -or $summary.delivery_total -lt 1 -or $summary.open_policy_violations -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected tenant operations summary with host, delivery, and policy data."
        exit 1
    }
    if (-not $summary.last_email -or $summary.last_email.channel -ne "email") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected latest email delivery proof."
        exit 1
    }
    if ($summary.priority_signals.Count -lt 1 -or $summary.upgrade_signals.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected priority and upgrade proof signals."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Customer Operations Cockpit", "Escalation Workbench", "Notification Delivery Board", "Upgrade Proof Pack", "Mail Delivery Proof", "Push Reach")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected dashboard HTML to include: $expected"
            exit 1
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 23 customer operations dashboard smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 23 backend pid=$($backend.Id)"
    }
}
