param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$TenantID = "family-varadha"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-live-server-provenance" -LogRoot "logs/local/test" | Out-Null

try {
    $cleanBaseUrl = $BaseUrl.TrimEnd("/")

    Write-TraceDeckLog -Level "INFO" -Message "Checking live server health at $cleanBaseUrl/health"
    $health = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/health"
    if ($health.status -ne "ok") {
        throw "Expected live server health status ok."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Checking live dashboard provenance markers"
    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$cleanBaseUrl/"
    foreach ($expected in @("theme-toggle-button", "server-status-light", "sourceBadge", "dashboard-page-nav", "browser-activity-button")) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected live dashboard marker '$expected'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Checking live browser activity provenance markers"
    $browserPage = Invoke-WebRequest -UseBasicParsing -Uri "$cleanBaseUrl/browser-activity"
    foreach ($expected in @("TraceDeck Browser Activity", "<th>Source</th>", "sourceBadge", "server-status-light")) {
        if ($browserPage.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected live browser activity marker '$expected'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Checking live browser activity API provenance"
    $viewer = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/tenants/$TenantID/browser-activity?limit=25"
    if ($viewer.summary.total -lt 1 -or $viewer.items.Count -lt 1) {
        throw "Expected at least one browser activity row."
    }
    foreach ($field in @("source_kind", "evidence_scope", "evidence_detail")) {
        if ([string]::IsNullOrWhiteSpace($viewer.items[0].$field)) {
            throw "Expected browser activity field '$field'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Checking live delivery API provenance"
    $devices = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices"
    if ($devices.count -lt 1) {
        throw "Expected at least one live device."
    }
    $deviceID = $devices.items[0].device_id
    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices/$deviceID/alert-deliveries"
    if ($deliveries.count -lt 1 -or [string]::IsNullOrWhiteSpace($deliveries.items[0].source_kind)) {
        throw "Expected alert delivery provenance."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Live server provenance passed base_url=$cleanBaseUrl tenant=$TenantID browser_rows=$($viewer.summary.total) device=$deviceID"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
