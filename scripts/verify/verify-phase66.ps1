Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase66" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Playwright Python setup" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/setup/install-playwright-python.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Render Phase 66 service manifests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1 -OutputRoot "data/local/service-manifests/phase66"
    }
    Invoke-TraceDeckLoggedCommand -Label "Windows task template verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-windows-task-template.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Autostart assurance verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-autostart-assurance.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Service manager dry-run verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-service-manager.ps1
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
    Invoke-TraceDeckLoggedCommand -Label "Script logging contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-script-logging.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/phase66"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 66 deployment readiness smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase66.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 66 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase66.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Cross-platform build verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-cross-platform-build.ps1 -BuildRoot "data/local/build/phase66"
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
