Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase67" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/67-premium-operations-hub"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 67 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase67.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a Premium Operations Hub that makes TraceDeck's dashboard feel like a monetisable endpoint productivity and risk observability product, with anomaly inbox, mail delivery proof, push notification route state, dashboard fallback, weekly reports, archive retention, deployment readiness, package value, and owner actions in one first-screen surface.

## Verification
- scripts/verify/verify-phase67.ps1
- scripts/local/smoke-phase67.ps1
- scripts/local/newman-phase67.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, raw provider payloads, keylogging, or hidden collection bypasses.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 67: premium operations hub" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 67 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/repository.go",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase67.postman_collection.json",
        "scripts/local/newman-phase67.ps1",
        "scripts/local/smoke-phase67.ps1",
        "scripts/repo/publish-phase67.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase67.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add premium operations hub")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 67"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a typed Premium Operations Hub API that composes revenue operations, deployment readiness, and portfolio metadata
- render a monetisable dashboard first screen for anomaly inbox, mail delivery proof, push notification route state, dashboard fallback, weekly reports, archive retention, deployment readiness, package value, and owner actions
- add Phase 67 smoke, Newman, DOM/layout, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase67.ps1 passed
- scripts/local/smoke-phase67.ps1 live-booted the seeded dashboard, verified premium operations UX markers, checked typed premium operations data, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase67.ps1 passed against a live dashboard demo
- backend, agent, dashboard contract, dashboard JavaScript syntax, service manifest rendering, cross-platform Windows/macOS/Linux builds, and root artifact checks passed locally

## Privacy
- metadata-only premium operations surface; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, raw provider payloads, keylogging, or hidden collection bypasses are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 67: premium operations hub" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 67 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 67 PR" -Command {
        gh pr merge $prURL --squash --delete-branch
    }

    Invoke-Git -Args @("checkout", "main")
    Invoke-Git -Args @("pull", "--ff-only", "origin", "main")

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
