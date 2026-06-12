Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase59" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/59-push-activation-center"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 59 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase59.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a monetisation-grade Push Activation Center that proves anomaly notifications can reach an owner with push route evidence, mail fallback, dashboard fallback, retry posture, provider-safe simulation status, preference coverage, escalation coverage, scenarios, and owner actions.

## Verification
- scripts/verify/verify-phase59.ps1
- scripts/local/smoke-phase59.ps1
- scripts/local/newman-phase59.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint payloads, private content, push endpoints, invoices, payment card data, push provider secrets, or raw push endpoints.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 59: push activation center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 59 issue #${IssueNumber}: $issueURL"
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
        "postman/tracedeck-backend-phase59.postman_collection.json",
        "scripts/local/newman-phase59.ps1",
        "scripts/local/smoke-phase59.ps1",
        "scripts/repo/publish-phase59.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase59.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add push activation center")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 59"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed tenant Push Activation Center API for push route proof, mail fallback, dashboard fallback, retry posture, notification preferences, escalation coverage, simulation readiness, anomaly scenarios, owner actions, and privacy guard
- add dashboard Push Activation Center section and command-nav entry so push notification monetisation is visible beside customer success and notification revenue surfaces
- add Phase 59 smoke, Newman, DOM, API, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase59.ps1 passed
- scripts/local/smoke-phase59.ps1 live-booted the seeded dashboard, verified push activation UX markers, exercised the typed push-activation-center API, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase59.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only Push Activation Center; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint payloads, private content, push endpoints, invoices, payment card data, raw push endpoints, or provider delivery secrets are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 59: push activation center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 59 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 59 PR" -Command {
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
