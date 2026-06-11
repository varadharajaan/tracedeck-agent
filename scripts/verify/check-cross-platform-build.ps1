param(
    [string]$BuildRoot = "data/local/build/phase1b"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "check-cross-platform-build" -LogRoot "logs/local/verify" | Out-Null

$originalGOOS = $env:GOOS
$originalGOARCH = $env:GOARCH
$originalCGO = $env:CGO_ENABLED

try {
    $outputRoot = Join-Path $script:TraceDeckRepoRoot $BuildRoot
    New-Item -ItemType Directory -Force -Path $outputRoot | Out-Null

    $targets = @(
        @{ GOOS = "windows"; GOARCH = "amd64"; Extension = ".exe" },
        @{ GOOS = "darwin"; GOARCH = "amd64"; Extension = "" },
        @{ GOOS = "linux"; GOARCH = "amd64"; Extension = "" }
    )

    foreach ($target in $targets) {
        $env:GOOS = $target.GOOS
        $env:GOARCH = $target.GOARCH
        $env:CGO_ENABLED = "0"

        $fileName = "tracedeck-agent-$($target.GOOS)-$($target.GOARCH)$($target.Extension)"
        $outputPath = Join-Path $outputRoot $fileName

        Invoke-TraceDeckLoggedCommand -Label "Cross-build $($target.GOOS)/$($target.GOARCH)" -Command {
            go build -trimpath -o $outputPath ./agent/cmd/tracedeck-agent
        }

        if (-not (Test-Path $outputPath)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected build output was not created: $outputPath"
            exit 1
        }
        Write-TraceDeckLog -Level "INFO" -Message "Build output exists: $outputPath"
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    $env:GOOS = $originalGOOS
    $env:GOARCH = $originalGOARCH
    $env:CGO_ENABLED = $originalCGO
}
