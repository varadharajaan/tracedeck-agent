param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase85" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Compile devctl and quality scripts" -Command {
        python -B -m py_compile ./devctl.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 85 strict Go quality gates" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-go-quality-gates.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 85 runtime Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase85.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Runtime doctor local/cloud assurance" -Command {
        python ./devctl.py doctor --no-cloud-refresh
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
