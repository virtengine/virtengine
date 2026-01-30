# Support Module

The `x/support` module provides an on-chain encrypted support request system for VirtEngine. It enables customers to create encrypted support tickets that can only be decrypted by authorized support staff.

## Overview

The support module provides:

- **Encrypted Tickets**: All ticket content is encrypted using multi-recipient envelopes from `x/encryption`
- **Role-Based Access Control**: Integrates with `x/roles` for SupportAgent and Administrator access
- **Ticket Lifecycle**: Create, assign, respond, resolve, close, and reopen tickets
- **Rate Limiting**: Prevents abuse with configurable limits per customer
- **Indexed Queries**: Efficient lookups by customer, provider, agent, or status

## Security Model

### Encryption

Ticket payloads are **never stored in plaintext** on-chain. The module uses `MultiRecipientEnvelope` from `x/encryption`:

```go
type MultiRecipientEnvelope struct {
    Version           uint32
    AlgorithmID       string           // "X25519-XSALSA20-POLY1305"
    PayloadCiphertext []byte           // Encrypted content
    PayloadNonce      []byte           // Unique nonce
    WrappedKeys       []WrappedKeyEntry // Per-recipient wrapped DEKs
    ClientSignature   []byte
    UserSignature     []byte
}
```

Recipients typically include:
- The ticket creator (customer)
- Assigned support agent(s)
- Support administrators

### Access Control

| Role | Can Create | Can View Own | Can View All | Can Assign | Can Respond | Can Resolve | Can Close |
|------|------------|--------------|--------------|------------|-------------|-------------|-----------|
| Customer | ✓ | ✓ | ✗ | ✗ | Own tickets | ✗ | Own tickets |
| SupportAgent | ✗ | ✗ | Assigned | ✗ | Assigned | Assigned | ✗ |
| Administrator | ✗ | ✗ | ✓ | ✓ | ✓ | ✓ | ✓ |

## Data Models

### SupportTicket

```go
type SupportTicket struct {
    TicketID         string                    // Unique identifier (TKT-00000001)
    CustomerAddress  string                    // Ticket creator
    ProviderAddress  string                    // Related provider (optional)
    ResourceRef      ResourceReference         // Related order/lease (optional)
    Status           TicketStatus              // Current status
    Priority         TicketPriority            // Low, Normal, High, Urgent
    Category         string                    // Issue category
    EncryptedPayload MultiRecipientEnvelope    // Encrypted content
    AssignedTo       string                    // Support agent address
    ResponseCount    uint32                    // Number of responses
    CreatedAt        time.Time
    UpdatedAt        time.Time
    ResolvedAt       *time.Time
    ClosedAt         *time.Time
    ResolutionRef    string                    // Resolution reference
}
```

### Ticket Status Flow

```
     ┌─────────┐
     │  Open   │◄────────────────────────────┐
     └────┬────┘                             │
          │ Assign                           │ Reopen
          ▼                                  │
    ┌──────────┐                             │
    │ Assigned │──────────────────┐          │
    └────┬─────┘                  │          │
         │ Start Work             │          │
         ▼                        │          │
   ┌───────────┐                  │          │
   │In Progress│◄─────────┐       │          │
   └─────┬─────┘          │       │          │
         │                │       │ Close    │
    ┌────┴────┐           │       │          │
    │         │           │       │          │
    ▼         ▼           │       │          │
┌─────────┐ ┌────────────┐│       │          │
│Pending  │ │  Resolved  │┼───────┼──────────┤
│Customer │ └─────┬──────┘│       │          │
└────┬────┘       │       │       ▼          │
     │            │       │  ┌─────────┐     │
     └────────────┴───────┴─►│ Closed  │─────┘
                             └─────────┘
```

### TicketPriority

| Priority | Description |
|----------|-------------|
| Low | Non-urgent issues |
| Normal | Default priority |
| High | Urgent issues |
| Urgent | Critical, immediate attention needed |

## Messages

### MsgCreateTicket

Creates a new support ticket.

```json
{
    "customer": "virtengine1...",
    "category": "technical",
    "priority": "normal",
    "provider_address": "virtengine1...",  // optional
    "resource_ref": {                       // optional
        "type": "lease",
        "id": "owner/dseq/gseq/oseq"
    },
    "encrypted_payload": { ... }
}
```

### MsgAssignTicket

Assigns a ticket to a support agent (admin only).

```json
{
    "sender": "virtengine1...",
    "ticket_id": "TKT-00000001",
    "assign_to": "virtengine1..."
}
```

