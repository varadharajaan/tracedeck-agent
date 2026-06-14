import argparse
import importlib.util
import json
import sys
from datetime import datetime, timezone
from pathlib import Path

from playwright.sync_api import sync_playwright


VIEWPORTS = (
    {"name": "desktop", "width": 1440, "height": 960},
    {"name": "mobile", "width": 390, "height": 844},
)

THEMES = ("light", "dark")


def load_lambda_module():
    module_path = Path("sam-app/frontend_function/app.py")
    spec = importlib.util.spec_from_file_location("tracedeck_lambda_frontend", module_path)
    module = importlib.util.module_from_spec(spec)
    assert spec.loader is not None
    spec.loader.exec_module(module)
    return module


def sample_summary() -> dict:
    return {
        "status": "ok",
        "bucket": "tracedeck-phase80-visual",
        "prefix": "tenant=family-varadha/",
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "privacy_boundary": "metadata-only Lambda visual test payload; no screenshots, raw URLs, page titles, cookies, tokens, private content, or provider secrets",
        "summary": {
            "objects": 2,
            "bytes": 18432,
            "sampled_rows": 3,
            "study_safe": 2,
            "non_study_youtube": 1,
            "latest_object_at": "2026-06-14T03:30:00Z",
        },
        "cache": {
            "hit": True,
            "hits": 3,
            "misses": 1,
            "hit_percent": 75,
            "miss_percent": 25,
            "ttl_seconds": 300,
            "cached_at": "2026-06-14T03:30:00Z",
        },
        "hosts": [
            {"label": "demo-study-laptop", "total": 3, "study_safe": 2, "non_study": 1, "last_observed_at": "2026-06-14T03:30:00Z"}
        ],
        "browsers": [
            {"label": "chrome", "total": 1, "study_safe": 1, "non_study": 0, "last_observed_at": "2026-06-14T03:28:00Z"},
            {"label": "edge", "total": 2, "study_safe": 1, "non_study": 1, "last_observed_at": "2026-06-14T03:30:00Z"},
        ],
        "browser_rows": [
            {
                "host_name": "demo-study-laptop",
                "device_id": "demo-study-laptop",
                "browser": "chrome",
                "domain": "docs.python.org",
                "category": "study",
                "study_safe": True,
                "source_kind": "s3_sample",
                "evidence_scope": "metadata_only",
                "evidence_detail": "Sampled from S3 archive metadata for cloud admin rendering.",
                "observed_at": "2026-06-14T03:28:00Z",
            },
            {
                "host_name": "demo-study-laptop",
                "device_id": "demo-study-laptop",
                "browser": "edge",
                "domain": "youtube.com",
                "category": "video-streaming",
                "study_safe": False,
                "source_kind": "s3_sample",
                "evidence_scope": "metadata_only",
                "evidence_detail": "Domain-only non-study YouTube review row.",
                "observed_at": "2026-06-14T03:30:00Z",
            },
        ],
        "objects": [
            {"key": "tenant=family-varadha/device=demo-study-laptop/sample.jsonl.gz", "size": 9216, "storage_class": "STANDARD", "last_modified": "2026-06-14T03:30:00Z"},
            {"key": "tenant=family-varadha/device=demo-study-laptop/sample-2.jsonl.gz", "size": 9216, "storage_class": "STANDARD", "last_modified": "2026-06-14T03:20:00Z"},
        ],
    }


