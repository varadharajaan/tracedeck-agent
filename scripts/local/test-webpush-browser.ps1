param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputRoot = "data/local/webpush-browser"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-webpush-browser" -LogRoot "logs/local/test" | Out-Null

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
    $reportPath = Join-Path $outputDir "webpush-browser-report.json"
    $profileDir = Join-Path $outputDir "chromium-profile"
    $checkerPath = Join-Path $script:TraceDeckRepoRoot "scripts/tools/webpush_browser_check.py"

    Invoke-TraceDeckLoggedCommand -Label "Dashboard Web Push browser activation check" -Command {
        python $checkerPath --base-url $BaseUrl --output $reportPath --profile-dir $profileDir
    }

    if (-not (Test-Path $reportPath)) {
        throw "Expected Web Push browser report was not created: $reportPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Web Push browser report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
