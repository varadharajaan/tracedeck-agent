param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "update-phase17-deps" -LogRoot "logs/local/setup" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Add AWS SESv2 SDK dependency" -Command {
        go get github.com/aws/aws-sdk-go-v2/service/sesv2@latest
    }

    Invoke-TraceDeckLoggedCommand -Label "Tidy Go modules" -Command {
        go mod tidy
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
