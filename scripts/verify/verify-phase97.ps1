param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase97" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 96 regression gate" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase96.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Runtime summary contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-summary.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Root artifact check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "git diff --check" -Command {
        git diff --check
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
