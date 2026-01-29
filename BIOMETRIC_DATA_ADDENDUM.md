# Biometric Data Addendum

**Last Updated:** January 29, 2026  
**Effective Date:** January 29, 2026

This Biometric Data Addendum ("Addendum") supplements the VirtEngine [Privacy Policy](PRIVACY_POLICY.md), [Terms of Service](TERMS_OF_SERVICE.md), and [Data Processing Agreement](DATA_PROCESSING_AGREEMENT.md) with respect to the collection, use, storage, and disclosure of biometric data through the VEID (Verifiable Electronic Identity) system.

This Addendum is required by and complies with the Illinois Biometric Information Privacy Act (BIPA), 740 ILCS 14/, and other applicable biometric privacy laws.

## 1. Applicability

This Addendum applies to:
- All users of VEID identity verification services
- Users located in jurisdictions with biometric privacy laws (Illinois, Texas, Washington, California, New York, Arkansas, etc.)
- Biometric data processors and controllers

**If you are an Illinois resident**, this Addendum contains mandatory BIPA disclosures and consent requirements.

## 2. Definitions

### 2.1 Biometric Information (BIPA Definition)

"Biometric information" means any information, regardless of how it is captured, converted, stored, or shared, based on an individual's biometric identifier used to identify an individual.

### 2.2 Biometric Identifier (BIPA Definition)

"Biometric identifier" means a retina or iris scan, fingerprint, voiceprint, or scan of hand or face geometry. 

**For VirtEngine VEID:** We collect facial geometry scans and face-based biometric identifiers.

### 2.3 Exclusions

Biometric identifiers do not include:
- Writing samples
- Written signatures
- Photographs (unless analyzed to extract facial geometry)
- Human biological samples used for scientific testing
- Demographic data (age, gender, race, ethnicity)
- Tattoo descriptions
- Physical descriptions (height, weight, eye color, hair color)

**VirtEngine Clarification:** Raw photographs are not biometric identifiers. Facial recognition templates extracted from photographs are biometric identifiers.

## 3. Notice of Biometric Data Collection (BIPA § 15(b))

### 3.1 Written Notice Requirement

Before collecting biometric identifiers or biometric information, VirtEngine provides this written notice informing you that we are collecting, capturing, or otherwise obtaining your biometric identifiers or biometric information.

### 3.2 Purpose of Collection

**Primary Purpose:** Identity verification for marketplace trust and safety.

**Specific Purposes:**
- Verify your identity when creating a VEID account
- Authenticate identity during high-value transactions
- Prevent fraud, impersonation, and synthetic identity attacks
- Detect liveness (prevent spoofing with photos, videos, or masks)
- Enable providers to verify tenant identities for lease agreements
- Comply with KYC (Know Your Customer) and AML (Anti-Money Laundering) regulations
- Generate trust scores for marketplace participants

**Secondary Purposes:**
- Improve machine learning model accuracy and determinism
- Detect and prevent system abuse
- Comply with legal obligations and law enforcement requests

### 3.3 Retention Period and Destruction

**Retention Period:** Biometric identifiers and biometric information are retained:
- **While Your Account Is Active:** For identity verification and fraud prevention
- **After Account Closure:** 3 years (to prevent re-registration of fraudulent accounts)
- **Legal Holds:** Longer retention if required by law (e.g., subpoena, investigation)
- **Maximum Retention:** 7 years from last use (KYC/AML compliance)

**Automatic Destruction:** At the end of the retention period, biometric data is:
- Permanently deleted from active systems
- Removed from backups according to backup rotation schedules (typically within 90 days)
- Encryption keys destroyed, rendering encrypted data unreadable

**User-Initiated Deletion:** You may request deletion of biometric data by:
- Closing your VEID account
- Submitting a deletion request to dpo@virtengine.com
- Using in-app "Delete My Biometric Data" function

**Deletion Timeline:** 30 days from request (or account closure), plus backup rotation period.

**Blockchain Consideration:** Biometric data is never stored unencrypted on the blockchain. Deletion includes destroying encryption keys, making on-chain encrypted data permanently unreadable.

