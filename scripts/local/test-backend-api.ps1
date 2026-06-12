param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-backend-api" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "go test backend api" -Command {
        go test ./backend/...
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
