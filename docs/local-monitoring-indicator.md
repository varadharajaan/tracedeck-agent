# Local Monitoring Indicator

Phase 112 adds a visible local monitoring indicator for endpoint-user trust.
It is implemented as a local status API plus generated JSON, text, and HTML
artifacts:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/get-local-monitoring-indicator.ps1
```

Default outputs:

```text
data/local/output/local-monitoring-indicator.json
data/local/output/local-monitoring-indicator.txt
data/local/output/local-monitoring-indicator.html
```

The API backing the indicator is:

```text
GET /api/v1/local-monitoring-indicator
```

The dashboard renders the same proof in the `Local Monitoring Indicator` panel
on the Deployment page.

## Proof Model

The indicator exposes:

- visible indicator readiness
- local status page path
- backend runtime readiness
- consent visibility
- denied sensitive collection proof
- typed transparency mode: `visible_indicator_required`
- metadata-only evidence scope

## Verification

Phase 112 can be rerun with:

```powershell
python ./devctl.py test phase112
python ./devctl.py test smoke112
python ./devctl.py test newman112
python ./devctl.py test verify112
```

The verifier covers backend API tests, dashboard DOM and JavaScript checks,
live smoke, Newman, contract audit refresh, and root hygiene.

## Privacy

The local indicator is metadata-only. It does not collect passwords,
screenshots, raw URLs, page titles, cookies, tokens, private content, provider
secrets, alert bodies, keylogging, hidden collection bypasses, payment data,
raw provider payloads, camera, or microphone data.
