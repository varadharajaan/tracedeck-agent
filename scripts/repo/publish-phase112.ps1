param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/112-local-monitoring-indicator",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase112" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 112 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase112.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 112 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 112 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a visible local monitoring indicator so endpoint users and admins can see TraceDeck monitoring status from the host.

## Scope
- add `GET /api/v1/local-monitoring-indicator`
- add typed local indicator models, proof rows, and action rows
- generate JSON, text, and HTML local status artifacts under `data/local/output`
- render a Local Monitoring Indicator dashboard panel and navigator entry
- replace the cramped Alert Delivery table with responsive delivery cards
- add explicit demo push truth: no web-push provider send and no screen notification
- update platform support, audit, docs, smoke, Newman, verifier, and devctl coverage

## Verification
- scripts/local/get-local-monitoring-indicator.ps1
- scripts/local/test-dashboard-delivery-ui.ps1
- scripts/local/smoke-dashboard-delivery-ui.ps1
- scripts/local/smoke-phase112.ps1
- scripts/local/newman-phase112.ps1
- scripts/verify/verify-phase112.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
The local indicator is metadata-only. It exposes status labels, local paths, command labels, consent visibility, runtime labels, and denied sensitive capability labels. It does not collect passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, provider secrets, alert bodies, keylogging data, hidden collection bypasses, payment data, raw provider payloads, camera, or microphone data.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 112: Local monitoring indicator" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 112 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "agent/internal/platform/support.go",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "devctl.py",
        "docs/backend-api.md",
        "docs/collection-policy.md",
        "docs/contract-completion-audit.md",
        "docs/dashboard.md",
        "docs/local-monitoring-indicator.md",
        "docs/phase-ledger.md",
        "docs/platform-support.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase112.postman_collection.json",
        "scripts/local/get-contract-completion-audit.ps1",
        "scripts/local/get-local-monitoring-indicator.ps1",
        "scripts/local/newman-phase112.ps1",
        "scripts/local/smoke-dashboard-delivery-ui.ps1",
        "scripts/local/smoke-phase112.ps1",
        "scripts/local/test-dashboard-delivery-ui.ps1",
        "scripts/repo/publish-phase112.ps1",
        "scripts/tools/dashboard_delivery_ui_check.py",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase112.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add local monitoring indicator", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 112"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed local monitoring indicator API, proof, and actions
- generate local JSON/text/HTML indicator artifacts through script
- render Local Monitoring Indicator in the dashboard and Workspace Navigator
- replace the Alert Delivery table with responsive delivery cards and demo-safe push proof labels
- update platform support, audit, docs, smoke, Newman, verifier, and devctl aliases

## Verification
- scripts/local/get-local-monitoring-indicator.ps1
- scripts/local/test-dashboard-delivery-ui.ps1
- scripts/local/smoke-dashboard-delivery-ui.ps1
- scripts/local/smoke-phase112.ps1
- scripts/local/newman-phase112.ps1
- scripts/verify/verify-phase112.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
The local indicator remains metadata-only and does not collect passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, provider secrets, alert bodies, keylogging data, hidden collection bypasses, payment data, raw provider payloads, camera, or microphone data.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 112: Local monitoring indicator" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 112 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 112 PR" -Command {
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
