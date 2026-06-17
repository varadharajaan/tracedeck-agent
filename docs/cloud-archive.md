# Cloud Archive

TraceDeck can upload compressed metadata batches to S3.

Current layout:

```text
s3://<bucket>/tenants/<tenant_id>/devices/<device_id>/hosts/<host_name>/date=YYYY-MM-DD/hour=HH/*.jsonl.gz
```

Useful checks:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/test-live-s3-archive.ps1
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/local/check-live-s3-metrics.ps1
```

Archive batches are metadata-only and should not contain passwords, screenshots,
raw URLs, page titles, cookies, tokens, or private content.
