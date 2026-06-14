#!/usr/bin/env python3
from __future__ import annotations

import argparse
import json
import os
import re
import shutil
import subprocess
import sys
import time
import urllib.error
import urllib.request
from datetime import datetime
from pathlib import Path


PROJECT_ROOT = Path(__file__).parent.resolve()
LOG_ROOT = PROJECT_ROOT / "logs" / "local" / "devctl"
OUTPUT_ROOT = PROJECT_ROOT / "data" / "local" / "output"
BACKEND_PID = PROJECT_ROOT / "data" / "local" / "backend" / "tracedeck-backend.pid"
BACKEND_STATE = PROJECT_ROOT / "data" / "local" / "backend" / "backend-state.json"
DEFAULT_ADDR = "127.0.0.1:18080"
SAM_DIR = PROJECT_ROOT / "sam-app"
SAM_TEMPLATE = SAM_DIR / "template.yaml"
SAM_CONFIG = SAM_DIR / "samconfig.toml"

LOG_ROOT.mkdir(parents=True, exist_ok=True)
OUTPUT_ROOT.mkdir(parents=True, exist_ok=True)
LOG_FILE = LOG_ROOT / f"devctl-{datetime.now().strftime('%Y%m%d-%H%M%S')}-{os.getpid()}.log"


def console_write(message: str, *, end: str = "\n") -> None:
    try:
        print(message, end=end)
    except UnicodeEncodeError:
        encoding = sys.stdout.encoding or "utf-8"
        safe = message.encode(encoding, errors="replace").decode(encoding, errors="replace")
        print(safe, end=end)


def log(level: str, message: str) -> None:
    line = f"{datetime.now().isoformat()} [{level}] {message}"
    with LOG_FILE.open("a", encoding="utf-8") as handle:
        handle.write(line + "\n")
    console_write(line)


def run(cmd: list[str], *, cwd: Path = PROJECT_ROOT, check: bool = True, stream: bool = False) -> subprocess.CompletedProcess[str]:
    log("INFO", "Running: " + " ".join(cmd))
    if stream:
        proc = subprocess.Popen(cmd, cwd=str(cwd))
        rc = proc.wait()
        if check and rc != 0:
            raise SystemExit(rc)
        return subprocess.CompletedProcess(cmd, rc, "", "")
    completed = subprocess.run(
        cmd,
        cwd=str(cwd),
        text=True,
        capture_output=True,
        encoding="utf-8",
        errors="replace",
    )
    if completed.stdout:
        for line in completed.stdout.splitlines():
            log("DEBUG", line)
    if completed.stderr:
        for line in completed.stderr.splitlines():
            log("DEBUG", line)
    if check and completed.returncode != 0:
        raise RuntimeError(f"command failed with exit code {completed.returncode}: {' '.join(cmd)}")
    return completed


def powershell(script: str, *args: str) -> list[str]:
    return ["powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", script, *args]


def health_url(addr: str) -> str:
    return f"http://{addr}/health"


def local_url(addr: str) -> str:
    return f"http://{addr}"


def check_health(addr: str) -> tuple[bool, str]:
    try:
        with urllib.request.urlopen(health_url(addr), timeout=3) as response:
            body = response.read().decode("utf-8", errors="replace")
        return True, body
    except (urllib.error.URLError, TimeoutError) as exc:
        return False, str(exc)


def read_url(url: str, *, timeout: int = 8) -> dict[str, object]:
    try:
        with urllib.request.urlopen(url, timeout=timeout) as response:
            body = response.read().decode("utf-8", errors="replace")
            return {"ok": True, "status_code": response.status, "body": body, "error": ""}
    except urllib.error.HTTPError as exc:
        body = exc.read().decode("utf-8", errors="replace") if exc.fp else ""
        return {"ok": False, "status_code": exc.code, "body": body, "error": str(exc)}
    except (urllib.error.URLError, TimeoutError) as exc:
        return {"ok": False, "status_code": 0, "body": "", "error": str(exc)}


def read_json(url: str, *, timeout: int = 8) -> dict[str, object]:
    result = read_url(url, timeout=timeout)
    if not result["ok"]:
        result["json"] = {}
        return result
    try:
        result["json"] = json.loads(str(result["body"] or "{}"))
    except json.JSONDecodeError as exc:
        result["ok"] = False
        result["error"] = f"invalid JSON: {exc}"
        result["json"] = {}
    return result


def required_markers(body: str, markers: list[str]) -> dict[str, object]:
    missing = [marker for marker in markers if marker not in body]
    return {"ok": not missing, "missing": missing, "expected": markers}


