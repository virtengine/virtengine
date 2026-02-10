# VirtEngine Rate Limiting - Client Guide

This guide helps API clients understand and work with VirtEngine's rate limiting system.

## Table of Contents

1. [Overview](#overview)
2. [Rate Limit Headers](#rate-limit-headers)
3. [Rate Limit Tiers](#rate-limit-tiers)
4. [Handling Rate Limits](#handling-rate-limits)
5. [Best Practices](#best-practices)
6. [Code Examples](#code-examples)
7. [Troubleshooting](#troubleshooting)

---

## Overview

VirtEngine implements comprehensive rate limiting to ensure fair usage and protect against abuse. Rate limits are applied at multiple levels:

- **IP-based**: Limits per IP address
- **User-based**: Limits per authenticated user
- **Endpoint-specific**: Different limits for different endpoints
- **Global**: System-wide limits

### Why Rate Limiting?

- **Fair Access**: Ensures all users get fair access to resources
- **DDoS Protection**: Prevents denial-of-service attacks
- **Cost Control**: Reduces infrastructure costs
- **Service Quality**: Maintains performance for all users

---

## Rate Limit Headers

Every API response includes rate limit information in the headers:

```http
HTTP/1.1 200 OK
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 995
X-RateLimit-Reset: 1672531200
```

### Header Definitions

| Header | Description | Example |
|--------|-------------|---------|
| `X-RateLimit-Limit` | Maximum requests allowed in the time window | `1000` |
| `X-RateLimit-Remaining` | Number of requests remaining | `995` |
| `X-RateLimit-Reset` | Unix timestamp when the limit resets | `1672531200` |
| `Retry-After` | Seconds to wait before retrying (only on 429) | `60` |

---

## Rate Limit Tiers

### Default Limits (Per Minute)

#### Anonymous (IP-based)

```
Requests per second: 10
Requests per minute: 300
Requests per hour: 10,000
Requests per day: 100,000
Burst allowance: 20
```

#### Authenticated Users

```
Requests per second: 50
Requests per minute: 1,000
Requests per hour: 50,000
Requests per day: 500,000
Burst allowance: 100
```

### Endpoint-Specific Limits

Some endpoints have stricter limits due to their computational cost:

#### VEID Verification Endpoints

```
/veid/verify
/veid/request-verification

Requests per second: 5
Requests per minute: 100
Requests per hour: 1,000
```

#### Market Endpoints

```
/market/order/*
/market/bid/*
/market/lease/*

Requests per second: 20
Requests per minute: 600
Requests per hour: 20,000
```

#### Public Query Endpoints

```
/status
/health
/metrics

No rate limits (monitoring purposes)
```

---

## Handling Rate Limits

### When You Hit a Rate Limit

When you exceed a rate limit, you'll receive a `429 Too Many Requests` response:

```http
HTTP/1.1 429 Too Many Requests
Content-Type: application/json
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1672531260
Retry-After: 60

{
  "error": "rate_limit_exceeded",
  "message": "Too many requests. Please try again later.",
  "limit": 1000,
  "remaining": 0,
  "retry_after": 60,
  "reset_at": "2024-01-01T12:21:00Z"
}
```

### Recommended Response

1. **Stop sending requests immediately**
2. **Wait for the duration specified in `Retry-After` header**
3. **Implement exponential backoff** for repeated failures
4. **Log the incident** for monitoring

---

## Best Practices

### 1. Implement Exponential Backoff

```python
import time
import requests

def make_request_with_backoff(url, max_retries=3):
    for attempt in range(max_retries):
        response = requests.get(url)

        if response.status_code == 429:
            retry_after = int(response.headers.get('Retry-After', 60))
            wait_time = retry_after * (2 ** attempt)  # Exponential backoff
            print(f"Rate limited. Waiting {wait_time} seconds...")
            time.sleep(wait_time)
            continue

        return response

    raise Exception("Max retries exceeded")
```

### 2. Monitor Rate Limit Headers

```javascript
async function makeRequest(url) {
  const response = await fetch(url);

  const limit = parseInt(response.headers.get('X-RateLimit-Limit'));
  const remaining = parseInt(response.headers.get('X-RateLimit-Remaining'));
  const resetTime = parseInt(response.headers.get('X-RateLimit-Reset'));

  // Warn when approaching limit
  if (remaining < limit * 0.1) {
    console.warn(`Approaching rate limit: ${remaining}/${limit} remaining`);
  }

  // Proactively wait if very close to limit
  if (remaining < 5) {
    const waitTime = (resetTime * 1000) - Date.now();
    console.log(`Proactively waiting ${waitTime}ms before next request`);
    await new Promise(resolve => setTimeout(resolve, waitTime));
  }

  return response;
}
```

### 3. Batch Requests

Instead of making many individual requests, batch them when possible:

```bash
# Bad: Multiple requests
curl https://api.virtengine.com/market/order/1
curl https://api.virtengine.com/market/order/2
curl https://api.virtengine.com/market/order/3

# Good: Single batched request
curl -X POST https://api.virtengine.com/market/orders/batch \
  -H "Content-Type: application/json" \
  -d '{"ids": [1, 2, 3]}'
```

### 4. Use Webhooks Instead of Polling

Instead of polling for updates:

```javascript
// Bad: Polling every second
setInterval(() => {
  fetch('/api/order/status?id=123')
}, 1000);

// Good: Subscribe to webhook
fetch('/api/webhooks/subscribe', {
  method: 'POST',
  body: JSON.stringify({
    event: 'order.updated',
    url: 'https://your-app.com/webhook'
  })
});
```

### 5. Cache Responses

Cache responses that don't change frequently:

```go
import (
    "time"
    "sync"
)

type Cache struct {
    data map[string]CacheEntry
    mu   sync.RWMutex
}

type CacheEntry struct {
    value     interface{}
    expiresAt time.Time
}

func (c *Cache) Get(key string) (interface{}, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()

    entry, ok := c.data[key]
    if !ok || time.Now().After(entry.expiresAt) {
        return nil, false
    }

    return entry.value, true
}

func makeRequestWithCache(url string, cache *Cache) (interface{}, error) {
    // Try cache first
    if cached, ok := cache.Get(url); ok {
        return cached, nil
    }

    // Make request if not cached
    response := makeRequest(url)
    cache.Set(url, response, 5*time.Minute)

    return response, nil
}
```

### 6. Authenticate Your Requests

Authenticated users get higher rate limits:

```bash
# Include authentication token
curl https://api.virtengine.com/market/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

---

## Code Examples

### Python

```python
import time
import requests
from typing import Optional

class VirtEngineClient:
    def __init__(self, base_url: str, api_key: Optional[str] = None):
        self.base_url = base_url
        self.session = requests.Session()
        if api_key:
            self.session.headers['Authorization'] = f'Bearer {api_key}'

    def request(self, method: str, endpoint: str, **kwargs):
        url = f"{self.base_url}{endpoint}"
        max_retries = 3

        for attempt in range(max_retries):
            response = self.session.request(method, url, **kwargs)

            # Check rate limit headers
            remaining = int(response.headers.get('X-RateLimit-Remaining', 1))
            if remaining < 10:
                print(f"Warning: Only {remaining} requests remaining")

            # Handle rate limiting
            if response.status_code == 429:
                retry_after = int(response.headers.get('Retry-After', 60))
                wait_time = retry_after * (2 ** attempt)
                print(f"Rate limited. Waiting {wait_time}s (attempt {attempt + 1}/{max_retries})")
                time.sleep(wait_time)
                continue

            response.raise_for_status()
            return response.json()

        raise Exception("Max retries exceeded due to rate limiting")

# Usage
client = VirtEngineClient(
    base_url="https://api.virtengine.com",
    api_key="your_api_key_here"
)

try:
    orders = client.request('GET', '/market/orders')
    print(f"Retrieved {len(orders)} orders")
except Exception as e:
    print(f"Error: {e}")
```

### Go

```go
package main

import (
    "fmt"
    "net/http"
    "strconv"
    "time"
)

type RateLimitClient struct {
    client  *http.Client
    baseURL string
    apiKey  string
}

func NewRateLimitClient(baseURL, apiKey string) *RateLimitClient {
    return &RateLimitClient{
        client:  &http.Client{Timeout: 10 * time.Second},
        baseURL: baseURL,
        apiKey:  apiKey,
    }
}

func (c *RateLimitClient) Request(method, endpoint string) (*http.Response, error) {
    url := c.baseURL + endpoint
    maxRetries := 3

    for attempt := 0; attempt < maxRetries; attempt++ {
        req, err := http.NewRequest(method, url, nil)
        if err != nil {
            return nil, err
        }

        if c.apiKey != "" {
            req.Header.Set("Authorization", "Bearer "+c.apiKey)
        }

        resp, err := c.client.Do(req)
        if err != nil {
            return nil, err
        }

        // Check rate limit headers
        remaining := resp.Header.Get("X-RateLimit-Remaining")
        if remaining != "" {
            if r, err := strconv.Atoi(remaining); err == nil && r < 10 {
                fmt.Printf("Warning: Only %d requests remaining\n", r)
            }
        }

        // Handle rate limiting
        if resp.StatusCode == http.StatusTooManyRequests {
            retryAfter := resp.Header.Get("Retry-After")
            waitSeconds := 60
            if retryAfter != "" {
                waitSeconds, _ = strconv.Atoi(retryAfter)
            }

            waitTime := time.Duration(waitSeconds*(1<<attempt)) * time.Second
            fmt.Printf("Rate limited. Waiting %v (attempt %d/%d)\n",
                waitTime, attempt+1, maxRetries)

            resp.Body.Close()
            time.Sleep(waitTime)
            continue
        }

        return resp, nil
    }

    return nil, fmt.Errorf("max retries exceeded due to rate limiting")
}

func main() {
    client := NewRateLimitClient(
        "https://api.virtengine.com",
        "your_api_key_here",
    )

    resp, err := client.Request("GET", "/market/orders")
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    defer resp.Body.Close()

    fmt.Println("Request successful!")
}
```

### JavaScript/TypeScript

```typescript
interface RateLimitInfo {
  limit: number;
  remaining: number;
  reset: number;
}

class VirtEngineClient {
  private baseURL: string;
  private apiKey?: string;
  private rateLimitInfo?: RateLimitInfo;

  constructor(baseURL: string, apiKey?: string) {
    this.baseURL = baseURL;
    this.apiKey = apiKey;
  }

  private async sleep(ms: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, ms));
  }

  private updateRateLimitInfo(headers: Headers): void {
    this.rateLimitInfo = {
      limit: parseInt(headers.get('X-RateLimit-Limit') || '0'),
      remaining: parseInt(headers.get('X-RateLimit-Remaining') || '0'),
      reset: parseInt(headers.get('X-RateLimit-Reset') || '0')
    };
  }

  async request(
    method: string,
    endpoint: string,
    options: RequestInit = {}
  ): Promise<any> {
    const url = `${this.baseURL}${endpoint}`;
    const maxRetries = 3;

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers
    };

    if (this.apiKey) {
      headers['Authorization'] = `Bearer ${this.apiKey}`;
    }

    for (let attempt = 0; attempt < maxRetries; attempt++) {
      // Proactive rate limit check
      if (this.rateLimitInfo && this.rateLimitInfo.remaining < 5) {
        const waitTime = (this.rateLimitInfo.reset * 1000) - Date.now();
        if (waitTime > 0) {
          console.log(`Proactively waiting ${waitTime}ms to avoid rate limit`);
          await this.sleep(waitTime);
        }
      }

      const response = await fetch(url, {
        ...options,
        method,
        headers
      });

      this.updateRateLimitInfo(response.headers);

      if (response.status === 429) {
        const retryAfter = parseInt(response.headers.get('Retry-After') || '60');
        const waitTime = retryAfter * 1000 * Math.pow(2, attempt);

        console.warn(
          `Rate limited. Waiting ${waitTime}ms (attempt ${attempt + 1}/${maxRetries})`
        );

        await this.sleep(waitTime);
        continue;
      }

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      return await response.json();
    }

    throw new Error('Max retries exceeded due to rate limiting');
  }

  getRateLimitInfo(): RateLimitInfo | undefined {
    return this.rateLimitInfo;
  }
}

// Usage
const client = new VirtEngineClient(
  'https://api.virtengine.com',
  'your_api_key_here'
);

async function main() {
  try {
    const orders = await client.request('GET', '/market/orders');
    console.log(`Retrieved ${orders.length} orders`);

    const rateLimitInfo = client.getRateLimitInfo();
    console.log(`Rate limit: ${rateLimitInfo?.remaining}/${rateLimitInfo?.limit} remaining`);
  } catch (error) {
    console.error('Error:', error);
  }
}

main();
```

---

## Troubleshooting

### Issue: Getting Rate Limited Frequently

**Symptoms:**
- Frequent 429 responses
- `X-RateLimit-Remaining` often at 0

**Solutions:**
1. Reduce request frequency
2. Implement request queuing
3. Use caching for repeated requests
4. Consider upgrading to authenticated access
5. Contact support for higher limits if needed

### Issue: Rate Limits Reset Unexpectedly

**Symptoms:**
- Rate limits reset at unexpected times
- Inconsistent `X-RateLimit-Reset` values

**Explanation:**
Rate limits are tracked in multiple time windows (second, minute, hour, day). Different windows reset at different times.

**Solution:**
Always check the `X-RateLimit-Reset` header and implement proper backoff.

### Issue: Still Rate Limited After Waiting

**Symptoms:**
- Waited for `Retry-After` duration but still getting 429

**Possible Causes:**
1. Multiple rate limit types (IP + user)
2. Endpoint-specific limits
3. System-wide load shedding

**Solution:**
Implement exponential backoff and check all rate limit headers.

### Issue: Different Limits Than Documentation

**Symptoms:**
- Actual limits don't match documented limits

**Possible Causes:**
1. System under load (graceful degradation active)
2. IP-based vs user-based limits
3. Special restrictions on your IP/account

**Solution:**
1. Always use the values from response headers
2. Contact support if limits seem incorrect
3. Check if you're hitting endpoint-specific limits

---

## Support

### Contact

- **Email**: support@virtengine.com
- **Discord**: https://discord.gg/virtengine
- **GitHub**: https://github.com/virtengine/virtengine/issues

### Rate Limit Increase Requests

If you need higher rate limits:

1. Email support@virtengine.com with:
   - Your use case description
   - Current request volume
   - Requested limits
   - Business justification

2. Enterprise customers can contact their account manager

### Monitoring Your Usage

Dashboard coming soon! In the meantime, track your usage by monitoring the `X-RateLimit-*` headers in your application.

---

## Changelog

- **2024-01-15**: Initial rate limiting implementation
- **2024-01-16**: Added endpoint-specific limits
- **2024-01-17**: Implemented graceful degradation

For the latest updates, see the [CHANGELOG](./CHANGELOG.md).
