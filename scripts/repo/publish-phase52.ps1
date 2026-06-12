Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase52" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/52-role-experiences"

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 52 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase52.ps1
    }

    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 52 role experiences in:title" --json number --jq ".[0].number" 2>$null
    if ([string]::IsNullOrWhiteSpace($existingIssue)) {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 52 role experiences" --body "Add a typed tenant role experience API and monetisation-grade dashboard center for parent, student, school admin, and business manager onboarding, role readiness, notification proof, trust, archive/report value, privacy-safe smoke/Newman coverage, docs, and full local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    else {
        $issueNumber = $existingIssue.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 52 issue: $issueNumber"

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
        "backend/internal/store/memory_test.go",
        "backend/internal/store/repository.go",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase52.postman_collection.json",
        "scripts/local/newman-phase52.ps1",
        "scripts/local/smoke-phase52.ps1",
        "scripts/repo/publish-phase52.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase52.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add role experiences")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 52"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$issueNumber

## Summary
- add tenant role experience API for parent, student, school admin, and business manager onboarding readiness
- add dashboard Role Experience Center and Paid Onboarding Checklist for role-specific monetisation packaging
- add Phase 52 smoke, Newman, layout, API, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase52.ps1 passed
- scripts/local/smoke-phase52.ps1 live-booted the seeded dashboard, verified role experience UX markers, exercised role-experiences and delivery-timeline APIs, checked privacy/forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase52.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only role experience center; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, tokens, cookies, private content, or endpoint payloads are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 52: role experiences" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 52 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 52 PR" -Command {
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
