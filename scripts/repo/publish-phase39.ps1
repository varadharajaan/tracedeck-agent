param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/39-autostart-assurance"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase39" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 39 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase39.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 39 autostart assurance in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 39 autostart assurance" --body "Harden scripted service/autostart assurance for Windows Task Scheduler hidden startup, typed status JSON, service dry-run plans, docs, smoke/Newman coverage, and full local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 39 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "docs/roadmap.md",
        "docs/service-management.md",
        "docs/testing.md",
        "docs/windows-autostart.md",
        "postman/tracedeck-backend-phase39.postman_collection.json",
        "scripts/local/get-windows-task-status.ps1",
        "scripts/local/newman-phase39.ps1",
        "scripts/local/register-windows-task.ps1",
        "scripts/local/smoke-phase39.ps1",
        "scripts/local/test-autostart-assurance.ps1",
        "scripts/repo/publish-phase39.ps1",
        "scripts/verify/verify-phase39.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "test: add autostart assurance")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 39"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase39.ps1 passed
- scripts/local/test-autostart-assurance.ps1 passed and verified hidden Windows Task Scheduler startup, logon delay, continuous agent mode, restart-on-failure, typed missing-task status JSON, and Windows service-manager dry-run plans
- scripts/local/test-service-manager.ps1 passed for Windows, macOS, and Linux dry-run lifecycle plans
- scripts/local/test-windows-task-template.ps1 passed
- scripts/local/smoke-phase39.ps1 live-booted the seeded dashboard after autostart assurance and verified service trust markers
- scripts/local/newman-phase39.ps1 passed against a live dashboard demo
- scripts/local/test-backend-api.ps1 passed
- go test ./agent/... passed
- scripts/local/test-dashboard-js.ps1 passed
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 39: autostart assurance" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 39 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 39 PR" -Command {
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
