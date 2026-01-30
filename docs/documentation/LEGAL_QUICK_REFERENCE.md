# VirtEngine Legal Documents - Quick Reference Guide

**Version:** 1.0  
**Last Updated:** January 29, 2026

## Document Overview

| Document | Purpose | Audience | Must Read Before |
|----------|---------|----------|------------------|
| [LEGAL_README.md](LEGAL_README.md) | Navigation and quick start | All users | Using any service |
| [TERMS_OF_SERVICE.md](TERMS_OF_SERVICE.md) | Master user agreement | All users | Account creation |
| [PRIVACY_POLICY.md](PRIVACY_POLICY.md) | Data protection practices | All users | Providing personal data |
| [BIOMETRIC_DATA_ADDENDUM.md](BIOMETRIC_DATA_ADDENDUM.md) | Biometric data notice | VEID users | Biometric verification |
| [COOKIE_POLICY.md](COOKIE_POLICY.md) | Cookie usage | Website users | First visit |
| [ACCEPTABLE_USE_POLICY.md](ACCEPTABLE_USE_POLICY.md) | Prohibited activities | Providers & Tenants | Deploying workloads |
| [DATA_PROCESSING_AGREEMENT.md](DATA_PROCESSING_AGREEMENT.md) | Processor obligations | Providers (data processors) | Processing tenant data |
| [CONSENT_FRAMEWORK.md](CONSENT_FRAMEWORK.md) | Technical consent system | Developers & Users | Implementing consent |
| [LEGAL_COMPLIANCE_CHECKLIST.md](LEGAL_COMPLIANCE_CHECKLIST.md) | Launch readiness | Internal teams | Production launch |

---

## Key Sections by Use Case

### "I want to use VirtEngine services"
1. Read [TERMS_OF_SERVICE.md](TERMS_OF_SERVICE.md) - Sections 1-5 (Introduction, Services, Eligibility)
2. Read [PRIVACY_POLICY.md](PRIVACY_POLICY.md) - Section 3 (What We Collect)
3. Review [ACCEPTABLE_USE_POLICY.md](ACCEPTABLE_USE_POLICY.md) - Section 3 (Prohibited Activities)

### "I want to use VEID identity verification"
1. Read [PRIVACY_POLICY.md](PRIVACY_POLICY.md) - Section 3.2 (Biometric Data)
2. **MUST READ:** [BIOMETRIC_DATA_ADDENDUM.md](BIOMETRIC_DATA_ADDENDUM.md) - Entire document
3. Review [CONSENT_FRAMEWORK.md](CONSENT_FRAMEWORK.md) - Section 3 (Identity Scopes)
4. Sign consent form (Section 13 of Biometric Addendum)

### "I want to become a Provider"
1. Read [TERMS_OF_SERVICE.md](TERMS_OF_SERVICE.md) - Section 6 (Provider Terms)
2. Read [DATA_PROCESSING_AGREEMENT.md](DATA_PROCESSING_AGREEMENT.md) - Entire document
3. Read [ACCEPTABLE_USE_POLICY.md](ACCEPTABLE_USE_POLICY.md) - Section 10.1 (Provider Responsibilities)
4. Execute DPA before processing tenant data

### "I want to become a Tenant"
1. Read [TERMS_OF_SERVICE.md](TERMS_OF_SERVICE.md) - Section 7 (Tenant Terms)
2. Read [ACCEPTABLE_USE_POLICY.md](ACCEPTABLE_USE_POLICY.md) - Sections 3.6.1, 3.7 (Prohibited Uses)
3. Review provider-specific policies (if more restrictive)

### "I'm building on VirtEngine (Developer)"
1. Read [CONSENT_FRAMEWORK.md](CONSENT_FRAMEWORK.md) - Section 8 (Implementation Guidelines)
2. Study `x/veid/types/consent.go` - Code reference
3. Review [DATA_PROCESSING_AGREEMENT.md](DATA_PROCESSING_AGREEMENT.md) - Section 5 (Security Measures)
4. Implement consent checks in code (see Consent Framework Section 4.3)

