param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-software-inventory-collector" -LogRoot "logs/local/tests" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Software inventory collector Go tests" -Command {
        go test ./agent/internal/platform ./agent/internal/collector/software ./agent/internal/alert ./agent/internal/app ./agent/internal/config
    }

    Write-TraceDeckLog -Level "INFO" -Message "Software inventory collector contract passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
