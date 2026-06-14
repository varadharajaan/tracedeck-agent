param(
    [string]$Addr = "127.0.0.1:18233"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase81" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase81/$timestamp"
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

    Invoke-TraceDeckLoggedCommand -Label "Read Phase 81 navigator labels" -Command {
        $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
        foreach ($expected in @(
            "Workspace Navigator",
            "class=""command-label"">Premium Operations</span>",
            "class=""command-label"">Deployment Readiness</span>",
            "class=""command-label"">Customer Control Room</span>",
            "class=""command-label"">Customer Success Packet</span>",
            "class=""command-label"">Provider Setup</span>",
            "class=""command-label"">Paid Operations</span>",
            "class=""command-label"">Delivery Assurance</span>",
            "class=""command-label"">Trust &amp; Consent</span>",
            "class=""command-meta"">waiting</span>"
        )) {
            if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
                throw "Expected Phase 81 navigator marker '$expected'."
            }
        }
        foreach ($forbidden in @(
            "data-jump-target=""deployment-readiness-section"">Deploy<span",
            "data-jump-target=""customer-control-section"">Control<span",
            "data-jump-target=""notification-provider-setup-section"">Setup<span",
            "data-jump-target=""paid-ops-section"">Paid Ops<span",
            "data-jump-target=""delivery-assurance-section"">Assurance<span",
            "data-jump-target=""premium-notification-section"">Notification Pro<span"
        )) {
            if ($dashboard.Content -match [regex]::Escape($forbidden)) {
                throw "Found stale Phase 81 shortcut marker '$forbidden'."
            }
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 81 product visual quality contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -BaseUrl $baseUrl -OutputRoot "data/local/dashboard-visual-quality/phase81-smoke"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 81 navigator clarity smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
