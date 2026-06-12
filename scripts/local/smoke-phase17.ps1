param(
    [string]$Addr = "127.0.0.1:18252"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase17" -LogRoot "logs/local/smoke" | Out-Null

$smtp = $null
$sleeper = $null
$oldHost = $env:TRACEDECK_SMTP_HOST
$oldPort = $env:TRACEDECK_SMTP_PORT
$oldUser = $env:TRACEDECK_SMTP_USERNAME
$oldPassword = $env:TRACEDECK_SMTP_PASSWORD
$oldTLS = $env:TRACEDECK_SMTP_SERVER_TLS

function Wait-TraceDeckReadyFile {
    param([string]$Path)

    $deadline = (Get-Date).AddSeconds(45)
    while ((Get-Date) -lt $deadline) {
        if (Test-Path -LiteralPath $Path) {
            return
        }
        Start-Sleep -Milliseconds 250
    }
    throw "Fake SMTP server did not become ready: $Path"
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase17/$timestamp"
    $outboxRoot = Join-Path $smokeRoot "outbox"
    $smtpRoot = Join-Path $smokeRoot "smtp"
    $readyFile = Join-Path $smokeRoot "fake-smtp.ready"
    $fakeSMTPPath = Join-Path $smokeRoot "fake-smtp.exe"
    New-Item -ItemType Directory -Force -Path $smokeRoot, $outboxRoot, $smtpRoot | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Build fake SMTP helper" -Command {
        go build -trimpath -o $fakeSMTPPath ./scripts/tools/fake-smtp
    }

    $smtpArgs = "--addr `"$Addr`" --out-dir `"$smtpRoot`" --ready-file `"$readyFile`""
    $smtp = Start-Process -FilePath $fakeSMTPPath -ArgumentList $smtpArgs -WorkingDirectory $script:TraceDeckRepoRoot -WindowStyle Hidden -RedirectStandardOutput (Join-Path $smokeRoot "fake-smtp.out.log") -RedirectStandardError (Join-Path $smokeRoot "fake-smtp.err.log") -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started fake SMTP server pid=$($smtp.Id) addr=$Addr"
    Wait-TraceDeckReadyFile -Path $readyFile

    $policyPath = Join-Path $smokeRoot "phase17-policy.yaml"
    $policy = Get-Content -Raw (Join-Path $script:TraceDeckRepoRoot "examples/policies/ai-btech-student.yaml")
    $policy = $policy -replace "provider: ses", "provider: smtp"
    $policy = $policy -replace "blocked_apps:\r?\n", "blocked_apps:`r`n  - powershell.exe`r`n"
    Set-Content -Path $policyPath -Value $policy

    $env:TRACEDECK_SMTP_HOST = ($Addr -split ":")[0]
    $env:TRACEDECK_SMTP_PORT = ($Addr -split ":")[1]
    $env:TRACEDECK_SMTP_USERNAME = ""
    $env:TRACEDECK_SMTP_PASSWORD = ""
    $env:TRACEDECK_SMTP_SERVER_TLS = "false"

    $sleeper = Start-Process -FilePath "powershell" -ArgumentList @("-NoProfile", "-Command", "Start-Sleep -Seconds 45") -WindowStyle Hidden -PassThru
    Write-TraceDeckLog -Level "INFO" -Message "Started Phase 17 sleeper process: $($sleeper.Id)"

    Invoke-TraceDeckLoggedCommand -Label "Run agent with provider-backed SMTP alert delivery" -Command {
        go run ./agent/cmd/tracedeck-agent run --once --config $policyPath --data-dir $smokeRoot --log-dir ./logs/local/agent --outbox-dir $outboxRoot --process-limit 512 --archive-once --archive-dry-run --alert-once --alert-dry-run=false --disable-browser-history
    }

    $messages = @(Get-ChildItem -Path $smtpRoot -Filter "*.eml" -File -ErrorAction SilentlyContinue)
    if ($messages.Count -lt 1) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected fake SMTP server to capture at least one email."
        exit 1
    }
    $latestMessage = $messages | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    $messageText = Get-Content -Raw -Path $latestMessage.FullName
    foreach ($expected in @("TraceDeck alert", "blocked_app_opened", "powershell", "varathu09@gmail.com")) {
        if ($messageText -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Expected SMTP message to contain: $expected"
            exit 1
        }
    }
    if ($messageText -match "https://" -or $messageText -match "TRACEDECK_SMTP_PASSWORD") {
        Write-TraceDeckLog -Level "ERROR" -Message "SMTP message leaked forbidden URL or secret marker."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 17 SMTP alert smoke passed: $($latestMessage.FullName)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
finally {
    $env:TRACEDECK_SMTP_HOST = $oldHost
    $env:TRACEDECK_SMTP_PORT = $oldPort
    $env:TRACEDECK_SMTP_USERNAME = $oldUser
    $env:TRACEDECK_SMTP_PASSWORD = $oldPassword
    $env:TRACEDECK_SMTP_SERVER_TLS = $oldTLS

    if ($sleeper -and -not $sleeper.HasExited) {
        Stop-Process -Id $sleeper.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped Phase 17 sleeper process: $($sleeper.Id)"
    }
    if ($smtp -and -not $smtp.HasExited) {
        Stop-Process -Id $smtp.Id -Force
        Write-TraceDeckLog -Level "INFO" -Message "Stopped fake SMTP server pid=$($smtp.Id)"
    }
}