### MsgRespondToTicket

Adds an encrypted response to a ticket.

```json
{
    "responder": "virtengine1...",
    "ticket_id": "TKT-00000001",
    "encrypted_payload": { ... }
}
```

### MsgResolveTicket

Marks a ticket as resolved (assigned agent or admin).

```json
{
    "sender": "virtengine1...",
    "ticket_id": "TKT-00000001",
    "resolution_ref": "Issue fixed by updating configuration"
}
```

### MsgCloseTicket

Closes a ticket (customer or admin).

```json
{
    "sender": "virtengine1...",
    "ticket_id": "TKT-00000001",
    "reason": "No longer needed"
}
```

### MsgReopenTicket

Reopens a closed ticket (within reopen window).

```json
{
    "sender": "virtengine1...",
    "ticket_id": "TKT-00000001",
    "reason": "Issue recurred"
}
```

## Queries

### Query Ticket

```bash
virtengine query support ticket TKT-00000001
```

### Query Tickets by Customer

```bash
virtengine query support tickets-by-customer virtengine1...
```

### Query Tickets by Agent

```bash
virtengine query support tickets-by-agent virtengine1...
```

### Query Open Tickets

```bash
virtengine query support open-tickets --priority high
```

### Query Ticket Responses

```bash
virtengine query support responses TKT-00000001
```

## Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `max_tickets_per_customer_per_day` | Daily ticket limit per customer | 5 |
| `max_responses_per_ticket` | Maximum responses per ticket | 100 |
| `ticket_cooldown_seconds` | Minimum seconds between tickets | 60 |
| `auto_close_after_days` | Days after resolution before auto-close | 7 |
| `max_open_tickets_per_customer` | Maximum open tickets per customer | 10 |
| `reopen_window_days` | Days after close when reopen is allowed | 30 |
| `allowed_categories` | Valid ticket categories | billing, technical, account, provider, deployment, security, other |

## Events

### EventTicketCreated

Emitted when a new ticket is created.

```json
{
    "ticket_id": "TKT-00000001",
    "customer": "virtengine1...",
    "provider": "virtengine1...",
    "category": "technical",
    "priority": "normal",
    "block_height": 12345,
    "timestamp": 1706620800
}
```

### EventTicketAssigned

Emitted when a ticket is assigned.

```json
{
    "ticket_id": "TKT-00000001",
    "assigned_to": "virtengine1...",
    "assigned_by": "virtengine1...",
    "block_height": 12346,
    "timestamp": 1706620860
}
```

### EventTicketResolved

Emitted when a ticket is resolved.

```json
{
    "ticket_id": "TKT-00000001",
    "resolved_by": "virtengine1...",
    "resolution": "issue-fixed",
    "block_height": 12350,
    "timestamp": 1706621100
}
```

### EventTicketClosed

Emitted when a ticket is closed.

```json
{
    "ticket_id": "TKT-00000001",
    "closed_by": "virtengine1...",
    "reason": "resolved",
    "block_height": 12360,
    "timestamp": 1706621700
}
```

## Integration

### With x/roles

The support module requires the roles keeper for authorization:

```go
func (k *Keeper) SetRolesKeeper(rolesKeeper RolesKeeper) {
    k.rolesKeeper = rolesKeeper
}
```

Required roles:
- `RoleSupportAgent` - Can view/respond to assigned tickets
- `RoleAdministrator` - Full access to all tickets

### With x/encryption

Ticket payloads must be valid `MultiRecipientEnvelope` structures. The encryption must include appropriate recipients based on who needs access.

## Audit Trail

All ticket lifecycle events are emitted as typed events for audit purposes:

- Ticket creation, assignment, response, resolution, and closure
- All events include block height and timestamp
- Events can be indexed for compliance and analysis

## Example Workflow

1. **Customer creates ticket**:
   - Encrypts ticket content with their key + support admin keys
   - Submits `MsgCreateTicket`
   - Event `ticket_created` emitted

2. **Admin assigns to agent**:
   - Submits `MsgAssignTicket`
   - Agent added as recipient (re-encryption needed off-chain)
   - Event `ticket_assigned` emitted

3. **Agent responds**:
   - Encrypts response with customer + admin keys
   - Submits `MsgRespondToTicket`
   - Event `ticket_responded` emitted

4. **Agent resolves**:
   - Submits `MsgResolveTicket`
   - Event `ticket_resolved` emitted

5. **Customer closes**:
   - Submits `MsgCloseTicket`
   - Event `ticket_closed` emitted
