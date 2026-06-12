param(
    [string]$TaskName = "\TraceDeck\TraceDeck Agent",
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$ConfigPath = "examples/policies/ai-btech-student.yaml",
    [string]$DataDir = "data/local",
    [string]$LogDir = "logs/local/agent",
    [string]$OutboxDir = "data/local/outbox",
    [string]$CollectionInterval = "10m",
    [string]$UserId = "$env:USERDOMAIN\$env:USERNAME",
    [string]$TaskXmlPath = "data/local/service-manifests/phase8/windows/tracedeck-agent-task.xml",
    [switch]$BuildAgent,
    [switch]$StartAfterRegister,
    [switch]$NoElevate
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "register-windows-task" -LogRoot "logs/local/service" | Out-Null

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Test-TraceDeckAdmin {
    $identity = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($identity)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Quote-TraceDeckArg {
    param([string]$Value)
    return '"' + ($Value -replace '"', '\"') + '"'
}

try {
    if (-not (Test-TraceDeckAdmin) -and -not $NoElevate) {
        $scriptPath = $MyInvocation.MyCommand.Path
        $args = @(
            "-NoProfile",
            "-ExecutionPolicy", "Bypass",
            "-File", (Quote-TraceDeckArg -Value $scriptPath),
            "-TaskName", (Quote-TraceDeckArg -Value $TaskName),
            "-AgentPath", (Quote-TraceDeckArg -Value $AgentPath),
            "-ConfigPath", (Quote-TraceDeckArg -Value $ConfigPath),
            "-DataDir", (Quote-TraceDeckArg -Value $DataDir),
            "-LogDir", (Quote-TraceDeckArg -Value $LogDir),
            "-OutboxDir", (Quote-TraceDeckArg -Value $OutboxDir),
            "-CollectionInterval", (Quote-TraceDeckArg -Value $CollectionInterval),
            "-UserId", (Quote-TraceDeckArg -Value $UserId),
            "-TaskXmlPath", (Quote-TraceDeckArg -Value $TaskXmlPath),
            "-NoElevate"
        )
        if ($BuildAgent) {
            $args += "-BuildAgent"
        }
        if ($StartAfterRegister) {
            $args += "-StartAfterRegister"
        }

        Write-TraceDeckLog -Level "INFO" -Message "Requesting UAC elevation to register scheduled task '$TaskName'."
        Start-Process -FilePath "powershell.exe" -ArgumentList ($args -join " ") -Verb RunAs -Wait
        Complete-TraceDeckScriptLog
        exit 0
    }

    $resolvedAgentPath = Resolve-TraceDeckPath -PathValue $AgentPath
    $resolvedTaskXmlPath = Resolve-TraceDeckPath -PathValue $TaskXmlPath

    if ($BuildAgent) {
        $agentParent = Split-Path -Parent $resolvedAgentPath
        New-Item -ItemType Directory -Force -Path $agentParent | Out-Null
        Invoke-TraceDeckLoggedCommand -Label "Build scheduled-task agent executable" -Command {
            go build -trimpath -ldflags "-H=windowsgui" -o $resolvedAgentPath ./agent/cmd/tracedeck-agent
        }
    }

    if (-not (Test-Path -LiteralPath $resolvedAgentPath)) {
        throw "Agent executable does not exist: $resolvedAgentPath. Rerun with -BuildAgent or provide -AgentPath."
    }

    Invoke-TraceDeckLoggedCommand -Label "Render Windows scheduled-task XML" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1 `
            -AgentPath $resolvedAgentPath `
            -ConfigPath $ConfigPath `
            -DataDir $DataDir `
            -LogDir $LogDir `
            -OutboxDir $OutboxDir `
            -CollectionInterval $CollectionInterval `
            -UserId $UserId `
            -OutputPath $resolvedTaskXmlPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Register Windows scheduled task" -Command {
        schtasks.exe /Create /TN $TaskName /XML $resolvedTaskXmlPath /F
    }

    if ($StartAfterRegister) {
        Invoke-TraceDeckLoggedCommand -Label "Start Windows scheduled task" -Command {
            schtasks.exe /Run /TN $TaskName
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Query Windows scheduled task" -Command {
        schtasks.exe /Query /TN $TaskName /V /FO LIST
    }

    Write-TraceDeckLog -Level "INFO" -Message "Registered scheduled task '$TaskName'. It starts TraceDeck at user logon after reboot."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
