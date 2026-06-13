param()

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "..\lib\logging.ps1")
Initialize-TraceDeckScriptLog -Name "test-lambda-frontend-contract" -LogRoot "logs/local/test" | Out-Null

try {
    $python = Get-Command python -ErrorAction SilentlyContinue
    if (-not $python) {
        throw "python is not installed or not on PATH"
    }

    Invoke-TraceDeckLoggedCommand -Label "Compile devctl and Lambda frontend Python" -Command {
        $compileRoot = Join-Path $script:TraceDeckRepoRoot "data/local/python-bytecode/lambda-frontend"
        New-Item -ItemType Directory -Force -Path $compileRoot | Out-Null
        @'
import py_compile
from pathlib import Path

targets = [
    ("devctl.py", "devctl.pyc"),
    ("sam-app/frontend_function/app.py", "lambda_frontend.pyc"),
]
out = Path("data/local/python-bytecode/lambda-frontend")
out.mkdir(parents=True, exist_ok=True)
for source, name in targets:
    py_compile.compile(source, cfile=str(out / name), doraise=True)
'@ | python -
    }

    $template = Get-Content -Raw -Path (Join-Path $script:TraceDeckRepoRoot "sam-app/template.yaml")
    foreach ($expected in @(
        "FunctionUrlConfig",
        "AuthType: NONE",
        "TRACEDECK_DATA_BUCKET",
        "TRACEDECK_CACHE_TTL_SECONDS",
        "TRACEDECK_LOCAL_BACKEND_URL",
        "FrontendFunctionUrl"
    )) {
        if ($template -notmatch [regex]::Escape($expected)) {
            throw "SAM template is missing expected marker '$expected'."
        }
    }
    foreach ($forbidden in @("AWS::Serverless::Api", "AWS::ApiGateway", "HttpApi", "ApiGateway")) {
        if ($template -match [regex]::Escape($forbidden)) {
            throw "SAM template contains forbidden API Gateway marker '$forbidden'."
        }
    }

    Invoke-TraceDeckLoggedCommand -Label "Lambda handler local contract" -Command {
        $env:TRACEDECK_DATA_BUCKET = ""
        $env:TRACEDECK_DATA_PREFIX = ""
        $env:PYTHONDONTWRITEBYTECODE = "1"
        @'
import importlib.util
import json
import sys
from pathlib import Path

sys.dont_write_bytecode = True
module_path = Path("sam-app/frontend_function/app.py")
spec = importlib.util.spec_from_file_location("tracedeck_lambda_frontend", module_path)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)

health = module.lambda_handler({"rawPath": "/api/health", "requestContext": {"http": {"method": "GET"}}}, None)
assert health["statusCode"] == 200, health
assert json.loads(health["body"])["status"] == "ok"

html = module.lambda_handler({"rawPath": "/", "requestContext": {"http": {"method": "GET"}}}, None)
assert html["statusCode"] == 200, html
body = html["body"]
for marker in [
    "TraceDeck Cloud Admin",
    "source-select",
    "Localhost 18080",
    "metric-hit",
    "metric-miss",
    "theme-button",
    "server-status",
]:
    assert marker in body, marker

agent_archive_record = {
    "ID": "phase72-agent-archive-row",
    "Type": "browser.domain.observed",
    "Source": "collector.browser.history",
    "Timestamp": "2026-06-13T08:00:00Z",
    "TenantID": "family-varadha",
    "DeviceID": "demo-study-laptop",
    "HostName": "demo-study-laptop",
    "AppName": "edge",
    "Metadata": {
        "browser_name": "edge",
        "domain": "youtube.com",
        "category": "video-streaming",
        "visit_count": "3",
        "youtube_study_match": "false",
        "stored_url_mode": "domain_only"
    }
}
row = module._browser_row(agent_archive_record, "tenant=family-varadha/device=demo-study-laptop/sample.jsonl.gz")
assert row["browser"] == "edge", row
assert row["domain"] == "youtube.com", row
assert row["study_safe"] is False, row
assert row["visit_count"] == 3, row

study_record = dict(agent_archive_record)
study_record["Metadata"] = dict(agent_archive_record["Metadata"])
study_record["Metadata"]["browser_name"] = "chrome"
study_record["Metadata"]["domain"] = "docs.python.org"
study_record["Metadata"]["category"] = "study"
study = module._browser_row(study_record, "tenant=family-varadha/device=demo-study-laptop/sample.jsonl.gz")
assert study["browser"] == "chrome", study
assert study["study_safe"] is True, study

summary = module.lambda_handler({"rawPath": "/api/s3-summary", "requestContext": {"http": {"method": "GET"}}}, None)
assert summary["statusCode"] == 200, summary
payload = json.loads(summary["body"])
assert payload["status"] == "not_configured", payload
assert "cache" in payload, payload
'@ | python -
    }

    Write-TraceDeckLog -Level "INFO" -Message "Lambda frontend contract passed."
    Complete-TraceDeckScriptLog
}
catch {
    Write-TraceDeckLog -Level "ERROR" -Message $_.Exception.Message
    throw
}
