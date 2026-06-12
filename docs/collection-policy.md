# Collection Policy

Collection policy controls transparency mode, browser metadata, media metadata,
and deny-only sensitive capabilities.

The default TraceDeck profile requires a visible monitoring indicator and
stores browser domain/category instead of full URLs.

Phase 3 browser collection reads Chrome, Edge, and Brave history databases when
available. The collector stores `browser.domain.observed` events with browser
name, domain, category, visit count, requested URL mode, and stored URL mode.
The stored URL mode is always `domain_only`; raw URLs and page titles are not
persisted.

Use `--disable-browser-history` for controlled local smokes or diagnostics that
should avoid reading live browser history.
