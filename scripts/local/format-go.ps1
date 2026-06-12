param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "format-go" -LogRoot "logs/local/format" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Format Go source" -Command {
        $files = @()
        foreach ($path in @("agent", "scripts/tools")) {
            if (Test-Path $path) {
                $files += Get-ChildItem -Path $path -Recurse -Filter "*.go" | ForEach-Object { $_.FullName }
            }
        }
        if ($files) { gofmt -w $files }
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
