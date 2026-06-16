param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputPath = "data/local/output/operator-assurance.json",
    [string]$TextOutputPath = "data/local/output/operator-assurance.txt"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-operator-assurance" -LogRoot "logs/local/ops" | Out-Null

try {
    $base = $BaseUrl.TrimEnd("/")
    if ([string]::IsNullOrWhiteSpace($base)) {
        $base = "http://127.0.0.1:18080"
    }
    $center = Invoke-RestMethod -Method "GET" -Uri "$base/api/v1/operator-assurance-center"

    $outputFullPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textFullPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $textFullPath) | Out-Null

    $center | ConvertTo-Json -Depth 24 | Set-Content -Path $outputFullPath -Encoding UTF8

    $summary = $center.summary
    $cards = @($center.cards)
    $actions = @($center.actions)
    $lines = @(
        "TraceDeck Operator Assurance",
        "Generated: $($center.generated_at)",
        "Status: $($summary.status)",
        "Headline: $($summary.headline)",
        "Can continue: $($summary.can_continue)",
        "Can promote: $($summary.can_promote)",
        "Scheduler: $($summary.scheduler_readback)",
        "Scheduler explanation: $($summary.scheduler_explanation)",
        "Frontend cache: $($summary.frontend_cache_status) $($summary.frontend_cache_hit_pct)%",
        "Git clean: $($summary.git_clean)",
        "Export path: $OutputPath",
        "Cards: $($cards.Count)",
        "Actions: $($actions.Count)",
        "Privacy: metadata-only operator proof; sensitive collection denied"
    )
    $lines | Set-Content -Path $textFullPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Operator assurance saved json=$OutputPath text=$TextOutputPath status=$($summary.status) cards=$($cards.Count) actions=$($actions.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
