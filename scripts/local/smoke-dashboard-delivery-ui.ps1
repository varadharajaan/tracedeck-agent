param(
    [string]$Addr = "127.0.0.1:18291",
    [string]$PidPath = "data/local/backend/dashboard-delivery-ui.pid",
    [string]$DataPath = "data/local/backend/dashboard-delivery-ui-state.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-dashboard-delivery-ui" -LogRoot "logs/local/smoke" | Out-Null

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting dashboard delivery UI backend addr=$Addr pid_path=$PidPath"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1 `
        -Addr $Addr `
        -PidPath $PidPath `
        -DataPath $DataPath
    if ($LASTEXITCODE -ne $null -and $LASTEXITCODE -ne 0) {
        throw "start-dashboard-demo.ps1 failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Dashboard delivery UI backend started addr=$Addr"

    try {
        Invoke-TraceDeckLoggedCommand -Label "Dashboard delivery card UI contract" -Command {
            powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-delivery-ui.ps1 `
                -BaseUrl "http://$Addr" `
                -OutputRoot "data/local/dashboard-delivery-ui/smoke"
        }
    }
    finally {
        Invoke-TraceDeckLoggedCommand -Label "Stop dashboard delivery UI backend" -Command {
            powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 `
                -PidPath $PidPath `
                -Addr $Addr
        }
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
