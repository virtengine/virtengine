# Market and Marketplace API Reference

VirtEngine currently exposes two related but distinct market surfaces:

- `x/market` legacy REST and gRPC order, bid, and lease APIs.
- `x/marketplace` gRPC query APIs for offering price calculation and allocation lookup.

This page keeps those surfaces separate so the docs only claim transport routes that
the current binary actually serves.

## Transport Summary

| Surface | Module | Boot Status | HTTP gateway | Primary docs |
|---------|--------|-------------|--------------|--------------|
| Market orders, bids, leases | `x/market` | Booting | Yes | OpenAPI `paths` and this page |
| Marketplace pricing and allocations | `x/marketplace` | Booting | No generated gateway handlers in this build | gRPC-only section below |

## Market REST Base URL

```text
/virtengine/market/v2beta1
```

## Marketplace gRPC Service

```text
virtengine.marketplace.v1.Query
```

The marketplace query service is registered on the gRPC server at boot. Its
gateway registration path is present, but the generated HTTP handlers are not
emitted in this build, so these methods are not exposed as REST paths.

## Authentication Requirements

| Operation Type | Authentication |
|----------------|----------------|
| Query endpoints | None required |
| Order creation | Wallet signature |
| Bid operations | Provider certificate |
| Lease management | Wallet signature |

## Legacy Market REST Endpoints

### List Orders

```http
GET /virtengine/market/v2beta1/orders/list
```

Filters orders by owner or state.

### List Bids

```http
GET /virtengine/market/v2beta1/bids/list
```

Filters bids by order owner, deployment sequence, provider, or state.

### List Leases

```http
GET /virtengine/market/v2beta1/leases/list
```

Filters leases by owner, provider, or state.

### Example

```bash
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.state=open"
```

## Marketplace gRPC Query Methods

### OfferingPrice

Calculates the price quote for an on-chain offering.

```text
grpc: virtengine.marketplace.v1.Query/OfferingPrice
```

**Request fields**

| Field | Type | Required | Notes |
|------|------|----------|-------|
| `offering_id` | string | Yes | Provider/sequence offering identifier |
| `resource_units` | `map[string]uint64` | No | Overrides keyed by resource type such as `cpu`, `ram`, `storage`, `gpu`, `network` |
| `quantity` | uint32 | No | Defaults to `1` when `0` or omitted |

**Response fields**

| Field | Type | Notes |
|------|------|-------|
| `total` | `Coin` | Final quoted amount |

**Errors**

| gRPC status | Module code | Condition |
|------------|-------------|-----------|
| `INVALID_ARGUMENT` | n/a | Empty request, missing `offering_id`, malformed offering ID, or invalid pricing inputs |
| `NOT_FOUND` | `marketplace:2200` | Offering not found |
| `INVALID_ARGUMENT` | `marketplace:2239` | Price calculation rejected the supplied resource units |

**Example**

```bash
grpcurl -plaintext \
  -d '{"offering_id":"virtengine1provider.../7","resource_units":{"cpu":8,"ram":32768},"quantity":2}' \
  localhost:9090 \
  virtengine.marketplace.v1.Query/OfferingPrice
```

### AllocationsByCustomer

Returns allocation summaries for a customer address.

```text
grpc: virtengine.marketplace.v1.Query/AllocationsByCustomer
```

**Request fields**

| Field | Type | Required |
|------|------|----------|
| `customer_address` | string | Yes |
| `pagination.offset` | uint64 | No |
| `pagination.limit` | uint64 | No |

**Response fields**

Each allocation entry contains:

- `allocation_id`
- `order_id`
- `offering_id`
- `provider_address`
- `customer_address`
- `state`
- `accepted_price`
- `created_at`
- `updated_at`
- `terminated_at`
- `state_reason`

**Errors**

| gRPC status | Module code | Condition |
|------------|-------------|-----------|
| `INVALID_ARGUMENT` | n/a | Empty request or missing `customer_address` |

**Example**

```bash
grpcurl -plaintext \
  -d '{"customer_address":"virtengine1customer...","pagination":{"limit":25}}' \
  localhost:9090 \
  virtengine.marketplace.v1.Query/AllocationsByCustomer
```

### AllocationsByProvider

Returns allocation summaries for a provider address.

```text
grpc: virtengine.marketplace.v1.Query/AllocationsByProvider
```

**Request fields**

| Field | Type | Required |
|------|------|----------|
| `provider_address` | string | Yes |
| `pagination.offset` | uint64 | No |
| `pagination.limit` | uint64 | No |

**Errors**

| gRPC status | Module code | Condition |
|------------|-------------|-----------|
| `INVALID_ARGUMENT` | n/a | Empty request or missing `provider_address` |

**Example**

```bash
grpcurl -plaintext \
  -d '{"provider_address":"virtengine1provider...","pagination":{"limit":25}}' \
  localhost:9090 \
  virtengine.marketplace.v1.Query/AllocationsByProvider
```

## State and Error Notes

### Allocation State

`state` in the marketplace allocation response is the raw
`virtengine.marketplace.v1.AllocationState` gRPC enum value.

### Error Handling

The marketplace query service mixes plain gRPC status failures with module
sentinel errors:

- Request-shape failures return standard gRPC `INVALID_ARGUMENT`.
- Missing offerings return `marketplace:2200`.
- Invalid pricing calculations return `marketplace:2239`.

See [Error Handling](../ERROR_HANDLING.md) for the shared error model.

## See Also

- [HPC API Reference](./hpc.md) - Workload template query surface used by HPC placement flows
- [Escrow Module](./escrow.md) - Payment escrow and settlement context
- [Provider Module](./provider.md) - Provider lifecycle and discovery
