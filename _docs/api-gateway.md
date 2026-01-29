# API Gateway & Management

This repository ships a DB-less Kong API gateway configuration for local development and CI. The gateway provides:

- Authentication/authorization (API keys + ACL)
- Request/response transformation
- API versioning routes
- Analytics/monitoring (Prometheus + Grafana)
- API key management via Kong Admin API
- Developer portal (Swagger UI)
- Basic lifecycle signals (deprecation headers + tags)

## Services & Ports

| Service | Port | Description |
|---------|------|-------------|
| Kong Proxy | 8000 | API gateway entrypoint |
| Kong Admin | 8001 | Gateway management API |
| Kong Metrics | 8100 | Prometheus scrape target |
| Dev Portal | 3001 | Swagger UI (gateway docs) |
| Prometheus | 9095 | Metrics UI |
| Grafana | 3002 | Dashboards |

## Routes

- `/v1/waldur` -> `waldur:8080` (deprecated)
- `/v2/waldur` -> `waldur:8080`
- `/v1/chain` -> `virtengine-node:1317`
- `/v1/provider` -> `provider-daemon:8443`
- `/portal` -> Swagger UI

## API Keys

Two default consumers are defined in `config/kong/kong.yaml`:

- `dev-portal` key: `dev-portal-key` (group: `public`)
- `admin-ops` key: `admin-ops-key` (group: `admin`)

Use the key in `x-api-key` or `apikey` headers.

## Lifecycle Signals

- `v1` routes include `Deprecation: true` and `Sunset: 2026-12-31` headers.
- Route tags (`lifecycle:deprecated`, `lifecycle:stable`, `lifecycle:beta`) help track lifecycle state.

## Example Requests

```bash
curl -H "x-api-key: dev-portal-key" http://localhost:8000/v2/waldur/
curl -H "x-api-key: dev-portal-key" http://localhost:8000/v1/chain/cosmos/base/tendermint/v1beta1/node_info
```
