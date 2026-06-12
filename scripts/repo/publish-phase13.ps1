param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/13-risky-software-detection"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase13" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 13 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase13.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 13 risky software detection in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 13 risky software detection" --body "Add privacy-safe risky software classification for process telemetry, alert evaluator support, dashboard watchlist, docs, Postman/Newman coverage, and local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 13 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "agent/internal/alert/evaluator.go",
        "agent/internal/alert/evaluator_test.go",
        "agent/internal/collector/process/collector.go",
        "agent/internal/constants/software.go",
        "agent/internal/software/classifier.go",
        "agent/internal/software/classifier_test.go",
        "backend/internal/api/web/dashboard.html",
        "docs/dashboard.md",
        "docs/policy-config.md",
        "docs/risky-software-detection.md",
        "docs/roadmap.md",
        "docs/telemetry-schema.md",
        "docs/testing.md",
        "examples/policies/ai-btech-student.yaml",
        "postman/tracedeck-backend-phase13.postman_collection.json",
        "scripts/local/newman-phase13.ps1",
        "scripts/local/smoke-phase13.ps1",
        "scripts/local/test-risky-software.ps1",
        "scripts/repo/publish-phase13.ps1",
        "scripts/verify/verify-phase13.ps1"
    )

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 13"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add risky software detection")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase13.ps1 passed
- scripts/local/test-risky-software.ps1 passed classifier and alert evaluator coverage
- scripts/local/smoke-phase13.ps1 live-booted backend and verified risky software dashboard visibility
- scripts/local/newman-phase13.ps1 passed authenticated dashboard/backend coverage against a live backend
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 13: risky software detection" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 13 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 13 PR" -Command {
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
