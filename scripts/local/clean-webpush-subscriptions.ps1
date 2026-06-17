param(
    [string]$SubscriptionFile = "data/local/webpush/subscriptions.json",
    [string]$OutputRoot = "data/local/webpush"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "clean-webpush-subscriptions" -LogRoot "logs/local/webpush" | Out-Null

try {
    $subscriptionPath = Join-Path $script:TraceDeckRepoRoot $SubscriptionFile
    $resolvedOutputRoot = Join-Path $script:TraceDeckRepoRoot $OutputRoot
    New-Item -ItemType Directory -Force -Path $resolvedOutputRoot | Out-Null

    if (-not (Test-Path -LiteralPath $subscriptionPath)) {
        Write-TraceDeckLog -Level "WARN" -Message "Subscription file does not exist: $SubscriptionFile"
        Complete-TraceDeckScriptLog
        return
    }

    $raw = Get-Content -LiteralPath $subscriptionPath -Raw
    $parsed = $raw | ConvertFrom-Json
    $subscriptions = @()
    if ($parsed.PSObject.Properties.Name -contains "subscriptions") {
        $subscriptions = @($parsed.subscriptions)
    }
    else {
        $subscriptions = @($parsed)
    }

    $kept = @()
    $removed = @()
    foreach ($subscription in $subscriptions) {
        $endpoint = [string]$subscription.endpoint
        $p256dh = [string]$subscription.keys.p256dh
        $auth = [string]$subscription.keys.auth
        $reason = ""
        if ([string]::IsNullOrWhiteSpace($endpoint) -or [string]::IsNullOrWhiteSpace($p256dh) -or [string]::IsNullOrWhiteSpace($auth)) {
            $reason = "missing_required_fields"
        }
        elseif ($endpoint -match "^https://push\.example\.test(/|$)") {
            $reason = "manual_example_endpoint"
        }

        if ($reason) {
            $endpointHost = "invalid"
            try {
                if (-not [string]::IsNullOrWhiteSpace($endpoint)) {
                    $endpointHost = ([Uri]$endpoint).Host
                }
            }
            catch {
                $endpointHost = "invalid"
            }
            $removed += [ordered]@{
                reason = $reason
                endpoint_host = $endpointHost
            }
            continue
        }

        $kept += [ordered]@{
            endpoint = $endpoint
            keys = [ordered]@{
                p256dh = $p256dh
                auth = $auth
            }
        }
    }

    [ordered]@{ subscriptions = $kept } | ConvertTo-Json -Depth 6 | Set-Content -LiteralPath $subscriptionPath -Encoding UTF8

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $reportPath = Join-Path $resolvedOutputRoot "webpush-cleanup-$timestamp.json"
    $report = [ordered]@{
        generated_at = (Get-Date).ToString("o")
        subscription_file = $SubscriptionFile
        kept = $kept.Count
        removed = $removed.Count
        removed_reasons = $removed
        privacy_boundary = "cleanup report stores endpoint host/reason only; no push endpoint tokens, credentials, cookies, screenshots, raw URLs, or private content"
    }
    $report | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $reportPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Web Push subscription cleanup kept=$($kept.Count) removed=$($removed.Count) report=$reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
