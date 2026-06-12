param(
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$Addr = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "stop-backend-dev" -LogRoot "logs/local/backend" | Out-Null

function Stop-TraceDeckProcess {
    param(
        [System.Diagnostics.Process]$Process,
        [string]$Reason
    )

    if (-not $Process) { return 0 }
    $path = ""
    try { $path = [string]$Process.Path } catch { $path = "" }
    $isTraceDeck = $Process.ProcessName -like "tracedeck*" -or $path.StartsWith($script:TraceDeckRepoRoot, [System.StringComparison]::OrdinalIgnoreCase)
    if (-not $isTraceDeck) {
        throw "Refusing to stop non-TraceDeck process pid=$($Process.Id) name=$($Process.ProcessName) path=$path"
    }
    Stop-Process -Id $Process.Id -Force
    Write-TraceDeckLog -Level "INFO" -Message "Stopped backend process $Reason pid=$($Process.Id) name=$($Process.ProcessName)"
    return 1
}

function Stop-TraceDeckPortListener {
    param([string]$ListenAddr)

    $parts = $ListenAddr -split ":", 2
    if ($parts.Count -ne 2) { return 0 }
    $port = 0
    if (-not [int]::TryParse($parts[1], [ref]$port)) { return 0 }

    $stopped = 0
    $connections = @(Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue)
    foreach ($connection in $connections) {
        $process = Get-Process -Id $connection.OwningProcess -ErrorAction SilentlyContinue
        if ($process) {
            $stopped += Stop-TraceDeckProcess -Process $process -Reason "listening on $ListenAddr"
        }
    }
    return $stopped
}

try {
    $stopped = 0
    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $PidPath
    if (Test-Path $pidFullPath) {
        $pidText = (Get-Content -Path $pidFullPath -Raw).Trim()
        if ($pidText) {
            $process = Get-Process -Id ([int]$pidText) -ErrorAction SilentlyContinue
            if ($process) {
                $stopped += Stop-TraceDeckProcess -Process $process -Reason "from pid file"
            }
        }
        Remove-Item -LiteralPath $pidFullPath -Force
    }

    if ($Addr) {
        $stopped += Stop-TraceDeckPortListener -ListenAddr $Addr
    }

    if (-not $Addr) {
        $orphans = @(Get-Process -ErrorAction SilentlyContinue | Where-Object {
            $_.ProcessName -like "tracedeck-backend*" -or $_.ProcessName -like "tracedeck-dashboard-demo*"
        })
        foreach ($orphan in $orphans) {
            $stopped += Stop-TraceDeckProcess -Process $orphan -Reason "by TraceDeck process name"
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Backend stop complete; stopped=$stopped"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
