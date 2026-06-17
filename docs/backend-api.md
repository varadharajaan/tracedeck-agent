# Backend API

The local backend runs at:

```text
http://127.0.0.1:18080
```

## Main Routes

| Route | Purpose |
| --- | --- |
| `GET /health` | Backend health |
| `GET /` | TraceDeck Console |
| `GET /browser-activity` | Browser Intelligence page |
| `GET /v1-old` | Legacy dashboard fallback |
| `GET /api/v1/devices` | Enrolled device list |
| `POST /api/v1/devices/enroll` | Enroll or update a device |
| `POST /api/v1/devices/{deviceId}/telemetry-events` | Ingest agent telemetry |
| `GET /api/v1/devices/{deviceId}/telemetry-status` | Stored event status |
| `GET /api/v1/tenants/{tenantId}/browser-activity` | Browser activity rows |
| `GET /api/v1/tenants/{tenantId}/delivery-assurance` | Delivery route proof |
| `GET /api/v1/runtime-status-center` | Runtime status summary |

## Source Labels

API responses distinguish source provenance:

- `live_ingested` for real agent data
- `s3_sample` for cloud archive samples
- `demo_seed` for demo-only rows

Default live views hide demo rows.
