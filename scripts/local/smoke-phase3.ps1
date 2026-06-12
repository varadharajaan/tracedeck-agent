param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase3" -LogRoot "logs/local/smoke" | Out-Null

function Read-GzipText {
    param(
        [string]$Path
    )

    $file = [System.IO.File]::OpenRead($Path)
    try {
        $gzip = [System.IO.Compression.GzipStream]::new($file, [System.IO.Compression.CompressionMode]::Decompress)
        try {
            $reader = [System.IO.StreamReader]::new($gzip)
            try {
                return $reader.ReadToEnd()
            }
            finally {
                $reader.Dispose()
            }
        }
        finally {
            $gzip.Dispose()
        }
    }
    finally {
        $file.Dispose()
    }
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase3/$timestamp"
    $outboxRoot = Join-Path $script:TraceDeckRepoRoot "data/local/outbox/smoke-phase3/$timestamp"
    $historyPath = Join-Path $smokeRoot "Google/Chrome/User Data/Default/History"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $outboxRoot | Out-Null

    Invoke-TraceDeckLoggedCommand -Label "Create browser history fixture" -Command {
        go run ./scripts/tools/browser-fixture --out $historyPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Run local agent with browser history fixture" -Command {
        go run ./agent/cmd/tracedeck-agent run --once --config ./examples/policies/ai-btech-student.yaml --data-dir $smokeRoot --log-dir ./logs/local/agent --outbox-dir $outboxRoot --process-limit 32 --browser-history-path $historyPath --browser-history-limit 16 --archive-once --archive-dry-run --alert-dry-run
    }

    $archiveFiles = @(Get-ChildItem -Path (Join-Path $outboxRoot "archive") -Filter "*.jsonl.gz" -File -ErrorAction SilentlyContinue)
    if (-not $archiveFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected browser smoke archive batch was not created."
        exit 1
    }

    $latestArchive = $archiveFiles | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    $archiveText = Read-GzipText -Path $latestArchive.FullName

    if ($archiveText -notmatch [regex]::Escape('"Type":"browser.domain.observed"')) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected browser.domain.observed event in archive."
        exit 1
    }
    if ($archiveText -notmatch [regex]::Escape('"domain":"youtube.com"')) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected youtube.com domain-only metadata in archive."
        exit 1
    }
    if ($archiveText -match "traceDeckSmoke123" -or $archiveText -match "private title must not persist" -or $archiveText -match "https://") {
        Write-TraceDeckLog -Level "ERROR" -Message "Browser smoke archive leaked raw URL or title data."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Browser smoke archive passed privacy checks: $($latestArchive.FullName)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
