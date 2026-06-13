param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/73-source-provenance",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase73" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 73 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase73.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make Phase 73 provenance hardening repeatable through first-class local scripts.

## Scope
- verify typed provenance fields on risk, delivery, browser, and cloud S3 rows
- prove local dashboard/browser source badges and Lambda Source columns
- add Phase 73 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase73.ps1
- scripts/local/smoke-phase73.ps1
- scripts/local/newman-phase73.ps1
- python ./devctl.py test phase73

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 73: source provenance verification harness" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 73 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "devctl.py",
        "docs/backend-api.md",
        "docs/cloud-frontend.md",
        "docs/dashboard.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase73.postman_collection.json",
        "scripts/local/newman-phase73.ps1",
        "scripts/local/smoke-phase73.ps1",
        "scripts/local/test-live-server-provenance.ps1",
        "scripts/repo/publish-phase73.ps1",
        "scripts/verify/verify-phase73.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: add phase 73 provenance harness", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 73"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add Phase 73 local smoke/Newman/verify/publish scripts
- add Phase 73 Postman provenance collection
- wire `python ./devctl.py test phase73`
- document source provenance verification and live readback

## Verification
- scripts/verify/verify-phase73.ps1
- scripts/local/smoke-phase73.ps1
- scripts/local/newman-phase73.ps1
- python ./devctl.py test live
- python ./devctl.py cloud newman

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 73: source provenance verification harness" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 73 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 73 PR" -Command {
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
