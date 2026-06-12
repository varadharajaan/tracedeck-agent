param(
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "stop-backend-dev" -LogRoot "logs/local/backend" | Out-Null

try {
    $stopped = 0
    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $PidPath
    if (Test-Path $pidFullPath) {
        $pidText = (Get-Content -Path $pidFullPath -Raw).Trim()
        if ($pidText) {
            $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
            if ($process) {
                Stop-Process -Id $process.Id -Force
                $stopped += 1
                Write-TraceDeckLog -Level "INFO" -Message "Stopped backend pid from pid file: $($process.Id)"
            }
        }
        Remove-Item -LiteralPath $pidFullPath -Force
    }

    $orphans = @(Get-Process -Name "tracedeck-backend" -ErrorAction SilentlyContinue)
    foreach ($orphan in $orphans) {
        Stop-Process -Id $orphan.Id -Force
        $stopped += 1
        Write-TraceDeckLog -Level "INFO" -Message "Stopped backend process by name: $($orphan.Id)"
    }

    Write-TraceDeckLog -Level "INFO" -Message "Backend stop complete; stopped=$stopped"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
