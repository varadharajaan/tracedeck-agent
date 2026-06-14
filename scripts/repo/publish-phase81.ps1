param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/81-dashboard-navigator-clarity",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase81" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 81 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase81.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make the embedded TraceDeck Console Workspace Navigator read like a product navigation surface, not an internal shortcut grid.

## Scope
- replace terse navigator labels such as Deploy, Control, Setup, Paid Ops, and Assurance with full product labels
- separate each navigation tile into a stable product label and a live metadata row
- harden the screenshot-free dashboard visual-quality contract for product labels and metadata rows
- add Phase 81 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase81.ps1
- scripts/local/smoke-phase81.ps1
- scripts/local/newman-phase81.ps1
- scripts/local/test-dashboard-visual-quality.ps1 through smoke-phase81

## Privacy
UI-only metadata rendering change. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 81: dashboard navigator clarity" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 81 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/web/dashboard.html",
        "devctl.py",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase81.postman_collection.json",
        "scripts/local/newman-phase81.ps1",
        "scripts/local/smoke-phase81.ps1",
        "scripts/repo/publish-phase81.ps1",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase81.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: clarify dashboard navigator labels", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 81"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- replace terse Workspace Navigator labels with full product labels
- split command tiles into command-label and command-meta rows
- harden the dashboard visual-quality checker against stale shortcut labels
- add Phase 81 smoke/Newman/verify/publish scripts, Postman coverage, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase81.ps1
- scripts/local/smoke-phase81.ps1
- scripts/local/newman-phase81.ps1

## Privacy
UI-only metadata rendering change. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 81: dashboard navigator clarity" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 81 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 81 PR" -Command {
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
