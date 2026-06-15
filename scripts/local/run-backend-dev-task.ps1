param(
    [string]$Addr = "127.0.0.1:18080",
    [string]$PidPath = "data/local/backend/tracedeck-backend.pid",
    [string]$DataPath = "data/local/backend/backend-state.json",
    [string]$ExePath = "data/local/backend/tracedeck-dashboard-demo.exe",
    [string]$ReadyPath = "data/local/backend/backend-task-ready.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "run-backend-dev-task" -LogRoot "logs/local/backend" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

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
    param([string]$BaseUrl)

    $deadline = (Get-Date).AddSeconds(60)
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
        Start-Sleep -Milliseconds 500
    }
    throw "Backend did not become healthy at $BaseUrl"
}

try {
    $baseUrl = "http://$Addr"
    $exeFullPath = Resolve-TraceDeckPath -PathValue $ExePath
    $pidFullPath = Resolve-TraceDeckPath -PathValue $PidPath
    $dataFullPath = Resolve-TraceDeckPath -PathValue $DataPath
    $readyFullPath = Resolve-TraceDeckPath -PathValue $ReadyPath
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $pidFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $dataFullPath) | Out-Null
    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $readyFullPath) | Out-Null

    if (-not (Test-Path -LiteralPath $exeFullPath)) {
        throw "Backend executable does not exist: $exeFullPath. Run scripts/local/start-backend-dev-task.ps1 first."
    }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $stdoutPath = Join-Path $script:TraceDeckRepoRoot "logs/local/backend/backend-task-$timestamp.out.log"
    $stderrPath = Join-Path $script:TraceDeckRepoRoot "logs/local/backend/backend-task-$timestamp.err.log"

    $process = Start-Process -FilePath $exeFullPath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$dataFullPath`""
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput $stdoutPath -RedirectStandardError $stderrPath -PassThru

    Set-Content -Path $pidFullPath -Value $process.Id
    Write-TraceDeckLog -Level "INFO" -Message "Started scheduled backend dev process pid=$($process.Id) addr=$Addr"
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

    $ready = [pscustomobject]@{
        addr = $Addr
        base_url = $baseUrl
        pid = $process.Id
        ready_at = (Get-Date).ToString("o")
        pid_path = $pidFullPath
        data_path = $dataFullPath
        stdout = $stdoutPath
        stderr = $stderrPath
    }
    Set-Content -Path $readyFullPath -Value ($ready | ConvertTo-Json -Depth 4) -Encoding UTF8
    Write-TraceDeckLog -Level "INFO" -Message "Scheduled backend dev task ready at $baseUrl"

    Wait-Process -Id $process.Id
    Write-TraceDeckLog -Level "WARN" -Message "Scheduled backend dev process exited pid=$($process.Id)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
