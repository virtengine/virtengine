# Service Desk Sync Operations Guide

This document describes the operational aspects of the Jira/Waldur service desk sync feature (VE-3E).

## Overview

The service desk bridge enables bi-directional synchronization between VirtEngine on-chain support tickets and external service desk systems (Jira Service Desk and Waldur).

### Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌────────────────┐
│  VirtEngine     │◄───►│  Service Desk    │◄───►│  Jira / Waldur │
│  On-Chain       │     │  Bridge Service  │     │  External APIs │
│  Support Module │     │                  │     │                │
└─────────────────┘     └──────────────────┘     └────────────────┘
         │                       │                       │
         │                       ▼                       │
         │              ┌──────────────────┐             │
         │              │  Audit Logger    │             │
         │              └──────────────────┘             │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐     ┌──────────────────┐     ┌────────────────┐
│  Artifact Store │     │  Retry Queue     │     │  Webhook Server│
│  (Attachments)  │     │                  │     │  (Callbacks)   │
└─────────────────┘     └──────────────────┘     └────────────────┘
```

## Configuration

### Environment Variables

```bash
# Jira Configuration
VIRTENGINE_SERVICEDESK_JIRA_URL=https://company.atlassian.net
VIRTENGINE_SERVICEDESK_JIRA_USERNAME=service-account
VIRTENGINE_SERVICEDESK_JIRA_API_TOKEN=<api-token>  # SENSITIVE
VIRTENGINE_SERVICEDESK_JIRA_PROJECT_KEY=VESUPPORT
VIRTENGINE_SERVICEDESK_JIRA_WEBHOOK_SECRET=<secret>  # SENSITIVE

# Waldur Configuration
VIRTENGINE_SERVICEDESK_WALDUR_URL=https://waldur.example.com/api
VIRTENGINE_SERVICEDESK_WALDUR_TOKEN=<api-token>  # SENSITIVE
VIRTENGINE_SERVICEDESK_WALDUR_ORG_UUID=<uuid>
VIRTENGINE_SERVICEDESK_WALDUR_WEBHOOK_SECRET=<secret>  # SENSITIVE

# Sync Configuration
VIRTENGINE_SERVICEDESK_SYNC_INTERVAL=30s
VIRTENGINE_SERVICEDESK_SYNC_BATCH_SIZE=50
VIRTENGINE_SERVICEDESK_CONFLICT_RESOLUTION=on_chain_wins

# Webhook Server
VIRTENGINE_SERVICEDESK_WEBHOOK_LISTEN=:8480
VIRTENGINE_SERVICEDESK_WEBHOOK_PATH_PREFIX=/webhooks
VIRTENGINE_SERVICEDESK_WEBHOOK_REQUIRE_SIGNATURE=true
```

### YAML Configuration

```yaml
servicedesk:
  enabled: true
  
  jira:
    base_url: "https://company.atlassian.net"
    project_key: "VESUPPORT"
    issue_type: "Service Request"
    timeout: 30s
    # username and api_token from secrets
  
  waldur:
    base_url: "https://waldur.example.com/api"
    organization_uuid: "org-uuid"
    timeout: 30s
    # token from secrets
  
  sync:
    sync_interval: 30s
    batch_size: 50
    conflict_resolution: on_chain_wins
    enable_inbound: true
    enable_outbound: true
    sync_attachments: true
  
  retry:
    max_retries: 5
    initial_backoff: 1s
    max_backoff: 5m
    backoff_multiplier: 2.0
  
  webhook:
    enabled: true
    listen_addr: ":8480"
    path_prefix: "/webhooks"
    require_signature: true
    rate_limit_per_second: 100
  
  audit:
    enabled: true
    log_level: info
    retention_days: 90
