param(
    [string]$Addr = "127.0.0.1:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-runtime-summary" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Generate runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-runtime-summary.ps1 -Addr $Addr
    }

    $jsonPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-summary.json"
    $textPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-summary.txt"
    if (!(Test-Path $jsonPath)) {
        throw "Expected runtime summary JSON at $jsonPath"
    }
    if (!(Test-Path $textPath)) {
        throw "Expected runtime summary text at $textPath"
    }

    $summary = Get-Content -Path $jsonPath -Raw | ConvertFrom-Json
    if ($summary.backend.runtime_ok -ne $true) {
        throw "Expected backend runtime_ok=true in runtime summary."
    }
    if ($summary.backend.health_ok -ne $true) {
        throw "Expected backend health_ok=true in runtime summary."
    }
    if ([string]::IsNullOrWhiteSpace($summary.backend.advisory.code)) {
        throw "Expected backend advisory code in runtime summary."
    }
    if ([string]::IsNullOrWhiteSpace($summary.backend.ready_pid_status)) {
        throw "Expected ready_pid_status in runtime summary."
    }
    if ($summary.doctor.overall -ne "ok") {
        throw "Expected runtime summary doctor overall ok."
    }
    if ($summary.doctor.local -ne "ok") {
        throw "Expected runtime summary doctor local ok."
    }
    if ($summary.verdict.can_continue -ne $true) {
        throw "Expected runtime summary verdict can_continue=true."
    }
    if ($summary.privacy.metadata_only -ne $true -or $summary.privacy.sensitive_collection -ne "denied") {
        throw "Expected metadata-only privacy boundary in runtime summary."
    }

    $text = Get-Content -Path $textPath -Raw
    foreach ($marker in @("TraceDeck runtime summary", "Backend:", "PID:", "Doctor:", "Verdict:")) {
        if ($text -notlike "*$marker*") {
            throw "Expected runtime summary text marker '$marker'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Runtime summary test passed json=$jsonPath text=$textPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
