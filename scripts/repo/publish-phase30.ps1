param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/30-offline-backend-sync"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase30" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 30 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase30.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 30 offline backend sync in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 30 offline backend sync" --body "Add durable SQLite backend sync cursor/backlog replay, offline-tolerant agent sync behavior, idempotent backend telemetry ingest by stable event ID, docs, Postman/Newman coverage, live offline-to-online smoke, and full local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 30 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/constants/project.go",
        "agent/internal/storage/sqlite/migrations/002_backend_sync.sql",
        "agent/internal/storage/sqlite/store.go",
        "agent/internal/storage/sqlite/store_test.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/memory_test.go",
        "docs/agent-telemetry-ingest.md",
        "docs/backend-api.md",
        "docs/offline-backend-sync.md",
        "docs/policy-config.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase30.postman_collection.json",
        "scripts/local/newman-phase30.ps1",
        "scripts/local/smoke-phase30.ps1",
        "scripts/repo/publish-phase30.ps1",
        "scripts/verify/verify-phase30.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: add offline backend sync")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 30"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase30.ps1 passed
- scripts/local/smoke-phase30.ps1 proved an offline agent run stores telemetry locally and a later online run replays it to the backend
- scripts/local/newman-phase30.ps1 passed against a live dashboard and verified idempotent replay by stable event ID
- scripts/local/test-backend-api.ps1 passed
- go test ./agent/... passed
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 30: offline backend sync" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 30 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 30 PR" -Command {
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