def git_value(args: list[str]) -> str:
    completed = subprocess.run(["git", *args], cwd=str(PROJECT_ROOT), text=True, capture_output=True, encoding="utf-8", errors="replace")
    return completed.stdout.strip() if completed.returncode == 0 else ""


def frontend_url_from_output() -> str:
    url_path = OUTPUT_ROOT / "frontend-url.txt"
    if url_path.exists():
        return url_path.read_text(encoding="utf-8").strip().rstrip("/")
    return ""


def runtime_doctor_report(args: argparse.Namespace) -> dict[str, object]:
    base_url = local_url(args.addr)
    frontend_url = (args.frontend_url or frontend_url_from_output()).rstrip("/")
    generated_at = datetime.now().isoformat()

    report: dict[str, object] = {
        "generated_at": generated_at,
        "repo_root": str(PROJECT_ROOT),
        "output_root": str(OUTPUT_ROOT),
        "git": {
            "branch": git_value(["branch", "--show-current"]),
            "head": git_value(["rev-parse", "--short", "HEAD"]),
            "clean": git_value(["status", "--short"]) == "",
        },
        "local": runtime_doctor_local(base_url=base_url, tenant_id=args.tenant_id),
        "cloud": {"skipped": True, "overall": "skipped", "reason": "cloud checks skipped"},
    }

    if not args.skip_cloud:
        report["cloud"] = runtime_doctor_cloud(frontend_url=frontend_url, refresh=not args.no_cloud_refresh)

    local_ok = dict(report["local"]).get("overall") == "ok"
    cloud = dict(report["cloud"])
    cloud_ok = cloud.get("overall") in ("ok", "skipped")
    report["overall"] = "ok" if local_ok and cloud_ok else "fail"
    return report


def runtime_doctor_local(*, base_url: str, tenant_id: str) -> dict[str, object]:
    health = read_json(f"{base_url}/health", timeout=4)
    health_json = dict(health.get("json") or {})
    health_ok = bool(health["ok"]) and health_json.get("status") == "ok"

    dashboard = read_url(f"{base_url}/", timeout=5)
    dashboard_markers = required_markers(
        str(dashboard.get("body") or ""),
        ["TraceDeck Console", "theme-toggle-button", "server-status-light", "dashboard-page-nav", "browser-activity-button", "Delivery Assurance Center"],
    )

    browser_page = read_url(f"{base_url}/browser-activity", timeout=5)
    browser_markers = required_markers(
        str(browser_page.get("body") or ""),
        ["TraceDeck Browser Activity", "theme-toggle-button", "server-status-light", "<th>Source</th>", "metadata-only guard"],
    )

    browser_api = read_json(f"{base_url}/api/v1/tenants/{tenant_id}/browser-activity?limit=25", timeout=6)
    browser_payload = dict(browser_api.get("json") or {})
    browser_items = list(browser_payload.get("items") or [])
    browser_summary = dict(browser_payload.get("summary") or {})
    provenance_ok = bool(browser_items) and all(
        bool(item.get("source_kind")) and bool(item.get("evidence_scope")) and bool(item.get("evidence_detail"))
        for item in browser_items
        if isinstance(item, dict)
    )

    devices = read_json(f"{base_url}/api/v1/devices", timeout=5)
    devices_payload = dict(devices.get("json") or {})
    device_items = list(devices_payload.get("items") or [])
    device_id = str(device_items[0].get("device_id", "")) if device_items and isinstance(device_items[0], dict) else ""
    delivery_ok = False
    delivery_count = 0
    delivery_source = ""
    if device_id:
        deliveries = read_json(f"{base_url}/api/v1/devices/{device_id}/alert-deliveries", timeout=5)
        delivery_payload = dict(deliveries.get("json") or {})
        delivery_items = list(delivery_payload.get("items") or [])
        delivery_count = int(delivery_payload.get("count") or len(delivery_items))
        if delivery_items and isinstance(delivery_items[0], dict):
            delivery_source = str(delivery_items[0].get("source_kind", ""))
        delivery_ok = bool(deliveries["ok"]) and delivery_count > 0 and bool(delivery_source)

    assurance = read_json(f"{base_url}/api/v1/tenants/{tenant_id}/delivery-assurance?limit=25", timeout=6)
    assurance_payload = dict(assurance.get("json") or {})
    assurance_summary = dict(assurance_payload.get("summary") or {})
    assurance_privacy = str(assurance_payload.get("privacy_boundary", ""))
    assurance_demo_only = int(assurance_summary.get("demo_only") or 0)
    assurance_retrying = int(assurance_summary.get("retrying") or 0)
    assurance_buyer_ready = bool(assurance_summary.get("buyer_ready"))
    assurance_ok = (
        bool(assurance["ok"])
        and int(assurance_summary.get("routes_total") or 0) > 0
        and "metadata-only" in assurance_privacy
        and (assurance_demo_only == 0 or not assurance_buyer_ready)
    )

    checks = {
        "health": health_ok,
        "dashboard": bool(dashboard["ok"]) and bool(dashboard_markers["ok"]),
        "browser_page": bool(browser_page["ok"]) and bool(browser_markers["ok"]),
        "browser_api": bool(browser_api["ok"]) and provenance_ok,
        "devices": bool(devices["ok"]) and bool(device_id),
        "deliveries": delivery_ok,
        "delivery_assurance": assurance_ok,
    }
    return {
        "base_url": base_url,
        "overall": "ok" if all(checks.values()) else "fail",
        "checks": checks,
        "health": {"ok": health_ok, "status": health_json.get("status", ""), "service": health_json.get("service", ""), "error": health.get("error", "")},
        "dashboard": {"ok": checks["dashboard"], "status_code": dashboard.get("status_code"), **dashboard_markers},
        "browser_page": {"ok": checks["browser_page"], "status_code": browser_page.get("status_code"), **browser_markers},
        "browser_api": {
            "ok": checks["browser_api"],
            "rows": int(browser_summary.get("total") or len(browser_items)),
            "source_kinds": sorted({str(item.get("source_kind", "")) for item in browser_items if isinstance(item, dict) and item.get("source_kind")}),
            "privacy_boundary": browser_payload.get("privacy_boundary", ""),
        },
        "devices": {"ok": checks["devices"], "count": int(devices_payload.get("count") or len(device_items)), "first_device_id": device_id},
        "deliveries": {"ok": delivery_ok, "count": delivery_count, "first_source_kind": delivery_source},
        "delivery_assurance": {
            "ok": assurance_ok,
            "status": assurance_summary.get("status", ""),
            "score": int(assurance_summary.get("assurance_score") or 0),
            "provider_confirmed": int(assurance_summary.get("provider_confirmed") or 0),
            "demo_only": assurance_demo_only,
            "retrying": assurance_retrying,
            "buyer_ready": assurance_buyer_ready,
            "privacy_boundary": assurance_privacy,
        },
    }


