param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase106" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    Invoke-TraceDeckLoggedCommand -Label "Python devctl syntax check" -Command {
        python -m py_compile ./devctl.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase ledger contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-phase-ledger.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Refresh phase ledger output" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-phase-ledger.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Cleanup generated root artifacts" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/clean-root-generated.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Root artifact re-check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
