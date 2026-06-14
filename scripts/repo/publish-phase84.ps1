param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/84-modern-admin-ui",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase84" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 84 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase84.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Replace the debug-looking dashboard, Browser Viewer, and Lambda Cloud Admin visual layer with a modern, customer-grade endpoint observability console.

## Scope
- add a Phase 84 product UI layer for the local dashboard
- align Browser Viewer with the same command-center visual system
- align Lambda Cloud Admin with the same light/dark theme, cards, source rail, and evidence-table styling
- keep server-connected indicators, dark theme toggles, host/browser filtering, cache metrics, and localhost fallback visible
- add Phase 84 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase84.ps1
- scripts/local/smoke-phase84.ps1
- scripts/local/newman-phase84.ps1
- scripts/local/test-dashboard-visual-quality.ps1 through verify-phase84
- scripts/local/test-dashboard-theme.ps1 through verify-phase84
- scripts/local/test-dashboard-layout.ps1 through verify-phase84
- scripts/local/test-lambda-frontend-visual.ps1 through verify-phase84

## Privacy
UI and rendered-layout verification only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 84: modern admin UI revamp" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 84 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/web/browser_activity.html",
        "backend/internal/api/web/dashboard.html",
        "devctl.py",
        "docs/cloud-frontend.md",
        "docs/dashboard.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase84.postman_collection.json",
        "sam-app/frontend_function/app.py",
        "scripts/local/newman-phase84.ps1",
        "scripts/local/smoke-phase84.ps1",
        "scripts/repo/publish-phase84.ps1",
        "scripts/verify/verify-phase84.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "style: modernize admin dashboard UI", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 84"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a customer-grade Phase 84 UI layer across the local dashboard, Browser Viewer, and Lambda Cloud Admin
- replace debug-like toolbar/shortcut styling with polished segmented navigation, larger chips, consistent cards, evidence tables, and dark theme treatment
- keep server connectivity, dark theme toggles, browser activity navigation, host filtering, cache metrics, and localhost fallback visible
- add Phase 84 smoke/Newman/verify/publish scripts, Postman coverage, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase84.ps1
- scripts/local/smoke-phase84.ps1
- scripts/local/newman-phase84.ps1
- scripts/local/test-dashboard-visual-quality.ps1
- scripts/local/test-dashboard-theme.ps1
- scripts/local/test-dashboard-layout.ps1
- scripts/local/test-lambda-frontend-visual.ps1

## Privacy
UI and rendered-layout verification only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 84: modern admin UI revamp" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 84 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 84 PR" -Command {
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
