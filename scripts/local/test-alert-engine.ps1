param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-alert-engine" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "go test alert engine" -Command {
        go test ./agent/internal/alert
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
