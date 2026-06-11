# TraceDeck Agent Plan

Last updated: 2026-06-11

## Project Overview

TraceDeck Agent is a transparent, privacy-first endpoint observability agent for
Windows, macOS, and Linux laptops and managed devices. The first MVP is
Windows-first, but the architecture uses platform adapters so macOS and Linux
can be added without rewriting the domain, policy, storage, archive, alerting,
or telemetry layers.

The implementation stack is Go-first. The product may still monitor developer
software such as `java.exe`, because TraceDeck tracks study and coding tools.

## Product Principles

- Transparent and consent-based. This is not spyware.
- Collection is governed by a strongly typed YAML policy with validated enums,
  safe defaults, and fail-fast startup checks.
- Credential capture, keylogging, cookies, tokens, private messages, camera, and
  microphone collection are hard-denied capabilities.
- Covert screenshot capture is not part of TraceDeck. Evidence should come from
  typed metadata such as app name, executable hash, domain, category, media file
  name, media path, and browser video classification when explicitly enabled.
- Browser tracking stores domain and category by default, not full URL.
- Local policy evaluation comes before cloud policy sync.
- The agent must remain lightweight, bounded, observable, and safe to run on a
  personal laptop.
- Local retention is bounded, with S3-backed archive and offline retry.
- OpenTelemetry is the export boundary for metrics, logs, and traces.
- Extensibility is designed in from Phase 1 through interfaces and clear
  component boundaries.

## Tech Stack

| Area | Choice | Why |
| --- | --- | --- |
| Agent language | Go | Small deployable binary, strong concurrency, cross-platform build support |
| CLI | Cobra | Clean commands for run, validate-config, service install, service uninstall |
| Config and policy | YAML via `gopkg.in/yaml.v3` | Human-readable policies and easy local editing |
| Config validation | Go structs, enums, `go-playground/validator`, generated JSON Schema | Strongly typed collection rules |
| Config reload | `fsnotify` | Watch policy/config changes without restart |
| Service manager | Windows Service, macOS launchd, Linux systemd | Native background execution per OS |
| Process monitoring | `github.com/shirou/gopsutil` plus OS adapters | Practical process inventory and metrics |
| Active window | OS-specific adapters | Foreground app attribution without screen capture |
| Local API | `net/http` with `go-chi/chi` if routing grows | Lightweight localhost API for browser extensions |
| Local DB | SQLite with pure-Go driver where practical | Local bounded buffer with simple deployment |
| DB migrations | `golang-migrate/migrate` | Repeatable SQLite schema changes |
| Cloud archive | AWS S3 via AWS SDK for Go v2 | Hourly offload, durable replay, lifecycle retention |
| Email alerts | Notifier interface with AWS SES/SNS and SMTP adapters | Automatic anomaly notifications |
| Telemetry | OpenTelemetry Go SDK and OTLP exporters | Standards-based metrics, logs, and traces |
| Logging | `log/slog` plus rotating file sink | Structured logs, 10 MiB rotation, bounded retention |
| Testing | Go test, Testify, gomock or mockery | Unit tests and interface-driven mocks |
| Linting | golangci-lint, staticcheck, gofmt, go vet | Enforced Go quality baseline |
| Security scans | govulncheck, gosec, Syft SBOM | Dependency and binary supply-chain checks |
| Packaging | GoReleaser | Repeatable Windows, macOS, and Linux builds and checksums |
| Local dependencies | Docker Compose | OpenTelemetry Collector and optional storage locally |
| Browser extension | TypeScript | Chrome, Edge, and Brave extension codebase |
| Backend later | Go service | Keeps the product Go-native end to end |
| Dashboard later | React | Product dashboard and policy management UI |
| Analytics later | Python optional | Offline anomaly experiments, not the always-running agent |

## Architecture Style

TraceDeck uses a clean architecture style with ports and adapters.