def runtime_doctor_cloud(*, frontend_url: str, refresh: bool) -> dict[str, object]:
    if not frontend_url:
        return {"skipped": False, "overall": "fail", "reason": "frontend URL is not configured"}

    health = read_json(f"{frontend_url}/api/health", timeout=8)
    health_json = dict(health.get("json") or {})
    health_ok = bool(health["ok"]) and health_json.get("status") == "ok"

    summary_path = "/api/s3-summary?refresh=true" if refresh else "/api/s3-summary"
    summary = read_json(f"{frontend_url}{summary_path}", timeout=20)
    summary_payload = dict(summary.get("json") or {})
    summary_body = dict(summary_payload.get("summary") or {})
    refresh_ok = bool(summary["ok"]) and summary_payload.get("status") == "ok"

    cached = read_json(f"{frontend_url}/api/s3-summary", timeout=12)
    cached_payload = dict(cached.get("json") or {})
    cache = dict(cached_payload.get("cache") or {})
    cache_ok = bool(cached["ok"]) and cache.get("hit") is True

    return {
        "skipped": False,
        "frontend_url": frontend_url,
        "overall": "ok" if health_ok and refresh_ok and cache_ok else "fail",
        "health": {"ok": health_ok, "status": health_json.get("status", ""), "service": health_json.get("service", ""), "error": health.get("error", "")},
        "s3_summary": {
            "ok": refresh_ok,
            "status": summary_payload.get("status", ""),
            "objects": int(summary_body.get("objects") or 0),
            "sampled_rows": int(summary_body.get("sampled_rows") or 0),
            "non_study_youtube": int(summary_body.get("non_study_youtube") or 0),
            "privacy_boundary": summary_payload.get("privacy_boundary", ""),
        },
        "cache": {
            "ok": cache_ok,
            "hit": bool(cache.get("hit")),
            "hit_percent": cache.get("hit_percent", 0),
            "miss_percent": cache.get("miss_percent", 0),
            "hits": cache.get("hits", 0),
            "misses": cache.get("misses", 0),
            "ttl_seconds": cache.get("ttl_seconds", 0),
        },
    }


