param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "check-root-clean" -LogRoot "logs/local/verify" | Out-Null

try {
    $rootFiles = Get-ChildItem -Path $script:TraceDeckRepoRoot -File -Force
    $forbidden = $rootFiles | Where-Object {
        $_.Name -eq "coverage.out" -or
        $_.Name -like "*.log" -or
        $_.Name -like "*.sqlite" -or
        $_.Name -like "*.db" -or
        $_.Name -like "*.prof" -or
        $_.Name -like "*.test" -or
        $_.Name -like "*.exe"
    }

    if ($forbidden) {
        $names = ($forbidden | Select-Object -ExpandProperty Name) -join ", "
        Write-TraceDeckLog -Level "ERROR" -Message "Root contains generated artifacts: $names"
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Root artifact check passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
