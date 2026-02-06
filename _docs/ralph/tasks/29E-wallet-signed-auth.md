# Task 29E: Wallet-Signed Request Authentication

**ID:** 29E  
**Title:** feat(portal): Wallet-signed request authentication  
**Priority:** P0 (Critical)  
**Wave:** 2 (Sequential after 29D)  
**Estimated LOC:** ~1500  
**Dependencies:** 29D (ProviderAPIClient)  
**Blocking:** 29G (Multi-provider aggregation)  

---

## Problem Statement

Provider APIs must verify that requests come from **authorized users who own the leases they're accessing**. The current HMAC approach uses a shared secret, which requires key distribution. The wallet-signed approach:

1. Uses user's existing Cosmos wallet for authentication
2. Provider verifies signature against user's on-chain address
3. Provider checks lease ownership on-chain
4. No shared secrets needed - decentralized auth

### Authentication Flow

```
┌─────────┐    1. Sign request     ┌───────────────┐
│ Browser │ ─────────────────────► │ Provider API  │
│ (Keplr) │    with wallet         │               │
└─────────┘                        │ 2. Verify sig │
                                   │ 3. Check lease│
                                   │    ownership  │
                                   └───────┬───────┘
                                           │
                                   ┌───────▼───────┐
                                   │  VirtEngine   │
                                   │    Chain      │
                                   │ (lease query) │
                                   └───────────────┘
```

---

## Acceptance Criteria

### AC-1: Client-Side Request Signing
- [ ] Implement `signRequest()` function using CosmJS
- [ ] Support ADR-036 arbitrary message signing
- [ ] Include timestamp for replay protection
- [ ] Include nonce for uniqueness
- [ ] Sign request body hash for integrity

### AC-2: Signature Message Format
- [ ] Define canonical message format
- [ ] Include: method, path, timestamp, nonce, body hash
- [ ] Ensure deterministic serialization
- [ ] Support both Amino and Direct sign modes

### AC-3: Provider-Side Verification (Go)
- [ ] Create `auth/cosmos_verify.go` middleware
- [ ] Parse signature from request headers
- [ ] Verify secp256k1 signature
- [ ] Validate timestamp within tolerance (5 min)
- [ ] Check nonce hasn't been used (anti-replay)

### AC-4: Lease Ownership Verification
- [ ] Query chain for lease by ID
- [ ] Verify signer is lease owner
- [ ] Cache ownership lookups (15 min TTL)
- [ ] Handle chain query failures gracefully

### AC-5: Authorization Levels
- [ ] **Owner**: Full access to deployment (logs, shell, actions)
- [ ] **Organization Member**: Role-based access via x/group
- [ ] **Public**: Health/status only (no auth required)
- [ ] Implement middleware for each level

### AC-6: Anti-Replay Protection
- [ ] Implement nonce store (in-memory + optional Redis)
- [ ] Reject requests with seen nonces
- [ ] Expire old nonces after timestamp window
- [ ] Handle concurrent requests safely

### AC-7: TypeScript Integration
- [ ] Update `ProviderAPIClient` with signing
- [ ] Auto-sign requests when wallet connected
- [ ] Handle signature failures gracefully
- [ ] Support unsigned requests for public endpoints

### AC-8: Testing
- [ ] Unit tests for signature generation
- [ ] Unit tests for signature verification
- [ ] Unit tests for nonce store
- [ ] E2E tests for full auth flow

---

## Technical Requirements

### Signature Message Format (ADR-036)

```typescript
// Canonical message format for signing
interface SignableMessage {
  chain_id: string;
  account_number: string;
  sequence: string;
  fee: {
    gas: string;
    amount: [];
  };
  msgs: [{
    type: "sign/MsgSignData";
    value: {
      signer: string;
      data: string;  // Base64 encoded request data
    };
  }];
  memo: string;
}

interface RequestData {
  method: string;       // GET, POST, etc.
  path: string;         // /api/v1/deployments/:id/logs
  timestamp: number;    // Unix milliseconds
  nonce: string;        // Random 32-byte hex
  body_hash: string;    // SHA256 of body (empty if no body)
}
```

### Client-Side Signing Implementation

