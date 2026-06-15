param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/90-runtime-doctor-delivery-truth",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase90" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 90 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase90.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Keep local runtime doctor and delivery proof language aligned with the demo provenance contract.

## Scope
- update `devctl.py doctor` to report default alert-delivery rows separately from opt-in demo rows
- prove default `/alert-deliveries` hides `source_kind=demo_seed`
- prove explicit `include_demo=true` still exposes labelled demo delivery proof
- keep buyer-ready notification proof false without provider-confirmed mail and push delivery
- update smoke, Newman, verifier, docs, and devctl test targets

## Verification
- scripts/verify/verify-phase90.ps1
- scripts/local/test-devctl-runtime-doctor.ps1
- scripts/local/smoke-phase90.ps1
- scripts/local/newman-phase90.ps1
- postman/tracedeck-backend-phase90.postman_collection.json

## Privacy
Metadata and provenance flags only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 90: runtime doctor delivery proof hardening" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 90 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/dashboard-demo-lifecycle.md",
        "docs/notification-route-registry.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase90.postman_collection.json",
        "scripts/local/newman-phase90.ps1",
        "scripts/local/smoke-phase90.ps1",
        "scripts/local/test-devctl-runtime-doctor.ps1",
        "scripts/repo/publish-phase90.ps1",
        "scripts/verify/check-root-clean.ps1",
        "scripts/verify/verify-phase90.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "fix: harden runtime delivery proof doctor", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 90"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- update runtime doctor delivery checks so default alert deliveries do not require demo rows
- report default delivery rows separately from explicit `include_demo=true` demo proof
- add Phase 90 smoke/Newman/verify scripts and devctl test targets
- document the demo versus provider-confirmed delivery boundary

## Verification
- scripts/verify/verify-phase90.ps1
- scripts/local/test-devctl-runtime-doctor.ps1
- scripts/local/smoke-phase90.ps1
- scripts/local/newman-phase90.ps1
- scripts/verify/verify-phase89.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Metadata and provenance flags only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 90: runtime doctor delivery proof hardening" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 90 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 90 PR" -Command {
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
