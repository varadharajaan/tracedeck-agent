param(
    [string]$OutputDir = "data/local/webpush"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "generate-webpush-vapid-keys" -LogRoot "logs/local/webpush" | Out-Null

try {
    $resolvedOutputDir = if ([System.IO.Path]::IsPathRooted($OutputDir)) {
        [System.IO.Path]::GetFullPath($OutputDir)
    }
    else {
        [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $OutputDir))
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate Web Push VAPID keys" -Command {
        go run ./scripts/tools/webpush-keys --out-dir $resolvedOutputDir
    }

    $publicPath = Join-Path $resolvedOutputDir "vapid-public.key"
    $privatePath = Join-Path $resolvedOutputDir "vapid-private.key"
    if (-not (Test-Path -LiteralPath $publicPath) -or -not (Test-Path -LiteralPath $privatePath)) {
        throw "Expected VAPID key files were not created in $resolvedOutputDir"
    }

    [pscustomobject]@{
        status = "created"
        public_key_path = $publicPath
        private_key_path = $privatePath
        subscription_file = (Join-Path $resolvedOutputDir "subscriptions.json")
    } | ConvertTo-Json -Depth 4

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
