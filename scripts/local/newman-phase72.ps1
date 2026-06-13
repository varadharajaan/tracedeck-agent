param(
    [string]$Bucket = "",
    [string]$Region = "ap-south-1",
    [string]$FrontendUrl = ""
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "newman-phase72" -LogRoot "logs/local/newman" | Out-Null

function Get-TraceDeckFrontendUrl {
    $path = Join-Path $script:TraceDeckRepoRoot "data/local/output/frontend-url.txt"
    if (!(Test-Path $path)) {
        throw "Frontend URL output is missing. Run 'python ./devctl.py sam outputs'."
    }
    return (Get-Content -Raw -Path $path).Trim().TrimEnd("/")
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $runRoot = "data/local/newman/phase72/$timestamp"
    $reportPath = "$runRoot/newman-report.json"
    $manifestPath = "$runRoot/cloud-sample-manifest.json"
    $reportDir = Split-Path -Parent (Join-Path $script:TraceDeckRepoRoot $reportPath)
    New-Item -ItemType Directory -Force -Path $reportDir | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Upload Phase 72 cloud sample before Newman" -Command {
        $uploadArgs = @(
            "-NoProfile",
            "-ExecutionPolicy", "Bypass",
            "-File", "./scripts/local/upload-cloud-sample-phase72.ps1",
            "-Region", $Region,
            "-ManifestPath", $manifestPath
        )
        if (-not [string]::IsNullOrWhiteSpace($Bucket)) {
            $uploadArgs += @("-Bucket", $Bucket)
        }
        powershell @uploadArgs
    }

    if ([string]::IsNullOrWhiteSpace($FrontendUrl)) {
        $FrontendUrl = Get-TraceDeckFrontendUrl
    }
    $FrontendUrl = $FrontendUrl.TrimEnd("/")

    Invoke-TraceDeckLoggedCommand -Label "Run Newman Phase 72 cloud collection" -Command {
        newman run ./postman/tracedeck-cloud-phase72.postman_collection.json --env-var "lambdaUrl=$FrontendUrl" --reporters "cli,json" --reporter-json-export $reportPath
    }

    Write-TraceDeckLog -Level "INFO" -Message "Newman Phase 72 report: $reportPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
