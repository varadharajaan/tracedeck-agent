Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase66" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/66-deployment-readiness-center"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 66 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase66.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a Deployment Readiness Center that proves Windows Task Scheduler, macOS launchd, Linux systemd, reboot persistence, background startup, live boot, offline replay, service manifests, archive backlog, and owner actions from one metadata-only product surface.

## Verification
- scripts/verify/verify-phase66.ps1
- scripts/local/smoke-phase66.ps1
- scripts/local/newman-phase66.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, raw provider payloads, keylogging, or hidden collection bypasses.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 66: deployment readiness center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 66 issue #${IssueNumber}: $issueURL"
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
        "backend/internal/store/repository.go",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/monetization.md",
        "docs/platform-support.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/service-management.md",
        "docs/testing.md",
        "docs/windows-autostart.md",
        "postman/tracedeck-backend-phase66.postman_collection.json",
        "scripts/local/newman-phase66.ps1",
        "scripts/local/smoke-phase66.ps1",
        "scripts/repo/publish-phase66.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase66.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add deployment readiness center")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 66"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a typed deployment readiness center API for Windows Task Scheduler, macOS launchd, Linux systemd, service manifest, live boot, background start, offline replay, archive backlog, and owner-action proof
- render the deployment readiness center in the monetisation dashboard with platform, manifest, boot/replay, and deployment action panels
- add Phase 66 smoke, Newman, service/autostart, DOM/layout, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase66.ps1 passed
- scripts/local/smoke-phase66.ps1 live-booted the seeded dashboard, verified deployment readiness UX markers, checked typed deployment readiness data, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase66.ps1 passed against a live dashboard demo
- service manifest rendering, Windows task template, autostart assurance, service manager dry-run, and cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only deployment readiness surface; no passwords, screenshots, raw URLs, page titles, alert bodies, provider secrets, push endpoints, endpoint payloads, private content, invoices, payment card data, tokens, cookies, raw provider payloads, keylogging, or hidden collection bypasses are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 66: deployment readiness center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 66 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 66 PR" -Command {
        gh pr merge $prURL --squash --delete-branch
    }

    Invoke-Git -Args @("checkout", "main")
    Invoke-Git -Args @("pull", "--ff-only", "origin", "main")

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
