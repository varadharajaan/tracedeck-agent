param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/75-delivery-assurance",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase75" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 75 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase75.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make notification delivery truth explicit so demo, retrying, dashboard-only, dry-run, and provider-confirmed states cannot be confused.

## Scope
- add tenant delivery assurance API
- add dashboard Delivery Assurance Center
- separate provider-confirmed, dry-run, dashboard-visible, demo-only, retrying, failed, disabled, and pending-provider states
- wire Phase 75 smoke, Newman, verify, publish, Postman, docs, and runtime doctor coverage

## Verification
- scripts/verify/verify-phase75.ps1
- scripts/local/smoke-phase75.ps1
- scripts/local/newman-phase75.ps1
- scripts/local/test-runtime-doctor.ps1 through smoke-phase75

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 75: delivery assurance truth center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 75 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/memory_test.go",
        "backend/internal/store/repository.go",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/notification-route-registry.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase75.postman_collection.json",
        "scripts/lib/logging.ps1",
        "scripts/local/newman-phase75.ps1",
        "scripts/local/smoke-phase75.ps1",
        "scripts/repo/publish-phase75.ps1",
        "scripts/verify/verify-phase75.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add delivery assurance truth center", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 75"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a typed tenant delivery assurance API
- add the dashboard Delivery Assurance Center
- label demo, retrying, dashboard-only, dry-run, failed, disabled, pending-provider, and provider-confirmed states separately
- add Phase 75 smoke/Newman/verify/publish scripts, Postman coverage, docs, and runtime doctor assertions

## Verification
- scripts/verify/verify-phase75.ps1
- scripts/local/smoke-phase75.ps1
- scripts/local/newman-phase75.ps1
- scripts/local/test-runtime-doctor.ps1 through smoke-phase75

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 75: delivery assurance truth center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 75 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 75 PR" -Command {
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
