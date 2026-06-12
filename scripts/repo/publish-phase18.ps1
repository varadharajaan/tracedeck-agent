param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/18-product-dashboard"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase18" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param(
        [Alias("Args")]
        [string[]]$GitArgs
    )
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 18 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase18.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 18 product-grade dashboard command center in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 18 product-grade dashboard command center" --body "Upgrade the embedded dashboard from technical panels into a product-grade command center with priority action, notification promise, commercial readiness, trust coverage, executive briefing, notification action queue, docs, Postman/Newman coverage, and local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 18 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "agent/internal/alert/email_notifier.go",
        "agent/internal/alert/email_notifier_test.go",
        "agent/internal/app/run.go",
        "agent/internal/config/types.go",
        "agent/internal/config/validate.go",
        "agent/internal/constants/config_fields.go",
        "agent/internal/constants/policy.go",
        "backend/internal/api/web/dashboard.html",
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase18.postman_collection.json",
        "scripts/local/newman-phase18.ps1",
        "scripts/local/smoke-phase18.ps1",
        "scripts/repo/publish-phase18.ps1",
        "scripts/tools/fake-smtp/main.go",
        "scripts/verify/verify-phase18.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: upgrade product dashboard")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 18"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase18.ps1 passed
- scripts/local/test-backend-api.ps1 passed backend API coverage
- scripts/local/smoke-phase18.ps1 live-booted backend and verified product dashboard sections
- scripts/local/newman-phase18.ps1 passed 5 requests and 5 assertions against a live backend
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 18: product dashboard command center" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 18 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 18 PR" -Command {
        gh pr merge "$prURL" --repo "$Owner/$RepoName" --squash --delete-branch
    }

    Invoke-Git -Args @("checkout", "main")
    Invoke-Git -Args @("pull", "--ff-only", "origin", "main")

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
