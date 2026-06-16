(function attachTraceDeckBrowserPrivacy(root) {
  "use strict";

  const constants = Object.freeze({
    defaultBackendOrigin: "http://127.0.0.1:18080",
    defaultTenantID: "family-varadha",
    defaultDeviceID: "browser-extension-device",
    defaultHostName: "browser-extension-host",
    defaultProfile: "ai_btech_student",
    defaultOSName: "browser_extension",
    eventTypeBrowserObserved: "browser.domain.observed",
    eventSourceBrowserExtension: "collector.browser.extension",
    sourceKindLiveIngested: "live_ingested",
    evidenceScopeLive: "live",
    evidenceDetail: "browser extension observed domain-only navigation metadata",
    urlModeDomainOnly: "domain_only",
    categoryStudy: "study",
    categoryVideoStreaming: "video_streaming",
    categorySocialMedia: "social_media",
    categoryGaming: "gaming",
    categoryShopping: "shopping",
    categoryUnknown: "unknown"
  });

  const defaultConfig = Object.freeze({
    backend_origin: constants.defaultBackendOrigin,
    tenant_id: constants.defaultTenantID,
    device_id: constants.defaultDeviceID,
    host_name: constants.defaultHostName,
    profile: constants.defaultProfile,
    os_name: constants.defaultOSName,
    youtube_study_keywords: ["python", "math", "system design", "course", "lecture", "tutorial"]
  });

  const forbiddenPayloadKeys = Object.freeze([
    "url",
    "raw_url",
    "title",
    "page_title",
    "cookie",
    "cookies",
    "token",
    "password",
    "screenshot",
    "page_content",
    "form_field",
    "private_message",
    "provider_secret",
    "alert_body"
  ]);

  function cleanConfig(input) {
    const config = Object.assign({}, defaultConfig, input || {});
    config.backend_origin = String(config.backend_origin || defaultConfig.backend_origin).replace(/\/+$/, "");
    config.tenant_id = nonEmpty(config.tenant_id, defaultConfig.tenant_id);
    config.device_id = nonEmpty(config.device_id, defaultConfig.device_id);
    config.host_name = nonEmpty(config.host_name, defaultConfig.host_name);
    config.profile = nonEmpty(config.profile, defaultConfig.profile);
    config.os_name = nonEmpty(config.os_name, defaultConfig.os_name);
    if (!Array.isArray(config.youtube_study_keywords)) {
      config.youtube_study_keywords = defaultConfig.youtube_study_keywords.slice();
    }
    config.youtube_study_keywords = config.youtube_study_keywords
      .map((item) => String(item || "").trim().toLowerCase())
      .filter(Boolean);
    return config;
  }

  function nonEmpty(value, fallback) {
    const normalized = String(value || "").trim();
    return normalized || fallback;
  }

  function isAllowedBackendOrigin(origin) {
    try {
      const parsed = new URL(String(origin || ""));
      if (parsed.protocol !== "http:") {
        return false;
      }
      return parsed.hostname === "127.0.0.1" || parsed.hostname === "localhost" || parsed.hostname === "::1";
    } catch (error) {
      return false;
    }
  }

  function telemetryEndpoint(config) {
    const safe = cleanConfig(config);
    if (!isAllowedBackendOrigin(safe.backend_origin)) {
      return "";
    }
    return `${safe.backend_origin}/api/v1/devices/${encodeURIComponent(safe.device_id)}/telemetry-events`;
  }

  function normalizeDomain(rawNavigationUrl) {
    try {
      const parsed = new URL(String(rawNavigationUrl || ""));
      if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
        return "";
      }
      let host = String(parsed.hostname || "").trim().toLowerCase();
      if (host.startsWith("www.")) {
        host = host.slice(4);
      }
      return host;
    } catch (error) {
      return "";
    }
  }

  function classifyDomain(domain, rawNavigationUrl, config) {
    const value = String(domain || "").toLowerCase();
    if (!value) {
      return { category: constants.categoryUnknown, youtube_study_match: false };
    }
    if (isYouTubeDomain(value)) {
      const study = containsStudyKeyword(rawNavigationUrl, config);
      return {
        category: study ? constants.categoryStudy : constants.categoryVideoStreaming,
        youtube_study_match: study
      };
    }
    if (matchesAny(value, ["docs.python.org", "learn.microsoft.com", "github.com", "khanacademy.org", "coursera.org", "edx.org"])) {
      return { category: constants.categoryStudy, youtube_study_match: false };
    }
    if (matchesAny(value, ["netflix.com", "primevideo.com", "hotstar.com", "disneyplus.com"])) {
      return { category: constants.categoryVideoStreaming, youtube_study_match: false };
    }
    if (matchesAny(value, ["facebook.com", "instagram.com", "x.com", "twitter.com", "reddit.com", "snapchat.com"])) {
      return { category: constants.categorySocialMedia, youtube_study_match: false };
    }
    if (matchesAny(value, ["steampowered.com", "epicgames.com", "roblox.com"])) {
      return { category: constants.categoryGaming, youtube_study_match: false };
    }
    if (matchesAny(value, ["amazon.com", "flipkart.com", "myntra.com"])) {
      return { category: constants.categoryShopping, youtube_study_match: false };
    }
    return { category: constants.categoryUnknown, youtube_study_match: false };
  }

  function isYouTubeDomain(domain) {
    return domain === "youtube.com" || domain.endsWith(".youtube.com") || domain === "youtu.be";
  }

  function matchesAny(domain, candidates) {
    return candidates.some((candidate) => domain === candidate || domain.endsWith(`.${candidate}`));
  }

  function containsStudyKeyword(rawNavigationUrl, config) {
    const safe = cleanConfig(config);
    const lower = String(rawNavigationUrl || "").toLowerCase();
    return safe.youtube_study_keywords.some((keyword) => keyword && lower.includes(keyword));
  }

  function detectBrowserName(userAgent, braveHint) {
    const value = String(userAgent || "").toLowerCase();
    if (braveHint) {
      return "brave";
    }
    if (value.includes("edg/")) {
      return "edge";
    }
    if (value.includes("chrome/") || value.includes("chromium/")) {
      return "chrome";
    }
    return "chromium";
  }

  function buildTelemetryEvent(input) {
    const config = cleanConfig(input && input.config);
    const domain = normalizeDomain(input && input.rawNavigationUrl);
    if (!domain) {
      return null;
    }
    const browserName = nonEmpty(input && input.browserName, "chromium");
    const classification = classifyDomain(domain, input && input.rawNavigationUrl, config);
    const observedAt = input && input.observedAt ? new Date(input.observedAt) : new Date();
    const eventID = nonEmpty(input && input.eventID, createEventID(domain, observedAt));
    return {
      id: eventID,
      type: constants.eventTypeBrowserObserved,
      source: constants.eventSourceBrowserExtension,
      observed_at: observedAt.toISOString(),
      tenant_id: config.tenant_id,
      device_id: config.device_id,
      host_name: config.host_name,
      app_name: browserName,
      process_id: 0,
      path_hash: "",
      metadata: {
        browser_name: browserName,
        domain: domain,
        category: classification.category,
        source_kind: constants.sourceKindLiveIngested,
        evidence_scope: constants.evidenceScopeLive,
        evidence_detail: constants.evidenceDetail,
        url_mode: constants.urlModeDomainOnly,
        stored_url_mode: constants.urlModeDomainOnly,
        visit_count: "1",
        youtube_study_match: String(classification.youtube_study_match)
      }
    };
  }

  function createEventID(domain, observedAt) {
    const timestamp = observedAt instanceof Date ? observedAt.getTime() : Date.now();
    const random = Math.random().toString(36).slice(2, 10);
    return `browser-extension-${timestamp}-${hashDomain(domain)}-${random}`;
  }

  function hashDomain(domain) {
    const value = String(domain || "");
    let hash = 2166136261;
    for (let i = 0; i < value.length; i += 1) {
      hash ^= value.charCodeAt(i);
      hash = Math.imul(hash, 16777619);
    }
    return (hash >>> 0).toString(16);
  }

  function buildIngestRequest(event, config) {
    if (!event) {
      return null;
    }
    const safe = cleanConfig(config);
    return {
      tenant_id: safe.tenant_id,
      device_id: safe.device_id,
      host_name: safe.host_name,
      profile: safe.profile,
      os_name: safe.os_name,
      events: [event]
    };
  }

  function hasForbiddenPayloadKeys(value) {
    if (Array.isArray(value)) {
      return value.some(hasForbiddenPayloadKeys);
    }
    if (value && typeof value === "object") {
      return Object.keys(value).some((key) => {
        const lowered = key.toLowerCase();
        return forbiddenPayloadKeys.includes(lowered) || hasForbiddenPayloadKeys(value[key]);
      });
    }
    return false;
  }

  const api = Object.freeze({
    constants,
    defaultConfig,
    cleanConfig,
    isAllowedBackendOrigin,
    telemetryEndpoint,
    normalizeDomain,
    classifyDomain,
    detectBrowserName,
    buildTelemetryEvent,
    buildIngestRequest,
    hasForbiddenPayloadKeys
  });

  root.TraceDeckBrowserPrivacy = api;
  if (typeof module !== "undefined" && module.exports) {
    module.exports = api;
  }
})(typeof globalThis !== "undefined" ? globalThis : this);
