param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase96" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Post-merge verification wrapper" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-postmerge.ps1 -PhaseTarget phase95 -SkipGitHub -AllowContentDiff
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
