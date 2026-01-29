# VirtEngine Consent Framework

**Version:** 1.0  
**Last Updated:** January 29, 2026

## 1. Overview

This document describes the technical implementation of the consent management system for VirtEngine's VEID (Verifiable Electronic Identity) service. It provides guidance for developers implementing consent mechanisms and users exercising consent rights.

The consent framework is based on the existing implementation in `x/veid/types/consent.go` and complies with:
- GDPR (General Data Protection Regulation)
- CCPA/CPRA (California Consumer Privacy Act / California Privacy Rights Act)
- BIPA (Biometric Information Privacy Act)
- Other applicable data protection laws

## 2. Architecture

### 2.1 Consent Granularity

VirtEngine implements **scope-based consent** with three levels of granularity:

1. **Global Settings:** Apply to all data processing
2. **Scope-Specific Consent:** Apply to specific data categories (identity scopes)
3. **Provider-Specific Consent:** Apply to specific marketplace providers

### 2.2 Core Components

**ConsentSettings Structure:**
```go
type ConsentSettings struct {
    ScopeConsents              map[string]ScopeConsent  // Per-scope consent
    ShareWithProviders         bool                     // Allow provider access
    ShareForVerification       bool                     // Allow verification requests
    AllowReVerification        bool                     // Allow re-verification
    AllowDerivedFeatureSharing bool                     // Allow feature hash sharing
    GlobalExpiresAt            *time.Time               // Global expiration
    LastUpdatedAt              time.Time                // Last modification
    ConsentVersion             uint32                   // Audit version
}
```

**ScopeConsent Structure:**
```go
type ScopeConsent struct {
    ScopeID            string      // Scope identifier
    Granted            bool        // Consent status
    GrantedAt          *time.Time  // Grant timestamp
    RevokedAt          *time.Time  // Revocation timestamp
    ExpiresAt          *time.Time  // Expiration
    Purpose            string      // Processing purpose
    GrantedToProviders []string    // Provider whitelist
    Restrictions       []string    // Additional restrictions
}
```

## 3. Identity Scopes

### 3.1 Standard Scopes

| Scope ID | Description | Data Included | Special Category (GDPR) |
|----------|-------------|---------------|-------------------------|
| `veid.biometric` | Biometric verification data | Facial templates, liveness scores | Yes (Art. 9) |
| `veid.document` | Identity document data | OCR-extracted text, document scans | No |
| `veid.basic` | Basic identity information | Name, date of birth, address | No |
| `veid.verification_history` | Verification events | Timestamps, scores, ML versions | No |
| `veid.trust_score` | Computed trust score | Score, factors, confidence | No |
| `veid.geo_location` | Location data | Country, region (not precise GPS) | No |
| `veid.device_fingerprint` | Device information | Browser, OS, device type | No |

### 3.2 Scope Dependencies

Some scopes require others:
- `veid.trust_score` requires `veid.verification_history`
- `veid.biometric` is independent (can be consented separately)
- `veid.document` is independent

### 3.3 Custom Scopes

Providers may define custom scopes for specific use cases. Custom scopes must:
- Use provider-prefixed naming: `provider.{address}.{scope}`
- Obtain explicit consent before collection
- Document purpose and data included

## 4. Consent Lifecycle

### 4.1 Initial Consent (Opt-In)

**Trigger:** First-time VEID enrollment

**Process:**
1. Display consent notice with purpose and scope details
2. User reviews data collection, retention, and sharing policies
3. User explicitly opts in (checkbox, button, or CLI command)
4. System records consent with timestamp and version
5. Consent stored on-chain (encrypted) and off-chain (indexed)

**GDPR Compliance:**
- Unbundled consent (separate from Terms of Service)
- Specific purpose stated
- Freely given (no consequences for refusal beyond feature access)
- Informed (links to Privacy Policy, Biometric Addendum)
- Unambiguous affirmative action required

**CLI Example:**
```bash
virtengine veid enroll \
  --consent-biometric=true \
  --consent-document=true \
  --consent-purpose="Marketplace identity verification" \
  --consent-expiration="2027-01-29T00:00:00Z"
```

