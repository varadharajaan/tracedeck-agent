param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
$logPath = Initialize-TraceDeckScriptLog -Name "test-script-logging" -LogRoot "logs/local/test"

try {
    $repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
    $resolvedLog = Resolve-Path $logPath
    $expectedRoot = Join-Path $repoRoot "logs\local"
    if (-not $resolvedLog.Path.StartsWith($expectedRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Script log escaped logs/local: $($resolvedLog.Path)"
    }

    $fileName = Split-Path -Leaf $resolvedLog.Path
    if ($fileName -notmatch "^test-script-logging-\d{8}-\d{6}-$PID\.log$") {
        throw "Script log filename does not include process id: $fileName"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Script logging filename contract passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
