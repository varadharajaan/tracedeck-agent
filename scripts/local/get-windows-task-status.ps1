param(
    [string]$TaskName = "\TraceDeck\TraceDeck Agent",
    [switch]$AllowMissing,
    [string]$OutputPath = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-windows-task-status" -LogRoot "logs/local/service" | Out-Null

function Write-TraceDeckTaskStatus {
    param([pscustomobject]$Status)

    $json = $Status | ConvertTo-Json -Depth 4
    Write-TraceDeckLog -Level "INFO" -Message $json
    if ($OutputPath) {
        $resolvedOutputPath = if ([System.IO.Path]::IsPathRooted($OutputPath)) {
            [System.IO.Path]::GetFullPath($OutputPath)
        }
        else {
            [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $OutputPath))
        }
        $parent = Split-Path -Parent $resolvedOutputPath
        New-Item -ItemType Directory -Force -Path $parent | Out-Null
        Set-Content -Path $resolvedOutputPath -Value $json -Encoding UTF8
    }
    Write-Output $json
}

try {
    $task = Get-ScheduledTask -TaskName (Split-Path -Leaf $TaskName) -TaskPath ((Split-Path -Parent $TaskName) + "\") -ErrorAction SilentlyContinue
    if (-not $task) {
        $message = "Scheduled task not found: $TaskName"
        if ($AllowMissing) {
            Write-TraceDeckLog -Level "WARN" -Message $message
            Write-TraceDeckTaskStatus -Status ([pscustomobject]@{
                task_name = $TaskName
                present = $false
                state = "missing"
                last_run_time = $null
                last_task_result = $null
                next_run_time = $null
                number_of_missed_runs = $null
            })
            Complete-TraceDeckScriptLog
            return
        }
        throw $message
    }

    $info = Get-ScheduledTaskInfo -TaskName $task.TaskName -TaskPath $task.TaskPath
    $status = [pscustomobject]@{
        task_name = $TaskName
        present = $true
        state = $task.State
        last_run_time = $info.LastRunTime
        last_task_result = $info.LastTaskResult
        next_run_time = $info.NextRunTime
        number_of_missed_runs = $info.NumberOfMissedRuns
    }

    Write-TraceDeckTaskStatus -Status $status
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
