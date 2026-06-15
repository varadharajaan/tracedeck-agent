param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$TaskName = "\TraceDeck\TraceDeck Backend Dev",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$ReadyPath = "data/local/backend/backend-task-ready.json",
    [string]$OutputPath = "data/local/backend/backend-task-status.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-backend-dev-task-status" -LogRoot "logs/local/backend" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

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
    $baseUrl = "http://$Addr"
    $pidFullPath = Resolve-TraceDeckPath -PathValue $PidPath
    $readyFullPath = Resolve-TraceDeckPath -PathValue $ReadyPath
    $outputFullPath = Resolve-TraceDeckPath -PathValue $OutputPath
    $taskParts = Split-TraceDeckTaskName -Name $TaskName
    $task = $null
    $taskQueryError = ""
    try {
        $task = Get-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name -ErrorAction Stop
    }
    catch {
        $taskQueryError = $_.Exception.Message
    }
    $info = $null
    $taskInfoError = ""
    if ($task) {
        try {
            $info = Get-ScheduledTaskInfo -TaskPath $taskParts.Path -TaskName $taskParts.Name -ErrorAction Stop
        }
        catch {
            $taskInfoError = $_.Exception.Message
        }
    }

    $healthOK = $false
    $healthError = ""
    try {
        $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health" -TimeoutSec 3
        $healthOK = $health.status -eq "ok"
    }
    catch {
        $healthError = $_.Exception.Message
    }

    $backendPid = $null
    $pidRunning = $false
    if (Test-Path -LiteralPath $pidFullPath) {
        $pidText = (Get-Content -Path $pidFullPath -Raw).Trim()
        if ($pidText) {
            $backendPid = [int]$pidText
            $pidRunning = $null -ne (Get-Process -Id $backendPid -ErrorAction SilentlyContinue)
        }
    }

    $ready = $null
    if (Test-Path -LiteralPath $readyFullPath) {
        $ready = Get-Content -Path $readyFullPath -Raw | ConvertFrom-Json
    }

    $status = [pscustomobject]@{
        task_name = $TaskName
        task_present = $null -ne $task
        task_state = if ($task) { [string]$task.State } elseif ($taskQueryError) { "inaccessible" } else { "missing" }
        task_query_error = $taskQueryError
        task_info_error = $taskInfoError
        last_run_time = if ($info) { $info.LastRunTime } else { $null }
        last_task_result = if ($info) { $info.LastTaskResult } else { $null }
        next_run_time = if ($info) { $info.NextRunTime } else { $null }
        base_url = $baseUrl
        health_ok = $healthOK
        health_error = $healthError
        pid = $backendPid
        pid_running = $pidRunning
        runtime_ok = $healthOK -and $pidRunning
        launch_task_verified = $null -ne $task
        ready = $ready
    }

    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputFullPath) | Out-Null
    $json = $status | ConvertTo-Json -Depth 8
    Set-Content -Path $outputFullPath -Value $json -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message $json
    Write-Output $json
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
