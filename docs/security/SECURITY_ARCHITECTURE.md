# VirtEngine Security Architecture

**Version:** 1.0.0  
**Date:** 2026-01-30  
**Status:** Authoritative Baseline  
**Task Reference:** DOCS-003  
**Classification:** Public

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Security Design Principles](#security-design-principles)
3. [Architecture Overview](#architecture-overview)
4. [Layer Security Model](#layer-security-model)
5. [Identity and Access Management](#identity-and-access-management)
6. [Cryptographic Architecture](#cryptographic-architecture)
7. [Data Protection](#data-protection)
8. [Network Security](#network-security)
9. [Secure Development Lifecycle](#secure-development-lifecycle)
10. [Security Monitoring and Response](#security-monitoring-and-response)
11. [Compliance Integration](#compliance-integration)
12. [Related Documentation](#related-documentation)

---

## Executive Summary

VirtEngine implements a defense-in-depth security architecture for its decentralized cloud computing platform. This document provides a comprehensive overview of security controls, cryptographic systems, and protection mechanisms across all platform layers.

### Key Security Properties

| Property | Implementation | Status |
|----------|----------------|--------|
| **Confidentiality** | End-to-end encryption, envelope encryption | ✅ Implemented |
| **Integrity** | Digital signatures, hash commitments, consensus | ✅ Implemented |
| **Availability** | Distributed consensus, provider redundancy | ✅ Implemented |
| **Authentication** | Multi-factor (FIDO2, TOTP), VEID verification | ✅ Implemented |
| **Authorization** | Role-based access control (RBAC) | ✅ Implemented |
| **Non-repudiation** | On-chain signatures, audit trails | ✅ Implemented |
| **Privacy** | Data minimization, encryption, GDPR compliance | ✅ Implemented |

### Security Certifications Target

| Certification | Status | Target Date |
|---------------|--------|-------------|
| SOC 2 Type II | In Progress | Q3 2026 |
| ISO 27001 | Planned | Q4 2026 |
| GDPR Compliance | ✅ Compliant | Current |

---

## Security Design Principles

### 1. Defense in Depth

Multiple layers of security controls ensure that compromise of one layer does not lead to complete system compromise.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        DEFENSE IN DEPTH LAYERS                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Layer 1: PERIMETER                                                         │
│  ├─ Rate limiting, WAF, DDoS protection, IP filtering                       │
│                                                                              │
│  Layer 2: NETWORK                                                            │
│  ├─ TLS 1.3, mTLS, network segmentation, P2P encryption                     │
│                                                                              │
│  Layer 3: APPLICATION                                                        │
│  ├─ Input validation, authentication, authorization, session management     │
│                                                                              │
│  Layer 4: DATA                                                               │
│  ├─ Encryption at rest/transit, key management, data classification         │
│                                                                              │
│  Layer 5: CONSENSUS                                                          │
│  ├─ Byzantine fault tolerance, validator staking, slashing                  │
│                                                                              │
│  Layer 6: CRYPTOGRAPHIC                                                      │
│  └─ HSM key storage, threshold signatures, ZK proofs                        │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### 2. Zero Trust Architecture

- **Never trust, always verify**: All requests are authenticated and authorized
- **Least privilege access**: Minimum permissions required for each operation
- **Microsegmentation**: Network isolation between components
- **Continuous verification**: Session validation on every request

### 3. Security by Design

- **Secure defaults**: All configurations default to most secure option
- **Privacy by design**: Data minimization and encryption from inception
- **Fail secure**: System fails to a secure state on error
- **Complete mediation**: All access requests are checked

### 4. Transparency and Auditability

- **On-chain audit trail**: Immutable record of all state changes
- **Comprehensive logging**: Security events logged with context
- **Open source**: Core codebase publicly auditable
- **Third-party audits**: Regular external security assessments

---

## Architecture Overview

### System Components

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    VIRTENGINE SECURITY ARCHITECTURE                          │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  CLIENT LAYER                                                        │   │
│  │  ┌───────────┐  ┌───────────┐  ┌───────────┐  ┌───────────┐        │   │
│  │  │ Web Portal│  │ Mobile App│  │    CLI    │  │    SDK    │        │   │
│  │  │  (TLS 1.3)│  │ (TLS 1.3) │  │ (TLS 1.3) │  │ (TLS 1.3) │        │   │
│  │  └───────────┘  └───────────┘  └───────────┘  └───────────┘        │   │
│  └──────────────────────────────┬──────────────────────────────────────┘   │
│                                 │                                           │
│  ┌──────────────────────────────▼──────────────────────────────────────┐   │
│  │  API GATEWAY                                                         │   │
│  │  • Rate Limiting (100 req/min standard, 10 req/min auth endpoints)  │   │
│  │  • WAF (OWASP Core Rule Set)                                        │   │
│  │  • DDoS Protection (Cloudflare/AWS Shield)                          │   │
│  │  • TLS Termination                                                   │   │
│  │  • Request Validation                                                │   │
│  └──────────────────────────────┬──────────────────────────────────────┘   │
│                                 │                                           │
│  ┌──────────────────────────────▼──────────────────────────────────────┐   │
│  │  BLOCKCHAIN LAYER                                                    │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │  ANTE HANDLERS                                               │   │   │
│  │  │  • Signature Verification    • MFA Enforcement               │   │   │
│  │  │  • Fee Validation            • Rate Limiting                 │   │   │
│  │  │  • VEID Score Check          • Permission Validation         │   │   │
│  │  └──────────────────────────────────────────────────────────────┘   │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │  SECURITY MODULES                                            │   │   │
│  │  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐           │   │   │
│  │  │  │  VEID   │ │   MFA   │ │Encryption│ │  Roles  │           │   │   │
│  │  │  │ Identity│ │  Auth   │ │ Envelope │ │  RBAC   │           │   │   │
│  │  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘           │   │   │
│  │  └──────────────────────────────────────────────────────────────┘   │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐   │   │
│  │  │  CONSENSUS (CometBFT)                                        │   │   │
│  │  │  • Byzantine Fault Tolerant                                  │   │   │
│  │  │  • 2/3+ Validator Agreement Required                         │   │   │
│  │  │  • Validator Staking & Slashing                              │   │   │
│  │  └──────────────────────────────────────────────────────────────┘   │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  PROVIDER LAYER                                                      │   │
│  │  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │   │
│  │  │Provider Daemon│  │ TEE Runtime   │  │Infrastructure │           │   │
│  │  │ (Ed25519 sig) │  │ (SGX/SEV/Nitro)│ │   (K8s/SLURM) │           │   │
│  │  └───────────────┘  └───────────────┘  └───────────────┘           │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Trust Boundaries

| Boundary | Trust Level | Controls |
|----------|-------------|----------|
| Public Internet → API Gateway | Untrusted | TLS, rate limiting, WAF |
| API Gateway → Blockchain Node | Semi-trusted | Authentication, authorization |
| Blockchain Node → State | Trusted (consensus) | Byzantine agreement, signatures |
| Provider Daemon → Customer Workload | Semi-trusted | TEE, encryption, isolation |
| Validator → Identity Data | Trusted (threshold) | Multi-party decryption |

---

## Layer Security Model

### Layer 1: Perimeter Security

**Rate Limiting:**
- Standard endpoints: 100 requests/minute/IP
- Authentication endpoints: 10 requests/minute/IP
- VEID upload: 3 requests/minute/account
- Burst allowance: 2x limit for 10 seconds

**Web Application Firewall (WAF):**
- OWASP Core Rule Set 3.3
- Custom rules for blockchain-specific attacks
- SQL injection prevention
- XSS protection
- Path traversal blocking

**DDoS Protection:**
- L3/L4 protection via CDN/cloud provider
- L7 protection via rate limiting and WAF
- Automatic traffic pattern analysis
- Geographic blocking capability

### Layer 2: Network Security

**Transport Layer Security:**

| Configuration | Value |
|---------------|-------|
| Protocol | TLS 1.3 only (TLS 1.2 deprecated) |
| Cipher Suites | TLS_AES_256_GCM_SHA384, TLS_CHACHA20_POLY1305_SHA256 |
| Key Exchange | ECDHE with X25519 or P-256 |
| Certificate | ECDSA P-256 or RSA 2048+ (legacy) |
| HSTS | Enabled (max-age=31536000, includeSubdomains) |
| Certificate Pins | Backup pins for root CAs |

**P2P Network Security:**
- Noise Protocol for validator communication
- Peer authentication via Ed25519 public keys
- Encrypted gossip protocol
- Peer reputation scoring

**Network Segmentation:**
- Validator nodes in isolated network
- Provider daemons in separate subnet
- API endpoints in DMZ
- Database layer not internet-accessible

### Layer 3: Application Security

**Input Validation:**
- All inputs validated before processing
- Type-safe protobuf message parsing
- Maximum message size enforcement
- Canonical encoding verification

**Authentication:**
- Primary: Secp256k1 wallet signatures
- Secondary: VEID identity verification
- Optional: MFA (FIDO2, TOTP)

**Session Security:**
- JWT tokens with short expiry (15 minutes)
- Refresh tokens with secure rotation
- Session binding to IP/device fingerprint
- Automatic invalidation on anomaly

### Layer 4: Data Security

See [Data Protection](#data-protection) section for comprehensive coverage.

### Layer 5: Consensus Security

**Byzantine Fault Tolerance:**
- CometBFT consensus algorithm
- Tolerates up to 1/3 malicious validators
- Finality after 2/3+ agreement

**Validator Security:**
- Minimum stake requirement (100,000 VE)
- Slashing for:
  - Double signing: 5% stake
  - Downtime: 0.01% stake per missed block
  - Malicious attestation: 100% stake
- Jail period before unjailing allowed

**Consensus Determinism:**
- No `time.Now()` in state transitions
- No `crypto/rand` in keeper methods
- Fixed ML model versions for scoring
- Reproducible execution across validators

### Layer 6: Cryptographic Security

See [Cryptographic Architecture](#cryptographic-architecture) section.

---

## Identity and Access Management

### VEID Identity System

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                         VEID IDENTITY FLOW                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│   1. Capture identity documents + selfie (Approved Client)                  │
│   2. Salt binding + client signature                                        │
│   3. User signs + submits to chain                                          │
│   4. Validators decrypt (threshold)                                         │
│   5. ML Scoring: Document authenticity, Liveness, Face match, Anti-spoofing│
│   6. Consensus on score                                                     │
│   7. VEID Score (0-100) stored on-chain                                     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Multi-Factor Authentication (MFA)

**Supported Methods:**

| Method | Security Level | Implementation |
|--------|----------------|----------------|
| FIDO2/WebAuthn | Highest | Hardware key, passkeys |
| TOTP | High | Authenticator apps |
| Email | Medium | OTP codes |
| SMS | Low (not recommended) | OTP codes |

**MFA Requirements by Operation:**

| Operation | VEID Score | MFA Required |
|-----------|------------|--------------|
| View balances | Any | No |
| Small transfers (<100 VE) | ≥30 | No |
| Large transfers (≥1000 VE) | ≥50 | Yes |
| Provider registration | ≥70 | Yes |
| Validator operations | ≥80 | Yes (FIDO2) |
| Key rotation | Any | Yes |
| GDPR data export | Any | Yes |

### Role-Based Access Control (RBAC)

**Standard Roles:**

| Role | Permissions | Assignment |
|------|-------------|------------|
| `anonymous` | Read public data | Default |
| `customer` | Own data, transactions | Account creation |
| `provider` | Provider operations | Registration + stake |
| `validator` | Consensus, identity decrypt | Validator registration |
| `auditor` | Read audit logs | Governance assignment |
| `admin` | Emergency operations | Multi-sig governance |

---

## Cryptographic Architecture

### Algorithm Standards

| Purpose | Algorithm | Key Size | Standard |
|---------|-----------|----------|----------|
| Key Exchange | X25519 | 256-bit | RFC 7748 |
| Symmetric Encryption | XSalsa20-Poly1305 | 256-bit | NaCl |
| Digital Signatures | Ed25519 | 256-bit | RFC 8032 |
| Wallet Signatures | secp256k1 | 256-bit | SEC 2 |
| Hashing | SHA-256 | 256-bit | FIPS 180-4 |
| Password Hashing | Argon2id | N/A | RFC 9106 |
| Key Derivation | HKDF | 256-bit | RFC 5869 |
| ZK Proofs | Groth16 (BN254) | 256-bit | Academic |

### Encryption Envelope Format

All sensitive on-chain data uses the EncryptedPayloadEnvelope format:

```go
type EncryptedPayloadEnvelope struct {
    Algorithm     string   // "X25519-XSalsa20-Poly1305"
    Version       uint32   // Protocol version
    EphemeralKey  []byte   // 32 bytes, sender's ephemeral public key
    Nonce         []byte   // 24 bytes, random nonce
    Ciphertext    []byte   // Encrypted data
    RecipientKeys []WrappedKey  // Per-recipient encrypted DEK
    Signature     []byte   // Sender's signature over envelope
}

type WrappedKey struct {
    Fingerprint   string   // Recipient key fingerprint
    EncryptedDEK  []byte   // Encrypted data encryption key
}
```

### Key Hierarchy

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           KEY HIERARCHY                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  LEVEL 0: Root Keys (Air-gapped)                                            │
│  ├─ Genesis Authority Key (multi-sig 3-of-5)                                │
│  └─ Emergency Recovery Key (multi-sig 5-of-7)                               │
│                                                                              │
│  LEVEL 1: Validator Keys (HSM)                                              │
│  ├─ Consensus Signing Key (Ed25519)                                         │
│  ├─ Identity Decryption Key (X25519, threshold)                             │
│  └─ Tendermint P2P Key (Ed25519)                                            │
│                                                                              │
│  LEVEL 2: Service Keys (Secure Enclave)                                     │
│  ├─ Provider Daemon Signing Key (Ed25519)                                   │
│  ├─ API Server TLS Key (ECDSA P-256)                                        │
│  └─ Approved Client Signing Key (Ed25519)                                   │
│                                                                              │
│  LEVEL 3: User Keys (User Custody)                                          │
│  ├─ Account Key (secp256k1)                                                 │
│  ├─ Encryption Key (X25519, derived)                                        │
│  └─ MFA Secrets (TOTP, encrypted on-chain)                                  │
│                                                                              │
│  LEVEL 4: Ephemeral Keys (Per-operation)                                    │
│  ├─ Session Keys (derived from account key)                                 │
│  ├─ Encryption Ephemeral Keys (per-envelope)                                │
│  └─ ZK Proof Random Values (off-chain generation)                           │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Cryptographic Audit Status

See [SECURITY-001-CRYPTOGRAPHIC-AUDIT.md](../../SECURITY-001-CRYPTOGRAPHIC-AUDIT.md) for detailed cryptographic audit findings and remediation status.

---

## Data Protection

### Data Classification

VirtEngine uses a 5-level data classification scheme:

| Level | Name | Description | Encryption | Access |
|-------|------|-------------|------------|--------|
| C0 | Public | Freely available | None required | Anyone |
| C1 | Internal | Not published | TLS in transit | Authenticated |
| C2 | Confidential | Business-sensitive | At rest + transit | Authorized |
| C3 | Restricted | Highly sensitive | End-to-end | Recipient only |
| C4 | Critical | Maximum protection | HSM + threshold | Multi-party |

See [_docs/data-classification.md](../../_docs/data-classification.md) for complete data inventory.

### Encryption at Rest

| Data Store | Encryption | Key Management |
|------------|------------|----------------|
| Blockchain state | Public ledger (some fields encrypted) | Per-recipient keys |
| Off-chain database | AES-256-GCM | AWS KMS / HashiCorp Vault |
| Backups | AES-256-GCM | Separate backup keys |
| Logs | AES-256-GCM | Automatic rotation |
| Provider storage | Customer-controlled | Customer keys |

### Encryption in Transit

| Communication Path | Encryption | Authentication |
|--------------------|------------|----------------|
| Client → API | TLS 1.3 | Server certificate |
| API → Blockchain Node | TLS 1.3 | mTLS (internal) |
| Validator → Validator | Noise Protocol | Peer Ed25519 keys |
| Provider → Customer | TLS 1.3 + app-layer | mTLS optional |

### Data Retention

| Data Type | Retention Period | Disposal Method |
|-----------|------------------|-----------------|
| Identity documents (encrypted) | Until user deletion | Key destruction |
| Transaction history | Permanent (blockchain) | N/A |
| Audit logs | 7 years | Secure deletion |
| Session data | 30 days | Automatic purge |
| Backups | 90 days rolling | Cryptographic erasure |

---

## Network Security

### Network Zones

| Zone | Purpose | Access Control |
|------|---------|----------------|
| ZONE 1: PUBLIC | Internet-facing (CDN, WAF, DDoS) | HTTPS only |
| ZONE 2: DMZ | API Gateways, Load Balancers | mTLS |
| ZONE 3: APPLICATION | Blockchain Nodes, App Servers | mTLS + Network ACLs |
| ZONE 4: DATA | Databases, Key Management | mTLS + Network ACLs |
| ZONE 5: VALIDATOR | HSMs, Consensus Keys | Air-gapped where possible |

### Firewall Rules (Summary)

| Source | Destination | Ports | Protocol |
|--------|-------------|-------|----------|
| Internet | CDN/WAF | 443 | HTTPS |
| CDN | API Gateway | 443 | HTTPS |
| API Gateway | Blockchain Node | 26657 | Tendermint RPC |
| Blockchain Nodes | Blockchain Nodes | 26656 | P2P (Noise) |
| Application | Database | 5432 | PostgreSQL (TLS) |
| Validator | HSM | Vendor-specific | PKCS#11 |

### Intrusion Detection

- Network-based IDS at zone boundaries
- Host-based IDS on critical servers
- Log correlation via SIEM
- Anomaly detection on traffic patterns
- Real-time alerting on suspicious activity

---

## Secure Development Lifecycle

### Security in Development

| Phase | Security Activities |
|-------|---------------------|
| DESIGN | Threat modeling, security requirements, data flow analysis |
| DEV | SAST, linting, code review, secrets scanning |
| TEST | DAST, fuzzing, pentest, regression |
| DEPLOY | Artifact signing, immutable deploy, rollback ready |
| MONITOR | Anomaly detection, vulnerability scanning, incident response |

### Code Security Controls

| Control | Tool | Frequency |
|---------|------|-----------|
| Static Analysis (SAST) | golangci-lint, gosec | Every commit |
| Dependency Scanning | Dependabot, Snyk | Daily |
| Secret Detection | gitleaks | Every commit |
| Container Scanning | Trivy | Every build |
| License Compliance | FOSSA | Weekly |

### Security Testing

| Test Type | Scope | Frequency |
|-----------|-------|-----------|
| Unit Tests (crypto) | All crypto code | Every commit |
| Fuzz Testing | Input parsers, encoding | Continuous |
| Integration Tests | Module interactions | Every PR |
| Penetration Testing | External attack surface | Annually |
| Blockchain Audit | Cosmos modules | Per release |

---

## Security Monitoring and Response

### Security Monitoring

**Metrics Collected:**
- Authentication failures
- Rate limit hits
- WAF blocks
- Transaction anomalies
- Validator misbehavior
- Key usage patterns

**Alerting Thresholds:**

| Alert | Threshold | Severity |
|-------|-----------|----------|
| Auth failures (single IP) | >100/hour | High |
| Auth failures (account) | >10/hour | Critical |
| WAF blocks (attack pattern) | >50/minute | High |
| Validator double-sign | Any | Critical |
| Key usage anomaly | >2σ from baseline | High |

### Incident Response

VirtEngine maintains a comprehensive incident response capability:

1. **Detection**: Automated monitoring + user reports
2. **Triage**: Severity classification (SEV-1 to SEV-4)
3. **Containment**: Immediate actions to limit impact
4. **Eradication**: Remove threat from environment
5. **Recovery**: Restore normal operations
6. **Lessons Learned**: Post-incident review

See [docs/sre/INCIDENT_RESPONSE.md](../sre/INCIDENT_RESPONSE.md) for detailed procedures.

### Security Incident Categories

| Category | Examples | Response SLA |
|----------|----------|--------------|
| Critical (SEV-1) | Active breach, data exfiltration | 15 minutes |
| High (SEV-2) | Vulnerability exploitation attempt | 1 hour |
| Medium (SEV-3) | Suspicious activity pattern | 4 hours |
| Low (SEV-4) | Policy violation, minor anomaly | 24 hours |

---

## Compliance Integration

### Compliance Frameworks

| Framework | Status | Scope |
|-----------|--------|-------|
| **GDPR** | ✅ Compliant | Personal data, EU users |
| **SOC 2 Type II** | In Progress | Security, availability, confidentiality |
| **ISO 27001** | Planned | Information security management |
| **HIPAA** | Not applicable | No healthcare data |
| **PCI DSS** | Partial | Payment-related operations |

### Compliance Controls Mapping

| Control Area | GDPR | SOC 2 | ISO 27001 |
|--------------|------|-------|-----------|
| Access Control | Art. 32 | CC6.1 | A.9 |
| Encryption | Art. 32 | CC6.1 | A.10 |
| Logging & Monitoring | Art. 30 | CC7.2 | A.12.4 |
| Incident Response | Art. 33 | CC7.3 | A.16 |
| Data Retention | Art. 5 | CC6.5 | A.8.2 |
| Vendor Management | Art. 28 | CC9.2 | A.15 |

See [COMPLIANCE_MATRIX.md](./COMPLIANCE_MATRIX.md) for detailed control mappings.

---

## Related Documentation

### Security Documentation

| Document | Purpose | Location |
|----------|---------|----------|
| Threat Model | STRIDE analysis of all threats | [_docs/threat-model.md](../../_docs/threat-model.md) |
| Data Classification | Data sensitivity levels | [_docs/data-classification.md](../../_docs/data-classification.md) |
| Key Management | Key lifecycle procedures | [_docs/key-management.md](../../_docs/key-management.md) |
| Cryptographic Audit | Algorithm security review | [SECURITY-001-CRYPTOGRAPHIC-AUDIT.md](../../SECURITY-001-CRYPTOGRAPHIC-AUDIT.md) |
| Penetration Testing | Security testing program | [PENETRATION_TESTING_PROGRAM.md](../../PENETRATION_TESTING_PROGRAM.md) |
| Incident Response | IR procedures | [docs/sre/INCIDENT_RESPONSE.md](../sre/INCIDENT_RESPONSE.md) |

### Compliance Documentation

| Document | Purpose | Location |
|----------|---------|----------|
| GDPR Compliance | EU data protection | [GDPR_COMPLIANCE.md](../../GDPR_COMPLIANCE.md) |
| Compliance Matrix | Multi-framework mapping | [COMPLIANCE_MATRIX.md](./COMPLIANCE_MATRIX.md) |
| Privacy Policy | User-facing privacy notice | [PRIVACY_POLICY.md](../../PRIVACY_POLICY.md) |
| Data Processing Agreement | Processor terms | [DATA_PROCESSING_AGREEMENT.md](../../DATA_PROCESSING_AGREEMENT.md) |

### Operational Documentation

| Document | Purpose | Location |
|----------|---------|----------|
| SRE Playbooks | Operational procedures | [docs/sre/](../sre/) |
| On-Call Runbook | Incident handling | [docs/operations/ON_CALL_RUNBOOK.md](../operations/ON_CALL_RUNBOOK.md) |
| Business Continuity | DR procedures | [_docs/business-continuity.md](../../_docs/business-continuity.md) |

---

**Document Owner**: Security Team  
**Last Updated**: 2026-01-30  
**Version**: 1.0.0  
**Next Review**: 2026-04-30