Core domain code does not depend on Windows, macOS, Linux, SQLite,
OpenTelemetry, or HTTP frameworks. Those integrations live behind interfaces.
This keeps the agent easy to test and lets future collectors, exporters, and
platform adapters be added without rewriting the policy engine or event
pipeline.

## Core Components

| Component | Responsibility |
| --- | --- |
| Process Collector | Samples running processes and emits app lifecycle observations |
| Active Window Collector | Detects foreground process and attributes active duration |
| Browser Local API | Receives active-tab domain events from browser extensions |
| Media Evidence Collector | Captures media player app, file name, and file path when policy allows |
| Software Inventory Collector | Detects install and uninstall changes in later phases |
| Policy Engine | Evaluates local YAML rules against app, browser, and install events |
| Domain Categorizer | Maps domains to education, ai-tools, video, social, adult, torrent, etc. |
| YouTube Study Classifier | Classifies video pages as study or non-study using typed local rules first |
| Anomaly Engine | Rule-based anomaly detection first, ML-ready later |
| Alert Notifier | Sends high and critical anomaly emails with dedupe and cooldown |
| Report Generator | Produces weekly AI summaries and PDF/email reports |
| Compliance Scorer | Scores productivity, risk, device health, and policy posture |
| Tamper Detector | Detects agent stop, extension disablement, policy edits, and clock changes |
| Policy Template Catalog | Provides packaged policies for family, student, school, and business use |
| Event Pipeline | Normalizes, validates, enriches, buffers, and exports events |
| SQLite Store | Bounded durable local queue and local summaries |
| S3 Archiver | Uploads hourly batches to S3 and retries when the laptop returns online |
| OTLP Exporter | Sends metrics, logs, and traces to OpenTelemetry Collector |
| Health Reporter | Emits agent CPU, memory, queue size, and export failure metrics |
| Local Indicator | Shows the user that monitoring is active |

## Design Patterns

- Ports and adapters for collectors, storage, exporters, categorization, and
  policy sync.
- Strategy pattern for domain categorization, policy rules, scoring, and anomaly
  rules.
- Repository pattern for SQLite persistence.
- Pipeline pattern for ingest -> normalize -> redact -> evaluate -> buffer ->
  export.
- Circuit breaker and retry policy around OTLP export.
- Outbox pattern for S3 uploads and email alerts.
- Idempotent batch manifests for cloud archive replay.
- Bounded queue with backpressure to protect laptop performance.
- Composition root in `cmd/tracedeck-agent` for explicit dependency wiring.
- Plugin-style collector registration so future collectors can be added without
  changing the pipeline core.

## Initial Repository Structure

```text
tracedeck-agent/
  README.md
  LICENSE
  go.mod
  go.sum
  .gitignore
  .golangci.yml
  Makefile
  plan.md
  docs/
    architecture.md
    privacy.md
    policy-config.md
    collection-policy.md
    cloud-archive.md
    alerting.md
    telemetry-schema.md
    roadmap.md
    security.md
    testing.md
  agent/
    cmd/
      tracedeck-agent/
        main.go
    internal/
      app/
        bootstrap.go
        lifecycle.go
      collector/
        process/
        activewindow/
        browser/
        media/
        software/
      domain/
        collection/
        event/
        policy/
        alert/
        anomaly/
        scoring/
        report/
        compliance/
        tamper/
        template/
        categorizer/
      pipeline/
      storage/
        sqlite/
      archive/
        s3/
      exporter/
        otlp/
      notifier/
        email/
      config/
      schema/
      logging/
      platform/
        windows/
        darwin/
        linux/
      service/
        windows/
        darwin/
        linux/
      health/
    migrations/
      sqlite/
    test/
      fixtures/
  browser-extension/
    shared/
    chrome/
    edge/
    brave/
  backend/
    README.md
  dashboard/
    README.md
  deployments/
    docker-compose.yml
    otel-collector-config.yaml
  examples/
    policies/
      ai-btech-student.yaml
  scripts/
    verify.ps1
    smoke-local.ps1
```

