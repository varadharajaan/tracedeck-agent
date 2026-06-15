param(
    [string]$Addr = "127.0.0.1:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase88" -LogRoot "logs/local/verify" | Out-Null

function Wait-TraceDeckBackendReady {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(60)
    while ((Get-Date) -lt $deadline) {
        try {
            $health = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/health" -TimeoutSec 2
            if ($health.status -eq "ok") {
                return
            }
        }
        catch {
            Start-Sleep -Milliseconds 500
        }
        Start-Sleep -Milliseconds 500
    }
    throw "TraceDeck backend did not become healthy at $BaseUrl"
}

try {
    $baseUrl = "http://$Addr"

    Invoke-TraceDeckLoggedCommand -Label "Phase 86 premium UI and provenance verifier" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase86.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Strict Go quality gate" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-go-quality-gates.ps1
    }
    Write-TraceDeckLog -Level "INFO" -Message "Starting: Restart persistent local backend"
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -Addr $Addr
    if ($LASTEXITCODE -ne 0) {
        throw "stop-backend-dev failed with exit code $LASTEXITCODE"
    }
    powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/start-backend-dev.ps1 -Addr $Addr
    if ($LASTEXITCODE -ne 0) {
        throw "start-backend-dev failed with exit code $LASTEXITCODE"
    }
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Restart persistent local backend"

    Invoke-TraceDeckLoggedCommand -Label "Wait for persistent local backend" -Command {
        Wait-TraceDeckBackendReady -BaseUrl $baseUrl
    }
    Invoke-TraceDeckLoggedCommand -Label "Live server provenance guard" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-server-provenance.ps1 -BaseUrl $baseUrl
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 88 smoke verifier" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase88.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Phase 88 Newman collection" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase88.ps1
    }
    Invoke-TraceDeckLoggedCommand -Label "Root artifact re-check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
