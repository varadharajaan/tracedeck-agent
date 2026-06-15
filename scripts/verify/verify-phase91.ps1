param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase91" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    Invoke-TraceDeckLoggedCommand -Label "Python devctl syntax check" -Command {
        python -m py_compile ./devctl.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 91 scheduled backend task smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase91.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 91 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase91.ps1
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