def main() -> int:
    parser = argparse.ArgumentParser(description="TraceDeck Lambda frontend visual quality contract")
    parser.add_argument("--output", required=True)
    args = parser.parse_args()

    module = load_lambda_module()
    html_response = module.lambda_handler({"rawPath": "/", "requestContext": {"http": {"method": "GET"}}}, None)
    html = html_response["body"]
    payload = sample_summary()

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "privacy_boundary": "rendered Lambda frontend metrics only; no screenshots, credentials, cookies, tokens, raw URLs, page titles, or private content capture",
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

                        def route_handler(route):
                            url = route.request.url
                            if url.rstrip("/") == "https://tracedeck-cloud.local":
                                route.fulfill(status=200, content_type="text/html; charset=utf-8", body=html)
                                return
                            if "/api/s3-summary" in url:
                                route.fulfill(status=200, content_type="application/json; charset=utf-8", body=json.dumps(payload))
                                return
                            route.fulfill(status=404, content_type="text/plain", body="not found")

                        page.route("**/*", route_handler)
                        page.goto("https://tracedeck-cloud.local/", wait_until="networkidle", timeout=60000)
                        page.wait_for_selector("#server-status.connected", state="visible", timeout=30000)
                        result = page.evaluate(
                            """({ theme }) => {
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
                              const status = document.getElementById("server-status");
                              const themeLabel = document.getElementById("theme-label");
                              const brandMark = document.querySelector(".brand-mark");
                              const appShell = document.querySelector(".app-shell");
                              const sideRail = document.querySelector(".side-rail");
                              const heroPanel = document.querySelector(".hero-panel");
                              const sourceCard = document.querySelector(".source-card");
                              const navLabels = Array.from(document.querySelectorAll(".tabs .tab")).map((button) => button.textContent.trim());
                              const bodyText = document.body.innerText || "";
                              const buttons = Array.from(document.querySelectorAll("button")).map((button) => button.textContent.trim());
                              const tinyInteractive = Array.from(document.querySelectorAll("button, .tab, .pill, .status"))
                                .filter((element) => {
                                  const rect = element.getBoundingClientRect();
                                  const style = window.getComputedStyle(element);
                                  return rect.width > 0 && rect.height > 0 && style.display !== "none" && style.visibility !== "hidden" && rect.height < 26;
                                })
                                .map((element) => `${element.textContent.trim().slice(0, 80)}:${element.getBoundingClientRect().height.toFixed(2)}`);
                              const bg = root.getPropertyValue("--bg").trim();
                              const surface = root.getPropertyValue("--surface").trim();
                              const bgLuma = luminance(parseColor(bg));
                              const surfaceLuma = luminance(parseColor(surface));
                              const documentWidth = Math.max(document.documentElement.scrollWidth, document.body.scrollWidth);
                              const viewportWidth = window.innerWidth;
                              const forbiddenText = /(Browser\\s*\\{|Center\\s*\\{|\\[[BCTR]\\]|\\{[BCTR]\\}|\\bTD\\b|^T$|\\bRev Ops\\b|\\bNotif Rev\\b|\\bNotify Pro\\b|\\bNotifs\\b)/i.test(bodyText);
                              return {
                                h1: h1 ? h1.textContent.trim() : "",
                                status_text: status ? status.textContent.trim() : "",
                                theme_label: themeLabel ? themeLabel.textContent.trim() : "",
                                button_labels: buttons,
                                nav_labels: navLabels,
                                checks: [
                                  {
                                    name: "expected-heading",
                                    ok: h1 && h1.textContent.trim() === "TraceDeck Cloud Admin",
                                    detail: h1 ? h1.textContent.trim() : "missing"
                                  },
                                  {
                                    name: "product-shell-present",
                                    ok: Boolean(appShell && sideRail && heroPanel && sourceCard),
                                    detail: `app=${Boolean(appShell)} side=${Boolean(sideRail)} hero=${Boolean(heroPanel)} source=${Boolean(sourceCard)}`
                                  },
                                  {
                                    name: "symbolic-brand-mark",
                                    ok: Boolean(brandMark && brandMark.textContent.trim() === "" && brandMark.querySelectorAll("span").length >= 3),
                                    detail: brandMark ? `text=${brandMark.textContent.trim() || "empty"} marks=${brandMark.querySelectorAll("span").length}` : "missing"
                                  },
                                  {
                                    name: "clear-page-navigation",
                                    ok: ["Overview", "Browser Activity", "S3 Archive", "Source & Cache"].every((label) => navLabels.includes(label)),
                                    detail: navLabels.join(", ")
                                  },
                                  {
                                    name: "source-controls-present",
                                    ok: Boolean(document.getElementById("source-select") && bodyText.includes("Workspace Source") && bodyText.includes("Lambda S3 Archive") && bodyText.includes("Localhost 18080")),
                                    detail: bodyText.includes("Workspace Source") ? "workspace source visible" : "workspace source missing"
                                  },
                                  {
                                    name: "connected-status-visible",
                                    ok: Boolean(status && status.classList.contains("connected") && status.getBoundingClientRect().height >= 26),
                                    detail: status ? status.textContent.trim() : "missing"
                                  },
                                  {
                                    name: "theme-label-current",
                                    ok: themeLabel && themeLabel.textContent.trim() === (theme === "dark" ? "Theme: Dark" : "Theme: Light"),
                                    detail: themeLabel ? themeLabel.textContent.trim() : "missing"
                                  },
                                  {
                                    name: "no-pseudo-letter-buttons",
                                    ok: buttons.every((label) => label !== "T" && !/[\\[\\]{}]/.test(label)),
                                    detail: buttons.join(", ")
                                  },
                                  {
                                    name: "no-stale-debug-copy",
                                    ok: !forbiddenText,
                                    detail: forbiddenText ? "found stale debug copy" : "no stale debug copy"
                                  },
                                  {
                                    name: "no-tiny-interactive-labels",
                                    ok: tinyInteractive.length === 0,
                                    detail: tinyInteractive.join(", ") || "all controls at least 26px tall"
                                  },
                                  {
                                    name: "theme-luminance",
                                    ok: theme === "dark" ? bgLuma < 0.08 && surfaceLuma < 0.14 : bgLuma > 0.78 && surfaceLuma > 0.92,
                                    detail: `bg=${bg}(${bgLuma.toFixed(3)}) surface=${surface}(${surfaceLuma.toFixed(3)})`
                                  },
                                  {
                                    name: "no-horizontal-overflow",
                                    ok: documentWidth <= viewportWidth + 2,
                                    detail: `${documentWidth}px document width within ${viewportWidth}px viewport`
                                  },
                                  {
                                    name: "cache-metrics-rendered",
                                    ok: (document.getElementById("metric-hit") || {}).textContent === "75%" && (document.getElementById("metric-miss") || {}).textContent === "25%",
                                    detail: `${(document.getElementById("metric-hit") || {}).textContent || ""}/${(document.getElementById("metric-miss") || {}).textContent || ""}`
                                  }
                                ]
                              };
                            }""",
                            {"theme": theme},
                        )
                        entry = {"page": "lambda_frontend", "theme": theme, "viewport": viewport, **result}
                        report["checks"].append(entry)
                        for check in result["checks"]:
                            if not check["ok"]:
                                failures.append(f"lambda_frontend {theme} {viewport['name']}: {check['name']} - {check['detail']}")
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
    print(f"Lambda frontend visual quality contract passed: {args.output}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
