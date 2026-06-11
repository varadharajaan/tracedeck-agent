param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "check-gofmt" -LogRoot "logs/local/verify" | Out-Null

try {
    $files = Get-ChildItem -Path "agent" -Recurse -Filter "*.go" | ForEach-Object { $_.FullName }
    if (-not $files) {
        Write-TraceDeckLog -Level "WARN" -Message "No Go files found under agent"
        Complete-TraceDeckScriptLog
        exit 0
    }

    $unformatted = gofmt -l $files
    if ($unformatted) {
        foreach ($file in $unformatted) {
            Write-TraceDeckLog -Level "ERROR" -Message "Unformatted Go file: $file"
        }
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "gofmt check passed"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
