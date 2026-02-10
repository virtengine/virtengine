# VirtEngine API Rate Limits & Quotas

This document describes the rate limiting and quota policies for the VirtEngine API.

## Overview

VirtEngine implements multi-layered rate limiting to ensure fair usage and platform stability:

| Layer | Purpose | Scope |
|-------|---------|-------|
| IP-based | Prevent abuse | Per IP address |
| API Key | Tier-based quotas | Per API key |
| User-based | Account limits | Per wallet address |
| Endpoint-specific | Protect critical paths | Per endpoint |

## Rate Limit Tiers

### Anonymous (IP-based)

No authentication required.

| Limit Type | Value |
|------------|-------|
| Requests/second | 10 |
| Requests/minute | 100 |
| Daily quota | 10,000 |
| Burst capacity | 20 |

**Use cases:**
- Public queries
- Exploratory access
- Testing

### Standard (API Key)

Free API key registration required.

| Limit Type | Value |
|------------|-------|
| Requests/second | 50 |
| Requests/minute | 1,000 |
| Daily quota | 100,000 |
| Burst capacity | 100 |

**Use cases:**
- Application integrations
- Development environments
- Small-scale production

### Premium (Paid Tier)

Paid subscription required.

| Limit Type | Value |
|------------|-------|
| Requests/second | 200 |
| Requests/minute | 5,000 |
| Daily quota | 1,000,000 |
| Burst capacity | 500 |

**Use cases:**
- High-volume production
- Enterprise integrations
- Provider operations

### Provider (Certificate-based)

Provider certificates required.

| Limit Type | Value |
|------------|-------|
| Requests/second | 500 |
| Requests/minute | 10,000 |
| Daily quota | Unlimited |
| Burst capacity | 1,000 |

**Use cases:**
- Provider daemon operations
- Cluster management
- High-frequency updates

## Per-Endpoint Rate Limits

### Query Endpoints

Most query endpoints use standard tier limits. Exceptions:

| Endpoint | Requests/sec | Burst | Reason |
|----------|--------------|-------|--------|
| `/veid/v1/identity/{addr}` | 10 | 20 | ML scoring overhead |
| `/veid/v1/score/{addr}` | 10 | 20 | ML inference cost |
| `/market/v2beta1/orders/list` | 20 | 50 | High-traffic endpoint |
| `/provider/v1beta4/providers` | 10 | 20 | Resource-intensive query |

### Transaction Endpoints

Transaction submission has stricter limits:

| Endpoint | Requests/sec | Burst | Reason |
|----------|--------------|-------|--------|
| `/veid/v1/scope/submit` | 5 | 10 | MFA + ML verification |
| `/market/v2beta1/bid/create` | 10 | 20 | Prevent bid spam |
| `/mfa/v1/challenge/create` | 5 | 10 | Security: prevent brute force |

## Rate Limit Headers

All API responses include rate limit headers:

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 47
X-RateLimit-Reset: 1704067260
X-RateLimit-Tier: standard
X-RateLimit-Burst-Remaining: 95
Retry-After: 60
```

### Header Descriptions

| Header | Description |
|--------|-------------|
| `X-RateLimit-Limit` | Requests allowed per second |
| `X-RateLimit-Remaining` | Requests remaining in current window |
| `X-RateLimit-Reset` | Unix timestamp when limit resets |
| `X-RateLimit-Tier` | Your rate limit tier |
| `X-RateLimit-Burst-Remaining` | Burst capacity remaining |
| `Retry-After` | Seconds until next allowed request (only when limited) |

## Rate Limit Responses

### 429 Too Many Requests

When rate limit is exceeded:

```json
{
  "error": {
    "code": "rate_limit:exceeded",
    "message": "Rate limit exceeded",
    "category": "rate_limit",
    "context": {
      "limit": 50,
      "window": "1s",
      "retry_after": 60
    }
  }
}
```

**Response headers:**

```http
HTTP/1.1 429 Too Many Requests
X-RateLimit-Limit: 50
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1704067260
Retry-After: 60
Content-Type: application/json
```

## Quota Management

### Daily Quotas

Daily quotas reset at 00:00 UTC.

**Check your quota:**

```bash
curl -H "x-api-key: YOUR_KEY" \
  https://api.virtengine.com/quota/status
```

**Response:**

```json
{
  "tier": "standard",
  "daily_quota": 100000,
  "used_today": 45230,
  "remaining": 54770,
  "reset_at": "2024-01-02T00:00:00Z"
}
```

### Quota Warnings

Email notifications sent at:
- 75% quota used
- 90% quota used
- 100% quota exhausted

## Best Practices

### 1. Implement Exponential Backoff

```typescript
async function requestWithBackoff(
  url: string,
  maxRetries: number = 5
): Promise<Response> {
  for (let i = 0; i < maxRetries; i++) {
    const response = await fetch(url);
    
    if (response.status !== 429) {
      return response;
    }
    
    const retryAfter = parseInt(response.headers.get('Retry-After') || '1');
    const delay = Math.min(retryAfter * 1000, 2 ** i * 1000);
    
    await new Promise(resolve => setTimeout(resolve, delay));
  }
  
  throw new Error('Max retries exceeded');
}
```

### 2. Monitor Rate Limit Headers

```go
func checkRateLimits(resp *http.Response) {
    remaining := resp.Header.Get("X-RateLimit-Remaining")
    limit := resp.Header.Get("X-RateLimit-Limit")
    
    if remaining, _ := strconv.Atoi(remaining); remaining < 10 {
        log.Warnf("Rate limit approaching: %s/%s", remaining, limit)
        // Implement adaptive throttling
    }
}
```

### 3. Cache Responses

```python
from functools import lru_cache
import time

