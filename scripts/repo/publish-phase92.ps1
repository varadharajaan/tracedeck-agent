param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/92-task-status-resilience",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase92" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 92 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase92.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make persistent local backend task status resilient when Windows allows runtime launch but denies Scheduler metadata readback from the current shell.

## Scope
- classify Scheduler query results as verified, inaccessible, missing, or query error
- accept `runtime_ok=true` plus `task_state=inaccessible` as healthy runtime proof instead of a false missing-task failure
- add a focused helper test for status classification and acceptance behavior
- add Phase 92 smoke, Newman, verifier, docs, and devctl test targets
- keep generated logs and reports under `logs/` and `data/`

## Verification
- scripts/verify/verify-phase92.ps1
- scripts/local/test-backend-task-status-resilience.ps1
- scripts/local/smoke-phase92.ps1
- scripts/local/newman-phase92.ps1
- postman/tracedeck-backend-phase92.postman_collection.json

## Privacy
Launch/status metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 92: backend task status resilience" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 92 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/dashboard-demo-lifecycle.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase92.postman_collection.json",
        "scripts/lib/backend-task-status.ps1",
        "scripts/local/get-backend-dev-task-status.ps1",
        "scripts/local/newman-phase92.ps1",
        "scripts/local/smoke-phase92.ps1",
        "scripts/local/test-backend-dev-task.ps1",
        "scripts/local/test-backend-task-status-resilience.ps1",
        "scripts/repo/publish-phase92.ps1",
        "scripts/verify/verify-phase92.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: harden backend task status resilience", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 92"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- classify Scheduler readback as verified, denied, missing, or query error
- keep `runtime_ok=true` as valid live proof when Windows denies Scheduler metadata readback
- add focused status helper tests plus Phase 92 smoke/Newman/verify scripts
- keep demo/live provenance guards in the Newman contract

## Verification
- scripts/verify/verify-phase92.ps1
- scripts/local/test-backend-task-status-resilience.ps1
- scripts/local/smoke-phase92.ps1
- scripts/local/newman-phase92.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Launch/status metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 92: backend task status resilience" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 92 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 92 PR" -Command {
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
