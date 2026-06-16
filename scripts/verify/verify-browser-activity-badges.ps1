param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-browser-activity-badges" -LogRoot "logs/local/verify" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    $env:GOCACHE = Join-Path $script:TraceDeckRepoRoot "data/local/go-cache/browser-activity-badges"
    $env:GOTMPDIR = Join-Path $script:TraceDeckRepoRoot "data/local/go-tmp/browser-activity-badges"
    New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
    New-Item -ItemType Directory -Force -Path $env:GOTMPDIR | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Python script syntax check" -Command {
        python -m py_compile ./devctl.py ./scripts/tools/browser_activity_badge_check.py
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser Activity DOM contract guard" -Command {
        go test ./backend/internal/api -run TestBrowserActivityDOMContract -count=1
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser Activity JavaScript syntax check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1 `
            -DashboardPath "backend/internal/api/web/browser_activity.html" `
            -OutputRoot "data/local/dashboard-js-check/browser-activity-badges"
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser Activity badge smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-browser-activity-badges.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Browser Activity Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase68.ps1 -Addr "127.0.0.1:18314"
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
