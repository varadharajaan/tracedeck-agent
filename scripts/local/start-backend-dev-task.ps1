param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$TaskName = "\TraceDeck\TraceDeck Backend Dev",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$DataPath = "data/local/backend/backend-state.json",
    [string]$ExePath = "data/local/backend/tracedeck-dashboard-demo.exe",
    [string]$ReadyPath = "data/local/backend/backend-task-ready.json",
    [string]$UserId = "$env:USERDOMAIN\$env:USERNAME",
    [switch]$ForceRegister
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "start-backend-dev-task" -LogRoot "logs/local/backend" | Out-Null

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

function Get-TraceDeckScheduledTask {
    param([string]$Name)

    $parts = Split-TraceDeckTaskName -Name $Name
    return Get-ScheduledTask -TaskPath $parts.Path -TaskName $parts.Name -ErrorAction SilentlyContinue
}

function Wait-TraceDeckBackend {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(75)
    while ((Get-Date) -lt $deadline) {
        try {
            $health = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/health" -TimeoutSec 2
            if ($health.status -eq "ok") {
                return
            }
        }
        catch {
            Start-Sleep -Milliseconds 500
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Scheduled backend did not become healthy at $BaseUrl"
}

try {
    $baseUrl = "http://$Addr"
    $exeFullPath = Resolve-TraceDeckPath -PathValue $ExePath
    $runnerPath = Resolve-TraceDeckPath -PathValue "scripts/local/run-backend-dev-task.ps1"
    $hiddenLauncherPath = Resolve-TraceDeckPath -PathValue "scripts/local/run-agent-task-hidden.vbs"
    $wscriptPath = if ($env:WINDIR) {
        Join-Path $env:WINDIR "System32\wscript.exe"
    }
    else {
        "wscript.exe"
    }
    $pidFullPath = Resolve-TraceDeckPath -PathValue $PidPath
    $dataFullPath = Resolve-TraceDeckPath -PathValue $DataPath
    $readyFullPath = Resolve-TraceDeckPath -PathValue $ReadyPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $exeFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $pidFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dataFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $readyFullPath) | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Stop stale backend dev listener before scheduled start" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $PidPath -Addr $Addr
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend dev executable for scheduled task" -Command {
        go build -trimpath -o $exeFullPath ./backend/cmd/tracedeck-backend
    }

    $taskParts = Split-TraceDeckTaskName -Name $TaskName
    $existing = Get-TraceDeckScheduledTask -Name $TaskName
    if (-not $existing -or $ForceRegister) {
        $argument = "`"$hiddenLauncherPath`" `"-File`" `"$runnerPath`" `"-Addr`" `"$Addr`" `"-PidPath`" `"$PidPath`" `"-DataPath`" `"$DataPath`" `"-ExePath`" `"$ExePath`" `"-ReadyPath`" `"$ReadyPath`""
        $action = New-ScheduledTaskAction -Execute $wscriptPath -Argument $argument -WorkingDirectory $script:TraceDeckRepoRoot
        $trigger = New-ScheduledTaskTrigger -AtLogOn -User $UserId
        $principal = New-ScheduledTaskPrincipal -UserId $UserId -LogonType Interactive -RunLevel Limited
        $settings = New-ScheduledTaskSettingsSet `
            -AllowStartIfOnBatteries `
            -DontStopIfGoingOnBatteries `
            -StartWhenAvailable `
            -MultipleInstances IgnoreNew `
            -ExecutionTimeLimit ([TimeSpan]::Zero)
        $settings.Hidden = $true
        $task = New-ScheduledTask -Action $action -Trigger $trigger -Principal $principal -Settings $settings

        Invoke-TraceDeckLoggedCommand -Label "Register backend dev scheduled task" -Command {
            Register-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name -InputObject $task -Force | Out-Null
        }
    }

    $registered = Get-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name -ErrorAction Stop
    $registeredAction = $registered.Actions | Select-Object -First 1
    if ($registeredAction.Execute -notmatch "wscript\.exe$" -or $registeredAction.Arguments -notmatch "run-agent-task-hidden\.vbs" -or $registeredAction.Arguments -notmatch "run-backend-dev-task\.ps1") {
        throw "Backend scheduled task is not using the silent hidden launcher."
    }

    Invoke-TraceDeckLoggedCommand -Label "Start backend dev scheduled task" -Command {
        Start-ScheduledTask -TaskPath $taskParts.Path -TaskName $taskParts.Name
    }

    Invoke-TraceDeckLoggedCommand -Label "Wait for scheduled backend health" -Command {
        Wait-TraceDeckBackend -BaseUrl $baseUrl
    }

    Write-TraceDeckLog -Level "INFO" -Message "Scheduled backend dev server ready: $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
