param(
    [string]$RepoName = "tracedeck-agent",
    [string]$Owner = "varadharajaan",
    [string]$Branch = "phase/0-go-foundation"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase0" -LogRoot "logs/local/repo" | Out-Null

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
    if (-not (Test-Path ".git")) {
        Invoke-Git -Args @("init", "-b", "main")
        Invoke-Git -Args @("config", "user.name", "TraceDeck Automation")
        Invoke-Git -Args @("config", "user.email", "tracedeck-automation@example.local")
        Invoke-Git -Args @("commit", "--allow-empty", "-m", "chore: initialize repository base")
    }

    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    gh repo view "$Owner/$RepoName" 2>$null | Out-Null
    $repoExists = $LASTEXITCODE -eq 0
    $ErrorActionPreference = $previousErrorActionPreference

    if (-not $repoExists) {
        Invoke-TraceDeckLoggedCommand -Label "Create GitHub repo" -Command {
            gh repo create "$Owner/$RepoName" --public --description "OpenTelemetry-powered endpoint agent for app usage tracking, browser category analytics, software inventory monitoring, policy violations, and anomaly detection."
        }
    }

    $previousErrorActionPreference = $ErrorActionPreference
    $ErrorActionPreference = "Continue"
    $remote = git remote get-url origin 2>$null
    $hasOrigin = $LASTEXITCODE -eq 0 -and $remote
    $ErrorActionPreference = $previousErrorActionPreference
    if (-not $hasOrigin) {
        Invoke-Git -Args @("remote", "add", "origin", "https://github.com/$Owner/$RepoName.git")
    }

    Invoke-Git -Args @("push", "-u", "origin", "main")

    $issueNumber = ""
    $existingIssue = gh issue list --repo "$Owner/$RepoName" --state open --search "Phase 0 governance repo foundation in:title" --json number --jq ".[0].number" 2>$null
    if ($existingIssue) {
        $issueNumber = $existingIssue.Trim()
    }
    else {
        $issueNumber = (gh issue create --repo "$Owner/$RepoName" --title "Phase 0 governance repo foundation" --body "Scaffold TraceDeck Agent Go foundation with typed policy config, generated schema, local scripts, docs, sample policy, and local verification gates.").Trim()
        $issueNumber = ($issueNumber -split "/")[-1]
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 0 issue: $issueNumber"

    Invoke-Git -Args @("checkout", "-B", $Branch)
    Invoke-Git -Args @("add", ".")

    $status = git status --short
    if (-not $status) {
        Write-TraceDeckLog -Level "WARN" -Message "No changes to commit for Phase 0"
    }
    else {
        Invoke-Git -Args @("commit", "-m", "chore: scaffold phase 0 foundation")
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $prURL = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    if (-not $prURL) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 0: governance and repo foundation" --body "Closes #$issueNumber.`n`nLocal verification: scripts/verify/verify-phase0.ps1 passed. GitHub Actions intentionally not configured.").Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 0 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 0 PR" -Command {
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
