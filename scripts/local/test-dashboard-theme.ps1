param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputRoot = "data/local/dashboard-theme"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-dashboard-theme" -LogRoot "logs/local/test" | Out-Null

try {
    $python = Get-Command python -ErrorAction SilentlyContinue
    if (-not $python) {
        throw "python is not installed or not on PATH"
    }

    Invoke-TraceDeckLoggedCommand -Label "Check Playwright Python package" -Command {
        python -c "from playwright.sync_api import sync_playwright"
    }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $outputDir = Join-Path $script:TraceDeckRepoRoot (Join-Path $OutputRoot $timestamp)
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
    $reportPath = Join-Path $outputDir "dashboard-theme-report.json"
    $checkerPath = Join-Path $script:TraceDeckRepoRoot "scripts/tools/dashboard_theme_check.py"

    Invoke-TraceDeckLoggedCommand -Label "Dashboard light and dark theme contract" -Command {
        python $checkerPath --base-url $BaseUrl --output $reportPath
    }

    if (-not (Test-Path $reportPath)) {
        throw "Expected dashboard theme report was not created: $reportPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Dashboard theme report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
