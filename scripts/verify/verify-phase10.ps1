param(
    [string]$Addr = "127.0.0.1:18086",
    [string]$PidPath = "data/local/backend/tracedeck-dashboard-demo.pid"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase10" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Start and verify seeded dashboard demo" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1 -Addr $Addr -PidPath $PidPath -StopAfterSeed
    }

    Invoke-TraceDeckLoggedCommand -Label "Root artifact re-check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
