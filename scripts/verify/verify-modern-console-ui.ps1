param(
    [string]$BaseUrl = "http://127.0.0.1:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-modern-console-ui" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"

    Invoke-TraceDeckLoggedCommand -Label "Python UI checker syntax" -Command {
        python -m py_compile ./devctl.py ./scripts/tools/dashboard_layout_check.py ./scripts/tools/dashboard_visual_quality_check.py ./scripts/tools/dashboard_theme_check.py ./scripts/tools/browser_activity_badge_check.py ./scripts/tools/legacy_dashboard_typography_check.py
    }

    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Dashboard DOM contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -OutputRoot "data/local/dashboard-js-check/modern-console-verify"
    }

    Invoke-TraceDeckLoggedCommand -Label "Browser Intelligence JavaScript syntax" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 -DashboardPath "backend/internal/api/web/browser_activity.html" -OutputRoot "data/local/dashboard-js-check/modern-console-browser-verify"
    }

    Invoke-TraceDeckLoggedCommand -Label "Dashboard layout contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $BaseUrl -OutputRoot "data/local/dashboard-layout/modern-console-verify"
    }

    Invoke-TraceDeckLoggedCommand -Label "Dashboard theme contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-theme.ps1 -BaseUrl $BaseUrl -OutputRoot "data/local/dashboard-theme/modern-console-verify"
    }

    Invoke-TraceDeckLoggedCommand -Label "Dashboard visual quality contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-visual-quality.ps1 -BaseUrl $BaseUrl -OutputRoot "data/local/dashboard-visual-quality/modern-console-verify"
    }

    Invoke-TraceDeckLoggedCommand -Label "Legacy dashboard typography contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-legacy-dashboard-typography.ps1 -BaseUrl $BaseUrl -OutputRoot "data/local/legacy-dashboard-typography/modern-console-verify"
    }

    Invoke-TraceDeckLoggedCommand -Label "Browser Activity badge integrity" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-browser-activity-badges.ps1 -BaseUrl $BaseUrl
    }

    Invoke-TraceDeckLoggedCommand -Label "Live server provenance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl $BaseUrl
    }

    Invoke-TraceDeckLoggedCommand -Label "Cleanup generated root artifacts" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/clean-root-generated.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Root artifact check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "git diff --check" -Command {
        git diff --check
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
