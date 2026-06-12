param(
    [string]$Addr = "127.0.0.1:18107",
    [string]$ApiKey = "phase21-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase21" -LogRoot "logs/local/smoke" | Out-Null

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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase21/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 21 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase21-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 21 backend pid=$($backend.Id) addr=$Addr"
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
        device_id = "phase21-policy-device"
        host_name = "phase21-policy-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody -Headers $authHeaders | Out-Null

    $seededGroups = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/device-groups" -Headers $authHeaders
    if ($seededGroups.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected seeded device groups."
        exit 1
    }

    $groupBody = @{
        name = "Exam Mode Devices"
        description = "Managed exam preparation laptops"
        profile = "school-laptop"
        device_ids = @("phase21-policy-device")
        policy_template_id = "school-laptop"
    } | ConvertTo-Json -Compress
    $group = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/device-groups" -Body $groupBody -Headers $authHeaders
    if ($group.name -ne "Exam Mode Devices" -or $group.device_ids.Count -ne 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected created device group with assigned device."
        exit 1
    }

    $assignmentBody = @{
        name = "Exam mode rollout"
        target_type = "device_group"
        target_id = $group.id
        policy_template_id = "school-laptop"
        alert_rule_ids = @("manual-rule-001")
        mode = "active"
    } | ConvertTo-Json -Compress
    $assignment = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants/family-varadha/policy-assignments" -Body $assignmentBody -Headers $authHeaders
    if ($assignment.name -ne "Exam mode rollout" -or $assignment.status -ne "active") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected active policy assignment."
        exit 1
    }

    $assignments = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/policy-assignments" -Headers $authHeaders
    if ($assignments.count -lt 2) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected seeded and custom policy assignments."
        exit 1
    }

    $audit = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/audit-events" -Headers $authHeaders
    if (-not ($audit.items | Where-Object { $_.action -eq "device_group.created" }) -or -not ($audit.items | Where-Object { $_.action -eq "policy_assignment.created" })) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected device group and policy assignment audit events."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    foreach ($expected in @("Device Groups", "Policy Assignments", "Managed family, school, and business device cohorts", "Tenant and group-level policy rollout status")) {
        if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected dashboard HTML to include: $expected"
            exit 1
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 21 device group and policy assignment smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 21 backend pid=$($backend.Id)"
    }
}
