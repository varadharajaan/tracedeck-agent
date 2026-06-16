param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/103-ready-pid-refresh",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase103" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 103 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase103.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 103 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 103 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a runnable remediation for stale ready-file PID proof so operators can refresh scheduled boot evidence without restarting a healthy backend.

## Scope
- add a metadata-only ready proof refresh script
- expose `python ./devctl.py server task-refresh-ready`
- update Runtime Status and Operator Assurance actions to point at the refresh command
- add Phase 103 smoke/Newman/verify/publish scripts and Postman coverage
- update docs and local verification notes

## Verification
- scripts/verify/verify-phase103.ps1
- scripts/local/test-refresh-backend-ready-proof.ps1
- scripts/local/smoke-phase103.ps1
- scripts/local/newman-phase103.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational runtime metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 103: ready PID refresh remediation" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 103 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/backend-api.md",
        "docs/dashboard.md",
        "docs/testing.md",
        "backend/internal/api/server.go",
        "backend/internal/api/server_test.go",
        "backend/internal/constants/constants.go",
        "postman/tracedeck-backend-phase103.postman_collection.json",
        "scripts/local/refresh-backend-ready-proof.ps1",
        "scripts/local/test-refresh-backend-ready-proof.ps1",
        "scripts/local/smoke-phase103.ps1",
        "scripts/local/newman-phase103.ps1",
        "scripts/verify/verify-phase103.ps1",
        "scripts/repo/publish-phase103.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add ready PID refresh remediation", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 103"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add `refresh-backend-ready-proof.ps1` to rewrite ready-file PID proof from healthy live PID and `/health`
- expose `python ./devctl.py server task-refresh-ready`
- point Runtime Status and Operator Assurance stale-ready-PID actions at the runnable refresh command
- add Phase 103 smoke/Newman/verify scripts, Postman coverage, and docs

## Verification
- scripts/verify/verify-phase103.ps1
- scripts/local/test-refresh-backend-ready-proof.ps1
- scripts/local/smoke-phase103.ps1
- scripts/local/newman-phase103.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
Operational runtime metadata only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 103: ready PID refresh remediation" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 103 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 103 PR" -Command {
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
