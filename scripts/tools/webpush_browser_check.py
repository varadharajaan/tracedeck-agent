import argparse
import json
import sys
from datetime import datetime, timezone
from urllib.parse import urljoin, urlparse

from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from playwright.sync_api import sync_playwright


PAGE_READY_STATE = "domcontentloaded"
NAVIGATION_TIMEOUT_MS = 120000
READY_TIMEOUT_MS = 120000
PUSH_TIMEOUT_MS = 30000


def origin_for(base_url: str) -> str:
    parsed = urlparse(base_url)
    return f"{parsed.scheme}://{parsed.netloc}"


def capability_probe(page):
    return page.evaluate(
        """() => ({
          secure_context: window.isSecureContext,
          notification_api: "Notification" in window,
          notification_permission: "Notification" in window ? Notification.permission : "unsupported",
          service_worker: "serviceWorker" in navigator,
          push_manager: "PushManager" in window,
          origin: window.location.origin
        })"""
    )


def ui_probe(page):
    return page.evaluate(
        """() => {
          const button = document.getElementById("push-setup-button");
          const status = document.getElementById("push-setup-status");
          return {
            button_text: button ? button.textContent.trim() : "missing",
            button_disabled: button ? button.disabled : null,
            status_text: status ? status.textContent.trim() : "missing",
            status_class: status ? status.className : "missing"
          };
        }"""
    )


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck dashboard Web Push browser activation check")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--profile-dir", required=True)
    args = parser.parse_args()

    base_url = args.base_url.rstrip("/") + "/"
    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": base_url,
        "privacy_boundary": "browser capability/status metadata only; no screenshots, cookies, tokens, raw page content, push endpoints, credentials, or private content capture",
        "ok": False,
        "capabilities_before": {},
        "capabilities_after": {},
        "ui_before": {},
        "ui_after": {},
        "console": [],
        "failures": [],
    }
    failures = []

    with sync_playwright() as playwright:
        context = playwright.chromium.launch_persistent_context(
            args.profile_dir,
            headless=True,
            viewport={"width": 1366, "height": 900},
        )
        try:
            context.grant_permissions(["notifications"], origin=origin_for(base_url))
            page = context.pages[0] if context.pages else context.new_page()
            page.on(
                "console",
                lambda message: report["console"].append(
                    {
                        "type": message.type,
                        "text": message.text[:500],
                    }
                ),
            )
            try:
                page.goto(urljoin(base_url, "/"), wait_until=PAGE_READY_STATE, timeout=NAVIGATION_TIMEOUT_MS)
                page.wait_for_selector("#push-setup-button", state="visible", timeout=READY_TIMEOUT_MS)
                page.wait_for_selector("#push-setup-status", state="visible", timeout=READY_TIMEOUT_MS)
                report["capabilities_before"] = capability_probe(page)
                report["ui_before"] = ui_probe(page)
                page.click("#push-setup-button")
                try:
                    page.wait_for_function(
                        """() => {
                          const status = document.getElementById("push-setup-status");
                          if (!status) return false;
                          return status.classList.contains("ready") || status.classList.contains("failed");
                        }""",
                        timeout=PUSH_TIMEOUT_MS,
                    )
                except PlaywrightTimeoutError:
                    failures.append("push setup did not reach ready or failed status within timeout")
                report["capabilities_after"] = capability_probe(page)
                report["ui_after"] = ui_probe(page)
            finally:
                page.close()
        finally:
            context.close()

    after = report["ui_after"]
    if "ready" not in after.get("status_class", ""):
        failures.append(f"push setup not ready: {after.get('status_text', 'missing status')}")
    for name, value in report["capabilities_after"].items():
        if name in ("secure_context", "notification_api", "service_worker", "push_manager") and not value:
            failures.append(f"missing browser capability: {name}")
    if report["capabilities_after"].get("notification_permission") != "granted":
        failures.append(
            f"notification permission is {report['capabilities_after'].get('notification_permission', 'unknown')}"
        )

    report["failures"] = failures
    report["ok"] = len(failures) == 0
    with open(args.output, "w", encoding="utf-8") as handle:
        json.dump(report, handle, indent=2)

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    print(f"Web Push browser activation check passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