```typescript
// lib/portal/src/auth/wallet-sign.ts
import { SigningStargateClient } from '@cosmjs/stargate';
import { sha256 } from '@cosmjs/crypto';
import { toHex, fromHex } from '@cosmjs/encoding';

export interface SignedRequestHeaders {
  'X-VE-Address': string;
  'X-VE-Timestamp': string;
  'X-VE-Nonce': string;
  'X-VE-Signature': string;
  'X-VE-PubKey': string;
}

export interface SignRequestOptions {
  method: string;
  path: string;
  body?: unknown;
  signer: any;  // Keplr/Leap offline signer
  address: string;
  chainId: string;
}

export async function signRequest(options: SignRequestOptions): Promise<SignedRequestHeaders> {
  const timestamp = Date.now();
  const nonce = generateNonce();
  const bodyHash = options.body
    ? toHex(sha256(new TextEncoder().encode(JSON.stringify(options.body))))
    : '';

  // Construct the data to sign
  const requestData: RequestData = {
    method: options.method,
    path: options.path,
    timestamp,
    nonce,
    body_hash: bodyHash,
  };

  const dataToSign = JSON.stringify(requestData);
  const dataBase64 = Buffer.from(dataToSign).toString('base64');

  // ADR-036 message format
  const signDoc = {
    chain_id: options.chainId,
    account_number: '0',
    sequence: '0',
    fee: {
      gas: '0',
      amount: [],
    },
    msgs: [{
      type: 'sign/MsgSignData',
      value: {
        signer: options.address,
        data: dataBase64,
      },
    }],
    memo: '',
  };

  // Sign using Keplr/Leap
  const signResponse = await options.signer.signAmino(
    options.chainId,
    options.address,
    signDoc,
  );

  return {
    'X-VE-Address': options.address,
    'X-VE-Timestamp': timestamp.toString(),
    'X-VE-Nonce': nonce,
    'X-VE-Signature': signResponse.signature.signature,
    'X-VE-PubKey': signResponse.signature.pub_key.value,
  };
}

function generateNonce(): string {
  const bytes = new Uint8Array(16);
  crypto.getRandomValues(bytes);
  return toHex(bytes);
}
```

### Updated ProviderAPIClient

```typescript
// lib/portal/src/provider-api/client.ts (updated)

export interface ProviderAPIClientOptions {
  endpoint: string;
  timeout?: number;
  retries?: number;
  // New: Wallet signing options
  wallet?: {
    signer: any;
    address: string;
    chainId: string;
  };
}

export class ProviderAPIClient {
  // ...existing code...

  private async request<T>(
    method: string,
    path: string,
    body?: unknown,
    requiresAuth: boolean = true,
  ): Promise<T> {
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
    };

    // Add wallet signature if available and required
    if (requiresAuth && this.wallet) {
      const authHeaders = await signRequest({
        method,
        path,
        body,
        signer: this.wallet.signer,
        address: this.wallet.address,
        chainId: this.wallet.chainId,
      });
      Object.assign(headers, authHeaders);
    }

    // ...rest of request logic...
  }

  // Public endpoint - no auth required
  async health(): Promise<HealthResponse> {
    return this.request<HealthResponse>('GET', '/api/v1/health', undefined, false);
  }

  // Protected endpoint - auth required
  async getDeploymentLogs(leaseId: string): Promise<string[]> {
    return this.request<string[]>('GET', `/api/v1/deployments/${leaseId}/logs`);
  }
}
```

### Provider-Side Verification (Go)

