param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputPath = "data/local/output/promotion-readiness.json",
    [string]$TextOutputPath = "data/local/output/promotion-readiness.txt"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-promotion-readiness" -LogRoot "logs/local/ops" | Out-Null

try {
    $base = $BaseUrl.TrimEnd("/")
    if ([string]::IsNullOrWhiteSpace($base)) {
        $base = "http://127.0.0.1:18080"
    }
    $center = Invoke-RestMethod -Method "GET" -Uri "$base/api/v1/promotion-readiness-center"

    $outputFullPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textFullPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $textFullPath) | Out-Null

    $center | ConvertTo-Json -Depth 24 | Set-Content -Path $outputFullPath -Encoding UTF8

    $summary = $center.summary
    $proof = @($center.proof)
    $actions = @($center.actions)
    $lines = @(
        "TraceDeck Promotion Readiness",
        "Generated: $($center.generated_at)",
        "Status: $($summary.status)",
        "Headline: $($summary.headline)",
        "Can promote: $($summary.can_promote)",
        "Runtime ready: $($summary.runtime_ready)",
        "Verification ready: $($summary.verification_ready)",
        "Assurance ready: $($summary.assurance_ready)",
        "Git clean: $($summary.git_clean)",
        "Ready PID: $($summary.ready_pid_status)",
        "Scheduler: $($summary.scheduler_readback)",
        "Gates: $($summary.gates_ok)/$($summary.gates_total)",
        "Watch count: $($summary.watch_count)",
        "Attention count: $($summary.attention_count)",
        "Next step: $($summary.operator_next_step)",
        "Promotion export: $OutputPath",
        "Operator assurance export: $($center.operator_assurance_path)",
        "Proof rows: $($proof.Count)",
        "Actions: $($actions.Count)",
        "Privacy: metadata-only promotion proof; sensitive collection denied"
    )
    $lines | Set-Content -Path $textFullPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Promotion readiness saved json=$OutputPath text=$TextOutputPath status=$($summary.status) can_promote=$($summary.can_promote) proof=$($proof.Count) actions=$($actions.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
