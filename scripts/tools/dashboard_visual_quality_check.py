import argparse
import json
import re
import sys
from datetime import datetime, timezone
from urllib.parse import urljoin

from playwright.sync_api import sync_playwright


PAGES = (
    {
        "name": "dashboard",
        "path": "/",
        "expected_h1": "TraceDeck Console",
        "status_id": "backend-status",
    },
    {
        "name": "browser_activity",
        "path": "/browser-activity",
        "expected_h1": "Browser Viewer",
        "status_id": "page-status",
    },
)

THEMES = ("light", "dark")

VIEWPORTS = (
    {"name": "desktop", "width": 1440, "height": 960},
    {"name": "mobile", "width": 390, "height": 844},
)

FORBIDDEN_TEXT_PATTERNS = (
    r"Browser\s*\{",
    r"Center\s*\{",
    r"\bRev Ops\b",
    r"\bNotif Rev\b",
    r"\bNotify Pro\b",
    r"\bNotifs\b",
    r"\[[BCTR]\]",
    r"\{[BCTR]\}",
)


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck dashboard visual quality contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": args.base_url.rstrip("/") + "/",
        "privacy_boundary": "rendered layout metrics only; no screenshots, credentials, cookies, tokens, page dumps, raw URLs, or private content capture",
        "checks": [],
    }
    failures = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        try:
            for viewport in VIEWPORTS:
                for theme in THEMES:
                    for page_def in PAGES:
                        page = browser.new_page(viewport={"width": viewport["width"], "height": viewport["height"]})
                        try:
                            page.add_init_script(
                                f"""window.localStorage.setItem("tracedeck.ui.theme", {json.dumps(theme)});"""
                            )
                            url = urljoin(report["base_url"], page_def["path"].lstrip("/"))
                            page.goto(url, wait_until="networkidle", timeout=60000)
                            page.wait_for_selector(f"#{page_def['status_id']}", state="visible", timeout=30000)
                            result = page.evaluate(
                                """({ theme, expectedH1, statusID }) => {
                                  const parseColor = (value) => {
                                    const probe = document.createElement("span");
                                    probe.style.color = value;
                                    document.body.appendChild(probe);
                                    const rgb = window.getComputedStyle(probe).color.match(/\\d+(?:\\.\\d+)?/g).map(Number);
                                    probe.remove();
                                    return rgb.slice(0, 3);
                                  };
                                  const luminance = (rgb) => {
                                    const [r, g, b] = rgb.map((channel) => {
                                      const normalized = channel / 255;
                                      return normalized <= 0.03928
                                        ? normalized / 12.92
                                        : Math.pow((normalized + 0.055) / 1.055, 2.4);
                                    });
                                    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
                                  };
                                  const root = window.getComputedStyle(document.body);
                                  const h1 = document.querySelector("h1");
                                  const status = document.getElementById(statusID);
                                  const bodyText = document.body.innerText || "";
                                  const toolbarLabels = Array.from(document.querySelectorAll(".toolbar button"))
                                    .map((button) => button.textContent.trim());
                                  const commandLabels = Array.from(document.querySelectorAll(".command-jump .command-label"))
                                    .map((label) => label.textContent.trim());
                                  const commandMetaCount = document.querySelectorAll(".command-jump .command-meta").length;
                                  const tinyInteractive = Array.from(document.querySelectorAll(".toolbar button, .dashboard-page-tab, .command-jump, .badge, .pill"))
                                    .filter((element) => {
                                      const rect = element.getBoundingClientRect();
                                      const style = window.getComputedStyle(element);
                                      return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden" && rect.height < 26;
                                    })
                                    .map((element) => ({
                                      text: element.textContent.trim().slice(0, 80),
                                      height: Number(element.getBoundingClientRect().height.toFixed(2)),
                                    }));
                                  const commandStacked = Array.from(document.querySelectorAll(".command-jump")).every((button) => {
                                    const span = button.querySelector(".command-meta");
                                    if (!span) return true;
                                    const buttonRect = button.getBoundingClientRect();
                                    const spanRect = span.getBoundingClientRect();
                                    return spanRect.top > buttonRect.top + 16;
                                  });
                                  const expectedCommandLabels = [
                                    "Premium Operations",
                                    "Onboarding Center",
                                    "Customer Settings",
                                    "Revenue Operations",
                                    "Deployment Readiness",
                                    "Customer Control Room",
                                    "Customer Success Packet",
                                    "Push Activation",
                                    "Host Portfolio",
                                    "Account Portfolio",
                                    "Executive Console",
                                    "Notification Revenue",
                                    "Provider Simulation",
                                    "Provider Setup",
                                    "Package Billing",
                                    "Paid Operations",
                                    "Growth Dashboard",
                                    "Notification Command",
                                    "Delivery Assurance",
                                    "Notification Proof",
                                    "Weekly Reports",
                                    "Archive Proof",
                                    "Trust & Consent",
                                    "Host Details"
                                  ];
                                  const staleCommandLabels = new Set(["Premium", "Onboard", "Settings", "Revenue Ops", "Deploy", "Control", "Success", "Push", "Portfolio", "Account", "Executive", "Provider", "Setup", "Packages", "Paid Ops", "Revenue", "Notification Pro", "Assurance", "Notifications", "Reports", "Archive", "Trust", "Hosts"]);
                                  const backgroundColor = root.getPropertyValue("--bg").trim();
                                  const surfaceColor = root.getPropertyValue("--surface").trim();
                                  const bgLuma = luminance(parseColor(backgroundColor));
                                  const surfaceLuma = luminance(parseColor(surfaceColor));
                                  const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                                  const viewportWidth = window.innerWidth;
                                  const checks = [
                                    {
                                      name: "expected-heading",
                                      ok: h1 && h1.textContent.trim() === expectedH1,
                                      detail: h1 ? h1.textContent.trim() : "missing"
                                    },
                                    {
                                      name: "status-visible",
                                      ok: Boolean(status && status.getBoundingClientRect().height >= 26),
                                      detail: status ? status.textContent.trim() : "missing"
                                    },
                                    {
                                      name: "no-debug-toolbar-labels",
                                      ok: toolbarLabels.every((label) => !/[\\[\\]{}]/.test(label) && !/^[BCTR]$/.test(label)),
                                      detail: toolbarLabels.join(", ")
                                    },
                                    {
                                      name: "no-tiny-interactive-labels",
                                      ok: tinyInteractive.length === 0,
                                      detail: tinyInteractive.map((item) => `${item.text}:${item.height}`).join(", ") || "all visible chips and buttons are at least 26px tall"
                                    },
                                    {
                                      name: "command-metadata-not-sidecar",
                                      ok: document.querySelectorAll(".command-jump").length === 0 || commandStacked,
                                      detail: commandStacked ? "shortcut metadata stacks under labels" : "shortcut metadata still renders as side labels"
                                    },
                                    {
                                      name: "command-labels-are-product-grade",
                                      ok: document.querySelectorAll(".command-jump").length === 0 || expectedCommandLabels.every((label) => commandLabels.includes(label)),
                                      detail: commandLabels.join(", ")
                                    },
                                    {
                                      name: "no-terse-command-labels",
                                      ok: commandLabels.every((label) => !staleCommandLabels.has(label)),
                                      detail: commandLabels.join(", ")
                                    },
                                    {
                                      name: "command-metadata-classed",
                                      ok: document.querySelectorAll(".command-jump").length === 0 || commandMetaCount === document.querySelectorAll(".command-jump").length,
                                      detail: `${commandMetaCount}/${document.querySelectorAll(".command-jump").length} command buttons have metadata rows`
                                    },
                                    {
                                      name: "theme-luminance",
                                      ok: theme === "dark" ? bgLuma < 0.08 && surfaceLuma < 0.14 : bgLuma > 0.78 && surfaceLuma > 0.92,
                                      detail: `bg=${backgroundColor}(${bgLuma.toFixed(3)}) surface=${surfaceColor}(${surfaceLuma.toFixed(3)})`
                                    },
                                    {
                                      name: "no-horizontal-overflow",
                                      ok: documentWidth <= viewportWidth + 2,
                                      detail: `${documentWidth}px document width within ${viewportWidth}px viewport`
                                    }
                                  ];
                                  return {
                                    h1: h1 ? h1.textContent.trim() : "",
                                    status_text: status ? status.textContent.trim() : "",
                                    toolbar_labels: toolbarLabels,
                                    command_labels: commandLabels,
                                    body_text: bodyText,
                                    checks,
                                  };
                                }""",
                                {
                                    "theme": theme,
                                    "expectedH1": page_def["expected_h1"],
                                    "statusID": page_def["status_id"],
                                },
                            )
                            body_text = result.pop("body_text")
                            forbidden_hits = [
                                pattern
                                for pattern in FORBIDDEN_TEXT_PATTERNS
                                if re.search(pattern, body_text, flags=re.IGNORECASE)
                            ]
                            result["checks"].append(
                                {
                                    "name": "no-stale-debug-copy",
                                    "ok": len(forbidden_hits) == 0,
                                    "detail": ", ".join(forbidden_hits) or "no stale abbreviations or brace labels",
                                }
                            )
                            entry = {
                                "page": page_def["name"],
                                "theme": theme,
                                "viewport": viewport,
                                **result,
                            }
                            report["checks"].append(entry)
                            for check in result["checks"]:
                                if not check["ok"]:
                                    failures.append(
                                        f"{page_def['name']} {theme} {viewport['name']}: {check['name']} - {check['detail']}"
                                    )
                        finally:
                            page.close()
        finally:
            browser.close()

    report["ok"] = len(failures) == 0
    report["failures"] = failures
    with open(args.output, "w", encoding="utf-8") as handle:
        json.dump(report, handle, indent=2)

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    print(f"Dashboard visual quality contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
