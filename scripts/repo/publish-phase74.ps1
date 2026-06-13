param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/74-runtime-doctor",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase74" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 74 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase74.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a single scripted runtime assurance check for local backend, browser viewer, Lambda Function URL, S3 summary, and cache status.

## Scope
- add `python ./devctl.py doctor`
- save runtime reports under `data/local/output`
- verify local dashboard, Browser Activity Viewer, browser activity API, device delivery provenance, Lambda health, S3 summary, and cache hit metrics
- add Phase 74 smoke, Newman, verify, publish, Postman, and docs

## Verification
- scripts/verify/verify-phase74.ps1
- scripts/local/smoke-phase74.ps1
- scripts/local/newman-phase74.ps1
- scripts/local/test-runtime-doctor.ps1
- python ./devctl.py doctor

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 74: runtime doctor assurance" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 74 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/backend-api.md",
        "docs/cloud-frontend.md",
        "docs/dashboard.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase74.postman_collection.json",
        "scripts/local/newman-phase74.ps1",
        "scripts/local/smoke-phase74.ps1",
        "scripts/local/test-runtime-doctor.ps1",
        "scripts/repo/publish-phase74.ps1",
        "scripts/verify/verify-phase74.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "test: add runtime doctor assurance", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 74"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add `python ./devctl.py doctor` runtime assurance reports
- verify local backend, dashboard controls, Browser Activity Viewer, delivery provenance, Lambda health, S3 summary, and cache hit metrics
- add Phase 74 smoke/Newman/verify/publish scripts and Postman collection
- document runtime doctor usage and output files

## Verification
- scripts/verify/verify-phase74.ps1
- scripts/local/smoke-phase74.ps1
- scripts/local/newman-phase74.ps1
- scripts/local/test-runtime-doctor.ps1
- python ./devctl.py doctor

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 74: runtime doctor assurance" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 74 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 74 PR" -Command {
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
