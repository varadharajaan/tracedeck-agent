param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/107-contract-completion-audit",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase107" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 107 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase107.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 107 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 107 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a metadata-only completion audit that compares the current repository to the TraceDeck contract and explicitly lists implemented, partial, and missing end-to-end deliverables.

## Scope
- add `docs/contract-completion-audit.md`
- add `python ./devctl.py audit`
- write contract audit JSON/text under `data/local/output`
- add a Phase 107 verifier for the audit contract
- update README, phase ledger, roadmap, and testing docs

## Verification
- scripts/verify/verify-phase107.ps1
- scripts/local/test-contract-completion-audit.ps1
- scripts/local/get-contract-completion-audit.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Repository metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 107: contract completion audit" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 107 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/contract-completion-audit.md",
        "docs/phase-ledger.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "scripts/local/get-contract-completion-audit.ps1",
        "scripts/local/test-contract-completion-audit.ps1",
        "scripts/repo/publish-phase107.ps1",
        "scripts/verify/verify-phase107.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "docs: add contract completion audit", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 107"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a contract completion audit and direct devctl command
- write metadata-only audit JSON/text under `data/local/output`
- identify implemented, partial, and missing TraceDeck contract deliverables
- update docs and add a focused verifier

## Verification
- scripts/verify/verify-phase107.ps1
- scripts/local/test-contract-completion-audit.ps1
- scripts/local/get-contract-completion-audit.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Repository metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 107: contract completion audit" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 107 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 107 PR" -Command {
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
