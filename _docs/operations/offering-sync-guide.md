# Offering Sync Guide

This document describes the automatic offering synchronization feature that syncs on-chain marketplace offerings to Waldur.

## Overview

The offering sync feature (VE-2D) eliminates manual mapping between VirtEngine on-chain offerings and Waldur marketplace offerings. When enabled, the provider daemon automatically:

1. Subscribes to on-chain offering events (create/update/terminate)
2. Synchronizes offerings to Waldur with the correct field mappings
3. Tracks sync state including checksums, versions, and errors
4. Handles retries with exponential backoff
5. Dead-letters persistently failing offerings
6. Runs periodic reconciliation to detect and fix drift

## Configuration

### CLI Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--waldur-offering-sync-enabled` | `false` | Enable automatic offering sync |
| `--waldur-offering-sync-state-file` | `data/offering_sync_state.json` | Path for sync state persistence |
| `--waldur-customer-uuid` | (required) | Waldur customer/organization UUID |
| `--waldur-category-map` | (optional) | Path to JSON file mapping categories to Waldur category UUIDs |
| `--waldur-offering-sync-interval` | `300` | Reconciliation interval in seconds |
| `--waldur-offering-sync-max-retries` | `5` | Max retries before dead-lettering |

### Environment Variables

All flags can be set via environment variables with the prefix `VIRTENGINE_`:

```bash
export VIRTENGINE_WALDUR_OFFERING_SYNC_ENABLED=true
export VIRTENGINE_WALDUR_CUSTOMER_UUID=<your-customer-uuid>
export VIRTENGINE_WALDUR_OFFERING_SYNC_INTERVAL=300
```

### Category Map JSON

Create a JSON file mapping VirtEngine offering categories to Waldur category UUIDs:

```json
{
  "compute": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "storage": "yyyyyyyy-yyyy-yyyy-yyyy-yyyyyyyyyyyy",
  "hpc": "zzzzzzzz-zzzz-zzzz-zzzz-zzzzzzzzzzzz"
}
```

## Migration from Manual Mapping

If you were previously using `--waldur-offering-map`:

1. The flag is now **deprecated** but still works for backward compatibility
2. When `--waldur-offering-sync-enabled=true`, the manual map is ignored
3. Existing offerings will be auto-discovered via Waldur `backend_id` lookup
4. New offerings will be created automatically

### Migration Steps

1. Ensure Waldur offerings have correct `backend_id` values (VirtEngine offering IDs)
2. Set `--waldur-offering-sync-enabled=true`
3. Provide `--waldur-customer-uuid` for new offering creation
4. Optionally provide `--waldur-category-map` for category mappings
5. Remove `--waldur-offering-map` after confirming sync works

## Field Mappings

On-chain offering fields are mapped to Waldur as follows:

| VirtEngine Field | Waldur Field | Notes |
|------------------|--------------|-------|
| `ID` | `backend_id` | Used for cross-reference |
| `Name` | `name` | Truncated to 255 chars |
| `Description` | `description` | Full description |
| `Category` | `type` | Mapped via category map |
| `State` | state action | `Active→activate`, `Paused→pause`, `Archived→archive` |
| `BasePrice` | `unit_price` | Divided by currency denominator |
| `EncryptedSecrets` | (excluded) | Never synced to Waldur |

## Sync States

Each offering can be in one of these sync states:

| State | Description |
|-------|-------------|
| `pending` | Initial state, never synced |
| `synced` | Successfully synced to Waldur |
| `out_of_sync` | Drift detected, needs re-sync |
| `failed` | Sync attempt failed |
| `retrying` | Waiting for retry attempt |
| `dead_lettered` | Failed after max retries |

## Retry Behavior

Failed syncs are retried with exponential backoff:

- Base backoff: 30 seconds
- Max backoff: 1 hour
- Max retries: 5 (configurable)
- Backoff formula: `min(base * 2^(attempt-1), max)`

Example retry schedule:
1. First retry: 30s
2. Second retry: 60s
3. Third retry: 120s
4. Fourth retry: 240s
5. Fifth retry: 480s
6. After 5 failures: Dead-lettered

## Dead Letter Queue

Offerings that fail after max retries are moved to the dead letter queue:

- Dead-lettered offerings are not automatically retried
- Use the reprocess command to retry dead-lettered items
- Dead letter entries include: offering ID, error message, attempt count, timestamp

### Reprocessing Dead Letters

```bash
# View dead-lettered offerings
virtengine provider-daemon dead-letter list

# Reprocess a specific offering
virtengine provider-daemon dead-letter reprocess <offering-id>

# Reprocess all dead-lettered offerings
virtengine provider-daemon dead-letter reprocess --all
```

## Reconciliation

The reconciliation job runs periodically to:

1. Detect offerings that are out of sync (checksum mismatch)
2. Find offerings missing from Waldur
3. Queue sync tasks for drifted offerings

