# VirtEngine System Architecture

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Status:** Authoritative Baseline  
**Task Reference:** VE-000

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Components and Boundaries](#system-components-and-boundaries)
3. [Module Interactions](#module-interactions)
4. [Data Flow Diagrams](#data-flow-diagrams)
5. [Role Model and Access Control](#role-model-and-access-control)
6. [Sensitive Transactions (MFA-Gated)](#sensitive-transactions-mfa-gated)
7. [Identity Lifecycle (VEID)](#identity-lifecycle-veid)
8. [SLA Requirements](#sla-requirements)
9. [Security Architecture Overview](#security-architecture-overview)
10. [Appendices](#appendices)

---

## Executive Summary

VirtEngine is a Cosmos SDK-based hybrid blockchain platform that combines:

1. **Decentralized Identity Verification (VEID)** - ML-powered identity scoring (0-100) with validator consensus
2. **On-chain MFA Module** - Multi-factor authentication gating for sensitive transactions
3. **Encrypted Data Subsystem** - Public-key encryption for all sensitive on-chain data
4. **Cloud Marketplace** - Waldur-backed marketplace with encrypted order/offering payloads
5. **Provider Daemon** - Automated bidding, provisioning (K8s/SLURM), and on-chain usage recording
6. **Supercomputer/HPC** - Distributed computing via SLURM clusters controlled through blockchain

The architecture prioritizes:
- **Privacy**: Sensitive data is never stored in plaintext on the public ledger
- **Decentralization**: No single point of failure for identity verification or order fulfillment
- **Verifiability**: Validators independently recompute identity scores during consensus
- **Compliance**: Role-based access control with audit trails

---

## System Components and Boundaries

### High-Level Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────────────────┐
│                              VIRTENGINE ECOSYSTEM                                    │
├─────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                      │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                         CLIENT LAYER                                         │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │   │
│  │  │  VE Portal   │  │ Mobile App   │  │   CLI/SDK    │  │ Waldur UI    │     │   │
│  │  │  (React)     │  │ (Approved)   │  │  (Go/TS)     │  │ (Django)     │     │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │   │
│  └─────────┼─────────────────┼─────────────────┼─────────────────┼─────────────┘   │
│            │                 │                 │                 │                  │
│            ▼                 ▼                 ▼                 ▼                  │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                         API GATEWAY LAYER                                    │   │
│  │  ┌──────────────────────────────────────────────────────────────────────┐   │   │
│  │  │  TLS-Encrypted REST/gRPC Endpoints (Cosmos LCD/gRPC + Waldur API)    │   │   │
│  │  └──────────────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                            │
│                                        ▼                                            │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                      BLOCKCHAIN LAYER (Cosmos SDK)                          │   │
│  │                                                                              │   │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌──────────┐  │   │
│  │  │  VEID   │ │   MFA   │ │ Encrypt │ │ Market  │ │ Escrow  │ │  Roles   │  │   │
│  │  │ Module  │ │ Module  │ │ Module  │ │ Module  │ │ Module  │ │  Module  │  │   │
│  │  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬─────┘  │   │
│  │       │           │           │           │           │           │         │   │
│  │  ┌────┴───────────┴───────────┴───────────┴───────────┴───────────┴────┐   │   │
│  │  │                    COSMOS SDK BASE MODULES                           │   │   │
│  │  │    (auth, bank, staking, gov, slashing, distribution, params)        │   │   │
│  │  └─────────────────────────────────────────────────────────────────────┘   │   │
│  │                                    │                                        │   │
│  │                                    ▼                                        │   │
│  │  ┌─────────────────────────────────────────────────────────────────────┐   │   │
│  │  │              TENDERMINT CONSENSUS (PoS Validators)                   │   │   │
│  │  └─────────────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                        │                                            │
│                                        ▼                                            │
│  ┌─────────────────────────────────────────────────────────────────────────────┐   │
│  │                      OFF-CHAIN SERVICES LAYER                               │   │
│  │                                                                              │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │   │
│  │  │  Provider    │  │   Waldur     │  │ Benchmarking │  │  ML Scoring  │    │   │
│  │  │   Daemon     │  │  Services    │  │   Daemon     │  │   Service    │    │   │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘    │   │
│  │         │                 │                 │                 │             │   │
│  │         ▼                 ▼                 ▼                 ▼             │   │
│  │  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │   │
│  │  │  Kubernetes  │  │   Django     │  │   Metrics    │  │  TensorFlow  │    │   │
│  │  │    SLURM     │  │   Postgres   │  │   Storage    │  │   Models     │    │   │
│  │  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                      │
└─────────────────────────────────────────────────────────────────────────────────────┘
```

### Component Descriptions

| Component | Description | Trust Boundary |
|-----------|-------------|----------------|
| **VE Portal** | React-based web interface for end-users and administrators | Client-side (untrusted input) |
| **Mobile App** | Approved capture client for identity document/selfie capture | Client-side with approved-client signing |
| **CLI/SDK** | Developer tools for programmatic access | Client-side (untrusted input) |
| **VE Blockchain** | Cosmos SDK chain with custom modules | Core trusted compute |
| **VEID Module** | Identity scope storage, ML scoring, verification status | On-chain (encrypted payloads) |
| **MFA Module** | Multi-factor authentication policies and gating | On-chain policy engine |
| **Encryption Module** | Public-key encryption primitives and envelope format | On-chain + off-chain decryption |
| **Market Module** | Orders, offerings, bids, leases | On-chain (encrypted sensitive fields) |
| **Escrow Module** | Payment holds and settlement | On-chain accounting |
| **Roles Module** | RBAC and account state management | On-chain access control |
| **Provider Daemon** | Off-chain service: bidding, provisioning, usage reporting | Provider-controlled (signed messages) |
| **Waldur Services** | Marketplace backend integration | External service (trusted integration) |
| **Benchmarking Daemon** | Provider performance metrics collection | Provider-controlled (signed metrics) |
| **ML Scoring Service** | TensorFlow inference for identity verification | Validator-controlled (deterministic) |

---

## Module Interactions

### Module Dependency Graph

```
                                    ┌─────────────┐
                                    │   Cosmos    │
                                    │  Base SDK   │
                                    │ (auth/bank) │
                                    └──────┬──────┘
                                           │
                    ┌──────────────────────┼──────────────────────┐
                    │                      │                      │
                    ▼                      ▼                      ▼
            ┌───────────────┐      ┌───────────────┐      ┌───────────────┐
            │     Roles     │◄────►│   Encryption  │◄────►│     Audit     │
            │    Module     │      │    Module     │      │    Module     │
            └───────┬───────┘      └───────┬───────┘      └───────────────┘
                    │                      │
        ┌───────────┼───────────┬──────────┴──────────┬───────────────────┐
        │           │           │                     │                   │
        ▼           ▼           ▼                     ▼                   ▼
┌───────────┐ ┌───────────┐ ┌───────────┐     ┌───────────┐       ┌───────────┐
│    MFA    │ │   VEID    │ │   Cert    │     │  Market   │       │  Provider │
│  Module   │ │  Module   │ │  Module   │     │  Module   │       │  Module   │
└─────┬─────┘ └─────┬─────┘ └───────────┘     └─────┬─────┘       └─────┬─────┘
      │             │                               │                   │
      │             └───────────────┐               │                   │
      │                             │               │                   │
      ▼                             ▼               ▼                   ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                        TRANSACTION GATING LAYER                             │
│  • MFA verification before sensitive tx execution                           │
│  • Identity score threshold checks for marketplace access                   │
│  • Role-based permission enforcement                                        │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                          ESCROW & SETTLEMENT                                │
│  • Payment holds on order creation                                          │
│  • Usage-based settlement from provider daemon reports                      │
│  • Staking rewards distribution                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Inter-Module Message Flow

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                      IDENTITY VERIFICATION FLOW                              │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐   │
│  │ Client  │───►│ Encrypt │───►│  VEID   │───►│Validator│───►│Consensus│   │
│  │ Capture │    │ Module  │    │ Module  │    │ ML Eval │    │  Vote   │   │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘   │
│       │              │              │              │              │         │
│       │ 1. Capture   │ 2. Encrypt   │ 3. Store     │ 4. Decrypt   │ 5. Vote │
│       │ doc+selfie   │ payload to   │ encrypted    │ + ML score   │ on score│
│       │ + sign       │ validator    │ scope refs   │ (0-100)      │         │
│       ▼              ▼              ▼              ▼              ▼         │
│  [salt+client     [envelope:     [state:        [score +       [block      │
│   sig+user sig]    pubkey,       scope_refs,    status in      includes    │
│                    cipher,       timestamps]    proposed       final       │
│                    nonce]                       block]         score]      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                       MARKETPLACE ORDER FLOW                                 │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐   │
│  │Customer │───►│ Market  │───►│ Escrow  │───►│Provider │───►│ Lease   │   │
│  │  Order  │    │ Module  │    │ Module  │    │ Daemon  │    │ Active  │   │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘    └─────────┘   │
│       │              │              │              │              │         │
│       │ 1. Create    │ 2. Store     │ 3. Hold     │ 4. Bid +     │ 5. Lease│
│       │ order with   │ order (enc   │ payment     │ provision    │ created │
│       │ encrypted    │ fields)      │ in escrow   │ workload     │ + usage │
│       │ details      │              │             │              │ start   │
│       ▼              ▼              ▼              ▼              ▼         │
│  [MFA required  [encrypted     [tokens        [signed        [on-chain   │
│   if high-value] order config]  locked]        bid tx]        lease ref] │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Data Flow Diagrams

### Encryption Envelope Format

```
┌─────────────────────────────────────────────────────────────────┐
│                    ENCRYPTED PAYLOAD ENVELOPE                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Header                                                     │ │
│  │  ├─ version: uint8 (envelope format version)               │ │
│  │  ├─ algorithm_id: uint8 (e.g., X25519-XChaCha20-Poly1305) │ │
│  │  ├─ recipient_pubkey: bytes[32]                            │ │
│  │  └─ sender_pubkey: bytes[32] (optional, for auth)         │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Cryptographic Material                                     │ │
│  │  ├─ nonce: bytes[24]                                        │ │
│  │  ├─ ciphertext: bytes[variable]                             │ │
│  │  └─ auth_tag: bytes[16] (included in ciphertext for AEAD)  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Metadata (plaintext, for routing/indexing)                 │ │
│  │  ├─ content_type: string (e.g., "identity_scope")          │ │
│  │  ├─ created_at: timestamp                                   │ │
│  │  └─ sender_signature: bytes[64] (signs header+ciphertext)  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Identity Verification Data Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    VEID: REGISTRATION → AUTHENTICATION → AUTHORIZATION       │
└─────────────────────────────────────────────────────────────────────────────┘

  ┌─────────────┐                                                              
  │  UNVERIFIED │  Account created, no identity scopes uploaded               
  │    (Tier 0) │  • Can browse marketplace                                   
  └──────┬──────┘  • Cannot place orders or access sensitive features         
         │                                                                     
         │ Upload identity scopes (doc + selfie + metadata)                   
         │ [Approved client signature + User signature required]              
         ▼                                                                     
  ┌─────────────┐                                                              
  │   PENDING   │  Identity scopes submitted, awaiting validator scoring      
  │  (Tier 0.5) │  • Scopes encrypted to validator pubkeys                    
  └──────┬──────┘  • Block proposer decrypts and runs ML scoring              
         │                                                                     
         │ Validator consensus on ML score (0-100)                            
         │ [All validators recompute and vote]                                
         ▼                                                                     
  ┌─────────────┐     Score < 50        ┌─────────────┐                       
  │  VERIFIED   │◄─────────────────────►│  REJECTED   │                       
  │  (Tier 1)   │     Score ≥ 50        │  (Tier 0)   │                       
  └──────┬──────┘                        └─────────────┘                       
         │  • Score 50-69: Basic marketplace access                            
         │  • Score 70-84: Standard offerings                                  
         │  • Score 85-100: Premium/high-value offerings                       
         │                                                                     
         │ Initiate sensitive transaction (MFA required)                      
         │ [Submit MFA factors per policy]                                    
         ▼                                                                     
  ┌─────────────┐                                                              
  │ AUTHORIZED  │  Temporary elevated state for sensitive action              
  │  (Tier 2)   │  • Key rotation, account recovery                           
  └─────────────┘  • High-value purchases, withdrawals                        
                   • Provider registration                                     
```

---

## Role Model and Access Control

### Role Hierarchy

| Role | Description | Trust Level | Key Permissions |
|------|-------------|-------------|-----------------|
| **GenesisAccount** | Initial chain authority | Highest | Nominate roles, governance proposals, emergency actions |
| **Administrator** | Platform operations | High | Manage account states, moderate content, config changes |
| **Moderator** | Content/user moderation | Medium-High | Review identity uploads, handle disputes, suspend users |
| **Staker/Validator** | Consensus participants | High | Block production, identity verification, governance voting |
| **ServiceProvider** | Infrastructure operators | Medium | List offerings, bid on orders, submit usage records |
| **Customer** | End users | Standard | Browse, place orders, upload identity scopes |
| **SupportAgent** | Customer support | Medium | Read support tickets, assist with account issues |

### Account State Machine

```
                    ┌───────────┐
                    │  CREATED  │
                    └─────┬─────┘
                          │ Identity upload
                          ▼
    ┌─────────────────────────────────────────┐
    │                 ACTIVE                   │
    │  ┌─────────────────────────────────────┐ │
    │  │  UNVERIFIED → PENDING → VERIFIED    │ │
    │  └─────────────────────────────────────┘ │
    └────────────┬────────────────┬────────────┘
                 │                │
     Admin/Gov   │                │ Violation
     action      ▼                ▼
          ┌───────────┐    ┌───────────┐
          │ SUSPENDED │    │  FLAGGED  │
          └─────┬─────┘    └─────┬─────┘
                │                │
                │ Appeal/        │ Review
                │ Resolution     │ Complete
                ▼                ▼
          ┌───────────┐    ┌───────────┐
          │ REINSTATED│    │TERMINATED │
          │ (→ACTIVE) │    │           │
          └───────────┘    └───────────┘
```

---

## Sensitive Transactions (MFA-Gated)

The following transactions require multi-factor authentication before execution:

### Tier 1: Always Require MFA

| Transaction Type | Risk Level | Required Factors | Rationale |
|-----------------|------------|------------------|-----------|
| `AccountRecovery` | Critical | VEID + FIDO2 + SMS/Email | Prevent unauthorized account takeover |
| `KeyRotation` | Critical | VEID + FIDO2 + Existing Key | Protect against key compromise |
| `ProviderRegistration` | High | VEID (score ≥70) + FIDO2 | Ensure provider legitimacy |
| `ValidatorRegistration` | Critical | VEID (score ≥85) + FIDO2 + Gov approval | Protect consensus integrity |
| `LargeWithdrawal` (>10,000 VE) | High | VEID + FIDO2 | Prevent theft |
| `GovernanceProposal` | High | VEID + FIDO2 | Prevent spam/malicious proposals |

### Tier 2: Conditional MFA (Based on Context)

| Transaction Type | Condition | Required Factors |
|-----------------|-----------|------------------|
| `OrderCreate` | Order value >1,000 VE | VEID + FIDO2 |
| `OrderCreate` | First order from account | VEID |
| `OfferingCreate` | Provider's first offering | VEID + FIDO2 |
| `TransferTokens` | To new address | FIDO2 |
| `TransferTokens` | Amount >5,000 VE | VEID + FIDO2 |
| `UpdateAccountSettings` | Change email/phone | SMS/Email verification |
| `SupportRequest` | Access to sensitive logs | VEID |

### Tier 3: Optional MFA (User-Configurable)

| Transaction Type | Default | User Can Require |
|-----------------|---------|------------------|
| `OrderCreate` (any) | No MFA | Yes |
| `TransferTokens` (any) | No MFA | Yes |
| `OfferingUpdate` | No MFA | Yes |
| `LeaseClose` | No MFA | Yes |

### MFA Factor Types

```
┌───────────────────────────────────────────────────────────────────┐
│                      MFA FACTOR REGISTRY                          │
├───────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │      VEID       │  │     FIDO2       │  │    SMS/Email    │  │
│  │  (Biometric)    │  │  (Hardware Key) │  │    (OTP)        │  │
│  ├─────────────────┤  ├─────────────────┤  ├─────────────────┤  │
│  │ Factor ID: 0x01 │  │ Factor ID: 0x02 │  │ Factor ID: 0x03 │  │
│  │ Strength: HIGH  │  │ Strength: HIGH  │  │ Strength: MEDIUM│  │
│  │ On-chain: Score │  │ On-chain: PubKey│  │ On-chain: Hash  │  │
│  │ Verify: ML+Face │  │ Verify: WebAuthn│  │ Verify: OTP     │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐                       │
│  │   TOTP App      │  │  Trusted Device │                       │
│  │ (Authenticator) │  │  (Remember Me)  │                       │
│  ├─────────────────┤  ├─────────────────┤                       │
│  │ Factor ID: 0x04 │  │ Factor ID: 0x05 │                       │
│  │ Strength: MEDIUM│  │ Strength: LOW   │                       │
│  │ On-chain: Seed* │  │ On-chain: DevID │                       │
│  │ Verify: TOTP    │  │ Verify: Cookie  │                       │
│  └─────────────────┘  └─────────────────┘                       │
│                                                                   │
│  * TOTP seed stored encrypted, decrypted only by user            │
│                                                                   │
└───────────────────────────────────────────────────────────────────┘
```

---

## Identity Lifecycle (VEID)

> **📘 Comprehensive Specification**: For complete VEID lifecycle details including state machine diagrams, capability matrices, and implementation guidelines, see the [VEID Flow Specification](./veid-flow-spec.md).

### Registration vs Authentication vs Authorization

| Phase | State | Actions Permitted | Identity Score | MFA Required |
|-------|-------|-------------------|----------------|--------------|
| **Registration** | Unverified | Browse marketplace, view public data | N/A | No |
| **Authentication** | Pending→Verified | Place orders (score-gated), basic marketplace | 0-100 | Per transaction |
| **Authorization** | Authorized (temp) | Sensitive actions (recovery, high-value) | ≥70 | Yes (always) |

### Score Thresholds and Capabilities

| Score Range | Tier | Marketplace Access | Max Order Value | Provider Registration |
|-------------|------|-------------------|-----------------|----------------------|
| 0-49 | Rejected | None | $0 | No |
| 50-69 | Basic | Basic offerings only | $500/order | No |
| 70-84 | Standard | Standard offerings | $10,000/order | Yes (with MFA) |
| 85-100 | Premium | All offerings | Unlimited | Yes (with MFA) |

---

## SLA Requirements

### Identity Scoring SLAs

| Metric | Target | Maximum | Failure Handling |
|--------|--------|---------|------------------|
| **Scoring Latency** | Within current block window (~6s) | 3 blocks (18s) | Set status=Pending, async finalization |
| **Async Finalization** | Within 30 blocks | 100 blocks | Auto-reject with TIMEOUT reason |
| **Score Determinism** | 100% match across validators | 0 tolerance | Block rejected if mismatch |
| **ML Model Sync** | All validators on same version | N/A | Validator offline until synced |

### Identity State Transitions

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    IDENTITY SCORING STATE MACHINE                            │
└─────────────────────────────────────────────────────────────────────────────┘

  Upload TX        Block N           Block N+1         Block N+K
  Submitted        (Proposer)        (Validators)      (K ≤ 3)
      │                │                  │                │
      ▼                ▼                  ▼                ▼
  ┌────────┐      ┌────────┐         ┌────────┐      ┌────────┐
  │SUBMITTED│─────►│SCORING │─────────►│PENDING │─────►│FINALIZED│
  └────────┘      └────────┘         └────────┘      └────────┘
      │                │                  │                │
      │                │                  │                ▼
      │                │                  │          ┌───────────┐
      │                │                  │          │ VERIFIED  │
      │                │                  │          │    or     │
      │                │                  │          │ REJECTED  │
      │                │                  │          └───────────┘
      │                │                  │
      ▼                ▼                  ▼
  [TX in mempool]  [Proposer runs   [Other validators  [Consensus reached,
                   ML inference,     recompute score,   score committed
                   includes score    vote on block      to state]
                   in block]         validity]
```

### Provisioning State Transitions

| State | Description | SLA Target | Timeout Action |
|-------|-------------|------------|----------------|
| `ORDER_OPEN` | Order submitted, awaiting bids | N/A | User can cancel |
| `ORDER_MATCHED` | Bid accepted, awaiting provision | 5 minutes | Auto-cancel, refund |
| `LEASE_ACTIVE` | Workload running | N/A (continuous) | Heartbeat required every 60s |
| `LEASE_CLOSING` | Termination requested | 2 minutes | Force terminate, final usage |
| `LEASE_CLOSED` | Workload terminated, settled | N/A | N/A |

### Provider Daemon SLAs

| Metric | Target | Maximum | Penalty |
|--------|--------|---------|---------|
| **Bid Response Time** | <2 seconds | 10 seconds | Miss opportunity |
| **Provision Time** | <5 minutes | 30 minutes | Order auto-cancels |
| **Usage Report Frequency** | Every 60 seconds | 5 minutes | Slashing warning |
| **Uptime** | 99.5% monthly | N/A | Reputation impact |

---

## Security Architecture Overview

### Defense in Depth

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           SECURITY LAYERS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Layer 1: TRANSPORT SECURITY                                                │
│  ├─ TLS 1.3 for all API communications                                      │
│  ├─ Certificate pinning for approved clients                                │
│  └─ Mutual TLS for provider daemon ↔ orchestrator                          │
│                                                                              │
│  Layer 2: AUTHENTICATION                                                    │
│  ├─ Cosmos account keypairs (secp256k1/ed25519)                            │
│  ├─ Transaction signatures verified on all messages                         │
│  └─ Approved-client signatures for identity uploads                         │
│                                                                              │
│  Layer 3: AUTHORIZATION                                                     │
│  ├─ Role-based access control (RBAC) on-chain                              │
│  ├─ Identity score thresholds for marketplace access                       │
│  └─ MFA gating for sensitive transactions                                  │
│                                                                              │
│  Layer 4: DATA PROTECTION                                                   │
│  ├─ Public-key encryption for all sensitive payloads                       │
│  ├─ Encrypted at rest (validator nodes, off-chain stores)                  │
│  └─ No plaintext sensitive data on public ledger                           │
│                                                                              │
│  Layer 5: CONSENSUS SECURITY                                                │
│  ├─ PoS with slashing for misbehavior                                      │
│  ├─ Validator recomputation of identity scores                             │
│  └─ Deterministic ML inference for consensus                                │
│                                                                              │
│  Layer 6: AUDIT & MONITORING                                                │
│  ├─ On-chain event logs for all state changes                              │
│  ├─ Off-chain metrics and alerting                                         │
│  └─ Incident response playbooks                                            │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Key Security Properties

| Property | Implementation | Verification |
|----------|---------------|--------------|
| **Confidentiality** | Public-key encryption (X25519-XChaCha20-Poly1305) | Only recipient can decrypt |
| **Integrity** | Digital signatures on all messages | Invalid signatures rejected |
| **Availability** | Decentralized validators, provider redundancy | No SPOF |
| **Non-repudiation** | Signed transactions, audit logs | On-chain evidence |
| **Authenticity** | Approved-client signatures, MFA | Multi-factor proof |

---

## Appendices

### A. Cryptographic Algorithms

| Purpose | Algorithm | Key Size | Notes |
|---------|-----------|----------|-------|
| Asymmetric Encryption | X25519 + XSalsa20-Poly1305 | 256-bit | AEAD, NaCl box compatible |
| Digital Signatures | Ed25519 or secp256k1 | 256-bit | Cosmos SDK standard |
| Hashing | SHA-256, BLAKE2b | 256-bit | Deterministic |
| Key Derivation | HKDF-SHA256 | Variable | For derived keys |
| Random Generation | CSPRNG | N/A | Cryptographically secure |

### A.1 Encryption Module Implementation (VE-101)

The encryption module (`x/encryption`) provides on-chain public-key encryption primitives and the canonical encrypted payload envelope format for all sensitive data.

#### Supported Algorithms

| Algorithm ID | Description | Status |
|--------------|-------------|--------|
| `X25519-XSALSA20-POLY1305` | X25519 key exchange + XSalsa20-Poly1305 AEAD (NaCl box) | Primary |
| `AGE-X25519` | age encryption format with X25519 | Reserved |

#### EncryptedPayloadEnvelope Structure

```go
type EncryptedPayloadEnvelope struct {
    Version          uint32            // Envelope format version (current: 1)
    AlgorithmID      string            // e.g., "X25519-XSALSA20-POLY1305"
    RecipientKeyIDs  []string          // Key fingerprints of intended recipients
    EncryptedKeys    [][]byte          // DEK encrypted for each recipient (multi-recipient mode)
    Nonce            []byte            // 24-byte IV/nonce for XSalsa20
    Ciphertext       []byte            // Encrypted data
    SenderSignature  []byte            // Signature over hash(version || algo || ciphertext || nonce || recipients)
    SenderPubKey     []byte            // 32-byte X25519 public key for verification/key exchange
    Metadata         map[string]string // Optional public metadata
}
```

#### Encryption Flow

```
┌──────────────────────────────────────────────────────────────────────────────┐
│                       ENCRYPTION SUBSYSTEM FLOW                              │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. KEY REGISTRATION (on-chain)                                             │
│     User → MsgRegisterRecipientKey → Keeper stores X25519 pubkey           │
│     • Key fingerprint: SHA256(pubkey)[:20] (hex-encoded)                   │
│     • Lookup by address or fingerprint                                      │
│                                                                              │
│  2. ENVELOPE CREATION (off-chain, using x/encryption/crypto)                │
│     Client → CreateEnvelope(plaintext, recipientPubKey, senderKeyPair)     │
│     • Generate random 24-byte nonce (crypto/rand)                          │
│     • Encrypt with NaCl box (X25519 + XSalsa20-Poly1305)                   │
│     • Sign envelope for authenticity                                        │
│                                                                              │
│  3. ENVELOPE STORAGE (on-chain)                                             │
│     Envelope stored as field in order/identity/support message             │
│     • Validators see only ciphertext                                       │
│     • Metadata may contain routing hints                                    │
│                                                                              │
│  4. ENVELOPE DECRYPTION (off-chain)                                         │
│     Recipient → OpenEnvelope(envelope, recipientPrivateKey)                 │
│     • Verify envelope structure                                            │
│     • Decrypt with NaCl box.Open                                           │
│     • Optionally verify sender signature                                   │
│                                                                              │
│  5. MULTI-RECIPIENT MODE                                                    │
│     For envelopes with multiple recipients:                                │
│     • Generate random DEK (data encryption key)                            │
│     • Encrypt data with DEK                                                │
│     • Encrypt DEK separately to each recipient                             │
│     • Each recipient can decrypt with their private key                    │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

#### Security Properties

- **Nonce uniqueness**: Generated with crypto/rand for every encryption
- **Forward secrecy**: Ephemeral sender keys can be used per message
- **Authentication**: Sender signature prevents tampering
- **Multi-recipient**: Single ciphertext, per-recipient encrypted keys
- **No private keys on-chain**: Only public keys stored in state

### B. Module Protobuf Definitions (Summary)

```protobuf
// Encryption envelope
message EncryptedEnvelope {
  uint32 version = 1;
  uint32 algorithm_id = 2;
  bytes recipient_pubkey = 3;
  bytes sender_pubkey = 4;
  bytes nonce = 5;
  bytes ciphertext = 6;
  bytes sender_signature = 7;
  string content_type = 8;
  google.protobuf.Timestamp created_at = 9;
}

// Identity scope reference
message IdentityScopeRef {
  string scope_id = 1;
  string scope_type = 2; // document, selfie, video, email, phone, domain
  EncryptedEnvelope encrypted_payload = 3;
  bytes client_signature = 4;
  bytes user_signature = 5;
  bytes salt = 6;
  google.protobuf.Timestamp uploaded_at = 7;
}

// Identity verification status
message IdentityStatus {
  string account = 1;
  uint32 score = 2; // 0-100
  enum Status {
    UNKNOWN = 0;
    PENDING = 1;
    VERIFIED = 2;
    REJECTED = 3;
  }
  Status status = 3;
  string model_version = 4;
  google.protobuf.Timestamp verified_at = 5;
  string reason_code = 6;
}

// MFA policy
message MFAPolicy {
  string account = 1;
  repeated string required_factors = 2; // factor IDs
  map<string, MFAFactorConfig> enrolled_factors = 3;
  bool trusted_device_reduction = 4;
  uint32 session_timeout_seconds = 5;
}
```

### C. References

- [Cosmos SDK Documentation](https://docs.cosmos.network/)
- [Tendermint Consensus](https://docs.tendermint.com/)
- [VirtEngine PRD](./ralph/prd.json)
- [Threat Model](./threat-model.md)
- [Data Classification](./data-classification.md)
- [VEID Flow Specification](./veid-flow-spec.md) - Registration, Authentication, and Authorization lifecycle

---

*Document maintained by VirtEngine Architecture Team*  
*Last updated: 2026-01-24*
