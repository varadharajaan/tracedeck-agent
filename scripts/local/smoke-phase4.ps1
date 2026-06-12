param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "smoke-phase4" -LogRoot "logs/local/smoke" | Out-Null

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

function Write-Phase4Policy {
    param(
        [string]$SourcePolicyPath,
        [string]$TargetPolicyPath
    )

    $policyText = Get-Content -Path $SourcePolicyPath -Raw
    $policyText = $policyText -replace "min_severity: high", "min_severity: medium"
    if ($policyText -notmatch "(?m)^blocked_domains:") {
        $policyText = $policyText -replace "(?m)^warn_categories:", "blocked_domains:`n  - instagram.com`n`nwarn_categories:"
    }
    if ($policyText -notmatch "(?m)^  blocked_domain_opened:") {
        $policyText = $policyText -replace "(?m)^  media_player_used:", "  blocked_domain_opened:`n    enabled: true`n    severity: high`n  media_player_used:"
    }

    New-Item -ItemType Directory -Force -Path (Split-Path -Parent $TargetPolicyPath) | Out-Null
    Set-Content -Path $TargetPolicyPath -Value $policyText -Encoding utf8
}

try {
    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $smokeRoot = Join-Path $script:TraceDeckRepoRoot "data/local/smoke-phase4/$timestamp"
    $outboxRoot = Join-Path $script:TraceDeckRepoRoot "data/local/outbox/smoke-phase4/$timestamp"
    $historyPath = Join-Path $smokeRoot "Google/Chrome/User Data/Default/History"
    $policyPath = Join-Path $smokeRoot "phase4-policy.yaml"
    $browserCacheDir = Join-Path $smokeRoot "browser-cache"
    New-Item -ItemType Directory -Force -Path $smokeRoot | Out-Null
    New-Item -ItemType Directory -Force -Path $outboxRoot | Out-Null

    Write-TraceDeckLog -Level "INFO" -Message "Starting: Create Phase 4 policy fixture"
    Write-Phase4Policy -SourcePolicyPath "./examples/policies/ai-btech-student.yaml" -TargetPolicyPath $policyPath
    Write-TraceDeckLog -Level "INFO" -Message "Completed: Create Phase 4 policy fixture"

    Invoke-TraceDeckLoggedCommand -Label "Validate Phase 4 policy fixture" -Command {
        go run ./agent/cmd/tracedeck-agent validate-config --config $policyPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Create browser history fixture" -Command {
        go run ./scripts/tools/browser-fixture --out $historyPath
    }

    Invoke-TraceDeckLoggedCommand -Label "Run local agent Phase 4 alert smoke" -Command {
        go run ./agent/cmd/tracedeck-agent run --once --config $policyPath --data-dir $smokeRoot --log-dir ./logs/local/agent --outbox-dir $outboxRoot --process-limit 32 --browser-history-path $historyPath --browser-history-limit 16 --browser-cache-dir $browserCacheDir --archive-once --archive-dry-run --alert-once --alert-dry-run
    }

    $alertFiles = @(Get-ChildItem -Path (Join-Path $outboxRoot "alerts") -Filter "*.json" -File -ErrorAction SilentlyContinue)
    if (-not $alertFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected Phase 4 alert outbox notification was not created."
        exit 1
    }

    $latestAlert = $alertFiles | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    $alertText = Get-Content -Path $latestAlert.FullName -Raw

    foreach ($expected in @('"rule_name": "non_study_youtube"', '"rule_name": "blocked_domain_opened"', '"domain": "youtube.com"', '"domain": "instagram.com"')) {
        if ($alertText -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Missing expected Phase 4 alert content: $expected"
            exit 1
        }
    }
    if ($alertText -match "traceDeckSmoke123" -or $alertText -match "private title must not persist" -or $alertText -match "https://") {
        Write-TraceDeckLog -Level "ERROR" -Message "Phase 4 alert outbox leaked raw URL or title data."
        exit 1
    }

    $archiveFiles = @(Get-ChildItem -Path (Join-Path $outboxRoot "archive") -Filter "*.jsonl.gz" -File -ErrorAction SilentlyContinue)
    if (-not $archiveFiles) {
        Write-TraceDeckLog -Level "ERROR" -Message "Expected Phase 4 archive batch was not created."
        exit 1
    }

    $latestArchive = $archiveFiles | Sort-Object LastWriteTime -Descending | Select-Object -First 1
    $archiveText = Read-GzipText -Path $latestArchive.FullName
    foreach ($expected in @('"Type":"browser.domain.observed"', '"domain":"youtube.com"', '"domain":"instagram.com"')) {
        if ($archiveText -notmatch [regex]::Escape($expected)) {
            Write-TraceDeckLog -Level "ERROR" -Message "Missing expected Phase 4 archive content: $expected"
            exit 1
        }
    }
    if ($archiveText -match "traceDeckSmoke123" -or $archiveText -match "private title must not persist" -or $archiveText -match "https://") {
        Write-TraceDeckLog -Level "ERROR" -Message "Phase 4 archive leaked raw URL or title data."
        exit 1
    }

    Write-TraceDeckLog -Level "INFO" -Message "Phase 4 smoke passed: alerts=$($latestAlert.FullName) archive=$($latestArchive.FullName)"
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
