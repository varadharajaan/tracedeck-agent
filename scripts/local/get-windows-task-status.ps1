param(
    [string]$TaskName = "\TraceDeck\TraceDeck Agent",
    [switch]$AllowMissing
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-windows-task-status" -LogRoot "logs/local/service" | Out-Null

try {
    $task = Get-ScheduledTask -TaskName (Split-Path -Leaf $TaskName) -TaskPath ((Split-Path -Parent $TaskName) + "\") -ErrorAction SilentlyContinue
    if (-not $task) {
        $message = "Scheduled task not found: $TaskName"
        if ($AllowMissing) {
            Write-TraceDeckLog -Level "WARN" -Message $message
            Complete-TraceDeckScriptLog
            return
        }
        throw $message
    }

    $info = Get-ScheduledTaskInfo -TaskName $task.TaskName -TaskPath $task.TaskPath
    $status = [pscustomobject]@{
        task_name = $TaskName
        state = $task.State
        last_run_time = $info.LastRunTime
        last_task_result = $info.LastTaskResult
        next_run_time = $info.NextRunTime
        number_of_missed_runs = $info.NumberOfMissedRuns
    }

    $json = $status | ConvertTo-Json -Depth 4
    Write-TraceDeckLog -Level "INFO" -Message $json
    Write-Output $json
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
