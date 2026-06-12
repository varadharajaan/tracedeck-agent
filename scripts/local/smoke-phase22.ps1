param(
    [string]$Addr = "127.0.0.1:18109",
    [string]$ApiKey = "phase22-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase22" -LogRoot "logs/local/smoke" | Out-Null

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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase22/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 22 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase22-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 22 backend pid=$($backend.Id) addr=$Addr"
    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody -Headers $authHeaders | Out-Null

    $exportBody = @{ format = "json"; scope = "tenant" } | ConvertTo-Json -Compress
    $export = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/data-exports" -Body $exportBody -Headers $authHeaders
    if ($export.status -ne "ready" -or -not $export.storage_key -or $export.resource_count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected ready tenant data export manifest."
        exit 1
    }

    $deleteBody = @{ scope = "tenant"; reason = "family account data cleanup request" } | ConvertTo-Json -Compress
    $delete = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/delete-requests" -Body $deleteBody -Headers $authHeaders
    if ($delete.status -ne "queued" -or -not $delete.due_at) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected queued delete request."
        exit 1
    }

    $exports = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/data-exports" -Headers $authHeaders
    $deletes = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/delete-requests" -Headers $authHeaders
    if ($exports.count -lt 1 -or $deletes.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected export and delete request list results."
        exit 1
    }

    $audit = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/audit-events" -Headers $authHeaders
    if (-not ($audit.items | Where-Object { $_.action -eq "data_export.created" }) -or -not ($audit.items | Where-Object { $_.action -eq "delete_request.created" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected export and delete request audit events."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Data Export Center", "Delete Request Queue", "Auditable tenant export manifests", "Non-destructive deletion requests")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected dashboard HTML to include: $expected"
            exit 1
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 22 export and delete workflow smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 22 backend pid=$($backend.Id)"
    }
}
