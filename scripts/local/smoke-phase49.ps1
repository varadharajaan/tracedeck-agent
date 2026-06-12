param(
    [string]$Addr = "127.0.0.1:18169"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase49" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase49/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$layoutRoot = "data/local/dashboard-layout/phase49-smoke/$timestamp"

function Start-TraceDeckDashboardDemo {
    param([string]$ListenAddr, [string]$RelativePidPath, [string]$RelativeDataPath)

    Write-TraceDeckLog -Level "INFO" -Message "Starting dashboard demo helper addr=$ListenAddr pid_path=$RelativePidPath"
    $helper = Start-Process -FilePath "powershell" -ArgumentList @(
        "-NoProfile",
        "-ExecutionPolicy", "Bypass",
        "-File", "./scripts/local/start-dashboard-demo.ps1",
        "-Addr", $ListenAddr,
        "-PidPath", $RelativePidPath,
        "-DataPath", $RelativeDataPath
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -PassThru

    $baseUrl = "http://$ListenAddr"
    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $RelativePidPath
    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        if ((Test-Path $pidFullPath)) {
            try {
                $devices = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/devices"
                if ($devices.count -ge 1) {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper completed readiness addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch { Start-Sleep -Milliseconds 500 }
        }
        elseif ($helper.HasExited -and $helper.ExitCode -ne 0) {
            throw "Dashboard demo helper failed with exit code $($helper.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo helper did not seed devices at $baseUrl"
}

try {
    $baseUrl = "http://$Addr"
    $tenantID = "family-varadha"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $dashboard = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/"
    foreach ($expected in @(
        "Notification Preference Center",
        "Preference Rule Matrix",
        "Study-Safe Suppression",
        "Preference Owner Actions",
        "notification-preference-section",
        "notification-preference-rule-list",
        "notification-preference-suppression-list",
        "notification-preference-action-list"
    )) {
        if ($dashboard.Content -notmatch [regex]::Escape($expected)) {
            throw "Expected Phase 49 dashboard marker '$expected'."
        }
    }

    $preferences = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/tenants/$tenantID/notification-preferences"
    if ($preferences.summary.rules_total -lt 3 -or -not $preferences.summary.email_enabled -or -not $preferences.summary.push_enabled -or -not $preferences.summary.dashboard_enabled) {
        throw "Expected notification preferences to expose seeded rules and channel coverage."
    }
    if ($preferences.summary.study_suppression_rules -lt 1 -or -not $preferences.quiet_hours.enabled -or -not $preferences.escalation.enabled) {
        throw "Expected study suppression, quiet hours, and escalation preference proof."
    }
    if ($preferences.privacy_boundary -notmatch "no passwords" -or $preferences.privacy_boundary -notmatch "screenshots") {
        throw "Expected strict notification preference privacy boundary."
    }
    $serialized = ($preferences | ConvertTo-Json -Depth 20).ToLowerInvariant()
    foreach ($forbidden in @("smtp_password", "provider_secret", "screenshot_bytes", "raw_url", "alert_body")) {
        if ($serialized.Contains($forbidden)) {
            throw "Notification preferences leaked forbidden field marker '$forbidden'."
        }
    }

    $body = @{
        digest_cadence = "daily"
        quiet_hours = @{
            enabled = $true
            start_local = "21:30"
            end_local = "06:00"
            timezone = "Asia/Calcutta"
        }
        escalation = @{
            enabled = $true
            after_minutes = 10
            repeat_every_minutes = 20
            max_repeats = 3
            channels = @("email", "push")
            owner = "parent escalation"
        }
        rules = @(
            @{
                name = "High-risk software immediate alert"
                event_type = "risky_software"
                severity = "high"
                channels = @("email", "push", "dashboard")
                mode = "immediate"
                recipient_group = "parent escalation"
                quiet_hours_bypass = $true
                paid_tier = "family_pro"
                delivery_sla = "10 minutes"
                next_action = "Verify delivery proof before relying on this rule."
                retention_evidence = "metadata-only alert and delivery proof"
            },
            @{
                name = "Study-safe digest"
                event_type = "non_study_youtube"
                severity = "low"
                channels = @("dashboard")
                mode = "silent"
                recipient_group = "dashboard archive"
                suppression_label = "study topics suppressed"
                study_safe = $true
                paid_tier = "free"
                delivery_sla = "dashboard only"
                next_action = "Keep study-safe learning out of noisy alert paths."
                retention_evidence = "category metadata only"
            }
        )
    } | ConvertTo-Json -Depth 12
    $updated = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/tenants/$tenantID/notification-preferences" -ContentType "application/json" -Body $body
    if ($updated.digest_cadence -ne "daily" -or $updated.summary.rules_total -ne 2 -or $updated.summary.silent_rules -ne 1) {
        throw "Expected notification preference update to apply typed policy changes."
    }

    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-layout.ps1 -BaseUrl $baseUrl -OutputRoot $layoutRoot

    Write-TraceDeckLog -Level "INFO" -Message "Phase 49 notification preference smoke passed addr=$Addr score=$($updated.summary.preference_score) rules=$($updated.summary.rules_total)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
