param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputPath = "data/local/output/local-monitoring-indicator.json",
    [string]$TextOutputPath = "data/local/output/local-monitoring-indicator.txt",
    [string]$HtmlOutputPath = "data/local/output/local-monitoring-indicator.html",
    [switch]$Open
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-local-monitoring-indicator" -LogRoot "logs/local/ops" | Out-Null

function ConvertTo-TraceDeckHtml {
    param([AllowNull()][object]$Value)
    return [System.Net.WebUtility]::HtmlEncode([string]$Value)
}

try {
    $base = $BaseUrl.TrimEnd("/")
    if ([string]::IsNullOrWhiteSpace($base)) {
        $base = "http://127.0.0.1:18080"
    }
    $center = Invoke-RestMethod -Method "GET" -Uri "$base/api/v1/local-monitoring-indicator"

    $outputFullPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textFullPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    $htmlFullPath = Join-Path $script:TraceDeckRepoRoot $HtmlOutputPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $textFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $htmlFullPath) | Out-Null

    $center | ConvertTo-Json -Depth 24 | Set-Content -Path $outputFullPath -Encoding UTF8

    $summary = $center.summary
    $proof = @($center.proof)
    $actions = @($center.actions)
    $lines = @(
        "TraceDeck Local Monitoring Indicator",
        "Generated: $($center.generated_at)",
        "Status: $($summary.status)",
        "Headline: $($summary.headline)",
        "Visible indicator ready: $($summary.visible_indicator_ready)",
        "Local status page ready: $($summary.local_status_page_ready)",
        "Runtime ready: $($summary.runtime_ready)",
        "Consent visible: $($summary.consent_visible)",
        "Sensitive collection denied: $($summary.sensitive_collection_denied)",
        "Transparency mode: $($summary.transparency_mode)",
        "Indicator surface: $($summary.indicator_surface)",
        "Dashboard route: $($summary.dashboard_route)",
        "JSON: $OutputPath",
        "Text: $TextOutputPath",
        "HTML: $HtmlOutputPath",
        "Proof rows: $($proof.Count)",
        "Actions: $($actions.Count)",
        "Privacy: metadata-only local indicator; passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, and raw provider payloads are denied"
    )
    $lines | Set-Content -Path $textFullPath -Encoding UTF8

    $proofRows = ($proof | ForEach-Object {
        "<tr><td>$((ConvertTo-TraceDeckHtml $_.label))</td><td>$((ConvertTo-TraceDeckHtml $_.value))</td><td>$((ConvertTo-TraceDeckHtml $_.status))</td><td>$((ConvertTo-TraceDeckHtml $_.detail))</td></tr>"
    }) -join [Environment]::NewLine
    $actionRows = ($actions | ForEach-Object {
        "<tr><td>$((ConvertTo-TraceDeckHtml $_.title))</td><td>$((ConvertTo-TraceDeckHtml $_.status))</td><td>$((ConvertTo-TraceDeckHtml $_.command))</td><td>$((ConvertTo-TraceDeckHtml $_.detail))</td></tr>"
    }) -join [Environment]::NewLine
    $statusClass = if ($summary.status -eq "ok") { "ok" } elseif ($summary.status -eq "watch") { "watch" } else { "attention" }

    $html = @"
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TraceDeck Local Monitoring Indicator</title>
  <style>
    :root { color-scheme: light dark; font-family: Segoe UI, Arial, sans-serif; background: #f7f8fb; color: #172033; }
    body { margin: 0; padding: 32px; }
    main { max-width: 1040px; margin: 0 auto; }
    header { display: flex; justify-content: space-between; gap: 20px; align-items: flex-start; border-bottom: 1px solid #d6dbe7; padding-bottom: 20px; }
    h1 { margin: 0 0 8px; font-size: 28px; }
    p { line-height: 1.55; }
    .badge { display: inline-flex; align-items: center; padding: 7px 11px; border-radius: 999px; font-weight: 700; text-transform: uppercase; letter-spacing: 0; font-size: 12px; }
    .ok { background: #d8f3dc; color: #14532d; }
    .watch { background: #fff3c4; color: #713f12; }
    .attention { background: #fee2e2; color: #7f1d1d; }
    .grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; margin: 22px 0; }
    .tile { border: 1px solid #d6dbe7; border-radius: 8px; padding: 14px; background: #fff; }
    .tile small { display: block; color: #64748b; margin-bottom: 8px; }
    .tile strong { font-size: 18px; }
    table { width: 100%; border-collapse: collapse; margin: 16px 0 24px; background: #fff; border: 1px solid #d6dbe7; }
    th, td { text-align: left; padding: 10px 12px; border-bottom: 1px solid #e7eaf1; vertical-align: top; }
    th { color: #475569; font-size: 12px; text-transform: uppercase; letter-spacing: 0; }
    .privacy { border-left: 4px solid #2563eb; padding: 12px 14px; background: #eff6ff; }
    @media (prefers-color-scheme: dark) {
      :root { background: #0f172a; color: #e5e7eb; }
      header { border-bottom-color: #334155; }
      .tile, table { background: #111827; border-color: #334155; }
      th, td { border-bottom-color: #263244; }
      .privacy { background: #172554; border-left-color: #60a5fa; }
    }
    @media (max-width: 760px) { body { padding: 18px; } header { display: block; } .grid { grid-template-columns: 1fr; } }
  </style>
</head>
<body>
  <main>
    <header>
      <div>
        <h1>TraceDeck Local Monitoring Indicator</h1>
        <p>$((ConvertTo-TraceDeckHtml $summary.headline))</p>
      </div>
      <span class="badge $statusClass">$((ConvertTo-TraceDeckHtml $summary.status))</span>
    </header>
    <section class="grid" aria-label="indicator summary">
      <div class="tile"><small>Visible</small><strong>$((ConvertTo-TraceDeckHtml $summary.visible_indicator_ready))</strong></div>
      <div class="tile"><small>Runtime</small><strong>$((ConvertTo-TraceDeckHtml $summary.runtime_ready))</strong></div>
      <div class="tile"><small>Consent</small><strong>$((ConvertTo-TraceDeckHtml $summary.consent_visible))</strong></div>
      <div class="tile"><small>Sensitive Data</small><strong>$((ConvertTo-TraceDeckHtml $summary.sensitive_collection_denied))</strong></div>
    </section>
    <section>
      <h2>Indicator Proof</h2>
      <table>
        <thead><tr><th>Label</th><th>Value</th><th>Status</th><th>Detail</th></tr></thead>
        <tbody>
$proofRows
        </tbody>
      </table>
    </section>
    <section>
      <h2>Actions</h2>
      <table>
        <thead><tr><th>Action</th><th>Status</th><th>Command</th><th>Detail</th></tr></thead>
        <tbody>
$actionRows
        </tbody>
      </table>
    </section>
    <section class="privacy">
      <strong>Privacy Boundary</strong>
      <p>$((ConvertTo-TraceDeckHtml $center.privacy_boundary))</p>
    </section>
  </main>
</body>
</html>
"@
    $html | Set-Content -Path $htmlFullPath -Encoding UTF8

    if ($Open) {
        Start-Process -FilePath $htmlFullPath | Out-Null
    }

    Write-TraceDeckLog -Level "INFO" -Message "Local monitoring indicator saved json=$OutputPath text=$TextOutputPath html=$HtmlOutputPath status=$($summary.status) proof=$($proof.Count) actions=$($actions.Count)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
