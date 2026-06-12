param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase28" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Regenerate policy schema" -Command {
        go run ./agent/cmd/tracedeck-agent schema --out ./docs/schema/policy-v1alpha1.schema.json
    }

    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Backend API tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Agent tests" -Command {
        go test ./agent/...
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 28 agent telemetry ingest smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase28.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 28 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase28.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Cross-platform build verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-cross-platform-build.ps1 -BuildRoot "data/local/build/phase28"
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
