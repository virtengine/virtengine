# VirtEngine Architecture Overview

**Module Duration:** 4 hours  
**Level:** Foundational  
**Prerequisites:** Basic blockchain concepts, familiarity with distributed systems  
**Version:** 1.0.0  
**Last Updated:** 2025-01-24

---

## Table of Contents

1. [Learning Objectives](#learning-objectives)
2. [Introduction to VirtEngine](#introduction-to-virtengine)
3. [Cosmos SDK Fundamentals](#cosmos-sdk-fundamentals)
4. [VirtEngine Module System](#virtengine-module-system)
5. [Blockchain Fundamentals](#blockchain-fundamentals)
6. [System Architecture](#system-architecture)
7. [Data Flow Patterns](#data-flow-patterns)
8. [Key Takeaways](#key-takeaways)
9. [Self-Assessment Questions](#self-assessment-questions)
10. [Practical Exercises](#practical-exercises)
11. [References](#references)

---

## Learning Objectives

By the end of this module, you will be able to:

- [ ] **Explain** the core architecture of VirtEngine and its relationship to Cosmos SDK
- [ ] **Identify** the key modules in the `x/` directory and their responsibilities
- [ ] **Describe** how transactions flow through the system from client to consensus
- [ ] **Understand** the separation between on-chain and off-chain components
- [ ] **Navigate** the codebase structure and locate key components
- [ ] **Explain** the role of validators, providers, and users in the ecosystem
- [ ] **Diagram** the interaction between core system components

---

## Introduction to VirtEngine

### What is VirtEngine?

VirtEngine is a **Cosmos SDK-based blockchain** designed for decentralized cloud computing with ML-powered identity verification. It combines:

| Capability | Description |
|------------|-------------|
| **VEID (Identity)** | ML-powered identity scoring (0-100) with validator consensus |
| **MFA Gating** | On-chain multi-factor authentication for sensitive transactions |
| **Encrypted Storage** | Public-key encryption for all sensitive on-chain data |
| **Cloud Marketplace** | Decentralized marketplace for cloud computing resources |
| **Provider Network** | Automated bidding, provisioning, and usage recording |
| **HPC Support** | Distributed computing via SLURM clusters |

### Design Principles

VirtEngine architecture prioritizes:

1. **Privacy** - Sensitive data never stored in plaintext on the public ledger
2. **Decentralization** - No single point of failure for verification or fulfillment
3. **Verifiability** - Validators independently verify all state transitions
4. **Compliance** - Role-based access control with complete audit trails
5. **Determinism** - All on-chain operations must be reproducible across validators

### Repository Structure

```
virtengine/
├── app/           → Cosmos SDK app wiring, ante handlers, genesis
├── x/             → Custom blockchain modules (core business logic)
│   ├── veid/      → Identity verification module
│   ├── mfa/       → Multi-factor authentication module
│   ├── encryption/→ Encryption primitives
│   ├── market/    → Marketplace orders, bids, leases
│   ├── escrow/    → Payment holds and settlement
│   ├── roles/     → RBAC and account states
│   └── hpc/       → High-performance computing
├── pkg/           → Off-chain services and libraries
├── cmd/           → CLI binaries and entry points
├── ml/            → Python ML pipelines
├── tests/         → Integration and E2E tests
└── _docs/         → Documentation
```

---

## Cosmos SDK Fundamentals

### What is Cosmos SDK?

The Cosmos SDK is a framework for building application-specific blockchains. Key concepts:

#### 1. Application-Specific Blockchains

Unlike general-purpose blockchains (Ethereum), Cosmos SDK chains are purpose-built:

```
┌─────────────────────────────────────────────────────┐
│                  Traditional Approach                │
│  ┌─────────────────────────────────────────────┐   │
│  │         Smart Contracts (Solidity)           │   │
│  └─────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────┐   │
│  │         General-Purpose VM (EVM)             │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│                  Cosmos SDK Approach                 │
│  ┌─────────────────────────────────────────────┐   │
│  │    Application Logic (Native Go Modules)    │   │
│  └─────────────────────────────────────────────┘   │
│  ┌─────────────────────────────────────────────┐   │
│  │         Tendermint/CometBFT Consensus        │   │
│  └─────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────┘
```

#### 2. Module Architecture

Cosmos SDK applications are composed of modules:

```go
// Each module implements the AppModule interface
type AppModule interface {
    Name() string
    RegisterServices(cfg module.Configurator)
    InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) []abci.ValidatorUpdate
    ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage
}
```

#### 3. Message-Based State Transitions

State changes happen through typed messages:

```go
// Example: Creating a marketplace order
type MsgCreateOrder struct {
    Owner       string
    GroupSpec   GroupSpec
    Deposit     sdk.Coin
}

func (msg MsgCreateOrder) ValidateBasic() error {
    // Validate message fields before processing
    if _, err := sdk.AccAddressFromBech32(msg.Owner); err != nil {
        return sdkerrors.ErrInvalidAddress
    }
    return nil
}
```

### Core Cosmos SDK Modules

VirtEngine builds on these base modules:

| Module | Purpose | Usage in VirtEngine |
|--------|---------|---------------------|
| `auth` | Account management | User and validator accounts |
| `bank` | Token transfers | VE token transfers, payments |
| `staking` | Validator staking | Validator set management |
| `gov` | Governance proposals | Module parameter updates |
| `slashing` | Validator penalties | Misbehavior punishment |
| `distribution` | Reward distribution | Staking rewards |
| `params` | Parameter storage | Module configuration |

### CometBFT Consensus

VirtEngine uses CometBFT (formerly Tendermint) for consensus:

```
┌──────────────────────────────────────────────────────────────┐
│                    CONSENSUS ROUND                            │
├──────────────────────────────────────────────────────────────┤
│                                                               │
│   ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│   │ Propose │───▶│Prevote  │───▶│Precommit│───▶│ Commit  │  │
│   └─────────┘    └─────────┘    └─────────┘    └─────────┘  │
│                                                               │
│   - Leader       - 2/3+ votes   - 2/3+ votes  - Block        │
│     proposes       to accept      to finalize   finalized    │
│     block                                                     │
│                                                               │
└──────────────────────────────────────────────────────────────┘
```

**Key Properties:**
- **Instant Finality** - Blocks are final once committed (no reorganizations)
- **Byzantine Fault Tolerance** - Tolerates up to 1/3 malicious validators
- **Deterministic** - All validators produce identical state transitions

---

## VirtEngine Module System

### Module Architecture Pattern

All VirtEngine modules follow the **Keeper Pattern**:

```go
// x/market/keeper/keeper.go - Standard pattern
type IKeeper interface {
    // Public interface for cross-module calls
    CreateOrder(ctx sdk.Context, gid dtypes.GroupID, spec dtypesBeta.GroupSpec) (types.Order, error)
    GetOrder(ctx sdk.Context, id mv1.OrderID) (types.Order, bool)
    WithOrders(ctx sdk.Context, fn func(types.Order) bool)
}

type Keeper struct {
    cdc       codec.BinaryCodec    // For serialization
    skey      storetypes.StoreKey  // For state storage
    authority string               // x/gov module account for params
}
```

### Core VirtEngine Modules

#### 1. VEID Module (`x/veid`)

**Purpose:** Identity verification with ML-powered scoring

```
┌─────────────────────────────────────────────────────────────┐
│                    VEID IDENTITY FLOW                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  User          Approved App        Blockchain    Validators  │
│   │                │                   │             │       │
│   │──Upload───────▶│                   │             │       │
│   │  ID Doc+Selfie │                   │             │       │
│   │                │──Sign & Submit───▶│             │       │
│   │                │   (Encrypted)     │             │       │
│   │                │                   │──Decrypt────▶│      │
│   │                │                   │  & Score     │      │
│   │                │                   │◀──Consensus──│      │
│   │                │                   │   Score      │      │
│   │◀────────────────────Score (0-100)──│             │       │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Key Concepts:**
- Identity scopes are encrypted before on-chain storage
- Validators decrypt and run ML scoring during consensus
- Scores range from 0-100 and are stored on-chain
- Three required signatures: client, user, salt binding

#### 2. MFA Module (`x/mfa`)

**Purpose:** Gate sensitive transactions with multi-factor authentication

```go
// MFA-gated operations
const (
    MFATypeNone         = 0  // No MFA required
    MFATypeTOTP         = 1  // Time-based OTP
    MFATypeWebAuthn     = 2  // Hardware security key
    MFATypeBiometric    = 3  // Biometric verification
)
```

**Sensitive Operations Requiring MFA:**
- Large fund transfers (above threshold)
- Key rotation
- Provider deregistration
- Validator operations

#### 3. Encryption Module (`x/encryption`)

**Purpose:** Public-key encryption for sensitive on-chain data

```go
// Envelope structure for all encrypted payloads
type EncryptionEnvelope struct {
    RecipientFingerprint string  // Validator's key fingerprint
    Algorithm            string  // "X25519-XSalsa20-Poly1305"
    Ciphertext           []byte
    Nonce                []byte
}
```

**Encryption Flow:**
1. Client generates ephemeral X25519 keypair
2. Derives shared secret with recipient public key
3. Encrypts payload with XSalsa20-Poly1305
4. Stores envelope on-chain

#### 4. Market Module (`x/market`)

**Purpose:** Decentralized marketplace for cloud resources

```
┌─────────────────────────────────────────────────────────────┐
│                  MARKETPLACE FLOW                            │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│   Order Created ──▶ Bids Received ──▶ Bid Accepted          │
│        │                  │                │                 │
│        ▼                  ▼                ▼                 │
│   ┌─────────┐      ┌───────────┐    ┌───────────┐          │
│   │ Escrow  │      │ Provider  │    │  Lease    │          │
│   │ Deposit │      │ Bids      │    │ Created   │          │
│   └─────────┘      └───────────┘    └───────────┘          │
│                                           │                  │
│                                           ▼                  │
│                                    ┌───────────┐            │
│                                    │ Resource  │            │
│                                    │ Deployed  │            │
│                                    └───────────┘            │
└─────────────────────────────────────────────────────────────┘
```

#### 5. Escrow Module (`x/escrow`)

**Purpose:** Payment holds and settlement

**State Machine:**
```
Open ──▶ Active ──▶ Closed
  │         │
  └─────────┴──▶ Overdrawn (insufficient funds)
```

#### 6. Roles Module (`x/roles`)

**Purpose:** Role-based access control (RBAC)

| Role | Permissions |
|------|-------------|
| `user` | Basic marketplace access |
| `verified_user` | Access to verified-only resources |
| `provider` | Submit bids, manage infrastructure |
| `validator` | Participate in consensus, score identities |
| `admin` | System administration |

#### 7. HPC Module (`x/hpc`)

**Purpose:** High-performance computing workloads

Manages:
- SLURM cluster integration
- Job submission and tracking
- Resource allocation
- Usage metering

### Module Dependency Graph

```
                              ┌─────────────┐
                              │   Cosmos    │
                              │  Base SDK   │
                              │ (auth/bank) │
                              └──────┬──────┘
                                     │
           ┌─────────────────────────┼─────────────────────────┐
           │                         │                         │
           ▼                         ▼                         ▼
    ┌─────────────┐          ┌─────────────┐          ┌─────────────┐
    │  encryption │◀─────────│    veid     │─────────▶│    roles    │
    └─────────────┘          └──────┬──────┘          └──────┬──────┘
           │                        │                        │
           │                        ▼                        │
           │                 ┌─────────────┐                 │
           └────────────────▶│     mfa     │◀────────────────┘
                             └──────┬──────┘
                                    │
                                    ▼
                             ┌─────────────┐
                             │   market    │◀──────┐
                             └──────┬──────┘       │
                                    │              │
                    ┌───────────────┼───────────┐  │
                    │               │           │  │
                    ▼               ▼           ▼  │
             ┌───────────┐  ┌───────────┐  ┌───────────┐
             │  escrow   │  │  provider │  │    hpc    │
             └───────────┘  └───────────┘  └───────────┘
```

---

## Blockchain Fundamentals

### Transaction Lifecycle

```
┌──────────────────────────────────────────────────────────────────────┐
│                    TRANSACTION LIFECYCLE                              │
├──────────────────────────────────────────────────────────────────────┤
│                                                                       │
│  1. CREATE          2. SIGN           3. BROADCAST       4. MEMPOOL  │
│  ┌─────────┐       ┌─────────┐       ┌─────────┐       ┌─────────┐  │
│  │ Build   │──────▶│ Sign w/ │──────▶│ Submit  │──────▶│ Pending │  │
│  │   Msg   │       │  Key    │       │ to Node │       │  Pool   │  │
│  └─────────┘       └─────────┘       └─────────┘       └─────────┘  │
│                                                              │        │
│                                                              ▼        │
│  7. FINALIZE        6. CONSENSUS     5. VALIDATE      ┌─────────┐   │
│  ┌─────────┐       ┌─────────┐       ┌─────────┐      │ Include │   │
│  │ State   │◀──────│  2/3+   │◀──────│CheckTx +│◀─────│ in Block│   │
│  │ Updated │       │ Commit  │       │DeliverTx│      └─────────┘   │
│  └─────────┘       └─────────┘       └─────────┘                     │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘
```

### State Storage

VirtEngine uses an IAVL tree for state storage:

```go
// Store keys define module state namespaces
const (
    ModuleName = "market"
    StoreKey   = ModuleName
    
    // Prefix keys for different data types
    OrderKeyPrefix    = 0x01
    BidKeyPrefix      = 0x02
    LeaseKeyPrefix    = 0x03
)

// Example: Storing an order
func (k Keeper) SetOrder(ctx sdk.Context, order types.Order) {
    store := ctx.KVStore(k.storeKey)
    key := types.OrderKey(order.OrderID)
    value := k.cdc.MustMarshal(&order)
    store.Set(key, value)
}
```

### Ante Handlers

Ante handlers validate transactions before execution:

```go
// app/ante.go - Transaction validation pipeline
func NewAnteHandler(options HandlerOptions) sdk.AnteHandler {
    return sdk.ChainAnteDecorators(
        ante.NewSetUpContextDecorator(),           // Set up context
        ante.NewValidateBasicDecorator(),          // Validate message format
        ante.NewTxTimeoutHeightDecorator(),        // Check timeout
        ante.NewValidateMemoDecorator(options.AccountKeeper),
        ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
        ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper),
        ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
        ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
        // VirtEngine-specific
        veidante.NewVerificationDecorator(options.VEIDKeeper),
        mfaante.NewMFAGatingDecorator(options.MFAKeeper),
    )
}
```

### Genesis and Upgrades

#### Genesis State

```go
// types/genesis.go
type GenesisState struct {
    Params  Params   `json:"params"`
    Orders  []Order  `json:"orders"`
    Bids    []Bid    `json:"bids"`
    Leases  []Lease  `json:"leases"`
}

func DefaultGenesisState() *GenesisState {
    return &GenesisState{
        Params: DefaultParams(),
    }
}

func (gs GenesisState) Validate() error {
    // Validate all genesis state
    if err := gs.Params.Validate(); err != nil {
        return err
    }
    // ... validate orders, bids, leases
    return nil
}
```

#### Upgrade Handlers

```go
// upgrades/v2/upgrade.go
func CreateUpgradeHandler(
    mm *module.Manager,
    configurator module.Configurator,
    keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
    return func(ctx sdk.Context, plan upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
        // Perform state migrations
        // Update parameters
        // Initialize new modules
        return mm.RunMigrations(ctx, configurator, fromVM)
    }
}
```

---

## System Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              VIRTENGINE ECOSYSTEM                            │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         CLIENT LAYER                                 │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │   │
│  │  │  VE Portal   │  │ Mobile App   │  │   CLI/SDK    │              │   │
│  │  │  (React)     │  │ (Approved)   │  │  (Go/TS)     │              │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘              │   │
│  └─────────┼─────────────────┼─────────────────┼───────────────────────┘   │
│            │                 │                 │                            │
│            ▼                 ▼                 ▼                            │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                         API GATEWAY LAYER                            │   │
│  │       TLS-Encrypted REST/gRPC Endpoints (LCD/gRPC)                  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      BLOCKCHAIN LAYER (Cosmos SDK)                   │   │
│  │                                                                      │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐       │   │
│  │  │  VEID   │ │   MFA   │ │ Encrypt │ │ Market  │ │ Escrow  │       │   │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘ └─────────┘       │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │              CometBFT CONSENSUS (PoS Validators)             │   │   │
│  │  └─────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                        │                                    │
│                                        ▼                                    │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                      OFF-CHAIN SERVICES LAYER                        │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐              │   │
│  │  │  Provider    │  │ Benchmarking │  │  ML Scoring  │              │   │
│  │  │   Daemon     │  │   Daemon     │  │   Service    │              │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘              │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### On-Chain vs Off-Chain Components

| Capability | On-Chain Module | Off-Chain Component | Notes |
|-----------|-----------------|---------------------|-------|
| Identity Verification | `x/veid` | ML scoring service | Encrypted scopes on-chain; deterministic scoring |
| MFA Gating | `x/mfa` | Factor providers | Policies on-chain; verification off-chain |
| Encryption | `x/encryption` | Client crypto libs | Envelope format and key metadata |
| Marketplace | `x/market`, `x/escrow` | Waldur services | Orders + escrow; encrypted fields |
| Provider Ops | `x/provider` | Provider daemon | Bids, provisioning, usage submission |
| HPC | `x/hpc` | SLURM adapters | Job tracking and lifecycle |

### Provider Daemon Architecture

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        PROVIDER DAEMON                                   │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐                  │
│  │  Bid Engine  │  │  Key Manager │  │ Usage Meter  │                  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘                  │
│         │                 │                 │                           │
│         └─────────────────┴─────────────────┘                           │
│                           │                                              │
│         ┌─────────────────┴─────────────────┐                           │
│         │         Adapter Layer             │                           │
│         │  ┌──────┐ ┌──────┐ ┌──────┐      │                           │
│         │  │ K8s  │ │SLURM │ │VMware│ ... │                           │
│         │  └──────┘ └──────┘ └──────┘      │                           │
│         └───────────────────────────────────┘                           │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

**Infrastructure Adapters:**

| Adapter | Backend | Use Case |
|---------|---------|----------|
| Kubernetes | K8s namespace isolation | Container workloads |
| SLURM | HPC job scheduler | Compute-intensive jobs |
| OpenStack | Waldur VM provisioning | Virtual machines |
| AWS | EC2/VPC/EBS via Waldur | Cloud VMs |
| Azure | Azure VMs via Waldur | Cloud VMs |
| VMware | vSphere integration | Enterprise VMs |

**Workload State Machine:**
```
Pending ──▶ Deploying ──▶ Running ──▶ Stopping ──▶ Stopped ──▶ Terminated
    │           │           │
    │           │           └──▶ Paused
    │           │
    └───────────┴──▶ Failed
```

---

## Data Flow Patterns

### Identity Verification Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│              IDENTITY VERIFICATION DATA FLOW                             │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. User uploads ID doc + selfie via approved mobile app                │
│  2. App signs payload (approved client signature)                       │
│  3. User signs transaction (wallet signature)                           │
│  4. Data encrypted to validators using X25519-XSalsa20-Poly1305        │
│  5. Encrypted envelope submitted to blockchain                          │
│  6. During consensus, validators:                                       │
│     a. Decrypt the envelope using their private key                     │
│     b. Run TensorFlow ML scoring (deterministic)                        │
│     c. Produce identical scores (0-100)                                 │
│  7. Score stored on-chain with verification status                      │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Marketplace Transaction Flow

```
┌─────────────────────────────────────────────────────────────────────────┐
│              MARKETPLACE DATA FLOW                                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  TENANT                 BLOCKCHAIN              PROVIDER                 │
│    │                        │                      │                     │
│    │──MsgCreateOrder───────▶│                      │                     │
│    │   (with deposit)       │                      │                     │
│    │                        │                      │                     │
│    │                        │◀──Watch for orders───│                     │
│    │                        │                      │                     │
│    │                        │◀──MsgCreateBid───────│                     │
│    │                        │   (price, specs)     │                     │
│    │                        │                      │                     │
│    │──MsgAcceptBid─────────▶│                      │                     │
│    │                        │                      │                     │
│    │                        │──Lease Created──────▶│                     │
│    │                        │                      │                     │
│    │                        │                      │──Deploy resources   │
│    │                        │                      │                     │
│    │                        │◀──Usage Records──────│                     │
│    │                        │   (signed)           │                     │
│    │                        │                      │                     │
│    │                        │──Escrow Settlement──▶│                     │
│    │                        │                      │                     │
└─────────────────────────────────────────────────────────────────────────┘
```

### Query vs Transaction

| Aspect | Query | Transaction |
|--------|-------|-------------|
| **Purpose** | Read state | Modify state |
| **Cost** | Free (no gas) | Requires gas fee |
| **Consensus** | Single node | All validators |
| **Finality** | Immediate | After block commit |
| **Endpoint** | gRPC query / REST GET | Broadcast tx |

---

## Key Takeaways

### Architecture Summary

1. **VirtEngine is built on Cosmos SDK** - Leverages proven blockchain infrastructure while adding custom modules for identity, marketplace, and compute.

2. **Modular design** - Each capability is encapsulated in a module (`x/`) with well-defined interfaces (IKeeper pattern).

3. **On-chain + off-chain separation** - Critical business logic on-chain for verifiability; compute-intensive operations off-chain with signed attestations.

4. **Privacy by default** - Sensitive data encrypted before on-chain storage; only authorized parties can decrypt.

5. **Determinism is critical** - All on-chain operations must produce identical results across validators for consensus.

### Key Patterns to Remember

| Pattern | Purpose | Example |
|---------|---------|---------|
| Keeper Pattern | Module interface encapsulation | `IKeeper` interface + concrete `Keeper` |
| Envelope Encryption | Secure on-chain storage | X25519-XSalsa20-Poly1305 envelopes |
| Signed Attestations | Off-chain verification | Provider usage records |
| State Machine | Lifecycle management | Workload states, escrow states |
| Ante Handlers | Transaction validation | MFA gating, signature verification |

---

## Self-Assessment Questions

### Knowledge Check

1. **What distinguishes a Cosmos SDK blockchain from a general-purpose blockchain like Ethereum?**
   <details>
   <summary>Click for answer</summary>
   Cosmos SDK builds application-specific blockchains with native Go modules instead of smart contracts on a general-purpose VM. This provides better performance, sovereignty, and customization.
   </details>

2. **Why must all on-chain operations be deterministic?**
   <details>
   <summary>Click for answer</summary>
   All validators must produce identical state transitions during consensus. Non-deterministic operations would cause validators to disagree, halting the chain.
   </details>

3. **What are the three required signatures for a VEID identity upload?**
   <details>
   <summary>Click for answer</summary>
   1. Client signature (approved capture app), 2. User signature (wallet), 3. Salt binding (prevents replay attacks).
   </details>

4. **Explain the purpose of the Keeper pattern in VirtEngine modules.**
   <details>
   <summary>Click for answer</summary>
   The Keeper pattern encapsulates module state access behind an interface (IKeeper), enabling dependency injection, cross-module calls, and testability through mocking.
   </details>

5. **What is the role of ante handlers in transaction processing?**
   <details>
   <summary>Click for answer</summary>
   Ante handlers validate transactions before execution, including signature verification, fee deduction, MFA checks, and message validation.
   </details>

### Scenario Questions

6. **A provider wants to participate in the marketplace. Trace the flow from registration to earning rewards.**
   <details>
   <summary>Click for answer</summary>
   1. Provider registers via MsgRegisterProvider
   2. Provider daemon watches for matching orders
   3. Daemon submits MsgCreateBid with pricing
   4. Tenant accepts bid, creating a lease
   5. Daemon provisions resources via adapter (K8s/SLURM)
   6. Daemon submits signed usage records
   7. Escrow settles payment to provider
   </details>

7. **How does VirtEngine ensure sensitive identity data remains private on a public blockchain?**
   <details>
   <summary>Click for answer</summary>
   All sensitive data is encrypted using X25519-XSalsa20-Poly1305 before on-chain storage. Only validators with the corresponding private keys can decrypt during consensus. Plaintext never appears on the public ledger.
   </details>

---

## Practical Exercises

### Exercise 1: Repository Navigation (30 minutes)

**Objective:** Familiarize yourself with the codebase structure.

1. Clone the VirtEngine repository
2. Identify the following files:
   - [ ] Main application entry point (`cmd/virtengine/main.go`)
   - [ ] Market module keeper (`x/market/keeper/keeper.go`)
   - [ ] VEID types definition (`x/veid/types/`)
   - [ ] Genesis configuration (`app/genesis.go`)
   - [ ] Ante handlers (`app/ante.go`)

3. Answer these questions:
   - How many custom modules are in `x/`?
   - What is the StoreKey for the escrow module?
   - Which base Cosmos SDK modules does VirtEngine import?

### Exercise 2: Trace a Transaction (45 minutes)

**Objective:** Understand transaction flow through the system.

1. Find the `MsgCreateOrder` message definition in the market module
2. Trace its path through:
   - [ ] Message validation (`ValidateBasic`)
   - [ ] Ante handler processing
   - [ ] Keeper method execution
   - [ ] State storage

3. Document:
   - What validations occur before the keeper is called?
   - What state changes when an order is created?
   - How does escrow integration work?

### Exercise 3: Module Dependency Analysis (30 minutes)

**Objective:** Understand module interactions.

1. Examine the keeper constructor for the market module
2. Identify which other module keepers it depends on
3. Draw a dependency diagram showing:
   - Direct dependencies (keeper references)
   - Indirect dependencies (through events/queries)

### Exercise 4: Run Local Network (45 minutes)

**Objective:** Experience the system in operation.

```bash
# Start localnet
./scripts/localnet.sh start

# Query chain status
virtengine status

# Create a test identity verification
virtengine tx veid create-scope --from test-user

# Query marketplace orders
virtengine query market orders
```

Document:
- Block time and validator count
- Transaction confirmation time
- Query response format

---

## References

### Internal Documentation

| Document | Path | Description |
|----------|------|-------------|
| Architecture Reference | `_docs/architecture.md` | Complete system architecture |
| Security Guidelines | `_docs/security-guidelines.md` | Secure coding practices |
| Development Environment | `_docs/development-environment.md` | Setup guide |
| Testing Guide | `_docs/testing-guide.md` | Test conventions |
| VEID Flow Spec | `_docs/veid-flow-spec.md` | Identity verification details |

### External Resources

| Resource | URL | Description |
|----------|-----|-------------|
| Cosmos SDK Docs | https://docs.cosmos.network | Official SDK documentation |
| CometBFT Docs | https://docs.cometbft.com | Consensus layer documentation |
| IBC Protocol | https://ibc.cosmos.network | Cross-chain communication |
| Tendermint Core | https://github.com/tendermint/tendermint | Consensus engine |

### Next Modules

After completing this module, proceed to:

1. **Security Fundamentals** (`security-fundamentals.md`) - Cryptographic foundations and secure coding
2. **Provider Operations** (`../provider/`) - Provider daemon and infrastructure management
3. **Validator Operations** (`../validator/`) - Running a validator node

---

*Module Version: 1.0.0 | Last Updated: 2025-01-24 | Maintainer: VirtEngine Training Team*
