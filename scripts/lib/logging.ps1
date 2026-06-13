param()

Set-StrictMode -Version Latest

function Initialize-TraceDeckScriptLog {
    param(
        [string]$Name,
        [string]$LogRoot = "logs/local"
    )

    $repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
    $logDir = Join-Path $repoRoot $LogRoot
    New-Item -ItemType Directory -Force -Path $logDir | Out-Null

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $script:TraceDeckLogFile = Join-Path $logDir "$Name-$timestamp-$PID.log"
    $script:TraceDeckRepoRoot = $repoRoot

    Update-TraceDeckProcessPath
    Write-TraceDeckLog -Level "INFO" -Message "Started script '$Name' from $repoRoot"
    return $script:TraceDeckLogFile
}

function Update-TraceDeckProcessPath {
    $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $goUserBin = Join-Path $env:USERPROFILE "go\bin"
    $paths = @($machinePath, $userPath, "C:\Program Files\Go\bin", $goUserBin) |
        Where-Object { $_ -and $_.Trim().Length -gt 0 }
    $env:Path = ($paths -join ";")
}

function Write-TraceDeckLog {
    param(
        [ValidateSet("TRACE", "DEBUG", "INFO", "WARN", "ERROR")]
        [string]$Level,
        [string]$Message
    )

    if (-not $script:TraceDeckLogFile) {
        throw "TraceDeck script log is not initialized."
    }

    $line = "{0} [{1}] {2}" -f (Get-Date -Format "o"), $Level, $Message
    Add-Content -Path $script:TraceDeckLogFile -Value $line
    Write-Host $line
}

function Invoke-TraceDeckLoggedCommand {
    param(
        [string]$Label,
        [scriptblock]$Command
    )

    Write-TraceDeckLog -Level "INFO" -Message "Starting: $Label"
    $previousErrorActionPreference = $ErrorActionPreference
    try {
        $ErrorActionPreference = "Continue"
        $global:LASTEXITCODE = 0
        & $Command 2>&1 | ForEach-Object {
            $message = ($_ | Out-String).Trim()
            if ($message.Length -gt 0) {
                Write-TraceDeckLog -Level "DEBUG" -Message $message
            }
        }
        if ($LASTEXITCODE -ne $null -and $LASTEXITCODE -ne 0) {
            throw "$Label failed with exit code $LASTEXITCODE"
        }
        Write-TraceDeckLog -Level "INFO" -Message "Completed: $Label"
    }
    catch {
        Write-TraceDeckLog -Level "ERROR" -Message "$Label failed: $($_.Exception.Message)"
        throw
    }
    finally {
        $ErrorActionPreference = $previousErrorActionPreference
    }
}

function Complete-TraceDeckScriptLog {
    Write-TraceDeckLog -Level "INFO" -Message "Completed script. Log file: $script:TraceDeckLogFile"
}
