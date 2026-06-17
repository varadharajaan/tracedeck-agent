param(
    [string]$ConfigPath = "data/local/config/tracedeck-live-this-machine.yaml",
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$DataDir = "data/local/agent-live-s3-test",
    [string]$LogDir = "logs/local/agent-live-s3-test",
    [string]$OutboxDir = "data/local/outbox-live-s3-test",
    [string]$OutputRoot = "data/local/agent-live-s3-test",
    [int]$WaitSeconds = 30
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-live-s3-archive" -LogRoot "logs/local/cloud" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)
    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Read-PolicyValue {
    param(
        [string]$ResolvedConfigPath,
        [string]$Name
    )
    $pattern = "^\s*$([regex]::Escape($Name))\s*:\s*(.+?)\s*$"
    foreach ($line in Get-Content -LiteralPath $ResolvedConfigPath) {
        if ($line -match $pattern) {
            return ($Matches[1] -replace '^"|"$', '').Trim()
        }
    }
    return ""
}

function Get-TraceDeckRelativePath {
    param(
        [string]$RootPath,
        [string]$ChildPath
    )

    $rootFull = [System.IO.Path]::GetFullPath($RootPath).TrimEnd([System.IO.Path]::DirectorySeparatorChar, [System.IO.Path]::AltDirectorySeparatorChar)
    $childFull = [System.IO.Path]::GetFullPath($ChildPath)
    $prefix = $rootFull + [System.IO.Path]::DirectorySeparatorChar
    if ($childFull.StartsWith($prefix, [System.StringComparison]::OrdinalIgnoreCase)) {
        return $childFull.Substring($prefix.Length)
    }
    return Split-Path -Leaf $childFull
}

try {
    $resolvedConfigPath = Resolve-TraceDeckPath -PathValue $ConfigPath
    $resolvedAgentPath = Resolve-TraceDeckPath -PathValue $AgentPath
    $resolvedDataDir = Resolve-TraceDeckPath -PathValue $DataDir
    $resolvedLogDir = Resolve-TraceDeckPath -PathValue $LogDir
    $resolvedOutboxDir = Resolve-TraceDeckPath -PathValue $OutboxDir
    $resolvedOutputRoot = Resolve-TraceDeckPath -PathValue $OutputRoot
    New-Item -ItemType Directory -Force -Path $resolvedDataDir, $resolvedLogDir, $resolvedOutboxDir, $resolvedOutputRoot | Out-Null

    $bucket = Read-PolicyValue -ResolvedConfigPath $resolvedConfigPath -Name "bucket"
    if ([string]::IsNullOrWhiteSpace($bucket)) {
        throw "Unable to read archive bucket from $resolvedConfigPath"
    }

    Invoke-TraceDeckLoggedCommand -Label "Run one live S3 archive upload cycle" -Command {
        & $resolvedAgentPath run `
            --config $resolvedConfigPath `
            --data-dir $resolvedDataDir `
            --log-dir $resolvedLogDir `
            --outbox-dir $resolvedOutboxDir `
            --once `
            --archive-once `
            --archive-dry-run=false `
            --alert-dry-run=true `
            --log-level debug
    }

    $deadline = (Get-Date).AddSeconds($WaitSeconds)
    $latest = $null
    while ((Get-Date) -lt $deadline) {
        $latest = Get-ChildItem -LiteralPath (Join-Path $resolvedOutboxDir "archive") -Recurse -File -Filter "*.jsonl.gz" -ErrorAction SilentlyContinue |
            Sort-Object LastWriteTime -Descending |
            Select-Object -First 1
        if ($latest) {
            break
        }
        Start-Sleep -Seconds 2
    }
    if (-not $latest) {
        throw "No local archive batch was written under $resolvedOutboxDir"
    }

    $key = ""
    $archiveRoot = Join-Path $resolvedOutboxDir "archive"
    $relative = Get-TraceDeckRelativePath -RootPath $archiveRoot -ChildPath $latest.FullName
    $keyCandidate = $relative -replace "\\", "/"
    $matches = aws s3api list-objects-v2 --bucket $bucket --region ap-south-1 --query "reverse(sort_by(Contents,&LastModified))[?contains(Key, '$($latest.BaseName)')].[Key,LastModified,Size] | [0]" --output json | ConvertFrom-Json
    if ($matches -and $matches.Count -ge 1) {
        $key = [string]$matches[0]
    }
    if ([string]::IsNullOrWhiteSpace($key)) {
        $objects = aws s3api list-objects-v2 --bucket $bucket --region ap-south-1 --max-items 10 --query "reverse(sort_by(Contents,&LastModified))[].[Key,LastModified,Size]" --output json | ConvertFrom-Json
        $candidate = $objects | Select-Object -First 1
        if ($candidate) {
            $key = [string]$candidate[0]
        }
    }
    if ([string]::IsNullOrWhiteSpace($key)) {
        throw "S3 upload could not be confirmed in bucket $bucket"
    }

    $report = [pscustomobject]@{
        status = "ok"
        bucket = $bucket
        confirmed_s3_key = $key
        local_archive = $latest.FullName
        local_archive_relative = $keyCandidate
        checked_at = (Get-Date).ToUniversalTime().ToString("o")
    }
    $reportPath = Join-Path $resolvedOutputRoot ("live-s3-archive-{0}.json" -f (Get-Date -Format "yyyyMMdd-HHmmss"))
    $report | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $reportPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Live S3 archive upload confirmed report=$reportPath key=$key"
    $report | ConvertTo-Json -Depth 5
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
