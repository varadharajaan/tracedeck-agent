from __future__ import annotations

import base64
import gzip
import html
import json
import os
import time
from datetime import datetime, timezone
from typing import Any


DEFAULT_BUCKET = os.environ.get("TRACEDECK_DATA_BUCKET", "")
DEFAULT_PREFIX = os.environ.get("TRACEDECK_DATA_PREFIX", "")
DEFAULT_CACHE_TTL = int(os.environ.get("TRACEDECK_CACHE_TTL_SECONDS", "300"))
DEFAULT_LOCAL_BACKEND = os.environ.get("TRACEDECK_LOCAL_BACKEND_URL", "http://127.0.0.1:18080")
MAX_KEYS = int(os.environ.get("TRACEDECK_FRONTEND_MAX_KEYS", "1000"))
MAX_SAMPLE_OBJECTS = int(os.environ.get("TRACEDECK_FRONTEND_SAMPLE_OBJECTS", "24"))
MAX_SAMPLE_BYTES = int(os.environ.get("TRACEDECK_FRONTEND_SAMPLE_BYTES", "1048576"))
SOURCE_KIND_S3_SAMPLE = "s3_sample"
EVIDENCE_SCOPE_METADATA_ONLY = "metadata_only"
EVIDENCE_DETAIL_S3_SAMPLE = "Sampled from S3 archive metadata for cloud admin rendering."

_CACHE: dict[str, Any] = {
    "key": "",
    "expires_at": 0.0,
    "payload": None,
    "hits": 0,
    "misses": 0,
    "cached_at": "",
}


def lambda_handler(event: dict[str, Any], _context: Any) -> dict[str, Any]:
    method = _method(event)
    if method == "OPTIONS":
        return _empty_response(204)

    path = event.get("rawPath") or event.get("path") or "/"
    if path == "/favicon.ico":
        return _empty_response(204)
    if path.startswith("/api/health"):
        return _json_response({"status": "ok", "service": "tracedeck-lambda-frontend"})
    if path.startswith("/api/s3-summary"):
        return _json_response(_cached_s3_summary(_query(event)))
    return _html_response(_index_html())


def _method(event: dict[str, Any]) -> str:
    return (
        event.get("requestContext", {}).get("http", {}).get("method")
        or event.get("httpMethod")
        or "GET"
    ).upper()


def _query(event: dict[str, Any]) -> dict[str, str]:
    raw = event.get("queryStringParameters") or {}
    return {str(key): str(value) for key, value in raw.items() if value is not None}


def _headers(content_type: str) -> dict[str, str]:
    return {
        "Content-Type": content_type,
        "Cache-Control": "no-store",
        "Access-Control-Allow-Origin": "*",
        "Access-Control-Allow-Headers": "*",
        "Access-Control-Allow-Methods": "GET,OPTIONS",
    }


def _json_response(payload: dict[str, Any], status: int = 200) -> dict[str, Any]:
    return {
        "statusCode": status,
        "headers": _headers("application/json; charset=utf-8"),
        "body": json.dumps(payload, separators=(",", ":"), default=str),
    }


def _html_response(markup: str) -> dict[str, Any]:
    return {"statusCode": 200, "headers": _headers("text/html; charset=utf-8"), "body": markup}


def _empty_response(status: int) -> dict[str, Any]:
    return {"statusCode": status, "headers": _headers("text/plain; charset=utf-8"), "body": ""}


def _cached_s3_summary(query: dict[str, str]) -> dict[str, Any]:
    bucket = query.get("bucket") or DEFAULT_BUCKET
    prefix = query.get("prefix", DEFAULT_PREFIX)
    refresh = query.get("refresh", "false").lower() == "true"
    now = time.time()
    cache_key = f"{bucket}|{prefix}|{MAX_KEYS}|{MAX_SAMPLE_OBJECTS}"

    if not refresh and _CACHE["key"] == cache_key and _CACHE["payload"] and now < _CACHE["expires_at"]:
        _CACHE["hits"] += 1
        payload = dict(_CACHE["payload"])
        payload["cache"] = _cache_stats(hit=True)
        return payload

    _CACHE["misses"] += 1
    payload = _load_s3_summary(bucket=bucket, prefix=prefix)
    _CACHE.update(
        {
            "key": cache_key,
            "expires_at": now + DEFAULT_CACHE_TTL,
            "payload": payload,
            "cached_at": datetime.now(timezone.utc).isoformat(),
        }
    )
    response = dict(payload)
    response["cache"] = _cache_stats(hit=False)
    return response


def _cache_stats(hit: bool) -> dict[str, Any]:
    hits = int(_CACHE["hits"])
    misses = int(_CACHE["misses"])
    total = max(1, hits + misses)
    return {
        "hit": hit,
        "hits": hits,
        "misses": misses,
        "hit_percent": round((hits / total) * 100, 1),
        "miss_percent": round((misses / total) * 100, 1),
        "ttl_seconds": DEFAULT_CACHE_TTL,
        "cached_at": _CACHE.get("cached_at", ""),
        "expires_at": datetime.fromtimestamp(float(_CACHE["expires_at"]), timezone.utc).isoformat()
        if _CACHE.get("expires_at")
        else "",
    }


