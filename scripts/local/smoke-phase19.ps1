param(
    [string]$Addr = "127.0.0.1:18103",
    [string]$ApiKey = "phase19-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase19" -LogRoot "logs/local/smoke" | Out-Null

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
    param([string]$BaseUrl)

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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase19/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 19 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase19-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 19 backend pid=$($backend.Id) addr=$Addr"
    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody -Headers $authHeaders | Out-Null

    $templates = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/alert-rule-templates" -Headers $authHeaders
    if ($templates.count -lt 4 -or -not ($templates.items | Where-Object { $_.id -eq "media_after_hours" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected alert rule template catalog."
        exit 1
    }

    $seededRules = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/alert-rules" -Headers $authHeaders
    if ($seededRules.count -lt 2) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected seeded tenant alert rules."
        exit 1
    }

    $ruleBody = @{
        template_id = "risky_software_detected"
        name = "Email when risky software appears"
        trigger = "risky_software"
        severity = "high"
        channels = @("email", "dashboard")
        condition = @{
            subject = "category"
            operator = "equals"
            value = "torrent_client"
            window_minutes = 0
            threshold = 0
        }
        enabled = $true
    } | ConvertTo-Json -Depth 5 -Compress
    $createdRule = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/alert-rules" -Body $ruleBody -Headers $authHeaders
    if ($createdRule.name -ne "Email when risky software appears" -or $createdRule.channels.Count -lt 2) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected custom alert rule creation."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("No-Code Alert Rules", "Rule Builder Recipes", "Saved tenant automations", "Paid templates for family, school, and business policy automation")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected dashboard HTML to include: $expected"
            exit 1
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 19 alert rule builder smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 19 backend pid=$($backend.Id)"
    }
}
