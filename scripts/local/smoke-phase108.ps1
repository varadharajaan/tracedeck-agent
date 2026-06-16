param(
    [string]$Addr = "127.0.0.1:18270"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase108" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase108/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"

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
        if (Test-Path $pidFullPath) {
            try {
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                if ($health.status -eq "ok") {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper ready addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch {
                Start-Sleep -Milliseconds 500
            }
        }
        elseif ($helper.HasExited -and $helper.ExitCode -ne 0) {
            throw "Dashboard demo helper failed with exit code $($helper.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo helper did not become healthy at $baseUrl"
}

try {
    $baseUrl = "http://$Addr"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-browser-extension-skeleton.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Browser extension skeleton test failed with exit code $LASTEXITCODE"
    }

    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $deviceID = "phase108-extension-device"
    $hostName = "phase108-extension-host"
    $tenantID = "family-varadha"

    $tenantBody = @{
        tenant_id = $tenantID
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Depth 8
    Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants" -ContentType "application/json" -Body $tenantBody | Out-Null

    $enrollBody = @{
        tenant_id = $tenantID
        device_id = $deviceID
        host_name = $hostName
        profile = "ai_btech_student"
        os_name = "browser_extension"
    } | ConvertTo-Json -Depth 8
    Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -ContentType "application/json" -Body $enrollBody | Out-Null

    $ingestBody = @{
        tenant_id = $tenantID
        device_id = $deviceID
        host_name = $hostName
        profile = "ai_btech_student"
        os_name = "browser_extension"
        events = @(
            @{
                id = "phase108-extension-domain-1"
                type = "browser.domain.observed"
                source = "collector.browser.extension"
                observed_at = (Get-Date).ToUniversalTime().ToString("o")
                tenant_id = $tenantID
                device_id = $deviceID
                host_name = $hostName
                app_name = "chrome"
                process_id = 0
                path_hash = ""
                metadata = @{
                    browser_name = "chrome"
                    domain = "docs.python.org"
                    category = "study"
                    source_kind = "live_ingested"
                    evidence_scope = "live"
                    evidence_detail = "browser extension observed domain-only navigation metadata"
                    url_mode = "domain_only"
                    stored_url_mode = "domain_only"
                    visit_count = "1"
                    youtube_study_match = "true"
                }
            }
        )
    } | ConvertTo-Json -Depth 16

    $ingest = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/devices/$deviceID/telemetry-events" -ContentType "application/json" -Body $ingestBody
    if ($ingest.accepted_events -ne 1 -or $ingest.backend_visible_host -ne $true) {
        throw "Expected one accepted browser extension telemetry event."
    }

    $viewer = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/browser-activity?device_id=$deviceID&browser=chrome&limit=10"
    if (@($viewer.items).Count -lt 1) {
        throw "Expected browser activity row from extension telemetry."
    }
    $row = @($viewer.items)[0]
    if ($row.domain -ne "docs.python.org" -or $row.source_kind -ne "live_ingested" -or $row.evidence_scope -ne "live") {
        throw "Unexpected browser extension activity row: $($row | ConvertTo-Json -Depth 12)"
    }
    $rowSerialized = ($row | ConvertTo-Json -Depth 24).ToLowerInvariant()
    foreach ($forbidden in @("https://", "raw_url", "page_title", "cookie", "password", "screenshot", "provider_secret", "alert_body")) {
        if ($rowSerialized.Contains($forbidden)) {
            throw "Browser extension smoke leaked forbidden marker '$forbidden'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 108 smoke passed addr=$Addr extension_rows=$($viewer.summary.total)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
