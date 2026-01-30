# Getting Started with VirtEngine API

This guide will help you make your first API calls to VirtEngine.

## Prerequisites

- A VirtEngine wallet address
- Basic understanding of REST APIs or gRPC
- (Optional) API key for higher rate limits

## Quick Start

### 1. Check Node Status

Verify the API is accessible:

```bash
# REST
curl https://api.virtengine.com/cosmos/base/tendermint/v1beta1/node_info

# Response
{
  "default_node_info": {
    "protocol_version": { ... },
    "network": "virtengine-1",
    "version": "0.38.6",
    ...
  }
}
```

### 2. Query Your Account Balance

```bash
curl "https://api.virtengine.com/cosmos/bank/v1beta1/balances/{your_address}"
```

### 3. Query Marketplace Orders

```bash
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list"
```

## Authentication

### Anonymous Access

Most query endpoints are accessible without authentication, subject to IP-based rate limits.

### API Key Authentication

For higher rate limits, include your API key:

```bash
curl -H "x-api-key: YOUR_API_KEY" \
  https://api.virtengine.com/virtengine/market/v2beta1/orders/list
```

### Wallet Signature Authentication

For sensitive operations, sign requests with your wallet:

```bash
# See Authentication Guide for full details
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-Cosmos-Signature: {signature}" \
  -H "X-Cosmos-Pubkey: {pubkey}" \
  https://api.virtengine.com/virtengine/veid/v1/identity/{address}
```

## Common Operations

### Identity Verification (VEID)

Query an identity record:

```bash
curl "https://api.virtengine.com/virtengine/veid/v1/identity/{address}"
```

### Marketplace

List active orders:

```bash
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.state=open"
```

Query specific order:

```bash
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/info?id.owner={owner}&id.dseq={dseq}&id.gseq={gseq}&id.oseq={oseq}"
```

### Providers

List providers:

```bash
curl "https://api.virtengine.com/virtengine/provider/v1beta4/providers/list"
```

### Deployments

Query deployments:

```bash
curl "https://api.virtengine.com/virtengine/deployment/v1beta5/deployments/list?filters.owner={address}"
```

## Using the Go SDK

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/virtengine/virtengine/sdk/go/node/market/v2beta1"
    "google.golang.org/grpc"
)

func main() {
    // Connect to gRPC endpoint
    conn, err := grpc.Dial("api.virtengine.com:9090", grpc.WithInsecure())
    if err != nil {
        panic(err)
    }
    defer conn.Close()
    
    // Create query client
    client := market.NewQueryClient(conn)
    
    // Query orders
    resp, err := client.Orders(context.Background(), &market.QueryOrdersRequest{})
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Found %d orders\n", len(resp.Orders))
}
```

## Using the TypeScript SDK

```typescript
import { VirtEngineClient } from '@virtengine/sdk';

async function main() {
  const client = new VirtEngineClient({
    rpcEndpoint: 'https://api.virtengine.com',
  });

  // Query orders
  const orders = await client.market.orders({});
  console.log(`Found ${orders.length} orders`);

  // Query identity
  const identity = await client.veid.identity({
    accountAddress: 'virtengine1...',
  });
  console.log('Identity:', identity);
}

main();
```

## Using cURL

### Query with Pagination

```bash
# First page
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?pagination.limit=10"

# Next page (use next_key from response)
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?pagination.limit=10&pagination.key={next_key}"
```

### Filter Results

```bash
# Filter by owner
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.owner={address}"

# Filter by state
curl "https://api.virtengine.com/virtengine/market/v2beta1/orders/list?filters.state=open"
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "market:1410",
    "message": "order not found",
    "category": "not_found",
    "retryable": false
  }
}
```

See the [Error Handling Guide](../ERROR_HANDLING.md) for complete error codes.

## Rate Limits

Default rate limits (per minute):

| Access Type | Requests/Second | Requests/Minute |
|-------------|-----------------|-----------------|
| Anonymous (IP) | 10 | 300 |
| Authenticated | 50 | 1,000 |

See the [Rate Limiting Guide](../../RATELIMIT_CLIENT_GUIDE.md) for details.

## Next Steps

- [Authentication Guide](./authentication.md) - Detailed auth setup
- [API Reference](../reference/) - Complete API documentation
- [Code Examples](../examples/) - More code samples
- [Versioning Guide](./versioning.md) - API version strategy

## Troubleshooting

### Connection Refused

Ensure you're using the correct endpoint:
- Mainnet: `https://api.virtengine.com`
- Testnet: `https://api.testnet.virtengine.com`

### 401 Unauthorized

For protected endpoints, ensure your API key or signature is correct.

### 429 Too Many Requests

You've hit the rate limit. Implement exponential backoff:

```javascript
async function withRetry(fn, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      if (error.status === 429) {
        const delay = Math.pow(2, i) * 1000;
        await new Promise(r => setTimeout(r, delay));
        continue;
      }
      throw error;
    }
  }
}
```

### Invalid Request

Check the API reference for required parameters and correct formats.
