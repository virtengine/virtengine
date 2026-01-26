# VirtEngine Data Classification

**Version:** 1.0.0  
**Date:** 2026-01-24  
**Status:** Authoritative Baseline  
**Task Reference:** VE-000

---

## Table of Contents

1. [Overview](#overview)
2. [Classification Levels](#classification-levels)
3. [Data Inventory](#data-inventory)
4. [Handling Rules by Classification](#handling-rules-by-classification)
5. [Encryption Requirements](#encryption-requirements)
6. [Access Control Matrix](#access-control-matrix)
7. [Data Lifecycle Management](#data-lifecycle-management)
8. [Compliance Mapping](#compliance-mapping)

---

## Overview

This document defines the data classification scheme for VirtEngine, establishing:

- **Sensitivity levels** for all data types
- **Handling rules** for storage, transit, and access
- **Encryption requirements** at each layer
- **Retention and disposal** policies

### Guiding Principles

1. **Sensitive data is never stored in plaintext on the public ledger**
2. **Encryption is applied at rest AND in transit**
3. **Access follows least-privilege principle**
4. **Audit trails are maintained for all sensitive data access**

---

## Classification Levels

| Level | Name | Description | Examples |
|-------|------|-------------|----------|
| **C0** | Public | Freely available, no restrictions | Block height, tx hashes, offering listings (non-sensitive fields) |
| **C1** | Internal | Not secret but not published | Node configurations, internal metrics, debug logs |
| **C2** | Confidential | Business-sensitive, access controlled | Account balances, order history, usage records |
| **C3** | Restricted | Highly sensitive, encrypted always | Identity documents, order credentials, private keys |
| **C4** | Critical | Highest sensitivity, maximum protection | Validator keys, ML training data, identity biometrics |

### Classification Decision Tree

```
                          ┌─────────────────────────────┐
                          │ Does disclosure cause harm? │
                          └──────────────┬──────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    │ NO                 │ YES                │
                    ▼                    │                    ▼
               ┌────────┐               │               ┌─────────┐
               │ PUBLIC │               │               │ Is it   │
               │  (C0)  │               │               │ PII/PHI?│
               └────────┘               │               └────┬────┘
                                        │                    │
                            ┌───────────┴───────────┐       │
                            │                       │       │
                            ▼                       ▼       ▼
                    ┌──────────────┐         ┌──────────────┐
                    │ Internal ops │         │ YES: Check   │
                    │ only?        │         │ sensitivity  │
                    └──────┬───────┘         └──────┬───────┘
                           │                        │
              ┌────────────┴────────────┐          │
              │ YES                NO   │          │
              ▼                    ▼    │          ▼
         ┌──────────┐       ┌──────────┐│    ┌───────────────┐
         │ INTERNAL │       │CONFIDENTIAL│    │ Identity doc? │
         │   (C1)   │       │   (C2)   ││    │ Biometric?    │
         └──────────┘       └──────────┘│    │ Private key?  │
                                        │    └───────┬───────┘
                                        │            │
                                        │    ┌───────┴───────┐
                                        │    │ YES       NO  │
                                        │    ▼           ▼   │
                                        │ ┌────────┐ ┌────────┐
                                        │ │CRITICAL│ │RESTRICTED│
                                        │ │  (C4)  │ │  (C3)  │
                                        │ └────────┘ └────────┘
                                        │
                                        └─────────────────────┘
```

---

## Data Inventory

### Identity Data (VEID Module)

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| Identity document images | **C4** (Critical) | On-chain (envelope) | Yes (to validators) | Until revoked + 7 years |
| Selfie/video captures | **C4** (Critical) | On-chain (envelope) | Yes (to validators) | Until revoked + 7 years |
| Face embeddings (vectors) | **C4** (Critical) | Validator memory only | Yes (ephemeral) | Not persisted |
| Document OCR text | **C4** (Critical) | Validator memory only | Yes (ephemeral) | Not persisted |
| Identity score (0-100) | **C2** (Confidential) | On-chain (state) | No (public result) | Permanent |
| Verification status | **C2** (Confidential) | On-chain (state) | No (public result) | Permanent |
| Verification timestamp | **C0** (Public) | On-chain (state) | No | Permanent |
| Model version used | **C0** (Public) | On-chain (state) | No | Permanent |
| Upload salt | **C1** (Internal) | On-chain (metadata) | No | Permanent |
| Client signature | **C1** (Internal) | On-chain (metadata) | No | Permanent |
| User signature | **C1** (Internal) | On-chain (metadata) | No | Permanent |

### Account Data

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| Account address | **C0** (Public) | On-chain | No | Permanent |
| Account public key | **C0** (Public) | On-chain | No | Permanent |
| Account private key | **C4** (Critical) | User custody only | Yes (user-managed) | User-controlled |
| Token balances | **C2** (Confidential) | On-chain (state) | No | Permanent |
| Transaction history | **C2** (Confidential) | On-chain | No | Permanent |
| Role assignments | **C1** (Internal) | On-chain (state) | No | Permanent |
| Account state (active/suspended) | **C1** (Internal) | On-chain (state) | No | Permanent |
| Email address (if provided) | **C3** (Restricted) | Off-chain (Waldur) | Yes | Until account deletion |
| Phone number (if provided) | **C3** (Restricted) | Off-chain (Waldur) | Yes | Until account deletion |

### MFA Data

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| FIDO2 credential ID | **C2** (Confidential) | On-chain (state) | No | Until revoked |
| FIDO2 public key | **C2** (Confidential) | On-chain (state) | No | Until revoked |
| TOTP secret seed | **C3** (Restricted) | On-chain (envelope) | Yes (to user) | Until revoked |
| SMS/Email verification hash | **C2** (Confidential) | On-chain (state) | No (hashed) | Until used |
| Trusted device IDs | **C2** (Confidential) | On-chain (state) | No | Until revoked |
| MFA policy config | **C1** (Internal) | On-chain (state) | No | Permanent |
| MFA audit log | **C2** (Confidential) | On-chain (events) | No | Permanent |

### Marketplace Data

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| Offering ID | **C0** (Public) | On-chain | No | Permanent |
| Offering description | **C0** (Public) | On-chain | No | Permanent |
| Offering pricing | **C0** (Public) | On-chain | No | Permanent |
| Provider address | **C0** (Public) | On-chain | No | Permanent |
| Order ID | **C0** (Public) | On-chain | No | Permanent |
| Order customer address | **C0** (Public) | On-chain | No | Permanent |
| Order offering reference | **C0** (Public) | On-chain | No | Permanent |
| Order configuration | **C3** (Restricted) | On-chain (envelope) | Yes (to provider) | Lease duration + 90 days |
| Order credentials (SSH keys, passwords) | **C4** (Critical) | On-chain (envelope) | Yes (to provider) | Lease duration + 30 days |
| Order IP addresses/hostnames | **C3** (Restricted) | On-chain (envelope) | Yes (to customer/provider) | Lease duration + 30 days |
| Lease ID | **C0** (Public) | On-chain | No | Permanent |
| Lease status | **C0** (Public) | On-chain (state) | No | Permanent |
| Usage records | **C2** (Confidential) | On-chain | No | Permanent |
| Bid details | **C1** (Internal) | On-chain | No | Permanent |

### Provider Data

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| Provider address | **C0** (Public) | On-chain | No | Permanent |
| Provider offerings | **C0** (Public) | On-chain | No | Permanent |
| Provider stake amount | **C0** (Public) | On-chain | No | Permanent |
| Provider reputation score | **C0** (Public) | On-chain (state) | No | Permanent |
| Provider daemon signing key | **C2** (Confidential) | Provider custody | Yes | Provider-controlled |
| Benchmarking metrics | **C0** (Public) | On-chain | No | Permanent |
| Infrastructure details (internal) | **C3** (Restricted) | Off-chain (provider) | Yes | Provider-controlled |
| Customer workload data | **C4** (Critical) | Provider infrastructure | Customer-controlled | Customer-controlled |

### Support & Audit Data

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| Support ticket ID | **C1** (Internal) | Off-chain (Waldur) | No | 3 years |
| Support ticket content | **C3** (Restricted) | Off-chain (Waldur) | Yes | 3 years |
| Support attachments | **C3** (Restricted) | Off-chain (Waldur) | Yes | 3 years |
| Audit events (on-chain) | **C2** (Confidential) | On-chain (events) | No | Permanent |
| Audit logs (off-chain) | **C2** (Confidential) | Off-chain (SIEM) | Yes (at rest) | 7 years |
| Admin action logs | **C2** (Confidential) | On-chain (events) | No | Permanent |

### Validator/Staking Data

| Data Element | Classification | Storage Location | Encrypted | Retention |
|--------------|---------------|------------------|-----------|-----------|
| Validator address | **C0** (Public) | On-chain | No | Permanent |
| Validator public key | **C0** (Public) | On-chain | No | Permanent |
| Validator private key | **C4** (Critical) | HSM/secure custody | Yes (HSM) | Validator-controlled |
| Identity decryption key | **C4** (Critical) | HSM/secure custody | Yes (HSM) | Rotated periodically |
| Delegations | **C0** (Public) | On-chain | No | Permanent |
| Slashing history | **C0** (Public) | On-chain | No | Permanent |
| Commission settings | **C0** (Public) | On-chain | No | Permanent |

---

## Handling Rules by Classification

### C0: Public Data

```
┌─────────────────────────────────────────────────────────────────┐
│                    C0: PUBLIC HANDLING RULES                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Storage:        Any location, no encryption required           │
│  Transit:        TLS recommended but not required               │
│  Access:         Unrestricted, public APIs                      │
│  Logging:        Standard logging, no special handling          │
│  Retention:      Per business needs                             │
│  Disposal:       No special requirements                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### C1: Internal Data

```
┌─────────────────────────────────────────────────────────────────┐
│                   C1: INTERNAL HANDLING RULES                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Storage:        Internal systems only, access controlled       │
│  Transit:        TLS required                                   │
│  Access:         Authenticated users, role-based                │
│  Logging:        Access logging required                        │
│  Retention:      As defined per data type                       │
│  Disposal:       Secure deletion when expired                   │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### C2: Confidential Data

```
┌─────────────────────────────────────────────────────────────────┐
│                 C2: CONFIDENTIAL HANDLING RULES                  │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Storage:        Encrypted at rest (AES-256)                    │
│  Transit:        TLS 1.3 required                               │
│  Access:         Authenticated + authorized, RBAC enforced      │
│  Logging:        All access logged with user ID, timestamp      │
│  Retention:      As defined, encrypted backups                  │
│  Disposal:       Cryptographic erasure or secure wipe           │
│                                                                  │
│  Additional:                                                    │
│  • No sharing with third parties without consent                │
│  • Anonymization required for analytics                         │
│  • Export requires approval                                     │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### C3: Restricted Data

```
┌─────────────────────────────────────────────────────────────────┐
│                  C3: RESTRICTED HANDLING RULES                   │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Storage:        Public-key encrypted to recipient(s)           │
│                  On-chain: encrypted envelope format            │
│                  Off-chain: AES-256 + access controls           │
│  Transit:        TLS 1.3 required, end-to-end encryption        │
│  Access:         Recipient-only decryption                      │
│                  MFA required for access                        │
│  Logging:        All access/decrypt attempts logged             │
│                  Alerts on anomalous access patterns            │
│  Retention:      Minimum necessary, auto-expiry                 │
│  Disposal:       Key deletion renders data unrecoverable        │
│                                                                  │
│  Additional:                                                    │
│  • Never displayed in logs or error messages                    │
│  • Never cached in plaintext                                    │
│  • Masking required in UI (show partial only)                   │
│  • Breach notification required if exposed                      │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### C4: Critical Data

```
┌─────────────────────────────────────────────────────────────────┐
│                   C4: CRITICAL HANDLING RULES                    │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Storage:        Hardware security modules (HSM) preferred      │
│                  If software: encrypted + access-gated          │
│                  On-chain: always encrypted envelope            │
│                  Ephemeral only where possible (RAM-only)       │
│  Transit:        TLS 1.3 + end-to-end encryption                │
│                  Additional: threshold encryption for keys      │
│  Access:         Multi-party authorization for keys             │
│                  MFA + VEID required                            │
│                  Time-limited access windows                    │
│  Logging:        Real-time monitoring + alerting                │
│                  Immutable audit trail                          │
│  Retention:      Absolute minimum, destroy when not needed      │
│  Disposal:       Cryptographic erasure + physical destruction   │
│                  for hardware keys                              │
│                                                                  │
│  Additional:                                                    │
│  • Never in logs, metrics, or error outputs                     │
│  • Never persisted outside secure boundary                      │
│  • Air-gapped generation for critical keys                      │
│  • Split knowledge / dual control for master keys               │
│  • Incident response plan for exposure                          │
│  • Regular key rotation where applicable                        │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

---

## Encryption Requirements

### Encryption Standards

| Layer | Algorithm | Key Size | Notes |
|-------|-----------|----------|-------|
| **Transport (TLS)** | TLS 1.3 | N/A | ECDHE key exchange, AES-256-GCM |
| **At-Rest (Disk)** | AES-256-GCM | 256-bit | For off-chain stores |
| **On-Chain Payload** | X25519-XChaCha20-Poly1305 | 256-bit | Public-key encryption |
| **Key Derivation** | HKDF-SHA256 | Variable | For derived session keys |
| **Digital Signatures** | Ed25519 / secp256k1 | 256-bit | Cosmos SDK standard |
| **Hashing** | SHA-256 / BLAKE2b | 256-bit | Deterministic |

### On-Chain Encrypted Payload Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    ON-CHAIN ENCRYPTION FLOW                                  │
└─────────────────────────────────────────────────────────────────────────────┘

  ┌─────────────┐                                              ┌─────────────┐
  │   SENDER    │                                              │  RECIPIENT  │
  │             │                                              │             │
  │ plaintext   │                                              │ plaintext   │
  │    data     │                                              │    data     │
  └──────┬──────┘                                              └──────▲──────┘
         │                                                            │
         │ 1. Get recipient                                          │ 6. Decrypt
         │    public key                                             │    with private
         ▼                                                           │    key
  ┌─────────────┐                                              ┌─────────────┐
  │ Generate    │                                              │  Retrieve   │
  │ ephemeral   │                                              │  from chain │
  │ keypair     │                                              │  state      │
  └──────┬──────┘                                              └──────▲──────┘
         │                                                            │
         │ 2. ECDH key                                               │
         │    agreement                                              │
         ▼                                                           │
  ┌─────────────┐                                              ┌─────────────┐
  │ Derive      │      3. Encrypt with                         │ On-Chain    │
  │ shared      │─────── derived key ─────►                    │ State       │
  │ secret      │                          ┌─────────────┐     │             │
  └─────────────┘                          │ Encrypted   │     │ ┌─────────┐ │
                                           │ Envelope:   │────►│ │envelope │ │
                                           │ • nonce     │     │ └─────────┘ │
                                           │ • cipher    │     │             │
                                           │ • sender pk │     └─────────────┘
                                           │ • recip pk  │
                                           │ • signature │
                                           └─────────────┘

  4. Submit TX with envelope
  5. Chain stores encrypted envelope
  6. Recipient retrieves and decrypts off-chain
```

### Key Management Requirements

| Key Type | Generation | Storage | Rotation | Backup |
|----------|------------|---------|----------|--------|
| User account keys | User device | User custody (wallet) | User-initiated | Mnemonic phrase |
| Validator consensus keys | Air-gapped, HSM | HSM | Governance-defined | Secure backup |
| Validator identity keys | Air-gapped, HSM | HSM | Annual or on compromise | Secure backup |
| Approved client keys | Secure build env | TEE/Secure Enclave | Version releases | Key escrow (multi-party) |
| Service encryption keys | HSM | HSM | Quarterly | HSM backup |
| TOTP seeds | User device | On-chain (encrypted) | User-initiated | User backup |

---

## Access Control Matrix

### Role-Based Access by Data Classification

| Role | C0 (Public) | C1 (Internal) | C2 (Confidential) | C3 (Restricted) | C4 (Critical) |
|------|-------------|---------------|-------------------|-----------------|---------------|
| **Anonymous** | Read | - | - | - | - |
| **Customer** | Read | - | Own data only | Own data only | Own keys only |
| **ServiceProvider** | Read | Own config | Own + assigned orders | Assigned only | Own keys only |
| **SupportAgent** | Read | Read | Read (with audit) | Read (encrypted)* | - |
| **Moderator** | Read | Read | Read | Review flags* | - |
| **Administrator** | Read | Read/Write | Read/Write | Manage (encrypted) | Emergency only |
| **Validator** | Read | Read | Read | Decrypt (identity)** | Own keys only |
| **GenesisAccount** | Read | Read/Write | Read/Write | Manage | Emergency only |

\* Cannot decrypt, only view metadata or flagged content  
\** Threshold decryption during verification only

### MFA Requirements by Access Type

| Access Type | MFA Required | Factors |
|-------------|--------------|---------|
| View own C2 data | No | - |
| Export own C2 data | Yes | FIDO2 |
| View C3 data (own) | Yes | VEID or FIDO2 |
| Modify C3 data | Yes | VEID + FIDO2 |
| Access C4 data (own keys) | Yes | VEID + FIDO2 |
| Admin access to C3 | Yes | VEID + FIDO2 + SMS |
| Validator identity decryption | Automatic | Threshold (t-of-n validators) |

---

## Data Lifecycle Management

### Lifecycle Stages

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       DATA LIFECYCLE STAGES                                  │
└─────────────────────────────────────────────────────────────────────────────┘

  ┌──────────┐     ┌──────────┐     ┌──────────┐     ┌──────────┐
  │ CREATION │────►│  STORAGE │────►│   USE    │────►│ ARCHIVE  │
  └────┬─────┘     └────┬─────┘     └────┬─────┘     └────┬─────┘
       │                │                │                │
       ▼                ▼                ▼                ▼
  • Classify        • Encrypt       • Access        • Move to
  • Encrypt         • Access        • Control       • Cold storage
  • Sign            • Control       • Audit log     • Encrypt
  • Validate        • Backup        • Monitor       • Retain
                                                         │
                                                         ▼
                                                   ┌──────────┐
                                                   │ DISPOSAL │
                                                   └────┬─────┘
                                                        │
                                                        ▼
                                                   • Crypto-shred
                                                   • Verify deletion
                                                   • Audit log
```

### Retention Periods by Data Type

| Data Category | Default Retention | Legal Hold | Post-Deletion |
|---------------|-------------------|------------|---------------|
| Identity documents (encrypted) | Account lifetime + 7 years | Suspend deletion | Crypto-shred |
| Account data | Account lifetime | Suspend deletion | Anonymize or delete |
| Transaction history | Permanent (blockchain) | N/A (immutable) | N/A |
| Order configuration | Lease + 90 days | Extend | Crypto-shred |
| Credentials (SSH, passwords) | Lease + 30 days | N/A | Crypto-shred |
| Support tickets | 3 years | Extend | Anonymize |
| Audit logs (off-chain) | 7 years | Extend | Secure delete |
| Usage records | Permanent (blockchain) | N/A (immutable) | N/A |
| Benchmarking data | Permanent (blockchain) | N/A (immutable) | N/A |

### Data Deletion Process

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                      DATA DELETION WORKFLOW                                  │
└─────────────────────────────────────────────────────────────────────────────┘

  User/Admin Request               System Process              Verification
        │                               │                          │
        ▼                               ▼                          ▼
  ┌───────────┐                   ┌───────────┐              ┌───────────┐
  │ Deletion  │                   │ Check     │              │ Audit log │
  │ Request   │──────────────────►│ retention │              │ entry     │
  │ (MFA req) │                   │ + holds   │              │ created   │
  └───────────┘                   └─────┬─────┘              └───────────┘
                                        │                          ▲
                                        ▼                          │
                                  ┌───────────┐                    │
                           ┌──────│ Eligible? │──────┐             │
                           │ NO   └───────────┘ YES  │             │
                           ▼                         ▼             │
                     ┌───────────┐            ┌───────────┐        │
                     │ Reject +  │            │ Delete    │        │
                     │ notify    │            │ encryption│        │
                     └───────────┘            │ keys      │────────┤
                                              └─────┬─────┘        │
                                                    │              │
                                                    ▼              │
                                              ┌───────────┐        │
                                              │ Mark data │        │
                                              │ as deleted│        │
                                              │ on-chain  │────────┘
                                              └───────────┘
```

---

## Compliance Mapping

### Regulatory Alignment

| Regulation | Relevant Data Types | VirtEngine Controls |
|------------|---------------------|---------------------|
| **GDPR** | Identity data, PII | Encryption, consent, data minimization, right to erasure (crypto-shred) |
| **CCPA** | California user data | Access requests, deletion rights, opt-out mechanisms |
| **SOC 2** | All operational data | Access controls, audit logging, encryption, incident response |
| **PCI-DSS** | Payment data (if applicable) | Encryption, access controls, key management |
| **HIPAA** | Health data (if applicable) | BAA required, encryption, access controls |

### GDPR Specific Controls

| GDPR Article | Requirement | VirtEngine Implementation |
|--------------|-------------|---------------------------|
| Art. 17 | Right to erasure | Crypto-shred encrypted payloads; on-chain data anonymized via key deletion |
| Art. 20 | Data portability | Export APIs for own data (C2/C3) |
| Art. 25 | Privacy by design | Encryption by default, data minimization |
| Art. 32 | Security measures | Encryption, access control, audit logs |
| Art. 33 | Breach notification | Incident response plan, 72-hour notification |

### Data Processing Roles

| Entity | GDPR Role | Responsibilities |
|--------|-----------|------------------|
| VirtEngine (chain) | Processor | On-chain data storage, validation |
| Validators | Sub-processor | Identity verification, temporary decryption |
| Providers | Processor | Customer workload processing |
| Waldur | Processor | Marketplace backend |
| Customers | Controller | Define data processing purposes |

---

## Appendix: Data Classification Labels

### Label Format

```
VE-[CLASSIFICATION]-[CATEGORY]-[VERSION]

Examples:
VE-C4-IDENTITY-v1     (Critical identity document)
VE-C3-ORDER-v1        (Restricted order configuration)
VE-C2-USAGE-v1        (Confidential usage record)
VE-C1-CONFIG-v1       (Internal configuration)
VE-C0-PUBLIC-v1       (Public offering data)
```

### Classification Header (for encrypted envelopes)

```json
{
  "classification": "C3",
  "category": "ORDER",
  "version": "v1",
  "retention_days": 120,
  "auto_expire": true,
  "legal_hold": false,
  "data_controller": "ve1abc...xyz",
  "created_at": "2026-01-24T12:00:00Z"
}
```

---

*Document maintained by VirtEngine Security Team*  
*Last updated: 2026-01-24*
