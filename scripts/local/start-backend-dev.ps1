param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$DataPath = "data/local/backend/backend-state.json",
    [string]$ApiKey = "",
    [string]$ApiKeyTenantId = "",
    [string]$ApiKeyActorId = "local_backend"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "start-backend-dev" -LogRoot "logs/local/backend" | Out-Null

try {
    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $PidPath
    $pidDir = Split-Path -Parent $pidFullPath
    New-Item -ItemType Directory -Force -Path $pidDir | Out-Null

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $stdoutPath = Join-Path $script:TraceDeckRepoRoot "logs/local/backend/backend-dev-$timestamp.out.log"
    $stderrPath = Join-Path $script:TraceDeckRepoRoot "logs/local/backend/backend-dev-$timestamp.err.log"
    $exePath = Join-Path $script:TraceDeckRepoRoot "data/local/backend/tracedeck-backend-dev.exe"

    Invoke-TraceDeckLoggedCommand -Label "Build backend dev executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $arguments = @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$DataPath`""
    )
    if ($ApiKey) {
        $arguments += @("--api-key", $ApiKey)
    }
    if ($ApiKeyTenantId) {
        $arguments += @("--api-key-tenant-id", $ApiKeyTenantId)
    }
    if ($ApiKeyActorId) {
        $arguments += @("--api-key-actor-id", $ApiKeyActorId)
    }

    $process = Start-Process -FilePath $exePath -ArgumentList $arguments -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru

    Set-Content -Path $pidFullPath -Value $process.Id
    Write-TraceDeckLog -Level "INFO" -Message "Started TraceDeck backend dev server pid=$($process.Id) addr=$Addr data_path=$DataPath pid_file=$pidFullPath"
    Write-TraceDeckLog -Level "INFO" -Message "stdout=$stdoutPath stderr=$stderrPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
