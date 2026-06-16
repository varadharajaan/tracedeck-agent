param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/100-operator-assurance-center",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase100" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 100 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase100.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 100 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 100 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make local handoff readiness obvious with one metadata-only operator assurance center.

## Scope
- add `/api/v1/operator-assurance-center`
- combine runtime status and verification evidence into typed assurance cards
- explain Scheduler denied vs runtime healthy states clearly
- add dashboard Operator Assurance Center panel
- add metadata-only operator assurance JSON/text export script
- add Phase 100 smoke/Newman/verify/publish scripts and Postman collection
- add bounded backend health wait helper for persistent dev-server refresh
- update docs and devctl aliases

## Verification
- scripts/verify/verify-phase100.ps1
- scripts/local/smoke-phase100.ps1
- scripts/local/newman-phase100.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational assurance metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 100: operator assurance center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 100 issue #${IssueNumber}: $issueURL"
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
        "postman/tracedeck-backend-phase100.postman_collection.json",
        "scripts/local/get-operator-assurance.ps1",
        "scripts/local/smoke-phase100.ps1",
        "scripts/local/newman-phase100.ps1",
        "scripts/local/wait-backend-health.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/tools/dashboard_theme_check.py",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase100.ps1",
        "scripts/repo/publish-phase100.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add operator assurance center", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 100"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- expose a typed Operator Assurance Center from local runtime and verification evidence
- render runtime, Scheduler, gate, frontend cache, git, export, and privacy proof in the dashboard
- add Phase 100 assurance export, smoke/Newman/verify scripts, and Postman coverage
- refresh the persistent dev backend through a bounded health wait during verification

## Verification
- scripts/verify/verify-phase100.ps1
- scripts/local/smoke-phase100.ps1
- scripts/local/newman-phase100.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational assurance metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 100: operator assurance center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 100 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 100 PR" -Command {
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
