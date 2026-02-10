# Lifecycle Control Operations Guide

VE-4E: Resource lifecycle control via Waldur with signed callbacks.

## Overview

This guide covers the operational aspects of managing resource lifecycle operations through the VirtEngine marketplace. Lifecycle actions (start, stop, restart, suspend, resume, resize, terminate) are executed via the Waldur backend with cryptographic callback verification.

## Architecture

```
┌─────────────────┐         ┌──────────────────┐         ┌─────────────┐
│   On-Chain      │         │  Provider Daemon │         │   Waldur    │
│  x/market       │◄───────►│  Lifecycle Ctrl  │◄───────►│   Backend   │
│                 │         │                  │         │             │
│  State Machine  │         │  Idempotency     │         │  Resources  │
│  Callback Valid │         │  Retry/Rollback  │         │  Actions    │
└─────────────────┘         └──────────────────┘         └─────────────┘
```

## Lifecycle Actions

### Action Types

| Action      | From State(s)           | To State     | Description                    |
|-------------|------------------------|--------------|--------------------------------|
| `start`     | Suspended              | Active       | Start a stopped resource       |
| `stop`      | Active                 | Suspended    | Stop a running resource        |
| `restart`   | Active                 | Active       | Restart a running resource     |
| `suspend`   | Active                 | Suspended    | Suspend with state preservation|
| `resume`    | Suspended              | Active       | Resume a suspended resource    |
| `resize`    | Active                 | Active       | Resize resource allocation     |
| `terminate` | Active, Suspended      | Terminating  | Permanently remove resource    |
| `provision` | Pending, Accepted      | Provisioning | Initial resource provisioning  |

### State Machine

```
               ┌──────────┐
               │ Pending  │
               └────┬─────┘
                    │ provision
                    ▼
               ┌──────────────┐
               │ Provisioning │
               └──────┬───────┘
                      │ complete
                      ▼
               ┌──────────┐◄────── start/resume ──────┐
   ┌──────────►│  Active  │                           │
   │           └────┬─────┘                           │
   │ restart        │                                 │
   └────────────────┤ stop/suspend                    │
                    ▼                           ┌─────┴─────┐
               ┌──────────┐                     │ Suspended │
               │Suspended │─────────────────────►           │
               └────┬─────┘                     └─────┬─────┘
                    │ terminate                       │
                    ▼                                 │
               ┌───────────┐◄─────── terminate ───────┘
               │Terminating│
               └─────┬─────┘
                     │ complete
                     ▼
               ┌───────────┐
               │Terminated │
               └───────────┘
```

## Operation States

Lifecycle operations progress through these states:

| State               | Description                                      |
|---------------------|--------------------------------------------------|
| `pending`           | Operation queued, not yet started                |
| `executing`         | Operation in progress at Waldur                  |
| `awaiting_callback` | Waiting for signed callback confirmation         |
| `completed`         | Operation finished successfully                  |
| `failed`            | Operation failed (may trigger rollback)          |
| `rolled_back`       | Operation was rolled back after failure          |
| `cancelled`         | Operation was cancelled before execution         |

## Provider Daemon Configuration

### Lifecycle Controller Settings

```yaml
# config.yaml
lifecycle:
  enabled: true
  state_file_path: "data/lifecycle_state.json"
  operation_timeout: 5m
  callback_ttl: 1h
  max_concurrent_ops: 10
  retry_interval: 30s
  cleanup_interval: 1h
  operation_retention_days: 7
  enable_audit_logging: true
  callback_url: "https://provider.example.com/callbacks/lifecycle"
```

### Configuration Parameters

| Parameter                 | Default   | Description                                    |
|---------------------------|-----------|------------------------------------------------|
| `enabled`                 | `true`    | Enable lifecycle control                       |
| `state_file_path`         | See above | Path for persisting operation state            |
| `operation_timeout`       | `5m`      | Timeout for individual operations              |
| `callback_ttl`            | `1h`      | How long callbacks remain valid                |
| `max_concurrent_ops`      | `10`      | Maximum concurrent lifecycle operations        |
| `retry_interval`          | `30s`     | Interval between retry attempts                |
| `cleanup_interval`        | `1h`      | Interval for cleaning up old operations        |
| `operation_retention_days`| `7`       | Days to keep completed operations              |
| `enable_audit_logging`    | `true`    | Enable audit logging for lifecycle actions     |
| `callback_url`            | —         | URL for Waldur callback notifications          |

## Signed Callbacks

### Callback Structure

Lifecycle operations require cryptographically signed callbacks to confirm state transitions:

