param(
    [string]$OutputRoot = "data/local/schema-check/phase35"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-policy-schema" -LogRoot "logs/local/test" | Out-Null

function Get-NormalizedSchemaHash {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $text = (Get-Content -Raw -Path $Path) -replace "`r`n", "`n"
    $bytes = [System.Text.Encoding]::UTF8.GetBytes($text.Trim())
    $sha = [System.Security.Cryptography.SHA256]::Create()
    try {
        return [System.BitConverter]::ToString($sha.ComputeHash($bytes)).Replace("-", "")
    }
    finally {
        $sha.Dispose()
    }
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $runRoot = Join-Path $script:TraceDeckRepoRoot (Join-Path $OutputRoot $timestamp)
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null
    $generatedPath = Join-Path $runRoot "policy-v1alpha1.schema.json"
    $checkedInPath = Join-Path $script:TraceDeckRepoRoot "docs/schema/policy-v1alpha1.schema.json"

    Invoke-TraceDeckLoggedCommand -Label "Generate versioned policy schema" -Command {
        go run ./agent/cmd/tracedeck-agent schema --version v1alpha1 --out $generatedPath
    }

    $generatedHash = Get-NormalizedSchemaHash -Path $generatedPath
    $checkedInHash = Get-NormalizedSchemaHash -Path $checkedInPath
    if ($generatedHash -ne $checkedInHash) {
        Write-TraceDeckLog -Level "ERROR" -Message "Generated schema differs from checked-in schema. generated=$generatedPath checked_in=$checkedInPath"
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Policy schema is current. generated=$generatedPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
