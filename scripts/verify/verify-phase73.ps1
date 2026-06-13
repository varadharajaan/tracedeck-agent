param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase73" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard DOM contract guard" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Backend API tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Agent tests" -Command {
        go test ./agent/...
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase73/dashboard"
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser activity JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -DashboardPath "backend/internal/api/web/browser_activity.html" -OutputRoot "data/local/dashboard-js-check/phase73/browser-activity"
    }
    Invoke-TraceDeckLoggedCommand -Label "Lambda frontend contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 73 local provenance smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase73.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 73 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase73.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 73 live server provenance readback" -Command {
        python ./devctl.py test live
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 73 cloud Lambda provenance Newman" -Command {
        python ./devctl.py cloud newman
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
