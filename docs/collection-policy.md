# Collection Policy

Collection policy controls transparency mode, browser metadata, foreground app
metadata, software inventory metadata, media metadata, and deny-only sensitive
capabilities.

The default TraceDeck profile requires a visible monitoring indicator and
stores browser domain/category instead of full URLs.

Phase 110 foreground app collection is controlled by:

```yaml
collection:
  foreground_app:
    enabled: true
    window_title_mode: none
```

The foreground app collector stores app name, process id, hashed executable
path, active state, and title mode only. Window titles are deny-only in the
current schema and are not read by the Windows adapter.

Phase 111 software inventory collection is controlled by:

```yaml
collection:
  software:
    enabled: true
    inventory_mode: metadata_only
```

The software collector stores display metadata and hashed snapshot identities
only. It emits `software.installed` and `software.uninstalled` after a local
baseline snapshot exists. It does not store install paths, file contents,
installer payloads, screenshots, passwords, cookies, tokens, raw URLs, page
titles, private content, provider secrets, alert bodies, payment data, or raw
provider payloads.

Phase 3 browser collection reads Chrome, Edge, and Brave history databases when
available. The collector stores `browser.domain.observed` events with browser
name, domain, category, visit count, requested URL mode, and stored URL mode.
The stored URL mode is always `domain_only`; raw URLs and page titles are not
persisted.

Use `--disable-browser-history` for controlled local smokes or diagnostics that
should avoid reading live browser history.

Phase 4 alert evaluation consumes the stored event metadata. Blocked-domain and
non-study YouTube alerts use the persisted domain/category fields, not raw
browser URLs or page titles. A YouTube video id may be hashed when configured,
but the raw id is not stored.
