param(
    [string]$OutputRoot = "data/local/go-quality/phase85",
    [switch]$SkipRace,
    [switch]$SkipSecurity
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-go-quality-gates" -LogRoot "logs/local/test" | Out-Null

function Assert-TraceDeckCommand {
    param([string]$Name)
    $command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $command) {
        throw "Required command '$Name' is not installed or not on PATH. Run scripts/setup/install-go-tools.ps1 first."
    }
    Write-TraceDeckLog -Level "INFO" -Message "Found command $Name at $($command.Source)"
}

try {
    Assert-TraceDeckCommand -Name "go"
    Assert-TraceDeckCommand -Name "golangci-lint"
    if (-not $SkipSecurity) {
        Assert-TraceDeckCommand -Name "govulncheck"
        Assert-TraceDeckCommand -Name "gosec"
    }

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $runRoot = Join-Path $script:TraceDeckRepoRoot (Join-Path $OutputRoot $timestamp)
    New-Item -ItemType Directory -Force -Path $runRoot | Out-Null

    $goVersionPath = Join-Path $runRoot "go-version.txt"
    $goTestPath = Join-Path $runRoot "go-test.txt"
    $racePath = Join-Path $runRoot "go-test-race.txt"
    $vetPath = Join-Path $runRoot "go-vet.txt"
    $lintPath = Join-Path $runRoot "golangci-lint.txt"
    $vulnPath = Join-Path $runRoot "govulncheck.txt"
    $gosecPath = Join-Path $runRoot "gosec.json"
    $summaryPath = Join-Path $runRoot "summary.json"

    Invoke-TraceDeckLoggedCommand -Label "Go version" -Command {
        go version | Tee-Object -FilePath $goVersionPath
    }
    Invoke-TraceDeckLoggedCommand -Label "Go format check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }
    $raceStatus = "pending"

    Invoke-TraceDeckLoggedCommand -Label "Go test all packages" -Command {
        go test ./... *>&1 | Tee-Object -FilePath $goTestPath
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
    if (-not $SkipRace) {
        $cgoEnabled = (go env CGO_ENABLED).Trim()
        if ($cgoEnabled -eq "1") {
            Invoke-TraceDeckLoggedCommand -Label "Go race test all packages" -Command {
                go test -race ./... *>&1 | Tee-Object -FilePath $racePath
                if ($LASTEXITCODE -ne 0) {
                    exit $LASTEXITCODE
                }
            }
            $raceStatus = "passed"
        }
        else {
            $raceStatus = "unsupported_cgo_disabled"
            "go test -race skipped because go env CGO_ENABLED=$cgoEnabled; the Go race detector requires cgo on this platform." | Set-Content -Path $racePath -Encoding UTF8
            Write-TraceDeckLog -Level "WARN" -Message "Skipping Go race test because CGO_ENABLED=$cgoEnabled; race detector is unsupported in this shell."
        }
    }
    else {
        $raceStatus = "skipped"
    }
    Invoke-TraceDeckLoggedCommand -Label "Go vet all packages" -Command {
        go vet ./... *>&1 | Tee-Object -FilePath $vetPath
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }
    Invoke-TraceDeckLoggedCommand -Label "golangci-lint all packages" -Command {
        golangci-lint run ./... *>&1 | Tee-Object -FilePath $lintPath
        if ($LASTEXITCODE -ne 0) {
            exit $LASTEXITCODE
        }
    }

    if (-not $SkipSecurity) {
        Invoke-TraceDeckLoggedCommand -Label "govulncheck all packages" -Command {
            govulncheck ./... *>&1 | Tee-Object -FilePath $vulnPath
            if ($LASTEXITCODE -ne 0) {
                exit $LASTEXITCODE
            }
        }
        Invoke-TraceDeckLoggedCommand -Label "gosec all packages" -Command {
            gosec -fmt=json -out $gosecPath ./...
            if ($LASTEXITCODE -ne 0) {
                exit $LASTEXITCODE
            }
        }
    }

    $summary = [ordered]@{
        generated_at = (Get-Date).ToUniversalTime().ToString("o")
        privacy_boundary = "quality reports only; no passwords, screenshots, raw URLs, page titles, cookies, tokens, private content, endpoint payloads, provider secrets, alert bodies, keylogging, hidden collection bypasses, or raw provider payloads"
        go_test = "passed"
        go_test_race = $raceStatus
        go_vet = "passed"
        golangci_lint = "passed"
        govulncheck = if ($SkipSecurity) { "skipped" } else { "passed" }
        gosec = if ($SkipSecurity) { "skipped" } else { "passed" }
        output_root = $runRoot
    }
    $summary | ConvertTo-Json -Depth 6 | Set-Content -Path $summaryPath -Encoding UTF8

    Write-TraceDeckLog -Level "INFO" -Message "Go quality gates passed. Summary: $summaryPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
