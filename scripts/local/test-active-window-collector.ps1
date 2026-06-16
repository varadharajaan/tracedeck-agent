param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-active-window-collector" -LogRoot "logs/local/tests" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Active window collector Go tests" -Command {
        go test ./agent/internal/platform ./agent/internal/collector/activewindow ./agent/internal/alert ./agent/internal/app
    }

    Write-TraceDeckLog -Level "INFO" -Message "Active window collector contract passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