## Go Package Boundaries

| Package Area | Rule |
| --- | --- |
| `domain/*` | Pure business logic. No OS, DB, HTTP, or OTLP imports |
| `collector/*` | Converts external signals into domain events |
| `pipeline` | Owns event flow, redaction, policy evaluation, buffering, export |
| `storage/sqlite` | Implements storage interfaces only |
| `exporter/otlp` | Implements telemetry exporter interfaces only |
| `platform/windows` | Contains Windows API wrappers |
| `platform/darwin` | Contains macOS API and command wrappers |
| `platform/linux` | Contains Linux API and desktop-environment wrappers |
| `service/windows` | Windows service install, uninstall, start, stop integration |
| `service/darwin` | macOS launchd plist install, unload, and status integration |
| `service/linux` | Linux systemd unit install, enable, start, stop integration |
| `cmd/tracedeck-agent` | CLI and dependency wiring only |

## Platform Support Matrix

| Capability | Windows | macOS | Linux |
| --- | --- | --- | --- |
| Background service | Windows Service | launchd agent/daemon | systemd user/system service |
| Process inventory | gopsutil + Win32 | gopsutil + ps/sysctl | gopsutil + /proc |
| Foreground app | Win32 foreground window APIs | Accessibility APIs / AppleScript fallback | X11/Wayland desktop adapters |
| Software inventory | Registry, uninstall keys, package hints | Applications folder, pkg receipts, Homebrew | dpkg/rpm/snap/flatpak/appimage hints |
| Media metadata | process args, file handles where available | process args, lsof where permitted | process args, lsof/proc where permitted |
| Browser extension | Chrome, Edge, Brave | Chrome, Edge, Brave, Safari later | Chrome, Edge, Brave, Firefox later |
| Local DB | SQLite | SQLite | SQLite |
| S3 archive | AWS SDK for Go v2 | AWS SDK for Go v2 | AWS SDK for Go v2 |
| OTel export | OTLP | OTLP | OTLP |
| Local indicator | tray/toast | menu bar item/notification | tray/notification where desktop exists |

Implementation priority:

1. Windows MVP.
2. macOS parity for process, foreground app, service lifecycle, browser events,
   local storage, S3 archive, and alerts.
3. Linux parity for process, service lifecycle, browser events, local storage,
   S3 archive, and alerts.
4. Linux foreground-app parity split by desktop stack because Wayland support
   varies by compositor and permissions.

## Domain Model

| Model | Key Fields |
| --- | --- |
| DeviceIdentity | tenant_id, device_id, user_id, host_name, agent_version |
| AppObservation | process_id, app_name, path_hash, started_at, ended_at |
| ForegroundInterval | app_name, path_hash, start_time, end_time, active_seconds |
| BrowserVisit | browser_name, profile_hash, domain, category, observed_at |
| BrowserVideo | browser_name, domain, video_id_hash, title, study_related, reason |
| MediaPlayback | app_name, file_name, file_path, path_hash, started_at, ended_at |
| SoftwareChange | app_name, publisher, version, action, observed_at |
| CollectionPolicy | typed capability settings, evidence modes, retention, export targets |
| Policy | allowed domains, blocked domains, categories, app lists, thresholds |
| PolicyViolation | rule_id, severity, action, reason, event_ref |
| Anomaly | type, severity, reason, score, event_ref |
| AlertNotification | recipient, severity, subject, event_ref, dedupe_key, sent_at |
| WeeklyReport | report_id, device_id, week_start, study_score, risk_score, summary |
| ComplianceScore | device_id, productivity_score, risk_score, health_score, posture |
| TamperEvent | type, severity, observed_at, reason, evidence |
| PolicyTemplate | template_id, audience, defaults, locked_capabilities, version |
| UploadBatch | batch_id, device_id, hour_window, local_path, s3_key, status |
| DailySummary | study time, coding time, ai-tool time, video/social/gaming time, score |

