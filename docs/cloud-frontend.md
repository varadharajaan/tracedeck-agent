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

Run locally:

```powershell
python ./devctl.py status
python ./devctl.py server restart
python ./devctl.py test phase69
```

Deploy with SAM:

```powershell
python ./devctl.py sam build
python ./devctl.py sam deploy
python ./devctl.py sam outputs
python ./devctl.py sam tail
```

`devctl.py sam deploy` saves CloudFormation outputs to:

```text
data/local/output/stack-outputs.txt
data/local/output/frontend-url.txt
```

Seed and verify cloud archive data:

```powershell
python ./devctl.py cloud seed
python ./devctl.py cloud smoke
python ./devctl.py cloud newman
python ./devctl.py test phase72
```

Phase 72 adds `scripts/local/upload-cloud-sample-phase72.ps1`, which writes a
small metadata-only JSONL gzip archive under `data/local/cloud-seed/`, uploads
it to the configured S3 bucket, and records a manifest under `data/local`.
`scripts/local/smoke-phase72.ps1` refreshes the deployed Lambda S3 summary,
verifies sampled browser rows, Chrome/Edge/Brave grouping, study-safe
inference, non-study YouTube, and then reads again to prove the Lambda memory
cache reports a hit.

The Lambda frontend intentionally renders safe metadata only: S3 object keys,
sizes, storage class, timestamps, host labels, browser names, domains,
categories, study-safe status, and counts. It does not render passwords,
cookies, tokens, page titles, private content, endpoint payloads, provider
secrets, push endpoints, alert bodies, payment data, or raw provider payloads.
