param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/76-dashboard-ui-revamp",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase76" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 76 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase76.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Revamp the embedded dashboard and Browser Activity Viewer so the UI reads like a monetisable endpoint operations product instead of an internal debug surface.

## Scope
- replace pseudo-letter toolbar markers with clear product controls
- tighten light and dark theme palettes
- improve card hierarchy, focus panels, chips, tables, and responsive wrapping
- keep dashboard page navigation and server connectivity indicators polished
- harden the screenshot-free layout contract against horizontal overflow
- add Phase 76 smoke, Newman, verify, publish, Postman, and docs coverage

## Verification
- scripts/verify/verify-phase76.ps1
- scripts/local/smoke-phase76.ps1
- scripts/local/newman-phase76.ps1
- scripts/local/test-dashboard-layout.ps1 through smoke-phase76

## Privacy
Layout and page-contract checks only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 76: dashboard UI revamp" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 76 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/web/browser_activity.html",
        "backend/internal/api/web/dashboard.html",
        "devctl.py",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase76.postman_collection.json",
        "scripts/local/newman-phase76.ps1",
        "scripts/local/smoke-phase76.ps1",
        "scripts/repo/publish-phase76.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase76.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "style: revamp dashboard product UI", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 76"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- revamp the dashboard and Browser Activity Viewer visual system
- remove pseudo-letter toolbar markers and noisy micro-label treatment
- improve light/dark palettes, card hierarchy, chips, tables, and responsive containment
- add Phase 76 smoke/Newman/verify/publish scripts, Postman coverage, docs, and layout assertions

## Verification
- scripts/verify/verify-phase76.ps1
- scripts/local/smoke-phase76.ps1
- scripts/local/newman-phase76.ps1
- scripts/local/test-dashboard-layout.ps1 through smoke-phase76

## Privacy
Layout and page-contract checks only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 76: dashboard UI revamp" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 76 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 76 PR" -Command {
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
