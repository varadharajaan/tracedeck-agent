param(
    [string]$Addr = "127.0.0.1:18095",
    [string]$ApiKey = "phase14-local-secret"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase14" -LogRoot "logs/local/smoke" | Out-Null

$backend = $null

function Invoke-TraceDeckJson {
    param(
        [string]$Method,
        [string]$Uri,
        [string]$Body = "",
        [hashtable]$Headers = @{}
    )

    $requestHeaders = @{ "Content-Type" = "application/json" }
    foreach ($key in $Headers.Keys) {
        $requestHeaders[$key] = $Headers[$key]
    }
    if ($Body) {
        return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $requestHeaders -Body $Body
    }
    return Invoke-RestMethod -Method $Method -Uri $Uri -Headers $requestHeaders
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
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase14/$timestamp"
    $exePath = Join-Path $smokeRoot "tracedeck-backend.exe"
    $statePath = Join-Path $smokeRoot "backend-state.json"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null

    $baseUrl = "http://$Addr"
    $authHeaders = @{
        "X-TraceDeck-API-Key" = $ApiKey
        "X-TraceDeck-Tenant-ID" = "family-varadha"
    }

    Invoke-TraceDeckLoggedCommand -Label "Build backend Phase 14 smoke executable" -Command {
        go build -trimpath -o $exePath ./backend/cmd/tracedeck-backend
    }

    $backend = Start-Process -FilePath $exePath -ArgumentList @(
        "--addr", $Addr,
        "--log-dir", "./logs/local/backend",
        "--data-path", "`"$statePath`"",
        "--api-key", $ApiKey,
        "--api-key-tenant-id", "family-varadha",
        "--api-key-actor-id", "phase14-smoke"
    ) -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "backend.out.log") -RedirectStandardError (Join-Path $smokeRoot "backend.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 14 backend pid=$($backend.Id) addr=$Addr"
    Wait-TraceDeckBackend -BaseUrl $baseUrl

    $tenantBody = @{
        tenant_id = "family-varadha"
        name = "Family Varadha"
        plan_id = "family_pro"
        retention_tier_id = "family_cloud_90_365_archive"
        primary_profile = "ai-btech-student"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/tenants" -Body $tenantBody -Headers $authHeaders | Out-Null

    $deviceBody = @{
        tenant_id = "family-varadha"
        device_id = "phase14-report-device"
        host_name = "phase14-report-host"
        profile = "ai-btech-student"
        os_name = "windows"
    } | ConvertTo-Json -Compress
    Invoke-TraceDeckJson -Method "POST" -Uri "$baseUrl/api/v1/devices/enroll" -Body $deviceBody -Headers $authHeaders | Out-Null

    $report = Invoke-TraceDeckJson -Method "GET" -Uri "$baseUrl/api/v1/devices/phase14-report-device/reports/weekly" -Headers $authHeaders
    if (-not $report.generated -or -not $report.email_ready -or -not $report.pdf_ready -or $report.highlights.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected generated weekly report with email/PDF readiness."
        exit 1
    }

    $pdf = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/api/v1/devices/phase14-report-device/reports/weekly/pdf" -Headers $authHeaders -UseBasicParsing
    if ($pdf.StatusCode -ne 200 -or $pdf.Headers["Content-Type"] -notmatch "application/pdf") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected weekly report PDF response."
        exit 1
    }
    $pdfText = [System.Text.Encoding]::ASCII.GetString($pdf.Content)
    if (-not $pdfText.StartsWith("%PDF-1.4")) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected PDF header."
        exit 1
    }

    $dashboard = Invoke-WebRequest -Method "GET" -Uri "$baseUrl/" -UseBasicParsing
    if ($dashboard.StatusCode -ne 200 -or $dashboard.Content -notmatch "Weekly Report") {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected weekly report dashboard panel."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 14 weekly report smoke passed at $baseUrl"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    if ($backend -and -not $backend.HasExited) {
        Stop-Process -Id $backend.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 14 backend pid=$($backend.Id)"
    }
}
