Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase69" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/69-admin-ui-lambda-frontend"
$IssueNumber = ""

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 69 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase69.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add TraceDeck admin UI polish and cloud admin access: multipage dashboard tabs, dark theme, server status lights, Browser Activity Viewer parity, a SAM Lambda Function URL frontend backed by S3 cache metrics, and a devctl controller for local server/test/SAM/log operations.

## Verification
- scripts/verify/verify-phase69.ps1
- scripts/local/smoke-phase69.ps1
- scripts/local/newman-phase69.ps1
- scripts/local/test-lambda-frontend-contract.ps1

## Deployment Contract
- Lambda Function URL only
- no API Gateway
- stack outputs saved under data/local/output/
- local backend remains localhost-bound

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 69: admin UI and Lambda frontend" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 69 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        ".gitignore",
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
        "docs/cloud-frontend.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase69.postman_collection.json",
        "sam-app/samconfig.toml",
        "sam-app/template.yaml",
        "sam-app/frontend_function/app.py",
        "scripts/local/newman-phase69.ps1",
        "scripts/local/smoke-phase69.ps1",
        "scripts/local/test-lambda-frontend-contract.ps1",
        "scripts/repo/publish-phase69.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase69.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add admin ui and lambda frontend")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 69"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add dashboard page tabs, dark theme toggle, and green/red server status lights
- add Browser Activity Viewer theme/status parity
- add a SAM Lambda Function URL frontend backed by S3 summaries, in-memory cache metrics, and a localhost source switch
- add devctl.py for local server, tests, SAM deployment, output saving, and logs
- add Phase 69 smoke, Newman, Lambda contract, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase69.ps1 passed
- scripts/local/smoke-phase69.ps1 live-booted the dashboard, verified UI markers, browser page parity, layout contract, Lambda frontend contract, and devctl status
- scripts/local/newman-phase69.ps1 passed against a live dashboard demo
- scripts/local/test-lambda-frontend-contract.ps1 verified Python compile, SAM Function URL/no-API-Gateway markers, and local handler routes

## Deployment Contract
- public Lambda Function URL only
- no API Gateway resources
- stack output URL saved under data/local/output/

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 69: admin UI and Lambda frontend" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 69 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 69 PR" -Command {
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
