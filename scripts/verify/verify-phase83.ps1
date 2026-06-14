param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase83" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Compile devctl, visual checkers, and Lambda frontend" -Command {
        python -B -m py_compile ./devctl.py ./scripts/tools/dashboard_visual_quality_check.py ./scripts/tools/lambda_frontend_visual_check.py ./sam-app/frontend_function/app.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Agent tests" -Command {
        go test ./agent/...
    }
    Invoke-TraceDeckLoggedCommand -Label "Backend API and store tests" -Command {
        go test ./backend/internal/api ./backend/internal/store
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard DOM contract guard" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase83/dashboard"
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser activity JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -DashboardPath "backend/internal/api/web/browser_activity.html" -OutputRoot "data/local/dashboard-js-check/phase83/browser-activity"
    }
    Invoke-TraceDeckLoggedCommand -Label "Lambda frontend contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 83 heartbeat smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase83.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 83 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase83.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard visual quality contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -OutputRoot "data/local/dashboard-visual-quality/phase83"
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard theme contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 -OutputRoot "data/local/dashboard-theme/phase83"
    }
    Invoke-TraceDeckLoggedCommand -Label "Lambda frontend visual contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-visual.ps1 -OutputRoot "data/local/lambda-frontend-visual/phase83"
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
