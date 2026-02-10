# VirtEngine API Versioning Guide

This guide documents the API versioning strategy, version lifecycle, and migration procedures.

## Version Format

VirtEngine APIs use semantic versioning with stability indicators:

```
v{major}{stability}{revision}
```

### Examples

| Version | Meaning |
|---------|---------|
| `v1` | Stable major version 1 |
| `v1beta1` | Beta version of v1 (breaking changes possible) |
| `v1beta2` | Second beta iteration |
| `v2beta1` | Beta version of upcoming v2 |

## Current API Versions

### Blockchain Modules

| Module | Stable | Beta | Deprecated |
|--------|--------|------|------------|
| VEID | v1 | - | - |
| MFA | v1 | - | - |
| Encryption | v1 | - | - |
| Market | - | v2beta1 | v1beta5 |
| Deployment | - | v1beta5 | v1beta4 |
| Provider | - | v1beta4 | v1beta3 |
| Escrow | v1 | - | v1beta3 |
| Cert | v1 | - | v1beta3 |
| Audit | v1 | - | v1beta3 |
| Roles | v1 | - | - |
| Staking | v1 | - | - |
| Take | v1 | - | v1beta3 |
| HPC | v1 | - | - |
| Enclave | v1 | - | - |

### Off-Chain Services

| Service | Stable | Beta |
|---------|--------|------|
| Provider RPC | v1 | - |
| Lease RPC | v1 | - |
| Inventory RPC | v1 | - |

## Version Lifecycle

```
┌────────────────────────────────────────────────────────────────────────┐
│                          API Version Lifecycle                          │
└────────────────────────────────────────────────────────────────────────┘

  Alpha         Beta            Stable          Deprecated      Removed
    │             │                │                 │             │
    ▼             ▼                ▼                 ▼             ▼
┌────────┐   ┌────────┐       ┌────────┐       ┌────────┐     ┌────────┐
│ v1alpha│──>│ v1beta1│──────>│   v1   │──────>│   v1   │────>│ Removed│
└────────┘   └────────┘       └────────┘       │(deprecated)│  └────────┘
                │                               └────────┘
                ▼                                   │
            ┌────────┐                              │
            │ v1beta2│──────────────────────────────┘
            └────────┘         (may skip directly to deprecated)
```

### Stage Descriptions

| Stage | Breaking Changes | Support Duration | Use Case |
|-------|------------------|------------------|----------|
| Alpha | Expected | Until beta | Internal testing |
| Beta | Possible | 6+ months | Early adopters, feedback |
| Stable | No | Until deprecated | Production use |
| Deprecated | No | 12 months min | Migration period |
| Removed | N/A | N/A | No longer available |

## API Gateway Routes

The API gateway provides versioned routes:

```yaml
# Current routing structure
routes:
  - /v1/chain/*    -> virtengine-node:1317 (deprecated, sunset: 2026-12-31)
  - /v2/chain/*    -> virtengine-node:1317 (stable)
  - /v1/provider/* -> provider-daemon:8443
```

### Deprecation Headers

Deprecated routes include lifecycle headers:

```http
HTTP/1.1 200 OK
Deprecation: true
Sunset: Sat, 31 Dec 2026 23:59:59 GMT
Link: </v2/chain>; rel="successor-version"
```

## Migration Procedures

### Identifying Breaking Changes

Check the changelog for breaking changes:

```bash
# View recent changes
git log --oneline --grep="BREAKING" -- sdk/proto/
```

Breaking changes include:
- Field removals or renames
- Type changes
- Semantic changes to existing fields
- Removed endpoints

### Migration Steps

#### 1. Check Current Version

```bash
# Query your current API version
curl https://api.virtengine.com/virtengine/market/v1beta5/params

# Check deprecation headers
curl -I https://api.virtengine.com/virtengine/market/v1beta5/params | grep -i deprecation
```

#### 2. Review Migration Guide

Each version includes a migration guide:

- [Market v1beta5 to v2beta1](./migrations/market-v2beta1.md)
- [Deployment v1beta4 to v1beta5](./migrations/deployment-v1beta5.md)
- [Provider v1beta3 to v1beta4](./migrations/provider-v1beta4.md)

#### 3. Update Client Code

##### Go SDK

```go
// Before: v1beta5
import marketv1beta5 "github.com/virtengine/virtengine/sdk/go/node/market/v1beta5"

client := marketv1beta5.NewQueryClient(conn)

// After: v2beta1
import marketv2beta1 "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"

client := marketv2beta1.NewQueryClient(conn)
```

##### TypeScript SDK

```typescript
// Before
const orders = await client.market.v1beta5.orders({});

// After
const orders = await client.market.v2beta1.orders({});
```

##### REST Endpoints

```bash
# Before
curl https://api.virtengine.com/virtengine/market/v1beta5/orders/list

# After
curl https://api.virtengine.com/virtengine/market/v2beta1/orders/list
```