```

## Credentials Management

### Secrets Storage

**CRITICAL: Never store credentials in configuration files or environment variables in plain text on disk.**

Recommended approaches:

1. **HashiCorp Vault** (preferred)
   ```bash
   vault kv put secret/virtengine/servicedesk \
     jira_api_token=<token> \
     jira_webhook_secret=<secret> \
     waldur_token=<token> \
     waldur_webhook_secret=<secret>
   ```

2. **Kubernetes Secrets**
   ```yaml
   apiVersion: v1
   kind: Secret
   metadata:
     name: virtengine-servicedesk-secrets
   type: Opaque
   data:
     jira-api-token: <base64>
     jira-webhook-secret: <base64>
     waldur-token: <base64>
     waldur-webhook-secret: <base64>
   ```

3. **AWS Secrets Manager / Azure Key Vault / GCP Secret Manager**

### Rotating Credentials

1. Generate new API tokens in Jira/Waldur
2. Update secrets in your secrets manager
3. Restart the service desk bridge service
4. Verify connectivity with health check endpoint
5. Revoke old tokens

## Operational Flow

### Outbound Sync (On-Chain → External)

1. User creates support ticket on-chain
2. `ticket_created` event is emitted
3. Bridge receives event and queues for sync
4. Bridge creates ticket in Jira/Waldur with mapped fields
5. External ticket reference is stored in sync record
6. Audit entry is logged

### Inbound Sync (External → On-Chain)

1. Agent updates ticket in Jira/Waldur
2. Webhook is sent to callback server
3. Signature is verified
4. Nonce is checked for replay prevention
5. Conflict detection runs
6. If no conflict, callback payload is processed
7. On-chain transaction is created (if enabled)
8. Audit entry is logged

### Conflict Resolution

| Strategy | Behavior |
|----------|----------|
| `on_chain_wins` | On-chain state is authoritative |
| `external_wins` | External state is authoritative |
| `newest_wins` | Most recent update wins |
| `manual` | Requires manual resolution |

## Monitoring

### Health Check Endpoint

```bash
curl http://localhost:8480/webhooks/health
```

Response:
```json
{
  "healthy": true,
  "jira_status": "healthy",
  "waldur_status": "healthy",
  "last_sync": "2024-01-15T10:30:00Z",
  "pending_events": 0,
  "failed_events": 0
}
```

### Metrics

The service exports Prometheus metrics:

- `servicedesk_sync_events_total{direction,status}` - Total sync events
- `servicedesk_sync_latency_seconds{direction,service}` - Sync latency
- `servicedesk_api_requests_total{service,method,status}` - API requests
- `servicedesk_queue_pending` - Pending events in queue
- `servicedesk_queue_failed` - Failed events

### Alerting

Recommended alerts:

```yaml
groups:
  - name: servicedesk
    rules:
      - alert: ServiceDeskSyncFailed
        expr: increase(servicedesk_sync_events_total{status="failed"}[5m]) > 5
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Service desk sync failures detected"
          
      - alert: ServiceDeskQueueBacklog
        expr: servicedesk_queue_pending > 100
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Service desk sync queue backlog"
          
      - alert: ServiceDeskUnhealthy
        expr: servicedesk_health_check != 1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Service desk bridge unhealthy"
```

## Webhook Setup

### Jira Webhook Configuration

1. Navigate to Jira Settings → System → Webhooks
2. Create new webhook with URL: `https://your-domain.com/webhooks/jira`
3. Set secret for signature verification
4. Select events:
   - Issue: created, updated, deleted
   - Comment: created, updated, deleted
5. Save and test

### Waldur Webhook Configuration

1. Navigate to Waldur admin panel
2. Configure webhook endpoint: `https://your-domain.com/webhooks/waldur`
3. Set secret for signature verification
4. Enable relevant event types
5. Save and verify

## Troubleshooting

### Common Issues

#### Sync Delays

- Check network connectivity to external APIs
- Verify rate limits are not being hit
- Check retry queue for backed-up events
- Review audit logs for errors

#### Authentication Failures

- Verify API tokens are valid and not expired
- Check username/email matches token owner
- Ensure service account has required permissions

#### Webhook Not Receiving

- Verify webhook URL is publicly accessible
- Check firewall rules allow incoming connections
- Verify signature configuration matches
- Check IP allowlist if configured

#### Conflict Errors

- Review audit log for conflict details
- Check sync timing and conflict resolution strategy
- Consider adjusting sync interval

### Log Analysis

```bash
# View sync errors
grep "sync failed" /var/log/virtengine/servicedesk.log

# View webhook events
grep "external callback" /var/log/virtengine/servicedesk.log

# View audit events
grep "audit" /var/log/virtengine/servicedesk.log | jq
```

### Manual Sync

To manually trigger sync for a specific ticket:

```bash
virtengine tx support sync-external <ticket-id> --direction=outbound
```

## Security Considerations

1. **Never log sensitive data** - API tokens, webhook secrets, and ticket content are never logged
2. **Signature verification** - All webhooks should require signature verification
3. **Nonce replay prevention** - Callback nonces prevent replay attacks
4. **TLS only** - All external API calls use HTTPS
5. **IP allowlisting** - Consider restricting webhook sources by IP
6. **Audit logging** - All external actions are logged for compliance

## Disaster Recovery

### Backup Considerations

- Sync records are stored on-chain (immutable)
- Audit logs should be backed up separately
- External ticket references are stored on-chain

### Recovery Procedure

1. If bridge service fails, events are queued for retry
2. On restart, pending events are processed
3. Manual sync can reconcile any missed updates
4. Audit log provides recovery information

## Capacity Planning

| Metric | Recommended Limit |
|--------|-------------------|
| Tickets per day | 10,000 |
| Sync interval | 30s minimum |
| Batch size | 50-100 |
| Retry queue | 10,000 events |
| Webhook rate | 100 req/s |
