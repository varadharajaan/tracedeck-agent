param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/93-backend-task-status-advisory",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase93" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 93 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase93.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make local backend task status actionable when Windows Scheduler metadata is verified, denied, missing, or failing.

## Scope
- add a typed advisory object to backend task status JSON
- explain whether the operator can continue, whether elevated Scheduler readback is recommended, and what command to run next
- expose the advisory through `python ./devctl.py server task-status`
- add Phase 93 smoke, Newman, verifier, docs, and devctl test targets
- keep generated logs and reports under `logs/` and `data/`

## Verification
- scripts/verify/verify-phase93.ps1
- scripts/local/test-backend-task-status-resilience.ps1
- scripts/local/smoke-phase93.ps1
- scripts/local/newman-phase93.ps1
- postman/tracedeck-backend-phase93.postman_collection.json

## Privacy
Launch/status metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 93: backend task status advisory" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 93 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/dashboard-demo-lifecycle.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase93.postman_collection.json",
        "scripts/lib/backend-task-status.ps1",
        "scripts/local/get-backend-dev-task-status.ps1",
        "scripts/local/newman-phase93.ps1",
        "scripts/local/smoke-phase93.ps1",
        "scripts/local/test-backend-task-status-resilience.ps1",
        "scripts/repo/publish-phase93.ps1",
        "scripts/verify/verify-phase93.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: add backend task status advisory", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 93"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed advisory metadata to backend task status output
- surface clear next actions in `devctl.py server task-status`
- keep Scheduler readback denial truthful while making local runtime proof understandable
- add Phase 93 smoke/Newman/verify scripts and docs

## Verification
- scripts/verify/verify-phase93.ps1
- scripts/local/test-backend-task-status-resilience.ps1
- scripts/local/smoke-phase93.ps1
- scripts/local/newman-phase93.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Launch/status metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 93: backend task status advisory" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 93 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 93 PR" -Command {
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
