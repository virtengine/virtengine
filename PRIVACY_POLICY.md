# VirtEngine Privacy Policy

**Last Updated:** January 29, 2026  
**Effective Date:** January 29, 2026

## 1. Introduction

DET-IO Pty. Ltd. ("VirtEngine", "we", "us", or "our") is committed to protecting your privacy and handling your personal data transparently and securely. This Privacy Policy explains how we collect, use, disclose, and protect your information when you use the VirtEngine decentralized cloud computing marketplace, blockchain network, and associated services (the "Services").

This Privacy Policy applies to all users of the Services, including tenants, providers, and validators.

**Key Points:**

- We process biometric data (special category data under GDPR) for identity verification
- Blockchain data is public and immutable
- You control data sharing through granular consent mechanisms
- We use end-to-end encryption for sensitive identity data
- We comply with GDPR, CCPA, BIPA, and other data protection laws

## 2. Data Controller Information

**Data Controller:**  
DET-IO Pty. Ltd.  
[Registered Office Address]  
Australia

**Data Protection Officer:**  
Email: dpo@virtengine.com  
Privacy Inquiries: privacy@virtengine.com

**EU Representative** (if applicable under GDPR):  
[EU Representative Details]

## 3. Information We Collect

### 3.1 Blockchain and Wallet Data

**What We Collect:**

- Wallet addresses (public keys)
- Transaction history on the VirtEngine blockchain
- Staking and delegation records
- Marketplace orders, bids, and leases
- Escrow transactions
- Validator voting and governance activities

**Legal Basis (GDPR):** Legitimate interests (providing blockchain services), contract performance

**Retention:** Permanent (blockchain immutability)

**Note:** Blockchain data is publicly accessible and cannot be deleted or modified once recorded.

### 3.2 Identity Verification Data (VEID)

**What We Collect:**

**Biometric Data:**

- Facial geometry and recognition templates
- Liveness detection metrics
- Facial feature hashes (derived data)
- Verification session metadata

**Identity Documents:**

