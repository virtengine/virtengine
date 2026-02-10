# VirtEngine Developer Guide

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Task Reference:** VE-803

---

## Table of Contents

1. [Overview](#overview)
2. [Getting Started](#getting-started)
3. [Module APIs](#module-apis)
4. [Encryption Envelope Specification](#encryption-envelope-specification)
5. [Event Schema](#event-schema)
6. [Transaction Examples](#transaction-examples)
7. [Local Devnet Setup](#local-devnet-setup)
8. [SDK Reference](#sdk-reference)
9. [Security Guidance](#security-guidance)

---

## Overview

VirtEngine is a Cosmos SDK-based blockchain with specialized modules for:

- **VEID**: Decentralized identity verification with ML-powered scoring
- **MFA**: Multi-factor authentication gating for sensitive transactions
- **Encryption**: Public-key encryption for sensitive on-chain data
- **Market**: Cloud marketplace with encrypted order payloads
- **HPC**: High-performance computing job scheduling
- **Benchmark**: Provider performance metrics and trust signals

## Getting Started

### Prerequisites

- Go 1.25.5+
- Node.js 18+ (for TypeScript SDK)
- Python 3.10+ (for Python helpers)
- Docker & Docker Compose

### Quick Start

```bash
# Clone repository
git clone https://github.com/virtengine/virtengine.git
cd virtengine

# Build
make build

# Run localnet
make localnet-start

# Run tests
make test
```

### Project Structure

```
virtengine/
├── app/           # Cosmos SDK app wiring
├── cmd/           # CLI binaries
├── x/             # Cosmos SDK modules
│   ├── veid/      # Identity verification
│   ├── mfa/       # Multi-factor authentication
│   ├── encryption/# Encryption primitives
│   ├── market/    # Marketplace
│   ├── hpc/       # HPC scheduling
│   └── benchmark/ # Provider benchmarking
├── client/        # Client utilities
├── pkg/           # Shared packages
├── tests/         # Integration tests
└── lib/           # Portal and SDK libraries
```

## Module APIs

### VEID Module

#### Query: Get Identity Score

```bash
# CLI
virtengine query veid score <address>

# REST
GET /virtengine/veid/v1/scores/{address}

# gRPC
rpc GetScore(QueryScoreRequest) returns (QueryScoreResponse)
```

Response:
```json
{
  "score": {
    "owner": "virtengine1...",
    "score": 85,
    "computed_at": "2026-01-24T12:00:00Z",
    "expires_at": "2026-02-24T12:00:00Z",
    "scopes_hash": "abc123..."
  }
}
```

#### Transaction: Upload Identity Scopes

```bash
virtengine tx veid upload-scopes \
    --scopes-file scopes.json \
    --from user \
    --keyring-backend test
```

Message format:
```go
type MsgUploadScopes struct {
    Owner           string                     `json:"owner"`
    Scopes          []EncryptedScope           `json:"scopes"`
    Salt            []byte                     `json:"salt"`
    ClientSignature []byte                     `json:"client_signature"`
    UserSignature   []byte                     `json:"user_signature"`
}
```

### MFA Module

#### Query: Get Account MFA Status

```bash
# REST
GET /virtengine/mfa/v1/accounts/{address}/status
```

Response:
```json
{
  "status": {
    "address": "virtengine1...",
    "enrolled_factors": ["totp", "fido2"],
    "policy": {
      "required_factors": 1,
      "allowed_factor_types": ["totp", "fido2", "backup_code"]
    },
    "trusted_devices": 2
  }
}
```

#### Transaction: Verify MFA Challenge

```go
type MsgVerifyChallenge struct {
    Address     string             `json:"address"`
    ChallengeID string             `json:"challenge_id"`
    Response    *ChallengeResponse `json:"response"`
}
```

### Encryption Module

#### Query: Get Recipient Keys

```bash
# REST
GET /virtengine/encryption/v1/keys/{address}
```

Response:
```json
{
  "keys": [
    {
      "fingerprint": "abc123...",
      "public_key": "base64...",
      "algorithm_id": "X25519-XSalsa20-Poly1305",
      "label": "Primary Key",
      "registered_at": "2026-01-24T00:00:00Z",
      "active": true
    }
  ]
}
```

### Market Module

#### Query: List Offerings

```bash
# REST
GET /virtengine/market/v1/offerings?region=us-east&min_score=50
```

#### Transaction: Create Order

```go
type MsgCreateOrder struct {
    Customer         string                   `json:"customer"`
    OfferingID       string                   `json:"offering_id"`
    Quantity         int32                    `json:"quantity"`
    EncryptedDetails *EncryptedPayloadEnvelope `json:"encrypted_details"`
}
```

### HPC Module

#### Transaction: Submit Job

```go
type MsgSubmitJob struct {
    Owner            string              `json:"owner"`
    Manifest         []byte              `json:"manifest"`
    Resources        HPCResources        `json:"resources"`
    EncryptedInputs  *EncryptedPayloadEnvelope `json:"encrypted_inputs,omitempty"`
}
```

## Encryption Envelope Specification

### Envelope Format

All sensitive data on VirtEngine uses the Encrypted Payload Envelope format:

```json
{
  "version": "1.0",
  "algorithm_id": "X25519-XSalsa20-Poly1305",
  "recipient_key_ids": ["fingerprint1", "fingerprint2"],
  "nonce": "base64_encoded_24_bytes",
  "ciphertext": "base64_encoded_ciphertext",
  "sender_pub_key": "base64_encoded_32_bytes",
  "sender_signature": "base64_encoded_signature",
  "encrypted_keys": ["base64_dek1", "base64_dek2"],
  "metadata": {
    "content_type": "application/json",
    "created_at": "2026-01-24T12:00:00Z"
  }
}
```

### Supported Algorithms

| Algorithm ID | Description | Key Size | Nonce Size |
|--------------|-------------|----------|------------|
| `X25519-XSalsa20-Poly1305` | NaCl box | 32 bytes | 24 bytes |

### Creating an Envelope (Go)

```go
import "github.com/virtengine/virtengine/x/encryption/crypto"

// Generate sender key pair
senderKP, _ := crypto.GenerateKeyPair()

// Create envelope for single recipient
envelope, err := crypto.CreateEnvelope(
    plaintext,
    recipientPublicKey,
    senderKP,
)

// Create envelope for multiple recipients
envelope, err := crypto.CreateMultiRecipientEnvelope(
    plaintext,
    [][]byte{validator1PubKey, validator2PubKey},
    senderKP,
)
```

### Creating an Envelope (TypeScript)

```typescript
import { createEnvelope, generateKeyPair } from '@virtengine/crypto';

const senderKeyPair = await generateKeyPair();

const envelope = await createEnvelope({
  plaintext: Buffer.from(JSON.stringify(data)),
  recipientPublicKey: recipientKey,
  senderKeyPair,
});
```

### Decrypting an Envelope

```go
import "github.com/virtengine/virtengine/x/encryption/crypto"

algorithm, _ := crypto.GetAlgorithm(envelope.AlgorithmID)

plaintext, err := algorithm.Decrypt(
    envelope.Ciphertext,
    envelope.Nonce,
    envelope.SenderPubKey,
    recipientPrivateKey,
)
```

## Event Schema

VirtEngine emits events for all significant state changes. Subscribe to these via WebSocket or query from blocks.

### VEID Events

```json
// veid.identity_uploaded
{
  "type": "veid.identity_uploaded",
  "attributes": {
    "owner": "virtengine1...",
    "scopes_hash": "abc123...",
    "scope_count": "3",
    "client_key_fingerprint": "def456..."
  }
}

// veid.score_computed
{
  "type": "veid.score_computed",
  "attributes": {
    "owner": "virtengine1...",
    "score": "85",
    "validator": "virtenginevaloper1...",
    "block_height": "12345"
  }
}
```

### Market Events

```json
// market.order_created
{
  "type": "market.order_created",
  "attributes": {
    "order_id": "order_123",
    "customer": "virtengine1...",
    "offering_id": "offering_456",
    "state": "PENDING"
  }
}

// market.bid_placed
{
  "type": "market.bid_placed",
  "attributes": {
    "order_id": "order_123",
    "bid_id": "bid_789",
    "provider": "virtengine1provider...",
    "price": "1000uve"
  }
}

// market.order_allocated
{
  "type": "market.order_allocated",
  "attributes": {
    "order_id": "order_123",
    "bid_id": "bid_789",
    "provider": "virtengine1provider..."
  }
}
```

### HPC Events

```json
// hpc.job_submitted
{
  "type": "hpc.job_submitted",
  "attributes": {
    "job_id": "job_abc",
    "owner": "virtengine1...",
    "resources": "{\"cpus\":4,\"memory\":8589934592}"
  }
}

// hpc.job_scheduled
{
  "type": "hpc.job_scheduled",
  "attributes": {
    "job_id": "job_abc",
    "provider": "virtengine1provider...",
    "scheduled_at": "2026-01-24T12:00:00Z"
  }
}
```

## Transaction Examples

### Complete Identity Upload Flow

```bash
# 1. Generate capture data (on approved client)
# Client captures document + selfie, generates salt

# 2. Encrypt scopes for validators
virtengine tx encryption get-validator-keys --output validator_keys.json

# 3. Create encrypted envelope (using SDK)
# See SDK examples below

# 4. Submit identity upload
virtengine tx veid upload-scopes \
    --scopes '{"document":{"envelope":...},"selfie":{"envelope":...}}' \
    --salt $(cat salt.hex) \
    --client-signature $(cat client_sig.hex) \
    --from user

# 5. Wait for scoring
virtengine query veid score $(virtengine keys show user -a) --watch

# 6. Query final score
virtengine query veid score $(virtengine keys show user -a)
```

### Complete Marketplace Order Flow

```bash
# 1. Browse offerings
virtengine query market offerings --region us-east --limit 10

# 2. Get offering details
virtengine query market offering offering_123

# 3. Create order
virtengine tx market create-order \
    --offering offering_123 \
    --quantity 1 \
    --config '{"duration_hours":24}' \
    --from customer

# 4. Wait for bids (provider daemons auto-bid)
virtengine query market bids --order order_456

# 5. Check allocation
virtengine query market order order_456

# 6. Once allocated, access provisioned resource
# Provider daemon provisions and updates status
```

## Local Devnet Setup

### Using Docker Compose

```bash
# Start full stack
docker compose -f docker-compose.yaml up -d

# Components started:
# - Chain node (port 26657, 9090, 1317)
# - Provider daemon (port 8443)
# - Waldur (port 8080)
# - Portal (port 3000)

# View logs
docker compose logs -f chain

# Stop
docker compose down
```

### Manual Setup

```bash
# Terminal 1: Start chain
make localnet-start

# Terminal 2: Start provider daemon
./cmd/provider-daemon/provider-daemon start \
    --node http://localhost:26657 \
    --config provider-config.yaml

# Terminal 3: Start portal
cd lib/portal && pnpm dev
```

### Test Accounts

| Account | Address | Mnemonic |
|---------|---------|----------|
| admin | virtengine1admin... | `abandon abandon ... abandon admin` |
| customer | virtengine1customer... | `abandon abandon ... abandon customer` |
| provider | virtengine1provider... | `abandon abandon ... abandon provider` |

> ⚠️ These are **test-only** accounts. Never use in production.

## SDK Reference

### TypeScript SDK

```typescript
import { VirtEngineClient, Wallet } from '@virtengine/sdk';

// Initialize client
const client = new VirtEngineClient({
  nodeUrl: 'http://localhost:26657',
  chainId: 'virtengine-localnet-1',
});

// Create wallet from mnemonic
const wallet = await Wallet.fromMnemonic(mnemonic);

// Query identity score
const score = await client.veid.getScore(address);
console.log(`Identity score: ${score.score}`);

// Create marketplace order
const order = await client.market.createOrder({
  offeringId: 'offering_123',
  quantity: 1,
  config: { region: 'us-east' },
  wallet,
});
console.log(`Order ID: ${order.orderId}`);
```

### Go SDK

```go
import (
    "github.com/virtengine/virtengine/client/go"
    "github.com/virtengine/virtengine/x/market/types"
)

// Create client
client, _ := virtengine.NewClient(virtengine.Config{
    NodeURL: "http://localhost:26657",
    ChainID: "virtengine-localnet-1",
})

// Query identity score
score, _ := client.VEID.GetScore(ctx, address)
fmt.Printf("Identity score: %d\n", score.Score)

// Create order
order, _ := client.Market.CreateOrder(ctx, &types.MsgCreateOrder{
    Customer:   customerAddr,
    OfferingID: "offering_123",
    Quantity:   1,
})
```

### Python SDK

```python
from virtengine import VirtEngineClient

# Initialize client
client = VirtEngineClient(
    node_url="http://localhost:26657",
    chain_id="virtengine-localnet-1",
)

# Query identity score
score = client.veid.get_score(address)
print(f"Identity score: {score.score}")

# For ML integration
from virtengine.ml import IdentityScorer

scorer = IdentityScorer(model_path="veid_scorer_v1.0.0.h5")
features = scorer.extract_features(document_image, selfie_image)
score = scorer.predict(features)
```

## Security Guidance

### Private Key Handling

> ⚠️ **NEVER** expose private keys in logs, errors, or API responses

```go
// BAD - logs private key
log.Printf("Using key: %s", privateKey)

// GOOD - logs key fingerprint only
log.Printf("Using key: %s", computeFingerprint(publicKey))
```

### Sensitive Data Encryption

All sensitive data MUST be encrypted before storage on-chain:

```go
// BAD - plaintext sensitive data
msg := &types.MsgCreateOrder{
    Config: map[string]string{"password": "secret123"},
}

// GOOD - encrypted sensitive data
envelope, _ := crypto.CreateEnvelope(sensitiveData, recipientKey, senderKP)
msg := &types.MsgCreateOrder{
    EncryptedConfig: envelope,
}
```

### Safe Logging Practices

```go
import "github.com/virtengine/virtengine/pkg/observability"

// Use structured logging with redaction
logger := observability.NewLogger(observability.Config{
    RedactPatterns: []string{
        "password", "secret", "key", "token", "mnemonic",
    },
})

// Sensitive fields are automatically redacted
logger.Info("Processing request",
    "user", userID,
    "api_key", apiKey, // Will be redacted
)
```

### Input Validation

Always validate inputs before processing:

```go
func validateOrder(order *types.MsgCreateOrder) error {
    if order.Quantity <= 0 {
        return ErrInvalidQuantity
    }
    if order.Quantity > MaxOrderQuantity {
        return ErrQuantityTooLarge
    }
    if !isValidOfferingID(order.OfferingID) {
        return ErrInvalidOfferingID
    }
    return nil
}
```

---

## Support

- Documentation: [docs.virtengine.com](https://docs.virtengine.com)
- Discord: [discord.gg/virtengine](https://discord.gg/virtengine)
- GitHub Issues: [github.com/virtengine/virtengine/issues](https://github.com/virtengine/virtengine/issues)
