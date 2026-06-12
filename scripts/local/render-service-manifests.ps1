param(
    [string]$AgentPath = "/opt/tracedeck/bin/tracedeck-agent",
    [string]$ConfigPath = "/etc/tracedeck/agent.yaml",
    [string]$LogDir = "/var/log/tracedeck",
    [string]$WorkingDir = "/opt/tracedeck",
    [string]$User = "tracedeck",
    [string]$OutputRoot = "data/local/service-manifests/phase7"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "render-service-manifests" -LogRoot "logs/local/service" | Out-Null

function Expand-TraceDeckTemplate {
    param(
        [string]$TemplatePath,
        [string]$OutputPath,
        [hashtable]$Values
    )

    $content = Get-Content -Raw -Path $TemplatePath
    foreach ($key in $Values.Keys) {
        $content = $content.Replace("{{$key}}", [string]$Values[$key])
    }
    $parent = Split-Path -Parent $OutputPath
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    Set-Content -Path $OutputPath -Value $content -Encoding utf8
}

try {
    $resolvedOutputRoot = Join-Path $script:TraceDeckRepoRoot $OutputRoot
    $values = @{
        AGENT_PATH = $AgentPath
        CONFIG_PATH = $ConfigPath
        LOG_DIR = $LogDir
        WORKING_DIR = $WorkingDir
        USER = $User
    }

    Write-TraceDeckLog -Level "INFO" -Message "Starting: Render macOS launchd manifest"
    Expand-TraceDeckTemplate `
        -TemplatePath (Join-Path $script:TraceDeckRepoRoot "deployments/service/darwin/io.tracedeck.agent.plist.tmpl") `
        -OutputPath (Join-Path $resolvedOutputRoot "darwin/io.tracedeck.agent.plist") `
        -Values $values
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Render macOS launchd manifest"

    Write-TraceDeckLog -Level "INFO" -Message "Starting: Render Linux systemd manifest"
    Expand-TraceDeckTemplate `
        -TemplatePath (Join-Path $script:TraceDeckRepoRoot "deployments/service/linux/tracedeck-agent.service.tmpl") `
        -OutputPath (Join-Path $resolvedOutputRoot "linux/tracedeck-agent.service") `
        -Values $values
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Render Linux systemd manifest"

    $darwinOutput = Join-Path $resolvedOutputRoot "darwin/io.tracedeck.agent.plist"
    $linuxOutput = Join-Path $resolvedOutputRoot "linux/tracedeck-agent.service"
    if (-not (Test-Path $darwinOutput) -or -not (Test-Path $linuxOutput)) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected service manifest outputs were not created."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Rendered service manifests under $resolvedOutputRoot"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
