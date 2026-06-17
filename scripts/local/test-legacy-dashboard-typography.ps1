param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputRoot = "data/local/legacy-dashboard-typography"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-legacy-dashboard-typography" -LogRoot "logs/local/test" | Out-Null

try {
    $env:PYTHONPYCACHEPREFIX = Join-Path $script:TraceDeckRepoRoot "data/local/pycache"
    $runRoot = Join-Path $script:TraceDeckRepoRoot (Join-Path $OutputRoot (Get-Date -Format "yyyyMMdd-HHmmss"))
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null
    $reportPath = Join-Path $runRoot "legacy-dashboard-typography-report.json"

    Invoke-TraceDeckLoggedCommand -Label "Check Playwright Python package" -Command {
        python -c "import playwright"
    }

    Invoke-TraceDeckLoggedCommand -Label "Legacy dashboard typography contract" -Command {
        python ./scripts/tools/legacy_dashboard_typography_check.py --base-url $BaseUrl --output $reportPath
    }

    Write-TraceDeckLog -Level "INFO" -Message "Legacy dashboard typography report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
