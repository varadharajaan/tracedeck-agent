param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/17-provider-email-alerts"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase17" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 17 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase17.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 17 provider-backed email alerts in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 17 provider-backed email alerts" --body "Add SMTP and AWS SESv2 provider-backed alert delivery, required typed sender policy, env-only provider credentials, fake SMTP live smoke, docs, schema updates, and local verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 17 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "go.mod",
        "go.sum",
        "agent/internal/alert/email_notifier.go",
        "agent/internal/alert/email_notifier_test.go",
        "agent/internal/app/run.go",
        "agent/internal/config/types.go",
        "agent/internal/config/validate.go",
        "agent/internal/constants/config_fields.go",
        "agent/internal/constants/policy.go",
        "docs/alerting.md",
        "docs/policy-config.md",
        "docs/roadmap.md",
        "docs/schema/policy-v1alpha1.schema.json",
        "docs/security.md",
        "docs/testing.md",
        "examples/policies/ai-btech-student.yaml",
        "scripts/local/smoke-phase17.ps1",
        "scripts/local/test-email-notifier.ps1",
        "scripts/local/update-phase17-deps.ps1",
        "scripts/repo/publish-phase17.ps1",
        "scripts/tools/fake-smtp/main.go",
        "scripts/verify/verify-phase17.ps1"
    )

    $hasStagedChanges = $true
    git diff --cached --quiet
    if ($LASTEXITCODE -eq 0) {
        $hasStagedChanges = $false
    }
    if ($hasStagedChanges) {
        Invoke-Git -Args @("commit", "-m", "feat: add provider-backed email alerts")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 17"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase17.ps1 passed
- scripts/local/test-email-notifier.ps1 passed focused alert/app/config tests
- scripts/local/smoke-phase0.ps1 regenerated schema and passed agent bootstrap smoke
- scripts/local/smoke-phase17.ps1 live-tested SMTP delivery through a local fake SMTP capture
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 17: provider-backed email alerts" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 17 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 17 PR" -Command {
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
