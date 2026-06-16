param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/102-runtime-pid-reconciliation",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase102" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 102 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase102.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 102 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 102 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make stale ready-file PID evidence explicit without marking a healthy live backend as down.

## Scope
- add typed ready PID status fields to backend task status and runtime summary
- add PID reconciliation proof/action rows to Runtime Status Center
- surface stale ready PID as watch in Operator Assurance while keeping live pid_and_health usable
- add Phase 102 smoke/Newman/verify/publish scripts and Postman coverage
- update docs and devctl aliases

## Verification
- scripts/verify/verify-phase102.ps1
- scripts/local/smoke-phase102.ps1
- scripts/local/newman-phase102.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational runtime metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 102: runtime PID reconciliation" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 102 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/testing.md",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "postman/tracedeck-backend-phase102.postman_collection.json",
        "scripts/lib/backend-task-status.ps1",
        "scripts/local/get-backend-dev-task-status.ps1",
        "scripts/local/get-runtime-summary.ps1",
        "scripts/local/test-backend-task-status-resilience.ps1",
        "scripts/local/test-runtime-summary.ps1",
        "scripts/local/clean-root-generated.ps1",
        "scripts/local/smoke-phase102.ps1",
        "scripts/local/newman-phase102.ps1",
        "scripts/verify/verify-phase102.ps1",
        "scripts/repo/publish-phase102.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add runtime PID reconciliation proof", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 102"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- reconcile live backend PID with ready-file PID in task status and runtime summary artifacts
- expose PID reconciliation proof and refresh action from Runtime Status Center
- show stale ready PID as watch in Operator Assurance while preserving healthy live runtime proof
- add Phase 102 smoke/Newman/verify scripts, Postman coverage, docs, and devctl aliases

## Verification
- scripts/verify/verify-phase102.ps1
- scripts/local/smoke-phase102.ps1
- scripts/local/newman-phase102.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational runtime metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 102: runtime PID reconciliation" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 102 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 102 PR" -Command {
        gh pr merge $prURL --merge --delete-branch
    }

    Invoke-Git -Args @("checkout", "main")
    Invoke-Git -Args @("pull", "--ff-only", "origin", "main")

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
