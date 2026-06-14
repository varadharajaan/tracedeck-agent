param(
    [string]$Addr = "127.0.0.1:18239"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase84" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase84/$timestamp"
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
    $activity = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/browser-activity?limit=25"
    $assurance = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/delivery-assurance?limit=25"

    $forbidden = @(
        "Browser\{",
        "Center\{",
        "\[[BCTR]\]",
        "\{[BCTR]\}",
        "\bTD\b",
        "\bRev Ops\b",
        "\bNotif Rev\b",
        "\bNotify Pro\b"
    )
    Assert-TraceDeckContent -Name "dashboard" -Content $dashboard.Content -RequiredMarkers @(
        "TraceDeck Console",
        "Phase 84 customer-grade UI layer",
        "Browser Activity",
        "Theme: Light",
        "server-status-light",
        "dashboard-page-nav",
        "Workspace Navigator",
        "Delivery Assurance Center"
    ) -ForbiddenPatterns $forbidden
    Assert-TraceDeckContent -Name "browser viewer" -Content $browser.Content -RequiredMarkers @(
        "TraceDeck Browser Activity",
        "Phase 84 browser viewer UI layer",
        "Browser Viewer",
        "Chrome, Brave, and Edge",
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

    Write-TraceDeckLog -Level "INFO" -Message "Phase 84 modern UI smoke passed addr=$Addr browser_rows=$($activity.summary.total) delivery_routes=$($assurance.summary.routes_total)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
