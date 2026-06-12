param(
    [string]$Addr = "127.0.0.1:18108",
    [string]$ApiKey = "phase21-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "newman-phase21" -LogRoot "logs/local/newman" | Out-Null

$backend = $null

function Wait-TraceDeckBackend {
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(45)
    while ((Get-Date) -lt $deadline) {
        try {
            $health = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/health"
            if ($health.status -eq "ok") {
                return
            }
        }
        catch {
            Start-Sleep -Milliseconds 500
        }
    }
    throw "Backend did not become healthy at $BaseUrl"
}

try {
    $newman = Get-Command newman -ErrorAction SilentlyContinue
    if (-not $newman) {
        throw "newman is not installed or not on PATH"
    }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $runRoot = Join-Path $script:TraceDeckRepoRoot "data/local/newman/phase21/$timestamp"
    $stdoutPath = Join-Path $runRoot "backend.out.log"
    $stderrPath = Join-Path $runRoot "backend.err.log"
    $reportPath = Join-Path $runRoot "newman-report.json"
    $exePath = Join-Path $runRoot "tracedeck-backend.exe"
    $statePath = Join-Path $runRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

    $baseUrl = "http://$Addr"
    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 21 Newman executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase21-newman"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started backend for Phase 21 Newman pid=$($backend.Id) addr=$Addr"

    Wait-TraceDeckBackend -BaseUrl $baseUrl

    Invoke-TraceDeckLoggedCommand -Label "Run Newman Phase 21 collection" -Command {
        newman run ./postman/tracedeck-backend-phase21.postman_collection.json --env-var "baseUrl=$baseUrl" --env-var "apiKey=$ApiKey" --reporters "cli,json" --reporter-json-export $reportPath
    }

    if (-not (Test-Path $reportPath)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected Newman report was not created: $reportPath"
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Newman Phase 21 report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped backend Phase 21 Newman process: $($backend.Id)"
    }
}