```json
{
  "id": "lcb_lco_abc123_8f3a2b1c",
  "operation_id": "lco_abc123def456",
  "allocation_id": "alloc-12345",
  "action": "start",
  "success": true,
  "result_state": "Active",
  "provider_address": "provider1abc...",
  "nonce": "8f3a2b1c4d5e6f7a",
  "timestamp": "2024-01-15T10:30:00Z",
  "expires_at": "2024-01-15T11:30:00Z",
  "signature": "base64-encoded-ed25519-signature"
}
```

### Callback Validation

The x/market keeper validates callbacks with these checks:

1. **Expiry Check**: Callback must not be expired
2. **Nonce Uniqueness**: Prevents replay attacks
3. **Signature Verification**: Ed25519 signature from authorized provider
4. **State Transition Validity**: Resulting state must be valid for the action
5. **Operation Matching**: Callback must match a pending operation

### Signature Generation

Providers sign callbacks using their registered keys:

```go
// Signing payload construction
h := sha256.New()
h.Write([]byte(callback.ID))
h.Write([]byte(callback.OperationID))
h.Write([]byte(callback.AllocationID))
h.Write([]byte(callback.Action))
h.Write([]byte(successStr))       // "success" or "failure"
h.Write([]byte(callback.ResultState.String()))
h.Write([]byte(callback.ProviderAddress))
h.Write([]byte(callback.Nonce))
h.Write([]byte(fmt.Sprintf("%d", callback.Timestamp.Unix())))
signingPayload := h.Sum(nil)

// Sign with Ed25519 key
signature := ed25519.Sign(privateKey, signingPayload)
```

## Idempotency

### Idempotency Keys

Operations are protected against duplicates using idempotency keys:

- Key format: SHA256 hash of `allocation_id:action:hour_timestamp`
- Keys are valid for a 1-hour window
- Duplicate requests return the existing operation result

### Example

```
Allocation: alloc-12345
Action: start
Timestamp: 2024-01-15T10:45:00Z (truncated to hour: 2024-01-15T10:00:00Z)

Idempotency Key: SHA256("alloc-12345:start:1705312800")[:16] → "8f3a2b1c4d5e6f7a"
```

## Rollback Policies

### Policy Types

| Policy      | Behavior                                           |
|-------------|---------------------------------------------------|
| `none`      | No automatic rollback; operation stays failed     |
| `automatic` | Automatically attempt rollback on failure         |
| `manual`    | Flag for manual intervention required             |
| `retry`     | Retry operation before considering rollback       |

### Retry Configuration

Default retry behavior:
- Maximum retries: 3
- Retry interval: 30 seconds (configurable)
- Exponential backoff: Not enabled by default

## Audit Logging

### Audit Entry Format

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "event_type": "lifecycle_action_completed",
  "operation_id": "lco_abc123def456",
  "allocation_id": "alloc-12345",
  "action": "start",
  "from_state": "Suspended",
  "to_state": "Active",
  "provider_address": "provider1abc...",
  "duration_ms": 2500,
  "success": true
}
```

### Audit Events

| Event                           | Description                              |
|---------------------------------|------------------------------------------|
| `lifecycle_action_requested`    | Lifecycle action was requested           |
| `lifecycle_action_started`      | Execution began                          |
| `lifecycle_action_completed`    | Action completed successfully            |
| `lifecycle_action_failed`       | Action failed                            |
| `lifecycle_callback_received`   | Signed callback received                 |
| `lifecycle_rollback_triggered`  | Rollback was initiated                   |

## Metrics

### Prometheus Metrics

The lifecycle controller exposes these metrics:

```
# Counter: Total lifecycle operations by action and status
virtengine_lifecycle_operations_total{action="start",status="success"} 150
virtengine_lifecycle_operations_total{action="start",status="failed"} 5

# Gauge: Currently executing operations
virtengine_lifecycle_operations_executing{action="start"} 2

# Histogram: Operation duration in milliseconds
virtengine_lifecycle_operation_duration_ms{action="start",quantile="0.5"} 1200
virtengine_lifecycle_operation_duration_ms{action="start",quantile="0.99"} 5000

# Counter: Callback validation results
virtengine_lifecycle_callbacks_total{result="valid"} 148
virtengine_lifecycle_callbacks_total{result="expired"} 2
virtengine_lifecycle_callbacks_total{result="invalid_signature"} 0
```

## Troubleshooting

### Common Issues

#### 1. Operation Stuck in "executing" State

**Symptoms**: Operation doesn't progress to "awaiting_callback"

**Causes**:
- Network issues between provider daemon and Waldur
- Waldur API is unresponsive
- Resource doesn't exist in Waldur

**Resolution**:
```bash
# Check Waldur connectivity
curl -H "Authorization: Token $WALDUR_TOKEN" \
  $WALDUR_API_URL/api/marketplace-resources/$RESOURCE_UUID/