def save_doctor_report(report: dict[str, object]) -> None:
    json_path = OUTPUT_ROOT / "runtime-doctor.json"
    text_path = OUTPUT_ROOT / "runtime-doctor.txt"
    json_path.write_text(json.dumps(report, indent=2, sort_keys=True), encoding="utf-8")

    local = dict(report.get("local") or {})
    cloud = dict(report.get("cloud") or {})
    browser_api = dict(local.get("browser_api") or {})
    delivery_assurance = dict(local.get("delivery_assurance") or {})
    cloud_cache = dict(cloud.get("cache") or {})
    cloud_summary = dict(cloud.get("s3_summary") or {})
    lines = [
        "TRACEDECK RUNTIME DOCTOR",
        "=" * 80,
        f"Generated: {report.get('generated_at', '')}",
        f"Overall:   {report.get('overall', '')}",
        f"Branch:    {dict(report.get('git') or {}).get('branch', '')}",
        f"Commit:    {dict(report.get('git') or {}).get('head', '')}",
        "",
        f"Local:     {local.get('overall', '')} {local.get('base_url', '')}",
        f"Browser:   rows={browser_api.get('rows', 0)} sources={', '.join(browser_api.get('source_kinds', []))}",
        f"Device:    {dict(local.get('devices') or {}).get('first_device_id', '')}",
        f"Delivery:  count={dict(local.get('deliveries') or {}).get('count', 0)} source={dict(local.get('deliveries') or {}).get('first_source_kind', '')}",
        f"Assurance: score={delivery_assurance.get('score', 0)} provider={delivery_assurance.get('provider_confirmed', 0)} demo={delivery_assurance.get('demo_only', 0)} retrying={delivery_assurance.get('retrying', 0)} buyer_ready={delivery_assurance.get('buyer_ready', False)}",
        "",
    ]
    if cloud.get("skipped"):
        lines.append("Cloud:     skipped")
    else:
        lines.extend(
            [
                f"Cloud:     {cloud.get('overall', '')} {cloud.get('frontend_url', '')}",
                f"S3:        objects={cloud_summary.get('objects', 0)} rows={cloud_summary.get('sampled_rows', 0)} non_study_youtube={cloud_summary.get('non_study_youtube', 0)}",
                f"Cache:     hit={cloud_cache.get('hit', False)} hit%={cloud_cache.get('hit_percent', 0)} miss%={cloud_cache.get('miss_percent', 0)} ttl={cloud_cache.get('ttl_seconds', 0)}s",
            ]
        )
    lines.extend(["", f"JSON:      {json_path}", f"Text:      {text_path}", ""])
    text_path.write_text("\n".join(lines), encoding="utf-8")
    log("INFO", f"Saved runtime doctor JSON: {json_path}")
    log("INFO", f"Saved runtime doctor text: {text_path}")


def server_start(args: argparse.Namespace) -> int:
    run(
        powershell(
            "./scripts/local/start-dashboard-demo.ps1",
            "-Addr",
            args.addr,
            "-PidPath",
            str(BACKEND_PID.relative_to(PROJECT_ROOT)),
            "-DataPath",
            str(BACKEND_STATE.relative_to(PROJECT_ROOT)),
        ),
        stream=True,
    )
    ok, body = check_health(args.addr)
    if not ok:
        raise RuntimeError(f"server did not become healthy: {body}")
    log("INFO", f"Server ready: {local_url(args.addr)}")
    return 0


def server_stop(args: argparse.Namespace) -> int:
    run(powershell("./scripts/local/stop-backend-dev.ps1", "-PidPath", str(BACKEND_PID.relative_to(PROJECT_ROOT)), "-Addr", args.addr))
    return 0


def server_status(args: argparse.Namespace) -> int:
    ok, body = check_health(args.addr)
    if ok:
        log("INFO", f"Server connected: {local_url(args.addr)}")
        log("INFO", body)
        return 0
    log("WARN", f"Server not connected at {local_url(args.addr)}: {body}")
    return 1


def cmd_server(args: argparse.Namespace) -> int:
    if args.action == "start":
        return server_start(args)
    if args.action == "stop":
        return server_stop(args)
    if args.action == "restart":
        server_stop(args)
        time.sleep(1)
        return server_start(args)
    return server_status(args)


