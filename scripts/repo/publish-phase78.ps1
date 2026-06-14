param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/78-notification-provider-setup-center",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase78" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 78 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase78.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a dashboard Notification Provider Setup Center so admins can clearly see why email or push has not reached a recipient, even when routes are configured or demo rows exist.

## Scope
- expose the typed notification-provider-setup API in the dashboard
- show provider-confirmed versus configured/demo-only/retrying states
- refresh the dashboard and Browser Viewer visual shell so paid-product screens no longer show debug-style labels
- add provider setup checklist and owner actions
- add workspace navigator, layout, theme, and visual-quality contract coverage
- add Phase 78 smoke, Newman, verify, publish, Postman, and docs coverage

## Verification
- scripts/verify/verify-phase78.ps1
- scripts/local/smoke-phase78.ps1
- scripts/local/newman-phase78.ps1
- scripts/local/test-dashboard-layout.ps1 through smoke-phase78
- scripts/local/test-dashboard-theme.ps1 through smoke-phase78
- scripts/local/test-dashboard-visual-quality.ps1 through smoke-phase78

## Privacy
Metadata-only provider setup proof. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 78: notification provider setup center" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 78 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "backend/internal/api/dashboard_contract_test.go",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/api/web/browser_activity.html",
        "backend/internal/api/web/dashboard.html",
        "backend/internal/constants/constants.go",
        "backend/internal/model/model.go",
        "backend/internal/store/memory.go",
        "backend/internal/store/repository.go",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase9.postman_collection.json",
        "postman/tracedeck-backend-phase42.postman_collection.json",
        "postman/tracedeck-backend-phase43.postman_collection.json",
        "postman/tracedeck-backend-phase74.postman_collection.json",
        "postman/tracedeck-backend-phase76.postman_collection.json",
        "postman/tracedeck-backend-phase78.postman_collection.json",
        "scripts/local/smoke-phase6.ps1",
        "scripts/local/smoke-phase9.ps1",
        "scripts/local/smoke-phase42.ps1",
        "scripts/local/smoke-phase76.ps1",
        "scripts/local/newman-phase78.ps1",
        "scripts/local/smoke-phase78.ps1",
        "scripts/local/start-dashboard-demo.ps1",
        "scripts/local/test-dashboard-visual-quality.ps1",
        "scripts/repo/publish-phase78.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/tools/dashboard_visual_quality_check.py",
        "scripts/verify/verify-phase78.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add notification provider setup center", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 78"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add dashboard Notification Provider Setup Center
- show configured versus provider-confirmed email/push/dashboard state
- revamp dashboard and Browser Viewer styling with product-grade shell, tabs, cards, chips, and navigator labels
- surface demo-only, retrying, buyer-ready, checklist, owner action, and privacy boundary proof
- add Phase 78 smoke/Newman/verify/publish scripts, Postman coverage, docs, layout, theme, and visual-quality assertions

## Verification
- scripts/verify/verify-phase78.ps1
- scripts/local/smoke-phase78.ps1
- scripts/local/newman-phase78.ps1

## Privacy
Metadata-only provider setup proof. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 78: notification provider setup center" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 78 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 78 PR" -Command {
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
