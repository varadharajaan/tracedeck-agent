Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase58" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/58-customer-success-packet"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 58 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase58.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a monetisation-grade Customer Success Packet that turns TraceDeck evidence into a buyer/admin review surface with anomaly proof, mail delivery, push notification evidence, report/archive readiness, package fit, provider rehearsal, privacy assurances, objection handling, and owner actions.

## Verification
- scripts/verify/verify-phase58.ps1
- scripts/local/smoke-phase58.ps1
- scripts/local/newman-phase58.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint payloads, private content, push endpoints, invoices, or payment card data.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 58: customer success packet" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 58 issue #${IssueNumber}: $issueURL"
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
        "postman/tracedeck-backend-phase58.postman_collection.json",
        "scripts/local/newman-phase58.ps1",
        "scripts/local/smoke-phase58.ps1",
        "scripts/repo/publish-phase58.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase58.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add customer success packet")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 58"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed tenant Customer Success Packet API for anomaly proof, mail delivery, push notification evidence, report/archive readiness, package fit, provider rehearsal, privacy assurances, objections, and owner actions
- add top-dashboard Customer Success Packet section and command-nav entry so the UI reads like a buyer/admin product review, not only a monitoring console
- add Phase 58 smoke, Newman, DOM, API, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase58.ps1 passed
- scripts/local/smoke-phase58.ps1 live-booted the seeded dashboard, verified customer-success UX markers, exercised the typed customer-success-packet API, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase58.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only Customer Success Packet; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint payloads, private content, push endpoints, invoices, or payment card data are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 58: customer success packet" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 58 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 58 PR" -Command {
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
