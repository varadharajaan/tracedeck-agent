import argparse
import json
import sys
from datetime import datetime, timezone
from urllib.parse import urljoin

from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from playwright.sync_api import sync_playwright


NAVIGATION_TIMEOUT_MS = 120000
READY_TIMEOUT_MS = 120000
VIEWPORTS = (
    {"name": "desktop", "width": 1280, "height": 900},
    {"name": "mobile", "width": 390, "height": 844},
)
THEMES = ("light", "dark")
STATUS_LABELS = {"attention", "healthy", "pending", "watch", "ok"}


def browser_activity_url(base_url: str) -> str:
    return urljoin(base_url.rstrip("/") + "/", "browser-activity")


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck Browser Activity badge layout contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "target": browser_activity_url(args.base_url),
        "privacy_boundary": (
            "rendered badge metrics only; no screenshots, credentials, cookies, tokens, "
            "raw URLs, page titles, private content, or browser history content capture"
        ),
        "checks": [],
    }
    failures = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch(headless=True)
        try:
            for viewport in VIEWPORTS:
                for theme in THEMES:
                    page = browser.new_page(viewport={"width": viewport["width"], "height": viewport["height"]})
                    try:
                        page.add_init_script(
                            f"""window.localStorage.setItem("tracedeck.ui.theme", {json.dumps(theme)});"""
                        )
                        page.goto(report["target"], wait_until="domcontentloaded", timeout=NAVIGATION_TIMEOUT_MS)
                        page.wait_for_selector("#activity-table", state="visible", timeout=READY_TIMEOUT_MS)
                        try:
                            page.wait_for_function(
                                """() => {
                                  const status = document.getElementById("page-status");
                                  return status && !/checking/i.test(status.textContent || "");
                                }""",
                                timeout=READY_TIMEOUT_MS,
                            )
                        except PlaywrightTimeoutError:
                            pass

                        result = page.evaluate(
                            """({ statusLabels }) => {
                              const table = document.getElementById("activity-table");
                              if (table) {
                                const row = document.createElement("tr");
                                row.setAttribute("data-contract-fixture", "badge-table-wrap");
                                row.innerHTML = `
                                  <td>contract-host</td>
                                  <td>contract-browser</td>
                                  <td>contract-domain</td>
                                  <td>contract-category</td>
                                  <td class="signal-cell"><span class="badge attention" data-contract-badge="table">attention</span><span class="muted">contract detail wraps here</span></td>
                                  <td>1</td>
                                  <td><span class="pill">contract-route</span></td>
                                  <td><span class="pill">contract-source</span></td>
                                  <td>contract-time</td>
                                `;
                                table.appendChild(row);
                              }

                              const fixture = document.createElement("div");
                              fixture.setAttribute("data-contract-fixture", "badge-narrow-wrap");
                              fixture.style.cssText = "position:absolute;left:0;top:0;width:44px;pointer-events:none;visibility:hidden;";
                              fixture.innerHTML = '<span class="badge attention" data-contract-badge="narrow">attention</span>';
                              document.body.appendChild(fixture);

                              const visible = (element) => {
                                const rect = element.getBoundingClientRect();
                                const style = window.getComputedStyle(element);
                                return rect.width > 0 && rect.height > 0 && style.display !== "none";
                              };
                              const number = (value) => {
                                const parsed = Number.parseFloat(value);
                                return Number.isFinite(parsed) ? parsed : 0;
                              };
                              const lineHeight = (style) => {
                                if (style.lineHeight === "normal") {
                                  return number(style.fontSize) * 1.2;
                                }
                                return number(style.lineHeight);
                              };
                              const badgeMetric = (element, index) => {
                                const rect = element.getBoundingClientRect();
                                const style = window.getComputedStyle(element);
                                const clean = (element.textContent || "").trim().toLowerCase();
                                const paddingY = number(style.paddingTop) + number(style.paddingBottom);
                                const borderY = number(style.borderTopWidth) + number(style.borderBottomWidth);
                                const minHeight = number(style.minHeight);
                                const singleLineHeight = Math.max(minHeight, lineHeight(style) + paddingY + borderY);
                                const folded = rect.height > singleLineHeight + 5;
                                const clipped = style.overflow !== "visible" && element.scrollWidth > element.clientWidth + 2;
                                const unsafeWrap = style.whiteSpace !== "nowrap" || style.overflowWrap !== "normal" || style.wordBreak === "break-all";
                                return {
                                  index,
                                  text: clean,
                                  contract: Boolean(element.dataset.contractBadge),
                                  short_status: statusLabels.includes(clean),
                                  width: Math.round(rect.width),
                                  height: Math.round(rect.height),
                                  client_width: element.clientWidth,
                                  scroll_width: element.scrollWidth,
                                  single_line_height: Math.round(singleLineHeight),
                                  white_space: style.whiteSpace,
                                  overflow_wrap: style.overflowWrap,
                                  word_break: style.wordBreak,
                                  overflow: style.overflow,
                                  folded,
                                  clipped,
                                  unsafe_wrap: unsafeWrap
                                };
                              };
                              const badges = Array.from(document.querySelectorAll(".badge"))
                                .filter(visible)
                                .map(badgeMetric);
                              return {
                                page_status: (document.getElementById("page-status")?.textContent || "").trim(),
                                badge_count: badges.length,
                                badges,
                                document_width: Math.max(document.documentElement.scrollWidth, document.body.scrollWidth),
                                viewport_width: window.innerWidth
                              };
                            }""",
                            {"statusLabels": sorted(STATUS_LABELS)},
                        )
                    finally:
                        page.close()

                    entry = {
                        "viewport": viewport,
                        "theme": theme,
                        **result,
                    }
                    report["checks"].append(entry)

                    contract_badges = [badge for badge in result["badges"] if badge["contract"]]
                    if len(contract_badges) < 2:
                        failures.append(f"{theme} {viewport['name']}: expected both badge contract fixtures")
                    for badge in result["badges"]:
                        if not (badge["contract"] or badge["short_status"]):
                            continue
                        label = f"{theme} {viewport['name']} badge '{badge['text']}'"
                        if badge["folded"]:
                            failures.append(
                                f"{label}: folded height {badge['height']}px exceeds single-line {badge['single_line_height']}px"
                            )
                        if badge["clipped"]:
                            failures.append(
                                f"{label}: clipped width {badge['scroll_width']}px scroll within {badge['client_width']}px client"
                            )
                        if badge["unsafe_wrap"]:
                            failures.append(
                                f"{label}: unsafe wrap css white-space={badge['white_space']} overflow-wrap={badge['overflow_wrap']} word-break={badge['word_break']}"
                            )
                    if result["document_width"] > result["viewport_width"] + 2:
                        failures.append(
                            f"{theme} {viewport['name']}: document width {result['document_width']}px exceeds viewport {result['viewport_width']}px"
                        )
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
    print(f"Browser Activity badge contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
