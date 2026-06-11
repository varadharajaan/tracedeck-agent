param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/2-archive-alert-outbox"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase2" -LogRoot "logs/local/repo" | Out-Null

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
    Invoke-TraceDeckLoggedCommand -Label "Phase 2 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase2.ps1
    }

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 2 archive and alert outbox in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 2 archive and alert outbox" --body "Add local compressed archive outbox, S3 upload adapter, blocked-app alert evaluator, dry-run alert notification outbox, and live Phase 2 smoke verification.").Trim()
        $issueNumber = ($issueURL -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 2 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @("add", "-A")

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 2"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "feat: add archive and alert outbox")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $body = @"
Closes #$issueNumber.

Local verification:
- scripts/verify/verify-phase2.ps1 passed
- scripts/local/smoke-phase2.ps1 passed with live blocked-app alert dry-run
- scripts/verify/check-cross-platform-build.ps1 passed for windows/amd64, darwin/amd64, linux/amd64
- scripts/verify/check-root-clean.ps1 passed
- go test -race skipped because CGO_ENABLED=0 in this Windows shell

GitHub Actions intentionally not configured.
"@
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 2: archive and alert outbox" --body $body).Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 2 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 2 PR" -Command {
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
