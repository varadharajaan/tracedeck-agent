param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/80-lambda-cloud-admin-visual-parity",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase80" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 80 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase80.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Bring the public Lambda Cloud Admin frontend up to the same product-grade visual baseline as the local TraceDeck Console.

## Scope
- refresh the Lambda Cloud Admin shell with a symbolic brand mark, Workspace Source sidebar, hero/status area, full page labels, controls, cards, chips, and tables
- remove pseudo-letter controls from the Function URL UI
- add screenshot-free Lambda frontend visual-quality checks
- add Phase 80 smoke, Newman, verify, publish, Postman, docs, and devctl hooks
- deploy SAM and verify the public Function URL

## Verification
- scripts/verify/verify-phase80.ps1
- scripts/local/smoke-phase80.ps1
- scripts/local/newman-phase80.ps1
- scripts/local/test-lambda-frontend-visual.ps1
- python ./devctl.py cloud visual
- python ./devctl.py doctor

## Privacy
Metadata-only Lambda frontend rendering. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 80: Lambda Cloud Admin visual parity" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 80 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/cloud-frontend.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-cloud-phase80.postman_collection.json",
        "sam-app/frontend_function/app.py",
        "scripts/local/newman-phase80.ps1",
        "scripts/local/smoke-phase80.ps1",
        "scripts/local/test-lambda-frontend-contract.ps1",
        "scripts/local/test-lambda-frontend-visual.ps1",
        "scripts/repo/publish-phase80.ps1",
        "scripts/tools/lambda_frontend_visual_check.py",
        "scripts/verify/verify-phase80.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: refresh lambda cloud admin UI", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 80"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- refresh the Lambda Cloud Admin Function URL UI with the TraceDeck admin shell, symbolic brand mark, Workspace Source sidebar, and full page labels
- remove the pseudo-letter theme control and use Theme: Light/Dark labels without debug-looking shortcuts
- add screenshot-free Lambda frontend visual-quality checks
- add Phase 80 smoke/Newman/verify/publish scripts, Postman coverage, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase80.ps1
- scripts/local/smoke-phase80.ps1
- scripts/local/newman-phase80.ps1
- python ./devctl.py cloud visual
- python ./devctl.py doctor

## Privacy
Metadata-only Lambda frontend rendering. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 80: Lambda Cloud Admin visual parity" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 80 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 80 PR" -Command {
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
