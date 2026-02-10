# VirtEngine Data Inventory and Classification

**Version:** 1.0  
**Last Updated:** January 30, 2026  
**Review Frequency:** Quarterly  
**GDPR Reference:** Article 30 (Records of Processing Activities)

This document provides a comprehensive inventory of all personal data processed by VirtEngine, including data classification, legal basis, retention periods, and processing activities.

---

## 1. Executive Summary

VirtEngine processes personal data across multiple categories:
- **Standard Personal Data:** Account information, transaction data, communication records
- **Special Category Data (GDPR Art. 9):** Biometric data for identity verification
- **Pseudonymized Data:** Blockchain addresses, hashed identifiers
- **Technical Data:** IP addresses, device fingerprints, logs

This inventory supports GDPR compliance requirements including:
- Records of Processing Activities (Art. 30)
- Data Protection Impact Assessments (Art. 35)
- Data Subject Rights fulfillment (Arts. 15-22)
- Cross-border transfer documentation (Art. 46)

---

## 2. Data Classification Framework

### 2.1 Classification Levels

| Level | Name | Description | Protection Requirements |
|-------|------|-------------|------------------------|
| C1 | Public | Blockchain data, public keys | Standard security |
| C2 | Internal | Analytics, aggregated metrics | Access controls |
| C3 | Confidential | Personal data, account info | Encryption, RBAC, audit logging |
| C4 | Restricted | Biometric data, identity documents | Enhanced encryption, MFA, strict access, secure enclaves |

### 2.2 Data Sensitivity Categories

| Category | GDPR Classification | Examples |
|----------|---------------------|----------|
| Standard PII | Personal Data (Art. 4) | Name, email, phone, address |
| Special Category | Art. 9 Data | Biometric templates, health data |
| Pseudonymous | Pseudonymized Data | Wallet addresses, hashed IDs |
| Anonymous | Non-Personal Data | Aggregated analytics |

---

## 3. Detailed Data Inventory

### 3.1 Identity Verification Data (VEID Module)

#### 3.1.1 Biometric Data (Classification: C4 - Restricted)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Facial geometry templates | Face embedding vectors | Off-chain (encrypted) | Active + 3 years | Explicit consent (Art. 9(2)(a)) | Special category |
| Facial recognition hashes | SHA-256 hash | On-chain (encrypted envelope) | Indefinite (blockchain) | Explicit consent | Special category |
| Liveness detection scores | Numeric scores | Off-chain | 90 days | Explicit consent | Special category |
| Face feature hashes | Derived hash | On-chain | Indefinite (blockchain) | Explicit consent | Special category |
| Raw facial images | JPEG/PNG | Temporary (memory) | Deleted after processing | Explicit consent | Special category |

**Processing Purpose:** Identity verification, fraud prevention, KYC/AML compliance

**Access Controls:**
- Validator public key encryption (X25519-XSalsa20-Poly1305)
- MFA required for administrative access
- Secure enclave processing where available
- Audit logging of all access

#### 3.1.2 Identity Document Data (Classification: C4 - Restricted)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Document scans | Encrypted images | Off-chain | 7 years (KYC) | Explicit consent, Legal obligation | Standard PII |
| OCR-extracted text | Structured data | Off-chain (encrypted) | 7 years (KYC) | Explicit consent | Standard PII |
| Document number | Hashed only | On-chain | Indefinite | Explicit consent | Standard PII |
| Document type | Category code | On-chain | Indefinite | Explicit consent | Standard PII |
| Document authenticity scores | Numeric | Off-chain | 7 years | Explicit consent | Standard PII |

**Processing Purpose:** KYC/AML compliance, identity verification, fraud prevention

#### 3.1.3 Verification Metadata (Classification: C3 - Confidential)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Verification timestamps | ISO 8601 | On-chain | Indefinite | Contract, Legitimate interest | Standard PII |
| ML model versions | Version string | On-chain | Indefinite | Legitimate interest | Non-personal |
| Verification scores | Numeric | On-chain | Indefinite | Contract | Standard PII |
| Verification attempt count | Counter | On-chain | Indefinite | Legitimate interest | Pseudonymous |
| Device fingerprints | Hashed | Off-chain | 12 months | Legitimate interest | Pseudonymous |
| IP addresses (verification) | IPv4/IPv6 | Off-chain | 12 months | Legitimate interest | Standard PII |

### 3.2 Wallet and Account Data

#### 3.2.1 Identity Wallet (Classification: C3 - Confidential)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Wallet address | Bech32 | On-chain | Indefinite | Contract | Pseudonymous |
| Wallet ID | UUID | On-chain | Indefinite | Contract | Pseudonymous |
| Consent settings | JSON | On-chain | Active | Consent | Standard PII |
| Trust scores | Numeric | On-chain | Indefinite | Contract | Pseudonymous |
| Delegation relationships | Addresses | On-chain | Indefinite | Contract | Pseudonymous |
| Wallet status | Enum | On-chain | Indefinite | Contract | Pseudonymous |