```go
// pkg/provider_daemon/auth/cosmos_verify.go
package auth

import (
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "strconv"
    "time"

    "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
    sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
    // Maximum age of a request signature (5 minutes)
    MaxTimestampAge = 5 * time.Minute
    
    // Header names
    HeaderAddress   = "X-VE-Address"
    HeaderTimestamp = "X-VE-Timestamp"
    HeaderNonce     = "X-VE-Nonce"
    HeaderSignature = "X-VE-Signature"
    HeaderPubKey    = "X-VE-PubKey"
)

// SignedRequest contains the parsed signature data
type SignedRequest struct {
    Address   string
    Timestamp time.Time
    Nonce     string
    Signature []byte
    PubKey    *secp256k1.PubKey
}

// RequestData is the canonical format of signed data
type RequestData struct {
    Method    string `json:"method"`
    Path      string `json:"path"`
    Timestamp int64  `json:"timestamp"`
    Nonce     string `json:"nonce"`
    BodyHash  string `json:"body_hash"`
}

// Verifier verifies wallet-signed requests
type Verifier struct {
    nonceStore NonceStore
    chainQuery ChainQuerier
}

func NewVerifier(nonceStore NonceStore, chainQuery ChainQuerier) *Verifier {
    return &Verifier{
        nonceStore: nonceStore,
        chainQuery: chainQuery,
    }
}

// Verify verifies the signature and returns the signer address
func (v *Verifier) Verify(r *http.Request) (*SignedRequest, error) {
    // Extract headers
    address := r.Header.Get(HeaderAddress)
    timestampStr := r.Header.Get(HeaderTimestamp)
    nonce := r.Header.Get(HeaderNonce)
    signatureB64 := r.Header.Get(HeaderSignature)
    pubKeyB64 := r.Header.Get(HeaderPubKey)

    if address == "" || timestampStr == "" || nonce == "" || signatureB64 == "" || pubKeyB64 == "" {
        return nil, errors.New("missing authentication headers")
    }

    // Parse timestamp
    timestampMs, err := strconv.ParseInt(timestampStr, 10, 64)
    if err != nil {
        return nil, fmt.Errorf("invalid timestamp: %w", err)
    }
    timestamp := time.UnixMilli(timestampMs)

    // Validate timestamp
    if time.Since(timestamp) > MaxTimestampAge {
        return nil, errors.New("request timestamp too old")
    }
    if timestamp.After(time.Now().Add(time.Minute)) {
        return nil, errors.New("request timestamp in future")
    }

    // Check nonce (anti-replay)
    if v.nonceStore.HasSeen(nonce) {
        return nil, errors.New("nonce already used")
    }

    // Decode signature and public key
    signature, err := base64.StdEncoding.DecodeString(signatureB64)
    if err != nil {
        return nil, fmt.Errorf("invalid signature encoding: %w", err)
    }

    pubKeyBytes, err := base64.StdEncoding.DecodeString(pubKeyB64)
    if err != nil {
        return nil, fmt.Errorf("invalid public key encoding: %w", err)
    }
    pubKey := &secp256k1.PubKey{Key: pubKeyBytes}

    // Verify address matches public key
    derivedAddr := sdk.AccAddress(pubKey.Address())
    if derivedAddr.String() != address {
        return nil, errors.New("address does not match public key")
    }

    // Reconstruct signed data
    bodyHash := ""
    if r.Body != nil {
        // Read and hash body
        body, err := io.ReadAll(r.Body)
        if err != nil {
            return nil, fmt.Errorf("failed to read body: %w", err)
        }
        r.Body = io.NopCloser(bytes.NewReader(body)) // Restore body
        
        if len(body) > 0 {
            hash := sha256.Sum256(body)
            bodyHash = hex.EncodeToString(hash[:])
        }
    }

    requestData := RequestData{
        Method:    r.Method,
        Path:      r.URL.Path,
        Timestamp: timestampMs,
        Nonce:     nonce,
        BodyHash:  bodyHash,
    }

    dataJSON, err := json.Marshal(requestData)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request data: %w", err)
    }

    // Create ADR-036 sign doc
    signDoc := createADR036SignDoc(address, base64.StdEncoding.EncodeToString(dataJSON))
    signDocBytes, err := json.Marshal(signDoc)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal sign doc: %w", err)
    }

    // Verify signature
    signDocHash := sha256.Sum256(signDocBytes)
    if !pubKey.VerifySignature(signDocHash[:], signature) {
        return nil, errors.New("signature verification failed")
    }

    // Mark nonce as used
    v.nonceStore.MarkSeen(nonce, timestamp.Add(MaxTimestampAge))

    return &SignedRequest{
        Address:   address,
        Timestamp: timestamp,
        Nonce:     nonce,
        Signature: signature,
        PubKey:    pubKey,
    }, nil
}

// VerifyLeaseOwnership checks if the address owns the lease
func (v *Verifier) VerifyLeaseOwnership(address string, leaseID string) error {
    lease, err := v.chainQuery.GetLease(leaseID)
    if err != nil {
        return fmt.Errorf("failed to query lease: %w", err)
    }

    if lease.Owner != address {
        return errors.New("address does not own this lease")
    }

    return nil
}
```

### Nonce Store Implementation

