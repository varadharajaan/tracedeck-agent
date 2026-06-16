param(
    [string]$OutputPath = "data/local/output/contract-completion-audit.json",
    [string]$TextOutputPath = "data/local/output/contract-completion-audit.txt",
    [string]$AuditDocPath = "docs/contract-completion-audit.md"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-contract-completion-audit" -LogRoot "logs/local/verify" | Out-Null

try {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-contract-completion-audit.ps1 -OutputPath $OutputPath -TextOutputPath $TextOutputPath
    if ($LASTEXITCODE -ne 0) {
        throw "get-contract-completion-audit.ps1 failed with exit code $LASTEXITCODE"
    }

    $jsonPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    $docPath = Join-Path $script:TraceDeckRepoRoot $AuditDocPath

    if (!(Test-Path $jsonPath)) {
        throw "Missing contract audit JSON output: $OutputPath"
    }
    if (!(Test-Path $textPath)) {
        throw "Missing contract audit text output: $TextOutputPath"
    }
    if (!(Test-Path $docPath)) {
        throw "Missing contract audit doc: $AuditDocPath"
    }

    $audit = Get-Content -Path $jsonPath -Raw | ConvertFrom-Json
    $text = Get-Content -Path $textPath -Raw
    $doc = Get-Content -Path $docPath -Raw

    if ($audit.evidence_scope -ne "metadata_only") {
        throw "Contract audit evidence_scope must be metadata_only"
    }
    if ($audit.overall_status -ne "attention") {
        throw "Contract audit should currently report attention until remaining gaps are implemented"
    }
    if ([int]$audit.summary.missing -lt 1) {
        throw "Contract audit should expose at least one missing end-to-end deliverable"
    }
    $ids = @($audit.requirements | ForEach-Object { $_.id })
    foreach ($requiredID in @("browser-extension-skeleton", "opentelemetry-exporter", "docker-compose-otel-stack", "release-sbom-packaging")) {
        if ($ids -notcontains $requiredID) {
            throw "Contract audit missing required finding id: $requiredID"
        }
    }
    if ($text -notmatch "TraceDeck is not end-to-end complete yet") {
        throw "Contract audit text must avoid claiming full completion"
    }
    if ($doc -notmatch "Remaining or partial") {
        throw "Contract audit doc must describe remaining or partial work"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Contract completion audit passed. ok=$($audit.summary.ok) attention=$($audit.summary.attention) missing=$($audit.summary.missing)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