## MVP Phases

### Phase 0: Governance and Repo Foundation

- [ ] Replace old prompt assumptions with Go-first TraceDeck rules.
- [ ] Add README, license, `.gitignore`, Go module, Makefile, and lint config.
- [ ] Add docs skeleton for architecture, privacy, policy, telemetry, roadmap,
      collection policy, cloud archive, alerting, security, and testing.
- [ ] Add sample `ai-btech-student.yaml` policy.
- [ ] Add strongly typed config structs, enum validation plan, and generated
      schema target.

### Phase 1: Local Windows Agent

- [ ] Build `tracedeck-agent run` command.
- [ ] Add config loader and policy loader.
- [ ] Add process collector.
- [ ] Add active foreground app collector.
- [ ] Track active duration only for foreground app.
- [ ] Add SQLite local buffer.
- [ ] Enforce 90-day local retention with disk-size guardrails.
- [ ] Add bounded queue and retry-safe pipeline.
- [ ] Export app usage metrics to OpenTelemetry Collector.
- [ ] Add local structured logs with rotation.
- [ ] Add media evidence collector for configured media players.
- [ ] Emit media-player policy violations with file name and path when policy
      allows metadata collection.
- [ ] Add unit tests for config, policy, pipeline, storage, and exporter seams.

### Phase 1B: Platform Adapter Contracts

- [ ] Define platform-neutral interfaces for process inventory, foreground app,
      software inventory, media evidence, service lifecycle, local indicator,
      and OS health.
- [ ] Add stub adapters for `windows`, `darwin`, and `linux`.
- [ ] Add build tags so OS-specific code cannot leak into domain packages.
- [ ] Add contract tests that every platform adapter must satisfy.

### Phase 2: S3 Archive and Email Alerting

- [ ] Create or configure one S3 bucket for TraceDeck archives.
- [ ] Use one bucket with prefixes by tenant, device, host, date, and hour.
- [ ] Add S3 lifecycle: Standard for 90 days, Standard-IA until day 365, archive
      tier after day 365.
- [ ] Add server-side encryption for S3 objects.
- [ ] Add hourly upload worker for local event batches.
- [ ] Add offline retry so missed hourly pushes upload the next time the laptop
      is online.
- [ ] Add idempotent upload manifests to avoid duplicate cloud objects.
- [ ] Add email notifier with recipient `varathu09@gmail.com`.
- [ ] Send high and critical anomaly alerts automatically.
- [ ] Add dedupe and cooldown to avoid alert storms.

### Phase 3: Browser Activity

- [ ] Build shared TypeScript extension code.
- [ ] Support Chrome, Edge, and Brave.
- [ ] Capture active tab domain only.
- [ ] Send events to localhost agent.
- [ ] Add domain categorizer.
- [ ] Add optional video-page metadata mode for YouTube classification.
- [ ] Classify YouTube videos as study or non-study using configured keywords
      and category rules.
- [ ] Suppress alerts for study-related YouTube videos.
- [ ] Alert on non-study YouTube/video-streaming usage beyond policy thresholds.
- [ ] Export browser domain and category duration metrics.
- [ ] Add privacy tests proving full URLs are redacted by default.

### Phase 4: Policy and Anomaly Engine

- [ ] Add allowlist and blocklist evaluation.
- [ ] Add warning and critical category rules.
- [ ] Detect policy violations.
- [ ] Add rule-based anomaly detection.
- [ ] Add daily productivity and study score.
- [ ] Add software install and uninstall detection.
- [ ] Add risky software detection for torrent clients, VPN/proxy tools,
      unknown browsers, game launchers, unsigned executables, and installers
      from Downloads.
- [ ] Add tamper detection for agent stop, browser extension disablement,
      policy file edits, clock changes, upload backlog, and exporter failures.

