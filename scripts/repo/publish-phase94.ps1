param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/94-deployment-service-advisory",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase94" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 94 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase94.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Surface deployment service advisories in the readiness API and dashboard without adding new collectors.

## Scope
- add typed deployment advisories for live boot, native autostart, background start, offline replay, archive backlog, and ready states
- render a Service Advisory panel in the Deployment Readiness Center
- add Phase 94 smoke, Newman, verifier, docs, and devctl test targets
- keep generated logs and reports under `logs/` and `data/`

## Verification
- scripts/verify/verify-phase94.ps1
- scripts/local/smoke-phase94.ps1
- scripts/local/newman-phase94.ps1
- postman/tracedeck-backend-phase94.postman_collection.json

## Privacy
Deployment metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 94: deployment service advisory center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 94 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/dashboard.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase94.postman_collection.json",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "scripts/local/newman-phase94.ps1",
        "scripts/local/smoke-phase94.ps1",
        "scripts/local/test-dashboard-js.ps1",
        "scripts/repo/publish-phase94.ps1",
        "scripts/verify/verify-phase94.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add deployment service advisories", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 94"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed deployment service advisories to the Deployment Readiness Center API
- render a Service Advisory panel in the dashboard rollout page
- add Phase 94 smoke/Newman/verify scripts and docs
- keep the change metadata-only and privacy-preserving

## Verification
- scripts/verify/verify-phase94.ps1
- scripts/local/smoke-phase94.ps1
- scripts/local/newman-phase94.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Deployment metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 94: deployment service advisory center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 94 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 94 PR" -Command {
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
