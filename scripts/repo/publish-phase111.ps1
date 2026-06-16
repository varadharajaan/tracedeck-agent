param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/111-software-install-collector",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase111" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 111 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase111.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 111 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 111 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add metadata-only software install/uninstall event collection behind TraceDeck platform adapters.

## Scope
- add typed `collection.software.inventory_mode: metadata_only` policy config
- add platform software inventory adapters for Windows, macOS, Linux, and explicit unsupported fallback
- diff local metadata snapshots to emit `software.installed` and `software.uninstalled`
- reuse risky-software alerting for install events and add `unknown_software_installed`
- add focused tests, smoke, Newman, schema, docs, audit, roadmap, devctl, and publish coverage

## Verification
- scripts/local/test-software-inventory-collector.ps1
- scripts/local/smoke-phase111.ps1
- scripts/local/newman-phase111.ps1
- scripts/verify/verify-phase111.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Software inventory collection is metadata-only. It does not collect install paths, file contents, screenshots, passwords, raw URLs, page titles, cookies, tokens, private content, provider secrets, alert bodies, keylogging data, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 111: Software install collector" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 111 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "agent/internal/alert/evaluator.go",
        "agent/internal/alert/evaluator_test.go",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/cli/root.go",
        "agent/internal/collector/activewindow/collector_test.go",
        "agent/internal/collector/browser/collector_test.go",
        "agent/internal/collector/heartbeat/collector_test.go",
        "agent/internal/collector/software/collector.go",
        "agent/internal/collector/software/collector_test.go",
        "agent/internal/config/config_test.go",
        "agent/internal/config/enums.go",
        "agent/internal/config/schema_enums.go",
        "agent/internal/config/types.go",
        "agent/internal/config/validate.go",
        "agent/internal/constants/config_fields.go",
        "agent/internal/constants/events.go",
        "agent/internal/constants/project.go",
        "agent/internal/constants/schema.go",
        "agent/internal/constants/software.go",
        "agent/internal/platform/adapter.go",
        "agent/internal/platform/adapter_test.go",
        "agent/internal/platform/current_darwin.go",
        "agent/internal/platform/current_linux.go",
        "agent/internal/platform/current_other.go",
        "agent/internal/platform/current_windows.go",
        "agent/internal/platform/support.go",
        "devctl.py",
        "docs/agent-telemetry-ingest.md",
        "docs/collection-policy.md",
        "docs/contract-completion-audit.md",
        "docs/phase-ledger.md",
        "docs/platform-support.md",
        "docs/privacy.md",
        "docs/risky-software-detection.md",
        "docs/roadmap.md",
        "docs/schema/policy-v1alpha1.schema.json",
        "docs/telemetry-schema.md",
        "docs/testing.md",
        "examples/policies/ai-btech-student.yaml",
        "postman/tracedeck-backend-phase111.postman_collection.json",
        "scripts/local/get-contract-completion-audit.ps1",
        "scripts/local/newman-phase111.ps1",
        "scripts/local/smoke-phase111.ps1",
        "scripts/local/test-software-inventory-collector.ps1",
        "scripts/repo/publish-phase111.ps1",
        "scripts/verify/verify-phase111.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add software install collector", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 111"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed software inventory collection config in metadata-only mode
- add platform software inventory adapters and local snapshot-diff change events
- emit `software.installed` and `software.uninstalled` without install paths or file contents
- reuse risky-software alerting for software install events and add unknown install alerts
- add smoke/Newman/verifier/publish scripts, schema, docs, audit, roadmap, and devctl aliases

## Verification
- scripts/local/test-software-inventory-collector.ps1
- scripts/local/smoke-phase111.ps1
- scripts/local/newman-phase111.ps1
- scripts/verify/verify-phase111.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Software inventory collection remains metadata-only and does not collect install paths, file contents, screenshots, passwords, raw URLs, page titles, cookies, tokens, private content, provider secrets, alert bodies, keylogging data, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 111: Software install collector" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 111 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 111 PR" -Command {
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
