param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
. (Join-Path $PSScriptRoot "..\lib\backend-task-status.ps1")
Initialize-TraceDeckScriptLog -Name "test-backend-task-status-resilience" -LogRoot "logs/local/test" | Out-Null

try {
    $classificationCases = @(
        @{ Name = "empty means missing"; Message = ""; Expected = $script:TraceDeckTaskStateMissing },
        @{ Name = "access denied means inaccessible"; Message = "Access denied "; Expected = $script:TraceDeckTaskStateInaccessible },
        @{ Name = "access is denied means inaccessible"; Message = "Access is denied."; Expected = $script:TraceDeckTaskStateInaccessible },
        @{ Name = "missing task means missing"; Message = "No MSFT_ScheduledTask objects found"; Expected = $script:TraceDeckTaskStateMissing },
        @{ Name = "missing path means missing"; Message = "ERROR: The system cannot find the path specified."; Expected = $script:TraceDeckTaskStateMissing },
        @{ Name = "rpc outage means query error"; Message = "The RPC server is unavailable."; Expected = $script:TraceDeckTaskStateQueryError }
    )

    foreach ($case in $classificationCases) {
        $actual = Get-TraceDeckTaskStateFromQueryError -Message $case.Message
        if ($actual -ne $case.Expected) {
            throw "Task status classification failed for '$($case.Name)': expected=$($case.Expected) actual=$actual"
        }
    }

    $statusCases = @(
        @{
            Name = "verified task and runtime healthy"
            Status = [pscustomobject]@{ task_present = $true; task_state = "Ready"; runtime_ok = $true }
            Expected = $true
        },
        @{
            Name = "scheduler denied but runtime healthy"
            Status = [pscustomobject]@{ task_present = $false; task_state = $script:TraceDeckTaskStateInaccessible; runtime_ok = $true }
            Expected = $true
        },
        @{
            Name = "missing task is not acceptable"
            Status = [pscustomobject]@{ task_present = $false; task_state = $script:TraceDeckTaskStateMissing; runtime_ok = $true }
            Expected = $false
        },
        @{
            Name = "query error is not acceptable"
            Status = [pscustomobject]@{ task_present = $false; task_state = $script:TraceDeckTaskStateQueryError; runtime_ok = $true }
            Expected = $false
        },
        @{
            Name = "dead runtime is not acceptable"
            Status = [pscustomobject]@{ task_present = $true; task_state = "Ready"; runtime_ok = $false }
            Expected = $false
        }
    )

    foreach ($case in $statusCases) {
        $actual = Test-TraceDeckBackendTaskStatusAcceptable -Status $case.Status
        if ($actual -ne $case.Expected) {
            throw "Backend task status acceptance failed for '$($case.Name)': expected=$($case.Expected) actual=$actual"
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Backend task status resilience checks passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
