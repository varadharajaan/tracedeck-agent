param(
    [string]$Addr = "127.0.0.1:18294"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "newman-phase113" -LogRoot "logs/local/newman" | Out-Null

$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$runRoot = "data/local/newman/phase113/$timestamp"
$pidPath = "$runRoot/tracedeck-backend.pid"
$dataPath = "$runRoot/backend-state.json"
$reportPath = "$runRoot/newman-report.json"
$webPushDir = "data/local/webpush"
$webPushBackupDir = "$runRoot/webpush-backup"
$webPushFileNames = @("subscriptions.json", "vapid-public.key", "vapid-private.key")

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
    $newman = Get-Command newman -ErrorAction SilentlyContinue
    if (-not $newman) {
        throw "newman is not installed or not on PATH"
    }

    Backup-WebPushState

    Invoke-TraceDeckLoggedCommand -Label "Generate Phase 113 Web Push keys" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/generate-webpush-vapid-keys.ps1
    }

    $baseUrl = "http://$Addr"
    Start-TraceDeckDashboardDemo -ListenAddr $Addr -RelativePidPath $pidPath -RelativeDataPath $dataPath

    $reportDir = Split-Path -Parent (Join-Path $script:TraceDeckRepoRoot $reportPath)
    New-Item -ItemType Directory -Force -Path $reportDir | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Run Newman Phase 113 collection" -Command {
        newman run ./postman/tracedeck-backend-phase113.postman_collection.json --env-var "baseUrl=$baseUrl" --reporters "cli,json" --reporter-json-export $reportPath
    }

    if (-not (Test-Path (Join-Path $script:TraceDeckRepoRoot $reportPath))) {
        throw "Expected Newman report was not created: $reportPath"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Newman Phase 113 report: $reportPath"
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
