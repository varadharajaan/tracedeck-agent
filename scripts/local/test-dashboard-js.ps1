param(
    [string]$DashboardPath = "backend/internal/api/web/dashboard.html",
    [string]$OutputRoot = "data/local/dashboard-js-check"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-dashboard-js" -LogRoot "logs/local/test" | Out-Null

try {
    $node = Get-Command node -ErrorAction SilentlyContinue
    if (-not $node) {
        throw "node is not installed or not on PATH"
    }

    $dashboardFullPath = Join-Path $script:TraceDeckRepoRoot $DashboardPath
    $html = Get-Content -Raw -Path $dashboardFullPath
    $matches = [regex]::Matches($html, "<script>(?<script>[\s\S]*?)</script>")
    if ($matches.Count -lt 1) {
        throw "No inline dashboard script block found in $DashboardPath"
    }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $outputDir = Join-Path $script:TraceDeckRepoRoot (Join-Path $OutputRoot $timestamp)
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
    $scriptPath = Join-Path $outputDir "dashboard.js"
    Set-Content -Path $scriptPath -Value $matches[$matches.Count - 1].Groups["script"].Value -Encoding UTF8

    $nodeParser = "const fs=require('fs'); const vm=require('vm'); const code=fs.readFileSync(process.argv[1], 'utf8'); new vm.Script(code, { filename: 'dashboard-inline.js' });"
    Invoke-TraceDeckLoggedCommand -Label "Dashboard JavaScript syntax check" -Command {
        & $node.Source -e $nodeParser $scriptPath
    }

    Write-TraceDeckLog -Level "INFO" -Message "Dashboard JavaScript syntax check passed: $scriptPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
