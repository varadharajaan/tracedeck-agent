param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase95" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 94 regression gate" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase94.ps1
    }

    $goCacheRoot = Join-Path $script:TraceDeckRepoRoot "data/local/go-build-cache"
    $goTmpRoot = Join-Path $script:TraceDeckRepoRoot "data/local/go-tmp"
    if (-not (Test-Path $goCacheRoot)) {
        throw "Expected repo-local Go build cache at $goCacheRoot"
    }
    if (-not (Test-Path $goTmpRoot)) {
        throw "Expected repo-local Go temp dir at $goTmpRoot"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Go cache uses repo-local paths: GOCACHE=$env:GOCACHE GOTMPDIR=$env:GOTMPDIR"

    Invoke-TraceDeckLoggedCommand -Label "Root artifact re-check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
