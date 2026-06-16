param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/98-runtime-status-center",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase98" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 98 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase98.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Surface the Phase 97 runtime summary as a typed local API and dashboard Runtime Status Center.

## Scope
- add `/api/v1/runtime-status-center`
- add typed runtime summary/status models and constants
- add a Rollout page Runtime Status Center dashboard panel
- add Phase 98 smoke/Newman/verify/publish scripts and Postman collection
- document the runtime status workflow

## Verification
- scripts/verify/verify-phase98.ps1
- scripts/local/smoke-phase98.ps1
- scripts/local/newman-phase98.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 98: runtime status center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 98 issue #${IssueNumber}: $issueURL"
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
        "postman/tracedeck-backend-phase98.postman_collection.json",
        "scripts/local/get-runtime-summary.ps1",
        "scripts/local/smoke-phase98.ps1",
        "scripts/local/newman-phase98.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase98.ps1",
        "scripts/repo/publish-phase98.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add runtime status center", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 98"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- expose a typed runtime status center from the Phase 97 summary artifact
- render a Rollout page Runtime Status Center in the local dashboard
- add Phase 98 smoke/Newman/verify scripts and Postman coverage

## Verification
- scripts/verify/verify-phase98.ps1
- scripts/local/smoke-phase98.ps1
- scripts/local/newman-phase98.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 98: runtime status center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 98 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 98 PR" -Command {
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
