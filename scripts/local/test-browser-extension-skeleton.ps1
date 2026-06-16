param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-browser-extension-skeleton" -LogRoot "logs/local/test" | Out-Null

try {
    $extensionRoot = Join-Path $script:TraceDeckRepoRoot "browser-extension"
    $manifestPath = Join-Path $extensionRoot "manifest.json"
    $backgroundPath = Join-Path $extensionRoot "src/background.js"
    $privacyCorePath = Join-Path $extensionRoot "src/privacy-core.js"
    $optionsPath = Join-Path $extensionRoot "src/options.html"
    $nodeTestPath = Join-Path $extensionRoot "test/privacy-core.test.js"

    foreach ($path in @($manifestPath, $backgroundPath, $privacyCorePath, $optionsPath, $nodeTestPath)) {
        if (-not (Test-Path $path)) {
            throw "Missing browser extension file: $path"
        }
    }

    $manifest = Get-Content -Path $manifestPath -Raw | ConvertFrom-Json
    if ([int]$manifest.manifest_version -ne 3) {
        throw "Browser extension manifest must use manifest_version 3."
    }
    if ($manifest.background.service_worker -ne "src/background.js") {
        throw "Browser extension service worker must be src/background.js."
    }
    $permissions = @($manifest.permissions)
    foreach ($required in @("storage", "webNavigation")) {
        if ($permissions -notcontains $required) {
            throw "Browser extension permission '$required' is required."
        }
    }
    foreach ($forbidden in @("tabs", "history", "cookies", "desktopCapture", "tabCapture", "scripting", "downloads", "bookmarks")) {
        if ($permissions -contains $forbidden) {
            throw "Browser extension must not request forbidden permission '$forbidden'."
        }
    }
    $hostPermissions = @($manifest.host_permissions)
    foreach ($requiredHost in @("<all_urls>", "http://127.0.0.1/*", "http://localhost/*")) {
        if ($hostPermissions -notcontains $requiredHost) {
            throw "Browser extension host permission '$requiredHost' is required."
        }
    }

    $background = Get-Content -Path $backgroundPath -Raw
    if ($background -notmatch "hasForbiddenPayloadKeys") {
        throw "Browser extension background worker must check forbidden payload keys before posting."
    }
    if ($background -match "chrome\.tabs" -or $background -match "chrome\.history" -or $background -match "chrome\.cookies") {
        throw "Browser extension worker must not use tabs, history, or cookies APIs."
    }
    $privacyCore = Get-Content -Path $privacyCorePath -Raw
    if ($privacyCore -notmatch "telemetry-events") {
        throw "Browser extension privacy core must target the existing telemetry-events route."
    }

    Invoke-TraceDeckLoggedCommand -Label "Node browser extension privacy contract" -Command {
        node --preserve-symlinks --preserve-symlinks-main ./browser-extension/test/privacy-core.test.js
    }

    Write-TraceDeckLog -Level "INFO" -Message "Browser extension skeleton contract passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
