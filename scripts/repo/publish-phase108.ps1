param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/108-browser-extension-skeleton",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase108" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 108 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase108.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 108 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 108 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a Chrome, Edge, and Brave browser extension skeleton that sends TraceDeck domain/category metadata to the existing localhost telemetry ingest route.

## Scope
- add `browser-extension/` manifest, background worker, options page, and privacy core
- keep the extension local-only and metadata-only
- add Node/static privacy contract tests
- add live smoke and Newman coverage for extension-shaped telemetry ingest
- update docs, phase ledger, roadmap, and devctl aliases

## Verification
- scripts/local/test-browser-extension-skeleton.ps1
- scripts/local/smoke-phase108.ps1
- scripts/local/newman-phase108.ps1
- scripts/verify/verify-phase108.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads are stored or transmitted.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 108: browser extension skeleton" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 108 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "browser-extension/README.md",
        "browser-extension/manifest.json",
        "browser-extension/src/background.js",
        "browser-extension/src/options.html",
        "browser-extension/src/options.js",
        "browser-extension/src/privacy-core.js",
        "browser-extension/test/privacy-core.test.js",
        "devctl.py",
        "docs/browser-extension.md",
        "docs/contract-completion-audit.md",
        "docs/phase-ledger.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-browser-extension-phase108.postman_collection.json",
        "scripts/local/newman-phase108.ps1",
        "scripts/local/smoke-phase108.ps1",
        "scripts/local/test-browser-extension-skeleton.ps1",
        "scripts/repo/publish-phase108.ps1",
        "scripts/verify/verify-phase108.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add browser extension skeleton", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 108"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add the TraceDeck Chromium browser extension skeleton for Chrome, Edge, and Brave
- normalize navigation observations into domain/category telemetry only
- post extension-shaped events to the existing localhost telemetry ingest route
- add privacy/static, live smoke, and Newman verification
- update docs, audit, ledger, roadmap, and devctl aliases

## Verification
- scripts/local/test-browser-extension-skeleton.ps1
- scripts/local/smoke-phase108.ps1
- scripts/local/newman-phase108.ps1
- scripts/verify/verify-phase108.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads are stored or transmitted.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 108: browser extension skeleton" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 108 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 108 PR" -Command {
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
