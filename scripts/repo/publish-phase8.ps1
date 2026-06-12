param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/8-windows-task-autostart"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase8" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 8 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase8.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 8 Windows scheduled task autostart in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 8 Windows scheduled task autostart" --body "Add Windows Task Scheduler XML, render/register/status scripts, docs, and verification for reboot persistence without console-window flicker.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 8 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @(
        "add", "--",
        "README.md",
        "deployments/service/windows/tracedeck-agent-task.xml.tmpl",
        "docs/platform-support.md",
        "docs/roadmap.md",
        "docs/security.md",
        "docs/testing.md",
        "docs/windows-autostart.md",
        "scripts/local/get-windows-task-status.ps1",
        "scripts/local/register-windows-task.ps1",
        "scripts/local/render-windows-task.ps1",
        "scripts/local/test-windows-task-template.ps1",
        "scripts/repo/publish-phase8.ps1",
        "scripts/verify/verify-phase8.ps1"
    )

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 8"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add windows task autostart")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase8.ps1 passed
- scripts/local/test-windows-task-template.ps1 passed
- scripts/local/render-windows-task.ps1 rendered Task Scheduler XML under data/local
- scripts/local/get-windows-task-status.ps1 supports status/query checks
- scripts/verify/check-cross-platform-build.ps1 passed for agent and backend on windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 8: Windows scheduled task autostart" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 8 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 8 PR" -Command {
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
