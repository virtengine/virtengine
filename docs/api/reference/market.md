# Market Module API Reference

The Market module manages the decentralized marketplace for compute resources including orders, bids, and leases.

## Overview

The Market module enables:
- Order creation and management
- Bid submission and acceptance
- Lease lifecycle management
- Escrow integration for payments
- Provider-tenant matching

## Base URL

```
/virtengine/market/v2beta1
```

## Authentication Requirements

| Operation Type | Authentication |
|----------------|----------------|
| Query endpoints | None required |
| Order creation | Wallet signature |
| Bid operations | Provider certificate |
| Lease management | Wallet signature |

---

## Query Endpoints

### List Orders

Queries orders with filters.

```http
GET /virtengine/market/v2beta1/orders/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `filters.owner` | query | string | No | Filter by owner address |
| `filters.state` | query | string | No | Filter by state (`open`, `active`, `closed`) |
| `filters.dseq` | query | string | No | Filter by deployment sequence |
| `pagination.limit` | query | integer | No | Max results (default: 100) |
| `pagination.key` | query | string | No | Pagination key |

**Response:**

```json
{
  "orders": [
    {
      "order_id": {
        "owner": "virtengine1abc...",
        "dseq": "12345",
        "gseq": 1,
        "oseq": 1
      },
      "state": "open",
      "spec": {
        "name": "web-service",
        "requirements": {
          "cpu": { "units": 1000 },
          "memory": { "size": "512Mi" },
          "storage": [{ "size": "1Gi" }]
        }
      },
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "next_key": "base64...",
    "total": "100"
  }
}
```

**Example:**

```bash
# List all open orders
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.state=open"

# List orders by owner
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.owner=virtengine1abc..."
```

---

### Get Order

Queries details of a specific order.

```http
GET /virtengine/market/v2beta1/orders/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `id.owner` | query | string | Yes | Owner address |
| `id.dseq` | query | string | Yes | Deployment sequence |
| `id.gseq` | query | integer | Yes | Group sequence |
| `id.oseq` | query | integer | Yes | Order sequence |

**Response:**

```json
{
  "order": {
    "order_id": {
      "owner": "virtengine1abc...",
      "dseq": "12345",
      "gseq": 1,
      "oseq": 1
    },
    "state": "open",
    "spec": {
      "name": "web-service",
      "requirements": {
        "cpu": { "units": 1000 },
        "memory": { "size": "512Mi" },
        "storage": [{ "size": "1Gi" }]
      },
      "resources": [
        {
          "resources": {
            "cpu": { "units": 1000 },
            "memory": { "size": "512Mi" }
          },
          "count": 1,
          "price": { "denom": "uvirt", "amount": "1000" }
        }
      ]
    },
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Example:**

```bash
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/info?id.owner=virtengine1abc...&id.dseq=12345&id.gseq=1&id.oseq=1"
```

---

### List Bids

Queries bids with filters.

```http
GET /virtengine/market/v2beta1/bids/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `filters.owner` | query | string | No | Filter by order owner |
| `filters.dseq` | query | string | No | Filter by deployment sequence |
| `filters.provider` | query | string | No | Filter by provider address |
| `filters.state` | query | string | No | Filter by state |
| `pagination.limit` | query | integer | No | Max results |

**Response:**

```json
{
  "bids": [
    {
      "bid_id": {
        "owner": "virtengine1abc...",
        "dseq": "12345",
        "gseq": 1,
        "oseq": 1,
        "provider": "virtengine1provider..."
      },
      "state": "open",
      "price": {
        "denom": "uvirt",
        "amount": "950"
      },
      "created_at": "2024-01-01T00:01:00Z"
    }
  ],
  "pagination": {
    "next_key": "base64...",
    "total": "5"
  }
}
```

---

### Get Bid

Queries details of a specific bid.

```http
GET /virtengine/market/v2beta1/bids/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `id.owner` | query | string | Yes | Order owner address |
| `id.dseq` | query | string | Yes | Deployment sequence |
| `id.gseq` | query | integer | Yes | Group sequence |
| `id.oseq` | query | integer | Yes | Order sequence |
| `id.provider` | query | string | Yes | Provider address |

---

### List Leases

Queries leases with filters.

```http
GET /virtengine/market/v2beta1/leases/list
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `filters.owner` | query | string | No | Filter by owner |
| `filters.provider` | query | string | No | Filter by provider |
| `filters.state` | query | string | No | Filter by state (`active`, `closed`, `insufficient_funds`) |
| `pagination.limit` | query | integer | No | Max results |

**Response:**

```json
{
  "leases": [
    {
      "lease_id": {
        "owner": "virtengine1abc...",
        "dseq": "12345",
        "gseq": 1,
        "oseq": 1,
        "provider": "virtengine1provider..."
      },
      "state": "active",
      "price": {
        "denom": "uvirt",
        "amount": "950"
      },
      "escrow_payment": {
        "account_id": "escrow_123",
        "payment_id": "pay_456",
        "balance": { "denom": "uvirt", "amount": "10000" },
        "withdrawn": { "denom": "uvirt", "amount": "500" }
      },
      "created_at": "2024-01-01T00:02:00Z"
    }
  ],
  "pagination": {
    "next_key": "base64...",
    "total": "10"
  }
}
```

---

### Get Lease

Queries details of a specific lease.

```http
GET /virtengine/market/v2beta1/leases/info
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `id.owner` | query | string | Yes | Owner address |
| `id.dseq` | query | string | Yes | Deployment sequence |
| `id.gseq` | query | integer | Yes | Group sequence |
| `id.oseq` | query | integer | Yes | Order sequence |
| `id.provider` | query | string | Yes | Provider address |

---

### Get Parameters

Queries the module parameters.

```http
GET /virtengine/market/v2beta1/params
```

**Response:**

```json
{
  "params": {
    "bid_min_deposit": { "denom": "uvirt", "amount": "500000" },
    "order_max_bids": 20,
    "bid_duration": "86400s",
    "order_min_duration": "3600s"
  }
}
```

---

## Transaction Messages

### Create Bid

Creates a bid on an open order (provider only).

```protobuf
message MsgCreateBid {
  OrderID order_id = 1;
  string provider = 2;
  cosmos.base.v1beta1.DecCoin price = 3;
  cosmos.base.v1beta1.Coin deposit = 4;
}
```

**Example (CLI):**

```bash
virtengine tx market create-bid \
  --owner=virtengine1abc... \
  --dseq=12345 \
  --gseq=1 \
  --oseq=1 \
  --price=950uvirt \
  --deposit=500000uvirt \
  --from provider-key \
  --chain-id virtengine-1
```

**Example (Go):**

```go
msg := &market.MsgCreateBid{
    OrderId: market.OrderID{
        Owner: ownerAddr,
        Dseq:  12345,
        Gseq:  1,
        Oseq:  1,
    },
    Provider: providerAddr,
    Price:    sdk.NewDecCoin("uvirt", sdk.NewInt(950)),
    Deposit:  sdk.NewCoin("uvirt", sdk.NewInt(500000)),
}
```

---

### Close Bid

Closes an open bid.

```protobuf
message MsgCloseBid {
  BidID bid_id = 1;
}
```

---

### Create Lease

Creates a lease from an accepted bid (called internally when bid is accepted).

```protobuf
message MsgCreateLease {
  BidID bid_id = 1;
}
```

---

### Withdraw Lease

Withdraws accumulated payment from a lease (provider only).

```protobuf
message MsgWithdrawLease {
  LeaseID lease_id = 1;
}
```

**Example (CLI):**

```bash
virtengine tx market withdraw-lease \
  --owner=virtengine1abc... \
  --dseq=12345 \
  --gseq=1 \
  --oseq=1 \
  --provider=virtengine1provider... \
  --from provider-key \
  --chain-id virtengine-1
```

---

### Close Lease

Closes an active lease.

```protobuf
message MsgCloseLease {
  LeaseID lease_id = 1;
}
```

---

## Order States

| State | Description |
|-------|-------------|
| `open` | Order is accepting bids |
| `active` | Order has an active lease |
| `closed` | Order is closed |

## Bid States

| State | Description |
|-------|-------------|
| `open` | Bid is pending acceptance |
| `active` | Bid was accepted, lease created |
| `lost` | Bid was not selected |
| `closed` | Bid was closed/withdrawn |

## Lease States

| State | Description |
|-------|-------------|
| `active` | Lease is active, workloads running |
| `insufficient_funds` | Escrow balance depleted |
| `closed` | Lease terminated |

---

## Error Codes

| Code | Message | Category | Action |
|------|---------|----------|--------|
| market:1400 | Invalid order | validation | Fix order parameters |
| market:1401 | Order not found | not_found | Verify order ID |
| market:1402 | Order already exists | conflict | Use existing order |
| market:1410 | Invalid bid | validation | Fix bid parameters |
| market:1411 | Bid not found | not_found | Verify bid ID |
| market:1420 | Invalid lease | validation | Fix lease parameters |
| market:1421 | Lease not found | not_found | Verify lease ID |
| market:1430 | Unauthorized | unauthorized | Check permissions |
| market:1440 | Insufficient deposit | validation | Increase deposit |

---

## Code Examples

### Go: Query Orders

```go
package main

import (
    "context"
    "fmt"
    
    market "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"
    "google.golang.org/grpc"
)

func main() {
    conn, _ := grpc.Dial("api.virtengine.com:9090", grpc.WithInsecure())
    defer conn.Close()
    
    client := market.NewQueryClient(conn)
    
    // Query open orders
    resp, err := client.Orders(context.Background(), &market.QueryOrdersRequest{
        Filters: market.OrderFilters{
            State: "open",
        },
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d open orders\n", len(resp.Orders))
    for _, order := range resp.Orders {
        fmt.Printf("  Order: %s/%d\n", order.OrderId.Owner, order.OrderId.Dseq)
    }
}
```

### TypeScript: Create Bid

```typescript
import { VirtEngineClient } from '@virtengine/sdk';

const client = new VirtEngineClient({
  rpcEndpoint: 'https://api.virtengine.com',
  wallet: providerWallet,
});

// Create bid on order
const result = await client.market.createBid({
  orderId: {
    owner: 'virtengine1abc...',
    dseq: '12345',
    gseq: 1,
    oseq: 1,
  },
  price: { denom: 'uvirt', amount: '950' },
  deposit: { denom: 'uvirt', amount: '500000' },
});

console.log('Bid created:', result.bidId);
```

### cURL: List Leases

```bash
# List all active leases for a provider
curl "https://api.virtengine.com/virtengine/market/v2beta1/leases/list?filters.provider=virtengine1provider...&filters.state=active"
```

---

## Marketplace Flow

```
┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
│  Tenant  │     │  Market  │     │ Provider │     │  Escrow  │
└────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
     │                │                │                │
     │ 1. Create      │                │                │
     │   Deployment   │                │                │
     │───────────────>│                │                │
     │                │                │                │
     │                │ 2. Order       │                │
     │                │   Created      │                │
     │                │───────────────>│                │
     │                │                │                │
     │                │ 3. Submit Bid  │                │
     │                │<───────────────│                │
     │                │                │                │
     │ 4. Accept Bid  │                │                │
     │───────────────>│                │                │
     │                │                │                │
     │                │ 5. Create      │                │
     │                │   Lease        │                │
     │                │───────────────>│                │
     │                │                │                │
     │                │ 6. Create      │                │
     │                │   Escrow       │                │
     │                │────────────────────────────────>│
     │                │                │                │
     │                │                │ 7. Deploy      │
     │                │                │   Workloads    │
     │                │                │                │
```

---

## See Also

- [Deployment Module](./deployment.md) - Deployment management
- [Escrow Module](./escrow.md) - Payment escrow
- [Provider Module](./provider.md) - Provider registration
- [Provider Daemon](./provider-daemon.md) - Provider services
