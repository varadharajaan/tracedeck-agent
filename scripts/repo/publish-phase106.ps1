param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/106-phase-ledger",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase106" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 106 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase106.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 106 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 106 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a concise phase ledger so the operator can answer how many phases are completed and how many numbered phases remain without reading long tracker files.

## Scope
- add `docs/phase-ledger.md`
- add `python ./devctl.py ledger`
- add scripted JSON/text ledger output under `data/local/output`
- add a Phase 106 verifier for the ledger contract
- update README, roadmap, and testing docs

## Verification
- scripts/verify/verify-phase106.ps1
- scripts/local/test-phase-ledger.ps1
- scripts/local/get-phase-ledger.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Repository metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 106: phase ledger and remaining count" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 106 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/phase-ledger.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "scripts/local/get-phase-ledger.ps1",
        "scripts/local/test-phase-ledger.ps1",
        "scripts/repo/publish-phase106.ps1",
        "scripts/verify/verify-phase106.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "docs: add phase ledger", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 106"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a tracked phase ledger and direct remaining-phase count
- add `python ./devctl.py ledger`
- write ledger JSON/text under `data/local/output`
- update roadmap/testing docs and add a focused verifier

## Verification
- scripts/verify/verify-phase106.ps1
- scripts/local/test-phase-ledger.ps1
- scripts/local/get-phase-ledger.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Repository metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 106: phase ledger and remaining count" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 106 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 106 PR" -Command {
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
