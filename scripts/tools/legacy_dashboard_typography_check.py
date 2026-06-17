import argparse
import json
import sys
from datetime import datetime, timezone
from urllib.parse import urljoin

from playwright.sync_api import sync_playwright


PAGE_READY_STATE = "domcontentloaded"
NAVIGATION_TIMEOUT_MS = 120000
READY_TIMEOUT_MS = 120000
ASYNC_LAYOUT_SETTLE_MS = 12000

PAGE_TARGETS = (
    "overview",
    "operations",
    "notifications",
    "deployment",
    "portfolio",
    "hosts",
    "admin",
)

FOCUS_SELECTORS = (
    ".executive-focus",
    ".monetisation-focus",
    ".business-focus",
    ".growth-focus",
    ".command-center-focus",
    ".buyer-ops-focus",
    ".paid-ops-focus",
    ".notification-revenue-focus",
    ".provider-simulation-focus",
    ".package-billing-focus",
    ".commercial-hero",
    ".suite-focus",
    ".hero-focus",
    ".commercial-focus",
    ".launch-focus",
)

CHECKS = (
    {"selector": "header h1", "max_size": 23.5, "max_weight": 660},
    {"selector": ".dashboard-page-tab.is-active", "max_size": 14.5, "max_weight": 620},
    {"selector": "#refresh-button", "max_size": 14.5, "max_weight": 620},
    {"selector": ".section-head h2", "max_size": 18.5, "max_weight": 640},
    {"selector": ".badge.pending", "max_size": 13.5, "max_weight": 620},
    {"selector": "#runtime-status-headline", "max_size": 28.5, "max_weight": 700},
    {"selector": "#local-indicator-headline", "max_size": 28.5, "max_weight": 700},
    {"selector": "#runtime-kpi-backend", "max_size": 22.5, "max_weight": 700},
    {"selector": "#runtime-status-source", "max_size": 14.5, "max_weight": 700},
    {"selector": '[data-page-target="deployment"]', "max_size": 15.5, "max_weight": 700},
)


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck v1-old typography contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    base_url = args.base_url.rstrip("/") + "/"
    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": base_url,
        "privacy_boundary": "computed style metrics only; no screenshots, credentials, cookies, tokens, raw URLs, or private content capture",
        "checks": [],
    }
    failures = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        try:
            page = browser.new_page(viewport={"width": 1440, "height": 1000})
            try:
                page.goto(urljoin(base_url, "v1-old"), wait_until=PAGE_READY_STATE, timeout=NAVIGATION_TIMEOUT_MS)
                page.wait_for_selector("#dashboard-page-nav", state="visible", timeout=READY_TIMEOUT_MS)
                page.click('[data-page-target="deployment"]')
                page.wait_for_selector("#runtime-status-headline", state="visible", timeout=READY_TIMEOUT_MS)
                page.wait_for_timeout(ASYNC_LAYOUT_SETTLE_MS)
                for check in CHECKS:
                    result = page.evaluate(
                        """({ selector, maxSize, maxWeight }) => {
                          const element = document.querySelector(selector);
                          if (!element) {
                            return {
                              selector,
                              ok: false,
                              detail: "missing",
                              size: 0,
                              weight: 0,
                              text: ""
                            };
                          }
                          const style = window.getComputedStyle(element);
                          const size = Number.parseFloat(style.fontSize || "0");
                          const weight = Number.parseFloat(style.fontWeight || "0");
                          const text = element.textContent.trim();
                          return {
                            selector,
                            ok: size <= maxSize && weight <= maxWeight,
                            detail: `${text}: ${size}px / ${weight}`,
                            size,
                            weight,
                            text
                          };
                        }""",
                        {
                            "selector": check["selector"],
                            "maxSize": check["max_size"],
                            "maxWeight": check["max_weight"],
                        },
                    )
                    report["checks"].append(result)
                    if not result["ok"]:
                        failures.append(
                            f"{check['selector']} exceeded legacy typography limit: {result['detail']}"
                        )
                width_result = page.evaluate(
                    """() => {
                      const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                      const viewportWidth = window.innerWidth;
                      return {
                        selector: "document-horizontal-overflow",
                        ok: documentWidth <= viewportWidth + 2,
                        detail: `${documentWidth}px document width within ${viewportWidth}px viewport`,
                        size: documentWidth,
                        weight: 0,
                        text: ""
                      };
                    }"""
                )
                report["checks"].append(width_result)
                if not width_result["ok"]:
                    failures.append(width_result["detail"])
                for target in PAGE_TARGETS:
                    page.click(f'[data-page-target="{target}"]')
                    page.wait_for_timeout(650)
                    page_result = page.evaluate(
                        """({ target, focusSelectors }) => {
                          const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                          const viewportWidth = window.innerWidth;
                          const visible = (element) => {
                            const rect = element.getBoundingClientRect();
                            const style = window.getComputedStyle(element);
                            return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden";
                          };
                          const focusWhitespace = Array.from(document.querySelectorAll(focusSelectors.join(",")))
                            .filter(visible)
                            .map((element) => {
                              const rect = element.getBoundingClientRect();
                              const children = Array.from(element.children).filter(visible);
                              if (children.length === 0) {
                                return {
                                  selector: element.className,
                                  height: Number(rect.height.toFixed(2)),
                                  bottomWhitespace: Number(rect.height.toFixed(2))
                                };
                              }
                              const first = children[0].getBoundingClientRect();
                              const last = children[children.length - 1].getBoundingClientRect();
                              const used = last.bottom - first.top;
                              return {
                                selector: element.className,
                                height: Number(rect.height.toFixed(2)),
                                bottomWhitespace: Number(Math.max(0, rect.height - used).toFixed(2))
                              };
                            });
                          const excessiveWhitespace = focusWhitespace.filter((item) => item.bottomWhitespace > 96);
                          return {
                            selector: `legacy-tab-${target}`,
                            ok: documentWidth <= viewportWidth + 2 && excessiveWhitespace.length === 0,
                            detail: `${documentWidth}px/${viewportWidth}px, whitespace=${excessiveWhitespace.map((item) => `${item.selector}:${item.bottomWhitespace}`).join(", ") || "ok"}`,
                            size: documentWidth,
                            weight: 0,
                            text: target,
                            focusWhitespace
                          };
                        }""",
                        {"target": target, "focusSelectors": list(FOCUS_SELECTORS)},
                    )
                    report["checks"].append(page_result)
                    if not page_result["ok"]:
                        failures.append(f"{target} legacy tab layout failed: {page_result['detail']}")
                page.evaluate("window.scrollTo(0, document.documentElement.scrollHeight)")
                page.wait_for_timeout(650)
                bottom_result = page.evaluate(
                    """() => {
                      const nav = document.querySelector("#command-navigation.command-nav");
                      if (!nav) {
                        return {
                          selector: "legacy-workspace-navigator-bottom-overlap",
                          ok: true,
                          detail: "workspace navigator absent",
                          size: 0,
                          weight: 0,
                          text: ""
                        };
                      }
                      const visible = (element) => {
                        const rect = element.getBoundingClientRect();
                        const style = window.getComputedStyle(element);
                        return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden";
                      };
                      const navRect = nav.getBoundingClientRect();
                      const overlaps = Array.from(document.querySelectorAll("main > section:not([hidden])"))
                        .filter((element) => element !== nav && visible(element))
                        .map((element) => {
                          const rect = element.getBoundingClientRect();
                          return {
                            id: element.id,
                            className: element.className,
                            overlaps: rect.right > navRect.left && rect.left < navRect.right && rect.bottom > navRect.top && rect.top < navRect.bottom
                          };
                        })
                        .filter((item) => item.overlaps);
                      return {
                        selector: "legacy-workspace-navigator-bottom-overlap",
                        ok: overlaps.length === 0,
                        detail: overlaps.map((item) => item.id || item.className || "section").join(", ") || "no overlap at bottom scroll",
                        size: overlaps.length,
                        weight: 0,
                        text: nav.textContent.trim().slice(0, 80)
                      };
                    }"""
                )
                report["checks"].append(bottom_result)
                if not bottom_result["ok"]:
                    failures.append(f"legacy workspace navigator overlaps bottom content: {bottom_result['detail']}")
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
    print(f"Legacy dashboard typography contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
