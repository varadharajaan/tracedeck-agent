param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-service-manager" -LogRoot "logs/local/test" | Out-Null

try {
    foreach ($platform in @("windows", "linux", "darwin")) {
        foreach ($action in @("install", "start", "stop", "status", "uninstall")) {
            Invoke-TraceDeckLoggedCommand -Label "Service manager dry-run $platform $action" -Command {
                powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform $platform -Action $action -DryRun
            }
        }
    }

    $plans = Get-ChildItem -Path (Join-Path $script:TraceDeckRepoRoot "data/local/service-actions/phase15") -Recurse -Filter "*.json"
    if ($plans.Count -lt 15) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected dry-run service action plans for all platforms/actions."
        exit 1
    }

    $allPlanText = ($plans | ForEach-Object { Get-Content -Raw -Path $_.FullName }) -join "`n"
    foreach ($expected in @(
        "schtasks.exe /Run",
        "sudo systemctl enable --now",
        "sudo systemctl disable --now",
        'launchctl bootstrap gui/$(id -u)',
        'launchctl kickstart -k gui/$(id -u)',
        "launchctl bootout gui/`$(id -u)"
    )) {
        if ($allPlanText -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected service manager plan command was missing: $expected"
            exit 1
        }
    }

    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
