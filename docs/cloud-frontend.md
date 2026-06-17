# Cloud Frontend

The cloud admin frontend lives under `sam-app/` and deploys with AWS SAM.

It uses:

- Lambda Function URL
- S3 archive as source of truth
- In-memory cache to reduce S3 calls
- Cache hit/miss metrics in the UI

## Commands

```powershell
python ./devctl.py sam build
python ./devctl.py sam deploy
python ./devctl.py sam outputs
python ./devctl.py doctor
```

Outputs are written under:

```text
data/local/output/
```

Important files:

- `data/local/output/stack-outputs.txt`
- `data/local/output/frontend-url.txt`
- `data/local/output/runtime-doctor.json`
