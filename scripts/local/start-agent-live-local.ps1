param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$TenantId = "family-varadha",
    [string]$Profile = "ai-btech-student",
    [string]$CollectionInterval = "10m",
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$ConfigPath = "data/local/config/tracedeck-live-this-machine.yaml",
    [string]$DataDir = "data/local/agent-live",
    [string]$LogDir = "logs/local/agent-live",
    [string]$OutboxDir = "data/local/outbox-live",
    [string]$PidPath = "data/local/agent-live/tracedeck-agent-live.pid",
    [string]$WebPushSubscriptionFile = "data/local/webpush/subscriptions.json",
    [string]$WebPushVAPIDPublicKeyFile = "data/local/webpush/vapid-public.key",
    [string]$WebPushVAPIDPrivateKeyFile = "data/local/webpush/vapid-private.key",
    [string]$WebPushVAPIDSubject = "mailto:varathu09@gmail.com",
    [switch]$SkipBuild,
    [switch]$LiveArchive,
    [switch]$LiveAlerts,
    [switch]$LivePush
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "start-agent-live-local" -LogRoot "logs/local/agent-live" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Get-TraceDeckDeviceId {
    $raw = (& hostname).Trim().ToLowerInvariant()
    $clean = ($raw -replace "[^a-z0-9_-]", "-").Trim("-")
    if ([string]::IsNullOrWhiteSpace($clean)) {
        return "this-machine"
    }
    return $clean
}

function Stop-ExistingAgent {
    param([string]$ResolvedPidPath)

    if (-not (Test-Path -LiteralPath $ResolvedPidPath)) {
        return
    }
    $pidText = (Get-Content -LiteralPath $ResolvedPidPath -Raw).Trim()
    if ($pidText -notmatch "^\d+$") {
        Remove-Item -LiteralPath $ResolvedPidPath -Force -ErrorAction SilentlyContinue
        return
    }
    $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
    if ($process) {
        Write-TraceDeckLog -Level "INFO" -Message "Stopping existing live agent pid=$pidText"
        Stop-Process -Id ([int]$pidText) -Force
        Start-Sleep -Milliseconds 500
    }
    Remove-Item -LiteralPath $ResolvedPidPath -Force -ErrorAction SilentlyContinue
}

function Get-WebPushPublicKey {
    param([string]$PublicKeyPath)

    if (-not (Test-Path -LiteralPath $PublicKeyPath)) {
        throw "Web Push public key file is required for -LivePush: $PublicKeyPath"
    }
    $value = (Get-Content -LiteralPath $PublicKeyPath -Raw).Trim()
    if ([string]::IsNullOrWhiteSpace($value)) {
        throw "Web Push public key file is empty: $PublicKeyPath"
    }
    return $value
}

function Ensure-BackendTenant {
    param(
        [string]$ResolvedBaseUrl,
        [string]$ResolvedTenantId,
        [string]$ResolvedProfile
    )

    try {
        $health = Invoke-RestMethod -Method "GET" -Uri "$ResolvedBaseUrl/health" -TimeoutSec 10
        if ($health.status -ne "ok") {
            throw "backend health returned $($health.status)"
        }
        $tenantBody = @{
            tenant_id = $ResolvedTenantId
            name = "Family Varadha"
            plan_id = "family_pro"
            retention_tier_id = "family_cloud_90_365_archive"
            primary_profile = $ResolvedProfile
        } | ConvertTo-Json -Compress
        Invoke-RestMethod -Method "POST" -Uri "$ResolvedBaseUrl/api/v1/tenants" -Headers @{ "Content-Type" = "application/json" } -Body $tenantBody -TimeoutSec 10 | Out-Null
        Write-TraceDeckLog -Level "INFO" -Message "Backend tenant ensured tenant=$ResolvedTenantId base_url=$ResolvedBaseUrl"
    }
    catch {
        Write-TraceDeckLog -Level "WARN" -Message "Backend tenant ensure skipped or failed: $($_.Exception.Message)"
    }
}

