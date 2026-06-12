param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/42-command-navigation"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase42" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 42 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase42.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 42 command navigation in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 42 command navigation" --body "Add a monetisation-grade command navigation strip with stable jump targets for paid ops, revenue, notifications, reports, archive, trust, and hosts, backed by typed dashboard data and covered by smoke/Newman/local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 42 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/web/dashboard.html",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase42.postman_collection.json",
        "scripts/local/newman-phase42.ps1",
        "scripts/local/smoke-phase42.ps1",
        "scripts/repo/publish-phase42.ps1",
        "scripts/verify/verify-phase42.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: add dashboard command navigation")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 42"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase42.ps1 passed
- scripts/local/smoke-phase42.ps1 live-booted the seeded dashboard and verified command navigation markers, jump targets, alert inbox KPI backing data, route KPI backing data, monetisation readiness, weekly report readiness, and root privacy-safe API continuity
- scripts/local/newman-phase42.ps1 passed against a live dashboard demo
- scripts/local/test-dashboard-contract.ps1 passed
- scripts/local/test-backend-api.ps1 passed
- go test ./agent/... passed
- scripts/local/test-dashboard-js.ps1 passed
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 42: dashboard command navigation" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 42 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 42 PR" -Command {
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