**API Example:**
```json
{
  "enroll_request": {
    "consent": {
      "scope_consents": {
        "veid.biometric": {
          "granted": true,
          "purpose": "Identity verification for marketplace trust",
          "expires_at": "2027-01-29T00:00:00Z"
        },
        "veid.document": {
          "granted": true,
          "purpose": "KYC/AML compliance",
          "expires_at": null
        }
      },
      "share_with_providers": false,
      "share_for_verification": true
    }
  }
}
```

### 4.2 Consent Management (Updates)

**Triggers:**
- User wants to change consent settings
- Provider requests access to new scope
- Consent expiration approaching

**Operations:**
- **Grant:** Add consent for a new scope
- **Revoke:** Remove consent for existing scope
- **Modify:** Change expiration or restrictions
- **Refresh:** Extend expiration

**CLI Example:**
```bash
# Grant consent to new scope
virtengine veid consent grant \
  --scope=veid.trust_score \
  --purpose="Provider risk assessment" \
  --expires-in=1y

# Revoke consent
virtengine veid consent revoke --scope=veid.biometric

# Add provider-specific consent
virtengine veid consent grant-provider \
  --scope=veid.basic \
  --provider=virtengine1abc... \
  --expires-in=30d
```

### 4.3 Consent Verification

**Before Processing:** All data processing operations must verify consent:

```go
// Example: Before sharing identity data with provider
func (k Keeper) ShareIdentityWithProvider(
    ctx sdk.Context,
    userAddress string,
    providerAddress string,
    scopeID string,
) error {
    // Get user's consent settings
    wallet, found := k.GetWallet(ctx, userAddress)
    if !found {
        return ErrWalletNotFound
    }
    
    // Check scope consent
    consent, found := wallet.ConsentSettings.GetScopeConsent(scopeID)
    if !found || !consent.Granted {
        return ErrConsentNotGranted.Wrapf("scope %s", scopeID)
    }
    
    // Check if consent is active (not expired)
    now := ctx.BlockTime()
    if !consent.IsActive(now) {
        return ErrConsentExpired.Wrapf("scope %s expired at %s", scopeID, consent.ExpiresAt)
    }
    
    // Check provider-specific consent
    if !consent.IsGrantedToProvider(providerAddress, now) {
        return ErrProviderNotAuthorized.Wrapf("provider %s", providerAddress)
    }
    
    // Proceed with data sharing
    return k.shareIdentityData(ctx, userAddress, providerAddress, scopeID)
}
```

### 4.4 Consent Revocation

**User-Initiated:**
- User can revoke consent at any time
- Immediate effect (next block)
- Ongoing processing stops
- Data deletion scheduled (subject to retention requirements)

**Automatic Revocation:**
- Consent expiration (time-based)
- Account closure
- Legal requirement (e.g., data breach, regulatory order)

**Effect:**
- Processing ceases immediately
- Provider access removed
- Verification status may be invalidated
- Data deletion initiated (subject to legal retention)

**Blockchain Consideration:**
- On-chain consent state updated immediately
- Off-chain systems notified via events
- Encryption keys can be destroyed (rendering on-chain data unreadable)

### 4.5 Consent Audit Trail

All consent actions are logged:
- Grant timestamp and version
- Revocation timestamp
- Expiration dates
- Purpose changes
- Provider additions/removals
- User actions (who, what, when, why)

**Audit Log Storage:**
- On-chain events (ConsentGranted, ConsentRevoked)
- Off-chain indexed logs (faster queries)
- Tamper-proof audit trail

## 5. Consent Expiration

### 5.1 Expiration Policies

**Time-Based Expiration:**
- User-specified expiration dates
- Default: No expiration (indefinite consent)
- Recommended: 1-year expiration for biometric consent

**Activity-Based Expiration:**
- Inactive accounts (no verification in 2 years)
- Dormant scopes (not accessed in 1 year)

