# TraceDeck Browser Metadata Bridge

This folder contains the TraceDeck Chromium extension skeleton for Chrome,
Microsoft Edge, and Brave.

The extension is a local-only bridge. It observes top-frame navigation events,
normalizes them to domain/category metadata, and posts `browser.domain.observed`
events to the existing local backend telemetry route:

```text
http://127.0.0.1:18080/api/v1/devices/{device_id}/telemetry-events
```

## Privacy Boundary

The extension does not store or transmit raw URLs, page titles, cookies,
tokens, passwords, screenshots, page content, form fields, private messages, or
provider secrets. Raw navigation URLs are used only transiently inside the
service worker so the domain and study/category labels can be derived.

The posted event metadata is limited to:

- browser name
- domain
- category
- URL mode and stored URL mode, both `domain_only`
- source/evidence labels
- visit count
- study-safe YouTube boolean

## Local Install

1. Open `chrome://extensions`, `edge://extensions`, or
   `brave://extensions`.
2. Enable developer mode.
3. Load this folder as an unpacked extension.
4. Open the extension options and confirm the local backend, tenant, device,
   host, profile, and OS labels.

The backend must be reachable on localhost. Remote backend origins are rejected
by the extension privacy guard.

## Verification

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-browser-extension-skeleton.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/smoke-phase108.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/newman-phase108.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/verify-phase108.ps1
```
