# VirtEngine API Error Handling Guide

This guide helps API clients handle errors from VirtEngine services.

## Error Response Format

All VirtEngine API errors follow a consistent JSON format:

```json
{
  "error": {
    "code": "veid:1001",
    "message": "invalid scope format",
    "category": "validation",
    "severity": "error",
    "retryable": false,
    "context": {
      "field": "email",
      "value": "invalid-email"
    }
  }
}
```

### Error Fields

- **code**: Module and error code in format `module:code` (e.g., `veid:1001`)
- **message**: Human-readable error message
- **category**: Error category (see below)
- **severity**: Error severity: `info`, `warning`, `error`, `critical`
- **retryable**: Boolean indicating if the operation can be retried
- **context**: Additional structured context (optional)

## Error Categories

### validation

Input validation errors. Check your request parameters.

**Example:**
```json
{
  "error": {
    "code": "veid:1001",
    "message": "invalid email format",
    "category": "validation",
    "retryable": false,
    "context": {
      "field": "email"
    }
  }
}
```

**Client Action:** Fix the input and retry.

### not_found

Resource not found.

**Example:**
```json
{
  "error": {
    "code": "veid:1010",
    "message": "identity not found: user123",
    "category": "not_found",
    "retryable": false,
    "context": {
      "resource_type": "identity",
      "resource_id": "user123"
    }
  }
}
```

**Client Action:** Verify the resource ID. Don't retry.

### conflict

Resource already exists or operation conflicts with current state.

**Example:**
```json
{
  "error": {
    "code": "veid:1020",
    "message": "scope already exists: scope456",
    "category": "conflict",
    "retryable": false
  }
}
```

**Client Action:** Check if the resource already exists. Don't retry.

### unauthorized

Authorization required or insufficient permissions.

**Example:**
```json
{
  "error": {
    "code": "mfa:1219",
    "message": "unauthorized: MFA verification required",
    "category": "unauthorized",
    "retryable": false,
    "context": {
      "reason": "MFA verification required"
    }
  }
}
```

**Client Action:** Authenticate or obtain required permissions. Don't retry without fixing auth.

### timeout

Operation timed out.

**Example:**
```json
{
  "error": {
    "code": "inference:251",
    "message": "timeout after 30s: ML inference",
    "category": "timeout",
    "retryable": true
  }
}
```

**Client Action:** Retry with exponential backoff.

### external

External service error.

**Example:**
```json
{
  "error": {
    "code": "waldur:650",
    "message": "Waldur API unavailable",
    "category": "external",
    "retryable": true,
    "context": {
      "service": "waldur",
      "operation": "create_vm"
    }
  }
}
```

**Client Action:** Retry with exponential backoff.

### internal

Internal system error.

**Example:**
```json
{
  "error": {
    "code": "veid:1060",
    "message": "model loading failed",
    "category": "internal",
    "severity": "critical",
    "retryable": false
  }
}
```

**Client Action:** Report to support. Don't retry immediately.

### rate_limit

Rate limit exceeded.

**Example:**
```json
{
  "error": {
    "code": "nli:980",
    "message": "rate limit exceeded: 100 requests allowed, resets at 2024-01-01T00:00:00Z",
    "category": "rate_limit",
    "retryable": true,
    "context": {
      "limit": 100,
      "remaining": 0,
      "reset_at": "2024-01-01T00:00:00Z"
    }
  }
}
```

**Client Action:** Wait until `reset_at` and retry.

## HTTP Status Codes

VirtEngine APIs use standard HTTP status codes:

| Status Code | Error Category | Description |
|-------------|----------------|-------------|
| 400 | validation | Bad request - invalid input |
| 401 | unauthorized | Unauthorized - authentication required |
| 403 | unauthorized | Forbidden - insufficient permissions |
| 404 | not_found | Resource not found |
| 409 | conflict | Conflict - resource already exists |
| 429 | rate_limit | Too many requests |
| 500 | internal | Internal server error |
| 502 | external | Bad gateway - external service error |
| 503 | external | Service unavailable |
| 504 | timeout | Gateway timeout |

## Error Code Ranges

Error codes are organized by module:

### Blockchain Modules (x/)

| Module | Code Range | Description |
|--------|------------|-------------|
| veid | 1000-1099 | Identity verification |
| mfa | 1200-1299 | Multi-factor authentication |
| encryption | 1300-1399 | Encryption services |
| market | 1400-1499 | Marketplace |
| escrow | 1500-1599 | Payment escrow |
| roles | 1600-1699 | Access control |
| hpc | 1700-1799 | HPC computing |
| provider | 1800-1899 | Provider management |
| deployment | 1900-1999 | Deployments |
| cert | 2000-2099 | Certificates |
| audit | 2100-2199 | Audit logs |
| settlement | 2200-2299 | Payment settlement |
| benchmark | 2300-2399 | Benchmarking |
| staking | 2400-2499 | Staking |

### Off-Chain Services (pkg/)

| Module | Code Range | Description |
|--------|------------|-------------|
| provider_daemon | 100-199 | Provider daemon |
| inference | 200-299 | ML inference |
| workflow | 300-399 | Workflow engine |
| benchmark_daemon | 400-499 | Benchmark daemon |
| nli | 900-999 | Natural language interface |
| capture_protocol | 3300-3399 | Identity capture |
| payment | 3400-3499 | Payment processing |

## Retry Strategies

### Retryable Errors

Retry these error categories with exponential backoff:

- `timeout`
- `external`
- `rate_limit` (after reset time)

### Exponential Backoff

