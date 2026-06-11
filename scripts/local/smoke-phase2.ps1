param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase2" -LogRoot "logs/local/smoke" | Out-Null

$sleeper = $null

try {
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase2"
    $outboxRoot = Join-Path $script:TraceDeckRepoRoot "data/local/outbox/smoke-phase2"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $outboxRoot | Out-Null

    $policyPath = Join-Path $smokeRoot "phase2-policy.yaml"
    $policy = Get-Content -Raw (Join-Path $script:TraceDeckRepoRoot "examples/policies/ai-btech-student.yaml")
    $policy = $policy -replace "blocked_apps:\r?\n", "blocked_apps:`r`n  - powershell.exe`r`n"
    Set-Content -Path $policyPath -Value $policy

    $sleeper = Start-Process -FilePath "powershell" -ArgumentList @("-NoProfile", "-Command", "Start-Sleep -Seconds 45") -WindowStyle Hidden -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started smoke sleeper process: $($sleeper.Id)"

    Invoke-TraceDeckLoggedCommand -Label "Run local agent archive and alert dry-run" -Command {
        go run ./agent/cmd/tracedeck-agent run --once --config $policyPath --data-dir $smokeRoot --log-dir ./logs/local/agent --outbox-dir $outboxRoot --process-limit 512 --archive-once --archive-dry-run --alert-once --alert-dry-run
    }

    $archiveFiles = Get-ChildItem -Path (Join-Path $outboxRoot "archive") -Filter "*.jsonl.gz" -File -ErrorAction SilentlyContinue
    if (-not $archiveFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected archive outbox batch was not created."
        exit 1
    }
    Write-TraceDeckLog -Level "INFO" -Message "Archive outbox batch count: $($archiveFiles.Count)"

    $alertFiles = Get-ChildItem -Path (Join-Path $outboxRoot "alerts") -Filter "*.json" -File -ErrorAction SilentlyContinue
    if (-not $alertFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected alert outbox notification was not created."
        exit 1
    }
    Write-TraceDeckLog -Level "INFO" -Message "Alert outbox notification count: $($alertFiles.Count)"

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($sleeper -and -not $sleeper.HasExited) {
        Stop-Process -Id $sleeper.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped smoke sleeper process: $($sleeper.Id)"
    }
}
