param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-windows-task-template" -LogRoot "logs/local/test" | Out-Null

try {
    $outputPath = "data/local/service-manifests/phase8/windows/test-tracedeck-agent-task.xml"
    Invoke-TraceDeckLoggedCommand -Label "Render Windows scheduled task XML" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/render-windows-task.ps1 -OutputPath $outputPath
    }

    $resolvedOutputPath = Join-Path $script:TraceDeckRepoRoot $outputPath
    [xml]$xml = Get-Content -Raw -Path $resolvedOutputPath
    $namespace = New-Object System.Xml.XmlNamespaceManager($xml.NameTable)
    $namespace.AddNamespace("task", "http://schemas.microsoft.com/windows/2004/02/mit/task")

    $command = $xml.SelectSingleNode("//task:Actions/task:Exec/task:Command", $namespace)
    $arguments = $xml.SelectSingleNode("//task:Actions/task:Exec/task:Arguments", $namespace)
    $hidden = $xml.SelectSingleNode("//task:Settings/task:Hidden", $namespace)
    $logon = $xml.SelectSingleNode("//task:Triggers/task:LogonTrigger", $namespace)

    if (-not $command -or $command.InnerText -notmatch "tracedeck-agent\.exe$") {
        Write-TraceDeckLog -Level "ERROR" -Message "Task XML command is missing or invalid."
        exit 1
    }
    if (-not $arguments -or $arguments.InnerText -notmatch "run --config" -or $arguments.InnerText -notmatch "--max-cycles 0") {
        Write-TraceDeckLog -Level "ERROR" -Message "Task XML arguments do not start continuous agent mode."
        exit 1
    }
    if (-not $hidden -or $hidden.InnerText -ne "true") {
        Write-TraceDeckLog -Level "ERROR" -Message "Task XML should be hidden by Task Scheduler to avoid a console flicker."
        exit 1
    }
    if (-not $logon) {
        Write-TraceDeckLog -Level "ERROR" -Message "Task XML must include a logon trigger for reboot persistence."
        exit 1
    }

    Invoke-TraceDeckLoggedCommand -Label "Query missing Windows task with allowed missing status" -Command {
        powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-windows-task-status.ps1 -TaskName "\TraceDeck\TraceDeck Agent Template Smoke Missing" -AllowMissing
    }

    Write-TraceDeckLog -Level "INFO" -Message "Windows scheduled task template smoke passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
