param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/43-buyer-ops-layout"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase43" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 43 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase43.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 43 buyer operations layout in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 43 buyer operations layout" --body "Add a monetisation-grade Buyer Operations Brief for anomaly alerting, mail proof, push notification dispatch, report readiness, archive retention, trust/audit, delivery command, package snapshot, and next commercial action, plus screenshot-free dashboard layout metrics across desktop/tablet/mobile with smoke/Newman/local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 43 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "backend/internal/api/web/dashboard.html",
        "docs/dashboard.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase43.postman_collection.json",
        "scripts/local/newman-phase43.ps1",
        "scripts/local/smoke-phase43.ps1",
        "scripts/local/test-dashboard-layout.ps1",
        "scripts/repo/publish-phase43.ps1",
        "scripts/setup/install-playwright-python.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase43.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: add buyer operations layout guard")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 43"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase43.ps1 passed
- scripts/local/smoke-phase43.ps1 live-booted the seeded dashboard, verified Buyer Operations Brief markers, and ran screenshot-free desktop/tablet/mobile layout metrics
- scripts/local/newman-phase43.ps1 passed against a live dashboard demo
- scripts/local/test-dashboard-layout.ps1 passed with no screenshots, video, credentials, or page content capture
- scripts/local/test-dashboard-contract.ps1 passed
- scripts/local/test-backend-api.ps1 passed
- go test ./agent/... passed
- scripts/local/test-dashboard-js.ps1 passed
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 43: buyer operations layout guard" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 43 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 43 PR" -Command {
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
