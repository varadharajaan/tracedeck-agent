param(
    [string]$LogRoot = "logs/local/setup"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
$logFile = Initialize-TraceDeckScriptLog -Name "install-go" -LogRoot $LogRoot
$repoRoot = $script:TraceDeckRepoRoot

function Update-CurrentProcessPath {
    $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $env:Path = "$machinePath;$userPath"

    $defaultGoBin = "C:\Program Files\Go\bin"
    if ((Test-Path $defaultGoBin) -and ($env:Path -notlike "*$defaultGoBin*")) {
        $env:Path = "$env:Path;$defaultGoBin"
    }
}

Write-TraceDeckLog -Level "INFO" -Message "TraceDeck Go setup started from $repoRoot"

$existingGo = Get-Command go -ErrorAction SilentlyContinue
if ($existingGo) {
    Write-TraceDeckLog -Level "INFO" -Message "Go already available at $($existingGo.Source)"
    Invoke-TraceDeckLoggedCommand -Label "Verify Go version" -Command { go version }
    Complete-TraceDeckScriptLog
    exit 0
}

$winget = Get-Command winget -ErrorAction SilentlyContinue
if (-not $winget) {
    Write-TraceDeckLog -Level "ERROR" -Message "winget is not available. Install Go manually or install App Installer, then rerun this script."
    exit 1
}

Write-TraceDeckLog -Level "INFO" -Message "winget found at $($winget.Source)"
Invoke-TraceDeckLoggedCommand -Label "Install Go using winget" -Command {
    winget install --id GoLang.Go --exact --accept-package-agreements --accept-source-agreements --silent
}

Update-CurrentProcessPath

$installedGo = Get-Command go -ErrorAction SilentlyContinue
if (-not $installedGo) {
    Write-TraceDeckLog -Level "ERROR" -Message "Go installation completed, but go.exe is still not visible on PATH in this process."
    exit 1
}

Write-TraceDeckLog -Level "INFO" -Message "Go installed at $($installedGo.Source)"
Invoke-TraceDeckLoggedCommand -Label "Verify Go version" -Command { go version }
Write-TraceDeckLog -Level "INFO" -Message "TraceDeck Go setup completed"
Complete-TraceDeckScriptLog
