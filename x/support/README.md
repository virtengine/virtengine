# Support Module

The `x/support` module provides lightweight on-chain references to external support tickets managed by operator software (Waldur/Jira).

## Architecture Decision

**Support tickets are managed by operator software (Waldur with native Jira integration), not on-chain.**

This design decision was made because:

1. **Operator software is purpose-built for ticketing** - Waldur and Jira have mature workflows, SLA tracking, agent assignment, and reporting capabilities
2. **No consensus overhead** - Ticket updates don't require blockchain transactions, reducing gas costs
3. **Waldur already supports Jira** - Native integration eliminates duplicate implementations
4. **Separation of concerns** - Blockchain handles value transfer and state; operators handle support workflows

## Overview

The support module provides:

- **External Ticket References**: Store mapping from on-chain resources to external ticket IDs
- **Traceability**: Link deployments, leases, and orders to their support tickets
- **Event Forwarding**: Emit events that the provider daemon forwards to Waldur

## Data Models

### ExternalTicketRef

Stores a reference to an external support ticket:

```go
type ExternalTicketRef struct {
    ResourceID       string    // On-chain resource ID (deployment, lease, order)
    ResourceType     string    // Type of resource ("deployment", "lease", "order")
    ExternalSystem   string    // "waldur" or "jira"
    ExternalTicketID string    // External ticket ID (e.g., "JIRA-123", waldur UUID)
    ExternalURL      string    // URL to the external ticket
    CreatedAt        time.Time // When the reference was created
    CreatedBy        string    // Address that created the reference
}
```

### Supported External Systems

| System | Description |
|--------|-------------|
| `waldur` | Waldur service desk (recommended) |
| `jira` | Jira Service Desk (via Waldur integration) |

## Messages

### MsgRegisterExternalTicket

Registers an external ticket reference for traceability:

```json
{
    "sender": "virtengine1...",
    "resource_id": "owner/dseq/gseq/oseq",
    "resource_type": "lease",
    "external_system": "waldur",
    "external_ticket_id": "abc-123-def",
    "external_url": "https://waldur.example.com/support/abc-123-def"
}
```

### MsgUpdateExternalTicket

Updates an external ticket reference:

```json
{
    "sender": "virtengine1...",
    "resource_id": "owner/dseq/gseq/oseq",
    "resource_type": "lease",
    "external_ticket_id": "abc-123-def",
    "external_url": "https://waldur.example.com/support/abc-123-def"
}
```

## Queries

### Query External Ticket by Resource

```bash
virtengine query support external-ticket --resource-id="owner/dseq/gseq/oseq" --resource-type="lease"
```

### Query External Tickets by Owner

```bash
virtengine query support external-tickets-by-owner virtengine1...
```

## Events

### EventExternalTicketRegistered

Emitted when an external ticket reference is registered:

```json
{
    "resource_id": "owner/dseq/gseq/oseq",
    "resource_type": "lease",
    "external_system": "waldur",
    "external_ticket_id": "abc-123-def",
    "block_height": 12345,
    "timestamp": 1706620800
}
```

## Integration with Waldur

The provider daemon (`pkg/provider_daemon`) handles the actual support ticket lifecycle:

1. **Ticket Creation**: Provider daemon creates tickets in Waldur when deployment issues occur
2. **Status Updates**: Waldur webhooks notify the daemon of ticket status changes
3. **Reference Storage**: Daemon registers external ticket references on-chain for traceability
4. **SLA Tracking**: Handled entirely by Waldur/Jira

### Provider Daemon Flow

```
┌─────────────────┐     ┌──────────────────┐     ┌────────────────┐
│  VirtEngine     │     │  Provider Daemon │     │  Waldur        │
│  On-Chain       │◄────│  (Bridge)        │────►│  (Jira native) │
│  References     │     │                  │     │                │
└─────────────────┘     └──────────────────┘     └────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
  Store external         Forward events          Manage full
  ticket refs            to Waldur              ticket lifecycle
```

### Configuration

See `pkg/servicedesk` for the bridge configuration:

```yaml
servicedesk:
  waldur:
    base_url: "https://waldur.example.com/api"
    organization_uuid: "org-uuid"
    # token from secrets manager
```

## Security Considerations

1. **No sensitive data on-chain** - Only ticket references are stored, not ticket content
2. **Access control** - Only resource owners can register ticket references
3. **External URLs validated** - URLs must match configured Waldur/Jira domains

## Migration Notes

This module replaces the previous full on-chain ticketing system. The rationale:

- **Previous**: Full ticket lifecycle on-chain (encrypted payloads, responses, SLA tracking)
- **Current**: Lightweight references only; operator software handles everything else

Benefits:
- 90%+ reduction in on-chain storage for support
- No gas costs for ticket updates
- Leverage mature operator tooling (Waldur/Jira)
