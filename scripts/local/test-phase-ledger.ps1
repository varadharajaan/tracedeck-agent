param(
    [string]$OutputPath = "data/local/output/phase-ledger.json",
    [string]$TextOutputPath = "data/local/output/phase-ledger.txt",
    [string]$LedgerDocPath = "docs/phase-ledger.md"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-phase-ledger" -LogRoot "logs/local/verify" | Out-Null

try {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-phase-ledger.ps1 -OutputPath $OutputPath -TextOutputPath $TextOutputPath
    if ($LASTEXITCODE -ne 0) {
        throw "get-phase-ledger.ps1 failed with exit code $LASTEXITCODE"
    }

    $jsonPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    $docPath = Join-Path $script:TraceDeckRepoRoot $LedgerDocPath

    if (!(Test-Path $jsonPath)) {
        throw "Missing phase ledger JSON output: $OutputPath"
    }
    if (!(Test-Path $textPath)) {
        throw "Missing phase ledger text output: $TextOutputPath"
    }
    if (!(Test-Path $docPath)) {
        throw "Missing phase ledger doc: $LedgerDocPath"
    }

    $ledger = Get-Content -Path $jsonPath -Raw | ConvertFrom-Json
    $doc = Get-Content -Path $docPath -Raw
    $text = Get-Content -Path $textPath -Raw

    if ($ledger.evidence_scope -ne "metadata_only") {
        throw "Ledger evidence_scope must be metadata_only"
    }
    if ([int]$ledger.counts.highest_tracked_phase -lt 106) {
        throw "Ledger highest tracked phase should include Phase 106"
    }
    if ([int]$ledger.counts.tracked_phase_verify_scripts -lt 100) {
        throw "Ledger tracked phase verifier count is unexpectedly low"
    }
    if ([int]$ledger.counts.remaining_planned_numbered_phases -ne 0) {
        throw "Ledger remaining planned numbered phases must be 0 until a future phase is explicitly planned"
    }
    if ($ledger.answer.statement -ne "0 currently defined numbered phases remain.") {
        throw "Ledger answer statement drifted"
    }
    if ($doc -notmatch "Remaining planned numbered phases: 0") {
        throw "Ledger doc must state the current remaining planned phase count"
    }
    if ($doc -notmatch "python ./devctl.py ledger") {
        throw "Ledger doc must expose the devctl ledger command"
    }
    if ($text -notmatch "Answer: 0 currently defined numbered phases remain\.") {
        throw "Ledger text output must contain the direct remaining-phase answer"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase ledger contract passed. highest_tracked_phase=$($ledger.counts.highest_tracked_phase) latest_merged_phase=$($ledger.counts.latest_merged_phase) remaining=$($ledger.counts.remaining_planned_numbered_phases)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
