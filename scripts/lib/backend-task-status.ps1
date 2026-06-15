Set-StrictMode -Version Latest

$script:TraceDeckTaskStateMissing = "missing"
$script:TraceDeckTaskStateInaccessible = "inaccessible"
$script:TraceDeckTaskStateQueryError = "query_error"
$script:TraceDeckSchedulerReadbackVerified = "verified"
$script:TraceDeckSchedulerReadbackDenied = "denied"
$script:TraceDeckSchedulerReadbackMissing = "missing"
$script:TraceDeckSchedulerReadbackError = "error"
$script:TraceDeckRuntimeEvidencePidAndHealth = "pid_and_health"
$script:TraceDeckRuntimeEvidenceHealthOnly = "health_only"
$script:TraceDeckRuntimeEvidencePidOnly = "pid_only"
$script:TraceDeckRuntimeEvidenceNone = "none"

function Get-TraceDeckTaskStateFromQueryError {
    param([string]$Message)

    if ([string]::IsNullOrWhiteSpace($Message)) {
        return $script:TraceDeckTaskStateMissing
    }

    $normalized = $Message.ToLowerInvariant()
    if ($normalized -match "access\s+denied" -or $normalized -match "access\s+is\s+denied" -or $normalized -match "unauthorized") {
        return $script:TraceDeckTaskStateInaccessible
    }
    if ($normalized -match "cannot\s+find" -or $normalized -match "not\s+found" -or $normalized -match "does\s+not\s+exist" -or $normalized -match "no\s+msft_scheduledtask" -or $normalized -match "path\s+specified") {
        return $script:TraceDeckTaskStateMissing
    }
    return $script:TraceDeckTaskStateQueryError
}

function Get-TraceDeckSchedulerReadbackState {
    param(
        [bool]$TaskPresent,
        [string]$TaskState
    )

    if ($TaskPresent) {
        return $script:TraceDeckSchedulerReadbackVerified
    }
    if ($TaskState -eq $script:TraceDeckTaskStateInaccessible) {
        return $script:TraceDeckSchedulerReadbackDenied
    }
    if ($TaskState -eq $script:TraceDeckTaskStateMissing) {
        return $script:TraceDeckSchedulerReadbackMissing
    }
    return $script:TraceDeckSchedulerReadbackError
}

function Get-TraceDeckRuntimeEvidenceState {
    param(
        [bool]$HealthOK,
        [bool]$PidRunning
    )

    if ($HealthOK -and $PidRunning) {
        return $script:TraceDeckRuntimeEvidencePidAndHealth
    }
    if ($HealthOK) {
        return $script:TraceDeckRuntimeEvidenceHealthOnly
    }
    if ($PidRunning) {
        return $script:TraceDeckRuntimeEvidencePidOnly
    }
    return $script:TraceDeckRuntimeEvidenceNone
}

function Test-TraceDeckBackendTaskStatusAcceptable {
    param([pscustomobject]$Status)

    if ($Status.runtime_ok -ne $true) {
        return $false
    }
    if ($Status.task_present -eq $true) {
        return $true
    }
    return [string]$Status.task_state -eq $script:TraceDeckTaskStateInaccessible
}

function Get-TraceDeckBackendTaskStatusFailureReason {
    param([pscustomobject]$Status)

    if ($Status.runtime_ok -ne $true) {
        return "backend runtime proof is not healthy"
    }
    if ($Status.task_present -eq $true) {
        return ""
    }
    if ([string]$Status.task_state -eq $script:TraceDeckTaskStateMissing) {
        return "scheduler task is missing"
    }
    if ([string]$Status.task_state -eq $script:TraceDeckTaskStateQueryError) {
        return "scheduler query failed with an unclassified error"
    }
    return "scheduler state is not acceptable"
}
