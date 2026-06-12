Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase64" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/64-customer-settings-center"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 64 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase64.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a customer settings center that turns TraceDeck's plan, retention, notification, route, archive, autostart, role, and privacy proof into buyer-safe configurable settings for activation and renewal.

## Verification
- scripts/verify/verify-phase64.ps1
- scripts/local/smoke-phase64.ps1
- scripts/local/newman-phase64.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 64: customer settings center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 64 issue #${IssueNumber}: $issueURL"
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
        "postman/tracedeck-backend-phase64.postman_collection.json",
        "scripts/local/newman-phase64.ps1",
        "scripts/local/smoke-phase64.ps1",
        "scripts/repo/publish-phase64.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase64.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add customer settings center")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 64"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a typed customer settings center API for plan, retention, notification, mail, push, dashboard, archive, autostart, role, and data-rights settings
- render the settings center in the monetisation dashboard with settings matrix, plan/retention options, channel proof, and owner actions
- add Phase 64 smoke, Newman, DOM/layout, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase64.ps1 passed
- scripts/local/smoke-phase64.ps1 live-booted the seeded dashboard, verified customer settings UX markers, checked typed settings data, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase64.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only customer settings surface; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, or raw provider payloads are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 64: customer settings center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 64 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 64 PR" -Command {
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
