import argparse
import json
import sys
from datetime import datetime, timezone
from urllib.parse import parse_qsl, urlencode, urlsplit, urlunsplit

from playwright.sync_api import sync_playwright


NAVIGATION_TIMEOUT_MS = 120000
READY_TIMEOUT_MS = 120000


def dashboard_url(base_url: str) -> str:
    parts = urlsplit(base_url.rstrip("/") + "/")
    query = dict(parse_qsl(parts.query, keep_blank_values=True))
    query["layout"] = "all"
    query["include_demo"] = "true"
    return urlunsplit((parts.scheme, parts.netloc, parts.path or "/", urlencode(query), parts.fragment))


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck delivery card UI contract")
    parser.add_argument("--base-url", required=True)
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "target": dashboard_url(args.base_url),
        "privacy_boundary": "rendered DOM metrics only; no screenshots, credentials, cookies, tokens, raw URLs, page titles, or private content capture",
        "checks": [],
    }
    failures = []

    with sync_playwright() as playwright:
        browser = playwright.chromium.launch()
        try:
            page = browser.new_page(viewport={"width": 1280, "height": 900})
            try:
                page.goto(report["target"], wait_until="domcontentloaded", timeout=NAVIGATION_TIMEOUT_MS)
                page.wait_for_selector("#delivery-list", state="visible", timeout=READY_TIMEOUT_MS)
                page.wait_for_function(
                    "document.querySelectorAll('#delivery-list .delivery-card').length >= 3",
                    timeout=READY_TIMEOUT_MS,
                )
                result = page.evaluate(
                    """() => {
                      const list = document.getElementById("delivery-list");
                      const cards = Array.from(document.querySelectorAll("#delivery-list .delivery-card"));
                      const text = list ? list.innerText : "";
                      const cardMetrics = cards.map((card) => {
                        const rect = card.getBoundingClientRect();
                        return {
                          className: card.className,
                          width: Math.round(rect.width),
                          scrollWidth: card.scrollWidth,
                          clientWidth: card.clientWidth,
                          scrollHeight: card.scrollHeight,
                          clientHeight: card.clientHeight
                        };
                      });
                      return {
                        tableAbsent: document.getElementById("delivery-table") === null,
                        listVisible: Boolean(list && list.offsetParent !== null),
                        cardCount: cards.length,
                        text,
                        cardMetrics
                      };
                    }"""
                )
            finally:
                page.close()
        finally:
            browser.close()

    checks = [
        ("delivery-table-removed", result["tableAbsent"], "old cramped table must not render"),
        ("delivery-list-visible", result["listVisible"], "#delivery-list is visible"),
        ("delivery-cards-present", result["cardCount"] >= 3, f"{result['cardCount']} delivery cards"),
        ("demo-copy-visible", "Demo only" in result["text"], "demo seed row has buyer-safe label"),
        (
            "push-notification-truth-visible",
            "Screen notification was not sent" in result["text"],
            "demo push cannot look like a real screen notification",
        ),
        (
            "provider-send-denial-visible",
            "No web-push provider send was attempted" in result["text"],
            "demo push text denies provider send",
        ),
    ]
    for name, ok, detail in checks:
        report["checks"].append({"name": name, "ok": bool(ok), "detail": detail})
        if not ok:
            failures.append(f"{name}: {detail}")

    for index, metric in enumerate(result["cardMetrics"], start=1):
        ok = metric["scrollWidth"] <= metric["clientWidth"] + 2
        detail = f"{metric['scrollWidth']}px scroll within {metric['clientWidth']}px client"
        report["checks"].append({"name": f"delivery-card-{index}-horizontal-fit", "ok": ok, "detail": detail})
        if not ok:
            failures.append(f"delivery-card-{index}-horizontal-fit: {detail}")

    report["ok"] = len(failures) == 0
    report["failures"] = failures
    with open(args.output, "w", encoding="utf-8") as handle:
        json.dump(report, handle, indent=2)

    if failures:
        for failure in failures:
            print(failure, file=sys.stderr)
        return 1
    print(f"Dashboard delivery UI contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
