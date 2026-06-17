param(
    [ValidateSet("install", "start", "stop", "status", "uninstall")]
    [string]$Action = "status",
    [ValidateSet("auto", "windows", "linux", "darwin")]
    [string]$Platform = "auto",
    [string]$AgentPath = "data/local/install/windows/tracedeck-agent.exe",
    [string]$ConfigPath = "examples/policies/ai-btech-student.yaml",
    [string]$DataDir = "data/local",
    [string]$LogDir = "logs/local/agent",
    [string]$OutboxDir = "data/local/outbox",
    [string]$CollectionInterval = "10m",
    [string]$PidPath = "",
    [string]$TaskName = "\TraceDeck\TraceDeck Agent",
    [string]$LaunchdLabel = "io.tracedeck.agent",
    [string]$SystemdUnit = "tracedeck-agent.service",
    [string]$ServiceUser = "tracedeck",
    [string]$WorkingDir = "",
    [switch]$BuildAgent,
    [switch]$DryRun
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "manage-agent-service" -LogRoot "logs/local/service" | Out-Null

function Get-TraceDeckPlatform {
    if ($Platform -ne "auto") {
        return $Platform
    }
    if ([System.Environment]::OSVersion.Platform -eq [System.PlatformID]::Win32NT) {
        return "windows"
    }
    $uname = ""
    try {
        $uname = (& uname -s 2>$null).Trim().ToLowerInvariant()
    }
    catch {
        $uname = ""
    }
    if ($uname -eq "darwin") {
        return "darwin"
    }
    return "linux"
}

function Resolve-TraceDeckPath {
    param([string]$PathValue)

    if ([System.IO.Path]::IsPathRooted($PathValue)) {
        return [System.IO.Path]::GetFullPath($PathValue)
    }
    return [System.IO.Path]::GetFullPath((Join-Path $script:TraceDeckRepoRoot $PathValue))
}

function Add-TraceDeckCommand {
    param(
        [System.Collections.Generic.List[string]]$Commands,
        [string]$Command
    )
    $Commands.Add($Command)
    Write-TraceDeckLog -Level "INFO" -Message "Planned service command: $Command"
}

function Save-TraceDeckPlan {
    param(
        [string]$ResolvedPlatform,
        [System.Collections.Generic.List[string]]$Commands
    )

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $planRoot = Join-Path $script:TraceDeckRepoRoot "data/local/service-actions/phase15/$timestamp"
    New-Item -ItemType Directory -Force -Path $planRoot | Out-Null
    $planPath = Join-Path $planRoot "$ResolvedPlatform-$Action.json"
    [pscustomobject]@{
        action = $Action
        platform = $ResolvedPlatform
        dry_run = [bool]$DryRun
        commands = @($Commands)
    } | ConvertTo-Json -Depth 5 | Set-Content -Path $planPath -Encoding utf8
    Write-TraceDeckLog -Level "INFO" -Message "Service action plan written: $planPath"
}

function Invoke-TraceDeckServiceCommand {
    param(
        [string]$Label,
        [scriptblock]$Command
    )

    if ($DryRun) {
        return
    }
    Invoke-TraceDeckLoggedCommand -Label $Label -Command $Command
}

function Invoke-WindowsServiceAction {
    param([System.Collections.Generic.List[string]]$Commands)

    switch ($Action) {
        "install" {
            Add-TraceDeckCommand -Commands $Commands -Command "scripts/local/register-windows-task.ps1 -BuildAgent:$BuildAgent -StartAfterRegister"
            Invoke-TraceDeckServiceCommand -Label "Register Windows scheduled task" -Command {
                $args = @(
                    "-TaskName", $TaskName,
                    "-AgentPath", $AgentPath,
                    "-ConfigPath", $ConfigPath,
                    "-DataDir", $DataDir,
                    "-LogDir", $LogDir,
                    "-OutboxDir", $OutboxDir,
                    "-CollectionInterval", $CollectionInterval,
                    "-PidPath", $PidPath,
                    "-StartAfterRegister"
                )
                if ($BuildAgent) {
                    $args += "-BuildAgent"
                }
                powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/register-windows-task.ps1 @args
            }
        }
        "start" {
            Add-TraceDeckCommand -Commands $Commands -Command "schtasks.exe /Run /TN `"$TaskName`""
            Invoke-TraceDeckServiceCommand -Label "Start Windows scheduled task" -Command {
                schtasks.exe /Run /TN $TaskName
            }
        }
        "stop" {
            Add-TraceDeckCommand -Commands $Commands -Command "schtasks.exe /End /TN `"$TaskName`""
            Invoke-TraceDeckServiceCommand -Label "Stop Windows scheduled task" -Command {
                schtasks.exe /End /TN $TaskName
            }
        }
        "status" {
            Add-TraceDeckCommand -Commands $Commands -Command "scripts/local/get-windows-task-status.ps1 -AllowMissing"
            Invoke-TraceDeckServiceCommand -Label "Query Windows scheduled task" -Command {
                powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1 -TaskName $TaskName -AllowMissing
            }
        }
        "uninstall" {
            Add-TraceDeckCommand -Commands $Commands -Command "schtasks.exe /Delete /TN `"$TaskName`" /F"
            Invoke-TraceDeckServiceCommand -Label "Delete Windows scheduled task" -Command {
                schtasks.exe /Delete /TN $TaskName /F
            }
        }
    }
}

function Invoke-LinuxServiceAction {
    param([System.Collections.Generic.List[string]]$Commands)

    $agent = if ([System.IO.Path]::IsPathRooted($AgentPath)) { $AgentPath } else { "/opt/tracedeck/bin/tracedeck-agent" }
    $config = if ([System.IO.Path]::IsPathRooted($ConfigPath)) { $ConfigPath } else { "/etc/tracedeck/agent.yaml" }
    $log = if ([System.IO.Path]::IsPathRooted($LogDir)) { $LogDir } else { "/var/log/tracedeck" }
    $work = if ($WorkingDir) { $WorkingDir } else { "/opt/tracedeck" }
    $manifestRoot = "data/local/service-manifests/phase15"
    $manifestPath = Join-Path (Resolve-TraceDeckPath -PathValue $manifestRoot) "linux/tracedeck-agent.service"

    switch ($Action) {
        "install" {
            Add-TraceDeckCommand -Commands $Commands -Command "render-service-manifests.ps1 - linux systemd"
            Invoke-TraceDeckServiceCommand -Label "Render Linux systemd manifest" -Command {
                powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1 -AgentPath $agent -ConfigPath $config -LogDir $log -WorkingDir $work -User $ServiceUser -OutputRoot $manifestRoot
            }
            Add-TraceDeckCommand -Commands $Commands -Command "sudo install -m 0644 `"$manifestPath`" /etc/systemd/system/$SystemdUnit"
            Invoke-TraceDeckServiceCommand -Label "Install Linux systemd manifest" -Command {
                sudo install -m 0644 $manifestPath "/etc/systemd/system/$SystemdUnit"
            }
            Add-TraceDeckCommand -Commands $Commands -Command "sudo systemctl daemon-reload"
            Invoke-TraceDeckServiceCommand -Label "Reload Linux systemd" -Command {
                sudo systemctl daemon-reload
            }
            Add-TraceDeckCommand -Commands $Commands -Command "sudo systemctl enable --now $SystemdUnit"
            Invoke-TraceDeckServiceCommand -Label "Enable and start Linux service" -Command {
                sudo systemctl enable --now $SystemdUnit
            }
        }
        "start" {
            Add-TraceDeckCommand -Commands $Commands -Command "sudo systemctl start $SystemdUnit"
            Invoke-TraceDeckServiceCommand -Label "Start Linux service" -Command {
                sudo systemctl start $SystemdUnit
            }
        }
        "stop" {
            Add-TraceDeckCommand -Commands $Commands -Command "sudo systemctl stop $SystemdUnit"
            Invoke-TraceDeckServiceCommand -Label "Stop Linux service" -Command {
                sudo systemctl stop $SystemdUnit
            }
        }
        "status" {
            Add-TraceDeckCommand -Commands $Commands -Command "systemctl status $SystemdUnit --no-pager"
            Invoke-TraceDeckServiceCommand -Label "Query Linux service" -Command {
                systemctl status $SystemdUnit --no-pager
            }
        }
        "uninstall" {
            Add-TraceDeckCommand -Commands $Commands -Command "sudo systemctl disable --now $SystemdUnit"
            Invoke-TraceDeckServiceCommand -Label "Disable Linux service" -Command {
                sudo systemctl disable --now $SystemdUnit
            }
            Add-TraceDeckCommand -Commands $Commands -Command "sudo rm -f /etc/systemd/system/$SystemdUnit"
            Invoke-TraceDeckServiceCommand -Label "Remove Linux systemd manifest" -Command {
                sudo rm -f "/etc/systemd/system/$SystemdUnit"
            }
            Add-TraceDeckCommand -Commands $Commands -Command "sudo systemctl daemon-reload"
            Invoke-TraceDeckServiceCommand -Label "Reload Linux systemd after uninstall" -Command {
                sudo systemctl daemon-reload
            }
        }
    }
}

function Invoke-DarwinServiceAction {
    param([System.Collections.Generic.List[string]]$Commands)

    $agent = if ([System.IO.Path]::IsPathRooted($AgentPath)) { $AgentPath } else { "/opt/tracedeck/bin/tracedeck-agent" }
    $config = if ([System.IO.Path]::IsPathRooted($ConfigPath)) { $ConfigPath } else { "/etc/tracedeck/agent.yaml" }
    $log = if ([System.IO.Path]::IsPathRooted($LogDir)) { $LogDir } else { "/var/log/tracedeck" }
    $work = if ($WorkingDir) { $WorkingDir } else { "/opt/tracedeck" }
    $manifestRoot = "data/local/service-manifests/phase15"
    $manifestPath = Join-Path (Resolve-TraceDeckPath -PathValue $manifestRoot) "darwin/io.tracedeck.agent.plist"
    $targetPath = "~/Library/LaunchAgents/$LaunchdLabel.plist"
    $launchdUserDomain = 'gui/$(id -u)'
    $launchdService = "$launchdUserDomain/$LaunchdLabel"

    switch ($Action) {
        "install" {
            Add-TraceDeckCommand -Commands $Commands -Command "render-service-manifests.ps1 - darwin launchd"
            Invoke-TraceDeckServiceCommand -Label "Render macOS launchd manifest" -Command {
                powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-service-manifests.ps1 -AgentPath $agent -ConfigPath $config -LogDir $log -WorkingDir $work -OutputRoot $manifestRoot
            }
            Add-TraceDeckCommand -Commands $Commands -Command "mkdir -p ~/Library/LaunchAgents"
            Invoke-TraceDeckServiceCommand -Label "Create macOS LaunchAgents directory" -Command {
                New-Item -ItemType Directory -Force -Path (Join-Path $HOME "Library/LaunchAgents") | Out-Null
            }
            Add-TraceDeckCommand -Commands $Commands -Command "cp `"$manifestPath`" $targetPath"
            Invoke-TraceDeckServiceCommand -Label "Install macOS launchd manifest" -Command {
                $actualTargetPath = Join-Path $HOME "Library/LaunchAgents/$LaunchdLabel.plist"
                Copy-Item -LiteralPath $manifestPath -Destination $actualTargetPath -Force
            }
            Add-TraceDeckCommand -Commands $Commands -Command "launchctl bootstrap $launchdUserDomain $targetPath"
            Invoke-TraceDeckServiceCommand -Label "Bootstrap macOS launchd service" -Command {
                $actualTargetPath = Join-Path $HOME "Library/LaunchAgents/$LaunchdLabel.plist"
                $uid = (& id -u).Trim()
                launchctl bootstrap "gui/$uid" $actualTargetPath
            }
            Add-TraceDeckCommand -Commands $Commands -Command "launchctl enable $launchdService"
            Invoke-TraceDeckServiceCommand -Label "Enable macOS launchd service" -Command {
                $uid = (& id -u).Trim()
                launchctl enable "gui/$uid/$LaunchdLabel"
            }
        }
        "start" {
            Add-TraceDeckCommand -Commands $Commands -Command "launchctl kickstart -k $launchdService"
            Invoke-TraceDeckServiceCommand -Label "Start macOS launchd service" -Command {
                $uid = (& id -u).Trim()
                launchctl kickstart -k "gui/$uid/$LaunchdLabel"
            }
        }
        "stop" {
            Add-TraceDeckCommand -Commands $Commands -Command "launchctl kill TERM $launchdService"
            Invoke-TraceDeckServiceCommand -Label "Stop macOS launchd service" -Command {
                $uid = (& id -u).Trim()
                launchctl kill TERM "gui/$uid/$LaunchdLabel"
            }
        }
        "status" {
            Add-TraceDeckCommand -Commands $Commands -Command "launchctl print $launchdService"
            Invoke-TraceDeckServiceCommand -Label "Query macOS launchd service" -Command {
                $uid = (& id -u).Trim()
                launchctl print "gui/$uid/$LaunchdLabel"
            }
        }
        "uninstall" {
            Add-TraceDeckCommand -Commands $Commands -Command "launchctl bootout $launchdUserDomain $targetPath"
            Invoke-TraceDeckServiceCommand -Label "Unload macOS launchd service" -Command {
                $actualTargetPath = Join-Path $HOME "Library/LaunchAgents/$LaunchdLabel.plist"
                $uid = (& id -u).Trim()
                launchctl bootout "gui/$uid" $actualTargetPath
            }
            Add-TraceDeckCommand -Commands $Commands -Command "rm -f $targetPath"
            Invoke-TraceDeckServiceCommand -Label "Remove macOS launchd manifest" -Command {
                $actualTargetPath = Join-Path $HOME "Library/LaunchAgents/$LaunchdLabel.plist"
                Remove-Item -LiteralPath $actualTargetPath -Force -ErrorAction SilentlyContinue
            }
        }
    }
}

try {
    $resolvedPlatform = Get-TraceDeckPlatform
    $commands = [System.Collections.Generic.List[string]]::new()
    Write-TraceDeckLog -Level "INFO" -Message "Managing TraceDeck service action=$Action platform=$resolvedPlatform dry_run=$([bool]$DryRun)"

    switch ($resolvedPlatform) {
        "windows" { Invoke-WindowsServiceAction -Commands $commands }
        "linux" { Invoke-LinuxServiceAction -Commands $commands }
        "darwin" { Invoke-DarwinServiceAction -Commands $commands }
        default { throw "unsupported platform: $resolvedPlatform" }
    }

    Save-TraceDeckPlan -ResolvedPlatform $resolvedPlatform -Commands $commands
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
