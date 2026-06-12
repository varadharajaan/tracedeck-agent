param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/5-backend-dashboard"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase5" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 5 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase5.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 5 backend and dashboard foundation in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 5 backend and dashboard foundation" --body "Add a localhost Go backend foundation, device enrollment APIs, policy template catalog, archive status, embedded dashboard shell, Postman/Newman coverage, and Phase 5 verification scripts.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 5 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "Makefile",
        "README.md",
        "backend",
        "docs/architecture.md",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/security.md",
        "docs/testing.md",
        "postman",
        "scripts/local/format-go.ps1",
        "scripts/local/newman-phase5.ps1",
        "scripts/local/smoke-phase5.ps1",
        "scripts/local/start-backend-dev.ps1",
        "scripts/local/stop-backend-dev.ps1",
        "scripts/local/stop-stale-verifiers.ps1",
        "scripts/local/test-backend-api.ps1",
        "scripts/repo/publish-phase5.ps1",
        "scripts/verify/check-cross-platform-build.ps1",
        "scripts/verify/check-gofmt.ps1",
        "scripts/verify/verify-phase0.ps1",
        "scripts/verify/verify-phase5.ps1"
    )

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 5"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add backend dashboard foundation")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase5.ps1 passed
- scripts/local/test-backend-api.ps1 passed
- scripts/local/smoke-phase5.ps1 passed against localhost backend
- scripts/local/newman-phase5.ps1 passed with 9 requests and 9 assertions
- scripts/verify/check-root-clean.ps1 passed
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- golangci-lint passed with 0 issues
- govulncheck found 0 called vulnerabilities
- go test -race skipped because CGO_ENABLED=0 in this Windows shell

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 5: backend and dashboard foundation" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 5 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 5 PR" -Command {
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