- Government-issued ID scans (driver's license, passport, national ID)
- OCR-extracted text data (name, date of birth, document number)
- Document authenticity scores

**Verification Metadata:**

- Verification timestamps
- ML model versions and scores
- Verification attempt history
- Device fingerprints
- IP addresses during verification

**Legal Basis (GDPR):** Explicit consent (GDPR Art. 9), contract performance

**Special Category Data Notice:** Biometric data is "special category" personal data under GDPR Article 9 and requires explicit consent. We will only process biometric data with your explicit, informed consent.

**Retention:**

- Biometric templates: Encrypted and retained while your VEID account is active, plus statutory retention periods
- Identity documents: Encrypted and retained for 7 years (KYC/AML compliance)
- Illinois users: See Biometric Data Addendum for BIPA-specific retention

### 3.3 Provider Infrastructure Data

**What We Collect from Providers:**

- Infrastructure capacity and specifications
- Resource usage metrics (CPU, memory, storage, GPU)
- Provider daemon logs and health checks
- Network performance metrics
- Service availability statistics
- Kubernetes/cloud platform configurations (anonymized)

**Legal Basis (GDPR):** Contract performance, legitimate interests (marketplace operation)

**Retention:** 3 years for billing and audit purposes

### 3.4 Tenant Usage Data

**What We Collect from Tenants:**

- Deployment manifests (anonymized)
- Resource consumption metrics
- Service usage patterns
- Billing and payment records
- Lease performance data

**Legal Basis (GDPR):** Contract performance, legitimate interests (service provision)

**Retention:** 3 years for billing and audit purposes

### 3.5 Communication Data

**What We Collect:**

- Email addresses (if provided)
- Support ticket correspondence
- Community forum posts
- Feedback and survey responses

**Legal Basis (GDPR):** Consent, legitimate interests (customer support)

**Retention:** 3 years after last interaction

### 3.6 Technical and Analytics Data

**What We Collect:**

- IP addresses
- Browser type and version
- Device information
- Operating system
- Access times and dates
- Pages viewed and interactions
- Referring URLs
- Error logs and diagnostics

**Legal Basis (GDPR):** Legitimate interests (security, performance, analytics)

**Retention:**

- Analytics: 26 months
- Security logs: 12 months
- Error logs: 90 days

## 4. How We Use Your Information

We use collected information for the following purposes:

### 4.1 Service Provision

- Operate the VirtEngine blockchain and marketplace
- Process identity verification requests
- Match marketplace orders with provider bids
- Execute and monitor leases
- Process payments and escrow settlements
- Provide customer support

### 4.2 Identity Verification and Fraud Prevention

- Verify user identities using ML-powered facial recognition
- Detect liveness and prevent spoofing attacks
- Validate identity documents
- Prevent fraud, money laundering, and terrorist financing
- Maintain marketplace trust and safety

### 4.3 Service Improvement

- Improve ML model accuracy and determinism
- Enhance marketplace matching algorithms
- Optimize network performance
- Develop new features and services
- Conduct research and analytics

### 4.4 Legal and Compliance

- Comply with KYC/AML regulations
- Respond to lawful requests from authorities
- Enforce Terms of Service
- Protect our rights and property
- Resolve disputes

### 4.5 Communications

- Send service notifications and updates
- Respond to support inquiries
- Send governance proposals (if you're a validator)
- Deliver newsletters (with consent)

## 5. Data Encryption and Security

### 5.1 Encryption Standards

**Identity Data Encryption:**

- Algorithm: X25519-XSalsa20-Poly1305 (NaCl Box)
- Key Management: Validator public keys, user-controlled decryption
- Envelope format: Recipient fingerprint, nonce, ciphertext

**Blockchain Data:**

- Sensitive data is encrypted before blockchain submission
- Transaction signing: secp256k1 ECDSA

**Data in Transit:**

- TLS 1.3 for all API communications
- gRPC with mutual TLS for validator communications

**Data at Rest:**

- AES-256 encryption for database storage
- Encrypted backups with key rotation

### 5.2 Security Measures

We implement technical and organizational measures including:

- Multi-factor authentication (MFA) for sensitive operations
- Hardware security module (HSM) support for provider keys
- Regular security audits and penetration testing
- Secure enclave support (SGX, SEV, Nitro) for sensitive computations
- Access controls and role-based permissions
- Security monitoring and incident response
- Employee background checks and confidentiality agreements

### 5.3 Data Breach Notification

In the event of a data breach affecting personal data, we will:

- Notify affected users within 72 hours (GDPR requirement)
- Notify relevant supervisory authorities as required
- Provide details of the breach, affected data, and mitigation steps
- Offer remediation measures (e.g., credit monitoring for identity theft risks)

## 6. Data Sharing and Disclosure

### 6.1 Consent-Based Sharing

You control how your VEID identity data is shared through granular consent mechanisms:

- **Scope-based consent:** Grant access to specific data scopes
- **Provider-specific consent:** Limit sharing to specific providers
- **Time-limited consent:** Set expiration dates
- **Revocable consent:** Withdraw consent at any time

### 6.2 Blockchain Data Sharing

Blockchain data is public by design and shared with:

- All blockchain node operators (validators, full nodes)
- Blockchain explorers and analytics services
- Any party with internet access (public blockchain)

**Important:** Do not submit unencrypted personal data to the blockchain.

### 6.3 Service Providers and Sub-Processors

We may share data with trusted third-party service providers:

- **Cloud infrastructure:** AWS, Azure, OpenStack (for provider services)
- **ML inference:** TensorFlow serving infrastructure
- **Payment processing:** Stripe, cryptocurrency payment gateways
- **Analytics:** Privacy-focused analytics platforms
- **Email services:** Transactional email providers
- **Security services:** DDoS protection, threat intelligence

All service providers are contractually obligated to protect your data and use it only for specified purposes.

### 6.4 Legal Disclosures

We may disclose information when required by law or to:

- Comply with legal obligations (subpoenas, court orders)
- Respond to lawful government requests
- Enforce Terms of Service
- Protect rights, property, or safety
- Prevent fraud or criminal activity

### 6.5 Business Transfers

In the event of a merger, acquisition, or sale of assets, user information may be transferred to the successor entity. You will be notified of any such transfer.

### 6.6 No Selling of Personal Data

**We do not sell your personal data.** Under CCPA, we do not sell personal information to third parties for monetary consideration.

**Biometric Data:** We do not sell, lease, or trade biometric data under any circumstances (BIPA compliance).

## 7. International Data Transfers

VirtEngine operates globally. Your data may be transferred to and processed in countries outside your jurisdiction, including:

- Australia (primary data controller location)
- United States (cloud infrastructure, service providers)
- European Union (EU users' data processing)
- Other countries where validators and providers operate

### 7.1 Transfer Mechanisms

For transfers from the EU/EEA, we rely on:

- **Standard Contractual Clauses (SCCs):** EU Commission-approved data transfer agreements
- **Adequacy Decisions:** Transfers to countries with adequate data protection (e.g., UK)
- **Explicit Consent:** For transfers with your consent

### 7.2 Blockchain Data

Blockchain data is globally distributed by nature. By using blockchain services, you acknowledge and consent to international data transfers inherent in decentralized networks.

## 8. Data Retention

### 8.1 Retention Periods

| Data Type              | Retention Period         | Rationale               |
| ---------------------- | ------------------------ | ----------------------- |
| Blockchain data        | Permanent                | Immutability            |
| Biometric templates    | Active account + 3 years | KYC/AML compliance      |
| Identity documents     | 7 years                  | Regulatory requirements |
| Transaction records    | 7 years                  | Tax, accounting, legal  |
| Usage metrics          | 3 years                  | Billing, audit          |
| Support communications | 3 years                  | Customer service        |
| Analytics data         | 26 months                | GDPR recommendation     |
| Security logs          | 12 months                | Incident investigation  |
| Error logs             | 90 days                  | Debugging               |

### 8.2 Deletion Procedures

Upon request or retention period expiration, we will:

- Securely delete personal data from active systems
- Remove data from backups according to backup rotation schedules
- Anonymize data where deletion is not feasible
- Retain data only where legally required

**Blockchain Exception:** Data recorded on the blockchain cannot be deleted due to immutability. We can only delete off-chain references and encryption keys.

## 9. Your Rights and Choices

### 9.1 Rights Under GDPR (EU/EEA Users)

You have the right to:

**Access:** Request copies of your personal data  
**Rectification:** Correct inaccurate or incomplete data  
**Erasure:** Request deletion ("right to be forgotten")  
**Restriction:** Limit how we process your data  
**Portability:** Receive data in machine-readable format  
**Object:** Object to processing based on legitimate interests  
**Withdraw Consent:** Revoke consent at any time  
**Lodge Complaint:** File complaints with supervisory authorities

**Exercise Your Rights:** Email dpo@virtengine.com or use in-app privacy controls.

### 9.2 Rights Under CCPA/CPRA (California Residents)

You have the right to:

**Know:** Request disclosure of data collection and sharing practices  
**Access:** Obtain copies of collected personal information  
**Delete:** Request deletion of personal information  
**Opt-Out:** Opt out of data "sales" (we don't sell data)  
**Non-Discrimination:** Not receive discriminatory treatment for exercising rights  
**Correct:** Request correction of inaccurate information  
**Limit Sensitive Data Use:** Limit use of sensitive personal information

**Exercise Your Rights:** Email privacy@virtengine.com or call [toll-free number].

### 9.3 Rights Under BIPA (Illinois Residents)

For biometric data, you have the right to:

- Receive written notice and consent forms before collection
- Know the purpose and retention period
- Receive written confirmation of data destruction

See the Biometric Data Addendum for full BIPA disclosures.

### 9.4 Consent Management

**VEID Consent Controls:**

- **Granular Scopes:** Control access to specific data types
- **Provider Restrictions:** Limit sharing to specific providers
- **Expiration Settings:** Set time limits on consent
- **Real-Time Revocation:** Withdraw consent instantly

Access consent controls at: [URL] or via CLI: `virtengine veid consent`

### 9.5 Blockchain Data Limitations

**Right to Erasure Limitations:** Blockchain data cannot be deleted due to technical immutability. However:

- We can delete encryption keys, rendering encrypted data unreadable
- We can remove off-chain indexes and references
- We can anonymize on-chain data where possible

## 10. Children's Privacy

The Services are not directed to children under 18 years of age. We do not knowingly collect personal information from children.

If we discover that we have collected data from a child without parental consent, we will delete such information promptly.

Parents or guardians who believe we have collected data from a child should contact privacy@virtengine.com.

## 11. Cookies and Tracking Technologies

We use cookies and similar technologies. For detailed information, see our Cookie Policy.

**Summary:**

- **Essential Cookies:** Required for service functionality
- **Analytics Cookies:** Measure usage and performance
- **Preference Cookies:** Remember your settings

You can control cookies through browser settings or our cookie consent tool.

## 12. Third-Party Services and Links

The Services may integrate with or link to third-party services (infrastructure providers, wallets, explorers). We are not responsible for third-party privacy practices.

**Third-Party Integrations:**

- Cryptocurrency wallets (Keplr, Ledger)
- Blockchain explorers
- Infrastructure providers (AWS, Azure, OpenStack)
- Identity verification services (government databases)

Review third-party privacy policies before using integrated services.

## 13. Automated Decision-Making

We use automated decision-making for:

- **Identity Verification:** ML models generate verification scores
- **Fraud Detection:** Automated risk scoring
- **Marketplace Matching:** Algorithm-based bid selection

**Your Rights:** You have the right to:

- Request human review of automated decisions
- Challenge adverse decisions
- Understand the logic behind decisions

Contact dpo@virtengine.com to exercise these rights.

## 14. Changes to This Privacy Policy

We may update this Privacy Policy to reflect changes in practices, technologies, or legal requirements.

**Notification of Changes:**

- Updated "Last Updated" date at the top of this policy
- Email notification for material changes (if email provided)
- In-app notifications
- 30-day notice period for material changes

Continued use of the Services after changes constitutes acceptance of the updated policy.

## 15. Regional Privacy Notices

### 15.1 European Union / EEA

**Legal Bases for Processing:**

- Consent (GDPR Art. 6(1)(a), Art. 9(2)(a) for biometric data)
- Contract performance (Art. 6(1)(b))
- Legal obligation (Art. 6(1)(c))
- Legitimate interests (Art. 6(1)(f))

**Supervisory Authority:**  
You may lodge complaints with your local data protection authority. EU supervisory authority list: https://edpb.europa.eu/about-edpb/board/members_en

### 15.2 United Kingdom

UK GDPR applies. UK residents have the same rights as EU residents. Contact the UK Information Commissioner's Office: https://ico.org.uk/

### 15.3 California

**CCPA Disclosure:**

- **Categories Collected:** Identifiers, biometric data, commercial information, internet activity
- **Business Purpose:** Service provision, fraud prevention, legal compliance
- **Third-Party Sharing:** Service providers (no selling)
- **Sensitive Data:** Biometric, government IDs (opt-in consent)

**Do Not Sell My Personal Information:** We do not sell personal information.

### 15.4 Nevada

Nevada residents may opt out of the "sale" of personal information. We do not sell personal information.

### 15.5 Brazil (LGPD)

Brazilian users have rights similar to GDPR, including access, correction, deletion, and portability. Contact our Data Protection Officer at dpo@virtengine.com.

### 15.6 Canada (PIPEDA)

Canadian users have rights to access and correct personal information. Contact privacy@virtengine.com.

### 15.7 Australia (Privacy Act)

As an Australian entity, we comply with the Australian Privacy Principles (APPs). Australian users may access and correct their information or lodge complaints with the Office of the Australian Information Commissioner (OAIC): https://www.oaic.gov.au/

## 16. Contact Us

For privacy questions, concerns, or to exercise your rights:

**Email:** privacy@virtengine.com  
**Data Protection Officer:** dpo@virtengine.com

**Response Time:** We will respond to privacy requests within 30 days (or as required by applicable law).

---

## Acknowledgment

By using the Services, you acknowledge that you have read and understood this Privacy Policy and consent to the collection, use, and disclosure of your information as described herein.
