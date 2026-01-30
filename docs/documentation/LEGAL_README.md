# VirtEngine Legal Documentation

This directory contains the legal documents governing the use of VirtEngine's decentralized cloud computing marketplace, blockchain network, and identity verification services.

## Core Legal Documents

### User Agreements

1. **[Terms of Service](TERMS_OF_SERVICE.md)** - Master agreement governing all VirtEngine services
   - Service descriptions (blockchain, VEID, marketplace)
   - User obligations and responsibilities
   - Provider and tenant terms
   - Intellectual property rights
   - Limitation of liability and disclaimers
   - Dispute resolution (Australian law, arbitration)

2. **[Acceptable Use Policy](ACCEPTABLE_USE_POLICY.md)** - Prohibited activities and conduct
   - Illegal activities prohibited
   - Identity fraud prevention
   - Network security requirements
   - Content restrictions
   - Provider and tenant specific rules

### Privacy and Data Protection

3. **[Privacy Policy](PRIVACY_POLICY.md)** - How we collect, use, and protect personal data
   - Data collection practices
   - Legal bases for processing (GDPR)
   - Data sharing and disclosure
   - International data transfers
   - User rights (access, deletion, portability)
   - Multi-jurisdiction compliance (GDPR, CCPA, BIPA, etc.)

4. **[Cookie Policy](COOKIE_POLICY.md)** - Cookie usage and tracking technologies
   - Types of cookies used (essential, analytics, functionality)
   - Cookie management and consent
   - Third-party cookies
   - Browser controls

5. **[Data Processing Agreement (DPA)](DATA_PROCESSING_AGREEMENT.md)** - Controller-processor relationship
   - GDPR Article 28 compliance
   - Processing instructions and limitations
   - Security measures
   - Sub-processor management
   - Standard Contractual Clauses (SCCs) for EU transfers
   - Data subject rights assistance

6. **[Biometric Data Addendum](BIOMETRIC_DATA_ADDENDUM.md)** - Biometric data specific disclosures
   - Illinois BIPA compliance
   - Biometric data collection notice
   - Retention and destruction schedules
   - Written consent requirements
   - No sale, lease, or trade of biometric data
   - Security measures for biometric data

### Technical Implementation

7. **[Consent Framework](CONSENT_FRAMEWORK.md)** - Technical consent management system
   - Scope-based consent architecture
   - Consent lifecycle (grant, modify, revoke)
   - Provider-specific consent
   - Implementation guidelines for developers
   - CLI commands and API examples

8. **[Legal Compliance Checklist](LEGAL_COMPLIANCE_CHECKLIST.md)** - Launch readiness validation
   - Multi-jurisdiction compliance validation
   - Consent mechanism verification
   - Data protection measures
   - Third-party management
   - Incident response readiness

## Quick Start for Users

### Before Using VirtEngine

1. **Read the Terms of Service** - Understand your rights and obligations
2. **Review the Privacy Policy** - Learn how your data is handled
3. **Understand Biometric Data Collection** (VEID users) - Read the Biometric Data Addendum
4. **Review Acceptable Use Policy** - Know what activities are prohibited

### Key Points to Understand

**Blockchain Transparency:**
- Blockchain data is public and immutable
- Never submit unencrypted personal data on-chain
- Transactions cannot be reversed

**Biometric Data:**
- Collection requires explicit consent
- You can revoke consent at any time
- Data is encrypted and securely stored
- Never sold, leased, or traded

**Your Rights:**
- Access your personal data
- Correct inaccurate data
- Request deletion (subject to blockchain limitations)
- Withdraw consent
- Export your data (portability)
- Lodge complaints with data protection authorities

## Jurisdiction-Specific Information

### European Union / EEA (GDPR)
- Data Protection Officer: dpo@virtengine.com
- EU Representative: [Contact if appointed]
- Legal basis: Consent, contract performance, legitimate interests
- Lodge complaints: https://edpb.europa.eu/about-edpb/board/members_en

### United Kingdom (UK GDPR)
- UK Representative: [Contact if appointed]
- Information Commissioner's Office: https://ico.org.uk/

