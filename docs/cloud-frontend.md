# Cloud Frontend

TraceDeck Phase 69 includes an optional admin frontend in `sam-app/`.

It is deployed with AWS SAM as a public Lambda Function URL:

- `FunctionUrlConfig` is enabled with `AuthType: NONE`.
- API Gateway resources are intentionally not defined.
- S3 is the source of truth for cloud archive summaries.
- Lambda module memory caches S3 summaries and reports cache hit/miss
  percentages in the UI.
- The UI includes a source selector for either Lambda/S3 or a browser-side
  localhost backend such as `http://127.0.0.1:18080`.
- Browser rows include provenance fields: `source_kind`, `evidence_scope`,
  and `evidence_detail`. S3 sampled rows are labeled as `s3_sample` so cloud
  archive evidence is not confused with live local telemetry.

Run locally:

```powershell
python ./devctl.py status
python ./devctl.py server restart
python ./devctl.py test live
python ./devctl.py test phase69
```

Deploy with SAM:

```powershell
python ./devctl.py sam build
python ./devctl.py sam deploy
python ./devctl.py sam outputs
python ./devctl.py sam tail
python ./devctl.py doctor
```

`devctl.py sam deploy` saves CloudFormation outputs to:

```text
data/local/output/stack-outputs.txt
data/local/output/frontend-url.txt
data/local/output/runtime-doctor.json
data/local/output/runtime-doctor.txt
```

Seed and verify cloud archive data:

```powershell
python ./devctl.py cloud seed
python ./devctl.py cloud smoke
python ./devctl.py cloud newman
python ./devctl.py test phase72
python ./devctl.py test phase73
```

Phase 72 adds `scripts/local/upload-cloud-sample-phase72.ps1`, which writes a
small metadata-only JSONL gzip archive under `data/local/cloud-seed/`, uploads
it to the configured S3 bucket, and records a manifest under `data/local`.
`scripts/local/smoke-phase72.ps1` refreshes the deployed Lambda S3 summary,
verifies sampled browser rows, Chrome/Edge/Brave grouping, study-safe
inference, non-study YouTube, and then reads again to prove the Lambda memory
cache reports a hit.

Phase 74 adds `python ./devctl.py doctor` as an operator assurance command.
With cloud checks enabled, it reads the saved Function URL, checks
`/api/health`, refreshes `/api/s3-summary`, reads `/api/s3-summary` again to
prove a cache hit, and saves a JSON/text report under `data/local/output/`.
Use `python ./devctl.py doctor --skip-cloud` when only the local backend should
be checked.

The Lambda frontend intentionally renders safe metadata only: S3 object keys,
sizes, storage class, timestamps, host labels, browser names, domains,
categories, study-safe status, and counts. It does not render passwords,
cookies, tokens, page titles, private content, endpoint payloads, provider
secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
