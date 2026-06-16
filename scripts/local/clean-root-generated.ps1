param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "clean-root-generated" -LogRoot "logs/local/cleanup" | Out-Null

function Resolve-TraceDeckRootChild {
    param([string]$Name)

    $candidate = [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $Name))
    $root = [System.IO.Path]::GetFullPath($script:TraceDeckRepoRoot)
    if (-not $candidate.StartsWith($root, [System.StringComparison]::OrdinalIgnoreCase)) {
        throw "Refusing to clean outside workspace: $candidate"
    }
    if ($candidate -eq $root) {
        throw "Refusing to clean workspace root directly."
    }
    return $candidate
}

try {
    $generatedRootDirs = @(
        "__pycache__",
        ".pytest_cache",
        "playwright-report",
        "test-results"
    )

    foreach ($dirName in $generatedRootDirs) {
        $path = Resolve-TraceDeckRootChild -Name $dirName
        if (Test-Path -LiteralPath $path) {
            Remove-Item -LiteralPath $path -Recurse -Force
            Write-TraceDeckLog -Level "INFO" -Message "Removed generated root directory: $dirName"
        }
    }

    Write-TraceDeckLog -Level "INFO" -Message "Root generated artifact cleanup completed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
