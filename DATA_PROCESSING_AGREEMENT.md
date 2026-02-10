# VirtEngine Data Processing Agreement (DPA)

**Last Updated:** January 29, 2026  
**Effective Date:** January 29, 2026

This Data Processing Agreement ("DPA") forms part of the VirtEngine Terms of Service and governs the processing of personal data by DET-IO Pty. Ltd. ("VirtEngine", "Data Processor", "we", "us", or "our") on behalf of users ("Data Controller", "Customer", "you").

## 1. Definitions

**Personal Data:** Any information relating to an identified or identifiable natural person as defined by applicable data protection laws (GDPR, CCPA, LGPD, etc.).

**Processing:** Any operation performed on Personal Data, including collection, storage, use, disclosure, deletion, or transmission.

**Data Subject:** The individual to whom Personal Data relates.

**Controller:** The entity that determines the purposes and means of processing Personal Data (typically the Customer).

**Processor:** The entity that processes Personal Data on behalf of the Controller (VirtEngine).

**Sub-processor:** A third party engaged by the Processor to process Personal Data.

**Supervisory Authority:** A government data protection authority (e.g., EU Data Protection Authorities, UK ICO, California Attorney General).

**Applicable Data Protection Law:** GDPR, UK GDPR, CCPA/CPRA, LGPD, PIPEDA, PDPA, and other applicable privacy laws.

## 2. Scope and Applicability

### 2.1 Application

This DPA applies when:
- You use VirtEngine as a **Provider** and process tenant data
- You use VirtEngine as a **Tenant** and we process your deployment data
- You use **VEID services** and we process identity verification data

### 2.2 Data Controller and Processor Roles

**VirtEngine as Data Processor:**
- When processing Provider infrastructure data on behalf of Providers
- When processing Tenant deployment data on behalf of Tenants
- When processing VEID verification data on behalf of Users

**VirtEngine as Data Controller:**
- For VirtEngine account and authentication data
- For blockchain transaction data
- For service analytics and improvement
- For legal compliance (KYC/AML)

**Note:** This DPA applies only when VirtEngine acts as a Data Processor.

## 3. Data Processing Terms

### 3.1 Processing Instructions

**Documented Instructions:** VirtEngine processes Personal Data only:
- As documented in this DPA
- As instructed through the Services (API calls, CLI commands, dashboard)
- As required by applicable law
- As set forth in the Terms of Service and Privacy Policy

**Prohibited Processing:** VirtEngine will not:
- Process Personal Data for purposes other than providing the Services
- Sell or rent Personal Data
- Use Personal Data for marketing without explicit consent
- Share Personal Data except as authorized by this DPA

**Legal Obligations:** If VirtEngine is required by law to process data differently, we will:
- Inform the Customer before processing (unless legally prohibited)
- Document the legal requirement
- Comply with the minimum legal obligation

### 3.2 Nature and Purpose of Processing

| Service | Purpose of Processing | Data Categories |
|---------|----------------------|-----------------|
| VEID Verification | Identity verification, fraud prevention | Biometric data, identity documents, verification metadata |
| Marketplace | Order matching, lease execution, payment processing | Transaction data, resource specifications, usage metrics |
| Provider Services | Infrastructure provisioning, deployment management | Deployment manifests, resource usage, logs |
| Blockchain | Transaction validation, consensus, state management | Addresses, transaction data, governance votes |

### 3.3 Duration of Processing

Processing continues for:
- **Active Services:** Duration of service usage
- **Retention Period:** As specified in the Privacy Policy (typically 3-7 years)
- **Legal Obligations:** As required by law (KYC/AML: 7 years)
- **Blockchain Data:** Permanent (immutability)

**Deletion:** Upon request or retention expiration, VirtEngine will delete or return Personal Data (subject to blockchain immutability constraints).

### 3.4 Categories of Data Subjects

- VirtEngine users (Tenants, Providers, Validators)
- Identity verification subjects (VEID users)
- Tenant application end-users (for Provider processing)
- Support ticket contacts

## 4. Customer Obligations as Data Controller

As Data Controller, you agree to:

- **Legal Basis:** Ensure you have a lawful basis for processing (consent, contract, legitimate interest, legal obligation)
- **Data Subject Rights:** Respond to data subject requests (access, rectification, erasure, objection)
- **Consent Management:** Obtain necessary consents for data processing
- **Privacy Notices:** Provide adequate privacy notices to data subjects
- **Data Minimization:** Only provide data necessary for the Services
- **Accuracy:** Ensure data accuracy and keep data updated
- **Security:** Implement appropriate security measures for data you control
- **Lawful Instructions:** Only provide lawful processing instructions
- **Third-Party Rights:** Respect third-party intellectual property and privacy rights

## 5. VirtEngine Obligations as Data Processor

### 5.1 Confidentiality

VirtEngine will:
- Treat Personal Data as confidential
- Ensure personnel with access to Personal Data are bound by confidentiality obligations
- Limit access to personnel with a need to know
- Implement access controls and authentication

### 5.2 Security Measures (GDPR Article 32)

VirtEngine implements technical and organizational measures including:

**Technical Measures:**
- **Encryption:** X25519-XSalsa20-Poly1305 for identity data, TLS 1.3 for transit, AES-256 at rest
- **Access Controls:** Role-based access control (RBAC), multi-factor authentication (MFA)
- **Pseudonymization:** Where feasible, data is pseudonymized
- **Secure Enclaves:** SGX, SEV, Nitro support for sensitive computations
- **Key Management:** HSM support, secure key rotation
- **Network Security:** Firewalls, intrusion detection, DDoS protection

**Organizational Measures:**
- Security policies and procedures
- Regular security training for personnel
- Background checks for employees with data access
- Incident response plan
- Business continuity and disaster recovery plans
- Regular security audits and penetration testing
- Vendor security assessments

**Blockchain-Specific:**
- Encryption before blockchain submission
- No unencrypted personal data on-chain
- Secure validator key management

### 5.3 Assistance with Data Subject Rights

VirtEngine will assist the Customer in responding to Data Subject requests:

**Access Requests:** Provide access to Personal Data held by VirtEngine  
**Rectification:** Correct inaccurate data  
**Erasure:** Delete data where technically feasible (blockchain limitations apply)  
**Restriction:** Limit processing as requested  
**Portability:** Provide data in machine-readable format  
**Objection:** Cease processing where applicable

**Process:**
1. Customer receives data subject request
2. Customer forwards request to dpo@virtengine.com
3. VirtEngine responds within 10 business days
4. Customer responds to data subject within legal timeframes

**Fees:** No fees for reasonable assistance. Excessive or repetitive requests may incur fees.

### 5.4 Data Breach Notification (GDPR Article 33)

**Incident Response:**
- VirtEngine maintains an incident response plan
- Security incidents are logged and investigated
- Affected Customers are notified promptly

**Notification Timeline:**
- **Discovery:** Detect and confirm breach
- **Assessment:** Determine scope and impact
- **Notification:** Notify affected Customers within 72 hours of discovery
- **Supervisory Authority:** Customer notifies authority if required by law

**Breach Notice Includes:**
- Nature of the breach
- Categories and approximate number of data subjects affected
- Likely consequences
- Measures taken or proposed to address the breach
- Contact point for more information

**Customer Obligations:** Customer notifies supervisory authorities and data subjects as required by applicable law.

### 5.5 Data Protection Impact Assessment (DPIA)

VirtEngine will:
- Assist Customers in conducting DPIAs when required
- Provide information about processing activities
- Cooperate in DPIA processes
- Inform Customers of high-risk processing

**High-Risk Processing:** Biometric verification, automated decision-making, large-scale special category data processing.

### 5.6 Audits and Inspections

**Customer Audit Rights:**
- Customers may audit VirtEngine's compliance with this DPA
- Audits must be reasonable, scheduled in advance, and not disrupt operations
- Frequency: Maximum once per year (unless breach or regulatory requirement)
- Scope: Limited to Customer's data and relevant processing activities

**Audit Alternatives:**
- VirtEngine may provide existing audit reports (SOC 2, ISO 27001)
- Third-party auditor reports
- Completed security questionnaires