# Check provider daemon logs
journalctl -u provider-daemon -f | grep lifecycle

# Force retry (if stuck > timeout)
virtengined provider lifecycle retry --operation-id lco_abc123def456
```

#### 2. Callback Signature Validation Failed

**Symptoms**: `ErrLifecycleCallbackInvalidSignature` in logs

**Causes**:
- Provider key mismatch
- Signing payload constructed incorrectly
- Timestamp drift

**Resolution**:
```bash
# Verify provider key registration
virtengined query market provider-info $PROVIDER_ADDRESS

# Check time synchronization
timedatectl status

# Verify signing payload matches callback data
virtengined debug callback validate --callback-file callback.json
```

#### 3. Idempotency Conflict

**Symptoms**: New operation returns existing operation ID

**Causes**:
- Duplicate request within same hour window
- Expected behavior for idempotent operations

**Resolution**:
```bash
# Check existing operation status
virtengined query market lifecycle-operation $OPERATION_ID

# Wait for existing operation to complete, or
# use a new hour window for retry
```

#### 4. Rollback Failed

**Symptoms**: `ErrLifecycleRollbackFailed` in logs

**Causes**:
- Resource in inconsistent state
- Rollback action not supported for this state

**Resolution**:
```bash
# Check current resource state
virtengined query market allocation $ALLOCATION_ID

# Manual intervention required
virtengined provider lifecycle cancel --operation-id lco_abc123def456
virtengined provider lifecycle sync --allocation-id $ALLOCATION_ID
```

### Log Analysis

```bash
# Filter lifecycle-related logs
journalctl -u provider-daemon | grep -E "(lifecycle|callback)"

# Monitor real-time lifecycle events
virtengined monitor lifecycle --provider $PROVIDER_ADDRESS

# Export lifecycle operation history
virtengined export lifecycle-ops --since 24h --format json > lifecycle_ops.json
```

### Health Checks

```bash
# Check lifecycle controller health
curl -s http://localhost:26660/health/lifecycle | jq

# Expected output:
{
  "status": "healthy",
  "pending_operations": 0,
  "executing_operations": 2,
  "failed_operations": 0,
  "avg_completion_time_ms": 1850
}
```

## Integration Testing

### Mock Waldur Setup

For testing lifecycle operations without a real Waldur instance:

```bash
# Start mock Waldur server
go run ./tests/mocks/waldur_server.go --port 8080

# Configure provider daemon to use mock
export WALDUR_API_URL=http://localhost:8080
export WALDUR_TOKEN=mock-token

# Run lifecycle integration tests
go test -v -tags=e2e.integration ./tests/integration/lifecycle_test.go
```

### Test Scenarios

1. **Happy Path**: Start → Stop → Start cycle
2. **Failure Handling**: Simulated Waldur timeout with retry
3. **Callback Validation**: Valid and invalid signature tests
4. **Concurrent Operations**: Multiple operations on same allocation
5. **Idempotency**: Duplicate request handling

## API Reference

### CLI Commands

```bash
# Request lifecycle action
virtengined tx market lifecycle-action \
  --allocation-id $ALLOCATION_ID \
  --action start \
  --from $WALLET_ADDRESS

# Query operation status
virtengined query market lifecycle-operation $OPERATION_ID

# List operations for allocation
virtengined query market lifecycle-operations \
  --allocation-id $ALLOCATION_ID \
  --state pending

# Cancel pending operation
virtengined tx market cancel-lifecycle-operation \
  --operation-id $OPERATION_ID \
  --from $WALLET_ADDRESS
```

### gRPC Endpoints

| Service                     | Method                    | Description                    |
|-----------------------------|---------------------------|--------------------------------|
| `market.Msg`                | `RequestLifecycleAction`  | Request a lifecycle action     |
| `market.Msg`                | `SubmitLifecycleCallback` | Submit signed callback         |
| `market.Msg`                | `CancelLifecycleOperation`| Cancel pending operation       |
| `market.Query`              | `LifecycleOperation`      | Query operation by ID          |
| `market.Query`              | `LifecycleOperations`     | List operations with filters   |

## See Also

- [Offering Sync Guide](./offering-sync-guide.md)
- [ServiceDesk Sync](./servicedesk-sync.md)
- [Provider Daemon Configuration](../provider-daemon-config.md)
