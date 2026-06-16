param(
    [string]$Phase = "phase99",
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$OutputPath = "data/local/output/verification-evidence.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-verification-evidence" -LogRoot "logs/local/ops" | Out-Null

function ConvertTo-TraceDeckRelativePath {
    param([string]$PathValue)
    if ([string]::IsNullOrWhiteSpace($PathValue)) {
        return ""
    }
    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        $full = [System.IO.Path]::GetFullPath($PathValue)
    }
    else {
        $full = [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
    }
    $root = [System.IO.Path]::GetFullPath($script:TraceDeckRepoRoot)
    if ($full.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
        return $full.Substring($root.Length).TrimStart("\", "/").Replace("\", "/")
    }
    return $PathValue.Replace("\", "/")
}

function Get-TraceDeckGitValue {
    param([string[]]$GitArgs)
    $excludeFile = Join-Path $script:TraceDeckRepoRoot ".git/info/exclude"
    try {
        $value = (& git -c "core.excludesFile=$excludeFile" @GitArgs 2>$null | Out-String).Trim()
        return $value
    }
    catch {
        return ""
    }
}

function Get-LatestTraceDeckFile {
    param(
        [string]$RelativeDir,
        [string]$Pattern
    )
    $dir = Join-Path $script:TraceDeckRepoRoot $RelativeDir
    if (-not (Test-Path $dir)) {
        return $null
    }
    return Get-ChildItem -Path $dir -Filter $Pattern -File |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
}

function Get-LatestTraceDeckNestedFile {
    param(
        [string]$RelativeDir,
        [string]$Pattern
    )
    $dir = Join-Path $script:TraceDeckRepoRoot $RelativeDir
    if (-not (Test-Path $dir)) {
        return $null
    }
    return Get-ChildItem -Path $dir -Filter $Pattern -File -Recurse |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
}

function Get-GateStatus {
    param([System.IO.FileInfo]$LogFile)
    if (-not $LogFile) {
        return "pending"
    }
    $tail = Get-Content -Path $LogFile.FullName -Tail 220 -ErrorAction SilentlyContinue
    $joined = ($tail -join "`n").ToLowerInvariant()
    if ($joined.Contains("[error]") -or $joined.Contains("failed with exit code")) {
        return "attention"
    }
    return "ok"
}

function New-GateEvidence {
    param(
        [string]$ID,
        [string]$Label,
        [string]$Command,
        [System.IO.FileInfo]$LogFile,
        [System.IO.FileInfo]$ReportFile,
        [string]$Detail
    )
    $status = Get-GateStatus -LogFile $LogFile
    $severity = if ($status -eq "ok") { "info" } elseif ($status -eq "pending") { "medium" } else { "high" }
    $completedAt = ""
    if ($LogFile) {
        $completedAt = $LogFile.LastWriteTime.ToString("o")
    }
    [ordered]@{
        id             = $ID
        label          = $Label
        command        = $Command
        status         = $status
        severity       = $severity
        log_path       = if ($LogFile) { ConvertTo-TraceDeckRelativePath -PathValue $LogFile.FullName } else { "" }
        report_path    = if ($ReportFile) { ConvertTo-TraceDeckRelativePath -PathValue $ReportFile.FullName } else { "" }
        detail         = $Detail
        completed_at   = $completedAt
        evidence_scope = "metadata_only"
    }
}

function New-ArtifactEvidence {
    param(
        [string]$ID,
        [string]$Label,
        [string]$RelativePath
    )
    $full = Join-Path $script:TraceDeckRepoRoot $RelativePath
    [ordered]@{
        id             = $ID
        label          = $Label
        path           = $RelativePath.Replace("\", "/")
        status         = if (Test-Path $full) { "ok" } else { "pending" }
        evidence_scope = "metadata_only"
    }
}

try {
    $phaseName = $Phase.Trim()
    if ([string]::IsNullOrWhiteSpace($phaseName)) {
        $phaseName = "phase99"
    }

    $newmanReport = Get-LatestTraceDeckNestedFile -RelativeDir "data/local/newman/$phaseName" -Pattern "newman-report.json"
    $gates = @(
        New-GateEvidence -ID "gofmt" -Label "Go format check" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/verify" -Pattern "check-gofmt-*.log") -ReportFile $null -Detail "Go source formatting gate."
        New-GateEvidence -ID "backend-api" -Label "Backend API tests" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-backend-api.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/test" -Pattern "test-backend-api-*.log") -ReportFile $null -Detail "Backend API and store regression tests."
        New-GateEvidence -ID "dashboard-contract" -Label "Dashboard DOM contract" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-contract.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/test" -Pattern "test-dashboard-contract-*.log") -ReportFile $null -Detail "Dashboard DOM, navigator, and forbidden marker contract."
        New-GateEvidence -ID "dashboard-js" -Label "Dashboard JavaScript syntax" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-dashboard-js.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/test" -Pattern "test-dashboard-js-*.log") -ReportFile $null -Detail "Dashboard JavaScript syntax extraction."
        New-GateEvidence -ID "runtime-summary" -Label "Runtime summary proof" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-runtime-summary.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/test" -Pattern "test-runtime-summary-*.log") -ReportFile $null -Detail "Runtime summary artifact and privacy proof."
        New-GateEvidence -ID "phase-smoke" -Label "$phaseName smoke" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-$phaseName.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/smoke" -Pattern "smoke-$phaseName-*.log") -ReportFile $null -Detail "Live isolated backend smoke coverage."
        New-GateEvidence -ID "phase-newman" -Label "$phaseName Newman" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-$phaseName.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/newman" -Pattern "newman-$phaseName-*.log") -ReportFile $newmanReport -Detail "Postman/Newman API contract coverage."
        New-GateEvidence -ID "phase-verify" -Label "$phaseName verifier" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-$phaseName.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/verify" -Pattern "verify-$phaseName-*.log") -ReportFile $null -Detail "Full scripted phase verifier."
        New-GateEvidence -ID "root-clean" -Label "Root artifact check" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/verify" -Pattern "check-root-clean-*.log") -ReportFile $null -Detail "Repo root artifact hygiene."
        New-GateEvidence -ID "live-provenance" -Label "Live server provenance" -Command "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl $BaseUrl" -LogFile (Get-LatestTraceDeckFile -RelativeDir "logs/local/test" -Pattern "test-live-server-provenance-*.log") -ReportFile $null -Detail "Live local server provenance and demo-boundary proof."
    )

    $attentionCount = @($gates | Where-Object { $_.status -eq "attention" }).Count
    $pendingCount = @($gates | Where-Object { $_.status -eq "pending" }).Count
    $okCount = @($gates | Where-Object { $_.status -eq "ok" }).Count
    $overallStatus = if ($attentionCount -gt 0) { "attention" } elseif ($pendingCount -gt 0) { "watch" } else { "ok" }
    $canPromote = $overallStatus -eq "ok"

    $artifacts = @(
        New-ArtifactEvidence -ID "runtime-summary-json" -Label "Runtime summary JSON" -RelativePath "data/local/output/runtime-summary.json"
        New-ArtifactEvidence -ID "runtime-summary-text" -Label "Runtime summary text" -RelativePath "data/local/output/runtime-summary.txt"
        New-ArtifactEvidence -ID "runtime-doctor-json" -Label "Runtime doctor JSON" -RelativePath "data/local/output/runtime-doctor.json"
        New-ArtifactEvidence -ID "frontend-url" -Label "Lambda Function URL output" -RelativePath "data/local/output/frontend-url.txt"
    )
    if ($newmanReport) {
        $artifacts += [ordered]@{
            id             = "newman-report"
            label          = "$phaseName Newman report"
            path           = ConvertTo-TraceDeckRelativePath -PathValue $newmanReport.FullName
            status         = "ok"
            evidence_scope = "metadata_only"
        }
    }

    $actions = @()
    if (-not $canPromote) {
        $actions += [ordered]@{
            id       = "run-phase-verifier"
            title    = "Run phase verifier"
            detail   = "Rerun the phase verifier, then regenerate verification evidence."
            command  = "python ./devctl.py test verify99"
            severity = "high"
            status   = "pending"
        }
    }
    $actions += [ordered]@{
        id       = "refresh-verification-evidence"
        title    = "Refresh verification evidence"
        detail   = "Regenerate the metadata-only evidence artifact after any gate run."
        command  = "powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-verification-evidence.ps1"
        severity = "info"
        status   = if ($canPromote) { "ok" } else { "watch" }
    }

    $artifact = [ordered]@{
        generated_at      = (Get-Date).ToString("o")
        phase             = $phaseName
        base_url          = $BaseUrl
        branch            = Get-TraceDeckGitValue -GitArgs @("branch", "--show-current")
        head              = Get-TraceDeckGitValue -GitArgs @("rev-parse", "--short", "HEAD")
        overall_status    = $overallStatus
        can_promote       = $canPromote
        gates             = $gates
        artifacts         = $artifacts
        actions           = $actions
        privacy           = [ordered]@{
            metadata_only        = $true
            sensitive_collection = "denied"
            forbidden_categories = @("credentials", "provider secrets", "screenshots", "raw web content", "payment data", "keylogging")
        }
        privacy_boundary  = "metadata-only verification evidence: gate labels, statuses, commands, timestamps, log paths, report paths, git branch/head labels, and operator actions only"
    }

    $outputFullPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $outputDir = Split-Path -Parent $outputFullPath
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null
    $artifact | ConvertTo-Json -Depth 16 | Set-Content -Path $outputFullPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Verification evidence saved path=$OutputPath phase=$phaseName status=$overallStatus ok=$okCount pending=$pendingCount attention=$attentionCount"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
