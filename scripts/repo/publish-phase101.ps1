param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/101-postmerge-verifier-hardening",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase101" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 101 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase101.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 101 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 101 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Harden the reusable post-merge verifier so phase checks that refresh a persistent local backend do not hang inside the output-capturing logger.

## Scope
- run the current phase verification directly in `verify-postmerge.ps1`
- keep subsequent post-merge checks logged through the existing helper
- add Phase 101 verifier/publish scripts and devctl aliases
- document the regression and rerun path

## Verification
- scripts/verify/verify-phase101.ps1
- scripts/verify/verify-postmerge.ps1 -PhaseTarget phase100 -IssueNumber 207 -PrNumber 208 -AllowContentDiff
- scripts/verify/check-root-clean.ps1

## Privacy
Verification metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 101: harden postmerge verifier" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 101 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/testing.md",
        "scripts/verify/verify-postmerge.ps1",
        "scripts/verify/verify-phase101.ps1",
        "scripts/repo/publish-phase101.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: harden postmerge verifier", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 101"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- run current phase verification directly from `verify-postmerge.ps1`
- keep backend task status, doctor, live provenance, root-clean, diff, and GitHub checks logged
- add Phase 101 verify/publish scripts and devctl/docs wiring

## Verification
- scripts/verify/verify-phase101.ps1
- scripts/verify/verify-postmerge.ps1 -PhaseTarget phase100 -IssueNumber 207 -PrNumber 208 -AllowContentDiff
- scripts/verify/check-root-clean.ps1

## Privacy
Verification metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 101: harden postmerge verifier" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 101 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 101 PR" -Command {
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
