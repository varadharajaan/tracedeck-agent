param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase89" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 89 focused activity feed provenance tests" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-activity-feed-provenance.ps1 -SkipLive
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 89 Go quality gate" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-go-quality-gates.ps1 -SkipRace -SkipSecurity -OutputRoot "data/local/go-quality/phase89"
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 89 smoke verifier" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase89.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 89 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase89.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 32 activity feed compatibility smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase32.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 41 delivery activity compatibility smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase41.ps1
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
