param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/72-cloud-s3-sample",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase72" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 72 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase72.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Make the deployed TraceDeck Lambda admin frontend prove real S3-backed data instead of an empty archive view.

## Scope
- parse real agent archive Metadata maps in the Lambda S3 summary reader
- upload a metadata-only browser activity sample archive to S3
- verify Lambda S3 summary rows, browser grouping, study-safe inference, non-study YouTube, and cache hit metrics
- add scripts, Newman collection, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase72.ps1
- scripts/local/smoke-phase72.ps1
- scripts/local/newman-phase72.ps1
- scripts/local/test-lambda-frontend-contract.ps1

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 72: cloud S3 sample and cache proof" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 72 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "devctl.py",
        "docs/cloud-frontend.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-cloud-phase72.postman_collection.json",
        "sam-app/frontend_function/app.py",
        "scripts/local/newman-phase72.ps1",
        "scripts/local/smoke-phase72.ps1",
        "scripts/local/test-lambda-frontend-contract.ps1",
        "scripts/local/upload-cloud-sample-phase72.ps1",
        "scripts/repo/publish-phase72.ps1",
        "scripts/verify/verify-phase72.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add cloud s3 sample proof", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 72"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- parse real agent archive Metadata maps inside the Lambda S3 reader
- add a Phase 72 S3 sample uploader that creates metadata-only JSONL gzip browser rows
- add live Lambda smoke and Newman checks for S3 rows, browser grouping, study-safe inference, non-study YouTube, and cache hit metrics
- add devctl Phase 72 hooks and docs

## Verification
- scripts/verify/verify-phase72.ps1
- scripts/local/smoke-phase72.ps1
- scripts/local/newman-phase72.ps1
- scripts/local/test-lambda-frontend-contract.ps1

## Deployment
- SAM Lambda frontend redeployed by the verifier
- stack output URL remains saved under data/local/output/

## Privacy
Metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 72: cloud S3 sample and cache proof" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 72 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 72 PR" -Command {
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