#### 4. Test Migration

```bash
# Run tests against new version
VIRTENGINE_API_VERSION=v2beta1 npm test

# Compare responses
diff <(curl .../v1beta5/orders/list) <(curl .../v2beta1/orders/list)
```

#### 5. Deploy Gradually

```yaml
# Feature flag for gradual rollout
api:
  market_version: "${MARKET_API_VERSION:-v1beta5}"
```

## Version Negotiation

### Accept Header

Request specific versions via Accept header:

```bash
curl -H "Accept: application/json; version=v2beta1" \
  https://api.virtengine.com/virtengine/market/orders/list
```

### Query Parameter

Override version via query parameter:

```bash
curl "https://api.virtengine.com/virtengine/market/orders/list?api-version=v2beta1"
```

### Default Version

If no version specified:
- Stable version is used if available
- Latest beta otherwise

## Protobuf Versioning

Proto files are versioned by package:

```protobuf
// sdk/proto/node/virtengine/market/v2beta1/query.proto
syntax = "proto3";
package virtengine.market.v2beta1;

option go_package = "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1";

service Query {
  // Orders queries orders with filters.
  rpc Orders(QueryOrdersRequest) returns (QueryOrdersResponse) {
    option (google.api.http).get = "/virtengine/market/v2beta1/orders/list";
  }
}
```

### Buf Configuration

Proto generation is managed via buf:

```yaml
# buf.gen.swagger.yaml
version: v1
plugins:
  - name: swagger
    out: ./.cache/tmp/swagger-gen
    opt: logtostderr=true,fqn_for_swagger_name=true,simple_operation_ids=true
```

## Deprecation Policy

### Minimum Support Periods

| Version Type | Minimum Support |
|--------------|-----------------|
| Stable → Deprecated | 12 months |
| Beta → Deprecated | 6 months |
| Deprecated → Removed | 6 months |

### Deprecation Notices

1. **API Headers**: `Deprecation: true` header added
2. **Documentation**: Warning banner on docs
3. **Changelog**: Entry in CHANGELOG.md
4. **SDK**: Deprecation warnings in code

### Example Deprecation Timeline

```
Jan 2025: v1beta5 released
Jul 2025: v2beta1 released, v1beta5 deprecated
Jan 2026: v2 stable released
Jul 2026: v1beta5 removed
```

## SDK Versioning

SDK releases are independent of API versions:

```
SDK v1.0.0 supports:
  - Market v1beta5, v2beta1
  - Provider v1beta3, v1beta4
  - VEID v1

SDK v1.1.0 supports:
  - Market v2beta1, v2 (new)
  - Provider v1beta4
  - VEID v1
  
SDK v2.0.0 supports:
  - Market v2 (v2beta1 removed)
  - Provider v1beta4
  - VEID v1
```

## Best Practices

### 1. Pin Versions Explicitly

```go
// Good: Explicit version
import market "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"

// Bad: Unversioned alias (if available)
import market "github.com/virtengine/virtengine/sdk/go/node/market"
```

### 2. Handle Version Errors

```typescript
try {
  await client.market.orders({});
} catch (error) {
  if (error.code === 'VERSION_DEPRECATED') {
    console.warn('API version deprecated, consider upgrading');
    // Fall back or notify
  }
  throw error;
}
```

### 3. Monitor Deprecation Headers

```go
func checkDeprecation(resp *http.Response) {
    if resp.Header.Get("Deprecation") == "true" {
        sunset := resp.Header.Get("Sunset")
        log.Warnf("API deprecated, sunset: %s", sunset)
    }
}
```

### 4. Subscribe to Updates

- Watch [GitHub releases](https://github.com/virtengine/virtengine/releases)
- Join [Discord](https://discord.gg/virtengine) #api-updates channel
- Subscribe to [changelog RSS](https://virtengine.com/changelog.rss)

## Common Migration Issues

### Field Renamed

```diff
// v1beta5
-message Order {
-  string owner_address = 1;
-}

// v2beta1
+message Order {
+  string owner = 1;
+}
```

**Fix**: Update field references in your code.

### Type Changed

```diff
// v1beta5
-string price = 1;

// v2beta1
+cosmos.base.v1beta1.Coin price = 1;
```

**Fix**: Update type handling:

```go
// Before
price := order.Price // string

// After
price := order.Price.Amount.String() // sdk.Coin
```

### Endpoint Moved

```diff
-GET /virtengine/market/v1beta5/orders
+GET /virtengine/market/v2beta1/orders/list
```

**Fix**: Update API paths in client configuration.

## See Also

- [API Changelog](../../CHANGELOG.md)
- [Proto Migration Guide](../../sdk/docs/api-proto-migration-guide.md)
- [SDK Release Notes](https://github.com/virtengine/virtengine/releases)
