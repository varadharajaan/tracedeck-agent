param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/99-verification-evidence-center",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase99" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 99 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase99.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make the autonomous verification contract visible as typed local evidence.

## Scope
- add `/api/v1/verification-evidence-center`
- add typed verification evidence models and centralized constants
- add a Rollout page Verification Evidence Center dashboard panel
- add a metadata-only evidence generator under `scripts/local/`
- add Phase 99 smoke/Newman/verify/publish scripts and Postman collection
- update docs and devctl aliases

## Verification
- scripts/verify/verify-phase99.ps1
- scripts/local/smoke-phase99.ps1
- scripts/local/newman-phase99.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational verification metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 99: verification evidence center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 99 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/testing.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "postman/tracedeck-backend-phase99.postman_collection.json",
        "scripts/local/get-verification-evidence.ps1",
        "scripts/local/smoke-phase99.ps1",
        "scripts/local/newman-phase99.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase99.ps1",
        "scripts/repo/publish-phase99.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add verification evidence center", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 99"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- expose a typed Verification Evidence Center from a local evidence artifact
- render scripted gate, report, action, and privacy proof in the dashboard
- add Phase 99 evidence generator, smoke/Newman/verify scripts, and Postman coverage

## Verification
- scripts/verify/verify-phase99.ps1
- scripts/local/smoke-phase99.ps1
- scripts/local/newman-phase99.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational verification metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 99: verification evidence center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 99 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 99 PR" -Command {
        gh pr merge $prURL --merge --delete-branch
    }

    Invoke-Git -Args @("checkout", "main")
    Invoke-Git -Args @("pull", "--ff-only", "origin", "main")

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
