param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "verify-phase0" -LogRoot "logs/local/verify" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Root artifact check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Go module tidy" -Command {
        go mod tidy
    }

    Invoke-TraceDeckLoggedCommand -Label "gofmt write" -Command {
        $files = Get-ChildItem -Path "agent" -Recurse -Filter "*.go" | ForEach-Object { $_.FullName }
        if ($files) { gofmt -w $files }
    }

    $goimports = Get-Command goimports -ErrorAction SilentlyContinue
    if ($goimports) {
        Invoke-TraceDeckLoggedCommand -Label "goimports write" -Command {
            $files = Get-ChildItem -Path "agent" -Recurse -Filter "*.go" | ForEach-Object { $_.FullName }
            if ($files) { goimports -w $files }
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "goimports not found; run scripts/setup/install-go-tools.ps1 to enable goimports verification."
    }

    Invoke-TraceDeckLoggedCommand -Label "gofmt check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-gofmt.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "go vet" -Command {
        go vet ./...
    }

    $golangciLint = Get-Command golangci-lint -ErrorAction SilentlyContinue
    if ($golangciLint) {
        Invoke-TraceDeckLoggedCommand -Label "golangci-lint" -Command {
            golangci-lint run
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "golangci-lint not found; run scripts/setup/install-go-tools.ps1 to enable lint verification."
    }

    Invoke-TraceDeckLoggedCommand -Label "go test" -Command {
        go test ./...
    }

    $cgoEnabled = (& go env CGO_ENABLED).Trim()
    if ($cgoEnabled -eq "1") {
        Invoke-TraceDeckLoggedCommand -Label "go test race" -Command {
            go test -race ./...
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "Skipping go test -race because CGO_ENABLED=$cgoEnabled in this Windows shell. Race test remains required when a race-capable toolchain is available."
    }

    $govulncheck = Get-Command govulncheck -ErrorAction SilentlyContinue
    if ($govulncheck) {
        Invoke-TraceDeckLoggedCommand -Label "govulncheck" -Command {
            govulncheck ./...
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "govulncheck not found; run scripts/setup/install-go-tools.ps1 to enable vulnerability verification."
    }

    $gosec = Get-Command gosec -ErrorAction SilentlyContinue
    if ($gosec) {
        Invoke-TraceDeckLoggedCommand -Label "gosec" -Command {
            gosec ./...
        }
    }
    else {
        Write-TraceDeckLog -Level "WARN" -Message "gosec not found; run scripts/setup/install-go-tools.ps1 to enable security verification."
    }

    Invoke-TraceDeckLoggedCommand -Label "Phase 0 smoke" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase0.ps1
    }

    Invoke-TraceDeckLoggedCommand -Label "Root artifact re-check" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
