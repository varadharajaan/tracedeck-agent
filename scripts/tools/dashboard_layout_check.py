import argparse
import json
import sys
from datetime import datetime, timezone
from urllib.parse import urljoin

from playwright.sync_api import sync_playwright


VIEWPORTS = (
    {"name": "desktop", "width": 1440, "height": 1000},
    {"name": "tablet", "width": 900, "height": 900},
    {"name": "mobile", "width": 390, "height": 844},
)

PAGE_READY_STATE = "domcontentloaded"
NAVIGATION_TIMEOUT_MS = 120000
READY_TIMEOUT_MS = 120000
ASYNC_LAYOUT_SETTLE_MS = 7000

EXPECTED_PAGES = (
    "overview",
    "hosts",
    "activity",
    "browser",
    "delivery",
    "revenue",
    "trust",
)

REQUIRED_IDS = (
    "dashboard-page-nav",
    "tenant-input",
    "device-select",
    "include-demo-toggle",
    "theme-toggle-button",
    "theme-toggle-label",
    "refresh-button",
    "push-setup-status",
    "legacy-dashboard-button",
    "backend-status",
    "server-status-light",
    "mode-badge",
    "pipeline-status",
    "operator-brief-list",
    "browser-activity-button",
    "browser-domain-table",
    "delivery-list",
    "package-list",
    "privacy-boundary",
)


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck modern dashboard layout contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    base_url = args.base_url.rstrip("/") + "/"
    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": base_url,
        "privacy_boundary": "layout metrics only; no screenshots, video, credentials, cookies, tokens, URLs, or private content capture",
        "viewports": [],
    }
    failures = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        try:
            for viewport in VIEWPORTS:
                page = browser.new_page(viewport={"width": viewport["width"], "height": viewport["height"]})
                try:
                    page.goto(urljoin(base_url, "/"), wait_until=PAGE_READY_STATE, timeout=NAVIGATION_TIMEOUT_MS)
                    page.wait_for_selector("#dashboard-page-nav", state="visible", timeout=READY_TIMEOUT_MS)
                    page.wait_for_selector("#backend-status", state="visible", timeout=READY_TIMEOUT_MS)
                    page.wait_for_timeout(ASYNC_LAYOUT_SETTLE_MS)
                    result = page.evaluate(
                        """({ expectedPages, requiredIDs }) => {
                          const rounded = (rect) => ({
                            x: Math.round(rect.x),
                            y: Math.round(rect.y),
                            width: Math.round(rect.width),
                            height: Math.round(rect.height),
                            right: Math.round(rect.right),
                            bottom: Math.round(rect.bottom)
                          });
                          const visible = (element) => {
                            if (!element) return false;
                            const rect = element.getBoundingClientRect();
                            const style = window.getComputedStyle(element);
                            return rect.width > 0 && rect.height > 0 && style.visibility !== "hidden" && style.display !== "none";
                          };
                          const checks = [];
                          const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                          const viewportWidth = window.innerWidth;
                          checks.push({
                            name: "document-horizontal-overflow",
                            ok: documentWidth <= viewportWidth + 2,
                            detail: `${documentWidth}px document width within ${viewportWidth}px viewport`
                          });
                          for (const id of requiredIDs) {
                            const element = document.getElementById(id);
                            const rect = element ? rounded(element.getBoundingClientRect()) : null;
                            checks.push({
                              name: `required-present-${id}`,
                              ok: Boolean(element),
                              detail: rect ? `${rect.width}x${rect.height} at ${rect.x},${rect.y}` : "missing",
                              rect
                            });
                          }
                          const nav = document.getElementById("dashboard-page-nav");
                          const sidebarTools = document.getElementById("sidebar-tools");
                          const legacyButton = document.getElementById("legacy-dashboard-button");
                          const navButtons = Array.from(document.querySelectorAll("[data-page-target]"));
                          const navLabels = navButtons.map((button) => button.textContent.trim());
                          checks.push({
                            name: "page-navigation-visible",
                            ok: visible(nav),
                            detail: nav ? `${rounded(nav.getBoundingClientRect()).width}px wide` : "missing"
                          });
                          checks.push({
                            name: "legacy-v1-switch-visible",
                            ok: Boolean(legacyButton && visible(legacyButton) && legacyButton.getAttribute("href") === "/v1-old"),
                            detail: legacyButton ? `${legacyButton.textContent.trim()} -> ${legacyButton.getAttribute("href")}` : "missing"
                          });
                          if (nav && sidebarTools && visible(nav) && visible(sidebarTools)) {
                            const navRect = nav.getBoundingClientRect();
                            const toolRect = sidebarTools.getBoundingClientRect();
                            const overlaps = navRect.right > toolRect.left && navRect.left < toolRect.right && navRect.bottom > toolRect.top && navRect.top < toolRect.bottom;
                            checks.push({
                              name: "nav-does-not-overlap-workspace-tools",
                              ok: !overlaps,
                              detail: `nav ${Math.round(navRect.left)}-${Math.round(navRect.right)} tools ${Math.round(toolRect.left)}-${Math.round(toolRect.right)}`
                            });
                          }
                          checks.push({
                            name: "expected-page-targets",
                            ok: expectedPages.every((name) => navButtons.some((button) => button.dataset.pageTarget === name)),
                            detail: navButtons.map((button) => button.dataset.pageTarget).join(", ")
                          });
                          for (const button of navButtons) {
                            const target = button.dataset.pageTarget;
                            const pageElement = document.getElementById(`${target}-page`);
                            checks.push({
                              name: `page-target-exists-${target}`,
                              ok: Boolean(pageElement),
                              detail: `${target}-page`
                            });
                            checks.push({
                              name: `page-tab-text-fit-${target}`,
                              ok: button.scrollWidth <= button.clientWidth + 2 && button.scrollHeight <= button.clientHeight + 2,
                              detail: `${button.scrollWidth}x${button.scrollHeight} scroll within ${button.clientWidth}x${button.clientHeight} client`
                            });
                          }
                          return {
                            viewport: { width: viewportWidth, height: window.innerHeight },
                            document_width: documentWidth,
                            nav_labels: navLabels,
                            checks
                          };
                        }""",
                        {"expectedPages": list(EXPECTED_PAGES), "requiredIDs": list(REQUIRED_IDS)},
                    )
                    page_results = []
                    for page_name in EXPECTED_PAGES:
                        page.click(f'[data-page-target="{page_name}"]')
                        page.wait_for_timeout(350)
                        page_result = page.evaluate(
                            """(pageName) => {
                              const activePage = document.getElementById(`${pageName}-page`);
                              const activeTab = document.querySelector(`[data-page-target="${pageName}"]`);
                              const rect = activePage ? activePage.getBoundingClientRect() : null;
                              const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                              const viewportWidth = window.innerWidth;
                              return {
                                page: pageName,
                                active_page_visible: Boolean(activePage && rect.width > 0 && rect.height > 0 && getComputedStyle(activePage).display !== "none"),
                                active_tab_selected: Boolean(activeTab && activeTab.classList.contains("is-active")),
                                document_width: documentWidth,
                                viewport_width: viewportWidth,
                                no_horizontal_overflow: documentWidth <= viewportWidth + 2
                              };
                            }""",
                            page_name,
                        )
                        page_results.append(page_result)
                        if not page_result["active_page_visible"]:
                            failures.append(f"{viewport['name']}: {page_name} page did not become visible")
                        if not page_result["active_tab_selected"]:
                            failures.append(f"{viewport['name']}: {page_name} tab did not become active")
                        if not page_result["no_horizontal_overflow"]:
                            failures.append(
                                f"{viewport['name']}: {page_name} overflow {page_result['document_width']}px > {page_result['viewport_width']}px"
                            )
                    result["name"] = viewport["name"]
                    result["page_results"] = page_results
                    report["viewports"].append(result)
                    for check in result["checks"]:
                        if not check["ok"]:
                            failures.append(f"{viewport['name']}: {check['name']} - {check['detail']}")
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
    print(f"Dashboard layout contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