#### 3.2.2 Consent Records (Classification: C3 - Confidential)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Scope consents | Map | On-chain | Active + 7 years | Legal obligation | Standard PII |
| Consent timestamps | ISO 8601 | On-chain | 7 years | Legal obligation | Standard PII |
| Consent purposes | Text | On-chain | 7 years | Legal obligation | Standard PII |
| Provider grants | Addresses | On-chain | Active + 7 years | Legal obligation | Standard PII |
| Revocation records | Timestamps | On-chain | 7 years | Legal obligation | Standard PII |

### 3.3 Marketplace Data

#### 3.3.1 Provider Data (Classification: C2 - Internal)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Provider address | Bech32 | On-chain | Indefinite | Contract | Pseudonymous |
| Infrastructure specs | JSON | On-chain | Active | Contract | Non-personal |
| Resource metrics | Time series | Off-chain | 3 years | Contract | Non-personal |
| Provider daemon logs | Text | Off-chain | 90 days | Legitimate interest | Non-personal |
| Service availability | Metrics | Off-chain | 3 years | Contract | Non-personal |

#### 3.3.2 Tenant Data (Classification: C2 - Internal)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Tenant address | Bech32 | On-chain | Indefinite | Contract | Pseudonymous |
| Deployment manifests | YAML/JSON | On-chain | Active | Contract | Non-personal |
| Resource consumption | Metrics | Off-chain | 3 years | Contract | Non-personal |
| Lease records | Structured | On-chain | Indefinite | Contract | Pseudonymous |
| Billing records | Structured | On-chain | 7 years | Legal obligation | Standard PII |

### 3.4 Blockchain Transaction Data

#### 3.4.1 On-Chain Transactions (Classification: C1 - Public)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Transaction hashes | Hex | On-chain | Indefinite | Contract | Non-personal |
| Sender addresses | Bech32 | On-chain | Indefinite | Contract | Pseudonymous |
| Transaction amounts | Numeric | On-chain | Indefinite | Contract | Pseudonymous |
| Block timestamps | Unix | On-chain | Indefinite | Contract | Non-personal |
| Message types | String | On-chain | Indefinite | Contract | Non-personal |
| Gas fees | Numeric | On-chain | Indefinite | Contract | Non-personal |

### 3.5 Communication Data

#### 3.5.1 Support Communications (Classification: C3 - Confidential)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Email addresses | Email | Off-chain | 3 years | Contract, Consent | Standard PII |
| Support tickets | Text | Off-chain | 3 years | Contract | Standard PII |
| Communication history | Text | Off-chain | 3 years | Contract | Standard PII |
| Feedback responses | Text | Off-chain | 3 years | Consent | Standard PII |

### 3.6 Technical and Analytics Data

#### 3.6.1 Security Logs (Classification: C2 - Internal)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| IP addresses | IPv4/IPv6 | Off-chain | 12 months | Legitimate interest | Standard PII |
| User agents | String | Off-chain | 12 months | Legitimate interest | Standard PII |
| Access timestamps | ISO 8601 | Off-chain | 12 months | Legitimate interest | Standard PII |
| Error logs | Text | Off-chain | 90 days | Legitimate interest | Non-personal |
| Security events | Structured | Off-chain | 7 years | Legal obligation | Standard PII |

#### 3.6.2 Analytics Data (Classification: C2 - Internal)

| Data Element | Type | Storage | Retention | Legal Basis | GDPR Category |
|--------------|------|---------|-----------|-------------|---------------|
| Page views | Counter | Off-chain | 26 months | Legitimate interest | Pseudonymous |
| Feature usage | Metrics | Off-chain | 26 months | Legitimate interest | Non-personal |
| Session data | Structured | Off-chain | 26 months | Legitimate interest | Pseudonymous |
| Geographic region | Country/region | Off-chain | 26 months | Legitimate interest | Pseudonymous |

---

## 4. Data Flow Diagrams

### 4.1 Identity Verification Data Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐    ┌──────────────┐
│  User       │───▶│  Capture App │───▶│  ML Pipeline │───▶│  Blockchain  │
│  Device     │    │  (Approved)  │    │  (Secure)    │    │  (On-chain)  │
└─────────────┘    └──────────────┘    └─────────────┘    └──────────────┘
      │                   │                   │                   │
      │                   │                   │                   │
      ▼                   ▼                   ▼                   ▼
 Raw Image          Encrypted           Embeddings           Hashes Only
 (temp)             Upload              (temp)               (permanent)
                                                                  │
                                              ┌───────────────────┘
                                              ▼
                                    ┌──────────────┐
                                    │  Off-chain   │
                                    │  Encrypted   │
                                    │  Storage     │
                                    └──────────────┘
                                          │
                                          ▼
                                    Encrypted embeddings
                                    (retention policy)
