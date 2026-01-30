# VirtEngine ↔ Waldur Marketplace Mapping Specification

> **Version:** 1.0.0  
> **Status:** Draft  
> **Authors:** VirtEngine Team  
> **Last Updated:** 2026-01-30

## Table of Contents

1. [Overview](#overview)
2. [Field-Level Mappings](#field-level-mappings)
   - [Offering Mapping](#offering-mapping)
   - [Order Mapping](#order-mapping)
   - [Resource Mapping](#resource-mapping)
   - [Provider Mapping](#provider-mapping)
3. [Pricing Normalization](#pricing-normalization)
4. [Region and Resource Mappings](#region-and-resource-mappings)
5. [Lifecycle State Alignment](#lifecycle-state-alignment)
6. [Sync and Reconciliation Policies](#sync-and-reconciliation-policies)
7. [Ownership and VEID Gating Rules](#ownership-and-veid-gating-rules)
8. [API Rate Limits and Pagination](#api-rate-limits-and-pagination)
9. [Error Handling](#error-handling)
10. [Example Payloads](#example-payloads)
11. [Test Fixtures](#test-fixtures)
12. [Operator Guide](#operator-guide)

---

## Overview

This specification defines the canonical mapping between VirtEngine's on-chain marketplace module (`x/market`) and Waldur's marketplace API. The integration enables:

- **Export:** On-chain offerings/orders sync to Waldur for visibility and management
- **Import:** Waldur-initiated actions (provisioning, termination) execute on-chain
- **Bi-directional state reconciliation:** Both systems remain consistent

### Architecture

```
┌─────────────────────┐                    ┌─────────────────────┐
│   VirtEngine Chain  │                    │   Waldur Platform   │
│                     │                    │                     │
│  x/market module    │◄──── Sync ────────►│  Marketplace API    │
│  - Offerings        │     (Export)       │  - Offerings        │
│  - Orders           │                    │  - Orders           │
│  - Allocations      │◄──── Callbacks ────│  - Resources        │
│  - Bids             │     (Import)       │  - Projects         │
│                     │                    │                     │
│  pkg/waldur         │                    │                     │
│  - MarketplaceClient│─────────────────────►│                   │
│  - OpenStackClient  │                    │                     │
│  - AWSClient        │                    │                     │
│  - AzureClient      │                    │                     │
│  - SLURMClient      │                    │                     │
└─────────────────────┘                    └─────────────────────┘
```

### Key Principles

1. **Chain as source of truth:** On-chain state is authoritative for ownership, pricing, and identity
2. **Waldur as operational layer:** Waldur handles infrastructure provisioning and usage tracking
3. **Idempotent operations:** All sync operations are safe to retry
4. **Security boundaries:** Encrypted data never leaves the chain unencrypted

---

## Field-Level Mappings

### Offering Mapping

#### VirtEngine → Waldur (Export)

| VirtEngine Field (x/market) | Waldur Field | Transform | Notes |
|----------------------------|--------------|-----------|-------|
| `ID.String()` | `backend_id` | Direct | Format: `{provider_address}/{sequence}` |
| `Name` | `name` | Direct | Max 255 chars in Waldur |
| `Description` | `description` | Direct | Markdown supported |
| `Category` | `category_uuid` | Lookup | Map category string to Waldur category UUID |
| `State` | `state` | Map | See [Offering State Map](#offering-state-map) |
| `Pricing.Model` | `type` | Map | See [Pricing Model Map](#pricing-model-map) |
| `Pricing.BasePrice` | `unit_price` | Normalize | Convert to Waldur currency format |
| `Pricing.Currency` | `components[].price` | Embed | Include in pricing component |
| `Regions[]` | `locations[]` | Lookup | Map region codes to Waldur location UUIDs |
| `Tags[]` | `attributes.tags` | Direct | Array of strings |
| `Specifications` | `attributes` | Merge | Flatten to Waldur attributes |
| `IdentityRequirement.MinScore` | `attributes.min_identity_score` | Direct | Custom attribute |
| `RequireMFAForOrders` | `attributes.require_mfa` | Direct | Custom attribute |
| `MaxConcurrentOrders` | `attributes.max_concurrent_orders` | Direct | Custom attribute |
| `PublicMetadata` | `attributes` | Merge | Prefix keys with `ve_` |
| `CreatedAt` | `created` | ISO8601 | UTC timestamp |
| `UpdatedAt` | `modified` | ISO8601 | UTC timestamp |
| `EncryptedSecrets` | **EXCLUDED** | N/A | Never synced - remains on-chain only |

#### Waldur → VirtEngine (Import - Read-Only Reference)

| Waldur Field | VirtEngine Usage | Notes |
|--------------|------------------|-------|
| `uuid` | `WaldurSyncRecord.WaldurID` | Track for reconciliation |
| `shared` | Visibility check | Public vs. private offerings |
| `billable` | Pricing validation | Must match chain config |
| `customer_uuid` | Provider validation | Map to provider address |

#### Offering State Map

| VirtEngine `OfferingState` | Waldur `state` |
|---------------------------|----------------|
| `OfferingStateActive` (1) | `Active` |
| `OfferingStatePaused` (2) | `Paused` |
| `OfferingStateSuspended` (3) | `Archived` |
| `OfferingStateDeprecated` (4) | `Paused` |
| `OfferingStateTerminated` (5) | `Archived` |

### Order Mapping

#### VirtEngine → Waldur (Export)

| VirtEngine Field | Waldur Field | Transform | Notes |
|-----------------|--------------|-----------|-------|
| `ID.String()` | `backend_id` | Direct | Format: `{customer_address}/{sequence}` |
| `OfferingID.String()` | `offering` | Lookup | Resolve to Waldur offering UUID |
| `State` | `state` | Map | See [Order State Map](#order-state-map) |
| `Region` | `attributes.region` | Direct | Location preference |
| `RequestedQuantity` | `limits.instances` | Direct | Resource quantity |
| `MaxBidPrice` | `attributes.max_price` | Normalize | For reference only |
| `AcceptedPrice` | `cost` | Normalize | After allocation |
| `PublicMetadata` | `attributes` | Merge | Prefix keys with `ve_` |
| `AllocatedProviderAddress` | `attributes.provider` | Direct | Post-allocation |
| `CreatedAt` | `created` | ISO8601 | UTC timestamp |
| `UpdatedAt` | `modified` | ISO8601 | UTC timestamp |
| `EncryptedConfig` | **EXCLUDED** | N/A | Never synced |
| `GatingResult` | **EXCLUDED** | N/A | Internal validation data |

#### Order State Map

| VirtEngine `OrderState` | Waldur `state` |
|------------------------|----------------|
| `OrderStatePendingPayment` (1) | `pending-consumer` |
| `OrderStateOpen` (2) | `pending-provider` |
| `OrderStateMatched` (3) | `executing` |
| `OrderStateProvisioning` (4) | `executing` |
| `OrderStateActive` (5) | `done` |
| `OrderStateSuspended` (6) | `done` (with suspended flag) |
| `OrderStatePendingTermination` (7) | `terminating` |
| `OrderStateTerminated` (8) | `terminated` |
| `OrderStateFailed` (9) | `erred` |
| `OrderStateCancelled` (10) | `canceled` |

### Resource Mapping

Waldur Resources correspond to VirtEngine Allocations after provisioning.

#### VirtEngine Allocation → Waldur Resource

| VirtEngine Field | Waldur Field | Transform |
|-----------------|--------------|-----------|
| `ID.String()` | `backend_id` | Format: `{customer}/{orderSeq}/{allocSeq}` |
| `OfferingID` | `offering_uuid` | Lookup Waldur UUID |
| `State` | `state` | Map allocation states |
| `ProviderAddress` | `attributes.provider` | Direct |
| `ServiceEndpoints` | `report.endpoints` | Public endpoints only |
| `TotalUsage` | `usages[]` | Transform to Waldur usage format |
| `ProvisioningStatus.Phase` | `report.phase` | Current phase |
| `ProvisioningStatus.Progress` | `report.progress` | 0-100 |

#### Allocation State Map

| VirtEngine `AllocationState` | Waldur Resource `state` |
|-----------------------------|------------------------|
| `AllocationStatePending` (1) | `Creating` |
| `AllocationStateAccepted` (2) | `Creating` |
| `AllocationStateProvisioning` (3) | `Creating` |
| `AllocationStateActive` (4) | `OK` |
| `AllocationStateSuspended` (5) | `OK` (suspended flag) |
| `AllocationStateTerminating` (6) | `Terminating` |
| `AllocationStateTerminated` (7) | `Terminated` |
| `AllocationStateRejected` (8) | `Erred` |
| `AllocationStateFailed` (9) | `Erred` |

### Provider Mapping

Provider profiles sync to Waldur customers/service providers.

| VirtEngine Field | Waldur Field | Notes |
|-----------------|--------------|-------|
| Provider Address | `backend_id` | Unique identifier |
| Provider Name | `name` | Display name |
| Provider Description | `description` | Rich text |
| Identity Score | `abbreviation` / attribute | Visibility indicator |
| Domain | `domain` | Verified domain |
| Regions | `locations[]` | Operating regions |
| Capabilities | `attributes.capabilities` | Feature flags |

---

## Pricing Normalization

### Currency Conversion

VirtEngine stores prices in the smallest token denomination. Waldur expects decimal currency.

```go
// VirtEngine → Waldur price conversion
func normalizePrice(chainPrice uint64, decimals int) string {
    divisor := math.Pow10(decimals)
    return fmt.Sprintf("%.6f", float64(chainPrice)/divisor)
}

// Waldur → VirtEngine price conversion (for reference)
func denormalizePrice(waldurPrice string, decimals int) (uint64, error) {
    price, err := strconv.ParseFloat(waldurPrice, 64)
    if err != nil {
        return 0, err
    }
    multiplier := math.Pow10(decimals)
    return uint64(price * multiplier), nil
}
```

### Pricing Model Map

| VirtEngine `PricingModel` | Waldur `type` | Component Structure |
|--------------------------|---------------|---------------------|
| `hourly` | `Support.PerHour` | `billing_type: usage`, `measured_unit: hour` |
| `daily` | `Support.PerDay` | `billing_type: usage`, `measured_unit: day` |
| `monthly` | `Support.Monthly` | `billing_type: fixed`, `measured_unit: month` |
| `usage_based` | `Support.Usage` | Multiple components per metric |
| `fixed` | `Support.OneTime` | `billing_type: one`, `measured_unit: item` |

### Usage Rate Components

For `usage_based` pricing, create Waldur components for each usage rate:

```json
{
  "components": [
    {
      "type": "cpu",
      "name": "CPU Hours",
      "measured_unit": "cpu_hour",
      "billing_type": "usage",
      "limit_period": "month",
      "price": "0.05"
    },
    {
      "type": "ram", 
      "name": "Memory GB-Hours",
      "measured_unit": "gb_hour",
      "billing_type": "usage",
      "limit_period": "month",
      "price": "0.01"
    },
    {
      "type": "storage",
      "name": "Storage GB-Months",
      "measured_unit": "gb_month",
      "billing_type": "usage",
      "limit_period": "month",
      "price": "0.10"
    },
    {
      "type": "gpu",
      "name": "GPU Hours",
      "measured_unit": "gpu_hour",
      "billing_type": "usage",
      "limit_period": "month",
      "price": "1.00"
    }
  ]
}
```

---

## Region and Resource Mappings

### Region Code Mapping

VirtEngine uses standardized region codes that map to Waldur locations:

| VirtEngine Region | Waldur Location | Cloud Provider Mapping |
|------------------|-----------------|----------------------|
| `us-east-1` | `{location_uuid}` | AWS us-east-1, Azure eastus |
| `us-west-2` | `{location_uuid}` | AWS us-west-2, Azure westus2 |
| `eu-west-1` | `{location_uuid}` | AWS eu-west-1, Azure westeurope |
| `eu-central-1` | `{location_uuid}` | AWS eu-central-1, Azure germanywestcentral |
| `ap-northeast-1` | `{location_uuid}` | AWS ap-northeast-1, Azure japaneast |

### Resource Type Mapping

#### Compute Resources

| VirtEngine Category | Waldur Offering Type | Backend Mapping |
|--------------------|---------------------|-----------------|
| `compute` | `Marketplace.VirtualMachine` | OpenStack Instance, AWS EC2, Azure VM |
| `gpu` | `Marketplace.VirtualMachine` | With GPU flavor |
| `hpc` | `SLURM.Allocation` | SLURM cluster allocation |

#### Storage Resources

| VirtEngine Category | Waldur Offering Type | Backend Mapping |
|--------------------|---------------------|-----------------|
| `storage` | `Marketplace.Volume` | OpenStack Volume, AWS EBS |
| `storage` (object) | `Rancher.ObjectStorage` | S3-compatible storage |

#### Infrastructure-Specific Mappings

**OpenStack Instance:**
```
VirtEngine Specifications → OpenStack/Waldur
----------------------------------------
specifications.vcpu       → cores
specifications.memory_gb  → ram (convert to MiB)
specifications.disk_gb    → disk (convert to MiB)
specifications.image      → image_name
specifications.flavor     → flavor_name
```

**AWS Instance:**
```
VirtEngine Specifications → AWS/Waldur
----------------------------------------
specifications.vcpu       → cores
specifications.memory_gb  → ram (convert to MiB)
specifications.disk_gb    → disk (convert to MiB)
specifications.instance_type → (derive from cores/ram)
```

**Azure VM:**
```
VirtEngine Specifications → Azure/Waldur
----------------------------------------
specifications.vcpu       → cores
specifications.memory_gb  → ram (convert to MiB)
specifications.disk_gb    → disk (convert to MiB)
specifications.size       → size_name
specifications.location   → location_name
```

**SLURM Allocation:**
```
VirtEngine Specifications → SLURM/Waldur
----------------------------------------
specifications.cpu_limit  → cpu_limit
specifications.gpu_limit  → gpu_limit
specifications.ram_gb     → ram_limit (convert to MiB)
```

---

## Lifecycle State Alignment

### Offering Lifecycle

```
VirtEngine                           Waldur
─────────────────────────────────────────────────────────────
CreateOffering                       →  MarketplaceOfferingsCreate
  │                                        │
  ▼                                        ▼
OfferingStateActive ──────────────────── Active
  │                                        │
  ├─ UpdateOffering ──────────────────── MarketplaceOfferingsUpdate
  │                                        │
  ├─ PauseOffering                         │
  │   └─ OfferingStatePaused ─────────── Paused
  │       └─ ResumeOffering                │
  │           └─ OfferingStateActive ──── Active
  │                                        │
  ├─ SuspendOffering (governance)          │
  │   └─ OfferingStateSuspended ──────── Archived
  │                                        │
  ├─ DeprecateOffering                     │
  │   └─ OfferingStateDeprecated ─────── Paused (no new orders)
  │                                        │
  └─ TerminateOffering                     │
      └─ OfferingStateTerminated ─────── Archived
```

### Order Lifecycle

```
VirtEngine                           Waldur
─────────────────────────────────────────────────────────────
CreateOrder                          →  (queue for sync)
  │
  ▼
OrderStatePendingPayment             →  pending-consumer
  │
  ├─ Payment confirmed
  │
  ▼
OrderStateOpen                       →  pending-provider
  │                                        │
  ├─ Provider bids                         │
  │                                        │
  ▼                                        │
OrderStateMatched                    →  executing
  │                                        │
  ├─ Allocation created                    │
  │   └─ Waldur order created ←───────── MarketplaceOrdersCreate
  │                                        │
  ▼                                        ▼
OrderStateProvisioning               →  executing
  │                                        │
  ├─ Provider provisions                   │
  │   └─ Waldur: ApproveByProvider ←───── MarketplaceOrdersApproveByProvider
  │                                        │
  ▼                                        ▼
OrderStateActive                     →  done
  │                                        │
  ├─ Usage reporting                       │
  │   └─ Waldur callbacks ─────────────── (periodic)
  │                                        │
  ├─ Suspend/Resume                        │
  │   └─ OrderStateSuspended ─────────── done (suspended attribute)
  │                                        │
  └─ Termination request                   │
      └─ OrderStatePendingTermination ─── (via callback)
          │                                │
          └─ Waldur: TerminateResource ──── MarketplaceResourcesTerminate
              │                            │
              ▼                            ▼
OrderStateTerminated ──────────────────── terminated
```

### Allocation Lifecycle

```
VirtEngine                           Waldur Resource
─────────────────────────────────────────────────────────────
CreateAllocation                     →  (order creates resource)
  │
  ▼
AllocationStatePending               →  Creating
  │
  ├─ Provider accepts
  │
  ▼
AllocationStateAccepted              →  Creating
  │
  ├─ Provisioning starts
  │
  ▼
AllocationStateProvisioning          →  Creating
  │                                        │
  ├─ Backend provisioning                  │
  │   └─ SetBackendID ←──────────────────── MarketplaceOrdersSetBackendId
  │                                        │
  ▼                                        ▼
AllocationStateActive                →  OK
  │                                        │
  ├─ Usage metering                        │
  │   └─ UsageReport ←───────────────────── (periodic callback)
  │                                        │
  ├─ Suspend (payment/policy)              │
  │   └─ AllocationStateSuspended ──────── OK (attributes.suspended=true)
  │                                        │
  └─ Terminate                             │
      └─ AllocationStateTerminating ────── Terminating
          │                                │
          ▼                                ▼
AllocationStateTerminated ─────────────── Terminated
```

### State Transition Matrix

#### Allowed Order Transitions

| From State | Allowed Next States |
|------------|-------------------|
| `PendingPayment` | `Open`, `Cancelled` |
| `Open` | `Matched`, `Cancelled` |
| `Matched` | `Provisioning`, `Failed`, `Cancelled` |
| `Provisioning` | `Active`, `Failed` |
| `Active` | `Suspended`, `PendingTermination` |
| `Suspended` | `Active`, `PendingTermination` |
| `PendingTermination` | `Terminated`, `Failed` |

---

## Sync and Reconciliation Policies

### Sync Strategy

#### Push-Based Sync (Chain → Waldur)

```
On-chain events trigger sync:
1. BeginBlocker checks for pending sync records
2. Group by entity type (offering, order, allocation)
3. Batch API calls (max 10 per batch)
4. Update sync records with results
5. Retry failed syncs with exponential backoff
```

#### Pull-Based Reconciliation (Waldur → Chain)

```
Periodic reconciliation (every 5 minutes):
1. List Waldur resources with VirtEngine backend_ids
2. Compare state/version with on-chain records
3. Identify conflicts:
   - Waldur state != expected from chain state
   - Missing Waldur resources for active allocations
   - Orphaned Waldur resources (no chain record)
4. Generate reconciliation events for review
```

### Sync Record Schema

```go
type WaldurSyncRecord struct {
    EntityType        WaldurSyncType  // offering, order, allocation, provider
    EntityID          string          // On-chain entity ID
    WaldurID          string          // Waldur UUID (after first sync)
    State             WaldurSyncState // pending, synced, failed, out_of_sync
    SyncVersion       uint64          // Last synced version
    ChainVersion      uint64          // Current chain version
    LastSyncedAt      *time.Time      // Last successful sync
    LastSyncAttemptAt *time.Time      // Last sync attempt
    FailureCount      uint32          // Consecutive failures
    LastError         string          // Last error message
    Checksum          string          // Data checksum for drift detection
}
```

### Retry Policy

| Failure Count | Retry Delay | Action |
|--------------|-------------|--------|
| 1 | 30 seconds | Immediate retry |
| 2 | 1 minute | Retry with backoff |
| 3 | 5 minutes | Retry with backoff |
| 4 | 15 minutes | Retry with backoff |
| 5+ | 1 hour | Mark for manual review |

### Conflict Resolution

| Conflict Type | Resolution Strategy |
|--------------|-------------------|
| State mismatch | Chain state wins; force-update Waldur |
| Missing in Waldur | Recreate if active; ignore if terminated |
| Missing on chain | Mark orphaned; operator must resolve |
| Price mismatch | Chain price wins; update Waldur |
| Data drift | Checksum comparison; full resync if needed |

### Idempotency Guarantees

All sync operations use:
1. **Entity versioning:** `ChainVersion` increments on each change
2. **Checksums:** SHA256 of exported data for drift detection
3. **Nonce tracking:** Prevent duplicate callback processing
4. **Transaction IDs:** Link Waldur operations to chain transactions

---

## Ownership and VEID Gating Rules

### Ownership Model

```
Provider Ownership:
- Provider address owns offerings
- Provider must have valid VEID score ≥ platform minimum
- Provider domain verification required for high-value offerings

Customer Ownership:
- Customer address owns orders and allocations
- Customer VEID gating checked at order creation
- Escrow binds customer funds to order

Waldur Mapping:
- Provider address → Waldur Customer/ServiceProvider
- Customer address → Waldur Project member
- Offering → Waldur Offering (owned by provider customer)
- Order → Waldur Order (placed by customer project)
```

### VEID Gating for Waldur Imports

When Waldur initiates actions (callbacks), VEID gating is enforced:

#### Callback Signature Verification

```go
type WaldurCallback struct {
    ID              string           // Unique callback ID
    ActionType      WaldurActionType // provision, terminate, etc.
    WaldurID        string           // Waldur entity UUID
    ChainEntityID   string           // On-chain entity ID
    Signature       []byte           // Ed25519 signature
    SignerID        string           // Authorized signer key ID
    Nonce           string           // Replay protection
    Timestamp       time.Time        // Creation time
    ExpiresAt       time.Time        // Expiry (1 hour default)
}
```

#### Gating Checks for Callbacks

| Action Type | Required Checks |
|-------------|----------------|
| `provision` | Provider identity verified, allocation exists, escrow funded |
| `terminate` | Valid allocation, authorized signer, not already terminated |
| `usage_report` | Active allocation, authorized provider, valid metrics |
| `status_update` | Valid entity, authorized signer, state transition allowed |

### Identity Score Requirements

| Offering Category | Minimum Provider Score | Minimum Customer Score |
|------------------|----------------------|----------------------|
| `compute` | 50 | 0 (configurable) |
| `storage` | 50 | 0 (configurable) |
| `gpu` | 70 | 30 |
| `hpc` | 70 | 30 |
| `ml` | 70 | 50 |

### MFA Gating

MFA is required when:
1. Offering has `RequireMFAForOrders = true`
2. Provider has `RequireMFAForAll = true`
3. Order value exceeds platform threshold
4. Customer is accessing encrypted configuration

---

## API Rate Limits and Pagination

### Waldur API Rate Limits

Default configuration in `pkg/waldur/client.go`:

```go
type Config struct {
    Timeout            time.Duration // 30 seconds
    MaxRetries         int           // 3
    RetryWaitMin       time.Duration // 1 second
    RetryWaitMax       time.Duration // 30 seconds
    RateLimitPerSecond int           // 10 requests/second
    UserAgent          string        // "VirtEngine-Provider-Daemon/1.0"
}
```

### Rate Limiting Strategy

1. **Token bucket algorithm:** 10 requests/second baseline
2. **Backoff on 429:** Exponential backoff up to 30 seconds
3. **Burst capacity:** Up to 10 requests can be made immediately
4. **Per-endpoint limits:** Some Waldur endpoints may have lower limits

### Pagination Strategy

#### List Operations

```go
type ListParams struct {
    Page     int // 1-indexed
    PageSize int // Default 10, max 100
}
```

#### Cursor-Based Pagination (Preferred)

For large datasets, use cursor-based pagination:

```go
func listAllOfferings(ctx context.Context, client *MarketplaceClient) ([]Offering, error) {
    var all []Offering
    page := 1
    pageSize := 100
    
    for {
        offerings, err := client.ListOfferings(ctx, ListOfferingsParams{
            Page:     page,
            PageSize: pageSize,
        })
        if err != nil {
            return nil, err
        }
        
        all = append(all, offerings...)
        
        if len(offerings) < pageSize {
            break // Last page
        }
        page++
    }
    
    return all, nil
}
```

#### Sync Batch Sizes

| Entity Type | Batch Size | Rationale |
|-------------|------------|-----------|
| Offerings | 50 | Moderate payload, infrequent changes |
| Orders | 100 | Smaller payload, frequent changes |
| Resources | 100 | Smaller payload, status updates |
| Usage reports | 500 | Small payload, high volume |

---

## Error Handling

### Error Categories

| Category | HTTP Codes | Retry | Action |
|----------|-----------|-------|--------|
| Authentication | 401 | No | Refresh token, alert operator |
| Authorization | 403 | No | Log, alert operator |
| Not Found | 404 | Conditional | Create if new, alert if existing |
| Conflict | 409 | No | Reconcile state, retry with fresh data |
| Rate Limited | 429 | Yes | Exponential backoff |
| Server Error | 5xx | Yes | Retry with backoff |
| Network Error | N/A | Yes | Retry with backoff |

### Error Mapping

```go
var (
    ErrNotConfigured   = errors.New("waldur client not configured")
    ErrInvalidToken    = errors.New("invalid API token")
    ErrUnauthorized    = errors.New("unauthorized: check API token")
    ErrForbidden       = errors.New("forbidden: insufficient permissions")
    ErrNotFound        = errors.New("resource not found")
    ErrConflict        = errors.New("resource conflict")
    ErrRateLimited     = errors.New("rate limited")
    ErrServerError     = errors.New("waldur server error")
    ErrTimeout         = errors.New("request timeout")
    ErrInvalidResponse = errors.New("invalid response from waldur")
)
```

### Error Recovery Procedures

#### Token Expiry

```
1. Detect 401 response
2. Attempt token refresh (if refresh token available)
3. Retry original request
4. If refresh fails, mark sync as failed
5. Alert operator for manual token update
```

#### State Conflict

```
1. Detect 409 response or state mismatch
2. Fetch current Waldur state
3. Compare with chain state
4. Apply conflict resolution rules
5. Retry with corrected data
```

#### Orphaned Resources

```
1. Reconciliation detects Waldur resource without chain record
2. Mark as orphaned with timestamp
3. Wait grace period (24 hours)
4. Alert operator for resolution
5. Operator can: link to chain record, or delete
```

---

## Example Payloads

### Offering Export

**VirtEngine Offering:**
```json
{
  "id": {
    "provider_address": "ve1abc123xyz...",
    "sequence": 42
  },
  "state": 1,
  "category": "compute",
  "name": "Standard Compute Instance",
  "description": "General-purpose compute instance with balanced CPU and memory",
  "version": "1.2.0",
  "pricing": {
    "model": "hourly",
    "base_price": 50000,
    "currency": "uvirt",
    "usage_rates": {},
    "minimum_commitment": 3600
  },
  "identity_requirement": {
    "min_score": 30,
    "required_status": "",
    "require_verified_email": true,
    "require_verified_domain": false,
    "require_mfa": false
  },
  "require_mfa_for_orders": false,
  "public_metadata": {
    "sla": "99.9%",
    "support": "24x7"
  },
  "specifications": {
    "vcpu": "4",
    "memory_gb": "16",
    "disk_gb": "100",
    "network": "1Gbps"
  },
  "tags": ["compute", "linux", "general-purpose"],
  "regions": ["us-east-1", "eu-west-1"],
  "created_at": "2026-01-15T10:00:00Z",
  "updated_at": "2026-01-20T14:30:00Z",
  "max_concurrent_orders": 100,
  "total_order_count": 1500,
  "active_order_count": 45
}
```

**Waldur Offering Export:**
```json
{
  "name": "Standard Compute Instance",
  "description": "General-purpose compute instance with balanced CPU and memory",
  "type": "Support.PerHour",
  "category": "uuid-of-compute-category",
  "shared": true,
  "billable": true,
  "attributes": {
    "ve_offering_id": "ve1abc123xyz.../42",
    "ve_version": "1.2.0",
    "min_identity_score": 30,
    "require_mfa": false,
    "require_verified_email": true,
    "max_concurrent_orders": 100,
    "sla": "99.9%",
    "support": "24x7",
    "vcpu": "4",
    "memory_gb": "16",
    "disk_gb": "100",
    "network": "1Gbps",
    "tags": ["compute", "linux", "general-purpose"]
  },
  "components": [
    {
      "type": "usage",
      "name": "Hourly Rate",
      "measured_unit": "hour",
      "billing_type": "usage",
      "limit_period": "month",
      "price": "0.050000"
    }
  ],
  "plans": [
    {
      "name": "Standard",
      "unit": "hour",
      "unit_price": "0.050000"
    }
  ],
  "locations": ["uuid-us-east-1", "uuid-eu-west-1"]
}
```

### Order Export

**VirtEngine Order:**
```json
{
  "id": {
    "customer_address": "ve1customer789...",
    "sequence": 101
  },
  "offering_id": {
    "provider_address": "ve1abc123xyz...",
    "sequence": 42
  },
  "state": 5,
  "public_metadata": {
    "project_name": "ML Training Pipeline",
    "environment": "production"
  },
  "region": "us-east-1",
  "requested_quantity": 2,
  "allocated_provider_address": "ve1provider456...",
  "max_bid_price": 60000,
  "accepted_price": 50000,
  "created_at": "2026-01-25T09:00:00Z",
  "updated_at": "2026-01-25T09:15:00Z",
  "matched_at": "2026-01-25T09:05:00Z",
  "activated_at": "2026-01-25T09:15:00Z",
  "bid_count": 3
}
```

**Waldur Order:**
```json
{
  "offering": "uuid-of-offering",
  "project": "uuid-of-customer-project",
  "type": "Create",
  "attributes": {
    "name": "ML Training Pipeline",
    "description": "VirtEngine order ve1customer789.../101",
    "ve_order_id": "ve1customer789.../101",
    "region": "us-east-1",
    "environment": "production",
    "provider": "ve1provider456..."
  },
  "limits": {
    "instances": 2
  }
}
```

### Waldur Callback (Terminate)

**Waldur → VirtEngine:**
```json
{
  "id": "wcb_ve1customer789.../101/1_a1b2c3d4",
  "action_type": "terminate",
  "waldur_id": "uuid-of-waldur-resource",
  "chain_entity_type": "allocation",
  "chain_entity_id": "ve1customer789.../101/1",
  "payload": {
    "reason": "customer_request",
    "requested_by": "user@example.com"
  },
  "signature": "base64-encoded-ed25519-signature",
  "signer_id": "waldur-bridge-signer-01",
  "nonce": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6",
  "timestamp": "2026-01-30T12:00:00Z",
  "expires_at": "2026-01-30T13:00:00Z"
}
```

---

## Test Fixtures

### Unit Test Fixtures

#### Offering Fixtures

```go
// pkg/waldur/testdata/fixtures.go

var TestOffering = marketplace.Offering{
    ID: marketplace.OfferingID{
        ProviderAddress: "ve1testprovider123",
        Sequence:        1,
    },
    State:       marketplace.OfferingStateActive,
    Category:    marketplace.OfferingCategoryCompute,
    Name:        "Test Compute Offering",
    Description: "Test offering for unit tests",
    Version:     "1.0.0",
    Pricing: marketplace.PricingInfo{
        Model:     marketplace.PricingModelHourly,
        BasePrice: 10000,
        Currency:  "uvirt",
    },
    IdentityRequirement: marketplace.IdentityRequirement{
        MinScore:             30,
        RequireVerifiedEmail: true,
    },
    Regions: []string{"us-east-1"},
    Tags:    []string{"test", "compute"},
    Specifications: map[string]string{
        "vcpu":      "2",
        "memory_gb": "8",
    },
}

var TestOrder = marketplace.Order{
    ID: marketplace.OrderID{
        CustomerAddress: "ve1testcustomer456",
        Sequence:        1,
    },
    OfferingID:        TestOffering.ID,
    State:             marketplace.OrderStateActive,
    Region:            "us-east-1",
    RequestedQuantity: 1,
    MaxBidPrice:       15000,
    AcceptedPrice:     10000,
}

var TestAllocation = marketplace.Allocation{
    ID: marketplace.AllocationID{
        OrderID:  TestOrder.ID,
        Sequence: 1,
    },
    OfferingID:      TestOffering.ID,
    ProviderAddress: TestOffering.ID.ProviderAddress,
    State:           marketplace.AllocationStateActive,
    AcceptedPrice:   10000,
}
```

#### Waldur Response Fixtures

```go
var TestWaldurOffering = waldur.Offering{
    UUID:        "550e8400-e29b-41d4-a716-446655440000",
    Name:        "Test Compute Offering",
    Description: "Test offering for unit tests",
    Type:        "Support.PerHour",
    State:       "Active",
    Shared:      true,
    Billable:    true,
}

var TestWaldurOrder = waldur.Order{
    UUID:        "550e8400-e29b-41d4-a716-446655440001",
    State:       "done",
    Type:        "Create",
    ProjectUUID: "550e8400-e29b-41d4-a716-446655440002",
}

var TestWaldurResource = waldur.Resource{
    UUID:         "550e8400-e29b-41d4-a716-446655440003",
    Name:         "test-resource-1",
    State:        "OK",
    OfferingUUID: TestWaldurOffering.UUID,
    ProjectUUID:  TestWaldurOrder.ProjectUUID,
    ResourceType: "Support.PerHour",
}
```

### Integration Test Fixtures

```go
// tests/integration/waldur_bridge_test.go

func setupTestWaldurBridge(t *testing.T) (*WaldurBridge, func()) {
    // Create test Waldur client with mock server
    mockServer := httptest.NewServer(http.HandlerFunc(waldurMockHandler))
    
    cfg := waldur.Config{
        BaseURL:            mockServer.URL,
        Token:              "test-token",
        Timeout:            5 * time.Second,
        MaxRetries:         1,
        RateLimitPerSecond: 100,
    }
    
    client, err := waldur.NewClient(cfg)
    require.NoError(t, err)
    
    bridge := NewWaldurBridge(client, testConfig())
    
    cleanup := func() {
        mockServer.Close()
    }
    
    return bridge, cleanup
}

func waldurMockHandler(w http.ResponseWriter, r *http.Request) {
    switch {
    case strings.HasPrefix(r.URL.Path, "/api/marketplace-offerings"):
        handleOfferingsMock(w, r)
    case strings.HasPrefix(r.URL.Path, "/api/marketplace-orders"):
        handleOrdersMock(w, r)
    case strings.HasPrefix(r.URL.Path, "/api/marketplace-resources"):
        handleResourcesMock(w, r)
    default:
        http.NotFound(w, r)
    }
}
```

---

## Operator Guide

### Initial Setup

#### 1. Configure Waldur Connection

```yaml
# config/waldur-bridge.yaml
waldur:
  enabled: true
  endpoint: "https://waldur.example.com/api"
  token: "${WALDUR_API_TOKEN}"  # Set via environment variable
  
  sync:
    interval_seconds: 60
    batch_size: 50
    max_retries: 3
    retry_backoff_seconds: 30
  
  callbacks:
    expiry_seconds: 3600
    nonce_window_seconds: 7200
    authorized_signers:
      - "waldur-signer-key-1"
      - "waldur-signer-key-2"
  
  rate_limits:
    requests_per_second: 10
    burst_size: 20
```

#### 2. Set Up Waldur Categories

Create Waldur categories matching VirtEngine offerings:

| VirtEngine Category | Waldur Category Name | UUID |
|--------------------|--------------------|------|
| `compute` | VirtEngine Compute | (create and note UUID) |
| `storage` | VirtEngine Storage | (create and note UUID) |
| `gpu` | VirtEngine GPU | (create and note UUID) |
| `hpc` | VirtEngine HPC | (create and note UUID) |
| `ml` | VirtEngine ML | (create and note UUID) |

Update category mapping:
```yaml
waldur:
  category_mapping:
    compute: "uuid-for-compute-category"
    storage: "uuid-for-storage-category"
    gpu: "uuid-for-gpu-category"
    hpc: "uuid-for-hpc-category"
    ml: "uuid-for-ml-category"
```

#### 3. Set Up Location Mapping

```yaml
waldur:
  location_mapping:
    us-east-1: "uuid-for-us-east-1"
    us-west-2: "uuid-for-us-west-2"
    eu-west-1: "uuid-for-eu-west-1"
    eu-central-1: "uuid-for-eu-central-1"
    ap-northeast-1: "uuid-for-ap-northeast-1"
```

### Monitoring and Alerting

#### Key Metrics

| Metric | Description | Alert Threshold |
|--------|-------------|-----------------|
| `waldur_sync_pending_count` | Entities pending sync | > 100 for 5 min |
| `waldur_sync_failed_count` | Failed sync records | > 10 |
| `waldur_sync_latency_seconds` | Sync latency | > 60 seconds |
| `waldur_callback_queue_size` | Pending callbacks | > 50 |
| `waldur_api_error_rate` | API error rate | > 5% |
| `waldur_rate_limit_hits` | Rate limit events | > 10/min |

#### Log Queries

```
# Find sync failures
level=error component=waldur_bridge

# Find state conflicts
level=warn component=waldur_bridge message="state conflict"

# Find orphaned resources
level=warn component=waldur_bridge message="orphaned resource"
```

### Troubleshooting

#### Sync Stuck in Pending

```bash
# Check sync records
virtengine query market waldur-sync-records --state pending

# Force immediate sync
virtengine tx market force-waldur-sync --entity-id <entity-id>

# Check Waldur API connectivity
curl -H "Authorization: Token $WALDUR_TOKEN" \
  https://waldur.example.com/api/users/me/
```

#### State Mismatch

```bash
# Compare chain and Waldur state
virtengine query market offering <offering-id>
curl -H "Authorization: Token $WALDUR_TOKEN" \
  https://waldur.example.com/api/marketplace-offerings/<waldur-uuid>/

# Force reconciliation
virtengine tx market reconcile-waldur --entity-type offering --entity-id <id>
```

#### Callback Processing Failures

```bash
# Check callback queue
virtengine query market waldur-callbacks --state pending

# Retry failed callback
virtengine tx market retry-waldur-callback --callback-id <id>

# Check nonce status
virtengine query market waldur-nonces --nonce <nonce>
```

### Maintenance Procedures

#### Token Rotation

1. Generate new Waldur API token
2. Update configuration (environment variable or secret)
3. Restart bridge service
4. Verify connectivity: `virtengine query market waldur-health`
5. Revoke old token in Waldur

#### Category/Location Mapping Updates

1. Create new categories/locations in Waldur
2. Update configuration with new UUIDs
3. Run reconciliation for affected entities
4. Verify sync status

#### Emergency Procedures

**Disable Waldur Sync:**
```bash
# Temporary disable
virtengine tx market update-waldur-config --enabled=false

# Or via config
waldur:
  enabled: false
```

**Clear Sync Queue:**
```bash
# Mark all pending as failed (for investigation)
virtengine tx market clear-waldur-queue --mark-failed

# Or cancel pending syncs
virtengine tx market clear-waldur-queue --cancel
```

---

## Appendix

### Changelog

| Version | Date | Changes |
|---------|------|---------|
| 1.1.0 | 2026-01-30 | Added Waldur-to-Chain ingestion specification (VE-3D) |
| 1.0.0 | 2026-01-30 | Initial specification |

---

## VE-3D: Waldur-to-Chain Ingestion

This section documents the ingestion of Waldur offerings into the on-chain marketplace.

### Overview

The ingestion worker fetches offerings from Waldur's marketplace API and creates corresponding on-chain offerings, enabling decentralized visibility of Waldur-managed resources.

### Architecture

```
Waldur Platform                    VirtEngine Chain
+-----------------+                +------------------+
| Marketplace API |   Ingest       | x/market module  |
| - Offerings     |--------------->| - Offerings      |
+-----------------+                +------------------+
        |                                   ^
        v                                   |
+-----------------+                +------------------+
| pkg/waldur      |                | pkg/provider_    |
| MarketplaceClient|               | daemon/          |
|                 |--------------->| WaldurIngest     |
|                 |                | Worker           |
+-----------------+                +------------------+
```

### Ingestion Workflow

1. **Fetch**: Worker polls Waldur API with pagination and rate limiting
2. **Validate**: Each offering is validated for required fields and provider mapping
3. **Map**: Waldur fields are mapped to on-chain schema
4. **Normalize**: Pricing, regions, and specs are normalized
5. **Submit**: Offerings are created/updated on-chain via the OfferingSubmitter
6. **Track**: Sync state is persisted for drift detection

### Configuration

```yaml
waldur_ingest:
  enabled: true
  
  # Provider mapping
  provider_address: "virtengine1abc..."
  waldur_customer_uuid: "550e8400-e29b-41d4-a716-..."
  
  # Schedules
  ingest_interval_seconds: 3600   # Full ingest every hour
  reconcile_interval_seconds: 300 # Reconcile every 5 minutes
  reconcile_on_startup: true
  
  # Pagination and rate limiting
  page_size: 50
  rate_limit_per_second: 2.0
  rate_limit_burst: 5
  
  # Retry policy
  max_retries: 5
  retry_backoff_seconds: 30
  max_backoff_seconds: 3600
  
  # Filters
  only_active_offerings: true
  skip_shared_offerings: false
  skip_non_billable_offerings: false
  
  # Mapping configuration
  category_map:
    "waldur-cat-uuid-1": "compute"
    "waldur-cat-uuid-2": "storage"
  customer_provider_map:
    "waldur-customer-uuid": "virtengine1provider..."
```

### Field Mapping (Waldur → Chain)

| Waldur Field | Chain Field | Transform |
|--------------|-------------|-----------|
| `uuid` | (tracking only) | Stored in IngestSyncRecord |
| `name` | `Name` | Direct (max 255 chars) |
| `description` | `Description` | Direct |
| `type` | `Category` | Map via TypeMap config |
| `state` | `State` | Map: Active→Active, Paused→Paused, Archived→Terminated |
| `components[].price` | `Pricing.BasePrice` | Multiply by CurrencyDenominator |
| `attributes.regions` | `Regions` | Direct array copy |
| `attributes.tags` | `Tags` | Direct array copy |
| `attributes.spec_*` | `Specifications` | Strip prefix |
| `attributes.ve_min_identity_score` | `IdentityRequirement.MinScore` | Direct |
| `attributes.ve_require_mfa` | `RequireMFAForOrders` | Direct |
| `customer_uuid` | `ID.ProviderAddress` | Lookup via CustomerProviderMap |

### Provider Ownership Validation

Before ingestion, the worker validates:

1. **Provider Mapping**: Waldur `customer_uuid` must be mapped to a valid provider address
2. **Provider Registration**: Provider must be registered on-chain (if required)
3. **VEID Score**: Provider must meet minimum identity score (if configured)

```go
// Validation flow
validation := offering.Validate(cfg)
if !validation.Valid {
    state.MarkSkipped(waldurUUID, validation.Errors)
    continue
}
if validation.NeedsProviderRegistration {
    // Provider must register first
    continue
}
```

### State Tracking

The `WaldurIngestState` tracks:

- **Records**: Map of Waldur UUID → IngestRecord
- **Dead Letter Queue**: Offerings that failed max retries
- **Cursor**: Pagination state for resumable ingestion
- **Metrics**: Ingestion statistics

State is persisted to disk and survives restarts.

### Drift Detection

The reconciliation loop detects drift by:

1. Comparing Waldur checksum with stored checksum
2. Identifying offerings in `pending`, `out_of_sync`, or `retrying` states
3. Re-queuing affected offerings for ingestion

### Audit Logging

All operations are logged via `WaldurIngestAuditLogger`:

```json
{
  "timestamp": "2026-01-30T10:00:00Z",
  "waldur_uuid": "550e8400-...",
  "offering_name": "Basic Compute",
  "action": "create",
  "success": true,
  "duration_ns": 1500000000,
  "provider_address": "virtengine1..."
}
```

### Prometheus Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `waldur_ingests_total` | Counter | Total ingestion attempts |
| `waldur_ingests_successful` | Counter | Successful ingestions |
| `waldur_ingests_failed` | Counter | Failed ingestion attempts |
| `waldur_ingests_dead_lettered` | Counter | Dead-lettered offerings |
| `waldur_offerings_created` | Counter | Offerings created on-chain |
| `waldur_offerings_updated` | Counter | Offerings updated on-chain |
| `waldur_drift_detected` | Counter | Drift detections |
| `waldur_ingest_duration_seconds` | Histogram | Ingestion duration |

### Operational Commands

```bash
# Check ingestion status
virtengine query market waldur-ingest-status

# Force re-ingestion of specific offering
virtengine tx market force-waldur-ingest --waldur-uuid <uuid>

# Reprocess dead-lettered offerings
virtengine tx market reprocess-dead-letter --waldur-uuid <uuid>

# View ingestion metrics
curl http://localhost:9090/metrics | grep waldur_ingest
```

### Error Handling

| Error Type | Action | Retry |
|------------|--------|-------|
| Validation failed | Mark skipped | No |
| Provider not mapped | Mark skipped | No |
| Chain submission failed | Retry with backoff | Yes |
| Rate limited | Wait and retry | Yes |
| Network error | Retry with backoff | Yes |
| Max retries exceeded | Move to dead letter | No |

---

### References

- [Waldur API Documentation](https://docs.waldur.com/api/)
- [VirtEngine Market Module](../x/market/README.md)
- [VEID Identity System](../x/veid/README.md)
- [Encryption Envelope Format](../x/encryption/README.md)

### Glossary

| Term | Definition |
|------|------------|
| Offering | A service listed by a provider for customers to order |
| Order | A customer request to consume an offering |
| Allocation | The binding of an order to a specific provider after bid matching |
| Bid | A provider's offer to fulfill an order at a specific price |
| VEID | VirtEngine Identity - the on-chain identity verification system |
| Sync Record | Tracks synchronization state between chain and Waldur |
| Callback | An action initiated by Waldur that affects on-chain state |
| Ingestion | The process of importing Waldur offerings onto the chain |
