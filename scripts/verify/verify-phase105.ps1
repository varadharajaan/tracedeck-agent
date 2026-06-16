param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase105" -LogRoot "logs/local/verify" | Out-Null

try {
    $defaultBackendAddr = "127.0.0.1:18080"

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
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase105"
    }
    Invoke-TraceDeckLoggedCommand -Label "Runtime summary contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-summary.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 105 promotion readiness smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase105.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 105 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase105.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Refresh Phase 105 verification evidence artifact" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1 -Phase "phase105"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Starting: Refresh default backend with Phase 105 endpoint"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -Addr $defaultBackendAddr
    if ($LASTEXITCODE -ne 0) {
        throw "Refresh default backend stop failed with exit code $LASTEXITCODE"
    }
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-backend-dev.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Refresh default backend start failed with exit code $LASTEXITCODE"
    }
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/wait-backend-health.ps1 -Addr $defaultBackendAddr
    if ($LASTEXITCODE -ne 0) {
        throw "Refresh default backend health wait failed with exit code $LASTEXITCODE"
    }
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/refresh-backend-ready-proof.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Refresh default backend ready proof failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Refresh default backend with Phase 105 endpoint"

    Invoke-TraceDeckLoggedCommand -Label "Restore default runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Refresh default operator assurance pack" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-operator-assurance.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Refresh default promotion readiness bundle" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-promotion-readiness.ps1
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
