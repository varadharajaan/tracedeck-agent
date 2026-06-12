param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/21-policy-assignment"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase21" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 21 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase21.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 21 device groups and policy assignments in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 21 device groups and policy assignments" --body "Add tenant-scoped device groups, policy assignment APIs, seeded rollout metadata, dashboard rollout panels, docs, Postman/Newman coverage, and local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 21 issue: $issueNumber"

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
        "backend/internal/store/memory_test.go",
        "backend/internal/store/repository.go",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/policy-assignment.md",
        "docs/roadmap.md",
        "docs/saas-readiness.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase21.postman_collection.json",
        "scripts/local/newman-phase21.ps1",
        "scripts/local/smoke-phase21.ps1",
        "scripts/repo/publish-phase21.ps1",
        "scripts/verify/verify-phase21.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: add policy assignment rollout")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 21"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase21.ps1 passed
- scripts/local/test-backend-api.ps1 passed device group, policy assignment, audit, and persistence coverage
- scripts/local/smoke-phase21.ps1 live-booted backend and verified seeded groups, custom group, active assignment, audit trail, and dashboard rollout panels
- scripts/local/newman-phase21.ps1 passed 9 requests and 9 assertions against a live API-key-protected backend
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 21: device groups and policy assignments" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 21 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 21 PR" -Command {
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
