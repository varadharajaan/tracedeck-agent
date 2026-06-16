param(
    [string]$Owner = "varadharajaan",
    [string]$RepoName = "tracedeck-agent",
    [string]$Branch = "phase/109-opentelemetry-exporter-stack",
    [string]$IssueNumber = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "publish-phase109" -LogRoot "logs/local/repo" | Out-Null

function Invoke-Git {
    param([Alias("Args")][string[]]$GitArgs)
    Invoke-TraceDeckLoggedCommand -Label "git $($GitArgs -join ' ')" -Command {
        git @GitArgs
    }
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Phase 109 verification"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase109.ps1
    if ($LASTEXITCODE -ne 0) {
        throw "Phase 109 verification failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Phase 109 verification"

    if ([string]::IsNullOrWhiteSpace($IssueNumber)) {
        $issueBody = @"
## Goal
Add a metadata-only OpenTelemetry OTLP exporter and local OpenTelemetry Collector stack for TraceDeck Agent.

## Scope
- add typed `observability.opentelemetry` policy config
- add OTLP/HTTP JSON log exporter with bounded attempts and drop metrics
- export from the persisted local event stream without raw URLs, page titles, credentials, screenshots, cookies, tokens, or private content
- add local fake OTLP receiver smoke coverage
- add Docker Compose and OpenTelemetry Collector config under deployments/otel
- add Newman, docs, schema, audit, ledger, and devctl coverage

## Verification
- scripts/local/test-otel-exporter.ps1
- scripts/local/smoke-phase109.ps1
- scripts/local/newman-phase109.ps1
- scripts/verify/verify-phase109.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
The exporter emits metadata-only OTLP logs. No passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payload bodies, provider secrets, alert bodies, keylogging, hidden collection bypasses, payment data, or raw provider payloads are stored or transmitted.
"@
        $issueURL = (gh issue create --repo "$Owner/$RepoName" --title "Phase 109: OpenTelemetry exporter and local collector stack" --body $issueBody).Trim()
        $IssueNumber = ($issueURL -split "/")[-1]
        Write-TraceDeckLog -Level "INFO" -Message "Created Phase 109 issue #${IssueNumber}: $issueURL"
    }

    Invoke-Git -Args @("checkout", "-B", $Branch)

    $files = @(
        "README.md",
        "agent/internal/app/run.go",
        "agent/internal/app/run_test.go",
        "agent/internal/config/config_test.go",
        "agent/internal/config/enums.go",
        "agent/internal/config/schema_enums.go",
        "agent/internal/config/types.go",
        "agent/internal/config/validate.go",
        "agent/internal/constants/config_fields.go",
        "agent/internal/constants/events.go",
        "agent/internal/constants/project.go",
        "agent/internal/constants/schema.go",
        "agent/internal/exporter/otlp.go",
        "agent/internal/exporter/otlp_test.go",
        "deployments/otel/docker-compose.yaml",
        "deployments/otel/otel-collector.yaml",
        "devctl.py",
        "docs/contract-completion-audit.md",
        "docs/opentelemetry-exporter.md",
        "docs/phase-ledger.md",
        "docs/roadmap.md",
        "docs/schema/policy-v1alpha1.schema.json",
        "docs/telemetry-schema.md",
        "docs/testing.md",
        "examples/policies/ai-btech-student.yaml",
        "postman/tracedeck-backend-phase109.postman_collection.json",
        "scripts/local/generate-policy-schema.ps1",
        "scripts/local/get-contract-completion-audit.ps1",
        "scripts/local/newman-phase109.ps1",
        "scripts/local/smoke-phase109.ps1",
        "scripts/local/test-otel-exporter.ps1",
        "scripts/repo/publish-phase109.ps1",
        "scripts/tools/fake-otlp/main.go",
        "scripts/verify/verify-phase109.ps1"
    )
    Invoke-Git -Args (@("add", "--") + $files)

    $staged = git diff --cached --name-only
    if (![string]::IsNullOrWhiteSpace($staged)) {
        Invoke-Git -Args @("commit", "-m", "feat: add opentelemetry exporter stack", "-m", "Refs #$IssueNumber")
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "No staged changes to commit for Phase 109"
    }

    Invoke-Git -Args @("push", "-u", "origin", $Branch, "--force-with-lease")

    $existingPR = gh pr list --repo "$Owner/$RepoName" --head $Branch --state open --json url --jq ".[0].url" 2>$null
    $body = @"
Closes #$IssueNumber

## Summary
- add typed OpenTelemetry policy config and schema updates
- add metadata-only OTLP/HTTP JSON log exporter with bounded retry/drop metrics
- wire OTel export into the agent run cycle from stored events
- add fake OTLP live smoke coverage and Phase 109 Newman regression
- add Docker Compose and OpenTelemetry Collector config under deployments/otel
- update docs, audit, ledger, roadmap, and devctl aliases

## Verification
- scripts/local/test-otel-exporter.ps1
- scripts/local/smoke-phase109.ps1
- scripts/local/newman-phase109.ps1
- scripts/verify/verify-phase109.ps1
- scripts/verify/check-root-clean.ps1

## Privacy
OTLP export is metadata-only. Sensitive markers are filtered before export and the smoke verifies no password, screenshot, raw URL, page title, cookie, token, provider secret, alert body, private content, payment data, or raw provider payload markers are present.
"@
    if ([string]::IsNullOrWhiteSpace($existingPR)) {
        $prURL = (gh pr create --repo "$Owner/$RepoName" --base main --head $Branch --title "Phase 109: OpenTelemetry exporter and local collector stack" --body $body).Trim()
    }
    else {
        $prURL = $existingPR.Trim()
    }
    Write-TraceDeckLog -Level "INFO" -Message "Phase 109 PR: $prURL"

    Invoke-TraceDeckLoggedCommand -Label "Merge Phase 109 PR" -Command {
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
