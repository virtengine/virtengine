# VirtEngine Support Ticket Schema Specification

**Version:** 1.0.0  
**Date:** 2026-01-30  
**Status:** Draft  
**Schema:** `docs/support/schemas/support_ticket.schema.json`

---

## Table of Contents

1. [Overview](#overview)
2. [Ticket States and Transitions](#ticket-states-and-transitions)
3. [Role Permissions](#role-permissions)
4. [Payload Schema](#payload-schema)
5. [Encryption Policy](#encryption-policy)
6. [Retention and PII Policy](#retention-and-pii-policy)
7. [Audit Logging Requirements](#audit-logging-requirements)
8. [Access Control Mapping](#access-control-mapping)
9. [Examples](#examples)
10. [Validation](#validation)

---

## Overview

This document defines the data model for VirtEngine support tickets, including:

- **Ticket lifecycle** with defined states and transitions
- **Payload fields** for orders, providers, timestamps, and attachments
- **Encryption envelope policy** for PII and sensitive data
- **Retention/deletion rules** by jurisdiction (GDPR, CCPA, HIPAA)
- **Audit logging** for compliance and security
- **Access control** integration with `x/roles` module

### Design Principles

1. **Sensitive data is encrypted at rest** using the `x/encryption` envelope format
2. **PII follows data minimization** - only collect what's necessary
3. **Retention is jurisdiction-aware** with automatic deletion
4. **All access is logged** for audit and compliance
5. **Role-based access control** via `x/roles` module

### Classification Level

Support tickets are classified as **C2 (Confidential)** to **C3 (Restricted)** depending on content:

| Field Type | Classification | Encrypted |
|------------|---------------|-----------|
| Ticket ID, status, timestamps | C0 (Public) | No |
| Order/Provider references | C2 (Confidential) | No |
| Message content | C2-C3 (Confidential-Restricted) | Yes |
| Attachments (screenshots, logs) | C3 (Restricted) | Yes |
| PII (email, phone, names) | C3 (Restricted) | Yes |

---

## Ticket States and Transitions

### State Definitions

| State | Description | Terminal |
|-------|-------------|----------|
| `OPEN` | Newly created ticket, awaiting triage | No |
| `ASSIGNED` | Assigned to a support agent | No |
| `IN_PROGRESS` | Agent actively working on ticket | No |
| `AWAITING_CUSTOMER` | Waiting for customer response | No |
| `AWAITING_PROVIDER` | Waiting for provider response | No |
| `ESCALATED` | Escalated to senior support/admin | No |
| `ON_HOLD` | Temporarily paused (customer request) | No |
| `RESOLVED` | Issue resolved, pending confirmation | No |
| `CLOSED` | Ticket closed, no further action | Yes |
| `CANCELLED` | Ticket cancelled by customer | Yes |

### State Transition Matrix

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                        Support Ticket State Machine                           │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                               │
│   ┌──────┐                                                                    │
│   │ OPEN │ ─────────────┬─────────────────────────────────────────────────┐  │
│   └──┬───┘              │                                                 │  │
│      │                  │ cancel                                          │  │
│      │ assign           ▼                                                 │  │
│      │            ┌───────────┐                                           │  │
│      ▼            │ CANCELLED │ (terminal)                                │  │
│   ┌──────────┐    └───────────┘                                           │  │
│   │ ASSIGNED │                                                            │  │
│   └────┬─────┘                                                            │  │
│        │ start_work                                                       │  │
│        ▼                                                                  │  │
│   ┌─────────────┐                                                         │  │
│   │ IN_PROGRESS │◄────────────────────────────────────┐                   │  │
│   └──────┬──────┘                                     │                   │  │
│          │                                            │                   │  │
│     ┌────┼────┬────────────┬───────────────┐         │                   │  │
│     │    │    │            │               │         │                   │  │
│     ▼    │    ▼            ▼               ▼         │                   │  │
│ ┌────────┴─┐ ┌───────────────┐ ┌───────────────┐ ┌─────────┐            │  │
│ │AWAIT_CUST│ │ AWAIT_PROVIDER│ │   ESCALATED   │ │ ON_HOLD │            │  │
│ └────┬─────┘ └───────┬───────┘ └───────┬───────┘ └────┬────┘            │  │
│      │               │                 │              │                  │  │
│      │ respond       │ respond         │ resolve      │ resume           │  │
│      └───────────────┴─────────────────┴──────────────┴──────────────────┘  │
│                                        │                                     │
│                                        │ resolve                             │
│                                        ▼                                     │
│                                   ┌──────────┐                               │
│                                   │ RESOLVED │                               │
│                                   └────┬─────┘                               │
│                                        │                                     │
│                           ┌────────────┼────────────┐                        │
│                           │ confirm    │ reopen     │                        │
│                           ▼            │            ▼                        │
│                      ┌────────┐        │       ┌─────────────┐               │
│                      │ CLOSED │        │       │ IN_PROGRESS │               │
│                      └────────┘        │       └─────────────┘               │
│                      (terminal)        │                                     │
│                                        └──► back to IN_PROGRESS              │
│                                                                               │
└──────────────────────────────────────────────────────────────────────────────┘
```

### Allowed Transitions by Role

| Current State | Target State | Allowed Roles |
|---------------|--------------|---------------|
| `OPEN` | `ASSIGNED` | `support_agent`, `moderator`, `administrator` |
| `OPEN` | `CANCELLED` | `customer` (owner), `administrator` |
| `ASSIGNED` | `IN_PROGRESS` | `support_agent` (assigned), `moderator`, `administrator` |
| `IN_PROGRESS` | `AWAITING_CUSTOMER` | `support_agent`, `moderator`, `administrator` |
| `IN_PROGRESS` | `AWAITING_PROVIDER` | `support_agent`, `moderator`, `administrator` |
| `IN_PROGRESS` | `ESCALATED` | `support_agent`, `moderator`, `administrator` |
| `IN_PROGRESS` | `ON_HOLD` | `support_agent`, `moderator`, `administrator` |
| `IN_PROGRESS` | `RESOLVED` | `support_agent`, `moderator`, `administrator` |
| `AWAITING_CUSTOMER` | `IN_PROGRESS` | `customer` (owner), `support_agent`, `administrator` |
| `AWAITING_PROVIDER` | `IN_PROGRESS` | `service_provider`, `support_agent`, `administrator` |
| `ESCALATED` | `IN_PROGRESS` | `moderator`, `administrator` |
| `ESCALATED` | `RESOLVED` | `moderator`, `administrator` |
| `ON_HOLD` | `IN_PROGRESS` | `customer` (owner), `support_agent`, `administrator` |
| `RESOLVED` | `CLOSED` | `customer` (owner), `support_agent`, `administrator` |
| `RESOLVED` | `IN_PROGRESS` | `customer` (owner) |

### Automatic Transitions

| Condition | Transition | Timeout |
|-----------|------------|---------|
| `AWAITING_CUSTOMER` with no response | `RESOLVED` | 7 days |
| `RESOLVED` with no confirmation | `CLOSED` | 3 days |
| `OPEN` with no assignment | Alert | 24 hours |

---

## Role Permissions

### Ticket-Specific Permissions

| Permission | Description |
|------------|-------------|
| `ticket:create` | Create new support tickets |
| `ticket:read` | View ticket details (own or assigned) |
| `ticket:read_all` | View all tickets (administrative) |
| `ticket:update` | Update ticket fields |
| `ticket:assign` | Assign tickets to agents |
| `ticket:escalate` | Escalate tickets |
| `ticket:close` | Close/cancel tickets |
| `ticket:delete` | Delete tickets (admin only) |
| `ticket:export` | Export ticket data |
| `ticket:attachment:upload` | Upload attachments |
| `ticket:attachment:download` | Download attachments |
| `ticket:message:read` | Read ticket messages |
| `ticket:message:create` | Create ticket messages |
| `ticket:pii:view` | View decrypted PII fields |

### Role-Permission Matrix

| Role | Permissions |
|------|-------------|
| `customer` | `ticket:create`, `ticket:read` (own), `ticket:update` (own), `ticket:close` (own), `ticket:attachment:upload`, `ticket:attachment:download` (own), `ticket:message:read` (own), `ticket:message:create` (own) |
| `service_provider` | `ticket:read` (related), `ticket:message:read` (related), `ticket:message:create` (related) |
| `support_agent` | `ticket:read`, `ticket:update`, `ticket:assign`, `ticket:close`, `ticket:attachment:*`, `ticket:message:*`, `ticket:pii:view` |
| `moderator` | All `support_agent` permissions + `ticket:escalate`, `ticket:read_all` |
| `administrator` | All permissions including `ticket:delete`, `ticket:export` |
| `genesis_account` | All permissions |
| `validator` | None (tickets not validator-related) |

---

## Payload Schema

### SupportTicket Structure

```go
// SupportTicket represents a customer support ticket
type SupportTicket struct {
    // Identification
    TicketID    string `json:"ticket_id"`    // Unique identifier (UUID v7)
    Version     uint32 `json:"version"`      // Schema version (1)
    
    // Classification
    Category    TicketCategory `json:"category"`    // billing, technical, identity, marketplace, other
    Priority    TicketPriority `json:"priority"`    // low, medium, high, critical
    Tags        []string       `json:"tags,omitempty"`
    
    // Lifecycle
    Status      TicketStatus `json:"status"`      // Current state
    CreatedAt   int64        `json:"created_at"`  // Unix timestamp (seconds)
    UpdatedAt   int64        `json:"updated_at"`  // Unix timestamp (seconds)
    ClosedAt    int64        `json:"closed_at,omitempty"` // Unix timestamp when closed
    
    // Parties
    CustomerAddress   string `json:"customer_address"`   // VirtEngine address
    AssignedAgentAddr string `json:"assigned_agent_addr,omitempty"` // Support agent address
    
    // References
    OrderRef       *OrderReference    `json:"order_ref,omitempty"`
    ProviderRef    *ProviderReference `json:"provider_ref,omitempty"`
    DeploymentRef  *DeploymentRef     `json:"deployment_ref,omitempty"`
    RelatedTickets []string           `json:"related_tickets,omitempty"`
    
    // Content (encrypted)
    Subject         string                    `json:"subject"`
    DescriptionEnvelope *EncryptedPayloadEnvelope `json:"description_envelope"`
    
    // Messages (encrypted)
    Messages []TicketMessage `json:"messages,omitempty"`
    
    // Attachments (encrypted, stored off-chain)
    Attachments []AttachmentRef `json:"attachments,omitempty"`
    
    // Contact Info (encrypted)
    ContactEnvelope *EncryptedPayloadEnvelope `json:"contact_envelope,omitempty"`
    
    // Retention
    RetentionPolicy   RetentionPolicyRef `json:"retention_policy"`
    JurisdictionCode  string             `json:"jurisdiction_code"` // ISO 3166-1 alpha-2
    ConsentRecordID   string             `json:"consent_record_id,omitempty"`
    
    // Metadata
    Metadata map[string]string `json:"metadata,omitempty"`
}
```

### Reference Types

```go
// OrderReference links ticket to a marketplace order
type OrderReference struct {
    OrderID    string `json:"order_id"`
    OwnerAddr  string `json:"owner_addr"`
    DSeq       uint64 `json:"dseq"`       // Deployment sequence
    GSeq       uint32 `json:"gseq"`       // Group sequence
    OSeq       uint32 `json:"oseq"`       // Order sequence
}

// ProviderReference links ticket to a service provider
type ProviderReference struct {
    ProviderAddr string `json:"provider_addr"`
    ProviderName string `json:"provider_name,omitempty"`
    LeaseID      string `json:"lease_id,omitempty"`
}

// DeploymentRef links ticket to a deployment
type DeploymentRef struct {
    DeploymentID string `json:"deployment_id"`
    OwnerAddr    string `json:"owner_addr"`
    DSeq         uint64 `json:"dseq"`
}

// TicketMessage represents a message in the ticket thread
type TicketMessage struct {
    MessageID       string                    `json:"message_id"` // UUID v7
    SenderAddr      string                    `json:"sender_addr"`
    SenderRole      string                    `json:"sender_role"` // customer, support_agent, etc.
    CreatedAt       int64                     `json:"created_at"`
    ContentEnvelope *EncryptedPayloadEnvelope `json:"content_envelope"`
    IsInternal      bool                      `json:"is_internal"` // Internal notes (not visible to customer)
}

// AttachmentRef references an encrypted attachment stored off-chain
type AttachmentRef struct {
    AttachmentID    string `json:"attachment_id"`    // UUID v7
    FileName        string `json:"file_name"`        // Original filename
    ContentType     string `json:"content_type"`     // MIME type
    SizeBytes       int64  `json:"size_bytes"`       // File size
    StorageLocation string `json:"storage_location"` // IPFS CID or storage path
    UploadedAt      int64  `json:"uploaded_at"`
    UploadedBy      string `json:"uploaded_by"`      // Uploader address
    ChecksumSHA256  string `json:"checksum_sha256"`  // Integrity verification
    
    // Encryption envelope for the attachment encryption key
    KeyEnvelope *EncryptedPayloadEnvelope `json:"key_envelope"`
}

// RetentionPolicyRef references the applicable retention policy
type RetentionPolicyRef struct {
    PolicyID          string `json:"policy_id"`
    RetentionDays     int    `json:"retention_days"`
    DeleteAfterClosed bool   `json:"delete_after_closed"`
    ArchiveAfterDays  int    `json:"archive_after_days,omitempty"`
}
```

### Enumerations

```go
type TicketStatus string

const (
    TicketStatusOpen             TicketStatus = "OPEN"
    TicketStatusAssigned         TicketStatus = "ASSIGNED"
    TicketStatusInProgress       TicketStatus = "IN_PROGRESS"
    TicketStatusAwaitingCustomer TicketStatus = "AWAITING_CUSTOMER"
    TicketStatusAwaitingProvider TicketStatus = "AWAITING_PROVIDER"
    TicketStatusEscalated        TicketStatus = "ESCALATED"
    TicketStatusOnHold           TicketStatus = "ON_HOLD"
    TicketStatusResolved         TicketStatus = "RESOLVED"
    TicketStatusClosed           TicketStatus = "CLOSED"
    TicketStatusCancelled        TicketStatus = "CANCELLED"
)

type TicketCategory string

const (
    TicketCategoryBilling     TicketCategory = "billing"
    TicketCategoryTechnical   TicketCategory = "technical"
    TicketCategoryIdentity    TicketCategory = "identity"
    TicketCategoryMarketplace TicketCategory = "marketplace"
    TicketCategorySecurity    TicketCategory = "security"
    TicketCategoryOther       TicketCategory = "other"
)

type TicketPriority string

const (
    TicketPriorityLow      TicketPriority = "low"
    TicketPriorityMedium   TicketPriority = "medium"
    TicketPriorityHigh     TicketPriority = "high"
    TicketPriorityCritical TicketPriority = "critical"
)
```

### Timestamp Fields

All timestamps use Unix epoch seconds (int64) for deterministic consensus:

| Field | Description | Required |
|-------|-------------|----------|
| `created_at` | Ticket creation time | Yes |
| `updated_at` | Last modification time | Yes |
| `closed_at` | Time when ticket reached terminal state | No |

---

## Encryption Policy

### Encrypted Fields

The following fields contain sensitive data and **MUST** be encrypted:

| Field | Content Type | Recipients |
|-------|--------------|------------|
| `description_envelope` | Initial ticket description | Customer + Assigned Agent + Support Team |
| `contact_envelope` | PII (email, phone, etc.) | Customer + Assigned Agent |
| `messages[].content_envelope` | Message content | Customer + Participants |
| `attachments[].key_envelope` | Attachment encryption key | Customer + Assigned Agent |

### Envelope Format

All encrypted fields use the VirtEngine `EncryptedPayloadEnvelope` format from `x/encryption`:

```go
type EncryptedPayloadEnvelope struct {
    Version          uint32            `json:"version"`           // Envelope version (1)
    AlgorithmID      string            `json:"algorithm_id"`      // "X25519-XSALSA20-POLY1305"
    AlgorithmVersion uint32            `json:"algorithm_version"` // Algorithm version
    RecipientKeyIDs  []string          `json:"recipient_key_ids"` // Key fingerprints
    WrappedKeys      []WrappedKeyEntry `json:"wrapped_keys"`      // Per-recipient DEKs
    Nonce            []byte            `json:"nonce"`             // Unique per encryption
    Ciphertext       []byte            `json:"ciphertext"`        // Encrypted data
    SenderSignature  []byte            `json:"sender_signature"`  // Authenticity
    SenderPubKey     []byte            `json:"sender_pub_key"`    // Sender's public key
    Metadata         map[string]string `json:"metadata,omitempty"`
}
```

### Recipient Selection

| Ticket Type | Encryption Recipients |
|-------------|----------------------|
| New ticket | Customer + Support Team Key (role-based) |
| Assigned ticket | Customer + Assigned Agent |
| Provider-related | Customer + Assigned Agent + Provider (for relevant fields) |
| Escalated | Customer + Assigned Agent + Escalation Team Key |

### Support Team Key

A **role-based encryption key** allows any authorized support agent to decrypt tickets:

```go
// Support team uses a shared key registered to the support_agent role
// This allows ticket reassignment without re-encryption
type RoleEncryptionKey struct {
    RoleID        string `json:"role_id"`        // "support_agent"
    PublicKey     []byte `json:"public_key"`     // X25519 public key
    KeyFingerprint string `json:"key_fingerprint"`
    RegisteredAt  int64  `json:"registered_at"`
    // Private key is managed by HSM accessible to role holders
}
```

### Key Rotation Strategy

| Key Type | Rotation Frequency | Procedure |
|----------|-------------------|-----------|
| Customer keys | User-controlled | Standard x/encryption key rotation |
| Support team role key | Annual or on compromise | Re-encrypt active tickets, archive old key |
| Agent individual keys | Annual | Re-encrypt assigned tickets on agent departure |
| Attachment DEKs | Per-upload (unique) | No rotation needed |

#### Role Key Rotation Procedure

1. **Generate new role key** in HSM
2. **Register new key** with `x/encryption` module
3. **Re-encrypt active tickets** (status != CLOSED, CANCELLED)
   - Decrypt with old key
   - Re-encrypt with new key + existing recipients
4. **Archive old key** for historical ticket access (read-only)
5. **Update key ID** in role configuration

```bash
# Key rotation command (governance-gated)
virtengine tx support rotate-role-key \
    --role support_agent \
    --new-key-file /path/to/new-key.pub \
    --from admin
```

### Attachment Encryption

Attachments use a separate encryption layer:

1. **Generate unique DEK** (AES-256-GCM) per attachment
2. **Encrypt file** with DEK
3. **Store encrypted file** off-chain (IPFS/S3)
4. **Encrypt DEK** using ticket envelope format
5. **Store `key_envelope`** in attachment reference

```
┌─────────────────────────────────────────────────────────────────┐
│                    Attachment Encryption Flow                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌────────────┐    AES-256-GCM    ┌──────────────────────────┐ │
│   │ Attachment │ ───────────────► │ Encrypted File           │ │
│   │ (plaintext)│      DEK          │ (stored off-chain)       │ │
│   └────────────┘                   └──────────────────────────┘ │
│                                                                  │
│   ┌────────────┐   X25519 Envelope ┌──────────────────────────┐ │
│   │    DEK     │ ───────────────► │ key_envelope             │ │
│   │ (256-bit)  │    Recipients     │ (stored on-chain)        │ │
│   └────────────┘                   └──────────────────────────┘ │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Retention and PII Policy

### Retention by Jurisdiction

| Jurisdiction | Regulation | Default Retention | Max Retention | Deletion Trigger |
|--------------|------------|-------------------|---------------|------------------|
| EU (GDPR) | GDPR Art. 5(1)(e) | 2 years | 6 years (legal hold) | Ticket closed + retention |
| California (US) | CCPA | 3 years | 7 years (legal hold) | Ticket closed + retention |
| US (General) | N/A | 5 years | 7 years | Ticket closed + retention |
| UK | UK GDPR | 2 years | 6 years (legal hold) | Ticket closed + retention |
| Global Default | N/A | 3 years | 7 years | Ticket closed + retention |

### Retention Policy Types

```go
const (
    // RetentionPolicyGDPR - EU GDPR compliance (2 years)
    RetentionPolicyGDPR = "gdpr"
    
    // RetentionPolicyCCPA - California CCPA compliance (3 years)
    RetentionPolicyCCPA = "ccpa"
    
    // RetentionPolicyUSGeneral - US general (5 years)
    RetentionPolicyUSGeneral = "us_general"
    
    // RetentionPolicyUKGDPR - UK GDPR (2 years)
    RetentionPolicyUKGDPR = "uk_gdpr"
    
    // RetentionPolicyGlobal - Global default (3 years)
    RetentionPolicyGlobal = "global"
    
    // RetentionPolicyLegalHold - Indefinite (litigation)
    RetentionPolicyLegalHold = "legal_hold"
)
```

### Jurisdiction Mapping

| Country Code | Policy | Retention Days |
|--------------|--------|----------------|
| AT, BE, BG, HR, CY, CZ, DK, EE, FI, FR, DE, GR, HU, IE, IT, LV, LT, LU, MT, NL, PL, PT, RO, SK, SI, ES, SE | `gdpr` | 730 |
| GB | `uk_gdpr` | 730 |
| US-CA | `ccpa` | 1095 |
| US-* | `us_general` | 1825 |
| * | `global` | 1095 |

### PII Field Inventory

| Field | Classification | PII Type | Encrypted | Deletion on Request |
|-------|---------------|----------|-----------|---------------------|
| `contact_envelope` | C3 | Email, Phone | Yes | Yes |
| `description_envelope` | C2-C3 | Varies | Yes | Pseudonymized |
| `messages[].content_envelope` | C2-C3 | Varies | Yes | Pseudonymized |
| `attachments[]` | C3 | Screenshots, logs | Yes | Yes |
| `customer_address` | C2 | Blockchain address | No | No (pseudonymous) |

### Data Subject Rights

#### Right to Access (GDPR Art. 15)

```bash
# Export all ticket data for a customer
virtengine query support export-customer-data \
    --customer-address virtengine1abc... \
    --format json
```

#### Right to Erasure (GDPR Art. 17)

```bash
# Request data deletion (requires consent verification)
virtengine tx support request-deletion \
    --ticket-id abc123... \
    --reason "customer_request" \
    --from customer-key
```

**Deletion Workflow:**

1. **Verify identity** - Confirm requester owns the ticket
2. **Check legal holds** - Cannot delete if under legal hold
3. **Anonymize references** - Replace PII with pseudonyms
4. **Delete encrypted content** - Remove `description_envelope`, `contact_envelope`
5. **Delete attachments** - Remove from off-chain storage
6. **Retain metadata** - Keep anonymized ticket shell for audit

#### Right to Rectification (GDPR Art. 16)

```bash
# Update PII (re-encrypts with new data)
virtengine tx support update-contact-info \
    --ticket-id abc123... \
    --contact-file updated-contact.json \
    --from customer-key
```

### Automatic Deletion Process

```
┌─────────────────────────────────────────────────────────────────┐
│                    Retention Deletion Flow                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   1. Daily Cron Job                                             │
│   ┌──────────────────────────────────────────────────────────┐  │
│   │ SELECT tickets WHERE                                      │  │
│   │   status IN (CLOSED, CANCELLED) AND                       │  │
│   │   closed_at + retention_days < NOW() AND                  │  │
│   │   legal_hold = false                                      │  │
│   └─────────────────────────────────────────────────────────┘  │
│                          │                                      │
│                          ▼                                      │
│   2. For each eligible ticket:                                 │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │ a. Delete encrypted envelopes (description, contact)    │  │
│   │ b. Delete message content envelopes                     │  │
│   │ c. Delete attachments from off-chain storage            │  │
│   │ d. Delete attachment key envelopes                      │  │
│   │ e. Anonymize customer address → hash                    │  │
│   │ f. Set deletion_timestamp                               │  │
│   │ g. Emit TicketDataDeleted event                         │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                  │
│   3. Retained (for audit):                                      │
│   ┌─────────────────────────────────────────────────────────┐  │
│   │ - ticket_id                                              │  │
│   │ - category, priority, status                             │  │
│   │ - created_at, closed_at, deletion_timestamp              │  │
│   │ - anonymized_customer_hash                               │  │
│   │ - retention_policy reference                             │  │
│   └─────────────────────────────────────────────────────────┘  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Legal Hold

Tickets can be placed under legal hold to prevent deletion:

```bash
# Place legal hold (admin only)
virtengine tx support set-legal-hold \
    --ticket-id abc123... \
    --reason "Litigation case #12345" \
    --from admin-key
```

Legal hold prevents:
- Automatic retention deletion
- Customer deletion requests
- Attachment purging

---

## Audit Logging Requirements

### Audit Events

| Event Type | Description | Logged Fields |
|------------|-------------|---------------|
| `TICKET_CREATED` | New ticket created | ticket_id, customer_addr, category, priority |
| `TICKET_VIEWED` | Ticket details accessed | ticket_id, viewer_addr, viewer_role, fields_accessed |
| `TICKET_UPDATED` | Ticket fields modified | ticket_id, modifier_addr, changed_fields |
| `TICKET_STATUS_CHANGED` | State transition | ticket_id, old_status, new_status, changed_by |
| `TICKET_ASSIGNED` | Agent assignment | ticket_id, agent_addr, assigned_by |
| `TICKET_ESCALATED` | Escalation event | ticket_id, escalated_by, escalation_reason |
| `MESSAGE_CREATED` | New message added | ticket_id, message_id, sender_addr, is_internal |
| `MESSAGE_VIEWED` | Message decrypted/read | ticket_id, message_id, viewer_addr |
| `ATTACHMENT_UPLOADED` | File attached | ticket_id, attachment_id, uploader_addr, size_bytes |
| `ATTACHMENT_DOWNLOADED` | File downloaded | ticket_id, attachment_id, downloader_addr |
| `ATTACHMENT_DELETED` | File removed | ticket_id, attachment_id, deleted_by |
| `PII_ACCESSED` | Contact info decrypted | ticket_id, accessor_addr, accessor_role, pii_fields |
| `TICKET_EXPORTED` | Data export requested | ticket_id, exporter_addr, export_format |
| `TICKET_DELETED` | Data deletion (GDPR) | ticket_id, deletion_reason, deleted_by |
| `LEGAL_HOLD_SET` | Legal hold applied | ticket_id, set_by, reason |
| `LEGAL_HOLD_REMOVED` | Legal hold lifted | ticket_id, removed_by, reason |

### Audit Log Entry Structure

```go
type SupportAuditEntry struct {
    EntryID        string            `json:"entry_id"`        // UUID v7
    Timestamp      int64             `json:"timestamp"`       // Unix seconds
    EventType      string            `json:"event_type"`      // From enum above
    TicketID       string            `json:"ticket_id"`
    ActorAddress   string            `json:"actor_address"`   // Who performed action
    ActorRole      string            `json:"actor_role"`      // Role at time of action
    ClientIP       string            `json:"client_ip"`       // Anonymized after 30 days
    UserAgent      string            `json:"user_agent,omitempty"`
    SessionID      string            `json:"session_id,omitempty"`
    
    // Event-specific details
    Details        map[string]interface{} `json:"details"`
    
    // Integrity
    PreviousHash   string            `json:"previous_hash"`   // Hash chain
    EntryHash      string            `json:"entry_hash"`      // SHA256 of entry
}
```

### Audit Log Retention

| Log Type | Retention | Archive |
|----------|-----------|---------|
| Access logs (`*_VIEWED`, `*_ACCESSED`) | 2 years | 5 years (cold) |
| Modification logs (`*_CREATED`, `*_UPDATED`) | 5 years | 10 years (cold) |
| Deletion logs (`*_DELETED`) | 10 years | Indefinite |
| Security events (`LEGAL_HOLD_*`) | 10 years | Indefinite |

### Audit Query Examples

```bash
# View audit trail for a ticket
virtengine query support audit-log \
    --ticket-id abc123... \
    --from-timestamp 1704067200 \
    --limit 100

# View PII access for compliance report
virtengine query support audit-log \
    --event-type PII_ACCESSED \
    --actor-role support_agent \
    --from-timestamp 1704067200

# Verify audit log integrity
virtengine query support verify-audit-integrity \
    --from-entry entry123... \
    --to-entry entry456...
```

---

## Access Control Mapping

### Integration with x/roles

The support ticket system integrates with the `x/roles` module for access control:

```go
// From x/roles/types/roles.go
const (
    RoleCustomer       Role = 6  // End users (ticket creators)
    RoleSupportAgent   Role = 7  // Customer support
    RoleModerator      Role = 4  // Escalation handling
    RoleAdministrator  Role = 2  // Full access
    RoleServiceProvider Role = 5 // Provider-related tickets
)
```

### Permission Checks

```go
// Check if actor can perform action on ticket
func (k Keeper) CanAccessTicket(ctx sdk.Context, actorAddr, ticketID string, permission Permission) error {
    // Get actor's role
    role, err := k.rolesKeeper.GetRole(ctx, actorAddr)
    if err != nil {
        return ErrUnauthorized
    }
    
    // Get ticket
    ticket, found := k.GetTicket(ctx, ticketID)
    if !found {
        return ErrTicketNotFound
    }
    
    // Check ownership
    isOwner := ticket.CustomerAddress == actorAddr
    isAssigned := ticket.AssignedAgentAddr == actorAddr
    isRelatedProvider := k.isProviderRelated(ctx, ticket, actorAddr)
    
    switch permission {
    case PermissionTicketRead:
        if isOwner || isAssigned || isRelatedProvider {
            return nil
        }
        if role == RoleSupportAgent || role == RoleModerator || role == RoleAdministrator {
            return nil
        }
        
    case PermissionTicketUpdate:
        if isOwner && ticket.Status == TicketStatusOpen {
            return nil
        }
        if isAssigned || role == RoleModerator || role == RoleAdministrator {
            return nil
        }
        
    case PermissionPIIView:
        // Only assigned agent or higher can view PII
        if isAssigned || role == RoleModerator || role == RoleAdministrator {
            return nil
        }
        
    case PermissionTicketDelete:
        // Admin only
        if role == RoleAdministrator || role == RoleGenesisAccount {
            return nil
        }
    }
    
    return ErrUnauthorized
}
```

### Trust Level Requirements

| Operation | Minimum Trust Level | Roles |
|-----------|---------------------|-------|
| Create ticket | 50 | `customer`, `support_agent`, `administrator` |
| View own ticket | 50 | `customer` |
| View any ticket | 60 | `support_agent`, `moderator`, `administrator` |
| Assign ticket | 60 | `support_agent`, `moderator`, `administrator` |
| Escalate | 70 | `moderator`, `administrator` |
| View PII | 60 | `support_agent` (assigned), `moderator`, `administrator` |
| Delete ticket | 90 | `administrator` |
| Export all | 90 | `administrator` |

---

## Examples

### Example 1: Basic Support Ticket

See: `docs/support/examples/basic_ticket.json`

### Example 2: Marketplace Order Issue

See: `docs/support/examples/marketplace_ticket.json`

### Example 3: Identity Verification Issue

See: `docs/support/examples/identity_ticket.json`

---

## Validation

### JSON Schema

The JSON Schema for validation is located at:
`docs/support/schemas/support_ticket.schema.json`

### Validation Rules

1. **TicketID**: Must be valid UUID v7 format
2. **Status**: Must be valid enum value
3. **Transitions**: Must follow allowed transition matrix
4. **Timestamps**: Must be valid Unix timestamps, `updated_at >= created_at`
5. **Addresses**: Must be valid VirtEngine bech32 addresses
6. **Envelopes**: Must pass `x/encryption` envelope validation
7. **Attachments**: Size must not exceed 50MB, allowed types only
8. **Retention**: Must have valid retention policy for jurisdiction

### Test Coverage

Tests are located at: `tests/support_ticket_schema_test.go`

- Schema validation tests
- State transition tests
- Encryption envelope tests
- Retention policy tests
- Access control tests

---

## Changelog

- **2026-01-30**: Initial specification (v1.0.0)
  - Defined ticket states and transitions
  - Defined payload schema with encryption
  - Defined retention policies by jurisdiction
  - Defined audit logging requirements
  - Mapped to x/roles access control
