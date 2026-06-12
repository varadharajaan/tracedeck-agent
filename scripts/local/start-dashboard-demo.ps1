param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [switch]$StopAfterSeed
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "start-dashboard-demo" -LogRoot "logs/local/backend" | Out-Null

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
    $baseUrl = "http://$Addr"
    $pidFullPath = Join-Path $script:TraceDeckRepoRoot $PidPath
    $pidDir = Split-Path -Parent $pidFullPath
    New-Item -ItemType Directory -Force -Path $pidDir | Out-Null

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $stdoutPath = Join-Path $script:TraceDeckRepoRoot "logs/local/backend/dashboard-demo-$timestamp.out.log"
    $stderrPath = Join-Path $script:TraceDeckRepoRoot "logs/local/backend/dashboard-demo-$timestamp.err.log"
    $exePath = Join-Path $script:TraceDeckRepoRoot "data/local/backend/tracedeck-dashboard-demo.exe"

    Invoke-TraceDeckLoggedCommand -Label "Build dashboard demo backend" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $process = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru

    Set-Content -Path $pidFullPath -Value $process.Id
    Write-TraceDeckLog -Level "INFO" -Message "Started dashboard demo backend pid=$($process.Id) addr=$Addr pid_file=$pidFullPath"
    Write-TraceDeckLog -Level "INFO" -Message "stdout=$stdoutPath stderr=$stderrPath"

    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody | Out-Null

    $deviceBody = @{
        tenant_id = "family-varadha"
        device_id = "demo-study-laptop"
        host_name = "demo-study-laptop"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody | Out-Null

    $overview = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/demo-study-laptop/overview"
    if ($overview.device.device_id -ne "demo-study-laptop" -or $overview.policy_violations.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Dashboard demo seed failed."
        exit 1
    }

    $dashboard = Invoke-WebRequest -UseBasicParsing -Method "GET" -Uri "$baseUrl/"
    if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch "TraceDeck Command Center") {
        Write-TraceDeckLog -Level "ERROR" -Message "Dashboard HTML verification failed."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Dashboard demo ready: $baseUrl/"
    if ($StopAfterSeed) {
        Stop-Process -Id $process.Id -Force
        Remove-Item -LiteralPath $pidFullPath -Force -ErrorAction SilentlyContinue
        Write-TraceDeckLog -Level "INFO" -Message "Stopped dashboard demo backend after verification pid=$($process.Id)"
    }
    else {
        Write-TraceDeckLog -Level "INFO" -Message "Stop it with: powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/stop-backend-dev.ps1 -PidPath $PidPath"
    }
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
