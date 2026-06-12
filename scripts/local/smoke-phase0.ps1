param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase0" -LogRoot "logs/local/smoke" | Out-Null

try {
    Invoke-TraceDeckLoggedCommand -Label "Validate sample policy" -Command {
        go run ./agent/cmd/tracedeck-agent validate-config --config ./examples/policies/ai-btech-student.yaml
    }

    Invoke-TraceDeckLoggedCommand -Label "Generate policy schema" -Command {
        go run ./agent/cmd/tracedeck-agent schema --out ./docs/schema/policy-v1alpha1.schema.json
    }

    Invoke-TraceDeckLoggedCommand -Label "Agent bootstrap smoke" -Command {
        go run ./agent/cmd/tracedeck-agent run --once --config ./examples/policies/ai-btech-student.yaml --disable-browser-history
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
