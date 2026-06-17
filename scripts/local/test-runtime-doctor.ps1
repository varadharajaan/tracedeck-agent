param(
    [string]$Addr = "127.0.0.1:18080",
    [switch]$IncludeCloud
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-runtime-doctor" -LogRoot "logs/local/test" | Out-Null

try {
    $doctorArgs = @("--addr", $Addr, "doctor")
    if (-not $IncludeCloud) {
        $doctorArgs += "--skip-cloud"
    }

    Invoke-TraceDeckLoggedCommand -Label "Run TraceDeck runtime doctor" -Command {
        python ./devctl.py @doctorArgs
    }

    $jsonPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-doctor.json"
    $textPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-doctor.txt"
    if (!(Test-Path $jsonPath)) {
        throw "Expected runtime doctor JSON at $jsonPath"
    }
    if (!(Test-Path $textPath)) {
        throw "Expected runtime doctor text at $textPath"
    }

    $report = Get-Content -Path $jsonPath -Raw | ConvertFrom-Json
    if ($report.overall -ne "ok") {
        throw "Expected runtime doctor overall ok."
    }
    if ($report.local.overall -ne "ok") {
        throw "Expected runtime doctor local overall ok."
    }
    if ($report.local.browser_api.rows -lt 1) {
        throw "Expected runtime doctor browser rows."
    }
    if ($report.local.dashboard.ok -ne $true -or $report.local.browser_page.ok -ne $true) {
        throw "Expected dashboard and browser page runtime checks to pass."
    }
    if ($report.local.deliveries.ok -ne $true) {
        throw "Expected runtime doctor delivery provenance."
    }
    if ($report.local.deliveries.default_demo_hidden -ne $true -or $report.local.deliveries.opt_in_demo_available -ne $true) {
        throw "Expected runtime doctor to hide demo delivery rows by default and expose opt-in demo proof."
    }
    if ($report.local.delivery_assurance.ok -ne $true) {
        throw "Expected runtime doctor delivery assurance."
    }
    if ($IncludeCloud -and $report.cloud.overall -ne "ok") {
        throw "Expected runtime doctor cloud overall ok."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Runtime doctor test passed addr=$Addr cloud=$($IncludeCloud.IsPresent) report=$jsonPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
