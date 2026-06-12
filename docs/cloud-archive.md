# Cloud Archive

TraceDeck archives hourly compressed JSONL batches to S3 when archive is
enabled. The archive path uses one bucket with tenant, device, host, date, and
hour prefixes.

The family profile bucket is:

```text
tracedeck-agent-family-varadha-996335889295-ap-south-1
```

Lifecycle target: Standard for 90 days, Standard-IA until day 365, then archive.

Phase 2 stages compressed JSONL batches under `data/local/outbox/archive/`.
S3 upload is available through the AWS SDK adapter and is skipped when
`--archive-dry-run` is enabled.

Phase 2B enables continuous mode. The first continuous cycle stages an archive
batch immediately, then subsequent staging follows `archive.upload_interval`.
