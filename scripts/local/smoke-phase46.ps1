param(
    [string]$Addr = "127.0.0.1:18163"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase46" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase46/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase46-smoke/$timestamp"

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
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($devices.count -ge 1) {
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
    throw "Dashboard demo helper did not seed devices at $baseUrl"
}

try {
    $baseUrl = "http://$Addr"
    $tenantID = "family-varadha"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Monetisation Command Center",
        "Delivery Remediation Center",
        "Remediation Action Ledger",
        "Remediation SLA",
        "delivery-remediation-action-list",
        "delivery-remediation-plan-list"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 46 dashboard marker '$expected'."
        }
    }

    $remediation = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/delivery-remediation"
    if ($remediation.summary.routes_total -lt 3 -or $remediation.privacy_boundary -notmatch "without live provider sends") {
        throw "Expected provider-safe delivery remediation summary."
    }

    $body = @{
        mode = "dry_run"
        channel = "push"
        action = "retry_plan"
        reason = "phase46 smoke route recovery plan"
        owner = "parent mobile push subscription"
    } | ConvertTo-Json -Compress
    $planned = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants/$tenantID/delivery-remediation" -ContentType "application/json" -Body $body
    if ($planned.summary.planned_actions -lt 1 -or $planned.recent_plans.Count -lt 1) {
        throw "Expected delivery remediation dry-run plan to be recorded."
    }
    if ($planned.recent_plans[0].action -ne "retry_plan" -or $planned.recent_plans[0].channel -ne "push") {
        throw "Expected push retry remediation plan in recent ledger."
    }

    try {
        Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants/$tenantID/delivery-remediation" -ContentType "application/json" -Body '{"mode":"send_live","action":"retry_plan"}' | Out-Null
        throw "Expected live send remediation mode to fail."
    }
    catch {
        if ($_.Exception.Response.StatusCode.value__ -ne 400) {
            throw
        }
    }

    $audit = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/audit-events"
    if (-not ($audit.items | Where-Object { $_.action -eq "delivery_remediation.planned" })) {
        throw "Expected delivery remediation audit event."
    }

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot

    Write-TraceDeckLog -Level "INFO" -Message "Phase 46 delivery remediation smoke passed addr=$Addr plans=$($planned.summary.planned_actions) problems=$($planned.summary.problems_open)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