**Cost:** Customer bears audit costs. VirtEngine may charge reasonable fees for extensive audits.

## 6. Sub-Processors

### 6.1 Authorization

Customer authorizes VirtEngine to engage sub-processors for processing Personal Data.

**Current Sub-Processors:**

| Sub-Processor | Service | Location | Data Processed |
|---------------|---------|----------|----------------|
| AWS | Cloud infrastructure | Global (customer-selected regions) | Encrypted backups, infrastructure |
| Azure | Cloud infrastructure | Global (customer-selected regions) | Provider workloads |
| OpenStack | Cloud infrastructure | Provider-dependent | Provider workloads |
| Stripe | Payment processing | US, EU | Payment information |
| [Email Provider] | Transactional emails | US, EU | Email addresses, communication |
| [Analytics Provider] | Usage analytics | EU | Anonymized usage data |

**Updated List:** See https://virtengine.com/legal/sub-processors for current list.

### 6.2 Sub-Processor Requirements

VirtEngine ensures sub-processors:
- Are bound by data protection obligations equivalent to this DPA
- Implement appropriate security measures
- Are subject to written contracts
- Are assessed for security and compliance

### 6.3 Notice of Changes

**New Sub-Processors:** VirtEngine will notify Customers at least 30 days before engaging new sub-processors or changing existing ones.

**Notification Method:**
- Email to Customer's registered address
- Notice on website: https://virtengine.com/legal/sub-processors
- In-app notification

**Objection Right:** Customer may object to a new sub-processor within 30 days of notice. If VirtEngine cannot accommodate the objection, Customer may terminate the affected Services.

## 7. International Data Transfers

### 7.1 Transfer Mechanisms

For transfers from the EU/EEA to third countries:

**Standard Contractual Clauses (SCCs):** VirtEngine adopts the EU Commission SCCs (Module 2: Controller-to-Processor or Module 3: Processor-to-Processor).

**Adequacy Decisions:** Transfers to countries with EU adequacy decisions (e.g., UK, Switzerland) rely on adequacy.

**Supplementary Measures:** VirtEngine implements supplementary security measures:
- End-to-end encryption
- Pseudonymization where feasible
- Access controls and authentication
- Contractual restrictions on government access

### 7.2 UK GDPR

For UK data transfers, VirtEngine complies with UK GDPR and UK International Data Transfer Addendum to SCCs.

### 7.3 Other Jurisdictions

**LGPD (Brazil):** International transfer clauses comply with LGPD requirements.  
**PIPEDA (Canada):** Transfers comply with Canadian adequacy and contractual requirements.  
**APEC CBPR:** VirtEngine may seek APEC Cross-Border Privacy Rules certification.

### 7.4 Customer Instructions on Transfers

Customer instructs VirtEngine to transfer Personal Data internationally as necessary to provide the Services, subject to the safeguards in this DPA.

## 8. Data Retention and Deletion

### 8.1 Retention

VirtEngine retains Personal Data for:
- **Active Use:** Duration of service usage
- **Legal Retention:** As required by law (typically 7 years for financial data)
- **Retention Policy:** As specified in Privacy Policy

### 8.2 Data Return

Upon termination or expiration of Services:
- **Customer Request:** Customer may request data export within 30 days
- **Format:** Machine-readable format (JSON, CSV)
- **Scope:** All data VirtEngine holds on behalf of Customer

### 8.3 Data Deletion

After data return or 30 days post-termination:
- VirtEngine securely deletes Customer Personal Data
- Deletion includes active systems and backups (subject to backup rotation)
- Certificate of deletion provided upon request

**Blockchain Exception:** Data on the blockchain cannot be deleted. VirtEngine will delete encryption keys rendering data unreadable.

**Legal Retention Exception:** Data required by law is retained for the legally mandated period.

## 9. Liability and Indemnification

### 9.1 Liability Allocation

**GDPR Compliance:** Under GDPR Article 82:
- Controller and Processor are jointly and severally liable for damages
- Each party is liable only for its own GDPR violations
- Customer is liable for unlawful processing instructions