def _load_s3_summary(bucket: str, prefix: str) -> dict[str, Any]:
    if not bucket:
        return {
            "status": "not_configured",
            "bucket": "",
            "prefix": prefix,
            "summary": _empty_summary(),
            "objects": [],
            "browser_rows": [],
            "hosts": [],
            "browsers": [],
            "generated_at": datetime.now(timezone.utc).isoformat(),
        }

    import boto3

    client = boto3.client("s3")
    objects: list[dict[str, Any]] = []
    paginator = client.get_paginator("list_objects_v2")
    page_iterator = paginator.paginate(Bucket=bucket, Prefix=prefix, PaginationConfig={"MaxItems": MAX_KEYS})
    for page in page_iterator:
        for obj in page.get("Contents", []):
            objects.append(
                {
                    "key": obj["Key"],
                    "size": int(obj.get("Size", 0)),
                    "last_modified": obj.get("LastModified").isoformat() if obj.get("LastModified") else "",
                    "storage_class": obj.get("StorageClass", "STANDARD"),
                }
            )

    latest = sorted(objects, key=lambda item: item.get("last_modified", ""), reverse=True)
    rows = _sample_browser_rows(client, bucket, latest[:MAX_SAMPLE_OBJECTS])
    summary = _summarize(objects, rows)
    return {
        "status": "ok",
        "bucket": bucket,
        "prefix": prefix,
        "summary": summary,
        "objects": latest[:50],
        "browser_rows": rows[:100],
        "hosts": _group_counts(rows, "host_name", "device_id"),
        "browsers": _group_counts(rows, "browser", ""),
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "privacy_boundary": "metadata-only S3 summary: object keys, sizes, safe browser domains, categories, browsers, hosts, counts, and timestamps only",
    }


def _empty_summary() -> dict[str, Any]:
    return {
        "objects": 0,
        "bytes": 0,
        "sampled_rows": 0,
        "study_safe": 0,
        "non_study_youtube": 0,
        "latest_object_at": "",
    }


def _summarize(objects: list[dict[str, Any]], rows: list[dict[str, Any]]) -> dict[str, Any]:
    total_bytes = sum(int(item.get("size", 0)) for item in objects)
    latest_object_at = max((item.get("last_modified", "") for item in objects), default="")
    return {
        "objects": len(objects),
        "bytes": total_bytes,
        "sampled_rows": len(rows),
        "study_safe": sum(1 for row in rows if row.get("study_safe")),
        "non_study_youtube": sum(1 for row in rows if row.get("domain") == "youtube.com" and not row.get("study_safe")),
        "latest_object_at": latest_object_at,
    }


def _sample_browser_rows(client: Any, bucket: str, objects: list[dict[str, Any]]) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for obj in objects:
        key = str(obj.get("key", ""))
        if int(obj.get("size", 0)) > MAX_SAMPLE_BYTES:
            continue
        try:
            body = client.get_object(Bucket=bucket, Key=key)["Body"].read()
            text = _decode_object(key, body)
            for record in _iter_records(text):
                row = _browser_row(record, key)
                if row:
                    rows.append(row)
        except Exception:
            continue
    return rows


def _decode_object(key: str, body: bytes) -> str:
    if key.endswith(".gz"):
        body = gzip.decompress(body)
    return body.decode("utf-8", errors="replace")


def _iter_records(text: str) -> list[dict[str, Any]]:
    clean = text.strip()
    if not clean:
        return []
    if clean.startswith("["):
        parsed = json.loads(clean)
        return [item for item in parsed if isinstance(item, dict)]
    if clean.startswith("{") and "\n" not in clean:
        parsed = json.loads(clean)
        return [parsed] if isinstance(parsed, dict) else []
    records = []
    for line in clean.splitlines():
        line = line.strip()
        if not line:
            continue
        try:
            parsed = json.loads(line)
            if isinstance(parsed, dict):
                records.append(parsed)
        except json.JSONDecodeError:
            continue
    return records


def _browser_row(record: dict[str, Any], key: str) -> dict[str, Any] | None:
    domain = _field(record, "domain", "Domain")
    browser = _field(record, "browser", "Browser", "browser_name", "BrowserName")
    if not domain and "browser" not in str(record.get("type", record.get("Type", ""))).lower():
        return None
    domain = domain or _domain_from_key(key)
    if not domain:
        return None
    category = _field(record, "category", "Category", "signal", "Signal") or "unknown"
    explicit_study_safe = _field(record, "study_safe", "StudySafe")
    study_safe = (
        _coerce_bool(explicit_study_safe)
        if explicit_study_safe is not None
        else _infer_study_safe(domain, category, _bool_field(record, "youtube_study_match", "YouTubeStudyMatch"))
    )
    observed = _field(record, "observed_at", "ObservedAt", "last_observed_at", "LastObservedAt", "timestamp", "Timestamp")
    return {
        "host_name": _field(record, "host_name", "HostName") or _field(record, "device_id", "DeviceID") or "unknown-host",
        "device_id": _field(record, "device_id", "DeviceID") or "unknown-device",
        "browser": (browser or "browser").lower(),
        "domain": domain.lower(),
        "category": category,
        "study_safe": study_safe,
        "visit_count": int(_field(record, "visit_count", "VisitCount") or 1),
        "observed_at": observed or "",
        "source_key": key,
        "source_kind": _field(record, "source_kind", "SourceKind") or SOURCE_KIND_S3_SAMPLE,
        "evidence_scope": _field(record, "evidence_scope", "EvidenceScope") or EVIDENCE_SCOPE_METADATA_ONLY,
        "evidence_detail": _field(record, "evidence_detail", "EvidenceDetail") or EVIDENCE_DETAIL_S3_SAMPLE,
    }