def sam_exe() -> str:
    candidates = [
        shutil.which("sam.cmd"),
        shutil.which("sam.exe"),
        shutil.which("sam"),
        Path(os.environ.get("APPDATA", "")) / "Python" / f"Python{sys.version_info.major}{sys.version_info.minor}" / "Scripts" / "sam.exe",
        Path(os.environ.get("APPDATA", "")) / "Python" / f"Python{sys.version_info.major}{sys.version_info.minor}" / "Scripts" / "sam.cmd",
        Path(r"C:\Program Files\Amazon\AWSSAMCLI\bin\sam.cmd"),
    ]
    for candidate in candidates:
        if not candidate:
            continue
        path = Path(candidate)
        if path.exists():
            return str(path)
    return "sam"


def aws_exe() -> str:
    return shutil.which("aws.cmd") or shutil.which("aws") or "aws"


def sam_stack_config() -> tuple[str, str]:
    text = SAM_CONFIG.read_text(encoding="utf-8") if SAM_CONFIG.exists() else ""
    stack = (re.search(r'stack_name\s*=\s*"([^"]+)"', text) or [None, "tracedeck-admin-frontend"])[1]
    region = (re.search(r'region\s*=\s*"([^"]+)"', text) or [None, "ap-south-1"])[1]
    return stack, region


def sam_build() -> Path:
    build_root = PROJECT_ROOT / "data" / "local" / "sam-build" / datetime.now().strftime("%Y%m%d-%H%M%S")
    build_root.mkdir(parents=True, exist_ok=True)
    run([sam_exe(), "build", "--template-file", str(SAM_TEMPLATE), "--build-dir", str(build_root)], cwd=SAM_DIR)
    return build_root / "template.yaml"


def save_stack_outputs(stack_name: str, region: str) -> None:
    completed = run(
        [
            aws_exe(),
            "cloudformation",
            "describe-stacks",
            "--stack-name",
            stack_name,
            "--region",
            region,
            "--query",
            "Stacks[0].Outputs",
            "--output",
            "json",
        ],
        check=True,
    )
    outputs = json.loads(completed.stdout or "[]")
    url = ""
    lines = [
        "=" * 80,
        "TRACEDECK ADMIN FRONTEND STACK OUTPUTS",
        "=" * 80,
        f"Saved:  {datetime.now().isoformat()}",
        f"Stack:  {stack_name}",
        f"Region: {region}",
        "",
    ]
    for item in sorted(outputs, key=lambda row: row.get("OutputKey", "")):
        key = item.get("OutputKey", "")
        value = item.get("OutputValue", "")
        desc = item.get("Description", "")
        lines.extend([key, f"  {desc}", f"  {value}", ""])
        if key == "FrontendFunctionUrl":
            url = value.rstrip("/")

    output_path = OUTPUT_ROOT / "stack-outputs.txt"
    output_path.write_text("\n".join(lines), encoding="utf-8")
    log("INFO", f"Saved stack outputs: {output_path}")
    if url:
        url_path = OUTPUT_ROOT / "frontend-url.txt"
        url_path.write_text(url + "\n", encoding="utf-8")
        log("INFO", f"Saved frontend URL: {url_path}")


def cmd_sam(args: argparse.Namespace) -> int:
    action = args.action
    stack, region = sam_stack_config()
    if action == "build":
        sam_build()
        return 0
    if action == "deploy":
        built_template = sam_build()
        run([sam_exe(), "deploy", "--template-file", str(built_template), "--config-file", str(SAM_CONFIG), "--no-confirm-changeset", "--no-fail-on-empty-changeset"], cwd=SAM_DIR, stream=True)
        save_stack_outputs(stack, region)
        return 0
    if action == "restart":
        built_template = sam_build()
        run([sam_exe(), "deploy", "--template-file", str(built_template), "--config-file", str(SAM_CONFIG), "--no-confirm-changeset", "--no-fail-on-empty-changeset"], cwd=SAM_DIR, stream=True)
        save_stack_outputs(stack, region)
        return 0
    if action == "local":
        run([sam_exe(), "local", "start-lambda", "--template-file", str(SAM_TEMPLATE)], cwd=SAM_DIR, stream=True)
        return 0
    if action == "outputs":
        save_stack_outputs(stack, region)
        return 0
    if action in ("logs", "tail"):
        cmd = [sam_exe(), "logs", "--stack-name", stack, "--region", region]
        if action == "tail":
            cmd.append("--tail")
        run(cmd, cwd=SAM_DIR, stream=True)
        return 0
    raise RuntimeError(f"unknown SAM action: {action}")