## 4. Written Consent Requirement (BIPA § 15(b))

### 4.1 Informed Consent

Before collecting your biometric identifiers or biometric information, VirtEngine obtains your **informed written consent** (or, in the context of a website, electronic consent through an affirmative opt-in mechanism).

**Consent Mechanism:**
- Checkbox or button labeled: "I consent to the collection and use of my biometric data as described in the Biometric Data Addendum"
- Link to this Addendum provided
- Consent is separate from general Terms of Service acceptance
- Consent is logged with timestamp and IP address

### 4.2 Consent Scope

By consenting, you authorize VirtEngine to:
- Collect facial geometry and biometric templates from photographs or video
- Extract facial features and generate biometric identifiers
- Store biometric identifiers in encrypted form
- Use biometric identifiers for identity verification
- Share biometric verification results (not raw biometric data) with providers (with your explicit consent per transaction)
- Use biometric data to improve ML models (in aggregate, anonymized form)

### 4.3 Voluntary Consent

**Biometric verification is optional.** You are not required to provide biometric data to use all VirtEngine services. However:
- Certain marketplace features may require biometric verification
- Providers may require verified tenants
- High-value transactions may trigger verification requirements

**Consequences of Refusal:**
- You can use non-identity-gated marketplace features
- You may face limitations on leases or bids requiring verification
- Some providers may reject unverified tenants

### 4.4 Withdrawal of Consent

**Right to Withdraw:** You may withdraw consent at any time by:
- Emailing dpo@virtengine.com with subject "Withdraw Biometric Consent"
- Using in-app "Revoke Biometric Consent" function
- Closing your VEID account

**Effect of Withdrawal:**
- VirtEngine ceases biometric processing (subject to legal retention)
- Your verification status is revoked
- Biometric data is deleted per the destruction schedule
- You can no longer use biometric verification features

## 5. Prohibitions and Limitations (BIPA § 15(c), (d))

### 5.1 No Sale, Lease, or Trade (BIPA § 15(c))

VirtEngine **does not and will not**:
- Sell biometric identifiers or biometric information
- Lease biometric identifiers or biometric information to third parties
- Trade biometric identifiers or biometric information for anything of value

**Absolute Prohibition:** This prohibition applies regardless of consent. Biometric data is never monetized.

### 5.2 No Disclosure Except as Authorized (BIPA § 15(d))

VirtEngine **does not and will not** disclose, redisclose, or otherwise disseminate biometric identifiers or biometric information unless:

**Consent:** You provide specific written consent for the disclosure

**Statutory Exceptions:**
1. **Completion of Transaction:** Disclosure is required to complete a financial transaction requested by you
2. **Service Providers:** Disclosure is to service providers bound by confidentiality agreements (e.g., ML inference servers, cloud storage providers)
3. **Law Enforcement:** Required by valid warrant, subpoena, or court order
4. **Ownership Transfer:** Business sale or merger (recipient must comply with BIPA)

**Our Practice:**
- We share **verification results** (pass/fail, trust score), not raw biometric data
- Service providers (ML servers, cloud storage) are contractually prohibited from retaining or using biometric data
- Law enforcement requests are carefully reviewed and contested when appropriate
- Biometric data is never shared with providers, tenants, or third parties for marketing

## 6. Data Security and Protection (BIPA § 15(e))

### 6.1 Standard of Care

VirtEngine stores, transmits, and protects biometric identifiers and biometric information using:

**Industry Standards:**
- Encryption standards equal to or exceeding standards for financial data
- Security standards equal to or exceeding those for other confidential and sensitive information
- Reasonable care (at least the same standard VirtEngine uses to protect other confidential information)

### 6.2 Specific Security Measures

**Encryption:**
- **At Rest:** AES-256 encryption with key rotation
- **In Transit:** TLS 1.3 for all network transmission
- **Biometric Templates:** X25519-XSalsa20-Poly1305 envelope encryption with validator public keys

**Access Controls:**
- Role-based access control (RBAC)
- Multi-factor authentication (MFA) for administrative access
- Audit logging of all biometric data access
- Principle of least privilege

