param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-platform-support" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "go test platform support" -Command {
        go test ./agent/internal/platform ./agent/internal/constants
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
