param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/95-repo-local-go-cache",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase95" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 95 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase95.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Keep scripted Go build/test artifacts inside the repo-local generated artifact area so nested verification does not depend on Windows profile cache ACLs.

## Scope
- set `GOCACHE` and `GOTMPDIR` from the shared PowerShell logging bootstrap
- document repo-local Go cache behavior
- add a Phase 95 verifier and devctl alias
- keep generated artifacts under `data/local` and logs under `logs/local`

## Verification
- scripts/verify/verify-phase95.ps1
- scripts/verify/verify-phase94.ps1
- scripts/local/test-backend-api.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Test infrastructure only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 95: repo-local Go cache for scripted verification" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 95 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/testing.md",
        "scripts/lib/logging.ps1",
        "scripts/repo/publish-phase95.ps1",
        "scripts/verify/verify-phase95.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: keep go cache repo-local", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 95"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- route scripted Go cache/temp artifacts to `data/local`
- add Phase 95 verification and devctl test target
- document the repo-local cache contract

## Verification
- scripts/verify/verify-phase95.ps1
- scripts/verify/verify-phase94.ps1
- scripts/local/test-backend-api.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Test infrastructure only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 95: repo-local Go cache for scripted verification" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 95 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 95 PR" -Command {
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