### Phase 5: Backend and Dashboard

- [ ] Add Go backend API contract.
- [ ] Add device enrollment model.
- [ ] Add policy sync API.
- [ ] Add usage aggregation API.
- [ ] Add React dashboard for study vs non-study time, apps, categories,
      anomalies, inventory, and policy violations.
- [ ] Add role-based dashboard views: parent, student self-view, school admin,
      and business manager.
- [ ] Add no-code alert rule builder.
- [ ] Add consent and audit center.
- [ ] Add policy template catalog and template versioning.
- [ ] Add weekly AI report generation and email/PDF delivery.
- [ ] Add device health and compliance score views.

### Phase 6: SaaS Readiness

- [ ] Add tenant-level policies.
- [ ] Add RBAC and audit logs.
- [ ] Add retention controls.
- [ ] Add reports.
- [ ] Add billing model.
- [ ] Add SIEM and OTLP integration options.
- [ ] Add tiered plans: Free, Family Pro, School, Business, Enterprise.
- [ ] Add cloud archive retention plans and cost controls.
- [ ] Add template marketplace packaging.

### Phase 7: macOS and Linux Endpoint Support

- [ ] Implement macOS launchd service support.
- [ ] Implement macOS process, foreground app, software inventory, media
      metadata, local indicator, and health adapters.
- [ ] Implement Linux systemd service support.
- [ ] Implement Linux process, software inventory, media metadata, local
      indicator, and health adapters.
- [ ] Add Linux foreground app support for X11 first and Wayland where
      compositor permissions allow it.
- [ ] Add cross-platform installer/package artifacts.
- [ ] Add platform-specific docs for permissions, consent, and limitations.

## OpenTelemetry Metrics

```text
tracedeck.device.active.duration_seconds
tracedeck.app.usage.duration_seconds
tracedeck.browser.domain.duration_seconds
tracedeck.browser.category.duration_seconds
tracedeck.software.install.count
tracedeck.software.remove.count
tracedeck.policy.violation.count
tracedeck.anomaly.count
tracedeck.productivity.score
tracedeck.agent.export.failure.count
tracedeck.agent.buffer.queue_size
tracedeck.agent.s3.upload.count
tracedeck.agent.s3.upload.failure.count
tracedeck.agent.s3.backlog.bytes
tracedeck.alert.email.sent.count
tracedeck.alert.email.failure.count
tracedeck.agent.cpu.percent
tracedeck.agent.memory.bytes
```

## Common Telemetry Attributes

```text
tenant.id
device.id
user.id
agent.version
os.name
os.version
host.name
app.name
app.path_hash
browser.name
browser.profile_hash
domain
category
policy.action
severity
alert.recipient_hash
s3.bucket_hash
s3.prefix
media.file_name
media.path_hash
browser.video_id_hash
browser.study_related
```

## Privacy Requirements

- Do not collect passwords or credentials.
- Do not collect keystrokes.
- Do not collect camera or microphone data.
- Do not collect private message content.
- Do not collect cookies, auth tokens, or browser storage.
- Do not collect covert screenshots.
- Do not store full URLs by default.
- Store domain and category by default.
- Collect media file name and media file path only when the typed policy enables
  media evidence metadata.
- Collect browser video title or video identifier only when the typed policy
  enables video classification metadata.
- Hash executable paths where practical.
- Keep local storage bounded.
- Keep local events for 90 days by default, with disk-size guardrails.
- Upload archive batches to S3 hourly when enabled.
- Retry missed S3 uploads when the laptop is next online.
- Encrypt data in transit.
- Encrypt S3 objects at rest.
- Prefer local redaction before persistence and export.
- Show a visible local indicator that monitoring is active.

## Policy Configuration

The policy file remains YAML-first and maps to strongly typed Go structs.
Unknown fields fail validation. String options use enums, not free-form input.

