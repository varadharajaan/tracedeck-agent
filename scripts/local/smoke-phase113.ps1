param(
    [string]$Addr = "127.0.0.1:18293"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase113" -LogRoot "logs/local/smoke" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$smokeRoot = "data/local/smoke-phase113/$timestamp"
$pidPath = "$smokeRoot/tracedeck-backend.pid"
$dataPath = "$smokeRoot/backend-state.json"
$webPushDir = "data/local/webpush"
$webPushBackupDir = "$smokeRoot/webpush-backup"
$webPushFileNames = @("subscriptions.json", "vapid-public.key", "vapid-private.key")
$subscriptionPath = "data/local/webpush/subscriptions.json"

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
        if (Test-Path $pidFullPath) {
            try {
                $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health"
                if ($health.status -eq "ok") {
                    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo helper ready addr=$ListenAddr helper_pid=$($helper.Id)"
                    return
                }
            }
            catch {
                Start-Sleep -Milliseconds 500
            }
        }
        elseif ($helper.HasExited -and $helper.ExitCode -ne 0) {
            throw "Dashboard demo helper failed with exit code $($helper.ExitCode)"
        }
        Start-Sleep -Milliseconds 500
    }
    throw "Dashboard demo helper did not become healthy at $baseUrl"
}

function Backup-WebPushState {
    $sourceDir = Join-Path $script:TraceDeckRepoRoot $webPushDir
    $backupDir = Join-Path $script:TraceDeckRepoRoot $webPushBackupDir
    New-Item -ItemType Directory -Force -Path $backupDir | Out-Null
    foreach ($fileName in $webPushFileNames) {
        $source = Join-Path $sourceDir $fileName
        if (Test-Path -LiteralPath $source) {
            Copy-Item -LiteralPath $source -Destination (Join-Path $backupDir $fileName) -Force
        }
    }
}

function Restore-WebPushState {
    $sourceDir = Join-Path $script:TraceDeckRepoRoot $webPushDir
    $backupDir = Join-Path $script:TraceDeckRepoRoot $webPushBackupDir
    New-Item -ItemType Directory -Force -Path $sourceDir | Out-Null
    foreach ($fileName in $webPushFileNames) {
        $source = Join-Path $sourceDir $fileName
        $backup = Join-Path $backupDir $fileName
        if (Test-Path -LiteralPath $backup) {
            Copy-Item -LiteralPath $backup -Destination $source -Force
        }
        elseif (Test-Path -LiteralPath $source) {
            Remove-Item -LiteralPath $source -Force
        }
    }
}

try {
    $baseUrl = "http://$Addr"
    Backup-WebPushState

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 113 Web Push keys" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/generate-webpush-vapid-keys.ps1
    }

    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $worker = Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/webpush-sw.js"
    if ($worker.StatusCode -ne 200 -or $worker.Headers["Content-Type"] -notmatch "application/javascript" -or $worker.Content -notmatch "showNotification") {
        throw "Expected Web Push service worker JavaScript."
    }

    $setup = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/api/v1/webpush/vapid-public-key"
    if ([string]::IsNullOrWhiteSpace($setup.public_key) -or $setup.subscription_url -ne "/api/v1/webpush/subscriptions" -or $setup.service_worker -ne "/webpush-sw.js") {
        throw "Expected Web Push setup metadata."
    }

    $subscription = @{
        endpoint = "https://push.example.test/phase113"
        keys = @{
            p256dh = "phase113-client-public-key"
            auth = "phase113-client-auth-secret"
        }
    } | ConvertTo-Json -Depth 4
    $subscribe = Invoke-RestMethod -Method "POST" -Uri "$baseUrl/api/v1/webpush/subscriptions" -Headers @{ "Content-Type" = "application/json" } -Body $subscription
    if ($subscribe.status -ne "subscribed" -or -not $subscribe.provider_configured) {
        throw "Expected Web Push subscription to be accepted."
    }
    $serializedResponse = ($subscribe | ConvertTo-Json -Depth 8).ToLowerInvariant()
    foreach ($forbidden in @("https://push.example.test/phase113", "phase113-client-auth-secret", "provider_secret", "smtp_password", "screenshot_bytes", "raw_url", "page_title", "alert_body")) {
        if ($serializedResponse.Contains($forbidden)) {
            throw "Web Push subscription response leaked forbidden marker '$forbidden'."
        }
    }

    $stored = Get-Content -Path (Join-Path $script:TraceDeckRepoRoot $subscriptionPath) -Raw | ConvertFrom-Json
    if (@($stored.subscriptions).Count -lt 1) {
        throw "Expected subscription to be written to $subscriptionPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 113 Web Push provider smoke passed addr=$Addr subscriptions=$($subscribe.subscriptions)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    Restore-WebPushState
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $pidPath -Addr $Addr
}
