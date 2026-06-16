param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$OutputRoot = "data/local/output",
    [string]$TaskName = "\TraceDeck\TraceDeck Backend Dev",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$ReadyPath = "data/local/backend/backend-task-ready.json",
    [string]$TaskStatusOutputPath = "data/local/backend/backend-task-status.json",
    [switch]$SkipDoctor
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-runtime-summary" -LogRoot "logs/local/ops" | Out-Null

function Get-TraceDeckProp {
    param(
        [object]$Object,
        [string]$Name,
        [object]$Default = $null
    )

    if ($null -eq $Object) {
        return $Default
    }
    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $Default
    }
    return $property.Value
}

function Read-TraceDeckJsonFile {
    param([string]$Path)

    if (!(Test-Path $Path)) {
        throw "Expected JSON file was not created: $Path"
    }
    return Get-Content -Path $Path -Raw | ConvertFrom-Json
}

try {
    $outputDir = Join-Path $script:TraceDeckRepoRoot $OutputRoot
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

    $summaryJsonPath = Join-Path $outputDir "runtime-summary.json"
    $summaryTextPath = Join-Path $outputDir "runtime-summary.txt"
    $taskStatusPath = Join-Path $script:TraceDeckRepoRoot $TaskStatusOutputPath
    $doctorJsonPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/runtime-doctor.json"
    $frontendUrlPath = Join-Path $script:TraceDeckRepoRoot "data/local/output/frontend-url.txt"

    Invoke-TraceDeckLoggedCommand -Label "Backend task status for runtime summary" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-backend-dev-task-status.ps1 `
            -Addr $Addr `
            -TaskName $TaskName `
            -PidPath $PidPath `
            -ReadyPath $ReadyPath `
            -OutputPath $TaskStatusOutputPath
    }
    $taskStatus = Read-TraceDeckJsonFile -Path $taskStatusPath
    $advisory = Get-TraceDeckProp -Object $taskStatus -Name "advisory" -Default $null
    $ready = Get-TraceDeckProp -Object $taskStatus -Name "ready" -Default $null

    $doctor = $null
    if (-not $SkipDoctor) {
        Invoke-TraceDeckLoggedCommand -Label "Runtime doctor for runtime summary" -Command {
            python ./devctl.py --addr $Addr doctor --skip-cloud
        }
        $doctor = Read-TraceDeckJsonFile -Path $doctorJsonPath
    }

    $frontendUrl = ""
    if (Test-Path $frontendUrlPath) {
        $frontendUrl = (Get-Content -Path $frontendUrlPath -Raw).Trim()
    }

    $trackedDiffRows = @(git -c core.excludesFile= diff --name-status 2>$null)
    $statusRows = @(git -c core.excludesFile= status --short --branch 2>$null)
    $gitBranch = (git -c core.excludesFile= branch --show-current 2>$null).Trim()
    $gitHead = (git -c core.excludesFile= rev-parse --short HEAD 2>$null).Trim()

    $backendRuntimeOk = [bool](Get-TraceDeckProp -Object $taskStatus -Name "runtime_ok" -Default $false)
    $backendHealthOk = [bool](Get-TraceDeckProp -Object $taskStatus -Name "health_ok" -Default $false)
    $doctorOverall = if ($SkipDoctor) { "skipped" } else { [string](Get-TraceDeckProp -Object $doctor -Name "overall" -Default "unknown") }
    $doctorLocal = if ($SkipDoctor) { "skipped" } else { [string](Get-TraceDeckProp -Object (Get-TraceDeckProp -Object $doctor -Name "local" -Default $null) -Name "overall" -Default "unknown") }
    $canContinue = $backendRuntimeOk -and $backendHealthOk -and ($doctorOverall -in @("ok", "skipped"))

    $nextActions = New-Object System.Collections.Generic.List[string]
    if (-not $backendRuntimeOk -or -not $backendHealthOk) {
        $nextActions.Add("Restart the backend with python ./devctl.py server task-restart, then rerun python ./devctl.py summary.")
    }
    $schedulerReadback = [string](Get-TraceDeckProp -Object $taskStatus -Name "scheduler_readback" -Default "unknown")
    if ($schedulerReadback -eq "denied") {
        $nextActions.Add("Use an elevated PowerShell session for full Scheduler readback if launch-task proof is required.")
    }
    if ($doctorOverall -notin @("ok", "skipped")) {
        $nextActions.Add("Inspect data/local/output/runtime-doctor.json and logs/local/devctl before promoting.")
    }
    if ($trackedDiffRows.Count -gt 0) {
        $nextActions.Add("Tracked content diff is present; commit/merge or inspect before final post-merge verification.")
    }
    if ($nextActions.Count -eq 0) {
        $nextActions.Add("No action needed.")
    }

    $summary = [ordered]@{
        generated_at = (Get-Date).ToString("o")
        base_url = "http://$Addr"
        output_json = $summaryJsonPath
        output_text = $summaryTextPath
        backend = [ordered]@{
            task_name = [string](Get-TraceDeckProp -Object $taskStatus -Name "task_name" -Default "")
            task_present = [bool](Get-TraceDeckProp -Object $taskStatus -Name "task_present" -Default $false)
            task_state = [string](Get-TraceDeckProp -Object $taskStatus -Name "task_state" -Default "unknown")
            scheduler_readback = $schedulerReadback
            launch_task_verified = [bool](Get-TraceDeckProp -Object $taskStatus -Name "launch_task_verified" -Default $false)
            runtime_ok = $backendRuntimeOk
            runtime_evidence = [string](Get-TraceDeckProp -Object $taskStatus -Name "runtime_evidence" -Default "")
            health_ok = $backendHealthOk
            pid = Get-TraceDeckProp -Object $taskStatus -Name "pid" -Default $null
            pid_running = [bool](Get-TraceDeckProp -Object $taskStatus -Name "pid_running" -Default $false)
            ready_file_present = [bool](Get-TraceDeckProp -Object $taskStatus -Name "ready_file_present" -Default $false)
            ready_at = [string](Get-TraceDeckProp -Object $ready -Name "ready_at" -Default "")
            advisory = [ordered]@{
                severity = [string](Get-TraceDeckProp -Object $advisory -Name "severity" -Default "unknown")
                code = [string](Get-TraceDeckProp -Object $advisory -Name "code" -Default "unknown")
                headline = [string](Get-TraceDeckProp -Object $advisory -Name "headline" -Default "")
                operator_action = [string](Get-TraceDeckProp -Object $advisory -Name "operator_action" -Default "")
                can_continue = [bool](Get-TraceDeckProp -Object $advisory -Name "can_continue" -Default $false)
            }
        }
        doctor = [ordered]@{
            skipped = [bool]$SkipDoctor
            overall = $doctorOverall
            local = $doctorLocal
            report_json = if ($SkipDoctor) { "" } else { $doctorJsonPath }
        }
        frontend = [ordered]@{
            url_present = -not [string]::IsNullOrWhiteSpace($frontendUrl)
            url = $frontendUrl
        }
        git = [ordered]@{
            branch = $gitBranch
            head = $gitHead
            tracked_content_diff = $trackedDiffRows.Count -gt 0
            tracked_content_diff_count = $trackedDiffRows.Count
            tracked_content_diff_rows = $trackedDiffRows
            status_rows = $statusRows
        }
        logs = [ordered]@{
            summary_log = $script:TraceDeckLogFile
            backend_stdout = [string](Get-TraceDeckProp -Object $ready -Name "stdout" -Default "")
            backend_stderr = [string](Get-TraceDeckProp -Object $ready -Name "stderr" -Default "")
        }
        verdict = [ordered]@{
            can_continue = $canContinue
            severity = if ($canContinue) { "ok" } else { "action_required" }
            headline = if ($canContinue) { "Runtime proof is healthy." } else { "Runtime proof needs attention." }
            next_actions = @($nextActions)
        }
        privacy = [ordered]@{
            metadata_only = $true
            sensitive_collection = "denied"
        }
    }

    $summaryJson = $summary | ConvertTo-Json -Depth 10
    Set-Content -Path $summaryJsonPath -Value $summaryJson -Encoding UTF8

    $textLines = @(
        "TraceDeck runtime summary",
        "Generated: $($summary.generated_at)",
        "Base URL: $($summary.base_url)",
        "Backend: runtime_ok=$($summary.backend.runtime_ok) health_ok=$($summary.backend.health_ok) scheduler=$($summary.backend.scheduler_readback) task_state=$($summary.backend.task_state)",
        "Advisory: $($summary.backend.advisory.severity)/$($summary.backend.advisory.code) - $($summary.backend.advisory.headline)",
        "Doctor: overall=$($summary.doctor.overall) local=$($summary.doctor.local)",
        "Git: branch=$($summary.git.branch) head=$($summary.git.head) tracked_diff_count=$($summary.git.tracked_content_diff_count)",
        "Frontend URL present: $($summary.frontend.url_present)",
        "Verdict: $($summary.verdict.severity) - $($summary.verdict.headline)"
    )
    foreach ($action in $summary.verdict.next_actions) {
        $textLines += "Action: $action"
    }
    Set-Content -Path $summaryTextPath -Value $textLines -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Runtime summary saved json=$summaryJsonPath text=$summaryTextPath can_continue=$canContinue"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
