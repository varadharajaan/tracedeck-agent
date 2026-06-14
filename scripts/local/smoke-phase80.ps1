param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase80" -LogRoot "logs/local/smoke" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 80 Lambda frontend contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-contract.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 80 Lambda frontend visual quality" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-visual.ps1 -OutputRoot "data/local/lambda-frontend-visual/phase80-smoke"
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 80 devctl cloud visual shortcut" -Command {
        python ./devctl.py cloud visual
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 80 Lambda Cloud Admin local smoke passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
