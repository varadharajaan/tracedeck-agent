import argparse
import json
import sys
from datetime import datetime, timezone

from playwright.sync_api import sync_playwright


VIEWPORTS = (
    {"name": "desktop", "width": 1440, "height": 1000},
    {"name": "tablet", "width": 900, "height": 900},
    {"name": "mobile", "width": 390, "height": 844},
)

REQUIRED_IDS = (
    "executive-console-section",
    "executive-console-status",
    "executive-console-headline",
    "executive-tile-list",
    "executive-alert-list",
    "executive-delivery-list",
    "executive-action-list",
    "notification-revenue-section",
    "notification-revenue-status",
    "notification-revenue-headline",
    "notification-revenue-kpi-list",
    "notification-revenue-scenario-list",
    "notification-revenue-channel-list",
    "notification-revenue-action-list",
    "provider-simulation-section",
    "provider-simulation-status",
    "provider-simulation-headline",
    "provider-simulation-route-list",
    "provider-simulation-scenario-list",
    "provider-simulation-action-list",
    "package-billing-section",
    "package-billing-status",
    "package-billing-headline",
    "package-billing-plan-list",
    "package-billing-feature-list",
    "package-billing-milestone-list",
    "package-billing-action-list",
    "business-dashboard-section",
    "business-dashboard-status",
    "business-dashboard-headline",
    "business-alert-list",
    "business-channel-list",
    "business-package-list",
    "business-action-list",
    "command-navigation",
    "command-nav-status",
    "command-nav-title",
    "growth-cockpit-section",
    "growth-cockpit-status",
    "growth-alert-ops-list",
    "growth-delivery-proof-list",
    "growth-owner-action-list",
    "notification-preference-section",
    "notification-preference-status",
    "notification-preference-rule-list",
    "notification-preference-suppression-list",
    "notification-preference-action-list",
    "role-experience-section",
    "role-experience-status",
    "role-experience-card-list",
    "role-onboarding-list",
    "monetization-command-center-section",
    "command-center-status",
    "command-center-inbox-list",
    "command-center-delivery-list",
    "command-center-action-list",
    "premium-notification-section",
    "premium-alert-funnel-list",
    "premium-delivery-proof-list",
    "premium-action-sla-list",
    "delivery-timeline-section",
    "delivery-timeline-status",
    "delivery-timeline-list",
    "buyer-ops-section",
    "buyer-ops-status",
    "buyer-delivery-list",
    "buyer-package-list",
    "buyer-action-list",
    "delivery-drilldown-section",
    "delivery-drill-route-list",
    "delivery-drill-action-list",
    "paid-ops-section",
    "revenue-section",
    "notification-proof-section",
    "push-activation-section",
    "push-activation-status",
    "push-activation-headline",
    "push-route-list",
    "push-scenario-list",
    "push-action-list",
    "push-guard-list",
    "portfolio-center-section",
    "portfolio-center-status",
    "portfolio-center-headline",
    "portfolio-alert-list",
    "portfolio-delivery-proof-list",
    "portfolio-host-list",
    "portfolio-segment-list",
    "portfolio-action-list",
    "portfolio-guard-list",
    "account-portfolio-section",
    "account-portfolio-status",
    "account-portfolio-headline",
    "account-tenant-list",
    "account-proof-list",
    "account-action-list",
    "mail-report-section",
    "archive-proof-section",
    "trust-proof-section",
    "host-detail-section",
)


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck dashboard layout contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "base_url": args.base_url,
        "privacy_boundary": "layout metrics only; no screenshots, video, credentials, or page content capture",
        "viewports": [],
    }
    failures = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        try:
            for viewport in VIEWPORTS:
                page = browser.new_page(viewport={"width": viewport["width"], "height": viewport["height"]})
                try:
                    page.goto(args.base_url, wait_until="networkidle", timeout=60000)
                    page.wait_for_selector("#command-navigation", state="visible", timeout=30000)
                    page.wait_for_function(
                        "() => document.getElementById('command-nav-title')"
                        " && !document.getElementById('command-nav-title').textContent.includes('No tenant loaded')",
                        timeout=30000,
                    )
                    result = page.evaluate(
                        """(requiredIDs) => {
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
                          const boxesOverlap = (left, right) => {
                            const horizontal = left.right > right.x + 1 && right.right > left.x + 1;
                            const vertical = left.bottom > right.y + 1 && right.bottom > left.y + 1;
                            return horizontal && vertical;
                          };

                          const checks = [];
                          const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                          const viewportWidth = window.innerWidth;
                          const allElements = Array.from(document.body.querySelectorAll("*")).map((element) => {
                            const rect = rounded(element.getBoundingClientRect());
                            return {
                              tag: element.tagName.toLowerCase(),
                              id: element.id || "",
                              className: typeof element.className === "string" ? element.className : "",
                              rect
                            };
                          }).filter((item) => item.rect.width > 0 && item.rect.height > 0)
                            .sort((left, right) => right.rect.right - left.rect.right)
                            .slice(0, 8);
                          checks.push({
                            name: "document-horizontal-overflow",
                            ok: documentWidth <= viewportWidth + 2,
                            detail: `${documentWidth}px document width within ${viewportWidth}px viewport`
                          });

                          for (const id of requiredIDs) {
                            const element = document.getElementById(id);
                            const rect = element ? rounded(element.getBoundingClientRect()) : null;
                            checks.push({
                              name: `required-visible-${id}`,
                              ok: visible(element),
                              detail: rect ? `${rect.width}x${rect.height} at ${rect.x},${rect.y}` : "missing",
                              rect
                            });
                            if (rect) {
                              checks.push({
                                name: `required-width-${id}`,
                                ok: rect.width <= viewportWidth + 2,
                                detail: `${rect.width}px within ${viewportWidth}px viewport`
                              });
                            }
                          }

                          const jumps = Array.from(document.querySelectorAll("[data-jump-target]"));
                          const jumpBoxes = jumps.map((element) => ({
                            target: element.getAttribute("data-jump-target"),
                            label: element.firstChild ? element.firstChild.textContent.trim() : element.textContent.trim(),
                            rect: rounded(element.getBoundingClientRect()),
                            scrollWidth: element.scrollWidth,
                            clientWidth: element.clientWidth,
                            scrollHeight: element.scrollHeight,
                            clientHeight: element.clientHeight
                          }));
                          checks.push({
                            name: "command-navigation-has-seventeen-jumps",
                            ok: jumpBoxes.length === 17,
                            detail: `${jumpBoxes.length} command jump buttons`
                          });
                          for (const item of jumpBoxes) {
                            checks.push({
                              name: `command-jump-text-fit-${item.target}`,
                              ok: item.scrollWidth <= item.clientWidth + 2 && item.scrollHeight <= item.clientHeight + 2,
                              detail: `${item.scrollWidth}x${item.scrollHeight} scroll within ${item.clientWidth}x${item.clientHeight} client`
                            });
                            checks.push({
                              name: `command-jump-target-exists-${item.target}`,
                              ok: Boolean(document.getElementById(item.target)),
                              detail: item.target
                            });
                          }
                          for (let leftIndex = 0; leftIndex < jumpBoxes.length; leftIndex += 1) {
                            for (let rightIndex = leftIndex + 1; rightIndex < jumpBoxes.length; rightIndex += 1) {
                              if (boxesOverlap(jumpBoxes[leftIndex].rect, jumpBoxes[rightIndex].rect)) {
                                checks.push({
                                  name: `command-jump-overlap-${jumpBoxes[leftIndex].target}-${jumpBoxes[rightIndex].target}`,
                                  ok: false,
                                  detail: "command jump buttons overlap"
                                });
                              }
                            }
                          }

                          const navStyle = window.getComputedStyle(document.getElementById("command-navigation"));
                          checks.push({
                            name: "command-navigation-position-contract",
                            ok: viewportWidth < 1180 ? navStyle.position === "static" : navStyle.position === "sticky",
                            detail: `${navStyle.position} at ${viewportWidth}px`
                          });

                          const activeBefore = document.querySelector("[data-jump-target].is-active");
                          const trustButton = document.querySelector('[data-jump-target="trust-proof-section"]');
                          if (trustButton) trustButton.click();
                          const activeAfter = document.querySelector("[data-jump-target].is-active");
                          checks.push({
                            name: "command-navigation-click-updates-active",
                            ok: Boolean(activeBefore) && Boolean(activeAfter) && activeAfter.getAttribute("data-jump-target") === "trust-proof-section",
                            detail: activeAfter ? activeAfter.getAttribute("data-jump-target") : "no active command target"
                          });

                          return {
                            viewport: { width: viewportWidth, height: window.innerHeight },
                            client_width: document.documentElement.clientWidth,
                            body_client_width: document.body.clientWidth,
                            document_width: documentWidth,
                            checks,
                            jump_boxes: jumpBoxes,
                            rightmost_elements: allElements
                          };
                        }""",
                        list(REQUIRED_IDS),
                    )
                    result["name"] = viewport["name"]
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
