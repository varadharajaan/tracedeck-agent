param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase2b" -LogRoot "logs/local/smoke" | Out-Null

$sleeper = $null

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase2b/$timestamp"
    $outboxRoot = Join-Path $script:TraceDeckRepoRoot "data/local/outbox/smoke-phase2b/$timestamp"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $outboxRoot | Out-Null

    $policyPath = Join-Path $smokeRoot "phase2b-policy.yaml"
    $policy = Get-Content -Raw (Join-Path $script:TraceDeckRepoRoot "examples/policies/ai-btech-student.yaml")
    $policy = $policy -replace "blocked_apps:\r?\n", "blocked_apps:`r`n  - powershell.exe`r`n"
    Set-Content -Path $policyPath -Value $policy

    $sleeper = Start-Process -FilePath "powershell" -ArgumentList @("-NoProfile", "-Command", "Start-Sleep -Seconds 45") -WindowStyle Hidden -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started continuous smoke sleeper process: $($sleeper.Id)"

    Invoke-TraceDeckLoggedCommand -Label "Run continuous agent for two cycles" -Command {
        go run ./agent/cmd/tracedeck-agent run --config $policyPath --data-dir $smokeRoot --log-dir ./logs/local/agent --outbox-dir $outboxRoot --process-limit 512 --collection-interval 1s --max-cycles 2 --archive-dry-run --alert-dry-run
    }

    $archiveFiles = @(Get-ChildItem -Path (Join-Path $outboxRoot "archive") -Filter "*.jsonl.gz" -File -ErrorAction SilentlyContinue)
    if (-not $archiveFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected continuous archive outbox batch was not created."
        exit 1
    }
    Write-TraceDeckLog -Level "INFO" -Message "Continuous archive outbox batch count: $($archiveFiles.Count)"

    $alertFiles = @(Get-ChildItem -Path (Join-Path $outboxRoot "alerts") -Filter "*.json" -File -ErrorAction SilentlyContinue)
    if (-not $alertFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected continuous alert outbox notification was not created."
        exit 1
    }
    Write-TraceDeckLog -Level "INFO" -Message "Continuous alert outbox notification count: $($alertFiles.Count)"

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($sleeper -and -not $sleeper.HasExited) {
        Stop-Process -Id $sleeper.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped continuous smoke sleeper process: $($sleeper.Id)"
    }
}
