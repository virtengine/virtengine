# VirtEngine Authentication Guide

This guide covers all authentication methods for the VirtEngine API.

## Overview

VirtEngine supports multiple authentication methods:

| Method | Use Case | Rate Limit Tier |
|--------|----------|-----------------|
| Anonymous (IP) | Public queries | Basic (10 req/s) |
| API Key | Applications, integrations | Standard (50 req/s) |
| Wallet Signature | User transactions | Standard (50 req/s) |
| MFA-Verified | Sensitive operations | Standard (50 req/s) |

## Anonymous Access

Most query endpoints are accessible without authentication:

```bash
curl https://api.virtengine.com/virtengine/market/v2beta1/orders/list
```

**Limitations:**
- Lower rate limits (10 req/s per IP)
- No access to user-specific data
- Cannot submit transactions

## API Key Authentication

### Obtaining an API Key

1. Create a VirtEngine account
2. Navigate to Developer Settings
3. Generate a new API key

### Using API Keys

Include the API key in your request headers:

```bash
# Using x-api-key header
curl -H "x-api-key: YOUR_API_KEY" \
  https://api.virtengine.com/virtengine/market/v2beta1/orders/list

# Using apikey header
curl -H "apikey: YOUR_API_KEY" \
  https://api.virtengine.com/virtengine/market/v2beta1/orders/list
```

### API Key Best Practices

1. **Never expose keys in client-side code**
2. **Rotate keys periodically**
3. **Use environment variables**

```bash
# Good: Use environment variables
export VIRTENGINE_API_KEY="your_key_here"
curl -H "x-api-key: $VIRTENGINE_API_KEY" https://api.virtengine.com/...
```

```go
// Go example
apiKey := os.Getenv("VIRTENGINE_API_KEY")
req.Header.Set("x-api-key", apiKey)
```

## Wallet Signature Authentication

For operations that require proof of wallet ownership, use Cosmos-compatible signatures.

### Signature Flow

1. **Construct the message** - Create a canonical JSON of the request
2. **Sign with private key** - Use secp256k1 signature
3. **Include in headers** - Add signature and public key to request

### Headers Required

| Header | Description |
|--------|-------------|
| `X-Cosmos-Signature` | Base64-encoded signature |
| `X-Cosmos-Pubkey` | Base64-encoded public key |
| `X-Cosmos-Timestamp` | Unix timestamp (±5 min tolerance) |
| `X-Cosmos-Sequence` | Account sequence number |

### Example: Signing a Request

```go
package main

import (
    "crypto/sha256"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "time"
    
    "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
)

func signRequest(privKey *secp256k1.PrivKey, method, path string, body interface{}) (map[string]string, error) {
    timestamp := time.Now().Unix()
    
    // Create canonical message
    message := map[string]interface{}{
        "method":    method,
        "path":      path,
        "body":      body,
        "timestamp": timestamp,
    }
    
    messageBytes, err := json.Marshal(message)
    if err != nil {
        return nil, err
    }
    
    // Hash and sign
    hash := sha256.Sum256(messageBytes)
    signature, err := privKey.Sign(hash[:])
    if err != nil {
        return nil, err
    }
    
    return map[string]string{
        "X-Cosmos-Signature": base64.StdEncoding.EncodeToString(signature),
        "X-Cosmos-Pubkey":    base64.StdEncoding.EncodeToString(privKey.PubKey().Bytes()),
        "X-Cosmos-Timestamp": fmt.Sprintf("%d", timestamp),
    }, nil
}
```

### TypeScript Example

```typescript
import { Secp256k1, sha256 } from '@cosmjs/crypto';
import { toBase64 } from '@cosmjs/encoding';

async function signRequest(
  privateKey: Uint8Array,
  method: string,
  path: string,
  body: any
): Promise<Record<string, string>> {
  const timestamp = Math.floor(Date.now() / 1000);
  
  const message = JSON.stringify({
    method,
    path,
    body,
    timestamp,
  });
  
  const hash = sha256(new TextEncoder().encode(message));
  const signature = await Secp256k1.createSignature(hash, privateKey);
  const pubkey = await Secp256k1.makeKeypair(privateKey);
  
  return {
    'X-Cosmos-Signature': toBase64(signature.toFixedLength()),
    'X-Cosmos-Pubkey': toBase64(pubkey.pubkey),
    'X-Cosmos-Timestamp': timestamp.toString(),
  };
}
```

## Multi-Factor Authentication (MFA)

Certain sensitive operations require MFA verification.

### MFA-Protected Operations

| Operation | MFA Required |
|-----------|--------------|
| Identity verification submission | Yes |
| Large fund transfers | Configurable |
| Provider registration changes | Yes |
| Key rotation | Yes |
| Account recovery | Yes |

### MFA Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │     │  VirtEngine │     │ MFA Module  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       │  1. Request Operation                 │
       │──────────────────>│                   │
       │                   │                   │
       │  2. MFA Required (401)                │
       │<──────────────────│                   │
       │                   │                   │
       │  3. Create MFA Challenge              │
       │──────────────────────────────────────>│
       │                   │                   │
       │  4. Return Challenge                  │
       │<──────────────────────────────────────│
       │                   │                   │
       │  5. Verify Challenge                  │
       │──────────────────────────────────────>│
       │                   │                   │
       │  6. Return Session Token              │
       │<──────────────────────────────────────│
       │                   │                   │
       │  7. Retry with Session Token          │
       │──────────────────>│                   │
       │                   │                   │
       │  8. Operation Success                 │
       │<──────────────────│                   │
       │                   │                   │
