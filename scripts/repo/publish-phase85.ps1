param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/85-go-quality-gates",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase85" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 85 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase85.ps1
    }

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a reusable strict Go quality gate for TraceDeck phases so the engineering contract is verified through scripts instead of tribal memory.

## Scope
- add `scripts/local/test-go-quality-gates.ps1`
- run gofmt, `go test ./...`, `go test -race ./...`, `go vet ./...`, `golangci-lint run ./...`, `govulncheck ./...`, and `gosec ./...`
- write quality reports under `data/local/go-quality/...` and logs under `logs/local/test`
- add Phase 85 verifier, Newman runtime guard, Postman collection, docs, and devctl hooks

## Verification
- scripts/verify/verify-phase85.ps1
- scripts/local/test-go-quality-gates.ps1
- scripts/local/newman-phase85.ps1
- python ./devctl.py doctor --no-cloud-refresh through verify-phase85

## Privacy
Quality reports and runtime metadata checks only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 85: add strict Go quality gates" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 85 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "devctl.py",
        "docs/roadmap.md",
        "docs/security.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase85.postman_collection.json",
        "scripts/local/newman-phase85.ps1",
        "scripts/local/test-go-quality-gates.ps1",
        "scripts/repo/publish-phase85.ps1",
        "scripts/verify/verify-phase85.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "build: add strict go quality gates", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 85"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add a reusable Phase 85 strict Go quality gate script
- verify gofmt, all Go tests, race tests, go vet, golangci-lint, govulncheck, and gosec through the local script
- add a Phase 85 runtime Newman guard, verifier, Postman collection, docs, and devctl aliases

## Verification
- scripts/verify/verify-phase85.ps1
- scripts/local/test-go-quality-gates.ps1
- scripts/local/newman-phase85.ps1

## Privacy
Quality reports and runtime metadata checks only. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, keylogging, hidden collection bypasses, provider secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 85: add strict Go quality gates" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 85 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 85 PR" -Command {
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
