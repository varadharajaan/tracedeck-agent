param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/16-monetisation-dashboard"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase16" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 16 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase16.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 16 monetisation dashboard in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 16 monetisation dashboard" --body "Upgrade the embedded host dashboard into a monetisation-ready operations surface with anomaly notification inbox, mail delivery center, push routing, alert route SLA visibility, paid packaging cues, docs, Postman/Newman coverage, and local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 16 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "backend/internal/api/web/dashboard.html",
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase16.postman_collection.json",
        "scripts/local/newman-phase16.ps1",
        "scripts/local/smoke-phase16.ps1",
        "scripts/repo/publish-phase16.ps1",
        "scripts/verify/verify-phase16.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: upgrade monetisation dashboard")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 16"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase16.ps1 passed
- scripts/local/test-backend-api.ps1 passed backend API coverage
- scripts/local/smoke-phase16.ps1 live-booted backend and verified dashboard, email, push, dashboard feed, weekly report, and host risk signals
- scripts/local/newman-phase16.ps1 passed authenticated API/dashboard coverage against a live backend
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 16: monetisation dashboard" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 16 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 16 PR" -Command {
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
