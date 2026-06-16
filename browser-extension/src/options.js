/* global chrome, TraceDeckBrowserPrivacy */
"use strict";

const fields = ["backend_origin", "tenant_id", "device_id", "host_name", "profile", "os_name"];
const statusNode = document.getElementById("status");

async function loadSettings() {
  const stored = await chrome.storage.local.get(Object.keys(TraceDeckBrowserPrivacy.defaultConfig));
  const config = TraceDeckBrowserPrivacy.cleanConfig(stored);
  fields.forEach((field) => {
    document.getElementById(field).value = config[field] || "";
  });
}

async function saveSettings(event) {
  event.preventDefault();
  const next = {};
  fields.forEach((field) => {
    next[field] = document.getElementById(field).value;
  });
  const config = TraceDeckBrowserPrivacy.cleanConfig(next);
  if (!TraceDeckBrowserPrivacy.isAllowedBackendOrigin(config.backend_origin)) {
    statusNode.textContent = "Backend must be localhost or 127.0.0.1.";
    return;
  }
  await chrome.storage.local.set(config);
  statusNode.textContent = "Settings saved.";
}

document.getElementById("settings-form").addEventListener("submit", saveSettings);
loadSettings().catch(() => {
  statusNode.textContent = "Settings could not be loaded.";
});
