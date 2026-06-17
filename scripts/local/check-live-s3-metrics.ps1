param(
    [string]$ConfigPath = "data/local/config/tracedeck-live-this-machine.yaml",
    [string]$OutputRoot = "data/local/cloud",
    [string]$Region = "ap-south-1",
    [int]$MaxItems = 5000
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "check-live-s3-metrics" -LogRoot "logs/local/cloud" | Out-Null

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

function Get-TraceDeckPropertyValue {
    param(
        [object]$Object,
        [string]$Name,
        [object]$DefaultValue = $null
    )

    if ($null -eq $Object) {
        return $DefaultValue
    }
    $property = $Object.PSObject.Properties[$Name]
    if (-not $property) {
        return $DefaultValue
    }
    return $property.Value
}

try {
    $resolvedConfigPath = Resolve-TraceDeckPath -PathValue $ConfigPath
    $resolvedOutputRoot = Resolve-TraceDeckPath -PathValue $OutputRoot
    New-Item -ItemType Directory -Force -Path $resolvedOutputRoot | Out-Null

    $tenantId = Read-PolicyValue -ResolvedConfigPath $resolvedConfigPath -Name "tenant_id"
    $deviceId = Read-PolicyValue -ResolvedConfigPath $resolvedConfigPath -Name "device_id"
    $bucket = Read-PolicyValue -ResolvedConfigPath $resolvedConfigPath -Name "bucket"
    if ([string]::IsNullOrWhiteSpace($tenantId) -or [string]::IsNullOrWhiteSpace($deviceId) -or [string]::IsNullOrWhiteSpace($bucket)) {
        throw "Unable to read tenant_id/device_id/bucket from $resolvedConfigPath"
    }

    $prefix = "tenants/$tenantId/devices/$deviceId/"
    $raw = aws s3api list-objects-v2 `
        --bucket $bucket `
        --region $Region `
        --prefix $prefix `
        --max-items $MaxItems `
        --output json
    $listing = $raw | ConvertFrom-Json
    $contents = Get-TraceDeckPropertyValue -Object $listing -Name "Contents" -DefaultValue @()
    $objects = @($contents)
    $archiveObjects = @($objects | Where-Object { $_.Key -like "*.jsonl.gz" })
    $ordered = @($archiveObjects | Sort-Object LastModified)
    $latest = $ordered | Select-Object -Last 1
    $earliest = $ordered | Select-Object -First 1
    $hours = @($archiveObjects | ForEach-Object {
        if ($_.Key -match "date=([0-9]{4}-[0-9]{2}-[0-9]{2})/hour=([0-9]{2})") {
            "$($Matches[1])T$($Matches[2])"
        }
    } | Sort-Object -Unique)

    $report = [pscustomobject]@{
        status = if ($archiveObjects.Count -gt 0) { "ok" } else { "empty" }
        bucket = $bucket
        prefix = $prefix
        archive_object_count = $archiveObjects.Count
        total_bytes = [int64](($archiveObjects | Measure-Object -Property Size -Sum).Sum)
        distinct_archive_hours = $hours.Count
        first_last_modified = if ($earliest) { [string]$earliest.LastModified } else { "" }
        latest_last_modified = if ($latest) { [string]$latest.LastModified } else { "" }
        latest_key = if ($latest) { [string](Get-TraceDeckPropertyValue -Object $latest -Name "Key" -DefaultValue "") } else { "" }
        latest_size = if ($latest) { [int64](Get-TraceDeckPropertyValue -Object $latest -Name "Size" -DefaultValue 0) } else { 0 }
        truncated = [bool](Get-TraceDeckPropertyValue -Object $listing -Name "IsTruncated" -DefaultValue $false)
        checked_at = (Get-Date).ToUniversalTime().ToString("o")
    }

    $reportPath = Join-Path $resolvedOutputRoot ("live-s3-metrics-{0}.json" -f (Get-Date -Format "yyyyMMdd-HHmmss"))
    $report | ConvertTo-Json -Depth 5 | Set-Content -LiteralPath $reportPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Live S3 metrics checked report=$reportPath count=$($report.archive_object_count) hours=$($report.distinct_archive_hours)"
    $report | ConvertTo-Json -Depth 5
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
