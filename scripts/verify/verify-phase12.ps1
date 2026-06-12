param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase12" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Device health focused tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-device-health.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Backend API and store tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 12 device health and dashboard smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase12.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 12 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase12.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Cross-platform build verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-cross-platform-build.ps1 -BuildRoot "data/local/build/phase12"
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
