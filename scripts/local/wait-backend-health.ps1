param(
    [string]$Addr = "127.0.0.1:18080",
    [int]$TimeoutSeconds = 45
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "wait-backend-health" -LogRoot "logs/local/backend" | Out-Null

try {
    $baseUrl = "http://$Addr"
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    $lastError = ""
    do {
        try {
            $body = & curl.exe --fail --silent --show-error --max-time 3 "$baseUrl/health"
            if ($LASTEXITCODE -eq 0 -and -not [string]::IsNullOrWhiteSpace($body)) {
                $health = $body | ConvertFrom-Json
                if ($health.status -eq "ok") {
                    Write-TraceDeckLog -Level "INFO" -Message "Backend healthy at $baseUrl"
                    Complete-TraceDeckScriptLog
                    exit 0
                }
                $lastError = "Unexpected health status: $($health.status)"
            }
        }
        catch {
            $lastError = $_.Exception.Message
        }
        Start-Sleep -Milliseconds 500
    } while ((Get-Date) -lt $deadline)

    throw "Backend did not become healthy at $baseUrl within $TimeoutSeconds seconds. Last error: $lastError"
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