**Infrastructure:**
- Secure enclaves (Intel SGX, AMD SEV, AWS Nitro) for ML inference where available
- HSM (Hardware Security Module) for key management
- Regular security audits and penetration testing
- Incident response and breach notification procedures

**Personnel:**
- Background checks for employees with biometric data access
- Confidentiality agreements and training
- Limited access to need-to-know personnel only

### 6.3 Third-Party Service Providers

Service providers with access to biometric data (e.g., cloud infrastructure) are:
- Contractually obligated to BIPA-equivalent standards
- Subject to data processing agreements (DPAs)
- Prohibited from retaining, using, or disclosing biometric data
- Regularly audited for compliance

## 7. Data Breach Notification

In the event of a data breach involving biometric identifiers or biometric information:

**Notification Timeline:**
- **Discovery:** Detection and confirmation of breach
- **Notice to Users:** Within 72 hours of discovery (faster for Illinois residents if feasible)
- **Notice to Authorities:** As required by applicable law

**Notice Contents:**
- Description of the breach
- Types of biometric data affected
- Number of affected individuals
- Steps taken to mitigate harm
- Resources for affected individuals (credit monitoring, identity theft protection)
- Contact information for questions

**Remediation:**
- Immediate steps to contain breach
- Investigation and root cause analysis
- Enhanced security measures to prevent recurrence
- Offer of identity theft protection services for affected users

## 8. User Rights

### 8.1 Access Rights

You have the right to:
- Request a copy of your biometric identifiers and biometric information
- Understand what biometric data we hold about you
- Receive information about how your biometric data is used

**Exercise Your Right:** Email dpo@virtengine.com with subject "Biometric Data Access Request"

**Response Time:** 30 days

### 8.2 Correction Rights

If your biometric data is inaccurate or outdated:
- Request correction by re-submitting verification
- Update identity documents if source data changed

### 8.3 Deletion Rights

You have the right to:
- Request deletion of biometric identifiers and biometric information
- Close your VEID account and have biometric data deleted
- Withdraw consent and cease processing

**Deletion Process:**
1. Submit deletion request to dpo@virtengine.com
2. VirtEngine confirms your identity
3. Biometric data deleted within 30 days
4. Confirmation of deletion provided

**Exceptions:** Legal holds, pending investigations, or retention required by law may delay deletion.

### 8.4 Portability Rights

Biometric data is typically not portable due to its sensitive nature and security risks. However:
- You may request verification records (dates, results, metadata)
- You may export identity documents (not biometric templates)

## 9. Illinois-Specific BIPA Provisions

### 9.1 Private Right of Action (BIPA § 20)

Illinois residents have a private right of action for BIPA violations, with statutory damages:
- **Negligent Violation:** $1,000 or actual damages (whichever is greater) per violation
- **Intentional or Reckless Violation:** $5,000 or actual damages (whichever is greater) per violation
- **Attorney's Fees and Costs:** Prevailing party may recover reasonable attorney's fees and costs

### 9.2 Waiver Prohibition (BIPA § 25)

**Any waiver of BIPA rights is void.** You cannot waive your BIPA rights through contract or agreement. If any provision of these Terms conflicts with BIPA, BIPA controls for Illinois residents.

### 9.3 Statute of Limitations

**5-Year Statute of Limitations:** BIPA claims must be brought within 5 years of the violation (or discovery of the violation for negligent violations).

**Continuing Violations:** Each scan or collection may constitute a separate violation.

### 9.4 Illinois Jurisdiction

BIPA applies to:
- Biometric data collected in Illinois
- Illinois residents, regardless of collection location
- Companies doing business in Illinois

## 10. Other State Biometric Privacy Laws

### 10.1 Texas (Tex. Bus. & Com. Code § 503.001)

Similar to BIPA but requires:
- Written consent before collection
- No sale without consent
- Reasonable care in storage and destruction
- **No private right of action** (enforcement by Attorney General only)

### 10.2 Washington (Wash. Rev. Code § 19.375)

- Consent required for biometric identifier enrollment
- No sale or lease without consent
- Notice of data breach
- **No private right of action** (enforcement by Attorney General only)

### 10.3 California (CCPA/CPRA)

