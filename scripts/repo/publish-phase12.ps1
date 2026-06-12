param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/12-device-health-dashboard"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase12" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 12 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase12.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 12 device health score and monetisation dashboard in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 12 device health score and monetisation dashboard" --body "Add privacy-safe device health scoring, typed backend health API, richer monetisation-ready dashboard panels, docs, Postman/Newman coverage, and live smoke verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 12 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/collector/health/collector.go",
        "agent/internal/collector/health/collector_test.go",
        "agent/internal/constants/events.go",
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
        "docs/device-health-score.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase12.postman_collection.json",
        "scripts/local/newman-phase12.ps1",
        "scripts/local/smoke-phase12.ps1",
        "scripts/local/test-device-health.ps1",
        "scripts/repo/publish-phase12.ps1",
        "scripts/verify/verify-phase12.ps1"
    )

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 12"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add device health dashboard")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase12.ps1 passed
- scripts/local/test-device-health.ps1 passed agent/backend health coverage
- scripts/local/smoke-phase12.ps1 live-booted the backend and verified health API plus dashboard panels
- scripts/local/newman-phase12.ps1 passed authenticated API coverage against a live backend
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 12: device health dashboard" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 12 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 12 PR" -Command {
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