### California (CCPA/CPRA)
- Exercise your rights: privacy@virtengine.com or [toll-free number]
- We do not sell personal information
- Right to opt out: N/A (no selling)
- Right to limit sensitive PI use: Available upon request

### Illinois (BIPA)
- Written consent required before biometric collection
- Purpose and retention period disclosed
- Data destroyed within 7 years maximum
- Private right of action available for violations

### Australia (Privacy Act)
- Office of the Australian Information Commissioner: https://www.oaic.gov.au/
- Notifiable Data Breaches scheme applies

### Other Jurisdictions
- Canada (PIPEDA): privacy@virtengine.com
- Brazil (LGPD): dpo@virtengine.com
- Singapore (PDPA): privacy@virtengine.com

## Contact Information

**General Legal Questions:**  
legal@virtengine.com

**Privacy and Data Protection:**  
privacy@virtengine.com

**Data Protection Officer:**  
dpo@virtengine.com

**Abuse Reports:**  
abuse@virtengine.com

**Security Vulnerabilities:**  
security@virtengine.com

**Mailing Address:**  
DET-IO Pty. Ltd.  
[Registered Office Address]  
Australia

## Document Updates

All legal documents are reviewed regularly and updated as needed to reflect:
- Changes in applicable laws and regulations
- New features or services
- User feedback
- Industry best practices

**Change Notifications:**
- Updated "Last Updated" date on each document
- Email notification for material changes (if email provided)
- In-app notifications
- 30-day notice period for material changes

**Change History:**  
See individual document headers for "Last Updated" dates.

## Open Source Licensing

VirtEngine is open source software licensed under the **Apache License 2.0**.

The software license (Apache 2.0) is separate from these Terms of Service and applies to:
- Source code contributions
- Software distribution
- Derivative works

See [LICENSE](../LICENSE) for the full Apache 2.0 license text.

**Important:** Using the VirtEngine network and services requires acceptance of these Terms of Service, even when running open source software.

## Patent Notice

VirtEngine's protocol is detailed in **patent AU2024203136B2**. By using the Services, you receive a limited license to utilize the patented technology solely for accessing the Services.

## Translations

Currently, legal documents are available in English only.

**Future Translations:**  
We plan to provide translations in:
- Spanish (Español)
- Simplified Chinese (简体中文)
- Portuguese (Português)
- German (Deutsch)
- French (Français)

**Translation Disclaimer:** In case of conflict between translated versions and the English version, the English version shall prevail.

## Legal Counsel Review

All documents in this directory have been drafted with legal considerations in mind and should be reviewed by qualified legal counsel before production launch.

**Recommended Review:**
- Corporate counsel familiar with Australian law
- Privacy/data protection specialist (GDPR, CCPA, BIPA)
- Cryptocurrency/blockchain legal expert
- International trade and export control counsel (if operating globally)

## Compliance Resources

**Data Protection Authorities:**
- EU: https://edpb.europa.eu/
- UK: https://ico.org.uk/
- California: https://oag.ca.gov/privacy
- Australia: https://www.oaic.gov.au/

**Privacy Frameworks:**
- GDPR Text: https://gdpr-info.eu/
- CCPA Text: https://oag.ca.gov/privacy/ccpa
- BIPA Text: https://www.ilga.gov/legislation/ilcs/ilcs3.asp?ActID=3004

**Industry Organizations:**
- International Association of Privacy Professionals (IAPP): https://iapp.org/
- Future of Privacy Forum: https://fpf.org/
- Electronic Frontier Foundation: https://www.eff.org/

## Developer Resources

For implementing consent mechanisms and privacy controls:
- [Consent Framework](CONSENT_FRAMEWORK.md) - Technical implementation guide
- Codebase: `x/veid/types/consent.go` - Consent data structures
- CLI: `virtengine veid consent` - Consent management commands
- API: gRPC/REST endpoints for consent operations

## Feedback

We welcome feedback on our legal documents and privacy practices:
- Email: legal@virtengine.com or privacy@virtengine.com
- GitHub Issues: [Repository URL]
- Community Forum: [Forum URL]

Your feedback helps us improve transparency and user experience.

---

**Last Updated:** January 29, 2026  
**Document Owner:** Legal Team, DET-IO Pty. Ltd.
