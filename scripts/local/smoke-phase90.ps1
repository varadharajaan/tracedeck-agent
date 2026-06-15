param(
    [string]$Addr = "127.0.0.1:18247"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase90" -LogRoot "logs/local/smoke" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 90 devctl runtime doctor provenance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-devctl-runtime-doctor.ps1 -Addr $Addr
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 90 smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
