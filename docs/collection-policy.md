# Collection Policy

Collection policy controls transparency mode, browser metadata, foreground app
metadata, media metadata, and deny-only sensitive capabilities.

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