Reconciliation interval is controlled by `--waldur-offering-sync-interval` (default: 5 minutes).

### Manual Reconciliation

Trigger immediate reconciliation via the admin API:

```bash
curl -X POST http://localhost:8080/admin/offering-sync/reconcile
```

## Monitoring

### Metrics

The sync worker exposes Prometheus-compatible metrics:

| Metric | Type | Description |
|--------|------|-------------|
| `offering_sync_total` | Counter | Total sync operations |
| `offering_sync_successful` | Counter | Successful syncs |
| `offering_sync_failed` | Counter | Failed syncs |
| `offering_sync_dead_lettered` | Counter | Dead-lettered items |
| `offering_sync_drift_detected` | Counter | Drift detections |
| `offering_sync_reconciliations` | Counter | Reconciliation runs |
| `offering_sync_queue_depth` | Gauge | Current queue depth |
| `offering_sync_active_syncs` | Gauge | Active sync operations |
| `offering_sync_dead_letter_size` | Gauge | Dead letter queue size |
| `offering_sync_duration_seconds` | Histogram | Sync operation duration |
| `offering_sync_events_received` | Counter | Events received from chain |
| `offering_sync_events_processed` | Counter | Events successfully processed |
| `offering_sync_events_dropped` | Counter | Events dropped (queue full) |

### Audit Logs

The sync worker logs structured JSON audit entries for:

- **Sync attempts**: Every create/update/disable operation
- **Reconciliation runs**: Periodic drift checks
- **Dead letter events**: Items dead-lettered or reprocessed

Audit log format:

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "offering_id": "ve1abc123...",
  "waldur_uuid": "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx",
  "action": "update",
  "success": true,
  "duration_ns": 1234567890,
  "retry_count": 0,
  "provider_address": "virtengine1...",
  "checksum": "sha256:abc123..."
}
```

## Troubleshooting

### Offering Not Syncing

1. Check sync state: `virtengine provider-daemon offering-sync status <offering-id>`
2. Verify provider address matches
3. Check Waldur API connectivity
4. Review audit logs for errors

### High Failure Rate

1. Check Waldur API health
2. Verify customer UUID is correct
3. Check category mappings
4. Review rate limiting (Waldur may be throttling)

### Drift Keeps Occurring

1. Check if external changes are made in Waldur
2. Verify checksum calculation is stable
3. Consider increasing reconciliation interval

### Dead Letter Queue Growing

1. Review error messages in dead-lettered items
2. Fix underlying issues (API errors, invalid data)
3. Reprocess after fixes

## Recovery Procedures

### State File Corruption

If the state file is corrupted:

1. Stop the provider daemon
2. Back up the corrupted file
3. Delete or rename the state file
4. Restart the daemon
5. Run manual reconciliation to rebuild state

```bash
mv data/offering_sync_state.json data/offering_sync_state.json.bak
virtengine provider-daemon start
curl -X POST http://localhost:8080/admin/offering-sync/reconcile
```

### Full Resync

To force a complete resync of all offerings:

1. Stop the provider daemon
2. Delete the state file
3. Restart with `--waldur-offering-sync-enabled=true`
4. The reconciliation job will sync all offerings

### Waldur API Recovery

If Waldur API was down and offerings are out of sync:

1. Wait for Waldur API to recover
2. Trigger manual reconciliation
3. Monitor dead letter queue for any permanent failures
4. Reprocess dead-lettered items after investigation

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                     Provider Daemon                              │
│                                                                 │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │   CometBFT   │───▶│    Event     │───▶│    Sync      │      │
│  │ Subscription │    │    Loop      │    │   Queue      │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│                                                   │              │
│                                                   ▼              │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────┐      │
│  │    State     │◀───│    Sync      │───▶│   Waldur     │      │
│  │    Store     │    │   Worker     │    │    API       │      │
│  └──────────────┘    └──────────────┘    └──────────────┘      │
│         │                    │                                   │
│         ▼                    ▼                                   │
│  ┌──────────────┐    ┌──────────────┐                          │
│  │  State File  │    │ Audit Logs / │                          │
│  │    (JSON)    │    │   Metrics    │                          │
│  └──────────────┘    └──────────────┘                          │
│                                                                 │
│  ┌──────────────┐                                               │
│  │ Reconcile    │ (runs every sync_interval)                    │
│  │    Loop      │                                               │
│  └──────────────┘                                               │
└─────────────────────────────────────────────────────────────────┘
```

## Security Considerations

- **Encrypted secrets are never synced**: The `EncryptedSecrets` field is excluded from Waldur sync
- **API tokens**: Waldur API token should be stored securely (env var, not CLI flag)
- **State file**: Contains offering IDs and Waldur UUIDs, protect accordingly
- **Audit logs**: May contain sensitive information, configure log retention

## Changelog

- **VE-2D**: Initial implementation of automatic offering sync