```

### 4.2 Consent Data Flow

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│  User       │───▶│  Consent UI  │───▶│  Blockchain  │
│             │    │  or CLI      │    │  (On-chain)  │
└─────────────┘    └──────────────┘    └─────────────┘
                         │                   │
                         │                   │
                         ▼                   ▼
                   Consent grants       Consent state
                   and revokes          (auditable)
```

---

## 5. International Data Transfers

### 5.1 Transfer Locations

| Data Category | Transfer Destinations | Transfer Mechanism | Safeguards |
|---------------|----------------------|-------------------|------------|
| Blockchain data | Global (decentralized) | Blockchain consensus | Encryption, public/pseudonymous |
| Biometric data | AU, US, EU | Standard Contractual Clauses | E2E encryption, access controls |
| Support data | AU, US | Standard Contractual Clauses | Encryption |
| Analytics | EU | Adequacy decision | Privacy-focused platform |

### 5.2 Sub-Processors

| Sub-Processor | Service | Location | Data Processed |
|---------------|---------|----------|----------------|
| AWS | Cloud infrastructure | Global | Encrypted backups |
| Azure | Cloud infrastructure | Global | Provider workloads |
| Stripe | Payment processing | US, EU | Payment data |

---

## 6. Data Subject Rights Mapping

### 6.1 Rights Implementation

| GDPR Right | Article | Implementation | Limitations |
|------------|---------|----------------|-------------|
| Access | Art. 15 | Data export via CLI/API | Blockchain public by design |
| Rectification | Art. 16 | Update via wallet | Blockchain immutable |
| Erasure | Art. 17 | Key destruction, off-chain deletion | Blockchain immutable (functional erasure) |
| Restriction | Art. 18 | Consent revocation | Processing stops |
| Portability | Art. 20 | JSON export | Machine-readable format |
| Object | Art. 21 | Consent revocation | Processing stops |
| Automated decisions | Art. 22 | Human review available | ML scoring |

### 6.2 Blockchain Immutability Considerations

**Challenge:** GDPR right to erasure conflicts with blockchain immutability.

**Solution: Functional Erasure**
1. All sensitive data encrypted before blockchain submission
2. Erasure implemented via encryption key destruction
3. On-chain data becomes permanently unreadable
4. Off-chain data deleted within 30 days
5. Backup rotation ensures complete removal within 90 days

---

## 7. Retention Schedule Summary

| Data Category | Minimum | Maximum | Trigger for Deletion |
|---------------|---------|---------|---------------------|
| Biometric templates | Active | Active + 3 years | Account closure + retention period |
| Identity documents | 7 years | 7 years | KYC/AML requirement |
| Consent records | Active | 7 years | Audit/legal requirement |
| Transaction data | 7 years | Indefinite (blockchain) | Tax/legal requirement |
| Support communications | 3 years | 3 years | Last interaction + 3 years |
| Security logs | 12 months | 12 months | Rolling deletion |
| Analytics | 26 months | 26 months | Rolling deletion |
| Raw images | 0 | 30 days | Delete after processing |

---

## 8. Data Protection Measures

### 8.1 Technical Measures

| Measure | Implementation | Data Categories Protected |
|---------|----------------|--------------------------|
| Encryption at rest | AES-256 | All off-chain data |
| Encryption in transit | TLS 1.3 | All data transfers |
| Identity data encryption | X25519-XSalsa20-Poly1305 | Biometric, identity docs |
| Access controls | RBAC | All data |
| MFA | Hardware/software tokens | Admin access |
| Secure enclaves | SGX, SEV, Nitro | Biometric processing |
| Audit logging | Append-only | All data access |
| Key management | HSM | Encryption keys |

### 8.2 Organizational Measures

| Measure | Description |
|---------|-------------|
| Data Protection Officer | Appointed and contactable |
| Privacy by design | Built into development process |
| Employee training | Annual GDPR training |
| Background checks | For data access roles |
| Confidentiality agreements | All employees |
| Incident response | 72-hour notification process |
| Vendor management | DPAs with all processors |

---

## 9. Audit and Review

### 9.1 Review Schedule

| Review Type | Frequency | Responsible Party |
|-------------|-----------|-------------------|
| Data inventory update | Quarterly | DPO |
| Retention compliance | Semi-annually | Compliance team |
| Access control audit | Annually | Security team |
| DPIA review | Annually or on change | DPO |
| Sub-processor review | Annually | Legal |

### 9.2 Document Control

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 1.0 | 2026-01-30 | DPO | Initial release |

---

## 10. Contacts

**Data Protection Officer:** dpo@virtengine.com  
**Privacy Inquiries:** privacy@virtengine.com  
**Security Team:** security@virtengine.com

---

*This document is confidential and intended for internal use and regulatory compliance purposes.*
