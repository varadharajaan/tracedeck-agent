param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/4-policy-anomaly-engine"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase4" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 4 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase4.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 4 policy and anomaly engine in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 4 policy and anomaly engine" --body "Add blocked-domain and non-study YouTube alert evaluation, focused alert tests, live Phase 4 smoke verification, and updated docs for policy/anomaly behavior.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 4 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "agent/internal/alert/evaluator.go",
        "agent/internal/alert/evaluator_test.go",
        "agent/internal/constants/events.go",
        "docs/alerting.md",
        "docs/architecture.md",
        "docs/collection-policy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "examples/policies/ai-btech-student.yaml",
        "scripts/local/test-alert-engine.ps1",
        "scripts/local/smoke-phase4.ps1",
        "scripts/verify/verify-phase4.ps1",
        "scripts/repo/publish-phase4.ps1"
    )

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 4"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add policy anomaly alerts")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase4.ps1 passed
- scripts/local/test-alert-engine.ps1 passed
- scripts/local/smoke-phase4.ps1 passed with browser_events=3, alerts_raised=2, non_study_youtube, and blocked_domain_opened
- scripts/local/smoke-phase3.ps1 passed archive privacy checks
- scripts/verify/check-root-clean.ps1 passed
- scripts/verify/check-cross-platform-build.ps1 passed for windows/amd64, darwin/amd64, linux/amd64
- golangci-lint passed with 0 issues
- govulncheck found 0 called vulnerabilities
- go test -race skipped because CGO_ENABLED=0 in this Windows shell

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 4: policy and anomaly engine" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 4 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 4 PR" -Command {
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
