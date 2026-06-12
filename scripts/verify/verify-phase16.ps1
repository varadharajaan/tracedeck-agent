param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase16" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Backend API tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 16 monetisation dashboard smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase16.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 16 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase16.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Cross-platform build verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-cross-platform-build.ps1 -BuildRoot "data/local/build/phase16"
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