**Damages:**
- VirtEngine is liable only for damages resulting from VirtEngine's GDPR violations
- Customer is liable for damages resulting from unlawful processing instructions or Customer's GDPR violations

### 9.2 Limitation of Liability

VirtEngine's total liability under this DPA is subject to the limitation of liability in the Terms of Service.

**Exception:** Liability limits do not apply to:
- GDPR fines imposed on Customer due to VirtEngine's breach
- Gross negligence or willful misconduct

### 9.3 Indemnification

**By VirtEngine:** VirtEngine indemnifies Customer for:
- Third-party claims arising from VirtEngine's GDPR violations
- Regulatory fines resulting from VirtEngine's breach of this DPA

**By Customer:** Customer indemnifies VirtEngine for:
- Third-party claims arising from Customer's unlawful processing instructions
- Customer's GDPR violations
- Customer's breach of this DPA

## 10. Term and Termination

### 10.1 Duration

This DPA remains in effect while VirtEngine processes Personal Data on behalf of Customer.

### 10.2 Survival

The following provisions survive termination:
- Data deletion obligations
- Confidentiality
- Liability and indemnification
- Audit rights (for 1 year post-termination)

### 10.3 Effect of Termination

Upon termination:
- VirtEngine ceases processing Personal Data
- Customer may request data return
- VirtEngine deletes Personal Data as specified in Section 8

## 11. Governing Law and Jurisdiction

### 11.1 Governing Law

This DPA is governed by:
- **General:** The laws of the Commonwealth of Australia
- **GDPR Clauses:** The laws of the EU member state where the Customer is established (for EU Customers)

### 11.2 Jurisdiction

Disputes are resolved according to the dispute resolution provisions in the Terms of Service.

**GDPR Exception:** EU/EEA data subjects and supervisory authorities may bring claims in EU member state courts as provided by GDPR.

## 12. Standard Contractual Clauses

### 12.1 Incorporation

For EU/EEA Customers, the EU Standard Contractual Clauses (SCCs) for international data transfers are incorporated by reference and attached as Annex A.

**Modules:**
- **Module Two:** Controller-to-Processor (when Customer is Controller, VirtEngine is Processor)
- **Module Three:** Processor-to-Processor (for sub-processor relationships)

### 12.2 Hierarchy

In case of conflict:
1. SCCs take precedence over conflicting DPA terms (for EU/EEA transfers)
2. This DPA supplements but does not contradict SCCs
3. Terms of Service apply to non-data protection matters

## 13. Amendment

This DPA may be amended:
- To reflect changes in data protection laws
- To update sub-processor lists
- To improve security measures
- To clarify provisions

**Notice:** Material amendments require 30 days' advance notice. Continued use constitutes acceptance.

## 14. Contact Information

**Data Protection Officer:**  
Email: dpo@virtengine.com

**DPA Inquiries:**  
Email: legal@virtengine.com

**EU Representative** (if applicable):  
[EU Representative Contact]

**Mailing Address:**  
DET-IO Pty. Ltd.  
Attn: Data Protection Officer  
Australia

---

## Annexes

### Annex A: Standard Contractual Clauses (SCCs)

[EU Commission SCCs for Controller-to-Processor data transfers - Module Two]

### Annex B: Sub-Processor List

See: https://virtengine.com/legal/sub-processors

### Annex C: Security Measures

See Section 5.2 and Technical Security Documentation at: https://virtengine.com/security

### Annex D: Data Processing Details

**Categories of Data:**
- Identity verification data (VEID)
- Transaction data (blockchain)
- Deployment data (marketplace)
- Support data (communications)

**Special Category Data (GDPR Article 9):**
- Biometric data (facial recognition, liveness detection)

**Processing Operations:**
- Collection, storage, retrieval, transmission, encryption, deletion, analysis

**Data Subject Categories:**
- VirtEngine users (tenants, providers, validators)
- Identity verification subjects
- End users (for provider processing)

---

**Effective Date:** This DPA is effective as of the date you accept the Terms of Service or begin using VirtEngine services as a Data Controller.

**Acknowledgment:** By using the Services, you acknowledge and agree to this DPA.