### "I'm launching VirtEngine to production (Internal)"
1. Complete [LEGAL_COMPLIANCE_CHECKLIST.md](LEGAL_COMPLIANCE_CHECKLIST.md) - All sections
2. Engage legal counsel for document review
3. Execute vendor agreements with DPA clauses
4. Appoint DPO and publish contact information

---

## Jurisdiction-Specific Quick Links

### European Union / EEA
**Applicable Laws:** GDPR, ePrivacy Directive  
**Key Sections:**
- Privacy Policy → Section 15.1 (EU Notice)
- DPA → Section 12.1 (Standard Contractual Clauses)
- Privacy Policy → Section 9.1 (GDPR Rights)

**Your Rights:**
- Access, rectify, erase, restrict, port, object
- Contact: dpo@virtengine.com

### United Kingdom
**Applicable Laws:** UK GDPR, PECR  
**Key Sections:**
- Privacy Policy → Section 15.2 (UK Notice)
- DPA → Section 12.1 (UK Addendum to SCCs)
- Cookie Policy → Section 8.1 (ePrivacy compliance)

**Your Rights:**
- Same as GDPR
- ICO: https://ico.org.uk/
- Contact: dpo@virtengine.com

### California (United States)
**Applicable Laws:** CCPA/CPRA  
**Key Sections:**
- Privacy Policy → Section 15.3 (CCPA Disclosure)
- Privacy Policy → Section 9.2 (CCPA Rights)
- Biometric Addendum → Section 10.1 (California biometric law)

**Your Rights:**
- Know, delete, correct, opt-out, limit sensitive PI
- Contact: privacy@virtengine.com

### Illinois (United States)
**Applicable Laws:** BIPA (Biometric Information Privacy Act)  
**Key Document:** [BIOMETRIC_DATA_ADDENDUM.md](BIOMETRIC_DATA_ADDENDUM.md)
**Key Sections:**
- Section 3 (Notice of Collection)
- Section 4 (Written Consent)
- Section 5 (Prohibitions - No Selling)
- Section 9 (Illinois-Specific Provisions)

**Your Rights:**
- Private right of action for violations
- Statutory damages ($1,000 negligent, $5,000 intentional)
- Contact: legal@virtengine.com (subject: "BIPA Inquiry")

### Australia
**Applicable Laws:** Privacy Act 1988, APPs  
**Key Sections:**
- Privacy Policy → Section 15.7 (Australia)
- OAIC: https://www.oaic.gov.au/

**Your Rights:**
- Access, correction, complaint
- Contact: privacy@virtengine.com

### Canada
**Applicable Laws:** PIPEDA  
**Key Sections:**
- Privacy Policy → Section 15.6 (Canada)
- DPA → Section 7.2 (International Transfers)

**Your Rights:**
- Access, correction, complaint
- Contact: privacy@virtengine.com

### Brazil
**Applicable Laws:** LGPD (Lei Geral de Proteção de Dados)  
**Key Sections:**
- Privacy Policy → Section 15.5 (Brazil)
- DPA → Section 7.2 (International Transfers)

**Your Rights:**
- Similar to GDPR (access, correction, deletion, portability)
- Contact: dpo@virtengine.com

### Singapore
**Applicable Laws:** PDPA (Personal Data Protection Act)  
**Key Sections:**
- Privacy Policy (general provisions apply)

**Your Rights:**
- Access, correction
- Contact: privacy@virtengine.com

---

## Common Questions

### "Do I have to accept all terms?"
**Yes.** Using VirtEngine services requires acceptance of the Terms of Service. For VEID biometric verification, you must also explicitly consent to the Biometric Data Addendum.

### "Can I use VirtEngine without biometric verification?"
**Yes.** Biometric verification (VEID) is optional. However, some marketplace features or providers may require verification.

### "How do I withdraw consent for biometric data?"
See [CONSENT_FRAMEWORK.md](CONSENT_FRAMEWORK.md) Section 7.1 or email dpo@virtengine.com with subject "Withdraw Biometric Consent".

