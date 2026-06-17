param(
    [string]$TaskPath = "\TraceDeck\",
    [string]$NamePrefix = "TraceDeck Backend Dev Phase",
    [switch]$WhatIfOnly
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "cleanup-stale-backend-dev-tasks" -LogRoot "logs/local/service" | Out-Null

function Stop-TraceDeckPidFile {
    param([string]$PidPath)

    if ([string]::IsNullOrWhiteSpace($PidPath)) {
        return
    }

    $resolvedPidPath = if ([System.IO.Path]::IsPathRooted($PidPath)) {
        [System.IO.Path]::GetFullPath($PidPath)
    }
    else {
        [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PidPath))
    }

    if (-not (Test-Path -LiteralPath $resolvedPidPath)) {
        return
    }

    $pidText = (Get-Content -LiteralPath $resolvedPidPath -Raw).Trim()
    if ($pidText -notmatch "^\d+$") {
        return
    }

    $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
    if (-not $process) {
        return
    }

    Write-TraceDeckLog -Level "INFO" -Message "Stopping stale backend child process pid=$pidText pid_path=$resolvedPidPath"
    if (-not $WhatIfOnly) {
        Stop-Process -Id ([int]$pidText) -Force
    }
}

try {
    $tasks = @(Get-ScheduledTask -TaskPath $TaskPath -ErrorAction SilentlyContinue |
        Where-Object { $_.TaskName.StartsWith($NamePrefix, [System.StringComparison]::OrdinalIgnoreCase) })

    $results = @()
    foreach ($task in $tasks) {
        $pidPath = ""
        $arguments = ($task.Actions | Select-Object -First 1).Arguments
        if ($arguments -match "-PidPath\s+`"([^`"]+)`"") {
            $pidPath = $Matches[1]
        }

        Write-TraceDeckLog -Level "INFO" -Message "Cleaning stale backend task $($task.TaskPath)$($task.TaskName) state=$($task.State)"
        if (-not $WhatIfOnly) {
            if ($task.State -eq "Running") {
                Stop-ScheduledTask -TaskName $task.TaskName -TaskPath $task.TaskPath -ErrorAction SilentlyContinue
                Start-Sleep -Milliseconds 500
            }
            Stop-TraceDeckPidFile -PidPath $pidPath
            Unregister-ScheduledTask -TaskName $task.TaskName -TaskPath $task.TaskPath -Confirm:$false
        }

        $results += [pscustomobject]@{
            task_name = "$($task.TaskPath)$($task.TaskName)"
            state = [string]$task.State
            pid_path = $pidPath
            removed = -not [bool]$WhatIfOnly
        }
    }

    $results | ConvertTo-Json -Depth 4
    Write-TraceDeckLog -Level "INFO" -Message "Stale backend task cleanup completed count=$($results.Count) what_if=$([bool]$WhatIfOnly)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
