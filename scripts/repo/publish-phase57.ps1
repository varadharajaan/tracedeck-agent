Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase57" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/57-customer-control-room"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 57 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase57.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a monetisation-grade Customer Control Room that goes beyond host filtering by showing anomaly command, mail delivery, push notification evidence, provider proof, package billing, archive/report readiness, and owner actions in one first-screen dashboard.

## Verification
- scripts/verify/verify-phase57.ps1
- scripts/local/smoke-phase57.ps1
- scripts/local/newman-phase57.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint payloads, private content, or payment card data.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 57: customer control room" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 57 issue #${IssueNumber}: $issueURL"
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
        "postman/tracedeck-backend-phase57.postman_collection.json",
        "scripts/local/newman-phase57.ps1",
        "scripts/local/smoke-phase57.ps1",
        "scripts/repo/publish-phase57.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase57.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add customer control room")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 57"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed tenant Customer Control Room API for anomaly command, mail delivery, push notification evidence, provider proof, report/archive readiness, package billing, and owner actions
- add first-screen dashboard section and command-nav entry so the UI reads like a monetisable product, not only a host-level monitor
- add Phase 57 smoke, Newman, DOM, API, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase57.ps1 passed
- scripts/local/smoke-phase57.ps1 live-booted the seeded dashboard, verified customer-control UX markers, exercised the typed customer-control-room API, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase57.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only Customer Control Room; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, endpoint payloads, private content, or payment card data are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 57: customer control room" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 57 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 57 PR" -Command {
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
