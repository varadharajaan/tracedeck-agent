param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/14-weekly-report-packaging"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase14" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 14 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase14.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 14 weekly report packaging in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 14 weekly report packaging" --body "Generate weekly report JSON from host overview, add email/PDF readiness metadata, add lightweight PDF endpoint, dashboard visibility, docs, Postman/Newman coverage, and local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 14 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/report.go",
        "backend/internal/store/report_test.go",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "docs/weekly-report.md",
        "postman/tracedeck-backend-phase14.postman_collection.json",
        "scripts/local/newman-phase14.ps1",
        "scripts/local/smoke-phase14.ps1",
        "scripts/repo/publish-phase14.ps1",
        "scripts/verify/verify-phase14.ps1"
    )

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 14"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add weekly report packaging")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase14.ps1 passed
- scripts/local/test-backend-api.ps1 passed report/API coverage
- scripts/local/smoke-phase14.ps1 live-booted backend and verified report JSON plus PDF endpoint
- scripts/local/newman-phase14.ps1 passed authenticated report/PDF coverage against a live backend
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 14: weekly report packaging" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 14 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 14 PR" -Command {
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
