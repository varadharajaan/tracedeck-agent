# Security

Security goals:

- Keep collection metadata-only.
- Use typed policy and generated schema.
- Keep provider secrets out of logs and dashboard responses.
- Keep generated data under `data/local/` and logs under `logs/local/`.
- Separate demo proof from live evidence.
- Use scripted verification for local, cloud, and UI checks.

Useful commands:

```powershell
python ./devctl.py test quality
powershell -NoProfile -ExecutionPolicy Bypass -File ./scripts/verify/check-root-clean.ps1
```
