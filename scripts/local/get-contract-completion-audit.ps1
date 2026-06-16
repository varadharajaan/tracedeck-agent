param(
    [string]$OutputPath = "data/local/output/contract-completion-audit.json",
    [string]$TextOutputPath = "data/local/output/contract-completion-audit.txt"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-contract-completion-audit" -LogRoot "logs/local/ops" | Out-Null

function ConvertTo-TraceDeckJsonString {
    param([AllowNull()][object]$Value)
    if ($null -eq $Value) {
        return "null"
    }
    $escaped = [string]$Value
    $escaped = $escaped.Replace("\", "\\")
    $escaped = $escaped.Replace('"', '\"')
    $escaped = $escaped.Replace("`r", "\r")
    $escaped = $escaped.Replace("`n", "\n")
    $escaped = $escaped.Replace("`t", "\t")
    return '"' + $escaped + '"'
}

function ConvertTo-TraceDeckJsonArray {
    param([object[]]$Values)
    $items = @()
    foreach ($value in @($Values)) {
        $items += ConvertTo-TraceDeckJsonString $value
    }
    return "[" + ($items -join ", ") + "]"
}

function Test-TraceDeckPath {
    param([string]$RelativePath)
    return Test-Path -Path (Join-Path $script:TraceDeckRepoRoot $RelativePath)
}

function Test-TraceDeckAnyPath {
    param([string[]]$RelativePaths)
    foreach ($path in $RelativePaths) {
        if (Test-TraceDeckPath -RelativePath $path) {
            return $true
        }
    }
    return $false
}

function Test-TraceDeckText {
    param(
        [string]$RelativePath,
        [string]$Pattern
    )
    $path = Join-Path $script:TraceDeckRepoRoot $RelativePath
    if (!(Test-Path $path)) {
        return $false
    }
    return [bool](Select-String -Path $path -Pattern $Pattern -SimpleMatch -Quiet)
}

function Add-TraceDeckAuditRequirement {
    param(
        [System.Collections.ArrayList]$Requirements,
        [string]$ID,
        [string]$Area,
        [string]$Title,
        [string]$Status,
        [string[]]$Evidence,
        [string[]]$Gaps,
        [string]$NextAction
    )
    [void]$Requirements.Add([ordered]@{
        id = $ID
        area = $Area
        title = $Title
        status = $Status
        evidence = @($Evidence)
        gaps = @($Gaps)
        next_action = $NextAction
    })
}

function ConvertTo-TraceDeckRequirementJson {
    param([object]$Requirement)
    return "    { ""id"": $(ConvertTo-TraceDeckJsonString $Requirement.id), ""area"": $(ConvertTo-TraceDeckJsonString $Requirement.area), ""title"": $(ConvertTo-TraceDeckJsonString $Requirement.title), ""status"": $(ConvertTo-TraceDeckJsonString $Requirement.status), ""evidence"": $(ConvertTo-TraceDeckJsonArray $Requirement.evidence), ""gaps"": $(ConvertTo-TraceDeckJsonArray $Requirement.gaps), ""next_action"": $(ConvertTo-TraceDeckJsonString $Requirement.next_action) }"
}

try {
    $requirements = [System.Collections.ArrayList]::new()

    $noJavaSpring = !(Test-TraceDeckAnyPath -RelativePaths @("pom.xml", "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts"))
    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "go-first-core" `
        -Area "architecture" `
        -Title "Go-first agent and backend core" `
        -Status ($(if ((Test-TraceDeckPath "go.mod") -and (Test-TraceDeckPath "agent/cmd/tracedeck-agent/main.go") -and (Test-TraceDeckPath "backend/cmd/tracedeck-backend/main.go") -and $noJavaSpring) { "ok" } else { "missing" })) `
        -Evidence @("go.mod", "agent/cmd/tracedeck-agent/main.go", "backend/cmd/tracedeck-backend/main.go", "no Java/Spring project files detected") `
        -Gaps @() `
        -NextAction "Keep new endpoint and backend work in Go unless explicitly approved otherwise."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "privacy-deny-baseline" `
        -Area "privacy" `
        -Title "Sensitive capabilities are deny-only and documented" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/config/validate.go") -and (Test-TraceDeckPath "docs/privacy.md") -and (Test-TraceDeckPath "docs/collection-policy.md") -and (Test-TraceDeckText "docs/schema/policy-v1alpha1.schema.json" "screenshots")) { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/config/validate.go", "docs/privacy.md", "docs/collection-policy.md", "docs/schema/policy-v1alpha1.schema.json") `
        -Gaps @() `
        -NextAction "Continue adding privacy regression tests for every collector or policy expansion."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "typed-policy-schema" `
        -Area "configuration" `
        -Title "Typed YAML policy and generated schema foundation" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/config/types.go") -and (Test-TraceDeckPath "agent/internal/config/enums.go") -and (Test-TraceDeckPath "agent/internal/schema/policy.go") -and (Test-TraceDeckPath "examples/policies/ai-btech-student.yaml")) { "ok" } else { "missing" })) `
        -Evidence @("agent/internal/config/types.go", "agent/internal/config/enums.go", "agent/internal/schema/policy.go", "examples/policies/ai-btech-student.yaml") `
        -Gaps @() `
        -NextAction "Keep schema version changes code-driven and test-covered."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "centralized-constants" `
        -Area "architecture" `
        -Title "Centralized constants for shared names and schema ids" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/constants/project.go") -and (Test-TraceDeckPath "agent/internal/constants/policy.go") -and (Test-TraceDeckPath "backend/internal/constants/constants.go")) { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/constants/*.go", "backend/internal/constants/constants.go") `
        -Gaps @() `
        -NextAction "Avoid new duplicated literals in business logic and scripts."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "sqlite-local-buffer" `
        -Area "storage" `
        -Title "SQLite local buffer and migrations" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/storage/sqlite/store.go") -and (Test-TraceDeckPath "agent/internal/storage/sqlite/migrations/001_events.sql")) { "ok" } else { "missing" })) `
        -Evidence @("agent/internal/storage/sqlite/store.go", "agent/internal/storage/sqlite/migrations/001_events.sql", "agent/internal/storage/sqlite/store_test.go") `
        -Gaps @() `
        -NextAction "Keep retention and disk-bounds checks covered as storage evolves."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "s3-archive-outbox" `
        -Area "archive" `
        -Title "S3 archive writer/uploader foundation" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/archive/s3.go") -and (Test-TraceDeckPath "agent/internal/archive/writer.go") -and (Test-TraceDeckPath "docs/cloud-archive.md")) { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/archive/s3.go", "agent/internal/archive/writer.go", "docs/cloud-archive.md") `
        -Gaps @() `
        -NextAction "Re-run S3 live verification when cloud credentials and network access are intentionally available."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "email-alerting" `
        -Area "alerting" `
        -Title "Email alert evaluator/notifier foundation" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/alert/evaluator.go") -and (Test-TraceDeckPath "agent/internal/alert/email_notifier.go") -and (Test-TraceDeckPath "docs/alerting.md")) { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/alert/evaluator.go", "agent/internal/alert/email_notifier.go", "agent/internal/alert/*_test.go", "docs/alerting.md") `
        -Gaps @() `
        -NextAction "Keep dedupe, cooldown, and provider-backed delivery proof verified before claiming live email delivery."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "platform-adapters" `
        -Area "platform" `
        -Title "Windows, macOS, and Linux platform adapter skeletons" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/platform/current_windows.go") -and (Test-TraceDeckPath "agent/internal/platform/current_darwin.go") -and (Test-TraceDeckPath "agent/internal/platform/current_linux.go")) { "ok" } else { "missing" })) `
        -Evidence @("agent/internal/platform/current_windows.go", "agent/internal/platform/current_darwin.go", "agent/internal/platform/current_linux.go", "deployments/service/*") `
        -Gaps @() `
        -NextAction "Expand platform collectors behind the same adapter contracts."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "service-autostart" `
        -Area "service" `
        -Title "Native service/autostart manifests and scripts" `
        -Status ($(if ((Test-TraceDeckPath "deployments/service/windows/tracedeck-agent-task.xml.tmpl") -and (Test-TraceDeckPath "deployments/service/darwin/io.tracedeck.agent.plist.tmpl") -and (Test-TraceDeckPath "deployments/service/linux/tracedeck-agent.service.tmpl") -and (Test-TraceDeckPath "scripts/local/manage-agent-service.ps1")) { "ok" } else { "attention" })) `
        -Evidence @("deployments/service/windows/tracedeck-agent-task.xml.tmpl", "deployments/service/darwin/io.tracedeck.agent.plist.tmpl", "deployments/service/linux/tracedeck-agent.service.tmpl", "scripts/local/manage-agent-service.ps1") `
        -Gaps @() `
        -NextAction "Run elevated/native registration only when the operator intentionally approves UAC/admin actions."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "backend-dashboard" `
        -Area "backend" `
        -Title "Local backend, dashboard, browser viewer, and API contracts" `
        -Status ($(if ((Test-TraceDeckPath "backend/internal/api/server.go") -and (Test-TraceDeckPath "backend/internal/api/web/dashboard.html") -and (Test-TraceDeckPath "backend/internal/api/web/browser_activity.html") -and (Test-TraceDeckPath "docs/backend-api.md")) { "ok" } else { "missing" })) `
        -Evidence @("backend/internal/api/server.go", "backend/internal/api/web/dashboard.html", "backend/internal/api/web/browser_activity.html", "docs/backend-api.md", "postman/*.json") `
        -Gaps @() `
        -NextAction "Continue using smoke/Newman/dashboard contract checks for every API or UI change."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "browser-domain-collector" `
        -Area "browser" `
        -Title "Browser domain/category metadata collector" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/collector/browser/collector.go") -and (Test-TraceDeckPath "agent/internal/collector/browser/collector_test.go") -and (Test-TraceDeckPath "docs/privacy.md")) { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/collector/browser/collector.go", "agent/internal/collector/browser/collector_test.go", "docs/privacy.md") `
        -Gaps @() `
        -NextAction "Keep default browser evidence domain/category-only and add extension proof before claiming active-tab support."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "browser-extension-skeleton" `
        -Area "browser" `
        -Title "Chrome, Edge, and Brave browser extension skeleton" `
        -Status ($(if (Test-TraceDeckPath "browser-extension") { "ok" } else { "missing" })) `
        -Evidence @() `
        -Gaps @("browser-extension/ directory is not present") `
        -NextAction "Add TypeScript browser extension skeletons that send domain/category metadata to localhost only."

    $foregroundCollectorPresent = (Test-TraceDeckPath "agent/internal/collector/activewindow/collector.go") -and
        (Test-TraceDeckPath "agent/internal/collector/activewindow/collector_test.go") -and
        (Test-TraceDeckPath "scripts/local/test-active-window-collector.ps1")

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "foreground-app-collector" `
        -Area "collector" `
        -Title "Active foreground app/window collection" `
        -Status ($(if ($foregroundCollectorPresent) { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/collector/activewindow/collector.go", "agent/internal/collector/activewindow/collector_test.go", "scripts/local/test-active-window-collector.ps1", "agent/internal/platform/support.go declares foreground app capability", "docs/platform-support.md documents permission differences") `
        -Gaps ($(if ($foregroundCollectorPresent) { @() } else { @("No full active-window collector package is present") })) `
        -NextAction ($(if ($foregroundCollectorPresent) { "Keep foreground app collection metadata-only and expand macOS/Linux native adapters behind the same contract." } else { "Implement foreground app collector behind platform adapters before claiming active-window MVP completion." }))

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "software-install-collector" `
        -Area "collector" `
        -Title "Software install/uninstall event collection" `
        -Status ($(if (Test-TraceDeckPath "agent/internal/collector/software") { "ok" } else { "attention" })) `
        -Evidence @("agent/internal/software/classifier.go", "docs/risky-software-detection.md") `
        -Gaps @("Classifier exists, but no full OS install-event collector package is present") `
        -NextAction "Add OS-specific install/uninstall collectors behind platform adapters."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "opentelemetry-exporter" `
        -Area "telemetry" `
        -Title "OpenTelemetry OTLP exporter implementation" `
        -Status ($(if ((Test-TraceDeckPath "agent/internal/exporter/otlp.go") -and (Test-TraceDeckPath "agent/internal/exporter/otlp_test.go")) { "ok" } else { "missing" })) `
        -Evidence @("agent/internal/exporter/otlp.go", "agent/internal/exporter/otlp_test.go", "docs/opentelemetry-exporter.md", "docs/telemetry-schema.md") `
        -Gaps ($(if ((Test-TraceDeckPath "agent/internal/exporter/otlp.go") -and (Test-TraceDeckPath "agent/internal/exporter/otlp_test.go")) { @() } else { @("Telemetry schema docs exist, but no agent/internal/exporter package is present") })) `
        -NextAction "Keep OTLP export bounded, metadata-only, and covered by smoke tests before adding additional signal types."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "visible-local-indicator" `
        -Area "transparency" `
        -Title "Visible local monitoring indicator" `
        -Status ($(if (Test-TraceDeckText "agent/internal/platform/support.go" "PlatformCapabilityLocalIndicator") { "attention" } else { "missing" })) `
        -Evidence @("agent/internal/platform/support.go", "docs/collection-policy.md") `
        -Gaps @("Capability is planned/documented, but no platform UI indicator implementation is present") `
        -NextAction "Implement a visible local indicator before expanding interactive monitoring."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "docker-compose-otel-stack" `
        -Area "local-dev" `
        -Title "Docker Compose and OpenTelemetry Collector local stack" `
        -Status ($(if ((Test-TraceDeckPath "deployments/otel/docker-compose.yaml") -and (Test-TraceDeckPath "deployments/otel/otel-collector.yaml")) { "ok" } else { "missing" })) `
        -Evidence @("deployments/otel/docker-compose.yaml", "deployments/otel/otel-collector.yaml", "scripts/local/test-otel-exporter.ps1") `
        -Gaps ($(if ((Test-TraceDeckPath "deployments/otel/docker-compose.yaml") -and (Test-TraceDeckPath "deployments/otel/otel-collector.yaml")) { @() } else { @("No Docker Compose or OTel Collector config was found in deployments/") })) `
        -NextAction "Use scripts/local/test-otel-exporter.ps1 and docker compose config before local collector demos."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "quality-and-security-gates" `
        -Area "verification" `
        -Title "Scripted Go quality and security gates" `
        -Status ($(if ((Test-TraceDeckPath "scripts/local/test-go-quality-gates.ps1") -and (Test-TraceDeckPath ".golangci.yml") -and (Test-TraceDeckPath "scripts/verify/verify-postmerge.ps1")) { "ok" } else { "attention" })) `
        -Evidence @("scripts/local/test-go-quality-gates.ps1", ".golangci.yml", "scripts/verify/verify-postmerge.ps1", "scripts/setup/install-go-tools.ps1") `
        -Gaps @() `
        -NextAction "Run full security gates when required tools are installed and phase scope changes Go behavior."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "release-sbom-packaging" `
        -Area "release" `
        -Title "GoReleaser and Syft SBOM release packaging" `
        -Status ($(if ((Test-TraceDeckAnyPath -RelativePaths @(".goreleaser.yml", ".goreleaser.yaml")) -and (Test-TraceDeckAnyPath -RelativePaths @("sbom", "dist/sbom"))) { "ok" } else { "missing" })) `
        -Evidence @() `
        -Gaps @("No GoReleaser config or Syft SBOM release flow found") `
        -NextAction "Add release packaging and SBOM generation before production distribution."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "cloud-admin-lambda" `
        -Area "cloud" `
        -Title "SAM Lambda Function URL admin frontend" `
        -Status ($(if ((Test-TraceDeckPath "sam-app/template.yaml") -and (Test-TraceDeckPath "sam-app/frontend_function/app.py") -and (Test-TraceDeckPath "docs/cloud-frontend.md")) { "ok" } else { "attention" })) `
        -Evidence @("sam-app/template.yaml", "sam-app/frontend_function/app.py", "docs/cloud-frontend.md", "python ./devctl.py sam ...") `
        -Gaps @() `
        -NextAction "Refresh SAM deployment only when cloud verification is intentionally requested."

    Add-TraceDeckAuditRequirement -Requirements $requirements `
        -ID "phase-ledger-and-promotion-proof" `
        -Area "operations" `
        -Title "Operator ledger, evidence, assurance, and promotion proof" `
        -Status ($(if ((Test-TraceDeckPath "docs/phase-ledger.md") -and (Test-TraceDeckPath "scripts/local/get-phase-ledger.ps1") -and (Test-TraceDeckPath "scripts/local/get-promotion-readiness.ps1")) { "ok" } else { "attention" })) `
        -Evidence @("docs/phase-ledger.md", "scripts/local/get-phase-ledger.ps1", "scripts/local/get-verification-evidence.ps1", "scripts/local/get-operator-assurance.ps1", "scripts/local/get-promotion-readiness.ps1") `
        -Gaps @() `
        -NextAction "Use these proof bundles before demos, PRs, and handoffs."

    $okCount = @($requirements | Where-Object { $_.status -eq "ok" }).Count
    $attentionCount = @($requirements | Where-Object { $_.status -eq "attention" }).Count
    $missingCount = @($requirements | Where-Object { $_.status -eq "missing" }).Count
    $overallStatus = if ($missingCount -gt 0 -or $attentionCount -gt 0) { "attention" } else { "ok" }
    $generatedAt = (Get-Date).ToUniversalTime().ToString("o")

    $outputFullPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textFullPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $textFullPath) | Out-Null

    $requirementJson = @()
    foreach ($requirement in $requirements) {
        $requirementJson += ConvertTo-TraceDeckRequirementJson -Requirement $requirement
    }
    $jsonLines = @(
        "{",
        "  ""generated_at"": $(ConvertTo-TraceDeckJsonString $generatedAt),",
        "  ""source"": ""phase107_contract_completion_audit"",",
        "  ""evidence_scope"": ""metadata_only"",",
        "  ""overall_status"": $(ConvertTo-TraceDeckJsonString $overallStatus),",
        "  ""summary"": { ""ok"": $okCount, ""attention"": $attentionCount, ""missing"": $missingCount, ""total"": $($requirements.Count) },",
        "  ""requirements"": [",
        ($requirementJson -join ",`n"),
        "  ],",
        "  ""privacy"": { ""note"": ""repository metadata only; sensitive collection denied"" }",
        "}"
    )
    $jsonLines | Set-Content -Path $outputFullPath -Encoding UTF8

    $lines = @(
        "TraceDeck Contract Completion Audit",
        "Generated: $generatedAt",
        "Status: $overallStatus",
        "OK: $okCount",
        "Attention: $attentionCount",
        "Missing: $missingCount",
        "Total: $($requirements.Count)",
        "JSON: $OutputPath",
        "Text: $TextOutputPath",
        "",
        "Direct answer:",
        "TraceDeck is not end-to-end complete yet; this audit found $attentionCount attention item(s) and $missingCount missing item(s).",
        "",
        "Findings:"
    )
    foreach ($requirement in $requirements) {
        $lines += "- $($requirement.status): $($requirement.id) - $($requirement.title)"
        if (@($requirement.gaps).Count -gt 0) {
            foreach ($gap in @($requirement.gaps)) {
                $lines += "  gap: $gap"
            }
        }
        $lines += "  next: $($requirement.next_action)"
    }
    $lines += ""
    $lines += "Privacy: repository metadata only; sensitive collection denied"
    $lines | Set-Content -Path $textFullPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Contract audit saved json=$OutputPath text=$TextOutputPath status=$overallStatus ok=$okCount attention=$attentionCount missing=$missingCount"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
