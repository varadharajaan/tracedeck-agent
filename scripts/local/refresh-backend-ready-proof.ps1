param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$DataPath = "data/local/backend/backend-state.json",
    [string]$ReadyPath = "data/local/backend/backend-task-ready.json",
    [string]$StdoutPath = "",
    [string]$StderrPath = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "refresh-backend-ready-proof" -LogRoot "logs/local/backend" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Get-TraceDeckProp {
    param(
        [object]$Object,
        [string]$Name,
        [object]$Default = $null
    )

    if ($null -eq $Object) {
        return $Default
    }
    $property = $Object.PSObject.Properties[$Name]
    if ($null -eq $property) {
        return $Default
    }
    return $property.Value
}

try {
    $baseUrl = "http://$Addr"
    $pidFullPath = Resolve-TraceDeckPath -PathValue $PidPath
    $dataFullPath = Resolve-TraceDeckPath -PathValue $DataPath
    $readyFullPath = Resolve-TraceDeckPath -PathValue $ReadyPath

    if (-not (Test-Path -LiteralPath $pidFullPath)) {
        throw "PID file does not exist: $pidFullPath"
    }
    $pidText = (Get-Content -Path $pidFullPath -Raw).Trim()
    if ([string]::IsNullOrWhiteSpace($pidText)) {
        throw "PID file is empty: $pidFullPath"
    }
    $livePid = 0
    if (-not [int]::TryParse($pidText, [ref]$livePid)) {
        throw "PID file does not contain an integer PID: $pidFullPath"
    }
    $process = Get-Process -Id $livePid -ErrorAction SilentlyContinue
    if (-not $process) {
        throw "PID file points to a non-running process: $livePid"
    }

    $health = Invoke-RestMethod -Method "GET" -Uri "$baseUrl/health" -TimeoutSec 5
    if ($health.status -ne "ok") {
        throw "Backend health is not ok at $baseUrl/health"
    }

    $existingReady = $null
    if (Test-Path -LiteralPath $readyFullPath) {
        try {
            $existingReady = Get-Content -Path $readyFullPath -Raw | ConvertFrom-Json
        }
        catch {
            Write-TraceDeckLog -Level "WARN" -Message "Existing ready proof could not be parsed and will be replaced: $($_.Exception.Message)"
        }
    }

    if ([string]::IsNullOrWhiteSpace($StdoutPath)) {
        $StdoutPath = [string](Get-TraceDeckProp -Object $existingReady -Name "stdout" -Default "")
    }
    if ([string]::IsNullOrWhiteSpace($StderrPath)) {
        $StderrPath = [string](Get-TraceDeckProp -Object $existingReady -Name "stderr" -Default "")
    }

    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $readyFullPath) | Out-Null
    $ready = [ordered]@{
        addr = $Addr
        base_url = $baseUrl
        pid = $livePid
        ready_at = (Get-Date).ToString("o")
        pid_path = $pidFullPath
        data_path = $dataFullPath
        stdout = $StdoutPath
        stderr = $StderrPath
        refreshed_by = "refresh-backend-ready-proof"
    }
    Set-Content -Path $readyFullPath -Value ($ready | ConvertTo-Json -Depth 6) -Encoding UTF8

    $json = $ready | ConvertTo-Json -Depth 6
    Write-TraceDeckLog -Level "INFO" -Message "Ready proof refreshed live_pid=$livePid ready_path=$readyFullPath"
    Write-Output $json
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
