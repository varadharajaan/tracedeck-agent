param(
    [string]$Addr = "127.0.0.1:18082"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase6" -LogRoot "logs/local/smoke" | Out-Null

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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase6/$timestamp"
    $stdoutPath = Join-Path $smokeRoot "backend.out.log"
    $stderrPath = Join-Path $smokeRoot "backend.err.log"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 6 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started backend Phase 6 smoke process: $($backend.Id)"

    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $plans = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/plans"
    if (-not ($plans.items | Where-Object { $_.id -eq "family_pro" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected Family Pro plan."
        exit 1
    }

    $roles = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/roles"
    if (-not ($roles.items | Where-Object { $_.id -eq "parent" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected parent role."
        exit 1
    }

    $retention = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/retention-tiers"
    if (-not ($retention.items | Where-Object { $_.id -eq "family_cloud_90_365_archive" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected family retention tier."
        exit 1
    }

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress

    $tenant = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody
    if ($tenant.tenant_id -ne "family-varadha" -or $tenant.plan_id -ne "family_pro") {
        Write-TraceDeckLog -Level "ERROR" -Message "Tenant creation smoke failed."
        exit 1
    }

    $tenantAudit = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/audit-events"
    if ($tenantAudit.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected tenant audit event."
        exit 1
    }

    $audit = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/audit-events"
    if ($audit.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected global audit event."
        exit 1
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Method "GET" -Uri "$baseUrl/"
    if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch "Plans") {
        Write-TraceDeckLog -Level "ERROR" -Message "Dashboard SaaS readiness smoke failed."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 6 backend smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped backend Phase 6 smoke process: $($backend.Id)"
    }
}
