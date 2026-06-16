param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase99" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    Invoke-TraceDeckLoggedCommand -Label "Python devctl syntax check" -Command {
        python -m py_compile ./devctl.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Backend API tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard DOM contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase99"
    }
    Invoke-TraceDeckLoggedCommand -Label "Runtime summary contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-summary.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Initial verification evidence artifact" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 -Phase "phase99"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 99 verification evidence smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase99.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 99 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase99.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Refresh Phase 99 verification evidence artifact" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 -Phase "phase99"
    }
    Invoke-TraceDeckLoggedCommand -Label "Restore default runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1
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
