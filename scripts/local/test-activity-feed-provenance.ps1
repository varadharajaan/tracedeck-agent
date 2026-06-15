param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$TenantID = "family-varadha",
    [switch]$SkipLive
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-activity-feed-provenance" -LogRoot "logs/local/test" | Out-Null

try {
    $cleanBaseUrl = $BaseUrl.TrimEnd("/")
    $demoSourceKind = "demo_seed"
    $demoMediaApp = "VLC media player"

    Invoke-TraceDeckLoggedCommand -Label "Focused activity feed provenance Go tests" -Command {
        go test ./backend/internal/store ./backend/internal/api -run "TestTenantActivityFeedFiltersRiskDeliveryAndTelemetry|TestHostDashboardRiskEndpoints" -count=1
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    if (-not $SkipLive) {
        Write-TraceDeckLog -Level "INFO" -Message "Checking live activity feed provenance at $cleanBaseUrl"
        $health = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/health"
        if ($health.status -ne "ok") {
            throw "Expected backend health status ok."
        }
        $devices = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/devices"
        if ($devices.count -lt 1) {
            throw "Expected at least one device on live backend."
        }
        $deviceID = $devices.items[0].device_id
        $activity = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/tenants/$TenantID/activity-feed?device_id=$deviceID&limit=10"
        $activityJson = $activity | ConvertTo-Json -Depth 20
        if ($activity.filters.include_demo -or $activityJson.Contains($demoMediaApp) -or $activityJson.Contains($demoSourceKind)) {
            throw "Default tenant activity feed leaked demo evidence for $deviceID."
        }
        $demoActivity = Invoke-RestMethod -Method "GET" -Uri "$cleanBaseUrl/api/v1/tenants/$TenantID/activity-feed?device_id=$deviceID&limit=10&include_demo=true"
        $demoActivityJson = $demoActivity | ConvertTo-Json -Depth 20
        if (-not $demoActivity.filters.include_demo -or -not $demoActivityJson.Contains($demoMediaApp) -or -not $demoActivityJson.Contains($demoSourceKind)) {
            throw "Expected opt-in tenant activity feed to return labelled demo evidence for $deviceID."
        }
        Write-TraceDeckLog -Level "INFO" -Message "Live activity feed provenance passed for device=$deviceID"
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
