param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase110" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    Invoke-TraceDeckLoggedCommand -Label "Python devctl syntax check" -Command {
        python -m py_compile ./devctl.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Active window collector contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-active-window-collector.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Policy schema drift check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-policy-schema.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Agent package tests" -Command {
        go test ./agent/...
    }
    Invoke-TraceDeckLoggedCommand -Label "Cross-platform build" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-cross-platform-build.ps1 -BuildRoot "data/local/build/phase110"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 110 foreground app smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase110.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 110 Newman" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase110.ps1
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