def cmd_test(args: argparse.Namespace) -> int:
    target = args.target
    if target == "smoke":
        run(powershell("./scripts/local/smoke-phase69.ps1"))
    elif target == "newman":
        run(powershell("./scripts/local/newman-phase69.ps1"))
    elif target == "verify":
        run(powershell("./scripts/verify/verify-phase69.ps1"))
    elif target == "phase72":
        run(powershell("./scripts/local/smoke-phase72.ps1"))
        run(powershell("./scripts/local/newman-phase72.ps1"))
    elif target == "smoke72":
        run(powershell("./scripts/local/smoke-phase72.ps1"))
    elif target == "newman72":
        run(powershell("./scripts/local/newman-phase72.ps1"))
    elif target == "verify72":
        run(powershell("./scripts/verify/verify-phase72.ps1"))
    elif target == "live":
        run(powershell("./scripts/local/test-live-server-provenance.ps1", "-BaseUrl", local_url(args.addr)))
    elif target == "phase73":
        run(powershell("./scripts/verify/verify-phase73.ps1"))
    elif target == "smoke73":
        run(powershell("./scripts/local/smoke-phase73.ps1"))
    elif target == "newman73":
        run(powershell("./scripts/local/newman-phase73.ps1"))
    elif target == "verify73":
        run(powershell("./scripts/verify/verify-phase73.ps1"))
    elif target == "phase74":
        run(powershell("./scripts/verify/verify-phase74.ps1"))
    elif target == "smoke74":
        run(powershell("./scripts/local/smoke-phase74.ps1"))
    elif target == "newman74":
        run(powershell("./scripts/local/newman-phase74.ps1"))
    elif target == "verify74":
        run(powershell("./scripts/verify/verify-phase74.ps1"))
    elif target == "phase75":
        run(powershell("./scripts/verify/verify-phase75.ps1"))
    elif target == "smoke75":
        run(powershell("./scripts/local/smoke-phase75.ps1"))
    elif target == "newman75":
        run(powershell("./scripts/local/newman-phase75.ps1"))
    elif target == "verify75":
        run(powershell("./scripts/verify/verify-phase75.ps1"))
    elif target == "phase76":
        run(powershell("./scripts/verify/verify-phase76.ps1"))
    elif target == "smoke76":
        run(powershell("./scripts/local/smoke-phase76.ps1"))
    elif target == "newman76":
        run(powershell("./scripts/local/newman-phase76.ps1"))
    elif target == "verify76":
        run(powershell("./scripts/verify/verify-phase76.ps1"))
    elif target == "phase78":
        run(powershell("./scripts/verify/verify-phase78.ps1"))
    elif target == "smoke78":
        run(powershell("./scripts/local/smoke-phase78.ps1"))
    elif target == "newman78":
        run(powershell("./scripts/local/newman-phase78.ps1"))
    elif target == "verify78":
        run(powershell("./scripts/verify/verify-phase78.ps1"))
    elif target == "phase80":
        run(powershell("./scripts/verify/verify-phase80.ps1"))
    elif target == "smoke80":
        run(powershell("./scripts/local/smoke-phase80.ps1"))
    elif target == "newman80":
        run(powershell("./scripts/local/newman-phase80.ps1"))
    elif target == "verify80":
        run(powershell("./scripts/verify/verify-phase80.ps1"))
    elif target == "phase81":
        run(powershell("./scripts/verify/verify-phase81.ps1"))
    elif target == "smoke81":
        run(powershell("./scripts/local/smoke-phase81.ps1"))
    elif target == "newman81":
        run(powershell("./scripts/local/newman-phase81.ps1"))
    elif target == "verify81":
        run(powershell("./scripts/verify/verify-phase81.ps1"))
    elif target == "phase82":
        run(powershell("./scripts/verify/verify-phase82.ps1"))
    elif target == "smoke82":
        run(powershell("./scripts/local/smoke-phase82.ps1"))
    elif target == "newman82":
        run(powershell("./scripts/local/newman-phase82.ps1"))
    elif target == "verify82":
        run(powershell("./scripts/verify/verify-phase82.ps1"))
    elif target == "phase83":
        run(powershell("./scripts/verify/verify-phase83.ps1"))
    elif target == "smoke83":
        run(powershell("./scripts/local/smoke-phase83.ps1"))
    elif target == "newman83":
        run(powershell("./scripts/local/newman-phase83.ps1"))
    elif target == "verify83":
        run(powershell("./scripts/verify/verify-phase83.ps1"))
    elif target == "phase84":
        run(powershell("./scripts/verify/verify-phase84.ps1"))
    elif target == "smoke84":
        run(powershell("./scripts/local/smoke-phase84.ps1"))
    elif target == "newman84":
        run(powershell("./scripts/local/newman-phase84.ps1"))
    elif target == "verify84":
        run(powershell("./scripts/verify/verify-phase84.ps1"))
    elif target == "phase85":
        run(powershell("./scripts/verify/verify-phase85.ps1"))
    elif target == "newman85":
        run(powershell("./scripts/local/newman-phase85.ps1"))
    elif target == "verify85":
        run(powershell("./scripts/verify/verify-phase85.ps1"))
    elif target == "phase86":
        run(powershell("./scripts/verify/verify-phase86.ps1"))
    elif target == "smoke86":
        run(powershell("./scripts/local/smoke-phase86.ps1"))
    elif target == "newman86":
        run(powershell("./scripts/local/newman-phase86.ps1"))
    elif target == "verify86":
        run(powershell("./scripts/verify/verify-phase86.ps1"))
    elif target == "phase87":
        run(powershell("./scripts/verify/verify-phase87.ps1"))
    elif target == "verify87":
        run(powershell("./scripts/verify/verify-phase87.ps1"))
    elif target == "quality":
        run(powershell("./scripts/local/test-go-quality-gates.ps1"))
    elif target == "theme":
        run(powershell("./scripts/local/test-dashboard-theme.ps1", "-BaseUrl", local_url(args.addr)))
    elif target == "visual":
        run(powershell("./scripts/local/test-dashboard-visual-quality.ps1", "-BaseUrl", local_url(args.addr)))
    else:
        run(powershell("./scripts/local/smoke-phase69.ps1"))
        run(powershell("./scripts/local/newman-phase69.ps1"))
    return 0