```go
// pkg/provider_daemon/auth/nonce_store.go
package auth

import (
    "sync"
    "time"
)

type NonceStore interface {
    HasSeen(nonce string) bool
    MarkSeen(nonce string, expiry time.Time)
}

// InMemoryNonceStore is a simple in-memory nonce store
type InMemoryNonceStore struct {
    mu     sync.RWMutex
    nonces map[string]time.Time
}

func NewInMemoryNonceStore() *InMemoryNonceStore {
    store := &InMemoryNonceStore{
        nonces: make(map[string]time.Time),
    }
    go store.cleanup()
    return store
}

func (s *InMemoryNonceStore) HasSeen(nonce string) bool {
    s.mu.RLock()
    defer s.mu.RUnlock()
    _, exists := s.nonces[nonce]
    return exists
}

func (s *InMemoryNonceStore) MarkSeen(nonce string, expiry time.Time) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.nonces[nonce] = expiry
}

func (s *InMemoryNonceStore) cleanup() {
    ticker := time.NewTicker(time.Minute)
    for range ticker.C {
        s.mu.Lock()
        now := time.Now()
        for nonce, expiry := range s.nonces {
            if now.After(expiry) {
                delete(s.nonces, nonce)
            }
        }
        s.mu.Unlock()
    }
}
```

### Auth Middleware

```go
// pkg/provider_daemon/portal_api.go (add middleware)

func (s *PortalAPIServer) authMiddleware(required bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Check if auth headers present
            if r.Header.Get(auth.HeaderAddress) == "" {
                if required {
                    http.Error(w, "authentication required", http.StatusUnauthorized)
                    return
                }
                // Allow unauthenticated for optional auth
                next.ServeHTTP(w, r)
                return
            }

            // Verify signature
            signedReq, err := s.verifier.Verify(r)
            if err != nil {
                http.Error(w, err.Error(), http.StatusUnauthorized)
                return
            }

            // Add verified address to context
            ctx := context.WithValue(r.Context(), "address", signedReq.Address)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

func (s *PortalAPIServer) leaseOwnerMiddleware() func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            address := r.Context().Value("address").(string)
            leaseID := chi.URLParam(r, "leaseId")

            if err := s.verifier.VerifyLeaseOwnership(address, leaseID); err != nil {
                http.Error(w, "not authorized for this lease", http.StatusForbidden)
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

---

## Files to Create/Modify

### New Files
| Path | Description | Est. Lines |
|------|-------------|------------|
| `lib/portal/src/auth/wallet-sign.ts` | Client-side signing | 150 |
| `lib/portal/src/auth/types.ts` | Auth type definitions | 50 |
| `lib/portal/src/auth/index.ts` | Module exports | 10 |
| `pkg/provider_daemon/auth/cosmos_verify.go` | Signature verification | 250 |
| `pkg/provider_daemon/auth/nonce_store.go` | Nonce tracking | 80 |
| `pkg/provider_daemon/auth/chain_query.go` | Lease ownership queries | 100 |
| `pkg/provider_daemon/auth/middleware.go` | HTTP middleware | 120 |
| `pkg/provider_daemon/auth/auth_test.go` | Unit tests | 300 |
| `tests/e2e/portal_auth_test.go` | E2E tests | 200 |

### Files to Modify
| Path | Changes |
|------|---------|
| `lib/portal/src/provider-api/client.ts` | Add wallet signing to requests |
| `pkg/provider_daemon/portal_api.go` | Add auth middleware |

**Total: ~1260 lines**

---

## Implementation Steps

### Step 1: Implement Client-Side Signing
Create `wallet-sign.ts` with ADR-036 support

### Step 2: Update ProviderAPIClient
Add wallet option and auto-signing

### Step 3: Implement Go Verifier
Create signature verification package

### Step 4: Implement Nonce Store
Create anti-replay protection

### Step 5: Add Chain Query
Implement lease ownership verification

### Step 6: Add Middleware
Integrate auth into portal_api.go routes

### Step 7: Write Tests
Unit and E2E tests

---

## Validation Checklist

- [ ] Signatures generated correctly in browser
- [ ] Go verifier accepts valid signatures
- [ ] Go verifier rejects invalid signatures
- [ ] Nonce replay protection works
- [ ] Timestamp validation works
- [ ] Lease ownership verified correctly
- [ ] Public endpoints accessible without auth
- [ ] Protected endpoints require auth
- [ ] E2E flow works end-to-end

---

## Vibe-Kanban Task ID

`5f06eefc-bf38-4c26-82a1-62c85061c173`
