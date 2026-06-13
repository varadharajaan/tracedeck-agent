param(
    [string]$Addr = "127.0.0.1:18222"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase73" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase73/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase73-smoke/$timestamp"

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

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "sourceBadge",
        "theme-toggle-button",
        "server-status-light",
        "dashboard-page-nav",
        "browser-activity-button"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 73 dashboard marker '$expected'."
        }
    }

    $browserPage = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/browser-activity"
    foreach ($expected in @(
        "TraceDeck Browser Activity",
        "<th>Source</th>",
        "sourceBadge",
        "metadata-only guard"
    )) {
        if ($browserPage.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 73 browser marker '$expected'."
        }
    }

    $viewer = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/family-varadha/browser-activity?limit=25"
    if ($viewer.summary.total -lt 3 -or $viewer.items.Count -lt 3) {
        throw "Expected Phase 73 browser activity rows."
    }
    foreach ($item in $viewer.items) {
        if ([string]::IsNullOrWhiteSpace($item.source_kind) -or [string]::IsNullOrWhiteSpace($item.evidence_scope) -or [string]::IsNullOrWhiteSpace($item.evidence_detail)) {
            throw "Expected browser provenance fields on every row."
        }
    }

    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices/demo-study-laptop/alert-deliveries"
    if ($deliveries.count -lt 1 -or $deliveries.items[0].source_kind -ne "demo_seed" -or $deliveries.items[0].evidence_scope -ne "delivery_proof") {
        throw "Expected demo delivery provenance proof."
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 73 dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 73 smoke passed addr=$Addr rows=$($viewer.summary.total) deliveries=$($deliveries.count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
