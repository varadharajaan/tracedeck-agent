param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$TenantID = "family-varadha"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
. (Join-Path $PSScriptRoot "..\lib\http-constants.ps1")
Initialize-TraceDeckScriptLog -Name "test-live-server-provenance" -LogRoot "logs/local/test" | Out-Null

try {
    $cleanBaseUrl = $BaseUrl.TrimEnd("/")
    $demoSourceKind = "demo_seed"
    $demoMediaApp = "VLC media player"
    $includeDemoQuery = "include_demo=true"

    Write-TraceDeckLog -Level "INFO" -Message "Checking live server health at $cleanBaseUrl/health"
    $healthResponse = Invoke-WebRequest -UseBasicParsing -Uri "$cleanBaseUrl/health"
    if ($healthResponse.Headers[$TraceDeckHeaderCacheControl] -ne $TraceDeckCacheNoStore) {
        throw "Expected no-store cache header on live health endpoint."
    }
    $health = $healthResponse.Content | ConvertFrom-Json
    if ($health.status -ne "ok") {
        throw "Expected live server health status ok."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Checking live dashboard provenance markers"
    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$cleanBaseUrl/"
    if ($dashboard.Headers[$TraceDeckHeaderCacheControl] -ne $TraceDeckCacheNoStore) {
        throw "Expected no-store cache header on live dashboard."
    }
    foreach ($expected in @("theme-toggle-button", "server-status-light", "sourceBadge", "dashboard-page-nav", "browser-activity-button")) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected live dashboard marker '$expected'."
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Checking live browser activity provenance markers"
    $browserPage = Invoke-WebRequest -UseBasicParsing -Uri "$cleanBaseUrl/browser-activity"
    if ($browserPage.Headers[$TraceDeckHeaderCacheControl] -ne $TraceDeckCacheNoStore) {
        throw "Expected no-store cache header on live browser activity page."
    }
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

    Write-TraceDeckLog -Level "INFO" -Message "Checking live risk and delivery API provenance"
    $devices = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices"
    if ($devices.count -lt 1) {
        throw "Expected at least one live device."
    }
    $deviceID = $devices.items[0].device_id
    $policy = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices/$deviceID/policy-violations"
    $policyJson = $policy | ConvertTo-Json -Depth 20
    if ($policy.count -ne 0 -or $policyJson.Contains($demoMediaApp) -or $policyJson.Contains($demoSourceKind)) {
        throw "Default policy endpoint leaked demo evidence for $deviceID."
    }

    $deliveries = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices/$deviceID/alert-deliveries"
    $deliveryJson = $deliveries | ConvertTo-Json -Depth 20
    if ($deliveries.count -ne 0 -or $deliveryJson.Contains($demoSourceKind)) {
        throw "Default alert delivery endpoint leaked demo delivery evidence for $deviceID."
    }

    $activity = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/tenants/$TenantID/activity-feed?device_id=$deviceID&limit=10"
    $activityJson = $activity | ConvertTo-Json -Depth 20
    if ($activity.filters.include_demo -or $activityJson.Contains($demoMediaApp) -or $activityJson.Contains($demoSourceKind)) {
        throw "Default tenant activity feed leaked demo evidence for $deviceID."
    }

    $demoPolicy = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices/$deviceID/policy-violations?$includeDemoQuery"
    if ($demoPolicy.count -lt 1 -or $demoPolicy.items[0].source_kind -ne $demoSourceKind) {
        throw "Expected opt-in demo policy provenance for $deviceID."
    }
    $demoDeliveries = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices/$deviceID/alert-deliveries?$includeDemoQuery"
    if ($demoDeliveries.count -lt 1 -or $demoDeliveries.items[0].source_kind -ne $demoSourceKind) {
        throw "Expected opt-in demo delivery provenance for $deviceID."
    }
    $demoActivity = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/tenants/$TenantID/activity-feed?device_id=$deviceID&limit=10&$includeDemoQuery"
    $demoActivityJson = $demoActivity | ConvertTo-Json -Depth 20
    if (-not $demoActivity.filters.include_demo -or -not $demoActivityJson.Contains($demoMediaApp) -or -not $demoActivityJson.Contains($demoSourceKind)) {
        throw "Expected opt-in demo activity feed provenance for $deviceID."
    }

    Write-TraceDeckLog -Level "INFO" -Message "Live server provenance passed base_url=$cleanBaseUrl tenant=$TenantID browser_rows=$($viewer.summary.total) device=$deviceID"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
