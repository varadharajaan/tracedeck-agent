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
. (Join-Path $PSScriptRoot "..\lib\backend-task-status.ps1")
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
    if (-not (Test-TraceDeckBackendTaskStatusAcceptable -Status $status)) {
        $reason = Get-TraceDeckBackendTaskStatusFailureReason -Status $status
        throw "Scheduled backend task is not healthy: $reason`: $($status | ConvertTo-Json -Depth 8)"
    }
    if ($status.task_state -eq $script:TraceDeckTaskStateInaccessible) {
        Write-TraceDeckLog -Level "WARN" -Message "Scheduler readback was denied, but runtime proof is healthy. status=$($status | ConvertTo-Json -Depth 8)"
    }

    $taskLeaf = Split-Path -Leaf $TaskName
    $taskParent = (Split-Path -Parent $TaskName) + "\"
    $task = Get-ScheduledTask -TaskName $taskLeaf -TaskPath $taskParent -ErrorAction Stop
    $action = $task.Actions | Select-Object -First 1
    if ($action.Execute -notmatch "wscript\.exe$" -or $action.Arguments -notmatch "run-agent-task-hidden\.vbs" -or $action.Arguments -notmatch "run-backend-dev-task\.ps1") {
        throw "Scheduled backend task should use the hidden wscript launcher, got execute=$($action.Execute) arguments=$($action.Arguments)"
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
        if (-not (Test-TraceDeckBackendTaskStatusAcceptable -Status $stableStatus)) {
            $reason = Get-TraceDeckBackendTaskStatusFailureReason -Status $stableStatus
            throw "Scheduled backend runtime did not survive stability window: $reason`: $($stableStatus | ConvertTo-Json -Depth 8)"
        }
        if ($stableStatus.task_state -eq $script:TraceDeckTaskStateInaccessible) {
            Write-TraceDeckLog -Level "WARN" -Message "Scheduler readback remained denied after stability window, but runtime proof is healthy."
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
