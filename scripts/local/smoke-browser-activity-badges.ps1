param(
    [string]$Addr = "127.0.0.1:18313",
    [string]$PidPath = "data/local/backend/browser-activity-badges.pid",
    [string]$DataPath = "data/local/backend/browser-activity-badges-state.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-browser-activity-badges" -LogRoot "logs/local/smoke" | Out-Null

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting Browser Activity badge backend addr=$Addr pid_path=$PidPath"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-dashboard-demo.ps1 `
        -Addr $Addr `
        -PidPath $PidPath `
        -DataPath $DataPath
    if ($LASTEXITCODE -ne $null -and $LASTEXITCODE -ne 0) {
        throw "start-dashboard-demo.ps1 failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Browser Activity badge backend started addr=$Addr"

    try {
        Invoke-TraceDeckLoggedCommand -Label "Browser Activity badge layout contract" -Command {
            powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-browser-activity-badges.ps1 `
                -BaseUrl "http://$Addr" `
                -OutputRoot "data/local/browser-activity-badges/smoke"
        }
    }
    finally {
        Invoke-TraceDeckLoggedCommand -Label "Stop Browser Activity badge backend" -Command {
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