def _field(record: dict[str, Any], *names: str) -> Any:
    containers = [record]
    for nested in ("payload", "Payload", "data", "Data", "attributes", "Attributes", "metadata", "Metadata"):
        value = record.get(nested)
        if isinstance(value, dict):
            containers.append(value)
    for container in containers:
        for name in names:
            if name in container and container[name] not in (None, ""):
                return container[name]
    lowered = {str(key).lower(): value for key, value in record.items()}
    for name in names:
        if name.lower() in lowered and lowered[name.lower()] not in (None, ""):
            return lowered[name.lower()]
    return None


def _bool_field(record: dict[str, Any], *names: str) -> bool:
    return _coerce_bool(_field(record, *names))


def _coerce_bool(value: Any) -> bool:
    if isinstance(value, bool):
        return value
    return str(value).lower() in ("true", "1", "yes", "study", "safe")


def _infer_study_safe(domain: str, category: str, youtube_study_match: bool) -> bool:
    category = str(category).lower().strip()
    domain = str(domain).lower().strip()
    if category == "study":
        return True
    return _is_youtube_domain(domain) and youtube_study_match


def _is_youtube_domain(domain: str) -> bool:
    return domain in ("youtube.com", "www.youtube.com", "youtu.be") or domain.endswith(".youtube.com")


def _domain_from_key(key: str) -> str:
    parts = [part for part in key.lower().replace("\\", "/").split("/") if "." in part and "json" not in part]
    return parts[-1] if parts else ""


def _group_counts(rows: list[dict[str, Any]], label_field: str, fallback_field: str) -> list[dict[str, Any]]:
    grouped: dict[str, dict[str, Any]] = {}
    for row in rows:
        label = str(row.get(label_field) or row.get(fallback_field) or "unknown")
        item = grouped.setdefault(label, {"label": label, "total": 0, "study_safe": 0, "non_study": 0, "last_observed_at": ""})
        item["total"] += 1
        item["study_safe"] += 1 if row.get("study_safe") else 0
        item["non_study"] += 0 if row.get("study_safe") else 1
        item["last_observed_at"] = max(item["last_observed_at"], str(row.get("observed_at") or ""))
    return sorted(grouped.values(), key=lambda item: item["total"], reverse=True)[:20]


