param(
    [string]$Addr = "127.0.0.1:18235"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase82" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase82/$timestamp"
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

try {
    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 82 polished dashboard shell markers" -Command {
        $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
        foreach ($expected in @(
            "Phase 82 product polish",
            "<span class=""brand-mark"" aria-hidden=""true""><span></span><span></span><span></span></span>",
            "class=""command-label"">Deployment Readiness</span>",
            "class=""command-label"">Customer Control Room</span>",
            "class=""command-label"">Delivery Assurance</span>",
            "class=""command-meta"">waiting</span>",
            "body.theme-dark",
            "server-status-light"
        )) {
            if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 82 dashboard marker '$expected'."
            }
        }
        foreach ($forbidden in @(
            "<span class=""brand-mark"" aria-hidden=""true"">TD</span>",
            "Browser{",
            "Center{",
            "[B]",
            "{C}",
            "Rev Ops",
            "Notif Rev",
            "Notify Pro"
        )) {
            if ($dashboard.Content -match [regex]::Escape($forbidden)) {
                throw "Found stale Phase 82 dashboard marker '$forbidden'."
            }
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 82 polished Browser Activity markers" -Command {
        $browser = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/browser-activity"
        foreach ($expected in @(
            "Browser Viewer",
            "Browser Activity Viewer",
            "Phase 82 product polish",
            "<span class=""brand-mark"" aria-hidden=""true""><span></span><span></span><span></span></span>",
            "Chrome, Brave, and Edge",
            "server-status-light"
        )) {
            if ($browser.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 82 browser marker '$expected'."
            }
        }
        foreach ($forbidden in @(
            "<span class=""brand-mark"" aria-hidden=""true"">TD</span>",
            "Browser{",
            "Center{",
            "[B]",
            "{C}",
            "raw_url",
            "page_title"
        )) {
            if ($browser.Content -match [regex]::Escape($forbidden)) {
                throw "Found stale Phase 82 browser marker '$forbidden'."
            }
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 82 dashboard visual quality contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-visual-quality/phase82-smoke"
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 82 dashboard theme contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-theme/phase82-smoke"
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 82 Lambda frontend visual contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-visual.ps1 -OutputRoot "data/local/lambda-frontend-visual/phase82-smoke"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 82 modern admin UI polish smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