try {
    $resolvedAgentPath = Resolve-TraceDeckPath -PathValue $AgentPath
    $resolvedConfigPath = Resolve-TraceDeckPath -PathValue $ConfigPath
    $resolvedDataDir = Resolve-TraceDeckPath -PathValue $DataDir
    $resolvedLogDir = Resolve-TraceDeckPath -PathValue $LogDir
    $resolvedOutboxDir = Resolve-TraceDeckPath -PathValue $OutboxDir
    $resolvedPidPath = Resolve-TraceDeckPath -PathValue $PidPath
    $resolvedWebPushSubscriptionFile = Resolve-TraceDeckPath -PathValue $WebPushSubscriptionFile
    $resolvedWebPushPublicKeyFile = Resolve-TraceDeckPath -PathValue $WebPushVAPIDPublicKeyFile
    $resolvedWebPushPrivateKeyFile = Resolve-TraceDeckPath -PathValue $WebPushVAPIDPrivateKeyFile
    $deviceId = Get-TraceDeckDeviceId
    $hostName = (& hostname).Trim()

    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedAgentPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedConfigPath) | Out-Null
    New-Item -ItemType Directory -Force -Path $resolvedDataDir | Out-Null
    New-Item -ItemType Directory -Force -Path $resolvedLogDir | Out-Null
    New-Item -ItemType Directory -Force -Path $resolvedOutboxDir | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedPidPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $resolvedWebPushSubscriptionFile) | Out-Null

    if (-not $SkipBuild) {
        Invoke-TraceDeckLoggedCommand -Label "Build hidden Windows agent executable" -Command {
            go build -trimpath -ldflags "-H=windowsgui" -o $resolvedAgentPath ./agent/cmd/tracedeck-agent
        }
    }

    $pushProvider = "none"
    $pushSubscriptionFile = ""
    $pushPublicKey = ""
    $pushPrivateKeyFile = ""
    $pushSubject = ""
    if ($LivePush) {
        $pushProvider = "web_push"
        $pushSubscriptionFile = $resolvedWebPushSubscriptionFile
        $pushPublicKey = Get-WebPushPublicKey -PublicKeyPath $resolvedWebPushPublicKeyFile
        if (-not (Test-Path -LiteralPath $resolvedWebPushPrivateKeyFile)) {
            throw "Web Push private key file is required for -LivePush: $resolvedWebPushPrivateKeyFile"
        }
        if (-not (Test-Path -LiteralPath $resolvedWebPushSubscriptionFile)) {
            throw "Web Push subscription file is required for -LivePush: $resolvedWebPushSubscriptionFile"
        }
        $pushPrivateKeyFile = $resolvedWebPushPrivateKeyFile
        $pushSubject = $WebPushVAPIDSubject
    }

    $config = @"
tenant_id: $TenantId
device_id: $deviceId
profile: $Profile

collection:
  transparency_mode: visible_indicator_required
  browser:
    url_mode: domain_only
    collect_page_title: false
    youtube_classification: enabled
    youtube_video_id_mode: hashed
  foreground_app:
    enabled: true
    window_title_mode: none
  software:
    enabled: true
    inventory_mode: metadata_only
  media:
    collect_file_name: true
    collect_file_path: true
    path_mode: full_path
  sensitive_capabilities:
    credentials: deny
    keystrokes: deny
    cookies: deny
    tokens: deny
    private_messages: deny
    screenshots: deny

retention:
  local_ttl_days: 90
  max_local_storage_mb: 2048

archive:
  enabled: true
  provider: s3
  bucket: tracedeck-agent-family-varadha-996335889295-ap-south-1
  prefix_template: tenants/{tenant_id}/devices/{device_id}/hosts/{host_name}/date={yyyy}-{mm}-{dd}/hour={hh}/
  upload_interval: 1h
  retry_when_online: true
  storage_class_days:
    standard: 90
    standard_ia_until: 365
    archive_after: 365

backend_sync:
  enabled: true
  base_url: $BaseUrl
  batch_limit: 100
  request_timeout: 10s

observability:
  opentelemetry:
    enabled: false
    protocol: otlp_http_json
    endpoint: http://127.0.0.1:4318/v1/logs
    batch_limit: 100
    request_timeout: 5s
    retry:
      max_attempts: 2

alerts:
  enabled: true
  email:
    provider: ses
    from: varathu09@gmail.com
    to:
      - varathu09@gmail.com
    min_severity: medium
    cooldown_minutes: 30
  push:
    provider: $pushProvider
    subscription_file: '$pushSubscriptionFile'
    vapid_public_key: '$pushPublicKey'
    vapid_private_key_file: '$pushPrivateKeyFile'
    vapid_subject: '$pushSubject'
    ttl_seconds: 3600
    min_severity: medium
    cooldown_minutes: 30

study_apps:
  - Code.exe
  - python.exe
  - java.exe
  - idea64.exe
  - pycharm64.exe
  - jupyter.exe
  - chrome.exe
  - msedge.exe
  - brave.exe

