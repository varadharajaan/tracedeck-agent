param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase82" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Compile devctl, visual checkers, and Lambda frontend" -Command {
        python -B -m py_compile ./devctl.py ./scripts/tools/dashboard_visual_quality_check.py ./scripts/tools/lambda_frontend_visual_check.py ./sam-app/frontend_function/app.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard DOM contract guard" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase82/dashboard"
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser activity JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -DashboardPath "backend/internal/api/web/browser_activity.html" -OutputRoot "data/local/dashboard-js-check/phase82/browser-activity"
    }
    Invoke-TraceDeckLoggedCommand -Label "Backend API tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 82 modern admin UI smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase82.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 82 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase82.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Clean Python bytecode artifacts" -Command {
        foreach ($relativePath in @("__pycache__", "scripts/tools/__pycache__", "sam-app/frontend_function/__pycache__")) {
            $bytecodePath = Join-Path $script:TraceDeckRepoRoot $relativePath
            if (Test-Path -LiteralPath $bytecodePath) {
                Remove-Item -LiteralPath $bytecodePath -Recurse -Force
            }
        }
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
