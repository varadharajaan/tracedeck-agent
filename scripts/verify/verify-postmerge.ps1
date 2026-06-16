param(
    [string]$PhaseTarget = "phase95",
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$IssueNumber = "",
    [string]$PrNumber = "",
    [switch]$AllowContentDiff,
    [switch]$SkipGitHub
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-postmerge" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Current phase verification target=$PhaseTarget" -Command {
        python ./devctl.py test $PhaseTarget
    }

    Invoke-TraceDeckLoggedCommand -Label "Backend task status" -Command {
        python ./devctl.py server task-status
    }

    Invoke-TraceDeckLoggedCommand -Label "Runtime doctor local" -Command {
        python ./devctl.py doctor --skip-cloud
    }

    Invoke-TraceDeckLoggedCommand -Label "Live server provenance" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl $BaseUrl
    }

    Invoke-TraceDeckLoggedCommand -Label "Root artifact check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "git diff --check" -Command {
        git diff --check
    }

    $contentDiff = git diff --name-status
    if (![string]::IsNullOrWhiteSpace($contentDiff)) {
        $level = if ($AllowContentDiff) { "WARN" } else { "ERROR" }
        Write-TraceDeckLog -Level $level -Message "Tracked content diff remains after post-merge verification:"
        foreach ($line in $contentDiff) {
            Write-TraceDeckLog -Level $level -Message $line
        }
        if (-not $AllowContentDiff) {
            throw "Tracked content diff remains after post-merge verification."
        }
    }
    else {
        Write-TraceDeckLog -Level "INFO" -Message "No tracked content diff remains."
    }

    $status = git status --short --branch
    foreach ($line in $status) {
        Write-TraceDeckLog -Level "INFO" -Message "git status: $line"
    }

    if (-not $SkipGitHub) {
        if ([string]::IsNullOrWhiteSpace($IssueNumber) -or [string]::IsNullOrWhiteSpace($PrNumber)) {
            throw "IssueNumber and PrNumber are required unless -SkipGitHub is set."
        }

        $issue = gh issue view $IssueNumber --repo varadharajaan/tracedeck-agent --json number,state,url | ConvertFrom-Json
        if ($issue.state -ne "CLOSED") {
            throw "Expected issue #$IssueNumber to be CLOSED, got $($issue.state)."
        }
        Write-TraceDeckLog -Level "INFO" -Message "GitHub issue #$IssueNumber is CLOSED: $($issue.url)"

        $pr = gh pr view $PrNumber --repo varadharajaan/tracedeck-agent --json number,state,mergedAt,mergeCommit,url | ConvertFrom-Json
        if ($pr.state -ne "MERGED" -or [string]::IsNullOrWhiteSpace($pr.mergedAt)) {
            throw "Expected PR #$PrNumber to be MERGED, got state=$($pr.state) merged_at=$($pr.mergedAt)."
        }
        Write-TraceDeckLog -Level "INFO" -Message "GitHub PR #$PrNumber is MERGED at $($pr.mergedAt): $($pr.url)"
    }
    else {
        Write-TraceDeckLog -Level "INFO" -Message "Skipped GitHub issue/PR state checks."
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
