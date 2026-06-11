param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "install-go-tools" -LogRoot "logs/local/setup" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Install goimports" -Command {
        go install golang.org/x/tools/cmd/goimports@latest
    }

    Invoke-TraceDeckLoggedCommand -Label "Install govulncheck" -Command {
        go install golang.org/x/vuln/cmd/govulncheck@latest
    }

    Invoke-TraceDeckLoggedCommand -Label "Install gosec" -Command {
        go install github.com/securego/gosec/v2/cmd/gosec@latest
    }

    Invoke-TraceDeckLoggedCommand -Label "Install golangci-lint" -Command {
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
