param(
    [string]$Addr = "127.0.0.1:18249"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase91" -LogRoot "logs/local/smoke" | Out-Null

try {
    $runRoot = "data/local/backend-task-phase91"
    Invoke-TraceDeckLoggedCommand -Label "Phase 91 scheduled backend dev task smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-dev-task.ps1 `
            -Addr $Addr `
            -TaskName "\TraceDeck\TraceDeck Backend Dev Phase91" `
            -PidPath "$runRoot/tracedeck-backend.pid" `
            -DataPath "$runRoot/backend-state.json" `
            -ExePath "$runRoot/tracedeck-dashboard-demo.exe" `
            -ReadyPath "$runRoot/backend-task-ready.json" `
            -StatusPath "$runRoot/backend-task-status.json" `
            -StabilitySeconds 15
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 91 smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