```javascript
async function retryWithBackoff(fn, maxRetries = 3) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error) {
      // Check if retryable
      if (!error.retryable || i === maxRetries - 1) {
        throw error;
      }
      
      // Exponential backoff: 1s, 2s, 4s
      const delay = Math.pow(2, i) * 1000;
      await sleep(delay);
    }
  }
}

// Usage
try {
  const result = await retryWithBackoff(async () => {
    return await api.createOrder(params);
  });
} catch (error) {
  console.error('Failed after retries:', error);
}
```

### Rate Limit Backoff

```javascript
async function handleRateLimit(fn) {
  try {
    return await fn();
  } catch (error) {
    if (error.category === 'rate_limit') {
      const resetAt = new Date(error.context.reset_at);
      const now = new Date();
      const waitMs = Math.max(0, resetAt - now);
      
      console.log(`Rate limited. Waiting ${waitMs}ms...`);
      await sleep(waitMs);
      
      // Retry once after reset
      return await fn();
    }
    throw error;
  }
}
```

## Common Error Codes

### Identity Verification (veid)

| Code | Message | Action |
|------|---------|--------|
| 1001 | Invalid scope format | Fix scope data format |
| 1010 | Scope not found | Verify scope ID |
| 1020 | Scope already exists | Use existing scope or delete first |
| 1030 | Unauthorized | Check permissions |
| 1036 | ML inference failed | Retry or contact support |
| 1040 | Validator key not found | Register validator key |

### Multi-Factor Authentication (mfa)

| Code | Message | Action |
|------|---------|--------|
| 1200 | Invalid address | Fix address format |
| 1207 | Invalid challenge | Generate new challenge |
| 1209 | Challenge expired | Generate new challenge |
| 1211 | Max attempts exceeded | Wait before retrying |
| 1213 | Verification failed | Check credentials |
| 1214 | MFA required | Complete MFA verification |

### Marketplace (market)

| Code | Message | Action |
|------|---------|--------|
| 1400 | Invalid order | Fix order parameters |
| 1410 | Order not found | Verify order ID |
| 1420 | Order already exists | Use existing order |
| 1430 | Unauthorized | Check permissions |

### Provider Daemon

| Code | Message | Action |
|------|---------|--------|
| 100 | Invalid configuration | Fix daemon config |
| 150 | External service error | Retry |
| 151 | Timeout | Retry |

## Client Libraries

### JavaScript/TypeScript

```typescript
interface VirtEngineError {
  error: {
    code: string;
    message: string;
    category: 'validation' | 'not_found' | 'conflict' | 'unauthorized' | 
             'timeout' | 'external' | 'internal' | 'rate_limit';
    severity: 'info' | 'warning' | 'error' | 'critical';
    retryable: boolean;
    context?: Record<string, any>;
  };
}

class VirtEngineClient {
  async handleError(response: Response): Promise<never> {
    const error: VirtEngineError = await response.json();
    
    switch (error.error.category) {
      case 'validation':
        throw new ValidationError(error);
      case 'not_found':
        throw new NotFoundError(error);
      case 'unauthorized':
        throw new UnauthorizedError(error);
      case 'rate_limit':
        throw new RateLimitError(error);
      default:
        throw new APIError(error);
    }
  }
}
```

### Python

```python
from enum import Enum
from typing import Optional, Dict, Any

class ErrorCategory(Enum):
    VALIDATION = "validation"
    NOT_FOUND = "not_found"
    CONFLICT = "conflict"
    UNAUTHORIZED = "unauthorized"
    TIMEOUT = "timeout"
    EXTERNAL = "external"
    INTERNAL = "internal"
    RATE_LIMIT = "rate_limit"

class VirtEngineError(Exception):
    def __init__(self, code: str, message: str, category: ErrorCategory, 
                 retryable: bool, context: Optional[Dict[str, Any]] = None):
        super().__init__(message)
        self.code = code
        self.category = category
        self.retryable = retryable
        self.context = context or {}

def parse_error(response: dict) -> VirtEngineError:
    error = response["error"]
    return VirtEngineError(
        code=error["code"],
        message=error["message"],
        category=ErrorCategory(error["category"]),
        retryable=error["retryable"],
        context=error.get("context")
    )
```

### Go

```go
package virtengine

import "github.com/virtengine/virtengine/pkg/errors"

type Client struct {
    // ... client fields
}

func (c *Client) CreateOrder(ctx context.Context, params OrderParams) (*Order, error) {
    resp, err := c.post(ctx, "/orders", params)
    if err != nil {
        return nil, errors.NewExternalError("virtengine_client", 0, "virtengine_api", "create_order", err.Error())
    }
    
    if resp.StatusCode != 200 {
        var apiErr APIError
        if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
            return nil, errors.Wrap(err, "failed to decode error response")
        }
        return nil, apiErr.ToError()
    }
    
    var order Order
    if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
        return nil, errors.Wrap(err, "failed to decode order")
    }
    
    return &order, nil
}
```

## Support

For issues with error handling or questions about specific error codes:

1. Check the [Error Catalog](../errors/ERROR_CATALOG.md) for complete error documentation
2. Review [Best Practices](../../_docs/ERROR_HANDLING.md) for developers
3. Open an issue on GitHub with the error code and context
4. Contact support with error details and request ID

## References

- Complete Error Catalog: [docs/errors/ERROR_CATALOG.md](../errors/ERROR_CATALOG.md)
- Developer Best Practices: [_docs/ERROR_HANDLING.md](../../_docs/ERROR_HANDLING.md)
- Error Code Registry: `pkg/errors/codes.go`
