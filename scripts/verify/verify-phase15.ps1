param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase15" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Service manager dry-run tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-service-manager.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Existing service manifest render verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1 -OutputRoot "data/local/service-manifests/phase15"
    }

    Invoke-TraceDeckLoggedCommand -Label "Windows task template verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-windows-task-template.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Cross-platform build verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-cross-platform-build.ps1 -BuildRoot "data/local/build/phase15"
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
