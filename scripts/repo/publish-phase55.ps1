Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase55" -LogRoot "logs/local/repo" | Out-Null

$Owner = "varadharajaan"
$RepoName = "tracedeck-agent"
$Branch = "phase/55-provider-simulation-lab"
$IssueNumber = "117"

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Invoke-TraceDeckLoggedCommand -Label "Phase 55 verification" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase55.ps1
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
        "docs/privacy.md",
        "docs/roadmap.md",
        "docs/testing.md",
        "postman/tracedeck-backend-phase55.postman_collection.json",
        "scripts/local/newman-phase55.ps1",
        "scripts/local/smoke-phase55.ps1",
        "scripts/repo/publish-phase55.ps1",
        "scripts/tools/dashboard_layout_check.py",
        "scripts/verify/verify-phase55.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add provider simulation lab")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 55"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed tenant provider simulation lab API for metadata-only email, push, and dashboard dry-run proof
- add dashboard provider simulation section with route proof, scenarios, action queue, privacy proof, and command-nav entry
- add Phase 55 smoke, Newman, DOM, API, docs, and local verification coverage

## Verification
- scripts/verify/verify-phase55.ps1 passed
- scripts/local/smoke-phase55.ps1 live-booted the seeded dashboard, verified provider simulation UX markers, exercised GET/POST provider-simulation-lab APIs, checked audit proof, checked forbidden markers, and ran screenshot-free layout metrics
- scripts/local/newman-phase55.ps1 passed against a live dashboard demo
- cross-platform Windows/macOS/Linux builds passed locally
- root artifact clean check passed

## Privacy
- metadata-only provider simulation; no provider secrets, SMTP passwords, push endpoint payloads, alert bodies, screenshots, raw URLs, tokens, cookies, private content, or endpoint payloads are collected or stored
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 55: provider simulation lab" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 55 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 55 PR" -Command {
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
