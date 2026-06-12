param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-device-health" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "go test agent health collector and app runner" -Command {
        go test ./agent/internal/collector/health ./agent/internal/app
    }

    Invoke-TraceDeckLoggedCommand -Label "go test backend health API and store" -Command {
        go test ./backend/internal/api ./backend/internal/store
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