**Legal Expiration:**
- GDPR: Re-consent required after significant changes
- BIPA: Maximum 7-year retention triggers auto-expiration

### 5.2 Expiration Handling

**Before Expiration (30-day warning):**
- Notify user via email/in-app notification
- Offer consent renewal
- Explain consequences of expiration

**At Expiration:**
- Consent status set to `Granted: false`
- Processing stops
- Data retention rules apply
- User can re-consent to resume services

**After Expiration:**
- Data deletion countdown begins
- Retention period: 90 days grace period
- Permanent deletion after grace period (subject to legal retention)

## 6. Provider-Specific Consent

### 6.1 Use Cases

Providers may request access to tenant identity data for:
- KYC/AML compliance
- Risk assessment and fraud prevention
- Service level customization
- Billing and invoicing

### 6.2 Request Flow

**Step 1: Provider Requests Scope**
```bash
# Provider creates access request
virtengine market provider request-identity-access \
  --tenant=virtengine1xyz... \
  --scope=veid.basic \
  --purpose="KYC for high-value lease" \
  --duration=30d
```

**Step 2: Tenant Reviews Request**
- Notification sent to tenant
- Tenant reviews provider reputation, purpose, and requested scopes
- Tenant approves or rejects

**Step 3: Tenant Grants Consent**
```bash
# Tenant grants provider-specific consent
virtengine veid consent grant-provider \
  --request-id=req_abc123 \
  --expires-in=30d
```

**Step 4: Provider Access**
- Provider can access consented scopes for specified duration
- Access logged and auditable
- Tenant can revoke at any time

### 6.3 Provider Obligations

Providers receiving identity data must:
- Use data only for stated purpose
- Implement data protection measures
- Delete data when no longer needed
- Respond to tenant revocation immediately
- Not share data with third parties

**Provider Agreement:**
Providers sign Data Processing Agreement (DPA) accepting these obligations.

## 7. Withdrawal and Deletion

### 7.1 Consent Withdrawal

**User Rights:**
- Withdraw consent at any time
- No penalty or discrimination
- Immediate effect

**Process:**
```bash
# Withdraw all biometric consent
virtengine veid consent revoke --scope=veid.biometric

# Withdraw all consents (nuclear option)
virtengine veid consent revoke-all
```

**Effect:**
- Processing stops immediately
- Provider access revoked
- Verification status invalidated
- Data deletion scheduled

### 7.2 Data Deletion

**Right to Erasure (GDPR Art. 17):**
- User requests deletion
- VirtEngine deletes data within 30 days
- Exceptions: Legal retention, ongoing legal proceedings, public interest

**Blockchain Immutability Challenge:**
- On-chain data cannot be deleted
- **Solution:** Encrypt data before blockchain submission
- **Deletion:** Destroy encryption keys, rendering data unreadable
- **Result:** Functional erasure (data inaccessible)

**Deletion Confirmation:**
- User receives confirmation email
- Certificate of deletion available upon request
- Audit log entry (deletion timestamp, scope)

## 8. Implementation Guidelines

### 8.1 For Developers

**Consent Checks:**
- Always check consent before processing personal data
- Use `IsActive()` to verify consent hasn't expired
- Check provider-specific grants with `IsGrantedToProvider()`
- Handle consent errors gracefully (return user-friendly messages)

**Consent UI:**
- Present consent requests clearly and conspicuously
- Use plain language (avoid legal jargon)
- Provide examples of how data will be used
- Make revocation as easy as granting
- Include links to Privacy Policy and Biometric Addendum

**Consent Storage:**
- Store consent settings in wallet object
- Index consent for fast lookups
- Log consent changes to audit trail
- Version consent settings for change tracking

### 8.2 For Users

**Best Practices:**
- Review consent requests carefully
- Grant consent only to trusted providers
- Use expiration dates for sensitive scopes
- Regularly review and update consent settings
- Revoke unused consents

**Consent Dashboard:**
- View all active consents
- See which providers have access
- Check expiration dates
- Revoke or modify consents
- Download consent history

