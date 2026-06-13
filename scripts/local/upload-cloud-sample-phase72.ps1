param(
    [string]$Bucket = "",
    [string]$Region = "ap-south-1",
    [string]$Prefix = "",
    [string]$TenantID = "family-varadha",
    [string]$DeviceID = "demo-study-laptop",
    [string]$HostName = "demo-study-laptop",
    [string]$ManifestPath = "",
    [switch]$SkipUpload
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "upload-cloud-sample-phase72" -LogRoot "logs/local/cloud" | Out-Null

function Get-TraceDeckStackOutput {
    param([string]$OutputKey)

    $outputPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/stack-outputs.txt"
    if (!(Test-Path $outputPath)) {
        return ""
    }

    $lines = Get-Content -Path $outputPath
    for ($index = 0; $index -lt $lines.Count; $index++) {
        if ($lines[$index].Trim() -eq $OutputKey) {
            for ($candidate = $index + 1; $candidate -lt $lines.Count; $candidate++) {
                $value = $lines[$candidate].Trim()
                if ($value -and $value -ne $OutputKey -and $value -notmatch "^S3 data bucket" -and $value -notmatch "^Public TraceDeck") {
                    return $value
                }
            }
        }
    }
    return ""
}

function New-BrowserArchiveRecord {
    param(
        [string]$ID,
        [string]$Browser,
        [string]$Domain,
        [string]$Category,
        [int]$VisitCount,
        [bool]$YouTubeStudyMatch,
        [datetime]$ObservedAt
    )

    return [ordered]@{
        ID = $ID
        Type = "browser.domain.observed"
        Source = "collector.browser.history"
        Timestamp = $ObservedAt.ToUniversalTime().ToString("o")
        TenantID = $TenantID
        DeviceID = $DeviceID
        HostName = $HostName
        AppName = $Browser
        Metadata = [ordered]@{
            profile = "ai-btech-student"
            operating_system = "windows"
            browser_name = $Browser
            domain = $Domain
            category = $Category
            source_kind = "s3_sample"
            evidence_scope = "metadata_only"
            evidence_detail = "Sampled from S3 archive metadata for cloud admin rendering."
            url_mode = "domain_only"
            stored_url_mode = "domain_only"
            visit_count = [string]$VisitCount
            youtube_study_match = $YouTubeStudyMatch.ToString().ToLowerInvariant()
        }
    }
}

function Resolve-TraceDeckLocalPath {
    param([string]$Path)

    if ([System.IO.Path]::IsPathRooted($Path)) {
        return $Path
    }
    return Join-Path $script:TraceDeckRepoRoot $Path
}

try {
    if ([string]::IsNullOrWhiteSpace($Bucket)) {
        $Bucket = Get-TraceDeckStackOutput -OutputKey "DataBucket"
    }
    if ([string]::IsNullOrWhiteSpace($Bucket)) {
        throw "S3 bucket is not configured. Run 'python ./devctl.py sam outputs' or pass -Bucket."
    }

    if (-not $SkipUpload) {
        $aws = Get-Command aws -ErrorAction SilentlyContinue
        if (-not $aws) {
            throw "AWS CLI is not installed or not on PATH."
        }
    }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $observedAt = (Get-Date).ToUniversalTime()
    $runRoot = Join-Path $script:TraceDeckRepoRoot "data/local/cloud-seed/phase72/$timestamp"
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

    $records = @(
        (New-BrowserArchiveRecord -ID "phase72-cloud-chrome-study" -Browser "chrome" -Domain "docs.python.org" -Category "study" -VisitCount 7 -YouTubeStudyMatch $false -ObservedAt $observedAt.AddMinutes(-5)),
        (New-BrowserArchiveRecord -ID "phase72-cloud-edge-youtube-review" -Browser "edge" -Domain "youtube.com" -Category "video-streaming" -VisitCount 3 -YouTubeStudyMatch $false -ObservedAt $observedAt.AddMinutes(-11)),
        (New-BrowserArchiveRecord -ID "phase72-cloud-brave-study" -Browser "brave" -Domain "github.com" -Category "study" -VisitCount 5 -YouTubeStudyMatch $false -ObservedAt $observedAt.AddMinutes(-18)),
        (New-BrowserArchiveRecord -ID "phase72-cloud-chrome-social-review" -Browser "chrome" -Domain "instagram.com" -Category "social-media" -VisitCount 2 -YouTubeStudyMatch $false -ObservedAt $observedAt.AddMinutes(-27)),
        (New-BrowserArchiveRecord -ID "phase72-cloud-edge-study-youtube" -Browser "edge" -Domain "youtube.com" -Category "study" -VisitCount 4 -YouTubeStudyMatch $true -ObservedAt $observedAt.AddMinutes(-33))
    )

    $jsonLines = $records | ForEach-Object { $_ | ConvertTo-Json -Compress -Depth 12 }
    $plainText = $jsonLines -join "`n"
    foreach ($forbidden in @("https://", "http://", "raw_url", "page_title", "password", "cookie_value", "token_value", "screenshot", "push_endpoint", "provider_secret", "keylogger")) {
        if ($plainText.ToLowerInvariant().Contains($forbidden)) {
            throw "Cloud sample contains forbidden marker '$forbidden'."
        }
    }

    $archivePath = Join-Path $runRoot "tracedeck-cloud-sample-$timestamp.jsonl.gz"
    $file = [System.IO.File]::Create($archivePath)
    try {
        $gzip = [System.IO.Compression.GzipStream]::new($file, [System.IO.Compression.CompressionMode]::Compress)
        try {
            $writer = [System.IO.StreamWriter]::new($gzip, [System.Text.UTF8Encoding]::new($false))
            try {
                foreach ($line in $jsonLines) {
                    $writer.WriteLine($line)
                }
            }
            finally {
                $writer.Dispose()
            }
        }
        finally {
            $gzip.Dispose()
        }
    }
    finally {
        $file.Dispose()
    }

    $date = $observedAt.ToString("yyyy-MM-dd")
    $hour = $observedAt.ToString("HH")
    $fileName = Split-Path -Leaf $archivePath
    $keyPrefix = "tenant=$TenantID/device=$DeviceID/date=$date/hour=$hour"
    if (-not [string]::IsNullOrWhiteSpace($Prefix)) {
        $keyPrefix = $Prefix.Trim("/")
    }
    $s3Key = "$keyPrefix/$fileName"
    $destination = "s3://$Bucket/$s3Key"

    if (-not $SkipUpload) {
        Invoke-TraceDeckLoggedCommand -Label "Upload Phase 72 S3 cloud sample" -Command {
            aws s3 cp $archivePath $destination --region $Region --content-type "application/jsonl"
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "SkipUpload enabled; not uploading $archivePath."
    }

    $manifest = [ordered]@{
        bucket = $Bucket
        region = $Region
        key = $s3Key
        destination = $destination
        local_path = $archivePath
        record_count = $records.Count
        tenant_id = $TenantID
        device_id = $DeviceID
        host_name = $HostName
        uploaded = -not $SkipUpload
        generated_at = (Get-Date).ToUniversalTime().ToString("o")
        privacy_boundary = "metadata-only S3 sample: domains, browser names, category, study-safe inference metadata, counts, host labels, and timestamps only"
    }

    if ([string]::IsNullOrWhiteSpace($ManifestPath)) {
        $ManifestPath = Join-Path $runRoot "manifest.json"
    }
    $resolvedManifestPath = Resolve-TraceDeckLocalPath -Path $ManifestPath
    $manifestDir = Split-Path -Parent $resolvedManifestPath
    if (!(Test-Path $manifestDir)) {
        New-Item -ItemType Directory -Path $manifestDir -Force | Out-Null
    }
    $manifest | ConvertTo-Json -Depth 8 | Set-Content -Path $resolvedManifestPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Phase 72 cloud sample manifest: $resolvedManifestPath"
    Write-TraceDeckLog -Level "INFO" -Message "Phase 72 cloud sample object: $destination"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
