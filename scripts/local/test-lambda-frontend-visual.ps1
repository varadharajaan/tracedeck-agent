param(
    [string]$OutputRoot = "data/local/lambda-frontend-visual"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-lambda-frontend-visual" -LogRoot "logs/local/test" | Out-Null

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
    $reportPath = Join-Path $outputDir "lambda-frontend-visual-report.json"
    $checkerPath = Join-Path $script:TraceDeckRepoRoot "scripts/tools/lambda_frontend_visual_check.py"

    Invoke-TraceDeckLoggedCommand -Label "Lambda frontend product visual quality contract" -Command {
        python $checkerPath --output $reportPath
    }

    if (-not (Test-Path $reportPath)) {
        throw "Expected Lambda frontend visual report was not created: $reportPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Lambda frontend visual report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
