param(
    [string]$OutputPath = "docs/schema/policy-v1alpha1.schema.json"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "generate-policy-schema" -LogRoot "logs/local/test" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Generate checked-in policy schema" -Command {
        go run ./agent/cmd/tracedeck-agent schema --version v1alpha1 --out $OutputPath
    }
    Write-TraceDeckLog -Level "INFO" -Message "Generated policy schema at $OutputPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
