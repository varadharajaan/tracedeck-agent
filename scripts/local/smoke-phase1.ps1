param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase1" -LogRoot "logs/local/smoke" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Run local agent once" -Command {
        go run ./agent/cmd/tracedeck-agent run --once --config ./examples/policies/ai-btech-student.yaml --data-dir ./data/local/smoke --log-dir ./logs/local/agent --process-limit 25
    }

    $dbPath = Join-Path $script:TraceDeckRepoRoot "data/local/smoke/tracedeck-agent.sqlite"
    if (-not (Test-Path $dbPath)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected SQLite DB was not created: $dbPath"
        exit 1
    }
    Write-TraceDeckLog -Level "INFO" -Message "SQLite DB exists: $dbPath"

    $logPath = Join-Path $script:TraceDeckRepoRoot "logs/local/agent/tracedeck-agent.log"
    if (-not (Test-Path $logPath)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected agent log was not created: $logPath"
        exit 1
    }
    Write-TraceDeckLog -Level "INFO" -Message "Agent log exists: $logPath"

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
