param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputRoot = "data/local/dashboard-visual-quality"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-dashboard-visual-quality" -LogRoot "logs/local/test" | Out-Null

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
    $reportPath = Join-Path $outputDir "dashboard-visual-quality-report.json"
    $checkerPath = Join-Path $script:TraceDeckRepoRoot "scripts/tools/dashboard_visual_quality_check.py"

    Invoke-TraceDeckLoggedCommand -Label "Dashboard product visual quality contract" -Command {
        python $checkerPath --base-url $BaseUrl --output $reportPath
    }

    if (-not (Test-Path $reportPath)) {
        throw "Expected dashboard visual quality report was not created: $reportPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Dashboard visual quality report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
