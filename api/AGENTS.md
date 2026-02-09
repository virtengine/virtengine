# API Layer (api/openapi) — AGENTS Guide

## Module Overview
- Purpose: OpenAPI specifications for VirtEngine blockchain and provider portal APIs, used for client generation and documentation.
- Use `api/openapi/virtengine-api.yaml` for chain-level queries and transactions; use `api/openapi/portal_api.yaml` for provider portal operations.
- Key assets:
  - `api/openapi/virtengine-api.yaml` entry point (`api/openapi/virtengine-api.yaml:1`).
  - `api/openapi/portal_api.yaml` entry point (`api/openapi/portal_api.yaml:1`).

## Architecture
- `api/openapi/virtengine-api.yaml` — blockchain API specification, shared tags for modules, and server definitions (`api/openapi/virtengine-api.yaml:1`).
- `api/openapi/portal_api.yaml` — provider portal REST API specification, with ADR-036 auth headers (`api/openapi/portal_api.yaml:10`).
- Both specs define `servers`, `security`, `tags`, `paths`, and `components` blocks for downstream generators.

## Core Concepts
- Authentication:
  - Chain API supports wallet signature auth and MFA for sensitive operations (`api/openapi/virtengine-api.yaml:11`).
  - Portal API uses ADR-036 wallet signature headers or optional HMAC auth (`api/openapi/portal_api.yaml:10`).
- Rate limits and tiers are documented in the chain spec (`api/openapi/virtengine-api.yaml:19`).
- Tags organize module endpoints; examples include Provider, Market, VEID (`api/openapi/virtengine-api.yaml:47`).

## Usage Examples

### Query a VEID identity
```bash
curl https://api.virtengine.com/virtengine/veid/v1/identity/<bech32_address>
```

### Provider portal health check
```bash
curl https://provider.example.com:8443/api/v1/health
```

### Portal request with wallet headers (ADR-036)
```
X-VE-Address: <bech32>
X-VE-Timestamp: <unix_ms>
X-VE-Nonce: <random_hex>
X-VE-Signature: <base64_sig>
X-VE-PubKey: <base64_pubkey>
```

## Implementation Patterns
- Add new endpoints in the appropriate spec file under `paths:` and assign a tag (`api/openapi/virtengine-api.yaml:85`, `api/openapi/portal_api.yaml:67`).
- Update shared schemas under `components:` to avoid copy/paste drift.
- Keep `servers` in sync with deployment environments (`api/openapi/virtengine-api.yaml:35`, `api/openapi/portal_api.yaml:37`).
- Anti-patterns:
  - Do not introduce new auth headers without documenting them in `security` and `description` blocks.
  - Do not add paths without rate-limit metadata for chain endpoints.

## API Reference
- Chain API metadata: `openapi`, `info`, `servers`, `security`, `tags` (`api/openapi/virtengine-api.yaml:1`).
- Portal API metadata: auth headers and server variables (`api/openapi/portal_api.yaml:8`).

## Dependencies & Environment
- Specs are documentation-only; no runtime dependencies.
- Client generation depends on OpenAPI tooling of choice (e.g., `openapi-generator`).

## Configuration
- No runtime configuration. Update server URLs and auth headers directly in the spec files.

## Testing
- Validate specs with your OpenAPI linter or generator before release.

## Troubleshooting
- OpenAPI lint errors:
  - Cause: Missing schema references or invalid YAML indentation.
  - Fix: Validate `api/openapi/virtengine-api.yaml` and `api/openapi/portal_api.yaml`.
