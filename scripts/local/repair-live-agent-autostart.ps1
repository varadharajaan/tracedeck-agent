param(
    [string]$TaskName = "\TraceDeck\TraceDeck Agent",
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$ConfigPath = "data/local/config/tracedeck-live-this-machine.yaml",
    [string]$DataDir = "data/local/agent-live",
    [string]$LogDir = "logs/local/agent-live",
    [string]$OutboxDir = "data/local/outbox-live",
    [string]$PidPath = "data/local/agent-live/tracedeck-agent-live.pid",
    [string]$CollectionInterval = "10m",
    [string]$ExtraArgs = "--archive-dry-run=false --alert-dry-run=false --log-level debug",
    [ValidateSet("HighestAvailable", "LeastPrivilege")]
    [string]$RunLevel = "LeastPrivilege",
    [switch]$SkipBuild
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "repair-live-agent-autostart" -LogRoot "logs/local/service" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Stop-TraceDeckLiveAgent {
    param([string]$ResolvedPidPath)

    if (-not (Test-Path -LiteralPath $ResolvedPidPath)) {
        return
    }

    $pidText = (Get-Content -LiteralPath $ResolvedPidPath -Raw).Trim()
    if ($pidText -notmatch "^\d+$") {
        Remove-Item -LiteralPath $ResolvedPidPath -Force -ErrorAction SilentlyContinue
        return
    }

    $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
    if ($process) {
        Write-TraceDeckLog -Level "INFO" -Message "Stopping previous live agent pid=$pidText before autostart repair"
        Stop-Process -Id ([int]$pidText) -Force
        Start-Sleep -Milliseconds 500
    }
    Remove-Item -LiteralPath $ResolvedPidPath -Force -ErrorAction SilentlyContinue
}

try {
    $resolvedPidPath = Resolve-TraceDeckPath -PathValue $PidPath
    $resolvedTaskXmlPath = Resolve-TraceDeckPath -PathValue "data/local/service-manifests/live/windows/tracedeck-agent-task.xml"

    $taskLeaf = Split-Path -Leaf $TaskName
    $taskParent = (Split-Path -Parent $TaskName) + "\"
    $existingTask = Get-ScheduledTask -TaskName $taskLeaf -TaskPath $taskParent -ErrorAction SilentlyContinue
    if ($existingTask -and $existingTask.State -eq "Running") {
        Write-TraceDeckLog -Level "INFO" -Message "Stopping existing scheduled task $TaskName before repair"
        Stop-ScheduledTask -TaskName $taskLeaf -TaskPath $taskParent -ErrorAction SilentlyContinue
        Start-Sleep -Milliseconds 500
    }

    Stop-TraceDeckLiveAgent -ResolvedPidPath $resolvedPidPath

    $registerArgs = @(
        "-TaskName", $TaskName,
        "-AgentPath", $AgentPath,
        "-ConfigPath", $ConfigPath,
        "-DataDir", $DataDir,
        "-LogDir", $LogDir,
        "-OutboxDir", $OutboxDir,
        "-PidPath", $PidPath,
        "-CollectionInterval", $CollectionInterval,
        "-ExtraArgs", $ExtraArgs,
        "-RunLevel", $RunLevel,
        "-TaskXmlPath", $resolvedTaskXmlPath,
        "-StartAfterRegister",
        "-NoElevate"
    )
    if (-not $SkipBuild) {
        $registerArgs += "-BuildAgent"
    }

    Invoke-TraceDeckLoggedCommand -Label "Register silent live agent scheduled task" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/register-windows-task.ps1 @registerArgs
    }

    $task = Get-ScheduledTask -TaskName $taskLeaf -TaskPath $taskParent
    $command = ($task.Actions | Select-Object -First 1).Execute
    $arguments = ($task.Actions | Select-Object -First 1).Arguments
    if ($command -notmatch "wscript\.exe$" -or $arguments -notmatch "run-agent-task-hidden\.vbs" -or $arguments -notmatch "run-agent-task\.ps1") {
        throw "Live agent task is not using the silent hidden runner."
    }

    [pscustomobject]@{
        status = "repaired"
        task_name = $TaskName
        state = [string]$task.State
        command = $command
        pid_path = Resolve-TraceDeckPath -PathValue $PidPath
        task_xml = $resolvedTaskXmlPath
    } | ConvertTo-Json -Depth 4

    Write-TraceDeckLog -Level "INFO" -Message "Live agent autostart repaired with silent runner task=$TaskName"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
