# VEID Module API Reference

The VEID (VirtEngine Identity) module provides decentralized identity verification with ML-powered scoring.

## Overview

The VEID module enables:
- Identity scope management (upload, verification, revocation)
- ML-based identity scoring
- Identity wallet management
- Consent-based data sharing
- Appeal system for verification disputes
- Compliance integration

## Base URL

```
/virtengine/veid/v1
```

## Authentication Requirements

| Operation Type | Authentication |
|----------------|----------------|
| Query endpoints | None required |
| Scope upload | Wallet signature |
| Verification submission | Wallet signature + MFA |
| Validator operations | Validator key + MFA |

---

## Query Endpoints

### Get Identity Record

Retrieves the identity record for an account.

```http
GET /virtengine/veid/v1/identity/{account_address}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `account_address` | path | string | Yes | Bech32 account address |

**Response:**

```json
{
  "identity": {
    "account_address": "virtengine1abc...",
    "scopes": [
      {
        "scope_id": "scope_xyz123",
        "scope_type": "FACIAL_BIOMETRIC",
        "status": "VERIFIED",
        "created_at": "2024-01-01T00:00:00Z",
        "verified_at": "2024-01-01T00:05:00Z"
      }
    ],
    "score": {
      "overall": 85,
      "factors": {
        "biometric": 90,
        "document": 80,
        "liveness": 85
      }
    },
    "status": "ACTIVE",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

**Example:**

```bash
curl "https://api.virtengine.com/virtengine/veid/v1/identity/virtengine1abc..."
```

---

### Get Identity Score

Retrieves the identity score for an account.

```http
GET /virtengine/veid/v1/score/{account_address}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `account_address` | path | string | Yes | Bech32 account address |

**Response:**

```json
{
  "score": {
    "overall": 85,
    "factors": {
      "biometric": 90,
      "document": 80,
      "liveness": 85
    },
    "last_updated": "2024-01-01T00:05:00Z",
    "model_version": "v2.1.0"
  }
}
```

---

### Get Scope

Retrieves a specific identity scope.

```http
GET /virtengine/veid/v1/scope/{account_address}/{scope_id}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `account_address` | path | string | Yes | Bech32 account address |
| `scope_id` | path | string | Yes | Unique scope identifier |

**Response:**

```json
{
  "scope": {
    "scope_id": "scope_xyz123",
    "scope_type": "FACIAL_BIOMETRIC",
    "status": "VERIFIED",
    "encrypted_payload": {
      "recipient_fingerprint": "validator1_fp",
      "algorithm": "X25519-XSalsa20-Poly1305",
      "ciphertext": "base64...",
      "nonce": "base64..."
    },
    "metadata": {
      "capture_device": "mobile_app_v2",
      "capture_time": "2024-01-01T00:00:00Z"
    },
    "verification": {
      "status": "VERIFIED",
      "verified_by": "virtengine1validator...",
      "verified_at": "2024-01-01T00:05:00Z",
      "score": 90
    },
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### List Scopes

Lists all scopes for an account.

```http
GET /virtengine/veid/v1/scopes/{account_address}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `account_address` | path | string | Yes | Bech32 account address |
| `pagination.limit` | query | integer | No | Max results (default: 100) |
| `pagination.offset` | query | integer | No | Offset for pagination |
| `pagination.key` | query | string | No | Key for keyset pagination |

**Response:**

```json
{
  "scopes": [
    {
      "scope_id": "scope_xyz123",
      "scope_type": "FACIAL_BIOMETRIC",
      "status": "VERIFIED",
      "created_at": "2024-01-01T00:00:00Z"
    }
  ],
  "pagination": {
    "next_key": "base64...",
    "total": "10"
  }
}
```

---

### List Scopes by Type

Lists all scopes of a specific type for an account.

```http
GET /virtengine/veid/v1/scopes/{account_address}/type/{scope_type}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `account_address` | path | string | Yes | Bech32 account address |
| `scope_type` | path | string | Yes | Scope type (e.g., `FACIAL_BIOMETRIC`) |

**Scope Types:**

| Type | Description |
|------|-------------|
| `FACIAL_BIOMETRIC` | Facial recognition data |
| `GOVERNMENT_ID` | Government-issued ID |
| `LIVENESS_CHECK` | Liveness verification |
| `ADDRESS_PROOF` | Address verification |
| `EMAIL_VERIFICATION` | Email verification |
| `PHONE_VERIFICATION` | Phone verification |

---

### Get Identity Wallet

Retrieves the identity wallet for an account.

```http
GET /virtengine/veid/v1/wallet/{account_address}
```

**Response:**

```json
{
  "wallet": {
    "account_address": "virtengine1abc...",
    "scope_refs": [
      {
        "scope_id": "scope_xyz123",
        "scope_type": "FACIAL_BIOMETRIC",
        "status": "ACTIVE"
      }
    ],
    "consent_settings": {
      "analytics_consent": true,
      "marketing_consent": false,
      "data_retention_days": 365
    },
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### Get Consent Settings

Retrieves consent settings for an account.

```http
GET /virtengine/veid/v1/consent/{account_address}
```

**Response:**

```json
{
  "consent_settings": {
    "analytics_consent": true,
    "marketing_consent": false,
    "data_sharing_consent": true,
    "third_party_consent": false,
    "data_retention_days": 365,
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

---

### Get Verification History

Retrieves the verification history for an account.

```http
GET /virtengine/veid/v1/history/{account_address}
```

**Response:**

```json
{
  "history": [
    {
      "event_type": "VERIFICATION_SUBMITTED",
      "scope_id": "scope_xyz123",
      "timestamp": "2024-01-01T00:00:00Z",
      "details": {
        "verifier": "virtengine1validator..."
      }
    },
    {
      "event_type": "VERIFICATION_COMPLETED",
      "scope_id": "scope_xyz123",
      "timestamp": "2024-01-01T00:05:00Z",
      "details": {
        "status": "VERIFIED",
        "score": 90
      }
    }
  ]
}
```

---

### Get Derived Features

Retrieves derived features metadata (consent-gated).

```http
GET /virtengine/veid/v1/derived_features/{account_address}
```

**Response:**

```json
{
  "derived_features": {
    "feature_types": ["face_embedding", "liveness_score"],
    "last_computed": "2024-01-01T00:05:00Z",
    "model_version": "v2.1.0"
  }
}
```

---

### Get Module Parameters

```http
GET /virtengine/veid/v1/params
```

**Response:**

```json
{
  "params": {
    "min_score_threshold": 60,
    "verification_timeout": "3600s",
    "max_scopes_per_account": 10,
    "allowed_scope_types": ["FACIAL_BIOMETRIC", "GOVERNMENT_ID", "LIVENESS_CHECK"],
    "encryption_algorithm": "X25519-XSalsa20-Poly1305"
  }
}
```

---

## Appeal Queries

### Get Appeal

```http
GET /virtengine/veid/v1/appeal/{appeal_id}
```

**Response:**

```json
{
  "appeal": {
    "appeal_id": "appeal_abc123",
    "account_address": "virtengine1abc...",
    "scope_id": "scope_xyz123",
    "reason": "Incorrect verification result",
    "status": "PENDING",
    "submitted_at": "2024-01-01T00:00:00Z"
  }
}
```

### List Appeals

```http
GET /virtengine/veid/v1/appeals/{account_address}
```

---

## Compliance Queries

### Get Compliance Status

```http
GET /virtengine/veid/v1/compliance/{account_address}
```

**Response:**

```json
{
  "compliance_status": {
    "account_address": "virtengine1abc...",
    "kyc_status": "VERIFIED",
    "aml_status": "CLEAR",
    "sanctions_status": "CLEAR",
    "last_check": "2024-01-01T00:00:00Z",
    "next_check_due": "2024-07-01T00:00:00Z"
  }
}
```

### List Compliance Providers

```http
GET /virtengine/veid/v1/compliance_providers
```

---

## Model Versioning Queries

### Get Active Model

```http
GET /virtengine/veid/v1/model/{model_type}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `model_type` | path | string | Yes | Model type (e.g., `facial_verification`) |

**Response:**

```json
{
  "model": {
    "model_type": "facial_verification",
    "version": "v2.1.0",
    "hash": "sha256:abc123...",
    "status": "ACTIVE",
    "activated_at": "2024-01-01T00:00:00Z"
  }
}
```

### List Active Models

```http
GET /virtengine/veid/v1/models
```

---

## Transaction Messages

### Upload Scope

Uploads an encrypted identity scope.

```protobuf
message MsgUploadScope {
  string sender = 1;
  string scope_id = 2;
  ScopeType scope_type = 3;
  EncryptedPayloadEnvelope encrypted_payload = 4;
  bytes client_signature = 5;
  bytes user_signature = 6;
  bytes salt = 7;
  ScopeMetadata metadata = 8;
}
```

**Example (CLI):**

```bash
virtengine tx veid upload-scope \
  --scope-id="scope_xyz123" \
  --scope-type="FACIAL_BIOMETRIC" \
  --payload-file="encrypted_payload.json" \
  --from mykey \
  --chain-id virtengine-1
```

**Example (gRPC):**

```go
msg := &veid.MsgUploadScope{
    Sender:           senderAddr,
    ScopeId:          "scope_xyz123",
    ScopeType:        veid.ScopeType_FACIAL_BIOMETRIC,
    EncryptedPayload: envelope,
    ClientSignature:  clientSig,
    UserSignature:    userSig,
    Salt:             salt,
}

resp, err := client.UploadScope(ctx, msg)
```

---

### Request Verification

Requests verification of a scope.

```protobuf
message MsgRequestVerification {
  string sender = 1;
  string scope_id = 2;
  string validator_address = 3;
}
```

---

### Revoke Scope

Revokes an identity scope.

```protobuf
message MsgRevokeScope {
  string sender = 1;
  string scope_id = 2;
  string reason = 3;
}
```

---

### Create Identity Wallet

Creates an identity wallet for portable identity.

```protobuf
message MsgCreateIdentityWallet {
  string sender = 1;
  ConsentSettings initial_consent = 2;
}
```

---

### Update Consent Settings

Updates consent settings.

```protobuf
message MsgUpdateConsentSettings {
  string sender = 1;
  ConsentSettings consent = 2;
}
```

---

### Submit Appeal

Submits an appeal against a verification decision.

```protobuf
message MsgSubmitAppeal {
  string sender = 1;
  string scope_id = 2;
  string reason = 3;
  bytes evidence = 4;
}
```

---

## Error Codes

| Code | Message | Category | Action |
|------|---------|----------|--------|
| veid:1001 | Invalid scope format | validation | Fix scope data |
| veid:1010 | Scope not found | not_found | Check scope ID |
| veid:1020 | Scope already exists | conflict | Use existing scope |
| veid:1030 | Unauthorized | unauthorized | Check permissions |
| veid:1036 | ML inference failed | internal | Retry or contact support |
| veid:1040 | Validator key not found | not_found | Register validator key |
| veid:1050 | Encryption failed | internal | Check payload format |
| veid:1060 | Model loading failed | internal | Contact support |

---

## Encryption Requirements

All sensitive identity data must be encrypted using the envelope format:

```json
{
  "recipient_fingerprint": "validator_key_fingerprint",
  "algorithm": "X25519-XSalsa20-Poly1305",
  "ciphertext": "base64_encoded_ciphertext",
  "nonce": "base64_encoded_nonce"
}
```

### Encryption Flow

1. Generate ephemeral X25519 keypair
2. Derive shared secret with recipient public key
3. Encrypt payload using XSalsa20-Poly1305
4. Include nonce and recipient fingerprint

---

## Code Examples

### Go: Query Identity

```go
package main

import (
    "context"
    "fmt"
    
    veid "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
    "google.golang.org/grpc"
)

func main() {
    conn, _ := grpc.Dial("api.virtengine.com:9090", grpc.WithInsecure())
    defer conn.Close()
    
    client := veid.NewQueryClient(conn)
    
    resp, err := client.Identity(context.Background(), &veid.QueryIdentityRequest{
        AccountAddress: "virtengine1abc...",
    })
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Identity score: %d\n", resp.Identity.Score.Overall)
}
```

### TypeScript: Upload Scope

```typescript
import { VirtEngineClient } from '@virtengine/sdk';
import { encryptPayload } from '@virtengine/sdk/crypto';

const client = new VirtEngineClient({ rpcEndpoint: 'https://api.virtengine.com' });

// Encrypt the payload
const envelope = await encryptPayload(
  payload,
  validatorPublicKey,
  'X25519-XSalsa20-Poly1305'
);

// Upload scope
const result = await client.veid.uploadScope({
  scopeId: 'scope_xyz123',
  scopeType: 'FACIAL_BIOMETRIC',
  encryptedPayload: envelope,
});

console.log('Scope uploaded:', result.scopeId);
```

### cURL: Query Scopes

```bash
curl "https://api.virtengine.com/virtengine/veid/v1/scopes/virtengine1abc..."
```

---

## See Also

- [MFA Module](./mfa.md) - Multi-factor authentication for sensitive operations
- [Encryption Module](./encryption.md) - Encryption services
- [Authentication Guide](../guides/authentication.md) - Authentication setup
- [VEID Flow Spec](../../_docs/veid-flow-spec.md) - Detailed verification flow
