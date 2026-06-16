param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/110-foreground-app-collector",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase110" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 110 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase110.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 110 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 110 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a metadata-only foreground app/window collector behind the TraceDeck platform adapter contract.

## Scope
- add typed `collection.foreground_app` policy config
- add Windows foreground app adapter using desktop APIs without window-title reads
- keep macOS and Linux explicit through unsupported/permission-aware adapter responses
- emit bounded `foreground_app.observed` metadata events with app name, PID, path hash, state, and window-title mode only
- reuse blocked-app and risky-software alert evaluation for foreground app events
- add focused tests, smoke, Newman, schema, docs, audit, ledger, and devctl coverage

## Verification
- scripts/local/test-active-window-collector.ps1
- scripts/local/smoke-phase110.ps1
- scripts/local/newman-phase110.ps1
- scripts/verify/verify-phase110.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Foreground app collection is metadata-only. It does not collect screenshots, window titles, raw URLs, page titles, passwords, cookies, tokens, private content, provider secrets, alert bodies, keylogging data, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 110: Foreground app collector" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 110 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "agent/internal/alert/evaluator.go",
        "agent/internal/alert/evaluator_test.go",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/collector/activewindow/collector.go",
        "agent/internal/collector/activewindow/collector_test.go",
        "agent/internal/collector/browser/collector_test.go",
        "agent/internal/collector/heartbeat/collector_test.go",
        "agent/internal/config/config_test.go",
        "agent/internal/config/enums.go",
        "agent/internal/config/schema_enums.go",
        "agent/internal/config/types.go",
        "agent/internal/config/validate.go",
        "agent/internal/constants/config_fields.go",
        "agent/internal/constants/events.go",
        "agent/internal/constants/policy.go",
        "agent/internal/constants/schema.go",
        "agent/internal/platform/adapter.go",
        "agent/internal/platform/current_darwin.go",
        "agent/internal/platform/current_linux.go",
        "agent/internal/platform/current_other.go",
        "agent/internal/platform/current_windows.go",
        "backend/internal/constants/constants.go",
        "devctl.py",
        "docs/agent-telemetry-ingest.md",
        "docs/collection-policy.md",
        "docs/contract-completion-audit.md",
        "docs/phase-ledger.md",
        "docs/platform-support.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/schema/policy-v1alpha1.schema.json",
        "docs/telemetry-schema.md",
        "docs/testing.md",
        "examples/policies/ai-btech-student.yaml",
        "postman/tracedeck-backend-phase110.postman_collection.json",
        "scripts/local/get-contract-completion-audit.ps1",
        "scripts/local/newman-phase110.ps1",
        "scripts/local/smoke-phase110.ps1",
        "scripts/local/test-active-window-collector.ps1",
        "scripts/repo/publish-phase110.ps1",
        "scripts/verify/check-cross-platform-build.ps1",
        "scripts/verify/verify-phase110.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add foreground app collector", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 110"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed foreground app collection policy config
- add Windows foreground app adapter without screenshot or window-title collection
- emit metadata-only `foreground_app.observed` events with app, PID, path hash, active state, and title mode
- reuse blocked-app and risky-software rules for foreground app observations
- add smoke/Newman/verifier/publish scripts, schema, docs, audit, ledger, and devctl aliases

## Verification
- scripts/local/test-active-window-collector.ps1
- scripts/local/smoke-phase110.ps1
- scripts/local/newman-phase110.ps1
- scripts/verify/verify-phase110.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Foreground app collection remains metadata-only and does not collect screenshots, passwords, raw URLs, page titles, window titles, cookies, tokens, private content, provider secrets, alert bodies, keylogging data, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 110: Foreground app collector" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 110 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 110 PR" -Command {
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
