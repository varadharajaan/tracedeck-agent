"use strict";

const assert = require("assert");
const fs = require("fs");
const path = require("path");
const privacy = require("../src/privacy-core.js");

const repoRoot = path.resolve(__dirname, "..", "..");
const manifest = JSON.parse(fs.readFileSync(path.join(repoRoot, "browser-extension", "manifest.json"), "utf8"));

assert.strictEqual(manifest.manifest_version, 3);
assert.strictEqual(manifest.background.service_worker, "src/background.js");
assert(manifest.permissions.includes("storage"));
assert(manifest.permissions.includes("webNavigation"));
["tabs", "history", "cookies", "desktopCapture", "tabCapture", "scripting", "downloads", "bookmarks"].forEach((permission) => {
  assert(!manifest.permissions.includes(permission), `forbidden permission present: ${permission}`);
});
assert(manifest.host_permissions.includes("<all_urls>"));
assert(manifest.host_permissions.includes("http://127.0.0.1/*"));
assert(manifest.host_permissions.includes("http://localhost/*"));

const config = privacy.cleanConfig({
  backend_origin: "http://127.0.0.1:18080",
  tenant_id: "family-varadha",
  device_id: "phase108-extension-device",
  host_name: "phase108-extension-host",
  profile: "ai_btech_student",
  os_name: "browser_extension",
  youtube_study_keywords: ["python", "system design"]
});

assert.strictEqual(privacy.telemetryEndpoint(config), "http://127.0.0.1:18080/api/v1/devices/phase108-extension-device/telemetry-events");
assert.strictEqual(privacy.isAllowedBackendOrigin("https://example.com"), false);
assert.strictEqual(privacy.isAllowedBackendOrigin("http://192.168.1.10:18080"), false);
assert.strictEqual(privacy.isAllowedBackendOrigin("http://localhost:18080"), true);
assert.strictEqual(privacy.normalizeDomain("https://www.YouTube.com/results?search_query=python+tutorial"), "youtube.com");

const event = privacy.buildTelemetryEvent({
  rawNavigationUrl: "https://www.youtube.com/results?search_query=python+tutorial&private=value",
  browserName: "chrome",
  config,
  observedAt: "2026-06-16T09:30:00Z",
  eventID: "phase108-extension-domain-1"
});
const body = privacy.buildIngestRequest(event, config);
const serialized = JSON.stringify(body).toLowerCase();

assert.strictEqual(event.type, "browser.domain.observed");
assert.strictEqual(event.source, "collector.browser.extension");
assert.strictEqual(event.metadata.browser_name, "chrome");
assert.strictEqual(event.metadata.domain, "youtube.com");
assert.strictEqual(event.metadata.category, "study");
assert.strictEqual(event.metadata.url_mode, "domain_only");
assert.strictEqual(event.metadata.stored_url_mode, "domain_only");
assert.strictEqual(event.metadata.youtube_study_match, "true");
assert.strictEqual(privacy.hasForbiddenPayloadKeys(body), false);
["https://", "search_query", "private=value", "raw_url", "page_title", "cookie", "password", "screenshot"].forEach((marker) => {
  assert(!serialized.includes(marker), `payload leaked forbidden marker: ${marker}`);
});

const background = fs.readFileSync(path.join(repoRoot, "browser-extension", "src", "background.js"), "utf8");
assert(background.includes("importScripts(\"privacy-core.js\")"));
assert(background.includes("hasForbiddenPayloadKeys"));

console.log("TraceDeck browser extension privacy contract passed");
