# Data Vault Operations Runbook

## Overview

The data vault provides encrypted storage for sensitive artifacts (VEID, support attachments, marketplace artifacts) with role/org access control and immutable audit logging.

## Key Components

- **VaultService** (`pkg/data_vault`): encrypt/decrypt and access control.
- **AuditLogger**: hash-chained audit events for every decrypt/read.
- **Portal API**: `/api/v1/vault/*` endpoints in provider daemon.
- **Metrics**: Prometheus counters under `virtengine_data_vault_*`.

## Access Control

- Owners always have access to their own blobs.
- Org membership is required for org-scoped blobs.
- Roles are enforced per scope:
  - `veid`: administrator, support_agent
  - `support`: administrator, support_agent
  - `market`: administrator, service_provider
  - `audit`: administrator, moderator

## Audit Logging

- Every decrypt/read is logged with requester, org, blob, purpose/reason, and timestamp.
- Hash chaining links each record via `previous_hash` → `hash`.

## Portal API Endpoints

All endpoints require wallet-signed or HMAC auth.

- `POST /api/v1/vault/blobs`
- `GET /api/v1/vault/blobs/{blobId}`
- `GET /api/v1/vault/blobs/{blobId}/metadata`
- `DELETE /api/v1/vault/blobs/{blobId}`
- `GET /api/v1/vault/audit`

## Metrics & Alerts

Prometheus counters:

- `virtengine_data_vault_access_total{scope,action,success}`
- `virtengine_data_vault_access_denied_total{scope,action}`
- `virtengine_data_vault_audit_failures_total`

Suggested alert:

- High rate of access denials per requester/org in a short window (possible misuse or credential compromise).

## Operational Checks

1. **Health**: ensure provider daemon API is up and responding.
2. **Audit Integrity**: verify audit log hash chain consistency periodically.
3. **Metrics**: review access denial spikes.

## Troubleshooting

- **Unauthorized errors**: confirm org membership + role assignment.
- **Audit gaps**: check audit logger configuration and Prometheus alerts.
- **Payload too large**: increase `VaultMaxPayloadBytes` or use streaming.

## References

- `pkg/data_vault/README.md`
- `docs/operations/` for operational procedures
