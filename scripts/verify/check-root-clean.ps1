param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "check-root-clean" -LogRoot "logs/local/verify" | Out-Null

try {
    $rootFiles = Get-ChildItem -Path $script:TraceDeckRepoRoot -File -Force
    $forbidden = $rootFiles | Where-Object {
        $_.Name -eq "coverage.out" -or
        $_.Name -like "*.log" -or
        $_.Name -like "*.sqlite" -or
        $_.Name -like "*.db" -or
        $_.Name -like "*.prof" -or
        $_.Name -like "*.test" -or
        $_.Name -like "*.exe"
    }

    if ($forbidden) {
        $names = ($forbidden | Select-Object -ExpandProperty Name) -join ", "
        Write-TraceDeckLog -Level "ERROR" -Message "Root contains generated artifacts: $names"
        exit 1
    }

    $localOnlyFiles = @(
        "plan.md",
        "checkpoint.md",
        "memory.md",
        "todo.md",
        "PLAN_MODE_PROMPT.md",
        "agent-strict-rules-prompts.md",
        "instructions.md",
        "tracedeck_agent_project_prompt.md",
        "tracedeck-agent-system-blueprint.md"
    )

    $trackedLocalOnly = @()
    foreach ($fileName in $localOnlyFiles) {
        $filePath = Join-Path $script:TraceDeckRepoRoot $fileName
        if (-not (Test-Path -LiteralPath $filePath)) {
            continue
        }

        $trackedPath = & git -C $script:TraceDeckRepoRoot ls-files -- $fileName
        if ($trackedPath) {
            $trackedLocalOnly += $fileName
        }
    }

    if ($trackedLocalOnly.Count -gt 0) {
        $names = $trackedLocalOnly -join ", "
        Write-TraceDeckLog -Level "ERROR" -Message "Local-only prompt/checkpoint files are tracked: $names"
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Root artifact check passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