**CLI Commands:**
```bash
# View current consent settings
virtengine veid consent show

# List providers with access
virtengine veid consent providers

# Export consent history
virtengine veid consent export --format=json --output=consent-history.json
```

## 9. Compliance Checklist

### 9.1 GDPR Compliance

- [x] Consent is freely given (no coercion)
- [x] Consent is specific (per-scope granularity)
- [x] Consent is informed (links to policies)
- [x] Consent is unambiguous (affirmative action)
- [x] Consent is withdrawable (easy revocation)
- [x] Consent is documented (audit trail)
- [x] Children's consent not collected (18+ only)
- [x] Special category data (biometric) has explicit consent

### 9.2 CCPA/CPRA Compliance

- [x] Notice at collection provided
- [x] Opt-out mechanism available (revocation)
- [x] Do Not Sell honored (we don't sell data)
- [x] Sensitive personal information limits respected
- [x] Consent for minors (N/A - 18+ only)

### 9.3 BIPA Compliance (Illinois)

- [x] Written notice provided (Biometric Data Addendum)
- [x] Purpose disclosed (identity verification)
- [x] Retention period disclosed (up to 7 years)
- [x] Informed written consent obtained (explicit opt-in)
- [x] No sale, lease, or trade of biometric data
- [x] Disclosure only with consent or legal exception
- [x] Reasonable security measures implemented

## 10. Consent Events and Monitoring

### 10.1 Blockchain Events

**ConsentGranted Event:**
```protobuf
message EventConsentGranted {
  string user_address = 1;
  string scope_id = 2;
  string purpose = 3;
  google.protobuf.Timestamp granted_at = 4;
  google.protobuf.Timestamp expires_at = 5;
}
```

**ConsentRevoked Event:**
```protobuf
message EventConsentRevoked {
  string user_address = 1;
  string scope_id = 2;
  google.protobuf.Timestamp revoked_at = 3;
  string reason = 4;  // Optional
}
```

**ConsentExpired Event:**
```protobuf
message EventConsentExpired {
  string user_address = 1;
  string scope_id = 2;
  google.protobuf.Timestamp expired_at = 3;
}
```

### 10.2 Monitoring and Alerts

**User Alerts:**
- Consent expiration warnings (30, 7, 1 day before)
- New provider access requests
- Consent revocation confirmations
- Data deletion confirmations

**Admin Monitoring:**
- Consent grant/revoke rates
- Expired consents requiring cleanup
- Provider access patterns
- Anomalous consent activity (potential fraud)

## 11. Troubleshooting

### 11.1 Common Issues

**Issue: "Consent not granted" error**
- **Cause:** User hasn't consented to required scope
- **Solution:** Grant consent via CLI or UI

**Issue: "Consent expired" error**
- **Cause:** Time-based expiration reached
- **Solution:** Renew consent with new expiration date

**Issue: "Provider not authorized" error**
- **Cause:** Provider not in allowed list
- **Solution:** Grant provider-specific consent

### 11.2 Support Contacts

**Consent Questions:** dpo@virtengine.com  
**Technical Issues:** support@virtengine.com  
**Privacy Concerns:** privacy@virtengine.com

## 12. Future Enhancements

### 12.1 Planned Features

- **Consent Templates:** Pre-configured consent bundles for common use cases
- **Conditional Consent:** Consent based on conditions (e.g., "only if provider rating > 4.5")
- **Consent Delegation:** Authorize third parties to manage consent (e.g., legal guardians)
- **Consent Blockchain Explorer:** Public view of consent practices (anonymized)
- **Machine-Readable Consent:** Export consent in standardized formats (ISO 27560)

### 12.2 Standards Compliance

- **ISO/IEC 27560:** Consent record information structure
- **W3C DPV (Data Privacy Vocabulary):** Semantic interoperability
- **FIDO Alliance:** Biometric authentication standards

---

**Document Maintenance:** This framework is reviewed quarterly and updated as needed to reflect changes in technology, regulations, and best practices.

**Version History:**
- **v1.0 (2026-01-29):** Initial release
