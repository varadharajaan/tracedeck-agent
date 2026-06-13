param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase74" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Compile devctl and Lambda frontend Python" -Command {
        python -B -m py_compile ./devctl.py ./sam-app/frontend_function/app.py
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
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase74/dashboard"
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser activity JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -DashboardPath "backend/internal/api/web/browser_activity.html" -OutputRoot "data/local/dashboard-js-check/phase74/browser-activity"
    }
    Invoke-TraceDeckLoggedCommand -Label "Lambda frontend contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-lambda-frontend-contract.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 74 runtime doctor smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase74.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 74 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase74.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 74 cloud runtime doctor smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase74.ps1 -Addr "127.0.0.1:18226" -IncludeCloud
    }
    Invoke-TraceDeckLoggedCommand -Label "Clean Python bytecode artifacts" -Command {
        foreach ($relativePath in @("__pycache__", "sam-app/frontend_function/__pycache__")) {
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