def cloud_args(script: str, args: argparse.Namespace) -> list[str]:
    command = powershell(script, "-Region", args.region)
    if args.bucket:
        command.extend(["-Bucket", args.bucket])
    if args.frontend_url and script.endswith(("smoke-phase72.ps1", "newman-phase72.ps1")):
        command.extend(["-FrontendUrl", args.frontend_url])
    return command


def cmd_cloud(args: argparse.Namespace) -> int:
    if args.action == "seed":
        run(cloud_args("./scripts/local/upload-cloud-sample-phase72.ps1", args))
    elif args.action == "visual":
        run(powershell("./scripts/local/test-lambda-frontend-visual.ps1"))
    elif args.action == "smoke":
        run(cloud_args("./scripts/local/smoke-phase72.ps1", args))
    elif args.action == "newman":
        run(cloud_args("./scripts/local/newman-phase72.ps1", args))
    else:
        run(cloud_args("./scripts/local/smoke-phase72.ps1", args))
        run(cloud_args("./scripts/local/newman-phase72.ps1", args))
    return 0


def cmd_logs(args: argparse.Namespace) -> int:
    roots = {
        "devctl": PROJECT_ROOT / "logs" / "local" / "devctl",
        "backend": PROJECT_ROOT / "logs" / "local" / "backend",
        "smoke": PROJECT_ROOT / "logs" / "local" / "smoke",
        "newman": PROJECT_ROOT / "logs" / "local" / "newman",
        "verify": PROJECT_ROOT / "logs" / "local" / "verify",
    }
    root = roots[args.kind]
    files = sorted(root.glob("*.log"), key=lambda path: path.stat().st_mtime, reverse=True) if root.exists() else []
    if not files:
        log("WARN", f"No logs found under {root}")
        return 1
    path = files[0]
    log("INFO", f"Showing {path}")
    if args.tail:
        with path.open("r", encoding="utf-8", errors="replace") as handle:
            handle.seek(0, os.SEEK_END)
            try:
                while True:
                    line = handle.readline()
                    if line:
                        console_write(line, end="")
                    else:
                        time.sleep(0.25)
            except KeyboardInterrupt:
                return 0
    lines = path.read_text(encoding="utf-8", errors="replace").splitlines()
    for line in lines[-args.lines :]:
        console_write(line)
    return 0


def cmd_status(args: argparse.Namespace) -> int:
    ok, body = check_health(args.addr)
    log("INFO", f"Repo: {PROJECT_ROOT}")
    log("INFO", f"Devctl log: {LOG_FILE}")
    log("INFO", f"Output folder: {OUTPUT_ROOT}")
    if ok:
        log("INFO", f"Local server connected: {local_url(args.addr)}")
        log("INFO", body)
    else:
        log("WARN", f"Local server not connected: {local_url(args.addr)}")
    stack, region = sam_stack_config()
    log("INFO", f"SAM stack: {stack} region={region}")
    url_path = OUTPUT_ROOT / "frontend-url.txt"
    if url_path.exists():
        log("INFO", f"Frontend URL: {url_path.read_text(encoding='utf-8').strip()}")
    return 0 if ok else 1


