param(
    [switch]$Force
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "install-sam-cli" -LogRoot "logs/local/cloud" | Out-Null

function Find-SamCli {
    $command = Get-Command "sam" -ErrorAction SilentlyContinue
    if ($command) {
        return $command.Source
    }
    $candidateRoots = @(
        "$env:ProgramFiles\Amazon\AWSSAMCLI\bin\sam.exe",
        "$env:LOCALAPPDATA\Programs\Amazon\AWSSAMCLI\bin\sam.exe"
    )
    foreach ($candidate in $candidateRoots) {
        if (Test-Path -LiteralPath $candidate) {
            return $candidate
        }
    }
    return ""
}

try {
    $samPath = Find-SamCli
    if ($samPath -and -not $Force) {
        Write-TraceDeckLog -Level "INFO" -Message "SAM CLI already available: $samPath"
        & $samPath --version
        Complete-TraceDeckScriptLog
        return
    }

    $winget = Get-Command "winget" -ErrorAction SilentlyContinue
    if (-not $winget) {
        throw "winget is required to install AWS SAM CLI automatically."
    }

    Invoke-TraceDeckLoggedCommand -Label "Install AWS SAM CLI with winget" -Command {
        winget install --id Amazon.AWS.SAM-CLI -e --silent --accept-package-agreements --accept-source-agreements
    }

    $samPath = Find-SamCli
    if (-not $samPath) {
        throw "SAM CLI install completed but sam.exe was not found on PATH or standard install paths."
    }

    Write-TraceDeckLog -Level "INFO" -Message "SAM CLI installed: $samPath"
    & $samPath --version
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
