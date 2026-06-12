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

        $agentFileName = "tracedeck-agent-$($target.GOOS)-$($target.GOARCH)$($target.Extension)"
        $agentOutputPath = Join-Path $outputRoot $agentFileName

        Invoke-TraceDeckLoggedCommand -Label "Cross-build agent $($target.GOOS)/$($target.GOARCH)" -Command {
            go build -trimpath -o $agentOutputPath ./agent/cmd/tracedeck-agent
        }

        if (-not (Test-Path $agentOutputPath)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected build output was not created: $agentOutputPath"
            exit 1
        }
        Write-TraceDeckLog -Level "INFO" -Message "Build output exists: $agentOutputPath"

        $backendFileName = "tracedeck-backend-$($target.GOOS)-$($target.GOARCH)$($target.Extension)"
        $backendOutputPath = Join-Path $outputRoot $backendFileName

        Invoke-TraceDeckLoggedCommand -Label "Cross-build backend $($target.GOOS)/$($target.GOARCH)" -Command {
            go build -trimpath -o $backendOutputPath ./backend/cmd/tracedeck-backend
        }

        if (-not (Test-Path $backendOutputPath)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected build output was not created: $backendOutputPath"
            exit 1
        }
        Write-TraceDeckLog -Level "INFO" -Message "Build output exists: $backendOutputPath"
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
