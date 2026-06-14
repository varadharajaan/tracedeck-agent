import argparse
import json
import sys
from datetime import datetime, timezone
from urllib.parse import urljoin

from playwright.sync_api import sync_playwright


PAGES = (
    {"name": "dashboard", "path": "/", "status_id": "backend-status", "theme_id": "theme-toggle-label"},
    {"name": "browser_activity", "path": "/browser-activity", "status_id": "page-status", "theme_id": "theme-toggle-label"},
)

THEMES = ("light", "dark")

VIEWPORTS = (
    {"name": "desktop", "width": 1440, "height": 960},
    {"name": "mobile", "width": 390, "height": 844},
)


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck screenshot-free light/dark theme contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": args.base_url.rstrip("/") + "/",
        "privacy_boundary": "theme metrics only; no screenshots, credentials, page content dumps, cookies, tokens, or private content capture",
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
                                """({ theme, statusID, themeID }) => {
                                  const status = document.getElementById(statusID);
                                  const themeLabel = document.getElementById(themeID);
                                  const toolbarButtons = Array.from(document.querySelectorAll(".toolbar button"))
                                    .map((button) => button.textContent.trim());
                                  const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                                  const viewportWidth = window.innerWidth;
                                  const bodyDark = document.body.classList.contains("theme-dark");
                                  const expectedLabel = theme === "dark" ? "Theme: Dark" : "Theme: Light";
                                  const badToolbarLabels = toolbarButtons.filter((label) => /[\\[\\]{}]/.test(label));
                                  const visible = (element) => {
                                    if (!element) return false;
                                    const rect = element.getBoundingClientRect();
                                    const style = window.getComputedStyle(element);
                                    return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden";
                                  };
                                  return {
                                    body_dark: bodyDark,
                                    theme_label: themeLabel ? themeLabel.textContent.trim() : "",
                                    status_visible: visible(status),
                                    status_text: status ? status.textContent.trim() : "",
                                    document_width: documentWidth,
                                    viewport_width: viewportWidth,
                                    toolbar_labels: toolbarButtons,
                                    bad_toolbar_labels: badToolbarLabels,
                                    checks: [
                                      {
                                        name: "theme-class",
                                        ok: theme === "dark" ? bodyDark : !bodyDark,
                                        detail: `body theme-dark=${bodyDark}`
                                      },
                                      {
                                        name: "theme-label",
                                        ok: themeLabel && themeLabel.textContent.trim() === expectedLabel,
                                        detail: themeLabel ? themeLabel.textContent.trim() : "missing"
                                      },
                                      {
                                        name: "status-visible",
                                        ok: visible(status),
                                        detail: status ? status.textContent.trim() : "missing"
                                      },
                                      {
                                        name: "no-horizontal-overflow",
                                        ok: documentWidth <= viewportWidth + 2,
                                        detail: `${documentWidth}px document width within ${viewportWidth}px viewport`
                                      },
                                      {
                                        name: "toolbar-labels-clean",
                                        ok: badToolbarLabels.length === 0,
                                        detail: badToolbarLabels.join(", ") || "toolbar labels are clean"
                                      }
                                    ]
                                  };
                                }""",
                                {"theme": theme, "statusID": page_def["status_id"], "themeID": page_def["theme_id"]},
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
                                    failures.append(f"{page_def['name']} {theme} {viewport['name']}: {check['name']} - {check['detail']}")
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
    print(f"Dashboard theme contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
