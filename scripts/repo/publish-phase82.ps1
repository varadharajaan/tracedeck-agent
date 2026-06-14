param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/82-modern-admin-ui-polish",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase82" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 82 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase82.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Revamp the local TraceDeck Console, Browser Activity drilldown, and Lambda Cloud Admin frontend so the UI reads like a modern monetisable admin product instead of an internal/debug prototype.

## Scope
- replace visible text-logo marks with a symbolic brand mark
- add a unified light/dark visual polish layer across local dashboard, browser drilldown, and Lambda frontend
- make command navigation, section badges, panels, KPI cards, and tables more readable and less tiny/debug-like
- harden static, Playwright, and Newman checks against stale `TD`, `Browser{}`, `Center{}`, bracket shortcut, and terse debug copy
- add Phase 82 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase82.ps1
- scripts/local/smoke-phase82.ps1
- scripts/local/newman-phase82.ps1
- scripts/local/test-dashboard-visual-quality.ps1 through smoke-phase82
- scripts/local/test-lambda-frontend-visual.ps1 through smoke-phase82

## Privacy
UI-only metadata rendering change. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 82: modern admin UI polish" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 82 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/web/browser_activity.html",
        "backend/internal/api/web/dashboard.html",
        "devctl.py",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase82.postman_collection.json",
        "sam-app/frontend_function/app.py",
        "scripts/local/newman-phase82.ps1",
        "scripts/local/smoke-phase82.ps1",
        "scripts/repo/publish-phase82.ps1",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase82.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: polish admin dashboard ui", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 82"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- replace the visible text brand mark with a symbolic product mark on local dashboard and Browser Activity
- apply a unified modern light/dark admin polish layer to local dashboard, Browser Activity, and Lambda Cloud Admin
- improve navigation, badges, panels, KPI cards, tables, and command tiles so they read like a paid product surface
- harden Go DOM, Playwright visual, Lambda visual, and Newman checks against stale debug-style UI copy
- add Phase 82 smoke/Newman/verify/publish scripts, Postman coverage, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase82.ps1
- scripts/local/smoke-phase82.ps1
- scripts/local/newman-phase82.ps1
- scripts/local/test-dashboard-visual-quality.ps1
- scripts/local/test-lambda-frontend-visual.ps1

## Privacy
UI-only metadata rendering change. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 82: modern admin UI polish" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 82 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 82 PR" -Command {
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