@lru_cache(maxsize=1000)
def get_provider_info(address: str, timestamp: int) -> dict:
    # timestamp rounded to 5 minutes for cache key
    response = requests.get(
        f"https://api.virtengine.com/virtengine/provider/v1beta4/providers/{address}"
    )
    return response.json()

# Use with rounded timestamp
provider = get_provider_info("virtengine1...", int(time.time() / 300))
```

### 4. Batch Requests

Instead of individual requests:

```bash
# Bad: Multiple requests
for addr in addresses; do
  curl "https://api.virtengine.com/virtengine/veid/v1/identity/$addr"
done
```

Use query filters:

```bash
# Good: Single request with filter
curl "https://api.virtengine.com/virtengine/veid/v1/identities?addresses=addr1,addr2,addr3"
```

### 5. Use WebSocket for Real-time Updates

Instead of polling:

```javascript
// Bad: Polling every second (3600 req/hour)
setInterval(async () => {
  const orders = await fetch('/virtengine/market/v2beta1/orders/list');
}, 1000);
```

Use WebSocket subscriptions:

```javascript
// Good: WebSocket subscription (1 connection)
const ws = new WebSocket('wss://api.virtengine.com/websocket');
ws.send(JSON.stringify({
  jsonrpc: '2.0',
  method: 'subscribe',
  params: { query: "tm.event='Tx' AND message.action='/virtengine.market.v2beta1.MsgCreateOrder'" },
  id: 1
}));
```

## Advanced Configuration

### Request Prioritization

Provider operations have priority over anonymous queries during high load:

| Priority | Tier | Queue Depth |
|----------|------|-------------|
| 1 (Highest) | Provider | Unlimited |
| 2 | Premium | 10,000 |
| 3 | Standard | 1,000 |
| 4 (Lowest) | Anonymous | 100 |

### Geographic Distribution

Rate limits are per-region:

| Region | Endpoint |
|--------|----------|
| US East | `us-east.api.virtengine.com` |
| US West | `us-west.api.virtengine.com` |
| EU | `eu.api.virtengine.com` |
| Asia Pacific | `ap.api.virtengine.com` |

**Example:**

```bash
# Use closest region for lower latency
curl https://eu.api.virtengine.com/virtengine/market/v2beta1/orders/list
```

### CDN Integration

Static query results are cached at CDN edge:

| Query Type | Cache TTL | Purge Strategy |
|------------|-----------|----------------|
| Provider list | 5 minutes | On provider update |
| Market orders | 2 seconds | On order creation |
| Identity score | 1 minute | On score recomputation |

**Cache headers:**

```http
HTTP/1.1 200 OK
X-Cache: HIT
X-Cache-Age: 45
Cache-Control: public, max-age=300
```

## Upgrading Tiers

### From Anonymous to Standard

1. Create account at https://portal.virtengine.com
2. Generate API key in Developer Settings
3. Use API key in requests:

```bash
curl -H "x-api-key: YOUR_KEY" \
  https://api.virtengine.com/...
```

### From Standard to Premium

1. Visit https://portal.virtengine.com/billing
2. Subscribe to Premium plan ($99/month)
3. API key automatically upgraded

### Provider Tier

1. Register as provider on-chain
2. Generate provider certificate:

```bash
virtengine tx cert generate provider \
  --from provider-key \
  --chain-id virtengine-1
```

3. Use certificate for mTLS authentication

## Rate Limit Bypass (Enterprise)

Custom rate limits available for enterprise customers:

- **Dedicated API instances**
- **Custom rate limit configurations**
- **Private endpoints**
- **SLA guarantees**

Contact: enterprise@virtengine.com

## Monitoring & Alerts

### Prometheus Metrics

Monitor your usage:

```prometheus
# Rate limit utilization
virtengine_api_rate_limit_utilization{tier="standard"} 0.82

# Quota usage
virtengine_api_quota_used{tier="standard"} 67000

# Throttled requests
virtengine_api_throttled_requests_total{tier="standard"} 45
```

### Grafana Dashboard

Import dashboard: https://grafana.com/dashboards/virtengine-api-metrics

### Alerts

Set up alerts for:

```yaml
- alert: RateLimitHighUsage
  expr: virtengine_api_rate_limit_utilization > 0.8
  for: 5m
  annotations:
    summary: "High rate limit usage"
    
- alert: QuotaApproaching
  expr: (virtengine_api_quota_used / virtengine_api_quota_limit) > 0.9
  for: 10m
  annotations:
    summary: "Daily quota approaching limit"
```

## Troubleshooting

### Issue: Frequently hitting rate limits

**Solutions:**
1. Implement caching (see Best Practices)
2. Use batching where available
3. Upgrade to higher tier
4. Implement exponential backoff

### Issue: Inconsistent rate limit responses

**Cause:** Multiple IPs or API keys in use

**Solution:** Verify single identity:

```bash
# Check which identity is being used
curl -I https://api.virtengine.com/quota/status
```

### Issue: Provider operations throttled

**Cause:** Certificate not recognized or expired

**Solution:**

```bash
# Check certificate status
virtengine query cert list --owner $(virtengine keys show provider-key -a)

# Renew if expired
virtengine tx cert generate provider --from provider-key
```

## API Changelog

Rate limit changes are announced 30 days in advance:

- **2024-01-01**: Premium tier daily quota increased to 1M
- **2023-12-15**: Burst capacity added for all tiers
- **2023-11-01**: Provider tier rate limits doubled

Subscribe: https://virtengine.com/changelog

## See Also

- [API Authentication Guide](./guides/authentication.md)
- [API Versioning Guide](./guides/versioning.md)
- [Error Handling](./ERROR_HANDLING.md)
- [Client Rate Limiting Guide](../../RATELIMIT_CLIENT_GUIDE.md)
