# Privacy

TraceDeck is designed as a metadata-first endpoint observability tool.

## Collected Metadata

- Process and app names
- Foreground app state where supported
- Browser domain/category activity
- Study-safe classification signals
- Device health metadata
- Software inventory metadata
- Agent health and archive status
- Delivery route status for email, push, and dashboard alerts

## Denied Sensitive Data

The policy boundary denies:

- Passwords
- Keystrokes
- Cookies
- Auth tokens
- Private messages
- Raw page content
- Raw URLs by default
- Page titles by default
- Camera or microphone capture
- Covert collection

Screenshot collection is not part of the safe baseline implemented in this
repo. Any future expansion must be explicit, consent-based, visible, typed in
policy, and separately reviewed.

## Demo Data

Seeded demo rows are labelled with `source_kind=demo_seed` and are hidden from
live views unless a route explicitly opts into demo proof.

## Storage

Local runtime data is kept under `data/local/`. Logs are kept under
`logs/local/`. S3 archive data is stored by tenant, device, host, date, and hour.