- Biometric information is "sensitive personal information"
- Right to limit use and disclosure of biometric data
- Enhanced penalties for violations involving minors
- **Private right of action for data breaches** only

### 10.4 New York (Proposed)

New York has proposed but not yet enacted biometric privacy legislation. VirtEngine monitors developments and will comply when enacted.

### 10.5 Arkansas (Ark. Code Ann. § 4-110-101)

- Notice and consent required
- Prohibition on sale without consent
- Reasonable care in storage
- **No private right of action** (enforcement by Attorney General only)

## 11. International Considerations

### 11.1 GDPR (EU/EEA)

Biometric data is "special category data" under GDPR Article 9, requiring:
- Explicit consent (GDPR Art. 9(2)(a))
- Data protection impact assessment (DPIA)
- Enhanced security measures
- Right to erasure (subject to exceptions)

### 11.2 UK GDPR

Similar to GDPR with UK-specific requirements. UK ICO guidance on biometric data applies.

### 11.3 Other Jurisdictions

VirtEngine complies with biometric data laws in all jurisdictions where we operate. Contact dpo@virtengine.com for jurisdiction-specific information.

## 12. Children's Biometric Data

**Age Restriction:** The Services are not available to individuals under 18 years of age.

**Parental Consent:** If we discover biometric data from a minor, we will:
- Delete the data immediately
- Notify the individual (if contact information available)
- Investigate how the data was collected

**Zero Tolerance:** We do not collect, store, or process children's biometric data under any circumstances.

## 13. Consent Form

**REQUIRED FOR ILLINOIS RESIDENTS AND RECOMMENDED FOR ALL USERS:**

Before using VEID biometric verification, you must affirmatively consent by checking the following:

---

**BIOMETRIC INFORMATION CONSENT FORM**

I acknowledge that I have read and understood the VirtEngine Biometric Data Addendum. I understand that VirtEngine collects, uses, and stores biometric identifiers and biometric information, specifically facial geometry and facial recognition templates.

I have been informed:
- [x] Of the specific purpose for collecting, using, and storing my biometric data (identity verification, fraud prevention)
- [x] Of the length of time my biometric data will be stored (up to 7 years)
- [x] That my biometric data will be permanently destroyed at the end of the retention period
- [x] That VirtEngine will not sell, lease, or trade my biometric data
- [x] That I have the right to withdraw consent and request deletion at any time

I hereby consent to VirtEngine's collection, use, storage, and disclosure of my biometric identifiers and biometric information as described in the Biometric Data Addendum.

**Consent Date:** [Timestamp]  
**User Address/ID:** [Wallet Address]  
**IP Address:** [IP]  
**Consent ID:** [Unique Identifier]

---

## 14. Amendment

This Addendum may be updated to reflect:
- Changes in biometric privacy laws
- New biometric processing activities
- Enhanced security measures
- User feedback

**Notice of Changes:** 30 days' advance notice for material changes. Re-consent required for expanded biometric processing.

## 15. Contact Information

**Biometric Data Questions:**  
Email: dpo@virtengine.com  
Subject: "Biometric Data Inquiry"

**Data Protection Officer:**  
Email: dpo@virtengine.com

**Privacy Inquiries:**  
Email: privacy@virtengine.com

**Illinois-Specific Inquiries:**  
Email: legal@virtengine.com  
Subject: "BIPA Inquiry"

**Mailing Address:**  
DET-IO Pty. Ltd.  
Attn: Data Protection Officer / Biometric Privacy  
Australia

**Response Time:** 30 days for non-urgent requests; 72 hours for urgent privacy concerns.

---

## Acknowledgment

**By using VEID biometric verification services, you acknowledge that:**
1. You have read and understood this Biometric Data Addendum
2. You have been informed of the collection, use, retention, and disclosure of your biometric data
3. You voluntarily consent to biometric data processing
4. You understand your rights, including the right to withdraw consent and request deletion
5. You understand that biometric data will not be sold, leased, or traded

**Illinois Residents:** You acknowledge that this Addendum complies with the Illinois Biometric Information Privacy Act (BIPA) and that your consent is informed and voluntary.
