param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "format-go" -LogRoot "logs/local/format" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Format Go source" -Command {
        gofmt -w ./agent
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
