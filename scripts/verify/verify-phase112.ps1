param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase112" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    Invoke-TraceDeckLoggedCommand -Label "Python script syntax check" -Command {
        python -m py_compile ./devctl.py ./scripts/tools/dashboard_delivery_ui_check.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Backend API tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard delivery card UI contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-dashboard-delivery-ui.ps1 `
            -Addr "127.0.0.1:18291" `
            -PidPath "data/local/backend/phase112-delivery-ui.pid" `
            -DataPath "data/local/backend/phase112-delivery-ui-state.json"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 112 local indicator smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase112.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 112 Newman" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase112.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Contract completion audit refresh" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-contract-completion-audit.ps1
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
