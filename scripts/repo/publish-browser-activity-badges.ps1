param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "hotfix/browser-activity-badge-wrap",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-browser-activity-badges" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Browser Activity badge hotfix verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-browser-activity-badges.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Fix Browser Activity status badges that can fold short labels like `attention` into multiple lines.

## Scope
- enforce final Browser Activity badge wrapping rules with enough selector specificity
- keep table Signal cells readable by separating the status chip from detail copy
- add screenshot-free rendered badge metrics across light/dark and desktop/mobile
- add local, smoke, verify, docs, and `devctl` coverage

## Verification
- scripts/verify/verify-browser-activity-badges.ps1
- scripts/local/test-browser-activity-badges.ps1 -BaseUrl http://127.0.0.1:18080
- scripts/local/test-dashboard-theme.ps1 -BaseUrl http://127.0.0.1:18080 -OutputRoot data/local/dashboard-theme/browser-activity-badges
- python ./devctl.py test browser-labels

## Privacy
Rendered badge metrics only. No screenshots, credentials, cookies, tokens, raw URLs, page titles, private content, browser history content, provider secrets, alert bodies, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Hotfix: Browser Activity badge wrapping" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Browser Activity badge issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/web/browser_activity.html",
        "devctl.py",
        "docs/testing.md",
        "scripts/local/smoke-browser-activity-badges.ps1",
        "scripts/local/test-browser-activity-badges.ps1",
        "scripts/repo/publish-browser-activity-badges.ps1",
        "scripts/tools/browser_activity_badge_check.py",
        "scripts/verify/verify-browser-activity-badges.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "fix: keep browser activity badges single-line", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Browser Activity badge hotfix"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- keep Browser Activity status badges such as `attention` single-line in hero and table layouts
- add a Signal cell class so the badge stays stable while detail text wraps normally
- add screenshot-free Playwright DOM metrics for desktop/mobile and light/dark badge layout
- expose the check through `python ./devctl.py test browser-labels`

## Verification
- scripts/verify/verify-browser-activity-badges.ps1
- scripts/local/test-browser-activity-badges.ps1 -BaseUrl http://127.0.0.1:18080
- scripts/local/test-dashboard-theme.ps1 -BaseUrl http://127.0.0.1:18080 -OutputRoot data/local/dashboard-theme/browser-activity-badges
- python ./devctl.py test browser-labels

## Privacy
Rendered badge metrics only. No screenshots, credentials, cookies, tokens, raw URLs, page titles, private content, browser history content, provider secrets, alert bodies, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Hotfix: Browser Activity badge wrapping" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Browser Activity badge PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Browser Activity badge PR" -Command {
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
