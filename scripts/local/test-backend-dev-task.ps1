param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$TaskName = "\TraceDeck\TraceDeck Backend Dev",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$DataPath = "data/local/backend/backend-state.json",
    [string]$ExePath = "data/local/backend/tracedeck-dashboard-demo.exe",
    [string]$ReadyPath = "data/local/backend/backend-task-ready.json",
    [string]$StatusPath = "data/local/backend/backend-task-status.json",
    [int]$StabilitySeconds = 15,
    [switch]$LeaveRunning
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-backend-dev-task" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Start scheduled backend dev task" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-backend-dev-task.ps1 `
            -Addr $Addr `
            -TaskName $TaskName `
            -PidPath $PidPath `
            -DataPath $DataPath `
            -ExePath $ExePath `
            -ReadyPath $ReadyPath `
            -ForceRegister
    }

    Start-Sleep -Seconds 5

    Invoke-TraceDeckLoggedCommand -Label "Query scheduled backend dev task" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
            -Addr $Addr `
            -TaskName $TaskName `
            -PidPath $PidPath `
            -ReadyPath $ReadyPath `
            -OutputPath $StatusPath
    }

    $statusFullPath = Join-Path $script:TraceDeckRepoRoot $StatusPath
    if (-not (Test-Path -LiteralPath $statusFullPath)) {
        throw "Expected backend task status output: $statusFullPath"
    }
    $status = Get-Content -Path $statusFullPath -Raw | ConvertFrom-Json
    if (-not $status.task_present -or -not $status.health_ok -or -not $status.pid_running) {
        throw "Scheduled backend task is not healthy: $($status | ConvertTo-Json -Depth 8)"
    }

    if ($StabilitySeconds -gt 0) {
        Start-Sleep -Seconds $StabilitySeconds
        Invoke-TraceDeckLoggedCommand -Label "Query scheduled backend dev task after stability window" -Command {
            powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
                -Addr $Addr `
                -TaskName $TaskName `
                -PidPath $PidPath `
                -ReadyPath $ReadyPath `
                -OutputPath $StatusPath
        }

        $stableStatus = Get-Content -Path $statusFullPath -Raw | ConvertFrom-Json
        if (-not $stableStatus.runtime_ok) {
            throw "Scheduled backend runtime did not survive stability window: $($stableStatus | ConvertTo-Json -Depth 8)"
        }
        if ($stableStatus.task_state -eq "missing") {
            throw "Scheduled backend task status reported missing without a query error: $($stableStatus | ConvertTo-Json -Depth 8)"
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Live provenance against scheduled backend" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl "http://$Addr"
    }

    Invoke-TraceDeckLoggedCommand -Label "Runtime doctor against scheduled backend" -Command {
        python ./devctl.py --addr $Addr doctor --skip-cloud
    }

    Write-TraceDeckLog -Level "INFO" -Message "Scheduled backend dev task smoke passed addr=$Addr leave_running=$($LeaveRunning.IsPresent)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if (-not $LeaveRunning) {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev-task.ps1 -Addr $Addr -TaskName $TaskName -PidPath $PidPath
    }
}
