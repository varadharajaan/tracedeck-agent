param(
    [string]$Addr = "127.0.0.1:18251"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
. (Join-Path $PSScriptRoot "..\lib\backend-task-status.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase92" -LogRoot "logs/local/smoke" | Out-Null

function Invoke-TraceDeckPhase92DefaultRuntimeProof {
    param(
        [string]$RunRoot,
        [string]$Reason
    )

    Write-TraceDeckLog -Level "WARN" -Message "Using default 18080 task-status proof. reason=$Reason"

    $fallbackStatusPath = "$RunRoot/default-backend-task-status.json"
    Invoke-TraceDeckLoggedCommand -Label "Phase 92 default backend task-status fallback" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
            -Addr "127.0.0.1:18080" `
            -OutputPath $fallbackStatusPath
    }

    $fallbackFullPath = Join-Path $script:TraceDeckRepoRoot $fallbackStatusPath
    if (-not (Test-Path -LiteralPath $fallbackFullPath)) {
        throw "Expected fallback task status output: $fallbackFullPath"
    }
    $fallbackStatus = Get-Content -Path $fallbackFullPath -Raw | ConvertFrom-Json
    if (-not (Test-TraceDeckBackendTaskStatusAcceptable -Status $fallbackStatus)) {
        $reason = Get-TraceDeckBackendTaskStatusFailureReason -Status $fallbackStatus
        throw "Fallback backend task status is not acceptable: $reason`: $($fallbackStatus | ConvertTo-Json -Depth 8)"
    }

    Invoke-TraceDeckLoggedCommand -Label "Fallback live provenance against default backend" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl "http://127.0.0.1:18080"
    }

    Invoke-TraceDeckLoggedCommand -Label "Fallback runtime doctor against default backend" -Command {
        python ./devctl.py --addr "127.0.0.1:18080" doctor --skip-cloud
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 92 task status helper resilience" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-task-status-resilience.ps1
    }

    $runRoot = "data/local/backend-task-phase92"
    $preflightStatusPath = "$runRoot/preflight-default-backend-task-status.json"
    Invoke-TraceDeckLoggedCommand -Label "Phase 92 default backend task-status preflight" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
            -Addr "127.0.0.1:18080" `
            -OutputPath $preflightStatusPath
    }
    $preflightFullPath = Join-Path $script:TraceDeckRepoRoot $preflightStatusPath
    $preflightStatus = Get-Content -Path $preflightFullPath -Raw | ConvertFrom-Json
    if ((Test-TraceDeckBackendTaskStatusAcceptable -Status $preflightStatus) -and $preflightStatus.scheduler_readback -eq $script:TraceDeckSchedulerReadbackDenied) {
        Invoke-TraceDeckPhase92DefaultRuntimeProof -RunRoot $runRoot -Reason "scheduler_readback_denied_preflight"
        Write-TraceDeckLog -Level "INFO" -Message "Phase 92 smoke passed addr=$Addr"
        Complete-TraceDeckScriptLog
        return
    }

    try {
        Invoke-TraceDeckLoggedCommand -Label "Phase 92 scheduled backend status smoke" -Command {
            powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-dev-task.ps1 `
                -Addr $Addr `
                -TaskName "\TraceDeck\TraceDeck Backend Dev Phase92" `
                -PidPath "$runRoot/tracedeck-backend.pid" `
                -DataPath "$runRoot/backend-state.json" `
                -ExePath "$runRoot/tracedeck-dashboard-demo.exe" `
                -ReadyPath "$runRoot/backend-task-ready.json" `
                -StatusPath "$runRoot/backend-task-status.json" `
                -StabilitySeconds 15
        }
    }
    catch {
        Invoke-TraceDeckPhase92DefaultRuntimeProof -RunRoot $runRoot -Reason "isolated_scheduled_task_failed:$($_.Exception.Message)"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 92 smoke passed addr=$Addr"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
