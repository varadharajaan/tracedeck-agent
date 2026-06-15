param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/91-persistent-local-backend-task",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase91" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 91 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase91.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Keep the local TraceDeck backend available after the devctl command session exits, with truthful status reporting for Windows Task Scheduler readback.

## Scope
- add scheduled-task backed local backend start, stop, status, and restart scripts
- add `devctl.py server task-start`, `task-stop`, `task-restart`, and `task-status`
- keep task status honest by distinguishing `Running`, `missing`, and `inaccessible`
- add isolated smoke, Newman, verifier, docs, and devctl test targets
- keep generated logs and reports under `logs/` and `data/`

## Verification
- scripts/verify/verify-phase91.ps1
- scripts/local/smoke-phase91.ps1
- scripts/local/newman-phase91.ps1
- scripts/local/test-backend-dev-task.ps1
- postman/tracedeck-backend-phase91.postman_collection.json

## Privacy
Metadata and provenance checks only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 91: persistent local backend task controls" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 91 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/dashboard-demo-lifecycle.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase91.postman_collection.json",
        "scripts/local/get-backend-dev-task-status.ps1",
        "scripts/local/newman-phase91.ps1",
        "scripts/local/run-backend-dev-task.ps1",
        "scripts/local/smoke-phase91.ps1",
        "scripts/local/start-backend-dev-task.ps1",
        "scripts/local/stop-backend-dev-task.ps1",
        "scripts/local/test-backend-dev-task.ps1",
        "scripts/repo/publish-phase91.ps1",
        "scripts/verify/verify-phase91.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add persistent local backend task controls", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 91"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add Windows scheduled-task backed local backend controls for devctl
- keep status truthful when Scheduler readback is inaccessible but runtime health is still OK
- add Phase 91 smoke/Newman/verify scripts and docs
- keep local runtime output under data/local and logs/local

## Verification
- scripts/verify/verify-phase91.ps1
- scripts/local/smoke-phase91.ps1
- scripts/local/newman-phase91.ps1
- scripts/local/test-backend-dev-task.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Metadata and provenance checks only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 91: persistent local backend task controls" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 91 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 91 PR" -Command {
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
