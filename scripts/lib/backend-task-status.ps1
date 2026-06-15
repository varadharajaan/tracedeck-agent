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
$script:TraceDeckTaskAdvisorySeverityOK = "ok"
$script:TraceDeckTaskAdvisorySeverityWatch = "watch"
$script:TraceDeckTaskAdvisorySeverityActionRequired = "action_required"
$script:TraceDeckTaskAdvisoryCodeReady = "scheduler_verified_runtime_ready"
$script:TraceDeckTaskAdvisoryCodeReadbackDenied = "runtime_ready_scheduler_readback_denied"
$script:TraceDeckTaskAdvisoryCodeTaskMissing = "runtime_ready_scheduler_task_missing"
$script:TraceDeckTaskAdvisoryCodeQueryError = "runtime_ready_scheduler_query_error"
$script:TraceDeckTaskAdvisoryCodeRuntimeUnhealthy = "runtime_unhealthy"

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

function New-TraceDeckBackendTaskAdvisory {
    param(
        [string]$Severity,
        [string]$Code,
        [string]$Headline,
        [string]$OperatorAction,
        [bool]$CanContinue,
        [bool]$AdminReadbackRecommended
    )

    return [pscustomobject]@{
        severity = $Severity
        code = $Code
        headline = $Headline
        operator_action = $OperatorAction
        can_continue = $CanContinue
        admin_readback_recommended = $AdminReadbackRecommended
    }
}

function Get-TraceDeckBackendTaskStatusAdvisory {
    param([pscustomobject]$Status)

    if ($Status.runtime_ok -ne $true) {
        return New-TraceDeckBackendTaskAdvisory `
            -Severity $script:TraceDeckTaskAdvisorySeverityActionRequired `
            -Code $script:TraceDeckTaskAdvisoryCodeRuntimeUnhealthy `
            -Headline "Backend runtime proof is unhealthy." `
            -OperatorAction "Run python ./devctl.py server task-restart, then rerun python ./devctl.py server task-status." `
            -CanContinue $false `
            -AdminReadbackRecommended $false
    }

    if ($Status.task_present -eq $true) {
        return New-TraceDeckBackendTaskAdvisory `
            -Severity $script:TraceDeckTaskAdvisorySeverityOK `
            -Code $script:TraceDeckTaskAdvisoryCodeReady `
            -Headline "Backend runtime and Scheduler readback are verified." `
            -OperatorAction "No action needed." `
            -CanContinue $true `
            -AdminReadbackRecommended $false
    }

    $taskState = [string]$Status.task_state
    if ($taskState -eq $script:TraceDeckTaskStateInaccessible) {
        return New-TraceDeckBackendTaskAdvisory `
            -Severity $script:TraceDeckTaskAdvisorySeverityWatch `
            -Code $script:TraceDeckTaskAdvisoryCodeReadbackDenied `
            -Headline "Backend is running, but this shell cannot read Scheduler metadata." `
            -OperatorAction "Use an elevated PowerShell session for full Scheduler readback; runtime proof is enough for local dashboard checks." `
            -CanContinue $true `
            -AdminReadbackRecommended $true
    }

    if ($taskState -eq $script:TraceDeckTaskStateMissing) {
        return New-TraceDeckBackendTaskAdvisory `
            -Severity $script:TraceDeckTaskAdvisorySeverityActionRequired `
            -Code $script:TraceDeckTaskAdvisoryCodeTaskMissing `
            -Headline "Backend is running, but the Scheduler task is missing." `
            -OperatorAction "Run python ./devctl.py server task-start before relying on reboot persistence." `
            -CanContinue $false `
            -AdminReadbackRecommended $false
    }

    return New-TraceDeckBackendTaskAdvisory `
        -Severity $script:TraceDeckTaskAdvisorySeverityActionRequired `
        -Code $script:TraceDeckTaskAdvisoryCodeQueryError `
        -Headline "Backend is running, but Scheduler readback failed with an unclassified error." `
        -OperatorAction "Inspect logs/local/backend and rerun python ./devctl.py server task-status from an elevated PowerShell session." `
        -CanContinue $false `
        -AdminReadbackRecommended $true
}