```

### MFA Example

#### Step 1: Attempt Operation (Gets 401)

```bash
curl -X POST \
  -H "X-Cosmos-Signature: ..." \
  https://api.virtengine.com/virtengine/veid/v1/scope/submit

# Response: 401
{
  "error": {
    "code": "mfa:1214",
    "message": "MFA verification required",
    "category": "unauthorized",
    "context": {
      "transaction_type": "veid.MsgSubmitScope"
    }
  }
}
```

#### Step 2: Create Challenge

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  https://api.virtengine.com/virtengine/mfa/v1/challenge/create \
  -d '{
    "address": "virtengine1...",
    "factor_type": "totp",
    "transaction_type": "veid.MsgSubmitScope"
  }'

# Response
{
  "challenge_id": "chal_abc123",
  "expires_at": "2024-01-01T12:05:00Z"
}
```

#### Step 3: Verify Challenge

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  https://api.virtengine.com/virtengine/mfa/v1/challenge/verify \
  -d '{
    "challenge_id": "chal_abc123",
    "response": "123456"
  }'

# Response
{
  "session_id": "sess_xyz789",
  "expires_at": "2024-01-01T12:10:00Z"
}
```

#### Step 4: Retry with Session Token

```bash
curl -X POST \
  -H "X-Cosmos-Signature: ..." \
  -H "X-MFA-Session: sess_xyz789" \
  https://api.virtengine.com/virtengine/veid/v1/scope/submit \
  -d '{ ... }'
```

### Supported MFA Factors

| Factor | Type Code | Description |
|--------|-----------|-------------|
| TOTP | `totp` | Time-based one-time password (Google Authenticator) |
| Hardware Key | `webauthn` | FIDO2/WebAuthn hardware keys |
| Recovery Code | `recovery` | One-time recovery codes |

## Provider Authentication

Providers use certificate-based authentication for provider daemon operations.

### Certificate Setup

1. Generate provider certificate:

```bash
virtengine tx cert generate client \
  --from provider-key \
  --chain-id virtengine-1
```

2. Publish certificate on-chain:

```bash
virtengine tx cert publish client \
  --from provider-key \
  --chain-id virtengine-1
```

3. Use mTLS for provider daemon:

```go
// Provider daemon configuration
config := &ProviderConfig{
    TLS: TLSConfig{
        CertFile: "/path/to/provider.crt",
        KeyFile:  "/path/to/provider.key",
        CAFile:   "/path/to/ca.crt",
    },
}
```

## JWT Tokens

Some off-chain services use JWT tokens.

### Token Format

```json
{
  "header": {
    "alg": "ES256",
    "typ": "JWT"
  },
  "payload": {
    "sub": "virtengine1...",
    "iat": 1704067200,
    "exp": 1704153600,
    "scope": ["market:read", "provider:read"]
  }
}
```

### Using JWT Tokens

```bash
curl -H "Authorization: Bearer eyJhbGciOiJFUzI1NiI..." \
  https://api.virtengine.com/provider/status
```

## Security Best Practices

### 1. Secure Key Storage

```go
// Use hardware security modules when possible
import "github.com/virtengine/virtengine/pkg/provider_daemon"

keyManager := provider_daemon.NewKeyManager(provider_daemon.KeyManagerConfig{
    StorageType: provider_daemon.KeyStorageHardware,
    HSMConfig: HSMConfig{
        Slot:  0,
        Label: "provider-key",
    },
})
```

### 2. Token Expiration

Always handle token expiration:

```typescript
class AuthClient {
  private token: string | null = null;
  private tokenExpiry: Date | null = null;

  async getToken(): Promise<string> {
    if (this.token && this.tokenExpiry && this.tokenExpiry > new Date()) {
      return this.token;
    }
    
    const response = await this.refreshToken();
    this.token = response.token;
    this.tokenExpiry = new Date(response.expires_at);
    return this.token;
  }
}
```

### 3. Request Signing

Always sign sensitive requests:

```go
func (c *Client) SubmitScope(scope *ScopeData) error {
    // Sign the request
    headers, err := c.signRequest("POST", "/veid/scope/submit", scope)
    if err != nil {
        return err
    }
    
    return c.post("/veid/scope/submit", scope, headers)
}
```

### 4. Audit Logging

Log all authentication events:

```go
log.Info("authentication_attempt",
    "method", "wallet_signature",
    "address", address,
    "success", true,
    "timestamp", time.Now(),
)
```

## Troubleshooting

### Invalid Signature

```json
{
  "error": {
    "code": "auth:001",
    "message": "invalid signature",
    "category": "unauthorized"
  }
}
```

**Solutions:**
1. Verify the message format matches exactly
2. Check timestamp is within ±5 minutes
3. Ensure public key matches the signing private key

### Expired Session

```json
{
  "error": {
    "code": "mfa:1209",
    "message": "session expired",
    "category": "unauthorized"
  }
}
```

**Solution:** Create a new MFA challenge and verify again.

### Certificate Error

```json
{
  "error": {
    "code": "cert:2001",
    "message": "certificate not found or expired",
    "category": "unauthorized"
  }
}
```

**Solutions:**
1. Verify certificate is published on-chain
2. Check certificate expiration date
3. Ensure using correct certificate for the provider address

## See Also

- [MFA Module Reference](../reference/mfa.md)
- [Certificate Module Reference](../reference/cert.md)
- [Rate Limiting Guide](../../RATELIMIT_CLIENT_GUIDE.md)
