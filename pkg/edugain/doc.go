// Package edugain provides EduGAIN federation integration for academic and research
// institution SSO (Single Sign-On) authentication in the VirtEngine platform.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
//
// EduGAIN (https://edugain.org) is a global academic authentication federation
// that interconnects identity federations from over 80 countries, enabling users
// from research and education institutions worldwide to access services using
// their home institution credentials.
//
// This package implements:
//   - SAML 2.0 Service Provider (SP) functionality with security best practices
//   - EduGAIN metadata federation parsing and validation
//   - Institution discovery service (WAYF - Where Are You From)
//   - eduPerson attribute mapping (eduPersonPrincipalName, eduPersonAffiliation, etc.)
//   - Secure session management with VEID integration
//   - Certificate validation and XML signature verification
//
// Architecture:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                       EduGAIN Service                               │
//	├─────────────────────────────────────────────────────────────────────┤
//	│  SAMLProvider       │  MetadataService    │  SessionManager        │
//	│  - AuthnRequest     │  - Federation parse │  - Session storage     │
//	│  - Response verify  │  - IdP discovery    │  - Token validation    │
//	│  - Signature check  │  - Metadata refresh │  - VEID integration    │
//	├─────────────────────────────────────────────────────────────────────┤
//	│                    Attribute Mapping                                │
//	├─────────────────────────────────────────────────────────────────────┤
//	│  eduPersonPrincipalName  │  eduPersonAffiliation  │  schacHomeOrg  │
//	│  eduPersonEntitlement    │  eduPersonScopedAffil  │  displayName   │
//	└─────────────────────────────────────────────────────────────────────┘
//
// Federation Structure:
//
//	┌─────────────────────────────────────────────────────────────────────┐
//	│                     EduGAIN Federation                              │
//	├───────────────────┬───────────────────┬─────────────────────────────┤
//	│   REFEDS (EU)     │   InCommon (US)   │   AAF (Australia)          │
//	│   ├─ Oxford       │   ├─ MIT          │   ├─ CSIRO                 │
//	│   ├─ CERN         │   ├─ Stanford     │   ├─ UniMelbourne          │
//	│   └─ ETH Zurich   │   └─ Harvard      │   └─ ANU                   │
//	└───────────────────┴───────────────────┴─────────────────────────────┘
//
// Security Considerations:
//   - All SAML responses are validated with XML signature verification
//   - Certificate pinning for known federation metadata sources
//   - Replay attack prevention via assertion ID tracking
//   - Session tokens are cryptographically signed and time-limited
//   - eduPerson attributes are hashed before VEID storage (privacy)
//   - Supports REFEDS MFA profile for step-up authentication
//
// Usage:
//
//	cfg := edugain.DefaultConfig()
//	service, err := edugain.NewService(cfg)
//	if err != nil {
//	    return err
//	}
//
//	// Refresh federation metadata
//	err = service.RefreshMetadata(ctx)
//
//	// Discover institutions
//	institutions, err := service.DiscoverInstitutions(ctx, "university")
//
//	// Generate AuthnRequest for IdP
//	request, err := service.CreateAuthnRequest(ctx, edugain.AuthnRequestParams{
//	    InstitutionID: "urn:mace:incommon:mit.edu",
//	    RelayState:    sessionID,
//	})
//
//	// Verify SAML response
//	assertion, err := service.VerifyResponse(ctx, samlResponseBase64)
//
//	// Create VEID scope from verified assertion
//	scope, err := service.CreateVEIDScope(ctx, assertion, walletAddress)
package edugain
