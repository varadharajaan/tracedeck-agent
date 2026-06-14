param(
    [string]$Addr = "127.0.0.1:18241"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase86" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase86/$timestamp"
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
        if ((Test-Path $pidFullPath)) {
            try {
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($health.status -eq "ok" -and $devices.count -ge 1) {
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
    throw "Dashboard demo helper did not seed devices at $baseUrl"
}

function Assert-TraceDeckContent {
    param(
        [string]$Name,
        [string]$Content,
        [string[]]$RequiredMarkers,
        [string[]]$ForbiddenPatterns
    )

    foreach ($marker in $RequiredMarkers) {
        if (-not $Content.Contains($marker)) {
            throw "$Name missing required marker '$marker'"
        }
    }
    foreach ($pattern in $ForbiddenPatterns) {
        if ($Content -cmatch $pattern) {
            throw "$Name contains forbidden debug/prototype pattern '$pattern'"
        }
    }
}

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Method "GET" -Uri "$baseUrl/"
    $browser = Invoke-WebRequest -UseBasicParsing -Method "GET" -Uri "$baseUrl/browser-activity"
    $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
    $activity = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/browser-activity?limit=25"
    $assurance = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/delivery-assurance?limit=25"
    if ($devices.count -lt 1) {
        throw "Expected seeded smoke device."
    }
    $deviceID = $devices.items[0].device_id
    $policy = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/policy-violations"
    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/alert-deliveries"
    $policyJson = $policy | ConvertTo-Json -Depth 20
    $deliveryJson = $deliveries | ConvertTo-Json -Depth 20
    if ($policy.count -ne 0 -or $policyJson.Contains("VLC media player") -or $policyJson.Contains("demo_seed")) {
        throw "Default policy endpoint leaked demo evidence."
    }
    if ($deliveries.count -ne 0 -or $deliveryJson.Contains("demo_seed")) {
        throw "Default alert delivery endpoint leaked demo evidence."
    }
    $demoPolicy = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/$deviceID/policy-violations?include_demo=true"
    if ($demoPolicy.count -lt 1 -or $demoPolicy.items[0].source_kind -ne "demo_seed") {
        throw "Expected opt-in demo policy evidence."
    }

    $forbidden = @(
        "Browser\{",
        "Center\{",
        "\[[BCTR]\]",
        "\{[BCTR]\}",
        "\bTD\b",
        "\bRev Ops\b",
        "\bNotif Rev\b",
        "\bNotify Pro\b",
        "\bNotifs\b"
    )
    Assert-TraceDeckContent -Name "dashboard" -Content $dashboard.Content -RequiredMarkers @(
        "TraceDeck Console",
        "Phase 86 monetisation-grade UI layer",
        "Premium endpoint activity",
        "Browser Evidence",
        "Sync View",
        "server-status-light",
        "dashboard-page-nav",
        "Workspace Navigator",
        "Delivery Assurance Center"
    ) -ForbiddenPatterns $forbidden
    Assert-TraceDeckContent -Name "browser viewer" -Content $browser.Content -RequiredMarkers @(
        "TraceDeck Browser Activity",
        "Phase 86 browser evidence UI layer",
        "Browser Viewer",
        "Chrome, Brave, and Edge evidence",
        "Sync View",
        "server-status-light",
        "<th>Source</th>",
        "metadata-only guard"
    ) -ForbiddenPatterns $forbidden

    if ($activity.privacy_boundary -notmatch "metadata-only") {
        throw "Expected browser activity privacy boundary to be metadata-only."
    }
    if ($activity.items.Count -gt 0) {
        $first = $activity.items | Select-Object -First 1
        if (-not $first.source_kind -or -not $first.evidence_scope -or -not $first.evidence_detail) {
            throw "Expected browser activity rows to include provenance fields."
        }
    }
    if ($assurance.privacy_boundary -notmatch "metadata-only") {
        throw "Expected delivery assurance privacy boundary to be metadata-only."
    }
    if ($assurance.summary.demo_only -gt 0 -and $assurance.summary.buyer_ready) {
        throw "Demo-only delivery rows must not be marked buyer ready."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 86 premium UI smoke passed addr=$Addr browser_rows=$($activity.summary.total) delivery_routes=$($assurance.summary.routes_total)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