def cmd_doctor(args: argparse.Namespace) -> int:
    report = runtime_doctor_report(args)
    save_doctor_report(report)
    local = dict(report.get("local") or {})
    cloud = dict(report.get("cloud") or {})
    log("INFO", f"Runtime doctor local={local.get('overall', '')} cloud={cloud.get('overall', '')} overall={report.get('overall', '')}")
    return 0 if report.get("overall") == "ok" else 1


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck development controller")
    parser.add_argument("--addr", default=DEFAULT_ADDR, help="Local backend address")
    sub = parser.add_subparsers(dest="cmd")

    status = sub.add_parser("status", help="Check local server, output paths, and SAM config")
    status.set_defaults(func=cmd_status)

    doctor = sub.add_parser("doctor", help="Write local/cloud runtime assurance reports under data/local/output")
    doctor.add_argument("--tenant-id", default="family-varadha", help="Tenant used for browser activity readback")
    doctor.add_argument("--frontend-url", default="", help="Override Lambda Function URL")
    doctor.add_argument("--skip-cloud", action="store_true", help="Check only the local backend and browser viewer")
    doctor.add_argument("--no-cloud-refresh", action="store_true", help="Read Lambda S3 summary without a forced refresh")
    doctor.set_defaults(func=cmd_doctor)

    server = sub.add_parser("server", help="Start, stop, restart, or check the local server")
    server.add_argument("action", choices=["start", "stop", "restart", "status"], nargs="?", default="status")
    server.set_defaults(func=cmd_server)

    sam = sub.add_parser("sam", help="Build, deploy, run, tail, or save outputs for the SAM frontend")
    sam.add_argument("action", choices=["build", "deploy", "restart", "local", "outputs", "logs", "tail"], nargs="?", default="build")
    sam.set_defaults(func=cmd_sam)

    test = sub.add_parser("test", help="Run TraceDeck smoke, Newman, live, or phase verification scripts")
    test.add_argument(
        "target",
        choices=[
            "phase69",
            "phase72",
            "phase73",
            "phase74",
            "phase75",
            "phase76",
            "phase78",
            "phase80",
            "phase81",
            "phase82",
            "phase83",
            "phase84",
            "phase85",
            "phase86",
            "phase87",
            "smoke",
            "newman",
            "verify",
            "smoke72",
            "newman72",
            "verify72",
            "smoke73",
            "newman73",
            "verify73",
            "smoke74",
            "newman74",
            "verify74",
            "smoke75",
            "newman75",
            "verify75",
            "smoke76",
            "newman76",
            "verify76",
            "smoke78",
            "newman78",
            "verify78",
            "smoke80",
            "newman80",
            "verify80",
            "smoke81",
            "newman81",
            "verify81",
            "smoke82",
            "newman82",
            "verify82",
            "smoke83",
            "newman83",
            "verify83",
            "smoke84",
            "newman84",
            "verify84",
            "newman85",
            "verify85",
            "smoke86",
            "newman86",
            "verify86",
            "verify87",
            "quality",
            "theme",
            "visual",
            "live",
        ],
        nargs="?",
        default="phase69",
    )
    test.set_defaults(func=cmd_test)

    cloud = sub.add_parser("cloud", help="Seed or verify the S3-backed Lambda frontend")
    cloud.add_argument("action", choices=["seed", "visual", "smoke", "newman", "phase72"], nargs="?", default="phase72")
    cloud.add_argument("--bucket", default="", help="Override S3 data bucket")
    cloud.add_argument("--region", default="ap-south-1", help="AWS region")
    cloud.add_argument("--frontend-url", default="", help="Override Lambda Function URL")
    cloud.set_defaults(func=cmd_cloud)

    logs = sub.add_parser("logs", help="Show latest local logs")
    logs.add_argument("--kind", choices=["devctl", "backend", "smoke", "newman", "verify"], default="devctl")
    logs.add_argument("--tail", action="store_true")
    logs.add_argument("--lines", type=int, default=60)
    logs.set_defaults(func=cmd_logs)

    args = parser.parse_args()
    if not args.cmd:
        parser.print_help()
        return 0
    try:
        return args.func(args) or 0
    except Exception as exc:
        log("ERROR", str(exc))
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
