param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/89-activity-feed-provenance",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase89" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 89 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase89.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Prevent seeded demo activity feed rows from appearing as live host evidence.

## Scope
- add typed `include_demo` to tenant activity feed filters
- hide `source_kind=demo_seed` rows from default tenant and host-scoped activity feed responses
- keep explicit `include_demo=true` demo behavior for seeded VLC, email, and push examples
- update backend tests, Postman collections, smoke scripts, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase89.ps1
- scripts/local/test-activity-feed-provenance.ps1
- scripts/local/smoke-phase89.ps1
- scripts/local/newman-phase89.ps1
- scripts/local/smoke-phase32.ps1
- scripts/local/smoke-phase41.ps1

## Privacy
Metadata and provenance flags only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 89: activity feed demo provenance hardening" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 89 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/memory_test.go",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase32.postman_collection.json",
        "postman/tracedeck-backend-phase41.postman_collection.json",
        "postman/tracedeck-backend-phase88.postman_collection.json",
        "postman/tracedeck-backend-phase89.postman_collection.json",
        "scripts/local/newman-phase89.ps1",
        "scripts/local/smoke-phase32.ps1",
        "scripts/local/smoke-phase41.ps1",
        "scripts/local/smoke-phase89.ps1",
        "scripts/local/test-activity-feed-provenance.ps1",
        "scripts/local/test-live-server-provenance.ps1",
        "scripts/repo/publish-phase89.ps1",
        "scripts/verify/verify-phase89.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "fix: hide demo rows from activity feed", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 89"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed `include_demo` to tenant activity feed filters
- hide seeded `VLC media player` and other `demo_seed` rows from default tenant activity feeds
- keep explicit `include_demo=true` demo responses labelled and reproducible
- update Phase 32/41/88 contracts plus Phase 89 smoke/Newman/verify/devctl/docs

## Verification
- scripts/verify/verify-phase89.ps1
- scripts/local/test-activity-feed-provenance.ps1
- scripts/local/smoke-phase89.ps1
- scripts/local/newman-phase89.ps1
- scripts/local/smoke-phase32.ps1
- scripts/local/smoke-phase41.ps1
- scripts/local/newman-phase32.ps1
- scripts/local/newman-phase41.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Metadata and provenance flags only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 89: activity feed demo provenance hardening" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 89 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 89 PR" -Command {
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