```yaml
tenant_id: family-varadha
device_id: laptop-cousin-001
profile: ai-btech-student

collection:
  transparency_mode: visible_indicator_required
  browser:
    url_mode: domain_only
    collect_page_title: false
    youtube_classification: enabled
    youtube_video_id_mode: hashed
  media:
    collect_file_name: true
    collect_file_path: true
    path_mode: full_path
  sensitive_capabilities:
    credentials: deny
    keystrokes: deny
    cookies: deny
    tokens: deny
    private_messages: deny
    screenshots: deny

retention:
  local_ttl_days: 90
  max_local_storage_mb: 2048

archive:
  enabled: true
  provider: s3
  bucket: tracedeck-agent-family-varadha-996335889295-ap-south-1
  prefix_template: tenants/{tenant_id}/devices/{device_id}/hosts/{host_name}/date={yyyy}-{mm}-{dd}/hour={hh}/
  upload_interval: 1h
  retry_when_online: true
  storage_class_days:
    standard: 90
    standard_ia_until: 365
    archive_after: 365

alerts:
  enabled: true
  email:
    provider: ses
    to:
      - varathu09@gmail.com
    min_severity: high
    cooldown_minutes: 30

study_apps:
  - Code.exe
  - python.exe
  - java.exe
  - jupyter.exe
  - chrome.exe
  - msedge.exe
  - brave.exe

blocked_apps:
  - vlc.exe
  - qbittorrent.exe
  - utorrent.exe
  - steam.exe
  - epicgameslauncher.exe

allowed_domains:
  - udemy.com
  - coursera.org
  - github.com
  - stackoverflow.com
  - docs.python.org
  - openai.com
  - huggingface.co
  - kaggle.com
  - microsoft.com
  - learn.microsoft.com

warn_categories:
  - video-streaming
  - social-media
  - gaming
  - shopping

critical_categories:
  - adult-content
  - torrent
  - malware
  - proxy-vpn

youtube_study_keywords:
  - python
  - system design
  - math
  - maths
  - machine learning
  - artificial intelligence
  - coding
  - java
  - data structures
  - algorithms

alert_rules:
  media_player_used:
    enabled: true
    severity: high
    include_media_file_metadata: true
  non_study_youtube:
    enabled: true
    severity: medium
    threshold_minutes_per_day: 30
  adult_or_torrent:
    enabled: true
    severity: critical
  blocked_app_opened:
    enabled: true
    severity: high
```

## S3 Archive Strategy

Use one S3 bucket with prefixes per tenant, device, host, date, and hour.
This is preferred over one bucket per host because bucket names are globally
unique, lifecycle rules are easier to manage centrally, and prefixes still give
clean isolation.

Example object layout:

```text
s3://tracedeck-agent-family-varadha-996335889295-ap-south-1/
  tenants/family-varadha/
    devices/laptop-cousin-001/
      hosts/HOSTNAME/
        date=2026-06-11/
          hour=14/
            events-20260611T140000Z-20260611T145959Z.jsonl.gz
            manifest.json
```

The agent writes hourly compressed JSONL batches from the SQLite outbox. If the
laptop is off or offline, batches remain local until the next online window.
Upload is idempotent: the manifest tracks batch id, row range, checksum, and S3
key.

Lifecycle:

- Day 0 to 90: S3 Standard.
- Day 91 to 365: S3 Standard-IA.
- After day 365: archive tier.

## Alerting Strategy

Email alerts are generated from anomaly and policy violation events.

Initial automatic alerts:

- VLC, Windows Media Player, Movies & TV, torrent clients, game launchers, or
  other blocked apps opened.
- Media player opened with media file name/path when available and enabled.
- Adult, torrent, malware, or proxy-vpn category accessed.
- YouTube or streaming usage classified as non-study and above threshold.
- Unknown executable launched from Downloads.
- New software installed outside allowlist.
- Late-night usage beyond configured hours.
- Repeated policy violations.

