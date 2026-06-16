# Browser Extension

Phase 108 adds the TraceDeck Browser Metadata Bridge for Chrome, Microsoft Edge,
and Brave.

The extension is a Manifest V3 Chromium extension under `browser-extension/`.
It observes top-frame browser navigation events, immediately reduces them to
domain/category metadata, and posts the existing TraceDeck telemetry event shape
to the local backend:

```text
POST /api/v1/devices/{device_id}/telemetry-events
```

## Privacy Boundary

The extension does not store or transmit raw URLs, page titles, cookies,
tokens, passwords, screenshots, page content, form fields, private messages,
provider secrets, alert bodies, payment data, or raw provider payloads.

Raw navigation URLs are used transiently inside the service worker only to
derive the normalized domain and a study/category label. The payload sent to the
backend contains `domain`, `category`, `browser_name`, `source_kind`,
`evidence_scope`, `evidence_detail`, `url_mode=domain_only`,
`stored_url_mode=domain_only`, `visit_count`, and `youtube_study_match`.

The backend origin must be localhost or `127.0.0.1`; remote origins are rejected
by the extension privacy core.

## Supported Browsers

- Google Chrome
- Microsoft Edge
- Brave

All three use the same Chromium extension package.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-browser-extension-skeleton.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase108.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase108.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase108.ps1
python ./devctl.py test phase108
```

The smoke and Newman checks start an isolated local backend, enroll a dedicated
extension device, ingest an extension-shaped `browser.domain.observed` event,
and verify the Browser Activity API exposes only domain/category metadata.
