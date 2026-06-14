param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase80" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Compile devctl, Lambda frontend, and visual checker" -Command {
        python -B -m py_compile ./devctl.py ./sam-app/frontend_function/app.py ./scripts/tools/lambda_frontend_visual_check.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 80 local Lambda smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase80.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Deploy SAM Lambda frontend for Phase 80" -Command {
        python ./devctl.py sam deploy
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 80 Lambda Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase80.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 80 runtime doctor with cloud" -Command {
        python ./devctl.py doctor
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