### "Can my blockchain transactions be reversed?"
**No.** Blockchain transactions are final and irreversible. See Terms of Service Section 5.1.

### "Can I delete my data from the blockchain?"
**Partially.** On-chain data is immutable, but we use encryption. Deleting encryption keys makes the data unreadable (functional erasure). See Privacy Policy Section 9.5.

### "Do you sell my data?"
**No.** We do not sell personal data. We especially never sell biometric data (absolute prohibition). See Privacy Policy Section 6.6 and Biometric Addendum Section 5.1.

### "How long do you keep my biometric data?"
**Maximum 7 years** from last use (KYC/AML compliance). Typically deleted 3 years after account closure. See Biometric Addendum Section 3.3.

### "What if there's a data breach?"
We notify affected users within **72 hours** and provide remediation resources. See Privacy Policy Section 5.3 and DPA Section 5.4.

### "Who do I contact for privacy questions?"
- **General Privacy:** privacy@virtengine.com
- **Data Protection Officer:** dpo@virtengine.com
- **BIPA (Illinois):** legal@virtengine.com (subject: "BIPA Inquiry")

### "Where can I report abuse?"
Email abuse@virtengine.com with evidence. See Acceptable Use Policy Section 12.

### "How do I report a security vulnerability?"
Email security@virtengine.com following responsible disclosure. See Acceptable Use Policy Section 7.

---

## Glossary of Key Terms

**Biometric Data:** Facial geometry, facial recognition templates, liveness detection metrics.

**Blockchain:** Public, immutable distributed ledger. Data cannot be deleted once recorded.

**Consent:** Explicit agreement to data processing. Can be withdrawn at any time.

**Data Controller:** Entity determining purposes and means of data processing (typically you or providers).

**Data Processor:** Entity processing data on behalf of controller (VirtEngine for provider services).

**DPO:** Data Protection Officer - oversees data protection compliance.

**GDPR:** General Data Protection Regulation (EU/EEA data protection law).

**Lease:** Agreement between tenant and provider for computing resources.

**Provider:** Entity offering computing capacity in the marketplace.

**Scope:** Category of identity data (e.g., biometric, document, basic).

**Special Category Data:** Sensitive data requiring explicit consent (biometric, health, genetic).

**Tenant:** User renting computing resources from providers.

**VEID:** Verifiable Electronic Identity - VirtEngine's ML-powered identity verification system.

---

## Document Relationships

```
LEGAL_README (Start Here)
    ↓
TERMS_OF_SERVICE (Master Agreement)
    ├── PRIVACY_POLICY (Data Protection)
    │   ├── BIOMETRIC_DATA_ADDENDUM (BIPA Compliance)
    │   └── COOKIE_POLICY (Web Tracking)
    ├── ACCEPTABLE_USE_POLICY (Prohibited Activities)
    └── DATA_PROCESSING_AGREEMENT (Provider Obligations)

CONSENT_FRAMEWORK (Technical Implementation)
    ↑
Referenced by all above documents

LEGAL_COMPLIANCE_CHECKLIST (Internal Launch Readiness)
    ↑
Validates all above documents
```

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-29 | Initial release - all documents drafted |

---

## Next Steps

1. **New User:** Start with [LEGAL_README.md](LEGAL_README.md)
2. **VEID User:** Read [BIOMETRIC_DATA_ADDENDUM.md](BIOMETRIC_DATA_ADDENDUM.md)
3. **Provider:** Read [DATA_PROCESSING_AGREEMENT.md](DATA_PROCESSING_AGREEMENT.md)
4. **Developer:** Read [CONSENT_FRAMEWORK.md](CONSENT_FRAMEWORK.md)
5. **Questions:** Email legal@virtengine.com or privacy@virtengine.com

---

**Need Help?**  
Email: legal@virtengine.com | privacy@virtengine.com | dpo@virtengine.com  
Web: https://virtengine.com/legal/
