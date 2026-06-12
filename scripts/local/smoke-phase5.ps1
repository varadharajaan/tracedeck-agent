param(
    [string]$Addr = "127.0.0.1:18080"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase5" -LogRoot "logs/local/smoke" | Out-Null

$backend = $null

function Invoke-TraceDeckJson {
    param(
        [string]$Method,
        [string]$Uri,
        [string]$Body = ""
    )

    $headers = @{ "Content-Type" = "application/json" }
    if ($Body) {
        return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $headers -Body $Body
    }
    return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $headers
}

function Wait-TraceDeckBackend {
    param(
        [string]$BaseUrl
    )

    $deadline = (Get-Date).AddSeconds(45)
    while ((Get-Date) -lt $deadline) {
        try {
            $health = Invoke-TraceDeckJson -Method "GET" -Uri "$BaseUrl/health"
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
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase5/$timestamp"
    $stdoutPath = Join-Path $smokeRoot "backend.out.log"
    $stderrPath = Join-Path $smokeRoot "backend.err.log"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    Invoke-TraceDeckLoggedCommand -Label "Build backend smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started backend smoke process: $($backend.Id)"

    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $version = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/version"
    if ($version.service -ne "tracedeck-backend") {
        Write-TraceDeckLog -Level "ERROR" -Message "Unexpected backend version response."
        exit 1
    }

    $deviceBody = @{
        tenant_id = "family-varadha"
        device_id = "phase5-smoke-device"
        host_name = "phase5-smoke-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress

    $device = Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody
    if ($device.device_id -ne "phase5-smoke-device") {
        Write-TraceDeckLog -Level "ERROR" -Message "Device enrollment smoke failed."
        exit 1
    }

    $devices = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices"
    if ($devices.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected at least one enrolled device."
        exit 1
    }

    $summary = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase5-smoke-device/summary/daily"
    if ($summary.device_id -ne "phase5-smoke-device") {
        Write-TraceDeckLog -Level "ERROR" -Message "Daily summary smoke failed."
        exit 1
    }

    $templates = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/policy-templates"
    if ($templates.count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected policy templates."
        exit 1
    }

    $archive = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/archive/status"
    if ($archive.provider -ne "s3") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected archive provider s3."
        exit 1
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Method "GET" -Uri "$baseUrl/"
    if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch "TraceDeck") {
        Write-TraceDeckLog -Level "ERROR" -Message "Dashboard smoke failed."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 5 backend smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped backend smoke process: $($backend.Id)"
    }
}
