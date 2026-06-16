/* global chrome, TraceDeckBrowserPrivacy, importScripts */
"use strict";

importScripts("privacy-core.js");

const privacy = TraceDeckBrowserPrivacy;

async function readConfig() {
  const stored = await chrome.storage.local.get(Object.keys(privacy.defaultConfig));
  return privacy.cleanConfig(stored);
}

async function detectBrowser() {
  let braveHint = false;
  try {
    braveHint = Boolean(navigator.brave && await navigator.brave.isBrave());
  } catch (error) {
    braveHint = false;
  }
  return privacy.detectBrowserName(navigator.userAgent, braveHint);
}

async function updateBadge(enabled) {
  if (!chrome.action) {
    return;
  }
  await chrome.action.setBadgeText({ text: enabled ? "ON" : "OFF" });
  await chrome.action.setBadgeBackgroundColor({ color: enabled ? "#15803d" : "#b91c1c" });
}

async function postDomainObservation(rawNavigationUrl) {
  const config = await readConfig();
  const endpoint = privacy.telemetryEndpoint(config);
  if (!endpoint) {
    await updateBadge(false);
    return;
  }

  const browserName = await detectBrowser();
  const event = privacy.buildTelemetryEvent({
    rawNavigationUrl,
    browserName,
    config
  });
  const body = privacy.buildIngestRequest(event, config);
  if (!body || privacy.hasForbiddenPayloadKeys(body)) {
    await updateBadge(false);
    return;
  }

  const response = await fetch(endpoint, {
    method: "POST",
    headers: {
      "Content-Type": "application/json"
    },
    body: JSON.stringify(body)
  });
  await updateBadge(response.ok);
}

chrome.runtime.onInstalled.addListener(async () => {
  const current = await chrome.storage.local.get(Object.keys(privacy.defaultConfig));
  await chrome.storage.local.set(privacy.cleanConfig(current));
  await updateBadge(true);
});

chrome.webNavigation.onCompleted.addListener((details) => {
  if (!details || details.frameId !== 0) {
    return;
  }
  postDomainObservation(details.url).catch(() => updateBadge(false));
});
