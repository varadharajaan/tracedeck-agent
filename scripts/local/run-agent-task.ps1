param(
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$ConfigPath = "examples/policies/ai-btech-student.yaml",
    [string]$DataDir = "data/local",
    [string]$LogDir = "logs/local/agent",
    [string]$OutboxDir = "data/local/outbox",
    [string]$CollectionInterval = "10m",
    [string]$PidPath = "",
    [string]$ExtraArgs = "",
    [string]$WorkingDir = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "run-agent-task" -LogRoot "logs/local/agent" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Stop-ExistingAgent {
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
        Write-TraceDeckLog -Level "INFO" -Message "Stopping existing scheduled agent pid=$pidText"
        Stop-Process -Id ([int]$pidText) -Force
        Start-Sleep -Milliseconds 500
    }
    Remove-Item -LiteralPath $ResolvedPidPath -Force -ErrorAction SilentlyContinue
}

function Split-TraceDeckExtraArgs {
    param([string]$Value)

    if ([string]::IsNullOrWhiteSpace($Value)) {
        return @()
    }

    $tokens = [System.Collections.Generic.List[string]]::new()
    $builder = [System.Text.StringBuilder]::new()
    $inQuote = $false
    $quoteChar = [char]0
    $chars = $Value.ToCharArray()

    for ($index = 0; $index -lt $chars.Length; $index++) {
        $ch = $chars[$index]
        if (($ch -eq '"' -or $ch -eq "'") -and (-not $inQuote)) {
            $inQuote = $true
            $quoteChar = $ch
            continue
        }
        if ($inQuote -and $ch -eq $quoteChar) {
            $inQuote = $false
            $quoteChar = [char]0
            continue
        }
        if ([char]::IsWhiteSpace($ch) -and -not $inQuote) {
            if ($builder.Length -gt 0) {
                $tokens.Add($builder.ToString())
                [void]$builder.Clear()
            }
            continue
        }
        [void]$builder.Append($ch)
    }

    if ($builder.Length -gt 0) {
        $tokens.Add($builder.ToString())
    }
    return @($tokens)
}

try {
    $resolvedAgentPath = Resolve-TraceDeckPath -PathValue $AgentPath
    $resolvedConfigPath = Resolve-TraceDeckPath -PathValue $ConfigPath
    $resolvedDataDir = Resolve-TraceDeckPath -PathValue $DataDir
    $resolvedLogDir = Resolve-TraceDeckPath -PathValue $LogDir
    $resolvedOutboxDir = Resolve-TraceDeckPath -PathValue $OutboxDir
    $resolvedWorkingDir = if ([string]::IsNullOrWhiteSpace($WorkingDir)) {
        [System.IO.Path]::GetFullPath($script:TraceDeckRepoRoot)
    }
    else {
        Resolve-TraceDeckPath -PathValue $WorkingDir
    }
    $resolvedPidPath = if ([string]::IsNullOrWhiteSpace($PidPath)) {
        Join-Path $resolvedDataDir "tracedeck-agent.pid"
    }
    else {
        Resolve-TraceDeckPath -PathValue $PidPath
    }

    if (-not (Test-Path -LiteralPath $resolvedAgentPath)) {
        throw "Agent executable does not exist: $resolvedAgentPath"
    }
    if (-not (Test-Path -LiteralPath $resolvedConfigPath)) {
        throw "Agent config does not exist: $resolvedConfigPath"
    }

    New-Item -ItemType Directory -Force -Path $resolvedDataDir | Out-Null
    New-Item -ItemType Directory -Force -Path $resolvedLogDir | Out-Null
    New-Item -ItemType Directory -Force -Path $resolvedOutboxDir | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedPidPath) | Out-Null

    Stop-ExistingAgent -ResolvedPidPath $resolvedPidPath

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $stdoutPath = Join-Path $resolvedLogDir "agent-task-$timestamp.out.log"
    $stderrPath = Join-Path $resolvedLogDir "agent-task-$timestamp.err.log"
    $arguments = @(
        "run",
        "--config", "`"$resolvedConfigPath`"",
        "--data-dir", "`"$resolvedDataDir`"",
        "--log-dir", "`"$resolvedLogDir`"",
        "--outbox-dir", "`"$resolvedOutboxDir`"",
        "--collection-interval", $CollectionInterval,
        "--max-cycles", "0"
    )
    $arguments += Split-TraceDeckExtraArgs -Value $ExtraArgs

    $process = Start-Process -FilePath $resolvedAgentPath `
        -ArgumentList $arguments `
        -WorkingDirectory $resolvedWorkingDir `
        -WindowStyle Hidden `
        -RedirectStandardOutput $stdoutPath `
        -RedirectStandardError $stderrPath `
        -PassThru

    Set-Content -LiteralPath $resolvedPidPath -Value $process.Id -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Started scheduled agent process pid=$($process.Id) stdout=$stdoutPath stderr=$stderrPath"

    Wait-Process -Id $process.Id
    Write-TraceDeckLog -Level "WARN" -Message "Scheduled agent process exited pid=$($process.Id)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
