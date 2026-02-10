# MFA Module API Reference

The MFA (Multi-Factor Authentication) module provides additional security for sensitive blockchain operations.

## Overview

The MFA module enables:
- Factor enrollment (TOTP, WebAuthn, Recovery Codes)
- Challenge-based verification
- Session management for MFA-verified operations
- Trusted device registration
- Sensitive transaction gating

## Base URL

```
/virtengine/mfa/v1
```

## Authentication Requirements

| Operation Type | Authentication |
|----------------|----------------|
| Query endpoints | None required |
| Factor enrollment | Wallet signature |
| Challenge verification | Wallet signature |
| Admin operations | Governance only |

---

## Query Endpoints

### Get MFA Policy

Retrieves the MFA policy for an account.

```http
GET /virtengine/mfa/v1/policy/{address}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `address` | path | string | Yes | Bech32 account address |

**Response:**

```json
{
  "policy": {
    "address": "virtengine1abc...",
    "enabled": true,
    "required_factors": 2,
    "enrolled_factors": [
      {
        "factor_id": "totp_123",
        "factor_type": "totp",
        "label": "Authenticator App",
        "enrolled_at": "2024-01-01T00:00:00Z"
      },
      {
        "factor_id": "webauthn_456",
        "factor_type": "webauthn",
        "label": "YubiKey",
        "enrolled_at": "2024-01-01T00:00:00Z"
      }
    ],
    "recovery_enabled": true,
    "updated_at": "2024-01-01T00:00:00Z"
  }
}
```

**Example:**

```bash
curl "https://api.virtengine.com/virtengine/mfa/v1/policy/virtengine1abc..."
```

---

### Get Factor Enrollments

Retrieves all factor enrollments for an account.

```http
GET /virtengine/mfa/v1/enrollments/{address}
```

**Response:**

```json
{
  "enrollments": [
    {
      "factor_id": "totp_123",
      "factor_type": "totp",
      "label": "Authenticator App",
      "status": "active",
      "enrolled_at": "2024-01-01T00:00:00Z",
      "last_used_at": "2024-01-15T10:30:00Z"
    },
    {
      "factor_id": "webauthn_456",
      "factor_type": "webauthn",
      "label": "YubiKey",
      "status": "active",
      "metadata": {
        "credential_id": "abc123...",
        "device_type": "security_key"
      },
      "enrolled_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

---

### Get Factor Enrollment

Retrieves a specific factor enrollment.

```http
GET /virtengine/mfa/v1/enrollment/{address}/{factor_id}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `address` | path | string | Yes | Account address |
| `factor_id` | path | string | Yes | Factor identifier |

---

### Get Challenge

Retrieves a challenge by ID.

```http
GET /virtengine/mfa/v1/challenge/{challenge_id}
```

**Response:**

```json
{
  "challenge": {
    "challenge_id": "chal_abc123",
    "address": "virtengine1abc...",
    "factor_type": "totp",
    "transaction_type": "veid.MsgSubmitScope",
    "status": "pending",
    "created_at": "2024-01-01T12:00:00Z",
    "expires_at": "2024-01-01T12:05:00Z",
    "attempts": 0,
    "max_attempts": 3
  }
}
```

---

### Get Pending Challenges

Retrieves pending challenges for an account.

```http
GET /virtengine/mfa/v1/challenges/{address}
```

**Response:**

```json
{
  "challenges": [
    {
      "challenge_id": "chal_abc123",
      "factor_type": "totp",
      "transaction_type": "veid.MsgSubmitScope",
      "status": "pending",
      "expires_at": "2024-01-01T12:05:00Z"
    }
  ]
}
```

---

### Get Authorization Session

Retrieves an authorization session by ID.

```http
GET /virtengine/mfa/v1/session/{session_id}
```

**Response:**

```json
{
  "session": {
    "session_id": "sess_xyz789",
    "address": "virtengine1abc...",
    "authorized_actions": ["veid.MsgSubmitScope", "veid.MsgRevokeScope"],
    "status": "active",
    "created_at": "2024-01-01T12:05:00Z",
    "expires_at": "2024-01-01T12:10:00Z",
    "used_count": 0
  }
}
```

---

### Get Trusted Devices

Retrieves trusted devices for an account.

```http
GET /virtengine/mfa/v1/devices/{address}
```

**Response:**

```json
{
  "devices": [
    {
      "device_id": "dev_123",
      "name": "MacBook Pro",
      "fingerprint": "abc123...",
      "trusted_at": "2024-01-01T00:00:00Z",
      "last_used_at": "2024-01-15T10:30:00Z",
      "ip_address": "192.168.1.1",
      "user_agent": "Mozilla/5.0..."
    }
  ]
}
```

---

### Get Sensitive Tx Config

Retrieves configuration for a sensitive transaction type.

```http
GET /virtengine/mfa/v1/sensitive_tx/{transaction_type}
```

**Parameters:**

| Name | In | Type | Required | Description |
|------|------|------|----------|-------------|
| `transaction_type` | path | string | Yes | Transaction type (e.g., `veid.MsgSubmitScope`) |

**Response:**

```json
{
  "config": {
    "transaction_type": "veid.MsgSubmitScope",
    "mfa_required": true,
    "min_factors": 1,
    "allowed_factors": ["totp", "webauthn"],
    "session_duration": "300s",
    "bypass_for_trusted_devices": false
  }
}
```

---

### Get All Sensitive Tx Configs

Retrieves all sensitive transaction configurations.

```http
GET /virtengine/mfa/v1/sensitive_tx
```

**Response:**

```json
{
  "configs": [
    {
      "transaction_type": "veid.MsgSubmitScope",
      "mfa_required": true,
      "min_factors": 1
    },
    {
      "transaction_type": "provider.MsgUpdateProvider",
      "mfa_required": true,
      "min_factors": 2
    },
    {
      "transaction_type": "bank.MsgSend",
      "mfa_required": false,
      "threshold_amount": "1000000uvirt"
    }
  ]
}
```

---

### Check MFA Required

Checks if MFA is required for a specific transaction.

```http
GET /virtengine/mfa/v1/required/{address}/{transaction_type}
```

**Response:**

```json
{
  "required": true,
  "factors_needed": 1,
  "allowed_factors": ["totp", "webauthn", "recovery"],
  "reason": "Sensitive transaction type",
  "bypass_available": false
}
```

---

### Get Parameters

```http
GET /virtengine/mfa/v1/params
```

**Response:**

```json
{
  "params": {
    "challenge_duration": "300s",
    "session_duration": "600s",
    "max_challenge_attempts": 3,
    "cooldown_after_failed_attempts": "900s",
    "max_trusted_devices": 10,
    "recovery_code_count": 10
  }
}
```

---

## Transaction Messages

### Enroll Factor

Enrolls a new MFA factor.

```protobuf
message MsgEnrollFactor {
  string sender = 1;
  FactorType factor_type = 2;
  bytes factor_data = 3;
  string label = 4;
}
```

**Factor Types:**

| Type | Description | Factor Data |
|------|-------------|-------------|
| `totp` | Time-based OTP | None (secret generated server-side) |
| `webauthn` | WebAuthn/FIDO2 | Credential response |
| `recovery` | Recovery codes | None (codes generated server-side) |

**Example (CLI):**

```bash
virtengine tx mfa enroll-factor \
  --factor-type=totp \
  --label="Authenticator App" \
  --from mykey \
  --chain-id virtengine-1
```

---

### Remove Factor

Removes an enrolled MFA factor.

```protobuf
message MsgRemoveFactor {
  string sender = 1;
  string factor_id = 2;
  bytes verification = 3;  // Verification from another factor
}
```

---

### Create Challenge

Creates an MFA challenge.

```protobuf
message MsgCreateChallenge {
  string sender = 1;
  FactorType factor_type = 2;
  string transaction_type = 3;
}
```

**Example:**

```go
msg := &mfa.MsgCreateChallenge{
    Sender:          senderAddr,
    FactorType:      mfa.FactorType_TOTP,
    TransactionType: "veid.MsgSubmitScope",
}
```

---

### Verify Challenge

Verifies an MFA challenge and creates a session.

```protobuf
message MsgVerifyChallenge {
  string sender = 1;
  string challenge_id = 2;
  bytes response = 3;
}
```

**Response Formats by Factor Type:**

| Factor Type | Response Format |
|-------------|-----------------|
| `totp` | 6-digit code as string |
| `webauthn` | WebAuthn assertion response (JSON) |
| `recovery` | Recovery code string |

**Example (CLI):**

```bash
virtengine tx mfa verify-challenge \
  --challenge-id=chal_abc123 \
  --response="123456" \
  --from mykey \
  --chain-id virtengine-1
```

---

### Register Trusted Device

Registers a trusted device.

```protobuf
message MsgRegisterTrustedDevice {
  string sender = 1;
  string device_name = 2;
  bytes device_fingerprint = 3;
}
```

---

### Revoke Trusted Device

Revokes a trusted device.

```protobuf
message MsgRevokeTrustedDevice {
  string sender = 1;
  string device_id = 2;
}
```

---

## Factor Types

| Type | Security Level | Description |
|------|----------------|-------------|
| `totp` | Medium | Time-based one-time password (Google Authenticator, Authy) |
| `webauthn` | High | Hardware security keys (YubiKey, Titan) |
| `recovery` | Low (backup) | One-time recovery codes |

---

## MFA Flow

### Standard MFA Flow

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │     │  VirtEngine │     │ MFA Module  │
└──────┬──────┘     └──────┬──────┘     └──────┬──────┘
       │                   │                   │
       │  1. Submit Tx     │                   │
       │──────────────────>│                   │
       │                   │                   │
       │                   │ 2. Check MFA Req  │
       │                   │──────────────────>│
       │                   │                   │
       │                   │ 3. MFA Required   │
       │                   │<──────────────────│
       │                   │                   │
       │  4. 401 MFA Req   │                   │
       │<──────────────────│                   │
       │                   │                   │
       │  5. Create Challenge                  │
       │──────────────────────────────────────>│
       │                   │                   │
       │  6. Challenge     │                   │
       │<──────────────────────────────────────│
       │                   │                   │
       │  [User enters code on device]         │
       │                   │                   │
       │  7. Verify Challenge                  │
       │──────────────────────────────────────>│
       │                   │                   │
       │  8. Session Token │                   │
       │<──────────────────────────────────────│
       │                   │                   │
       │  9. Retry Tx with Session             │
       │──────────────────>│                   │
       │                   │                   │
       │  10. Success      │                   │
       │<──────────────────│                   │
```

### TOTP Setup Flow

```
┌─────────────┐     ┌─────────────┐
│   Client    │     │ MFA Module  │
└──────┬──────┘     └──────┬──────┘
       │                   │
       │  1. Enroll TOTP   │
       │──────────────────>│
       │                   │
       │  2. Secret + QR   │
       │<──────────────────│
       │                   │
       │  [User scans QR]  │
       │                   │
       │  3. Confirm Code  │
       │──────────────────>│
       │                   │
       │  4. Enrolled      │
       │<──────────────────│
```

---

## Error Codes

| Code | Message | Category | Action |
|------|---------|----------|--------|
| mfa:1200 | Invalid address | validation | Fix address format |
| mfa:1201 | Factor not found | not_found | Check factor ID |
| mfa:1205 | Invalid factor type | validation | Use valid factor type |
| mfa:1207 | Invalid challenge | validation | Create new challenge |
| mfa:1209 | Challenge expired | timeout | Create new challenge |
| mfa:1211 | Max attempts exceeded | rate_limit | Wait for cooldown |
| mfa:1213 | Verification failed | unauthorized | Check code/response |
| mfa:1214 | MFA required | unauthorized | Complete MFA flow |
| mfa:1219 | Session expired | unauthorized | Create new session |

---

## Code Examples

### Go: Check MFA Required

```go
package main

import (
    "context"
    "fmt"
    
    mfa "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
    "google.golang.org/grpc"
)

func main() {
    conn, _ := grpc.Dial("api.virtengine.com:9090", grpc.WithInsecure())
    defer conn.Close()
    
    client := mfa.NewQueryClient(conn)
    
    resp, err := client.MFARequired(context.Background(), &mfa.QueryMFARequiredRequest{
        Address:         "virtengine1abc...",
        TransactionType: "veid.MsgSubmitScope",
    })
    if err != nil {
        panic(err)
    }
    
    if resp.Required {
        fmt.Printf("MFA required: %d factor(s) needed\n", resp.FactorsNeeded)
        fmt.Printf("Allowed factors: %v\n", resp.AllowedFactors)
    } else {
        fmt.Println("MFA not required")
    }
}
```

### TypeScript: Complete MFA Flow

```typescript
import { VirtEngineClient } from '@virtengine/sdk';

const client = new VirtEngineClient({
  rpcEndpoint: 'https://api.virtengine.com',
  wallet: userWallet,
});

async function submitWithMFA(tx: any) {
  try {
    // Try to submit transaction
    return await client.signAndBroadcast(tx);
  } catch (error) {
    if (error.code === 'mfa:1214') {
      // MFA required - create challenge
      const challenge = await client.mfa.createChallenge({
        factorType: 'totp',
        transactionType: tx.typeUrl,
      });

      // Prompt user for TOTP code
      const code = await promptUser('Enter your 2FA code:');

      // Verify challenge
      const session = await client.mfa.verifyChallenge({
        challengeId: challenge.challengeId,
        response: code,
      });

      // Retry with session
      return await client.signAndBroadcast(tx, {
        mfaSession: session.sessionId,
      });
    }
    throw error;
  }
}
```

### cURL: Create and Verify Challenge

```bash
# 1. Create challenge
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-Cosmos-Signature: ..." \
  https://api.virtengine.com/virtengine/mfa/v1/tx/create-challenge \
  -d '{
    "factor_type": "totp",
    "transaction_type": "veid.MsgSubmitScope"
  }'

# 2. Verify challenge
curl -X POST \
  -H "Content-Type: application/json" \
  -H "X-Cosmos-Signature: ..." \
  https://api.virtengine.com/virtengine/mfa/v1/tx/verify-challenge \
  -d '{
    "challenge_id": "chal_abc123",
    "response": "123456"
  }'
```

---

## See Also

- [VEID Module](./veid.md) - Identity verification (requires MFA)
- [Authentication Guide](../guides/authentication.md) - Full authentication setup
- [Error Handling](../ERROR_HANDLING.md) - Error code reference
