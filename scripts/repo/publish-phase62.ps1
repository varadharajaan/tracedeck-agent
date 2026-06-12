Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase62" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/62-monetisation-overview"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 62 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase62.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a buyer-grade Monetisation Overview as the dashboard opening surface. It should make TraceDeck feel like a paid endpoint productivity and risk observability product by showing account coverage, host coverage, anomaly pressure, push notification proof, mail delivery proof, report readiness, archive posture, package fit, owner actions, and trust guardrails before drilldowns.

## Verification
- scripts/verify/verify-phase62.ps1
- scripts/local/smoke-phase62.ps1
- scripts/local/newman-phase62.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 62: monetisation overview" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 62 issue #${IssueNumber}: $issueURL"
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
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase62.postman_collection.json",
        "scripts/local/newman-phase62.ps1",
        "scripts/local/smoke-phase62.ps1",
        "scripts/repo/publish-phase62.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase62.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add monetisation overview dashboard")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 62"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a top-of-dashboard Monetisation Overview that presents account, host, anomaly, mail, push, report, archive, trust, package, and owner-action proof before drilldowns
- wire the overview to existing typed metadata-only APIs, including account portfolio, portfolio center, customer control, weekly reports, and delivery evidence
- add Phase 62 smoke, Newman, DOM/layout, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase62.ps1 passed
- scripts/local/smoke-phase62.ps1 live-booted the seeded dashboard, verified monetisation overview UX markers, checked account proof data, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase62.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only dashboard overview; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, or raw provider payloads are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 62: monetisation overview dashboard" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 62 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 62 PR" -Command {
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
