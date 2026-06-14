param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/83-agent-heartbeat-telemetry",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase83" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 83 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase83.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a metadata-only agent heartbeat telemetry event so admins can prove each host agent is alive, configured, syncing, and visible in backend telemetry without collecting sensitive content.

## Scope
- add `agent.health.heartbeat` from `collector.agent.heartbeat`
- emit one heartbeat per agent collection cycle
- include typed readiness metadata for agent version, collection mode/interval, archive state, backend sync state, alert state, profile, and operating system
- surface heartbeat type/source counts through telemetry status and tenant sync-health proof
- keep the dashboard, Browser Viewer, and Lambda Cloud Admin on the polished Phase 84 visual layer
- add Phase 83 smoke, Newman, verify, publish, Postman, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase83.ps1
- scripts/local/smoke-phase83.ps1
- scripts/local/newman-phase83.ps1
- scripts/local/test-dashboard-visual-quality.ps1 through verify-phase83
- scripts/local/test-dashboard-theme.ps1 through verify-phase83
- scripts/local/test-lambda-frontend-visual.ps1 through verify-phase83

## Privacy
Metadata-only readiness telemetry. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 83: agent heartbeat telemetry" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 83 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/collector/heartbeat/collector.go",
        "agent/internal/collector/heartbeat/collector_test.go",
        "agent/internal/constants/events.go",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/web/browser_activity.html",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/store/memory_test.go",
        "devctl.py",
        "docs/agent-telemetry-ingest.md",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/telemetry-schema.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase83.postman_collection.json",
        "sam-app/frontend_function/app.py",
        "scripts/local/newman-phase83.ps1",
        "scripts/local/smoke-phase83.ps1",
        "scripts/repo/publish-phase83.ps1",
        "scripts/verify/verify-phase83.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: emit agent heartbeat telemetry", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 83"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a metadata-only `agent.health.heartbeat` collector and emit one heartbeat per agent cycle
- include typed readiness metadata for agent health, version, collection mode, interval, archive state, backend sync state, alert state, profile, and operating system
- count heartbeat type/source in backend telemetry status and prove backend-visible replay through tenant sync-health
- keep local dashboard, Browser Viewer, and Lambda Cloud Admin on the polished product UI layer
- add Phase 83 smoke/Newman/verify/publish scripts, Postman coverage, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase83.ps1
- scripts/local/smoke-phase83.ps1
- scripts/local/newman-phase83.ps1
- scripts/local/test-dashboard-visual-quality.ps1
- scripts/local/test-dashboard-theme.ps1
- scripts/local/test-lambda-frontend-visual.ps1

## Privacy
Metadata-only readiness telemetry. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 83: agent heartbeat telemetry" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 83 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 83 PR" -Command {
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
