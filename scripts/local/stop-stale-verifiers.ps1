param(
    [switch]$IncludeAgentExecutables
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "stop-stale-verifiers" -LogRoot "logs/local/ops" | Out-Null

try {
    $escapedRoot = [WildcardPattern]::Escape($script:TraceDeckRepoRoot)
    $scriptPatterns = @(
        "*verify-phase0.ps1*",
        "*verify-phase1.ps1*",
        "*verify-phase1b.ps1*",
        "*verify-phase2.ps1*",
        "*verify-phase2b.ps1*",
        "*smoke-phase0.ps1*",
        "*smoke-phase1.ps1*",
        "*smoke-phase2.ps1*",
        "*smoke-phase2b.ps1*",
        "*./agent/cmd/tracedeck-agent*",
        "*.\agent\cmd\tracedeck-agent*"
    )
    $staleProcesses = Get-CimInstance Win32_Process |
        Where-Object {
            if (-not $_.CommandLine) {
                $false
            }
            elseif ($_.ProcessId -eq $PID) {
                $false
            }
            elseif ($_.Name -notin @("powershell.exe", "pwsh.exe", "go.exe", "gosec.exe", "govulncheck.exe", "golangci-lint.exe", "tracedeck-agent.exe")) {
                $false
            }
            else {
                $matchesRepoRoot = $_.CommandLine -like "*$escapedRoot*"
                $matchesTraceDeckScript = $false
                foreach ($pattern in $scriptPatterns) {
                    if ($_.CommandLine -like $pattern) {
                        $matchesTraceDeckScript = $true
                        break
                    }
                }
                $matchesOrphanAgent = $IncludeAgentExecutables -and $_.Name -eq "tracedeck-agent.exe"

                $matchesRepoRoot -or $matchesTraceDeckScript -or $matchesOrphanAgent
            }
        }

    if (-not $staleProcesses) {
        Write-TraceDeckLog -Level "INFO" -Message "No stale verifier processes found."
        Complete-TraceDeckScriptLog
        return
    }

    foreach ($process in $staleProcesses) {
        Write-TraceDeckLog -Level "WARN" -Message "Stopping stale process $($process.ProcessId) $($process.Name)"
        Stop-Process -Id $process.ProcessId -Force
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
