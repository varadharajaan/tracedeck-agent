Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase68" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/68-browser-activity-viewer"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 68 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase68.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a Browser Activity Viewer page linked from the main dashboard, backed by a typed tenant API for Chrome, Brave, and Edge domain activity with host filters, category filters, study-safe suppression, non-study YouTube review, notification proof, and privacy guardrails.

## Verification
- scripts/verify/verify-phase68.ps1
- scripts/local/smoke-phase68.ps1
- scripts/local/newman-phase68.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, or alert bodies.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 68: browser activity viewer" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 68 issue #${IssueNumber}: $issueURL"
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
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase68.postman_collection.json",
        "scripts/local/newman-phase68.ps1",
        "scripts/local/smoke-phase68.ps1",
        "scripts/repo/publish-phase68.ps1",
        "scripts/verify/verify-phase68.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add browser activity viewer")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 68"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a typed tenant Browser Activity API with Chrome, Edge, Brave, host, category, study-safe, non-study YouTube, and notification proof summary
- add a dedicated Browser Activity Viewer page and dashboard redirect button
- add Phase 68 Go, DOM, smoke, Newman, JavaScript, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase68.ps1 passed
- scripts/local/smoke-phase68.ps1 live-booted the seeded dashboard, verified the browser viewer page, checked typed browser activity data, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase68.ps1 passed against a live dashboard demo
- backend, agent, dashboard contract, dashboard JavaScript syntax, browser activity JavaScript syntax, service manifest rendering, cross-platform Windows/macOS/Linux builds, and root artifact checks passed locally

## Privacy
- metadata-only browser activity surface; no passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, or alert bodies are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 68: browser activity viewer" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 68 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 68 PR" -Command {
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
