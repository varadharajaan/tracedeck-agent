param(
    [string]$OutputPath = "data/local/output/phase-ledger.json",
    [string]$TextOutputPath = "data/local/output/phase-ledger.txt"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "get-phase-ledger" -LogRoot "logs/local/ops" | Out-Null

function Invoke-TraceDeckGitText {
    param([string[]]$GitArgs)
    $result = & git --no-pager @GitArgs 2>$null
    if ($LASTEXITCODE -ne 0) {
        return ""
    }
    return ($result -join "`n").Trim()
}

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

function ConvertTo-TraceDeckJsonNumber {
    param([AllowNull()][object]$Value)
    if ($null -eq $Value) {
        return "null"
    }
    return ([string]$Value)
}

try {
    Write-TraceDeckLog -Level "INFO" -Message "Collecting phase verifier files"
    $verifyRoot = Join-Path $script:TraceDeckRepoRoot "scripts/verify"
    $verifyFiles = Get-ChildItem -Path $verifyRoot -Filter "verify-phase*.ps1" -File
    $trackedPhaseValues = @(
        foreach ($file in $verifyFiles) {
            if ($file.BaseName -match "^verify-phase(?<phase>\d+)$") {
                [int]$Matches.phase
            }
        }
    )
    $trackedPhases = @($trackedPhaseValues | Sort-Object -Unique)

    Write-TraceDeckLog -Level "INFO" -Message "Collecting merged phase PR subjects"
    $mergeText = Invoke-TraceDeckGitText -GitArgs @("log", "--grep=Merge pull request #", "--format=%s", "-n", "250")
    $mergeSubjects = @($mergeText -split "`r?`n" | Where-Object { -not [string]::IsNullOrWhiteSpace($_) })
    $phaseMerges = @()
    foreach ($subject in $mergeSubjects) {
        if ($subject -match "Merge pull request #(?<pr>\d+) from .*/phase/(?<phase>\d+)-(?<slug>[A-Za-z0-9-]+)") {
            $phaseMerges += [pscustomobject]@{
                phase = [int]$Matches.phase
                pr = [int]$Matches.pr
                slug = $Matches.slug
                subject = $subject
            }
        }
    }

    $latestPhaseMerge = $null
    if ($phaseMerges.Count -gt 0) {
        $latestPhaseMerge = $phaseMerges[0]
    }

    Write-TraceDeckLog -Level "INFO" -Message "Collecting latest issue reference"
    $latestIssue = $null
    $latestRefsBody = Invoke-TraceDeckGitText -GitArgs @("log", "--grep=Refs #[0-9]", "--format=%B", "-n", "1")
    if ($latestRefsBody -match "Refs #(?<issue>\d+)") {
        $latestIssue = [int]$Matches.issue
    }

    $plannedNumberedPhases = @()
    $highestTrackedPhase = if ($trackedPhases.Count -gt 0) { [int]($trackedPhases | Select-Object -Last 1) } else { $null }
    $latestMergedPhase = if ($null -ne $latestPhaseMerge) { [int]$latestPhaseMerge.phase } else { $null }
    $latestMergedPr = if ($null -ne $latestPhaseMerge) { [int]$latestPhaseMerge.pr } else { $null }
    $remainingPlannedCount = @($plannedNumberedPhases).Count
    Write-TraceDeckLog -Level "INFO" -Message "Collecting git branch and head"
    $branch = Invoke-TraceDeckGitText -GitArgs @("branch", "--show-current")
    $head = Invoke-TraceDeckGitText -GitArgs @("rev-parse", "--short", "HEAD")

    Write-TraceDeckLog -Level "INFO" -Message "Composing ledger object"
    $latestMergeObject = $null
    if ($null -ne $latestPhaseMerge) {
        $latestMergeObject = [ordered]@{
            phase = $latestPhaseMerge.phase
            pr = $latestPhaseMerge.pr
            slug = $latestPhaseMerge.slug
            subject = $latestPhaseMerge.subject
        }
    }
    $recentMerges = @()
    foreach ($merge in @($phaseMerges | Select-Object -First 15)) {
        $recentMerges += [ordered]@{
            phase = $merge.phase
            pr = $merge.pr
            slug = $merge.slug
            subject = $merge.subject
        }
    }

    $generatedAt = (Get-Date).ToUniversalTime().ToString("o")
    $deniedPrivacy = @(
        "passwords",
        "screenshots",
        "raw_urls",
        "page_titles",
        "cookies",
        "tokens",
        "private_content",
        "endpoint_payloads",
        "provider_secrets",
        "alert_bodies",
        "keylogging",
        "hidden_collection_bypasses",
        "payment_data",
        "raw_provider_payloads"
    )

    $outputFullPath = Join-Path $script:TraceDeckRepoRoot $OutputPath
    $textFullPath = Join-Path $script:TraceDeckRepoRoot $TextOutputPath
    Write-TraceDeckLog -Level "INFO" -Message "Writing phase ledger outputs"
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $outputFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $textFullPath) | Out-Null

    $trackedPhaseJson = @($trackedPhases | ForEach-Object { [string]$_ }) -join ", "
    $plannedPhaseJson = @($plannedNumberedPhases | ForEach-Object { [string]$_ }) -join ", "
    $privacyJson = @($deniedPrivacy | ForEach-Object { ConvertTo-TraceDeckJsonString $_ }) -join ", "
    $recentMergeJsonRows = @(
        foreach ($merge in $recentMerges) {
            "    { ""phase"": $($merge.phase), ""pr"": $($merge.pr), ""slug"": $(ConvertTo-TraceDeckJsonString $merge.slug), ""subject"": $(ConvertTo-TraceDeckJsonString $merge.subject) }"
        }
    )
    $recentMergeJson = $recentMergeJsonRows -join ",`n"
    $latestMergeJson = "null"
    if ($null -ne $latestMergeObject) {
        $latestMergeJson = "{ ""phase"": $($latestMergeObject.phase), ""pr"": $($latestMergeObject.pr), ""slug"": $(ConvertTo-TraceDeckJsonString $latestMergeObject.slug), ""subject"": $(ConvertTo-TraceDeckJsonString $latestMergeObject.subject) }"
    }
    $jsonLines = @(
        "{",
        "  ""generated_at"": $(ConvertTo-TraceDeckJsonString $generatedAt),",
        "  ""source"": ""phase106_phase_ledger"",",
        "  ""evidence_scope"": ""metadata_only"",",
        "  ""repo_root"": $(ConvertTo-TraceDeckJsonString $script:TraceDeckRepoRoot),",
        "  ""git"": { ""branch"": $(ConvertTo-TraceDeckJsonString $branch), ""head"": $(ConvertTo-TraceDeckJsonString $head) },",
        "  ""counts"": {",
        "    ""highest_tracked_phase"": $(ConvertTo-TraceDeckJsonNumber $highestTrackedPhase),",
        "    ""tracked_phase_verify_scripts"": $(@($trackedPhases).Count),",
        "    ""latest_merged_phase"": $(ConvertTo-TraceDeckJsonNumber $latestMergedPhase),",
        "    ""merged_phase_pr_commits"": $(@($phaseMerges).Count),",
        "    ""latest_merged_pr"": $(ConvertTo-TraceDeckJsonNumber $latestMergedPr),",
        "    ""latest_issue_reference"": $(ConvertTo-TraceDeckJsonNumber $latestIssue),",
        "    ""remaining_planned_numbered_phases"": $remainingPlannedCount",
        "  },",
        "  ""answer"": {",
        "    ""remaining_planned_numbered_phases"": $remainingPlannedCount,",
        "    ""next_planned_numbered_phase"": null,",
        "    ""statement"": ""0 currently defined numbered phases remain.""",
        "  },",
        "  ""latest_merged_phase"": $latestMergeJson,",
        "  ""recent_merged_phase_prs"": [",
        $recentMergeJson,
        "  ],",
        "  ""tracked_phases"": [$trackedPhaseJson],",
        "  ""planned_numbered_phases"": [$plannedPhaseJson],",
        "  ""privacy"": {",
        "    ""note"": ""repository metadata only; sensitive collection denied"",",
        "    ""denied"": [$privacyJson]",
        "  }",
        "}"
    )
    $jsonLines | Set-Content -Path $outputFullPath -Encoding UTF8

    $lines = @(
        "TraceDeck Phase Ledger",
        "Generated: $generatedAt",
        "Highest tracked verifier phase: $highestTrackedPhase",
        "Tracked phase verifier scripts: $(@($trackedPhases).Count)",
        "Latest merged phase: $latestMergedPhase",
        "Latest merged PR: #$latestMergedPr",
        "Latest issue reference: #$latestIssue",
        "Merged phase PR commits: $(@($phaseMerges).Count)",
        "Remaining planned numbered phases: $remainingPlannedCount",
        "Next planned numbered phase: none",
        "Answer: 0 currently defined numbered phases remain.",
        "JSON: $OutputPath",
        "Text: $TextOutputPath",
        "Privacy: repository metadata only; sensitive collection denied"
    )
    $lines | Set-Content -Path $textFullPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Phase ledger saved json=$OutputPath text=$TextOutputPath highest_tracked_phase=$highestTrackedPhase latest_merged_phase=$latestMergedPhase remaining=$remainingPlannedCount"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