blocked_apps:
  - vlc.exe
  - qbittorrent.exe
  - utorrent.exe
  - steam.exe
  - epicgameslauncher.exe

ignored_apps:
  - System
  - Registry
  - svchost.exe

allowed_domains:
  - udemy.com
  - coursera.org
  - github.com
  - stackoverflow.com
  - docs.python.org
  - openai.com
  - huggingface.co
  - kaggle.com
  - tensorflow.org
  - pytorch.org
  - microsoft.com
  - learn.microsoft.com

blocked_domains:
  - instagram.com
  - netflix.com
  - primevideo.com
  - hotstar.com
  - steampowered.com
  - epicgames.com

warn_categories:
  - video-streaming
  - social-media
  - gaming
  - shopping

critical_categories:
  - adult-content
  - torrent
  - malware
  - proxy-vpn

thresholds:
  max_video_minutes_per_day: 60
  max_social_minutes_per_day: 30
  max_unknown_app_minutes_per_day: 45
  late_night_usage_start: "23:30"
  late_night_usage_end: "05:00"

youtube_study_keywords:
  - python
  - system design
  - math
  - maths
  - machine learning
  - artificial intelligence
  - coding
  - java
  - data structures
  - algorithms

alert_rules:
  adult_content:
    enabled: true
    severity: critical
  blocked_app_opened:
    enabled: true
    severity: high
  blocked_domain_opened:
    enabled: true
    severity: high
  media_player_used:
    enabled: true
    severity: high
    include_media_file_metadata: true
  unknown_software_installed:
    enabled: true
    severity: high
  risky_software_detected:
    enabled: true
    severity: high
  non_study_youtube:
    enabled: true
    severity: medium
    threshold_minutes_per_day: 30
  excessive_video_usage:
    enabled: true
    severity: medium
  excessive_social_usage:
    enabled: true
    severity: medium
  late_night_usage:
    enabled: true
    severity: medium
"@
    Set-Content -LiteralPath $resolvedConfigPath -Value $config -Encoding UTF8

    Invoke-TraceDeckLoggedCommand -Label "Validate live agent config" -Command {
        & $resolvedAgentPath validate-config --config $resolvedConfigPath
    }

    Ensure-BackendTenant -ResolvedBaseUrl $BaseUrl -ResolvedTenantId $TenantId -ResolvedProfile $Profile
    Stop-ExistingAgent -ResolvedPidPath $resolvedPidPath

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $stdoutPath = Join-Path $resolvedLogDir "agent-live-$timestamp.out.log"
    $stderrPath = Join-Path $resolvedLogDir "agent-live-$timestamp.err.log"
    $archiveDryRunValue = if ($LiveArchive) { "false" } else { "true" }
    $alertDryRunValue = if ($LiveAlerts) { "false" } else { "true" }
    $arguments = @(
        "run",
        "--config", "`"$resolvedConfigPath`"",
        "--data-dir", "`"$resolvedDataDir`"",
        "--log-dir", "`"$resolvedLogDir`"",
        "--outbox-dir", "`"$resolvedOutboxDir`"",
        "--collection-interval", $CollectionInterval,
        "--max-cycles", "0",
        "--archive-dry-run=$archiveDryRunValue",
        "--alert-dry-run=$alertDryRunValue",
        "--log-level", "debug"
    )
    $process = Start-Process -FilePath $resolvedAgentPath -ArgumentList $arguments -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru
    Set-Content -LiteralPath $resolvedPidPath -Value $process.Id -Encoding UTF8

    [pscustomobject]@{
        status = "started"
        pid = $process.Id
        tenant_id = $TenantId
        device_id = $deviceId
        host_name = $hostName
        base_url = $BaseUrl
        config_path = $resolvedConfigPath
        data_dir = $resolvedDataDir
        log_dir = $resolvedLogDir
        outbox_dir = $resolvedOutboxDir
        pid_path = $resolvedPidPath
        collection_interval = $CollectionInterval
        archive_dry_run = -not [bool]$LiveArchive
        alert_dry_run = -not [bool]$LiveAlerts
    } | ConvertTo-Json -Depth 4

    Write-TraceDeckLog -Level "INFO" -Message "Live local agent started pid=$($process.Id) tenant=$TenantId device=$deviceId interval=$CollectionInterval archive_dry_run=$archiveDryRunValue alert_dry_run=$alertDryRunValue"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
