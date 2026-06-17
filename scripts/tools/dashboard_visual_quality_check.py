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
        "ready_id": "dashboard-page-nav",
        "expected_nav": (
            "Command Center",
            "Host Portfolio",
            "Signal Queue",
            "Browser Intelligence",
            "Delivery Assurance",
            "Revenue Packaging",
            "Trust Center",
        ),
    },
    {
        "name": "browser_activity",
        "path": "/browser-activity",
        "expected_h1": "Browser Intelligence",
        "status_id": "page-status",
        "ready_id": "activity-table",
        "expected_nav": (),
    },
)

THEMES = ("light", "dark")

VIEWPORTS = (
    {"name": "desktop", "width": 1440, "height": 960},
    {"name": "mobile", "width": 390, "height": 844},
)

PAGE_READY_STATE = "domcontentloaded"
NAVIGATION_TIMEOUT_MS = 120000
READY_TIMEOUT_MS = 120000
ASYNC_LAYOUT_SETTLE_MS = 7000

FORBIDDEN_TEXT_PATTERNS = (
    r"Browser\s*\{",
    r"Center\s*\{",
    r"\bTD\b",
    r"\bRev Ops\b",
    r"\bNotif Rev\b",
    r"\bNotify Pro\b",
    r"\bNotifs\b",
    r"\[[BCTR]\]",
    r"\{[BCTR]\}",
    r"Workspace Navigator",
    r"Premium Operations",
    r"Phase 82 product polish",
)


def wait_for_contract_ready(page, page_def):
    page.wait_for_selector(f"#{page_def['status_id']}", state="visible", timeout=READY_TIMEOUT_MS)
    page.wait_for_selector(f"#{page_def['ready_id']}", timeout=READY_TIMEOUT_MS)
    page.wait_for_timeout(ASYNC_LAYOUT_SETTLE_MS)


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck modern UI visual quality contract")
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
                            page.goto(url, wait_until=PAGE_READY_STATE, timeout=NAVIGATION_TIMEOUT_MS)
                            wait_for_contract_ready(page, page_def)
                            result = page.evaluate(
                                """({ theme, expectedH1, statusID, expectedNav }) => {
                                  const parseColor = (value) => {
                                    const probe = document.createElement("span");
                                    probe.style.color = value;
                                    document.body.appendChild(probe);
                                    const match = window.getComputedStyle(probe).color.match(/\\d+(?:\\.\\d+)?/g);
                                    const rgb = match ? match.map(Number).slice(0, 3) : [0, 0, 0];
                                    probe.remove();
                                    return rgb;
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
                                  const visible = (element) => {
                                    if (!element) return false;
                                    const rect = element.getBoundingClientRect();
                                    const style = window.getComputedStyle(element);
                                    return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden";
                                  };
                                  const root = window.getComputedStyle(document.body);
                                  const h1 = document.querySelector("h1");
                                  const status = document.getElementById(statusID);
                                  const brandMark = document.querySelector(".brand-mark");
                                  const bodyText = document.body.innerText || "";
                                  const toolbarLabels = Array.from(document.querySelectorAll(".toolbar button, .toolbar a"))
                                    .map((item) => item.textContent.trim());
                                  const navLabels = Array.from(document.querySelectorAll("[data-page-target]"))
                                    .map((button) => button.textContent.trim());
                                  const tinyInteractive = Array.from(document.querySelectorAll(".toolbar button, .toolbar a, [data-page-target], .badge, .pill"))
                                    .filter((element) => {
                                      const rect = element.getBoundingClientRect();
                                      const style = window.getComputedStyle(element);
                                      return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden" && rect.height < 26;
                                    })
                                    .map((element) => ({
                                      text: element.textContent.trim().slice(0, 80),
                                      height: Number(element.getBoundingClientRect().height.toFixed(2)),
                                    }));
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
                                      ok: Boolean(status && status.getBoundingClientRect().height >= 28),
                                      detail: status ? status.textContent.trim() : "missing"
                                    },
                                    {
                                      name: "symbolic-brand-mark",
                                      ok: Boolean(brandMark && brandMark.textContent.trim() === "" && brandMark.querySelectorAll("span").length >= 3),
                                      detail: brandMark ? `text=${brandMark.textContent.trim() || "empty"} marks=${brandMark.querySelectorAll("span").length}` : "missing"
                                    },
                                    {
                                      name: "toolbar-labels-clean",
                                      ok: toolbarLabels.every((label) => !/[\\[\\]{}]/.test(label) && !/^[BCTR]$/.test(label)),
                                      detail: toolbarLabels.join(", ")
                                    },
                                    {
                                      name: "no-tiny-interactive-labels",
                                      ok: tinyInteractive.length === 0,
                                      detail: tinyInteractive.map((item) => `${item.text}:${item.height}`).join(", ") || "all visible chips and controls are at least 26px tall"
                                    },
                                    {
                                      name: "navigation-labels-product-grade",
                                      ok: expectedNav.length === 0 || expectedNav.every((label) => navLabels.includes(label)),
                                      detail: navLabels.join(", ") || "no page navigation on this page"
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
                                    nav_labels: navLabels,
                                    body_text: bodyText,
                                    checks,
                                  };
                                }""",
                                {
                                    "theme": theme,
                                    "expectedH1": page_def["expected_h1"],
                                    "statusID": page_def["status_id"],
                                    "expectedNav": list(page_def["expected_nav"]),
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
                                    "detail": ", ".join(forbidden_hits) or "no stale abbreviations or legacy phase labels",
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
