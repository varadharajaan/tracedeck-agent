param(
    [string]$Addr = "127.0.0.1:18229"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase76" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase76/$timestamp"
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

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 76 dashboard product UI markers" -Command {
        $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
        foreach ($expected in @(
            "TraceDeck Command Center",
            "Browser Activity",
            "Light mode",
            "server-status-light",
            "dashboard-page-nav",
            "overflow-x: clip",
            "TraceDeck keeps it in session storage only"
        )) {
            if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 76 dashboard marker '$expected'."
            }
        }
        foreach ($forbidden in @("button-icon", ">B<", ">C<", ">T<", ">R<", "Browser{", "Center{")) {
            if ($dashboard.Content.Contains($forbidden)) {
                throw "Dashboard still contains unprofessional marker '$forbidden'."
            }
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 76 browser activity product UI markers" -Command {
        $browserPage = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/browser-activity"
        foreach ($expected in @(
            "TraceDeck Browser Activity",
            "Command Center",
            "Light mode",
            "Browser Activity Viewer",
            "Browser Domain Activity",
            "overflow-x: clip"
        )) {
            if ($browserPage.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 76 browser page marker '$expected'."
            }
        }
        foreach ($forbidden in @("button-icon", ">B<", ">C<", ">T<", ">R<", "Browser{", "Center{")) {
            if ($browserPage.Content.Contains($forbidden)) {
                throw "Browser Activity page still contains unprofessional marker '$forbidden'."
            }
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 76 screenshot-free layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-layout/phase76-smoke"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 76 UI revamp smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
