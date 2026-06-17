param(
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$ConfigPath = "examples/policies/ai-btech-student.yaml",
    [string]$DataDir = "data/local",
    [string]$LogDir = "logs/local/agent",
    [string]$OutboxDir = "data/local/outbox",
    [string]$CollectionInterval = "10m",
    [string]$PidPath = "",
    [string]$RunnerScriptPath = "scripts/local/run-agent-task.ps1",
    [string]$HiddenLauncherPath = "scripts/local/run-agent-task-hidden.vbs",
    [string]$ExtraArgs = "",
    [ValidateSet("HighestAvailable", "LeastPrivilege")]
    [string]$RunLevel = "HighestAvailable",
    [string]$UserId = "$env:USERDOMAIN\$env:USERNAME",
    [string]$OutputPath = "data/local/service-manifests/phase8/windows/tracedeck-agent-task.xml"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "render-windows-task" -LogRoot "logs/local/service" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Escape-XmlValue {
    param([string]$Value)
    return [System.Security.SecurityElement]::Escape($Value)
}

try {
    $templatePath = Join-Path $script:TraceDeckRepoRoot "deployments/service/windows/tracedeck-agent-task.xml.tmpl"
    $resolvedOutputPath = Resolve-TraceDeckPath -PathValue $OutputPath
    $resolvedAgentPath = Resolve-TraceDeckPath -PathValue $AgentPath
    $resolvedConfigPath = Resolve-TraceDeckPath -PathValue $ConfigPath
    $resolvedDataDir = Resolve-TraceDeckPath -PathValue $DataDir
    $resolvedLogDir = Resolve-TraceDeckPath -PathValue $LogDir
    $resolvedOutboxDir = Resolve-TraceDeckPath -PathValue $OutboxDir
    $resolvedWorkingDir = [System.IO.Path]::GetFullPath($script:TraceDeckRepoRoot)
    $null = $PidPath
    $null = $RunnerScriptPath
    $null = $HiddenLauncherPath

    $values = @{
        USER_ID = $UserId
        AGENT_PATH = $resolvedAgentPath
        CONFIG_PATH = $resolvedConfigPath
        DATA_DIR = $resolvedDataDir
        LOG_DIR = $resolvedLogDir
        OUTBOX_DIR = $resolvedOutboxDir
        COLLECTION_INTERVAL = $CollectionInterval
        EXTRA_ARGS = $ExtraArgs
        RUN_LEVEL = $RunLevel
        WORKING_DIR = $resolvedWorkingDir
    }

    $content = Get-Content -Raw -Path $templatePath
    foreach ($key in $values.Keys) {
        $content = $content.Replace("{{$key}}", (Escape-XmlValue -Value ([string]$values[$key])))
    }

    $parent = Split-Path -Parent $resolvedOutputPath
    New-Item -ItemType Directory -Force -Path $parent | Out-Null
    Set-Content -Path $resolvedOutputPath -Value $content -Encoding Unicode

    [xml]$xml = Get-Content -Raw -Path $resolvedOutputPath
    if ($xml.Task.Actions.Exec.Command -ne $resolvedAgentPath) {
        Write-TraceDeckLog -Level "ERROR" -Message "Rendered task XML command path mismatch."
        exit 1
    }
    foreach ($requiredValue in @("run", "--config", $resolvedConfigPath, "--data-dir", $resolvedDataDir, "--log-dir", $resolvedLogDir, "--outbox-dir", $resolvedOutboxDir, "--collection-interval", $CollectionInterval, "--max-cycles", "0")) {
        if ($xml.Task.Actions.Exec.Arguments -notmatch [regex]::Escape($requiredValue)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Rendered task XML missing required agent argument: $requiredValue"
            exit 1
        }
    }
    if ($xml.Task.Actions.Exec.Command -match "powershell|wscript|cscript|cmd\.exe" -or $xml.Task.Actions.Exec.Arguments -match "powershell|run-agent-task|wscript|cscript|cmd\.exe") {
        Write-TraceDeckLog -Level "ERROR" -Message "Rendered task XML should launch the GUI agent directly without script hosts."
        exit 1
    }

    if ((Get-Content -Raw -Path $resolvedOutputPath) -match "\{\{") {
        Write-TraceDeckLog -Level "ERROR" -Message "Rendered task XML still contains an unresolved placeholder."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Rendered Windows scheduled task XML: $resolvedOutputPath"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
