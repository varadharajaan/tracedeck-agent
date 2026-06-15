param(
    [string]$Addr = "127.0.0.1:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
. (Join-Path $PSScriptRoot "..\lib\backend-task-status.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase93" -LogRoot "logs/local/smoke" | Out-Null

try {
    $runRoot = "data/local/backend-task-phase93"
    $statusPath = "$runRoot/backend-task-status.json"

    Invoke-TraceDeckLoggedCommand -Label "Phase 93 advisory helper contract" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-task-status-resilience.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 93 live backend task status advisory" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
            -Addr $Addr `
            -OutputPath $statusPath
    }

    $statusFullPath = Join-Path $script:TraceDeckRepoRoot $statusPath
    if (-not (Test-Path -LiteralPath $statusFullPath)) {
        throw "Expected Phase 93 status output: $statusFullPath"
    }
    $status = Get-Content -Path $statusFullPath -Raw | ConvertFrom-Json
    if (-not $status.advisory) {
        throw "Expected backend task status advisory object."
    }
    if ([string]::IsNullOrWhiteSpace($status.advisory.code)) {
        throw "Expected backend task advisory code."
    }
    if ([string]::IsNullOrWhiteSpace($status.advisory.headline)) {
        throw "Expected backend task advisory headline."
    }
    if ([string]::IsNullOrWhiteSpace($status.advisory.operator_action)) {
        throw "Expected backend task advisory operator action."
    }
    if (-not (Test-TraceDeckBackendTaskStatusAcceptable -Status $status)) {
        $reason = Get-TraceDeckBackendTaskStatusFailureReason -Status $status
        throw "Phase 93 live backend task status is not acceptable: $reason`: $($status | ConvertTo-Json -Depth 8)"
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 93 live provenance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl "http://$Addr"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 93 smoke passed addr=$Addr advisory=$($status.advisory.code)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
