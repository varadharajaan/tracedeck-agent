param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-autostart-assurance" -LogRoot "logs/local/test" | Out-Null

function Get-RequiredNode {
    param(
        [xml]$Xml,
        [System.Xml.XmlNamespaceManager]$Namespace,
        [string]$XPath
    )

    $node = $Xml.SelectSingleNode($XPath, $Namespace)
    if (-not $node) {
        throw "Missing required scheduled-task XML node: $XPath"
    }
    return $node
}

function Assert-NodeText {
    param(
        [xml]$Xml,
        [System.Xml.XmlNamespaceManager]$Namespace,
        [string]$XPath,
        [string]$Expected
    )

    $node = Get-RequiredNode -Xml $Xml -Namespace $Namespace -XPath $XPath
    if ($node.InnerText -ne $Expected) {
        throw "Expected $XPath to be '$Expected', got '$($node.InnerText)'"
    }
}

try {
    $outputPath = "data/local/service-manifests/phase39/windows/tracedeck-agent-task.xml"
    $missingStatusPath = "data/local/service-status/phase39/missing-task.json"

    Invoke-TraceDeckLoggedCommand -Label "Render Phase 39 Windows scheduled task XML" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1 -OutputPath $outputPath
    }

    $resolvedOutputPath = Join-Path $script:TraceDeckRepoRoot $outputPath
    $raw = Get-Content -Raw -Path $resolvedOutputPath
    if ($raw -match "\{\{") {
        throw "Rendered scheduled-task XML still contains unresolved placeholders."
    }

    [xml]$xml = $raw
    $namespace = New-Object System.Xml.XmlNamespaceManager($xml.NameTable)
    $namespace.AddNamespace("task", "http://schemas.microsoft.com/windows/2004/02/mit/task")

    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:Hidden" -Expected "true"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Triggers/task:LogonTrigger/task:Enabled" -Expected "true"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Triggers/task:LogonTrigger/task:Delay" -Expected "PT30S"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:StartWhenAvailable" -Expected "true"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:MultipleInstancesPolicy" -Expected "IgnoreNew"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:ExecutionTimeLimit" -Expected "PT0S"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:RestartOnFailure/task:Interval" -Expected "PT5M"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:RestartOnFailure/task:Count" -Expected "3"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:DisallowStartIfOnBatteries" -Expected "false"
    Assert-NodeText -Xml $xml -Namespace $namespace -XPath "//task:Settings/task:StopIfGoingOnBatteries" -Expected "false"

    $command = Get-RequiredNode -Xml $xml -Namespace $namespace -XPath "//task:Actions/task:Exec/task:Command"
    $arguments = Get-RequiredNode -Xml $xml -Namespace $namespace -XPath "//task:Actions/task:Exec/task:Arguments"
    if ($command.InnerText -notmatch "wscript\.exe$") {
        throw "Expected scheduled-task command to target wscript.exe."
    }
    foreach ($expected in @("run-agent-task-hidden.vbs", "run-agent-task.ps1", "-AgentPath", "-CollectionInterval", "10m", "-LogDir", "-OutboxDir", "-PidPath")) {
        if ($arguments.InnerText -notmatch [regex]::Escape($expected)) {
            throw "Expected scheduled-task arguments to include '$expected'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Query missing Windows task as typed status JSON" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1 `
            -TaskName "\TraceDeck\TraceDeck Agent Phase 39 Missing" `
            -AllowMissing `
            -OutputPath $missingStatusPath
    }

    $resolvedStatusPath = Join-Path $script:TraceDeckRepoRoot $missingStatusPath
    $status = Get-Content -Raw -Path $resolvedStatusPath | ConvertFrom-Json
    if ($status.present -ne $false -or $status.state -ne "missing" -or -not $status.PSObject.Properties.Name.Contains("last_task_result")) {
        throw "Expected missing task status JSON with present=false, state=missing, and last_task_result field."
    }

    Invoke-TraceDeckLoggedCommand -Label "Dry-run Windows service install plan" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action install -DryRun
    }
    Invoke-TraceDeckLoggedCommand -Label "Dry-run Windows service status plan" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/manage-agent-service.ps1 -Platform windows -Action status -DryRun
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 39 autostart assurance passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
