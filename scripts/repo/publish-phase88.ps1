param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/88-cache-visual-contract-hardening",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase88" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 88 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase88.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Patch the Phase 87 follow-up contract so live provenance cache guards and screenshot-free Playwright checks are complete and reproducible.

## Scope
- add the missing shared PowerShell HTTP constants helper used by live provenance checks
- verify dashboard, browser activity, and API no-store headers
- keep seeded VLC/media rows hidden from default host APIs and explicit under `include_demo=true`
- replace brittle Playwright `networkidle` waits with DOM and hydrated-control readiness waits
- add Phase 88 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase88.ps1
- scripts/verify/verify-phase87.ps1
- scripts/local/smoke-phase88.ps1
- scripts/local/newman-phase88.ps1
- scripts/local/test-live-server-provenance.ps1

## Privacy
Metadata and rendered layout metrics only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 88: cache and visual contract hardening" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 88 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/backend-api.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase88.postman_collection.json",
        "scripts/lib/http-constants.ps1",
        "scripts/local/newman-phase88.ps1",
        "scripts/local/smoke-phase88.ps1",
        "scripts/local/test-live-server-provenance.ps1",
        "scripts/repo/publish-phase87.ps1",
        "scripts/repo/publish-phase88.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/tools/dashboard_theme_check.py",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase88.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: harden cache and visual contracts", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 88"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add the shared PowerShell HTTP constants helper required by live provenance checks
- assert no-store headers for dashboard, browser activity, health, and host evidence APIs
- keep default host policy evidence free of seeded VLC/demo rows
- replace Playwright network-idle waits with DOM and hydrated-control readiness waits
- add Phase 88 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase88.ps1
- scripts/verify/verify-phase87.ps1
- scripts/local/smoke-phase88.ps1
- scripts/local/newman-phase88.ps1
- scripts/local/test-live-server-provenance.ps1

## Privacy
Metadata and rendered layout metrics only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 88: cache and visual contract hardening" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 88 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 88 PR" -Command {
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
