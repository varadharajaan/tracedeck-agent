param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$TaskName = "\TraceDeck\TraceDeck Backend Dev",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [switch]$Unregister
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "stop-backend-dev-task" -LogRoot "logs/local/backend" | Out-Null

function Split-TraceDeckTaskName {
    param([string]$Name)

    $normalized = $Name
    if (-not $normalized.StartsWith("\")) {
        $normalized = "\" + $normalized
    }
    $leaf = Split-Path -Leaf $normalized
    $parent = Split-Path -Parent $normalized
    if ([string]::IsNullOrWhiteSpace($parent) -or $parent -eq "\") {
        $path = "\"
    }
    else {
        $path = $parent.TrimEnd("\") + "\"
    }
    return [pscustomobject]@{ Path = $path; Name = $leaf }
}

try {
    $taskParts = Split-TraceDeckTaskName -Name $TaskName
    $task = Get-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name -ErrorAction SilentlyContinue
    if ($task) {
        Invoke-TraceDeckLoggedCommand -Label "Stop backend dev scheduled task" -Command {
            Stop-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name -ErrorAction SilentlyContinue
        }
    }
    elseif ($Unregister) {
        Write-TraceDeckLog -Level "WARN" -Message "Scheduled task already missing: $TaskName"
    }

    Invoke-TraceDeckLoggedCommand -Label "Stop backend dev listener and pid process" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $PidPath -Addr $Addr
    }

    if ($Unregister -and $task) {
        Invoke-TraceDeckLoggedCommand -Label "Unregister backend dev scheduled task" -Command {
            Unregister-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name -Confirm:$false
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Backend dev scheduled task stop complete task=$TaskName unregister=$($Unregister.IsPresent)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
