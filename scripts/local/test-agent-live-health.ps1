param(
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$ConfigPath = "data/local/config/tracedeck-live-this-machine.yaml",
    [string]$PidPath = "data/local/agent-live/tracedeck-agent-live.pid",
    [string]$LogDir = "logs/local/agent-live",
    [string]$OutputRoot = "data/local/agent-live/health",
    [int]$WaitSeconds = 90,
    [int]$MinStoredEvents = 1
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-agent-live-health" -LogRoot "logs/local/agent-live" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Read-PolicyValue {
    param(
        [string]$ResolvedConfigPath,
        [string]$Name
    )
    $pattern = "^\s*$([regex]::Escape($Name))\s*:\s*(.+?)\s*$"
    foreach ($line in Get-Content -LiteralPath $ResolvedConfigPath) {
        if ($line -match $pattern) {
            return ($Matches[1] -replace '^"|"$', '').Trim()
        }
    }
    return ""
}

function Get-TraceDeckAgentProcessProof {
    param(
        [string]$ResolvedPidPath,
        [string]$ResolvedConfigPath
    )

    if (Test-Path -LiteralPath $ResolvedPidPath) {
        $pidText = (Get-Content -LiteralPath $ResolvedPidPath -Raw).Trim()
        if ($pidText -match "^\d+$") {
            $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
            if ($process) {
                return [pscustomobject]@{
                    pid = $pidText
                    process_alive = $true
                    evidence = "pid_file"
                }
            }
        }
    }

    $escapedConfig = [regex]::Escape($ResolvedConfigPath)
    $candidate = Get-CimInstance Win32_Process -ErrorAction SilentlyContinue |
        Where-Object { $_.Name -ieq "tracedeck-agent.exe" -and $_.CommandLine -match $escapedConfig } |
        Sort-Object ProcessId |
        Select-Object -First 1

    if ($candidate) {
        return [pscustomobject]@{
            pid = [string]$candidate.ProcessId
            process_alive = $true
            evidence = "command_line"
        }
    }

    return [pscustomobject]@{
        pid = ""
        process_alive = $false
        evidence = "none"
    }
}

try {
    $resolvedConfigPath = Resolve-TraceDeckPath -PathValue $ConfigPath
    $resolvedPidPath = Resolve-TraceDeckPath -PathValue $PidPath
    $resolvedLogDir = Resolve-TraceDeckPath -PathValue $LogDir
    $resolvedOutputRoot = Resolve-TraceDeckPath -PathValue $OutputRoot
    New-Item -ItemType Directory -Force -Path $resolvedOutputRoot | Out-Null

    $tenantId = Read-PolicyValue -ResolvedConfigPath $resolvedConfigPath -Name "tenant_id"
    $deviceId = Read-PolicyValue -ResolvedConfigPath $resolvedConfigPath -Name "device_id"
    if ([string]::IsNullOrWhiteSpace($tenantId) -or [string]::IsNullOrWhiteSpace($deviceId)) {
        throw "Unable to read tenant_id/device_id from $resolvedConfigPath"
    }

    $deadline = (Get-Date).AddSeconds($WaitSeconds)
    $lastError = ""
    $status = $null
    $pidValue = ""
    $processAlive = $false
    $processEvidence = "none"
    while ((Get-Date) -lt $deadline) {
        try {
            $processProof = Get-TraceDeckAgentProcessProof -ResolvedPidPath $resolvedPidPath -ResolvedConfigPath $resolvedConfigPath
            $pidValue = $processProof.pid
            $processAlive = [bool]$processProof.process_alive
            $processEvidence = $processProof.evidence
            if (-not $processAlive) {
                throw "Agent process is not alive; pid_file=$resolvedPidPath config=$resolvedConfigPath"
            }
            $health = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/health" -TimeoutSec 10
            if ($health.status -ne "ok") {
                throw "Backend health is $($health.status)"
            }
            $status = Invoke-RestMethod -Method "GET" -Uri "$BaseUrl/api/v1/devices/$deviceId/telemetry-status" -TimeoutSec 10
            if ([int]$status.stored_events -ge $MinStoredEvents) {
                break
            }
            $lastError = "Telemetry stored_events=$($status.stored_events), waiting for >= $MinStoredEvents"
        }
        catch {
            $lastError = $_.Exception.Message
        }
        Start-Sleep -Seconds 3
    }

    if (-not $status -or [int]$status.stored_events -lt $MinStoredEvents -or -not $processAlive) {
        throw "Live agent health check failed: $lastError"
    }

    $latestLog = Get-ChildItem -LiteralPath $resolvedLogDir -File -ErrorAction SilentlyContinue |
        Sort-Object LastWriteTime -Descending |
        Select-Object -First 1
    $report = [pscustomobject]@{
        status = "ok"
        tenant_id = $tenantId
        device_id = $deviceId
        pid = $pidValue
        process_alive = $processAlive
        process_evidence = $processEvidence
        backend_url = $BaseUrl
        stored_events = [int]$status.stored_events
        last_observed_at = $status.last_observed_at
        last_ingested_at = $status.last_ingested_at
        counts_by_type = $status.counts_by_type
        counts_by_source = $status.counts_by_source
        latest_log = if ($latestLog) { $latestLog.FullName } else { "" }
        checked_at = (Get-Date).ToUniversalTime().ToString("o")
    }
    $reportPath = Join-Path $resolvedOutputRoot ("agent-live-health-{0}.json" -f (Get-Date -Format "yyyyMMdd-HHmmss"))
    $report | ConvertTo-Json -Depth 8 | Set-Content -LiteralPath $reportPath -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Live agent health passed report=$reportPath stored_events=$($status.stored_events) pid=$pidValue"
    $report | ConvertTo-Json -Depth 8
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
