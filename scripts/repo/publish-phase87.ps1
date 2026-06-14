param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/87-trust-quality-ui-hardening",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase87" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 87 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase87.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Package the verified local hardening work: strict Go quality gates, premium UI recovery, and demo-provenance trust fixes.

## Scope
- add strict Go quality-gate scripts and reports under `data/local/go-quality/...`
- add Phase 86 premium dashboard and Browser Viewer UI verification
- hide seeded `demo_seed` risk/delivery evidence from default host APIs
- keep seeded VLC/media-playback rows available only through `?include_demo=true`
- prevent generated weekly reports or demo email rows from claiming delivered email proof
- update smoke, Newman, live provenance, docs, Postman, and devctl hooks

## Verification
- scripts/verify/verify-phase87.ps1
- scripts/verify/verify-phase86.ps1
- scripts/local/test-go-quality-gates.ps1
- scripts/local/smoke-phase86.ps1
- scripts/local/newman-phase86.ps1
- scripts/local/test-live-server-provenance.ps1

## Privacy
Quality, UI, and metadata-only provenance checks. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 87: trust, quality, and UI hardening" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 87 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "agent/internal/constants/policy.go",
        "agent/internal/storage/sqlite/store.go",
        "agent/internal/storage/sqlite/store_test.go",
        "agent/internal/syncer/client.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/browser_activity.html",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/report.go",
        "backend/internal/store/report_test.go",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase85.postman_collection.json",
        "postman/tracedeck-backend-phase86.postman_collection.json",
        "scripts/local/newman-phase85.ps1",
        "scripts/local/newman-phase86.ps1",
        "scripts/local/smoke-phase86.ps1",
        "scripts/local/start-dashboard-demo.ps1",
        "scripts/local/test-go-quality-gates.ps1",
        "scripts/local/test-live-server-provenance.ps1",
        "scripts/repo/publish-phase85.ps1",
        "scripts/repo/publish-phase87.ps1",
        "scripts/verify/verify-phase85.ps1",
        "scripts/verify/verify-phase86.ps1",
        "scripts/verify/verify-phase87.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "build: harden trust quality and premium UI gates", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 87"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add strict Go quality gates and Phase 85 runtime coverage
- add Phase 86 premium dashboard/Browser Viewer UI coverage
- make demo risk and delivery rows opt-in with `include_demo=true`
- prevent seeded VLC/media playback and demo email rows from being shown as live host truth or delivered proof
- add Phase 87 verifier/publish script to prove the whole bundle before merge

## Verification
- scripts/verify/verify-phase87.ps1
- scripts/verify/verify-phase86.ps1
- scripts/local/test-go-quality-gates.ps1
- scripts/local/test-live-server-provenance.ps1

## Privacy
Quality, UI, and metadata-only provenance checks. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 87: trust, quality, and UI hardening" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 87 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 87 PR" -Command {
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
