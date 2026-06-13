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
        )
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
    else:
        run(powershell("./scripts/local/smoke-phase69.ps1"))
        run(powershell("./scripts/local/newman-phase69.ps1"))
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


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck development controller")
    parser.add_argument("--addr", default=DEFAULT_ADDR, help="Local backend address")
    sub = parser.add_subparsers(dest="cmd")

    status = sub.add_parser("status", help="Check local server, output paths, and SAM config")
    status.set_defaults(func=cmd_status)

    server = sub.add_parser("server", help="Start, stop, restart, or check the local server")
    server.add_argument("action", choices=["start", "stop", "restart", "status"], nargs="?", default="status")
    server.set_defaults(func=cmd_server)

    sam = sub.add_parser("sam", help="Build, deploy, run, tail, or save outputs for the SAM frontend")
    sam.add_argument("action", choices=["build", "deploy", "restart", "local", "outputs", "logs", "tail"], nargs="?", default="build")
    sam.set_defaults(func=cmd_sam)

    test = sub.add_parser("test", help="Run Phase 69 smoke, Newman, or verify scripts")
    test.add_argument("target", choices=["phase69", "smoke", "newman", "verify"], nargs="?", default="phase69")
    test.set_defaults(func=cmd_test)

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