Emails should include:

- Device id and host name.
- Event time.
- Severity.
- App or browser.
- Domain/category when relevant.
- Media file name/path when relevant and enabled.
- Why the alert fired.
- Local event id and S3 object key when already uploaded.

Emails must not include passwords, cookies, tokens, screenshots, or private
message content.

## Monetization Strategy

TraceDeck should be packaged as privacy-aware endpoint activity, productivity,
and risk observability. The first paid wedge is Family Pro, but the architecture
must support schools, coaching centers, BYOD, and small businesses.

Paid feature set:

- Weekly AI summary reports with study hours, coding hours, entertainment time,
  top apps/sites, anomalies, late-night usage, software installs, and risk
  score.
- Device compliance score covering productivity posture, risky software, device
  health, agent health, and archive/export health.
- Risky software detection for torrent clients, VPN/proxy tools, game launchers,
  unsigned executables, outdated apps, unknown browsers, and installers from
  Downloads.
- Policy template catalog for AI BTech Student, School Laptop, Coaching Center,
  Developer Workstation, Family Pro, Exam Mode, and Small Business Productivity.
- Role-based dashboard views for parent, student self-view, school admin, and
  business manager.
- No-code alert rule builder for app, category, time-window, install, tamper,
  and threshold rules.
- Consent and audit center showing what is collected, who receives alerts,
  policy history, pause controls, export, and delete controls.
- Cloud archive retention plans with local-only, 90-day, 1-year, and custom
  retention tiers.
- AI category classifier for YouTube, websites, apps, and unknown usage.
- Tamper detection as a paid trust feature.

Suggested pricing:

| Plan | Target | Features |
| --- | --- | --- |
| Free | Trial/single device | 1 device, local only, 7-day local retention, basic report |
| Family Pro | Families | 5 devices, S3 archive, email alerts, weekly AI reports |
| School | Schools/coaching | Per student/device, templates, admin dashboard, reports |
| Business | SMB/BYOD | Per endpoint/month, inventory, compliance, audit logs |
| Enterprise | Large orgs | SSO, RBAC, SIEM/OTLP, custom retention, policy APIs |

## Quality Gates

- `gofmt` and `goimports` pass.
- `go test ./...` passes.
- `go test -race ./...` passes where supported.
- `golangci-lint run` passes.
- `govulncheck ./...` passes.
- `gosec ./...` passes or findings are explicitly accepted.
- SQLite migrations apply cleanly.
- Local OpenTelemetry Collector receives app usage metrics.
- Browser events are rejected unless they come from the allowed localhost
  boundary.
- Privacy tests prove forbidden data types are not collected.
- Config schema generation passes.
- S3 lifecycle config is validated before bucket provisioning.
- Hourly archive replay works after simulated offline time.
- Email alert dedupe and cooldown tests pass.

## Initial Definition of Done

Phase 1 is done when a Windows laptop can run the local agent, load the sample
policy, track foreground app active time, store events in SQLite, export app
usage metrics to an OpenTelemetry Collector, record configured media metadata
for blocked media playback, and pass the local quality gates. Cross-platform
support is done when the same domain pipeline runs on macOS and Linux through
platform adapters with documented OS-specific limitations.

## Out of Scope for Phase 1

- Cloud backend.
- React dashboard.
- Billing.
- ML anomaly detection.
- Remote policy sync.
- Browser extension.
- Blocking applications at OS level.
- Capturing window titles unless explicitly enabled by typed policy.
- Capturing full URLs.
- Capturing screenshots, passwords, credentials, cookies, tokens, private
  messages, camera, microphone, or keystrokes.

## Next Action

Scaffold Phase 0: create the Go module, README, `.gitignore`, Makefile,
`.golangci.yml`, docs skeleton, typed sample policy, cloud archive docs, alerting
docs, and the first agent package layout.
