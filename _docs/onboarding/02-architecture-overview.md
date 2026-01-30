# Architecture Overview

This guide provides a high-level understanding of VirtEngine's architecture for new developers.

## Table of Contents

1. [System Overview](#system-overview)
2. [Layer Architecture](#layer-architecture)
3. [Blockchain Modules](#blockchain-modules)
4. [Off-Chain Services](#off-chain-services)
5. [Data Flow Patterns](#data-flow-patterns)
6. [Security Architecture](#security-architecture)
7. [Key Concepts](#key-concepts)

---

## System Overview

VirtEngine is a Cosmos SDK-based hybrid blockchain platform combining:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         VIRTENGINE ECOSYSTEM                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      CLIENT APPLICATIONS                         │   │
│  │   VE Portal │ Mobile App │ CLI/SDK │ Provider Dashboard          │   │
│  └──────────────────────────────┬──────────────────────────────────┘   │
│                                 │                                       │
│                                 ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                      API GATEWAY LAYER                           │   │
│  │         REST/gRPC Endpoints │ Rate Limiting │ TLS                │   │
│  └──────────────────────────────┬──────────────────────────────────┘   │
│                                 │                                       │
│                                 ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    BLOCKCHAIN LAYER                              │   │
│  │  ┌─────────┐ ┌─────┐ ┌─────────┐ ┌────────┐ ┌────────┐          │   │
│  │  │  VEID   │ │ MFA │ │Encrypt  │ │ Market │ │ Escrow │ + more   │   │
│  │  └─────────┘ └─────┘ └─────────┘ └────────┘ └────────┘          │   │
│  │                    Cosmos SDK Base Modules                       │   │
│  │                    CometBFT Consensus                            │   │
│  └──────────────────────────────┬──────────────────────────────────┘   │
│                                 │                                       │
│                                 ▼                                       │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                   OFF-CHAIN SERVICES                             │   │
│  │  Provider Daemon │ ML Inference │ Waldur │ Benchmarking          │   │
│  └─────────────────────────────────────────────────────────────────┘   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

### Core Design Principles

| Principle | Description |
|-----------|-------------|
| **Privacy** | Sensitive data is never stored in plaintext on-chain |
| **Decentralization** | No single point of failure for identity or order fulfillment |
| **Verifiability** | Validators independently verify identity scores |
| **Determinism** | All consensus operations are reproducible |
| **Modularity** | Clean separation between on-chain and off-chain logic |

---

## Layer Architecture

### 1. Client Layer

Applications that interact with VirtEngine:

| Client | Technology | Purpose |
|--------|------------|---------|
| VE Portal | React | Web-based user interface |
| Mobile App | React Native | Identity capture with approved signing |
| CLI | Go | Command-line operations |
| SDKs | Go/TS/Python | Programmatic integration |

### 2. API Gateway Layer

Handles external API requests:

- **REST API** (port 1317) - Cosmos LCD endpoints
- **gRPC** (port 9090) - High-performance queries
- **Tendermint RPC** (port 26657) - Node operations
- **Rate Limiting** - Kong-based protection
- **TLS Termination** - Encrypted transport

### 3. Blockchain Layer

The core consensus layer built on Cosmos SDK:

```
┌─────────────────────────────────────────────────────────────────┐
│                     CUSTOM MODULES (x/)                         │
│  veid │ mfa │ encryption │ market │ escrow │ roles │ hpc       │
├─────────────────────────────────────────────────────────────────┤
│                    COSMOS SDK BASE MODULES                      │
│  auth │ bank │ staking │ gov │ slashing │ distribution         │
├─────────────────────────────────────────────────────────────────┤
│                      COMETBFT CONSENSUS                         │
│         Tendermint BFT │ PoS Validators │ State Machine         │
└─────────────────────────────────────────────────────────────────┘
```

### 4. Off-Chain Services Layer

Services that extend blockchain functionality:

- **Provider Daemon** - Automated bidding and provisioning
- **ML Inference** - TensorFlow-based identity scoring
- **Waldur Services** - External resource management
- **Benchmarking Daemon** - Provider performance metrics

---

## Blockchain Modules

### Module Overview

| Module | Path | Purpose |
|--------|------|---------|
| VEID | `x/veid` | Identity verification with ML scoring |
| MFA | `x/mfa` | Multi-factor authentication gating |
| Encryption | `x/encryption` | Public-key encryption primitives |
| Market | `x/market` | Orders, bids, leases lifecycle |
| Escrow | `x/escrow` | Payment holds and settlement |
| Roles | `x/roles` | Role-based access control |
| HPC | `x/hpc` | High-performance computing jobs |
| Provider | `x/provider` | Provider registration |
| Benchmark | `x/benchmark` | Provider performance metrics |

### Module Interaction Diagram

```
                            ┌─────────────┐
                            │   Cosmos    │
                            │  Base SDK   │
                            └──────┬──────┘
                                   │
            ┌──────────────────────┼──────────────────────┐
            │                      │                      │
            ▼                      ▼                      ▼
    ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
    │     Roles     │◄────►│   Encryption  │◄────►│     Audit     │
    └───────┬───────┘      └───────┬───────┘      └───────────────┘
            │                      │
  ┌─────────┼─────────┬────────────┴────────┬───────────────────┐
  │         │         │                     │                   │
  ▼         ▼         ▼                     ▼                   ▼
┌─────┐ ┌───────┐ ┌───────┐           ┌──────────┐       ┌──────────┐
│ MFA │ │ VEID  │ │ Cert  │           │  Market  │       │ Provider │
└──┬──┘ └───┬───┘ └───────┘           └────┬─────┘       └────┬─────┘
   │        │                              │                  │
   └────────┴──────────────────────────────┴──────────────────┘
                                │
                                ▼
                    ┌───────────────────────┐
                    │   Escrow & Settlement │
                    └───────────────────────┘
```

### VEID Module (Identity)

Handles decentralized identity verification:

- **Identity Scopes**: Encrypted document + selfie data
- **ML Scoring**: 0-100 verification score
- **Validator Consensus**: Score verification by validators
- **Score Thresholds**: Access levels based on score

```go
// Key types in x/veid
type IdentityScore struct {
    Owner      string
    Score      uint8    // 0-100
    ComputedAt time.Time
    ExpiresAt  time.Time
}

type EncryptedScope struct {
    Envelope   EncryptionEnvelope
    ScopeType  ScopeType  // DOCUMENT, SELFIE, etc.
}
```

### Market Module (Marketplace)

Manages the cloud computing marketplace:

- **Orders**: Customer requests for compute resources
- **Bids**: Provider responses to orders
- **Leases**: Matched order-bid pairs
- **Lifecycle**: PENDING → OPEN → MATCHED → ACTIVE → CLOSED

```go
// Order lifecycle states
const (
    OrderStateOpen    = "open"
    OrderStateMatched = "matched"
    OrderStateActive  = "active"
    OrderStateClosed  = "closed"
)
```

### Encryption Module

Provides cryptographic primitives:

- **Algorithm**: X25519-XSalsa20-Poly1305 (NaCl box)
- **Envelope Format**: Standardized encrypted payload structure
- **Key Management**: Recipient key registration and lookup

---

## Off-Chain Services

### Provider Daemon (`pkg/provider_daemon/`)

The provider daemon bridges on-chain orders to infrastructure:

```
┌─────────────────────────────────────────────────────────────────┐
│                      PROVIDER DAEMON                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────┐   ┌───────────────┐   ┌──────────────┐       │
│  │  Bid Engine  │   │ Key Manager   │   │ Usage Meter  │       │
│  └──────┬───────┘   └───────────────┘   └──────┬───────┘       │
│         │                                       │               │
│         ▼                                       ▼               │
│  ┌───────────────────────────────────────────────────────────┐ │
│  │                    ADAPTERS LAYER                          │ │
│  │  Kubernetes │ OpenStack │ AWS │ Azure │ VMware │ SLURM     │ │
│  └───────────────────────────────────────────────────────────┘ │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

**Components**:

| Component | File | Purpose |
|-----------|------|---------|
| Bid Engine | `bid_engine.go` | Watches orders, computes pricing, submits bids |
| Key Manager | `key_manager.go` | Provider key storage (hardware support) |
| Manifest Parser | `manifest.go` | Validates deployment manifests |
| Usage Meter | `usage_meter.go` | Collects metrics, submits usage records |

**Adapters** (workload lifecycle management):

| Adapter | Backend |
|---------|---------|
| Kubernetes | K8s namespace isolation |
| OpenStack | VM provisioning via Waldur |
| AWS | EC2/VPC/EBS via Waldur |
| Azure | Azure VMs via Waldur |
| VMware | vSphere integration |
| SLURM | HPC job scheduling |

### ML Inference (`pkg/inference/`)

TensorFlow-based scoring for identity verification:

```go
// Determinism is CRITICAL for consensus
type DeterminismConfig struct {
    ForceCPU         bool   // true - no GPU variance
    RandomSeed       int64  // fixed seed (42)
    DeterministicOps bool   // true - TF deterministic ops
}
```

> **Important**: All ML scoring must be deterministic for validators to reach consensus.

---

## Data Flow Patterns

### Identity Verification Flow

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│ Client  │───►│ Encrypt │───►│  VEID   │───►│Validator│───►│Consensus│
│ Capture │    │ Module  │    │ Module  │    │ ML Eval │    │  Vote   │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
     │              │              │              │              │
     │ 1. Capture   │ 2. Encrypt   │ 3. Store     │ 4. Decrypt   │ 5. Vote
     │ doc+selfie   │ payload to   │ encrypted    │ + ML score   │ on score
     │ + sign       │ validator    │ scope refs   │ (0-100)      │
     ▼              ▼              ▼              ▼              ▼
[salt+client    [envelope:      [state:        [score +       [block
 sig+user sig]   pubkey,        scope_refs,    status in      includes
                 cipher,        timestamps]    proposed       final
                 nonce]                        block]         score]
```

### Marketplace Order Flow

```
┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐
│Customer │───►│ Market  │───►│ Escrow  │───►│Provider │───►│ Lease   │
│  Order  │    │ Module  │    │ Module  │    │ Daemon  │    │ Active  │
└─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘
     │              │              │              │              │
     │ 1. Create    │ 2. Store     │ 3. Hold      │ 4. Bid +     │ 5. Lease
     │ order with   │ order (enc   │ payment      │ provision    │ created
     │ encrypted    │ fields)      │ in escrow    │ workload     │ + usage
     │ details      │              │              │              │ start
     ▼              ▼              ▼              ▼              ▼
[MFA required   [encrypted     [tokens        [signed        [on-chain
 if high-value]  order config]  locked]        bid tx]        lease ref]
```

---

## Security Architecture

### Encryption Envelope Format

All sensitive data uses the standardized envelope format:

```
┌─────────────────────────────────────────────────────────────────┐
│                    ENCRYPTED PAYLOAD ENVELOPE                    │
├─────────────────────────────────────────────────────────────────┤
│  Header                                                         │
│  ├─ version: uint8                                              │
│  ├─ algorithm_id: uint8 (X25519-XSalsa20-Poly1305)             │
│  ├─ recipient_pubkey: bytes[32]                                 │
│  └─ sender_pubkey: bytes[32]                                    │
├─────────────────────────────────────────────────────────────────┤
│  Cryptographic Material                                         │
│  ├─ nonce: bytes[24]                                            │
│  ├─ ciphertext: bytes[variable]                                 │
│  └─ auth_tag: bytes[16] (AEAD)                                  │
├─────────────────────────────────────────────────────────────────┤
│  Metadata (plaintext)                                           │
│  ├─ content_type: string                                        │
│  ├─ created_at: timestamp                                       │
│  └─ sender_signature: bytes[64]                                 │
└─────────────────────────────────────────────────────────────────┘
```

### MFA Gating

Sensitive transactions require multi-factor authentication:

| Transaction | Risk Level | Required Factors |
|-------------|------------|------------------|
| AccountRecovery | Critical | VEID + FIDO2 + SMS |
| KeyRotation | Critical | VEID + FIDO2 |
| ProviderRegistration | High | VEID (score ≥70) + FIDO2 |
| LargeWithdrawal | High | VEID + FIDO2 |

### Trust Boundaries

| Component | Trust Level | Notes |
|-----------|-------------|-------|
| Blockchain consensus | Highest | Core trusted compute |
| Validator nodes | High | Decrypts identity data |
| Provider daemon | Medium | Signed messages |
| Client applications | Untrusted | All input validated |

---

## Key Concepts

### Cosmos SDK Patterns

Understanding these Cosmos SDK concepts is essential:

| Concept | Description |
|---------|-------------|
| **Module** | Self-contained blockchain feature (`x/module`) |
| **Keeper** | Module's data access layer |
| **Msg** | Transaction message type |
| **Query** | Read-only state query |
| **Genesis** | Initial state configuration |
| **Ante Handler** | Pre-transaction validation |

### VirtEngine-Specific Concepts

| Concept | Description |
|---------|-------------|
| **Identity Score** | 0-100 ML-computed trust score |
| **Encrypted Scope** | Encrypted identity data bundle |
| **Lease** | Active resource allocation |
| **Usage Record** | Provider-submitted resource consumption |
| **Workload State** | Infrastructure provisioning status |

---

## Related Documentation

- [Full Architecture Document](../architecture.md) - Comprehensive system design
- [Module Development Guide](./06-module-development.md) - Building blockchain modules
- [Developer Guide](../developer-guide.md) - API reference and examples
- [Testing Guide](./04-testing-guide.md) - Testing architecture
