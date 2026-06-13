param(
    [string]$Bucket = "",
    [string]$Region = "ap-south-1",
    [string]$FrontendUrl = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase72" -LogRoot "logs/local/smoke" | Out-Null

function Get-TraceDeckFrontendUrl {
    $path = Join-Path $script:TraceDeckRepoRoot "data/local/output/frontend-url.txt"
    if (!(Test-Path $path)) {
        throw "Frontend URL output is missing. Run 'python ./devctl.py sam outputs'."
    }
    return (Get-Content -Raw -Path $path).Trim().TrimEnd("/")
}

function Assert-Condition {
    param(
        [bool]$Condition,
        [string]$Message
    )
    if (-not $Condition) {
        throw $Message
    }
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $runRoot = "data/local/smoke-phase72/$timestamp"
    $manifestPath = "$runRoot/cloud-sample-manifest.json"
    New-Item -ItemType Directory -Force -Path (Join-Path $script:TraceDeckRepoRoot $runRoot) | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Upload Phase 72 cloud sample archive" -Command {
        $uploadArgs = @(
            "-NoProfile",
            "-ExecutionPolicy", "Bypass",
            "-File", "./scripts/local/upload-cloud-sample-phase72.ps1",
            "-Region", $Region,
            "-ManifestPath", $manifestPath
        )
        if (-not [string]::IsNullOrWhiteSpace($Bucket)) {
            $uploadArgs += @("-Bucket", $Bucket)
        }
        powershell @uploadArgs
    }

    $manifestFullPath = Join-Path $script:TraceDeckRepoRoot $manifestPath
    $manifest = Get-Content -Raw -Path $manifestFullPath | ConvertFrom-Json
    Assert-Condition -Condition ($manifest.uploaded -eq $true) -Message "Expected cloud sample manifest to mark upload as true."
    Assert-Condition -Condition ($manifest.record_count -ge 5) -Message "Expected at least five cloud sample rows."

    if ([string]::IsNullOrWhiteSpace($FrontendUrl)) {
        $FrontendUrl = Get-TraceDeckFrontendUrl
    }
    $FrontendUrl = $FrontendUrl.TrimEnd("/")

    Invoke-TraceDeckLoggedCommand -Label "Read Lambda S3 summary refresh" -Command {
        $script:phase72Summary = Invoke-RestMethod -Method "GET" -Uri "$FrontendUrl/api/s3-summary?refresh=true"
    }

    Assert-Condition -Condition ($script:phase72Summary.status -eq "ok") -Message "Expected Lambda S3 summary status ok."
    Assert-Condition -Condition ($script:phase72Summary.bucket -eq $manifest.bucket) -Message "Expected Lambda summary bucket to match uploaded bucket."
    Assert-Condition -Condition ($script:phase72Summary.summary.objects -ge 1) -Message "Expected at least one S3 object in Lambda summary."
    Assert-Condition -Condition ($script:phase72Summary.summary.sampled_rows -ge 5) -Message "Expected Lambda to sample uploaded browser rows."
    Assert-Condition -Condition ($script:phase72Summary.summary.study_safe -ge 2) -Message "Expected study-safe rows from uploaded archive."
    Assert-Condition -Condition ($script:phase72Summary.summary.non_study_youtube -ge 1) -Message "Expected non-study YouTube row from uploaded archive."

    $browserLabels = @($script:phase72Summary.browsers | ForEach-Object { $_.label })
    foreach ($browser in @("chrome", "edge", "brave")) {
        Assert-Condition -Condition ($browserLabels -contains $browser) -Message "Expected Lambda summary to include browser '$browser'."
    }

    $serialized = ($script:phase72Summary | ConvertTo-Json -Depth 20).ToLowerInvariant()
    foreach ($forbidden in @("https://", "http://", "raw_url", "page_title", "password", "cookie_value", "token_value", "screenshot", "push_endpoint", "provider_secret", "keylogger")) {
        Assert-Condition -Condition (-not $serialized.Contains($forbidden)) -Message "Lambda S3 summary exposed forbidden marker '$forbidden'."
    }
    Assert-Condition -Condition ($script:phase72Summary.privacy_boundary -match "metadata-only") -Message "Expected metadata-only privacy boundary."

    Invoke-TraceDeckLoggedCommand -Label "Read Lambda S3 summary cache hit" -Command {
        $script:phase72Cached = Invoke-RestMethod -Method "GET" -Uri "$FrontendUrl/api/s3-summary"
    }
    Assert-Condition -Condition ($script:phase72Cached.cache.hit -eq $true) -Message "Expected second Lambda summary read to be a cache hit."
    Assert-Condition -Condition ($script:phase72Cached.cache.hit_percent -gt 0) -Message "Expected cache hit percentage to be greater than zero."

    Write-TraceDeckLog -Level "INFO" -Message "Phase 72 cloud S3 smoke passed url=$FrontendUrl key=$($manifest.key) rows=$($script:phase72Summary.summary.sampled_rows) cache_hit=$($script:phase72Cached.cache.hit_percent)%"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