def _index_html() -> str:
    local_backend = html.escape(DEFAULT_LOCAL_BACKEND, quote=True)
    return f"""<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>TraceDeck Cloud Admin</title>
  <style>
    :root {{
      color-scheme: light;
      --bg: #f4f7fb;
      --bg-band: #e7edf4;
      --surface: #ffffff;
      --surface-2: #f8fafc;
      --surface-3: #eef3f7;
      --ink: #17212b;
      --muted: #647386;
      --line: #d8e1ea;
      --line-strong: #a7b7c6;
      --teal: #087f7b;
      --teal-strong: #055f5b;
      --green: #17764c;
      --amber: #ad6412;
      --red: #b84438;
      --blue: #285fba;
      --accent-soft: #e5f6f3;
      --success-surface: #e8f7ed;
      --warning-surface: #fff4df;
      --danger-surface: #fff0ed;
      --info-surface: #eaf3ff;
      --shadow-soft: 0 14px 34px rgba(23, 33, 43, 0.08);
      --shadow-line: 0 1px 0 rgba(167, 183, 198, 0.28);
    }}
    body.theme-dark {{
      color-scheme: dark;
      --bg: #0f1214;
      --bg-band: #171c20;
      --surface: #181d21;
      --surface-2: #20272c;
      --surface-3: #2a3339;
      --ink: #f4f7f6;
      --muted: #aab6bb;
      --line: #334047;
      --line-strong: #60707a;
      --teal: #5ad1c8;
      --teal-strong: #9ce7df;
      --green: #7bd998;
      --amber: #efbf72;
      --red: #f08a82;
      --blue: #8ebcff;
      --accent-soft: #163733;
      --success-surface: #173324;
      --warning-surface: #382a15;
      --danger-surface: #3a2020;
      --info-surface: #1a2a3d;
      --shadow-soft: 0 16px 38px rgba(0, 0, 0, 0.30);
      --shadow-line: 0 1px 0 rgba(96, 112, 122, 0.30);
    }}
    * {{ box-sizing: border-box; }}
    html {{ max-width: 100%; overflow-x: clip; background: var(--bg); }}
    body {{ margin: 0; max-width: 100%; overflow-x: clip; background: linear-gradient(180deg, var(--bg-band) 0, var(--bg) 340px, var(--bg) 100%); color: var(--ink); font: 14px/1.45 Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif; letter-spacing: 0; }}
    h1, h2, h3, p {{ margin: 0; }}
    h1 {{ font-size: 24px; font-weight: 840; line-height: 1.1; }}
    h2 {{ font-size: 20px; font-weight: 820; line-height: 1.2; }}
    h3 {{ font-size: 14px; font-weight: 790; line-height: 1.2; }}
    button, select, input {{ min-height: 42px; border: 1px solid color-mix(in srgb, var(--line) 88%, transparent); border-radius: 8px; background: color-mix(in srgb, var(--surface) 96%, var(--surface-2)); color: var(--ink); font: inherit; }}
    button {{ display: inline-flex; align-items: center; justify-content: center; gap: 8px; padding: 0 14px; cursor: pointer; font-weight: 780; white-space: nowrap; }}
    button:hover {{ border-color: var(--line-strong); background: var(--surface); }}
    button.primary {{ border-color: #12312f; background: #12312f; color: #fff; box-shadow: 0 10px 22px rgba(8, 127, 123, 0.16); }}
    body.theme-dark button.primary {{ border-color: color-mix(in srgb, var(--teal-strong) 78%, #fff); background: color-mix(in srgb, var(--teal-strong) 88%, #fff); color: #0f1718; box-shadow: none; }}
    input, select {{ width: 100%; padding: 0 12px; }}
    .muted {{ color: var(--muted); }}
    .eyebrow {{ color: var(--teal-strong); font-size: 12px; font-weight: 820; text-transform: uppercase; }}
    .topbar {{ position: sticky; top: 0; z-index: 12; display: grid; grid-template-columns: minmax(340px, 1fr) minmax(0, auto); gap: 18px; align-items: center; min-height: 84px; padding: 16px 30px; border-bottom: 1px solid color-mix(in srgb, var(--line) 82%, transparent); background: color-mix(in srgb, var(--surface) 94%, transparent); box-shadow: var(--shadow-line), 0 16px 36px rgba(23, 33, 43, 0.06); backdrop-filter: blur(14px); }}
    body.theme-dark .topbar {{ box-shadow: var(--shadow-line), 0 18px 42px rgba(0, 0, 0, 0.24); }}
    .brand-lockup {{ display: flex; align-items: center; gap: 14px; min-width: 0; }}
    .brand-lockup p.muted {{ margin-top: 4px; max-width: 760px; }}
    .brand-mark {{ display: grid; width: 46px; height: 46px; flex: 0 0 46px; grid-template-columns: repeat(3, 1fr); gap: 4px; padding: 9px; border: 1px solid color-mix(in srgb, var(--teal) 42%, var(--line)); border-radius: 8px; background: color-mix(in srgb, var(--surface) 90%, var(--accent-soft)); box-shadow: inset 0 1px 0 color-mix(in srgb, #fff 40%, transparent); }}
    .brand-mark span {{ border-radius: 3px; background: var(--teal); }}
    .brand-mark span:nth-child(2) {{ background: var(--blue); transform: translateY(5px); }}
    .brand-mark span:nth-child(3) {{ background: var(--green); transform: translateY(10px); }}
    body.theme-dark .brand-mark {{ background: color-mix(in srgb, var(--surface-2) 88%, var(--accent-soft)); box-shadow: none; }}
    .toolbar, .status-row, .meta-row, .hero-actions {{ display: flex; align-items: center; gap: 10px; flex-wrap: wrap; min-width: 0; }}
    .toolbar {{ justify-content: flex-end; }}
    .status {{ display: inline-flex; align-items: center; gap: 8px; min-height: 42px; padding: 0 12px; border: 1px solid #b8c7da; border-radius: 8px; background: var(--surface-2); color: #475467; font-weight: 780; white-space: nowrap; }}
    .status.checking {{ border-color: #b8c7da; background: var(--surface-2); color: #475467; }}
    .status.connected {{ border-color: #abefc6; background: var(--success-surface); color: var(--green); }}
    .status.disconnected {{ border-color: #fecdca; background: var(--danger-surface); color: var(--red); }}
    body.theme-dark .status {{ border-color: #4b5563; background: var(--surface-2); color: #c7d2de; }}
    body.theme-dark .status.connected {{ border-color: rgba(100, 197, 139, 0.45); background: rgba(30, 93, 64, 0.35); color: var(--green); }}
    body.theme-dark .status.disconnected {{ border-color: rgba(244, 118, 109, 0.5); background: rgba(92, 35, 36, 0.38); color: var(--red); }}
    .status-light {{ width: 9px; height: 9px; border-radius: 999px; background: currentColor; box-shadow: 0 0 0 4px rgba(2, 122, 72, 0.13); }}
    .app-shell {{ width: calc(100% - 32px); max-width: 1460px; margin: 22px auto 34px; display: grid; grid-template-columns: 286px minmax(0, 1fr); gap: 18px; min-width: 0; }}
    .side-rail {{ position: sticky; top: 106px; align-self: start; display: grid; gap: 14px; min-width: 0; }}
    .source-card, .panel, .metric, .hero-panel {{ min-width: 0; border: 1px solid color-mix(in srgb, var(--line) 90%, transparent); border-radius: 8px; background: color-mix(in srgb, var(--surface) 98%, var(--surface-2)); box-shadow: var(--shadow-soft); }}
    .source-card, .panel, .hero-panel {{ padding: 18px; overflow: hidden; }}
    .rail-title {{ display: grid; gap: 4px; margin-bottom: 14px; }}
    .rail-title strong {{ font-size: 15px; }}
    .source-grid {{ display: grid; gap: 12px; }}
    .field {{ display: grid; gap: 6px; }}
    .field label {{ color: var(--muted); font-size: 12.5px; font-weight: 780; }}
    .tabs {{ display: grid; gap: 8px; padding: 8px; border: 1px solid color-mix(in srgb, var(--line) 86%, transparent); border-radius: 8px; background: color-mix(in srgb, var(--surface) 95%, transparent); box-shadow: var(--shadow-line); }}
    .tab {{ justify-content: flex-start; width: 100%; min-height: 46px; border-color: transparent; background: transparent; color: var(--muted); box-shadow: none; }}
    .tab:hover {{ border-color: color-mix(in srgb, var(--line-strong) 60%, transparent); background: var(--surface-2); color: var(--ink); }}
    .tab.is-active {{ border-color: color-mix(in srgb, var(--teal) 48%, var(--line)); background: color-mix(in srgb, var(--teal) 12%, var(--surface)); color: var(--teal-strong); box-shadow: inset 3px 0 0 var(--teal); }}
    .content-stack {{ display: grid; gap: 16px; min-width: 0; }}
    .hero-panel {{ display: grid; grid-template-columns: minmax(0, 1fr) minmax(240px, auto); gap: 16px; align-items: center; background: linear-gradient(135deg, color-mix(in srgb, var(--surface) 96%, var(--accent-soft)), var(--surface)); }}
    body.theme-dark .hero-panel {{ background: linear-gradient(135deg, color-mix(in srgb, var(--surface) 92%, var(--accent-soft)), var(--surface)); }}
    .hero-panel h2 {{ max-width: 720px; margin-top: 4px; }}
    .hero-panel p.muted {{ margin-top: 6px; max-width: 760px; }}
    .page {{ display: grid; gap: 14px; min-width: 0; }}
    .page[hidden] {{ display: none !important; }}
    .grid {{ display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }}
    .two {{ display: grid; grid-template-columns: minmax(0, 0.92fr) minmax(0, 1.08fr); gap: 12px; }}
    .metric {{ min-height: 136px; padding: 15px; display: grid; align-content: space-between; gap: 8px; border-top: 3px solid color-mix(in srgb, var(--teal) 72%, var(--line)); background: var(--surface); }}
    .metric strong {{ color: var(--ink); font-size: 30px; font-weight: 840; line-height: 1.05; overflow-wrap: anywhere; }}
    .metric-label {{ color: var(--muted); font-size: 12.5px; font-weight: 780; }}
    .section-header {{ display: flex; align-items: flex-start; justify-content: space-between; gap: 12px; margin-bottom: 12px; }}
    .section-header p {{ margin-top: 4px; }}
    .list {{ display: grid; gap: 8px; }}
    .item {{ display: grid; gap: 7px; padding: 12px; border: 1px solid color-mix(in srgb, var(--line) 84%, transparent); border-radius: 8px; background: color-mix(in srgb, var(--surface-2) 90%, var(--surface)); overflow-wrap: anywhere; }}
    .settings-grid {{ display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }}
    .settings-grid .item:nth-child(2) {{ grid-column: span 2; }}
    .pill {{ display: inline-flex; align-items: center; min-height: 28px; padding: 4px 10px; border: 1px solid var(--line); border-radius: 8px; background: color-mix(in srgb, var(--surface-2) 86%, var(--surface)); color: var(--muted); font-size: 12.5px; font-weight: 760; line-height: 1.18; }}
    .strong-pill {{ border-color: color-mix(in srgb, var(--teal) 45%, var(--line)); background: var(--accent-soft); color: var(--teal-strong); }}
    table {{ width: 100%; min-width: 980px; border-collapse: collapse; }}
    th, td {{ padding: 12px; border-bottom: 1px solid var(--line); text-align: left; vertical-align: top; overflow-wrap: anywhere; }}
    th {{ color: var(--muted); font-size: 12.5px; font-weight: 790; background: color-mix(in srgb, var(--surface-2) 92%, var(--surface)); }}
    .table-wrap {{ max-width: 100%; overflow-x: auto; border: 1px solid color-mix(in srgb, var(--line) 90%, transparent); border-radius: 8px; background: var(--surface); box-shadow: var(--shadow-soft); }}
    .empty {{ min-height: 90px; display: grid; place-items: center; padding: 14px; border: 1px dashed var(--line); border-radius: 8px; background: var(--surface-2); color: var(--muted); text-align: center; }}
    @media (max-width: 980px) {{ .topbar, .app-shell, .hero-panel, .two {{ grid-template-columns: 1fr; }} .topbar {{ position: static; padding: 14px 16px; }} .toolbar {{ justify-content: stretch; }} .toolbar > * {{ flex: 1 1 180px; min-width: 0; }} .side-rail {{ position: static; }} .tabs {{ grid-template-columns: repeat(2, minmax(0, 1fr)); }} .grid {{ grid-template-columns: repeat(2, minmax(0, 1fr)); }} }}
    @media (max-width: 560px) {{ .app-shell {{ width: auto; margin: 14px 10px 28px; }} .grid, .settings-grid, .tabs {{ grid-template-columns: 1fr; }} .settings-grid .item:nth-child(2) {{ grid-column: auto; }} .toolbar > * {{ flex-basis: 100%; }} }}
  </style>
</head>
<body>
  <header class="topbar">
    <div class="brand-lockup">
      <span class="brand-mark" aria-hidden="true"><span></span><span></span><span></span></span>
      <div>
        <p class="eyebrow">Endpoint Risk Observability</p>
        <h1>TraceDeck Cloud Admin</h1>
        <p class="muted">Archive-backed activity review, host posture, cache efficiency, and localhost fallback in one admin console.</p>
      </div>
    </div>
    <div class="toolbar">
      <button id="theme-button" type="button" aria-label="Toggle cloud admin theme"><span id="theme-label">Theme: Light</span></button>
      <button id="refresh-button" class="primary" type="button">Refresh</button>
      <span id="server-status" class="status"><span class="status-light" aria-hidden="true"></span><span id="server-status-text">Checking</span></span>
    </div>
  </header>
  <main class="app-shell">
    <aside class="side-rail">
      <section class="source-card">
        <div class="rail-title">
          <strong>Workspace Source</strong>
          <span class="muted">Choose archive or localhost data.</span>
        </div>
        <div class="source-grid">
          <div class="field">
            <label for="source-select">Source</label>
            <select id="source-select">
              <option value="s3">Lambda S3 Archive</option>
              <option value="local">Localhost 18080</option>
            </select>
          </div>
          <div class="field">
            <label for="tenant-input">Tenant</label>
            <input id="tenant-input" value="family-varadha" autocomplete="off">
          </div>
          <div class="field">
            <label for="local-input">Local Backend</label>
            <input id="local-input" value="{local_backend}" autocomplete="off">
          </div>
          <button id="force-refresh-button" type="button">Bypass Cache</button>
        </div>
      </section>
      <nav class="tabs" aria-label="Cloud admin pages">
        <button class="tab is-active" type="button" data-page="overview">Overview</button>
        <button class="tab" type="button" data-page="browser">Browser Activity</button>
        <button class="tab" type="button" data-page="archive">S3 Archive</button>
        <button class="tab" type="button" data-page="settings">Source &amp; Cache</button>
      </nav>
    </aside>
    <div class="content-stack">
      <section class="hero-panel">
        <div>
          <p class="eyebrow">Live Archive Console</p>
          <h2>Activity, cache, and host signal from the selected source</h2>
          <p class="muted">Designed for quick admin review without exposing passwords, screenshots, raw URLs, page titles, tokens, cookies, or private content.</p>
        </div>
        <div class="hero-actions">
          <span id="cache-state" class="pill strong-pill">cache waiting</span>
          <span id="last-refresh-pill" class="pill">waiting for refresh</span>
        </div>
      </section>
      <section id="overview-page" class="page">
        <div class="grid">
          <div class="metric"><span class="metric-label">Archive Objects</span><strong id="metric-objects">-</strong><span id="metric-objects-sub" class="muted">waiting</span></div>
          <div class="metric"><span class="metric-label">Sampled Rows</span><strong id="metric-rows">-</strong><span id="metric-rows-sub" class="muted">waiting</span></div>
          <div class="metric"><span class="metric-label">Cache Hit Rate</span><strong id="metric-hit">-</strong><span id="metric-hit-sub" class="muted">waiting</span></div>
          <div class="metric"><span class="metric-label">Cache Miss Rate</span><strong id="metric-miss">-</strong><span id="metric-miss-sub" class="muted">waiting</span></div>
        </div>
        <div class="two">
          <div class="panel">
            <div class="section-header">
              <div><h2>Hosts</h2><p class="muted">Reporting hosts from the current source.</p></div>
              <span id="host-count" class="pill">0 hosts</span>
            </div>
            <div id="host-list" class="list"><div class="empty">No host rows loaded.</div></div>
          </div>
          <div class="panel">
            <div class="section-header">
              <div><h2>Browsers</h2><p class="muted">Chrome, Edge, Brave, and other browser summaries.</p></div>
              <span id="browser-count" class="pill">0 browsers</span>
            </div>
            <div id="browser-list" class="list"><div class="empty">No browser rows loaded.</div></div>
          </div>
        </div>
      </section>
      <section id="browser-page" class="page" hidden>
        <div class="panel">
          <div class="section-header">
            <div><h2>Browser Domain Activity</h2><p class="muted">Metadata-only domain, category, study status, and source proof.</p></div>
            <span id="row-count" class="pill">0 rows</span>
          </div>
          <div class="table-wrap">
            <table><thead><tr><th>Host</th><th>Browser</th><th>Domain</th><th>Category</th><th>Study</th><th>Source</th><th>Observed</th></tr></thead><tbody id="browser-table"><tr><td colspan="7"><div class="empty">No browser data loaded.</div></td></tr></tbody></table>
          </div>
        </div>
      </section>
      <section id="archive-page" class="page" hidden>
        <div class="panel">
          <div class="section-header">
            <div><h2>S3 Archive Objects</h2><p class="muted">Recent archive objects sampled for admin rendering.</p></div>
            <span id="object-count" class="pill">0 objects</span>
          </div>
          <div class="table-wrap">
            <table><thead><tr><th>Key</th><th>Size</th><th>Storage</th><th>Modified</th></tr></thead><tbody id="object-table"><tr><td colspan="4"><div class="empty">No S3 objects loaded.</div></td></tr></tbody></table>
          </div>
        </div>
      </section>
      <section id="settings-page" class="page" hidden>
        <div class="panel">
          <div class="section-header">
            <div><h2>Source &amp; Cache</h2><p class="muted">Current bucket, prefix, refresh time, privacy boundary, and localhost target.</p></div>
          </div>
          <div class="settings-grid">
            <div class="item"><span class="muted">Bucket</span><strong id="bucket-label">Bucket pending</strong><span id="prefix-label" class="muted">Prefix pending</span></div>
            <div class="item"><span class="muted">Last Generated</span><strong id="generated-label">Generated pending</strong><span id="privacy-label" class="muted">Metadata-only rendering</span></div>
            <div class="item"><span class="muted">Local Backend</span><strong id="local-label">Local backend pending</strong><span class="muted">Local mode uses your browser to contact the selected localhost URL.</span></div>
          </div>
        </div>
      </section>
    </div>
  </main>
  <script>
    const storage = {{ theme: "tracedeck.ui.theme", source: "tracedeck.cloud.source", page: "tracedeck.cloud.page" }};
    let lastPayload = null;

    function text(value) {{ return value === 0 ? "0" : (value || "-"); }}
    function setText(id, value) {{ const element = document.getElementById(id); if (element) element.textContent = text(value); }}
    function escapeHTML(value) {{ return String(text(value)).replace(/[&<>"']/g, c => ({{ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }}[c])); }}
    function formatBytes(value) {{
      const size = Number(value || 0);
      if (size < 1024) return size + " B";
      if (size < 1048576) return (size / 1024).toFixed(1) + " KB";
      if (size < 1073741824) return (size / 1048576).toFixed(1) + " MB";
      return (size / 1073741824).toFixed(1) + " GB";
    }}
    function setStatus(kind, label) {{
      const status = document.getElementById("server-status");
      status.classList.remove("checking", "connected", "disconnected");
      status.classList.add(kind || "checking");
      setText("server-status-text", label);
    }}
    function sourceBadge(row) {{
      const source = String(row.source_kind || "source_pending").replace(/_/g, " ");
      const scope = String(row.evidence_scope || "scope pending").replace(/_/g, " ");
      return '<span class="pill">' + escapeHTML(source) + '</span><br><span class="muted">' + escapeHTML(scope) + '</span>';
    }}
    function setTheme(theme) {{
      const next = theme === "dark" ? "dark" : "light";
      document.body.classList.toggle("theme-dark", next === "dark");
      localStorage.setItem(storage.theme, next);
      setText("theme-label", next === "dark" ? "Theme: Dark" : "Theme: Light");
    }}
    function setPage(page) {{
      const next = page || "overview";
      document.querySelectorAll(".page").forEach(el => el.hidden = el.id !== next + "-page");
      document.querySelectorAll("[data-page]").forEach(button => {{
        const active = button.dataset.page === next;
        button.classList.toggle("is-active", active);
        if (active) button.setAttribute("aria-current", "page"); else button.removeAttribute("aria-current");
      }});
      localStorage.setItem(storage.page, next);
    }}
    async function load(force) {{
      try {{
        setStatus("", "Checking");
        const source = document.getElementById("source-select").value;
        localStorage.setItem(storage.source, source);
        if (source === "local") await loadLocal();
        else await loadS3(force);
        setStatus("connected", "Connected");
      }} catch (error) {{
        setStatus("disconnected", "Not connected");
        setText("privacy-label", error.message || "Load failed");
      }}
    }}
    async function loadS3(force) {{
      const response = await fetch("/api/s3-summary" + (force ? "?refresh=true" : ""));
      if (!response.ok) throw new Error("S3 summary failed with " + response.status);
      const payload = await response.json();
      lastPayload = payload;
      renderS3(payload);
    }}
    async function loadLocal() {{
      const base = document.getElementById("local-input").value.replace(/\\/$/, "");
      const tenant = encodeURIComponent(document.getElementById("tenant-input").value || "family-varadha");
      const health = await fetch(base + "/health");
      if (!health.ok) throw new Error("Local health failed with " + health.status);
      const response = await fetch(base + "/api/v1/tenants/" + tenant + "/browser-activity?limit=50");
      if (!response.ok) throw new Error("Local browser activity failed with " + response.status);
      const viewer = await response.json();
      const payload = {{
        status: "local",
        bucket: "localhost",
        prefix: base,
        summary: {{ objects: 0, bytes: 0, sampled_rows: viewer.summary?.total || 0, study_safe: viewer.summary?.study_safe || 0, non_study_youtube: viewer.summary?.non_study_youtube || 0, latest_object_at: viewer.summary?.last_observed_at || "" }},
        cache: {{ hits: 0, misses: 0, hit_percent: 0, miss_percent: 0, hit: false, ttl_seconds: 0 }},
        browser_rows: viewer.items || [],
        hosts: (viewer.hosts || []).map(item => ({{ label: item.host_name || item.device_id, total: item.total, study_safe: item.study_safe, non_study: item.non_study, last_observed_at: item.last_observed_at }})),
        browsers: (viewer.browsers || []).map(item => ({{ label: item.name, total: item.total, study_safe: item.study_safe, non_study: item.non_study_youtube, last_observed_at: item.last_observed_at }})),
        objects: [],
        generated_at: new Date().toISOString(),
        privacy_boundary: viewer.privacy_boundary || "metadata-only local browser activity"
      }};
      lastPayload = payload;
      renderS3(payload);
    }}
    function renderS3(payload) {{
      const summary = payload.summary || {{}};
      const cache = payload.cache || {{}};
      setText("metric-objects", summary.objects || 0);
      setText("metric-objects-sub", formatBytes(summary.bytes || 0) + " archived");
      setText("metric-rows", summary.sampled_rows || 0);
      setText("metric-rows-sub", (summary.study_safe || 0) + " study-safe, " + (summary.non_study_youtube || 0) + " YouTube review");
      setText("metric-hit", (cache.hit_percent ?? 0) + "%");
      setText("metric-hit-sub", (cache.hits || 0) + " hits, TTL " + (cache.ttl_seconds || 0) + "s");
      setText("metric-miss", (cache.miss_percent ?? 0) + "%");
      setText("metric-miss-sub", (cache.misses || 0) + " misses");
      setText("cache-state", cache.hit ? "cache hit" : "cache miss");
      setText("last-refresh-pill", summary.latest_object_at ? "latest object " + summary.latest_object_at : "waiting for archive");
      setText("bucket-label", payload.bucket || "Bucket pending");
      setText("prefix-label", payload.prefix ? "Prefix " + payload.prefix : "No prefix configured");
      setText("generated-label", payload.generated_at || "Generated pending");
      setText("privacy-label", payload.privacy_boundary || "Metadata-only rendering");
      setText("local-label", document.getElementById("local-input").value || "Local backend pending");
      renderGroups("host-list", "host-count", payload.hosts || [], "hosts");
      renderGroups("browser-list", "browser-count", payload.browsers || [], "browsers");
      renderRows(payload.browser_rows || []);
      renderObjects(payload.objects || []);
    }}
    function renderGroups(targetID, countID, rows, label) {{
      setText(countID, rows.length + " " + label);
      const target = document.getElementById(targetID);
      if (!rows.length) {{ target.innerHTML = '<div class="empty">No rows loaded.</div>'; return; }}
      target.innerHTML = rows.map(item => '<div class="item"><strong>' + escapeHTML(item.label) + '</strong><div class="meta-row"><span class="pill">' + escapeHTML(item.total || 0) + ' total</span><span class="pill">' + escapeHTML(item.study_safe || 0) + ' safe</span><span class="pill">' + escapeHTML(item.non_study || 0) + ' review</span></div><span class="muted">' + escapeHTML(item.last_observed_at || "not observed") + '</span></div>').join("");
    }}
    function renderRows(rows) {{
      setText("row-count", rows.length + " rows");
      const target = document.getElementById("browser-table");
      if (!rows.length) {{ target.innerHTML = '<tr><td colspan="7"><div class="empty">No browser data loaded.</div></td></tr>'; return; }}
      target.innerHTML = rows.map(row => '<tr><td>' + escapeHTML(row.host_name || row.device_id) + '</td><td>' + escapeHTML(row.browser) + '</td><td><strong>' + escapeHTML(row.domain) + '</strong></td><td>' + escapeHTML(row.category) + '</td><td>' + escapeHTML(row.study_safe ? "study-safe" : "review") + '</td><td>' + sourceBadge(row) + '<br><span class="muted">' + escapeHTML(row.evidence_detail || "") + '</span></td><td>' + escapeHTML(row.observed_at || "not observed") + '</td></tr>').join("");
    }}
    function renderObjects(objects) {{
      setText("object-count", objects.length + " objects");
      const target = document.getElementById("object-table");
      if (!objects.length) {{ target.innerHTML = '<tr><td colspan="4"><div class="empty">No S3 objects loaded.</div></td></tr>'; return; }}
      target.innerHTML = objects.map(item => '<tr><td>' + escapeHTML(item.key) + '</td><td>' + escapeHTML(formatBytes(item.size)) + '</td><td>' + escapeHTML(item.storage_class || "STANDARD") + '</td><td>' + escapeHTML(item.last_modified || "") + '</td></tr>').join("");
    }}
    document.getElementById("theme-button").addEventListener("click", () => setTheme(document.body.classList.contains("theme-dark") ? "light" : "dark"));
    document.getElementById("refresh-button").addEventListener("click", () => load(false));
    document.getElementById("force-refresh-button").addEventListener("click", () => load(true));
    document.getElementById("source-select").addEventListener("change", () => load(false));
    document.querySelectorAll("[data-page]").forEach(button => button.addEventListener("click", () => setPage(button.dataset.page)));
    setTheme(localStorage.getItem(storage.theme) || (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"));
    document.getElementById("source-select").value = localStorage.getItem(storage.source) || "s3";
    setPage(localStorage.getItem(storage.page) || "overview");
    load(false);
  </script>
</body>
</html>"""
