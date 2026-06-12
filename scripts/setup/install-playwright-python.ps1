param(
    [string]$Version = "1.52.0"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "install-playwright-python" -LogRoot "logs/local/setup" | Out-Null

try {
    $python = Get-Command python -ErrorAction SilentlyContinue
    if (-not $python) {
        throw "python is not installed or not on PATH"
    }

    Invoke-TraceDeckLoggedCommand -Label "Install Python Playwright package" -Command {
        python -m pip install --user "playwright==$Version"
    }

    Invoke-TraceDeckLoggedCommand -Label "Install Playwright Chromium browser" -Command {
        python -m playwright install chromium
    }

    Invoke-TraceDeckLoggedCommand -Label "Verify Python Playwright import" -Command {
        python -c "from playwright.sync_api import sync_playwright"
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
