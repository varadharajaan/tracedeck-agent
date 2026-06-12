param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/28-agent-telemetry-ingest"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase28" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 28 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase28.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 28 agent telemetry ingest in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 28 agent telemetry ingest" --body "Add metadata-only agent-to-backend telemetry ingest, backend status proof, dashboard live telemetry panels, schema/docs updates, Postman/Newman coverage, and real agent live smoke verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 28 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/config/config_test.go",
        "agent/internal/config/types.go",
        "agent/internal/config/validate.go",
        "agent/internal/constants/config_fields.go",
        "agent/internal/constants/project.go",
        "agent/internal/syncer/client.go",
        "agent/internal/syncer/client_test.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/memory_test.go",
        "backend/internal/store/repository.go",
        "docs/agent-telemetry-ingest.md",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/policy-config.md",
        "docs/roadmap.md",
        "docs/schema/policy-v1alpha1.schema.json",
        "examples/policies/ai-btech-student.yaml",
        "postman/tracedeck-backend-phase28.postman_collection.json",
        "scripts/local/newman-phase28.ps1",
        "scripts/local/smoke-phase28.ps1",
        "scripts/repo/publish-phase28.ps1",
        "scripts/verify/verify-phase28.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: add agent telemetry ingest")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 28"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase28.ps1 passed
- scripts/local/smoke-phase28.ps1 ran the real agent once with backend sync enabled and verified backend telemetry status
- scripts/local/newman-phase28.ps1 passed against a live dashboard with the Phase 28 collection
- scripts/local/test-backend-api.ps1 passed
- go test ./agent/... passed
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 28: agent telemetry ingest" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 28 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 28 PR" -Command {
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
