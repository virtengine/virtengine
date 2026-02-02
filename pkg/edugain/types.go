// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"errors"
	"time"
)

// ============================================================================
// Errors
// ============================================================================

var (
	// ErrServiceNotInitialized is returned when the service is not initialized
	ErrServiceNotInitialized = errors.New("edugain service not initialized")

	// ErrMetadataNotLoaded is returned when federation metadata is not loaded
	ErrMetadataNotLoaded = errors.New("federation metadata not loaded")

	// ErrMetadataExpired is returned when federation metadata has expired
	ErrMetadataExpired = errors.New("federation metadata has expired")

	// ErrMetadataParseError is returned when metadata cannot be parsed
	ErrMetadataParseError = errors.New("failed to parse federation metadata")

	// ErrInvalidMetadataSignature is returned when metadata signature is invalid
	ErrInvalidMetadataSignature = errors.New("invalid federation metadata signature")

	// ErrInstitutionNotFound is returned when institution is not found
	ErrInstitutionNotFound = errors.New("institution not found in federation")

	// ErrInvalidSAMLResponse is returned when SAML response is invalid
	ErrInvalidSAMLResponse = errors.New("invalid SAML response")

	// ErrSAMLSignatureInvalid is returned when SAML signature verification fails
	ErrSAMLSignatureInvalid = errors.New("SAML signature verification failed")

	// ErrAssertionExpired is returned when SAML assertion has expired
	ErrAssertionExpired = errors.New("SAML assertion has expired")

	// ErrAssertionNotYetValid is returned when assertion is not yet valid
	ErrAssertionNotYetValid = errors.New("SAML assertion not yet valid")

	// ErrAudienceRestriction is returned when audience restriction fails
	ErrAudienceRestriction = errors.New("SAML audience restriction failed")

	// ErrReplayDetected is returned when assertion replay is detected
	ErrReplayDetected = errors.New("SAML assertion replay detected")

	// ErrMissingRequiredAttribute is returned when required attribute is missing
	ErrMissingRequiredAttribute = errors.New("required eduPerson attribute missing")

	// ErrInvalidSessionToken is returned when session token is invalid
	ErrInvalidSessionToken = errors.New("invalid session token")

	// ErrSessionExpired is returned when session has expired
	ErrSessionExpired = errors.New("session has expired")

	// ErrSessionNotFound is returned when session is not found
	ErrSessionNotFound = errors.New("session not found")

	// ErrCertificateExpired is returned when IdP certificate has expired
	ErrCertificateExpired = errors.New("identity provider certificate expired")

	// ErrCertificateNotTrusted is returned when certificate is not in trust chain
	ErrCertificateNotTrusted = errors.New("certificate not in trusted chain")

	// ErrUnsupportedBinding is returned when SAML binding is not supported
	ErrUnsupportedBinding = errors.New("unsupported SAML binding")

	// ErrMFARequired is returned when MFA is required but not satisfied
	ErrMFARequired = errors.New("multi-factor authentication required")

	// ErrVEIDIntegrationFailed is returned when VEID integration fails
	ErrVEIDIntegrationFailed = errors.New("VEID integration failed")

	// ErrWeakSignatureAlgorithm is returned when a weak signature algorithm (e.g., SHA-1) is used
	ErrWeakSignatureAlgorithm = errors.New("weak signature algorithm detected: SHA-1 is not allowed")

	// ErrWeakDigestAlgorithm is returned when a weak digest algorithm (e.g., SHA-1) is used
	ErrWeakDigestAlgorithm = errors.New("weak digest algorithm detected: SHA-1 is not allowed")
)

// ============================================================================
// Constants
// ============================================================================

// Version constants
const (
	// EduGAINVersion is the current version of the EduGAIN integration
	EduGAINVersion = "1.0.0"

	// SAMLVersion is the supported SAML version
	SAMLVersion = "2.0"
)

// Default configuration values
const (
	// DefaultMetadataRefreshInterval is the default metadata refresh interval
	DefaultMetadataRefreshInterval = 6 * time.Hour

	// DefaultSessionDuration is the default session duration
	DefaultSessionDuration = 8 * time.Hour

	// DefaultAssertionMaxAge is the maximum age of a SAML assertion
	DefaultAssertionMaxAge = 5 * time.Minute

	// DefaultClockSkew is the allowed clock skew for time validation
	DefaultClockSkew = 2 * time.Minute

	// DefaultReplayWindowDuration is the duration to track assertion IDs for replay
	DefaultReplayWindowDuration = 24 * time.Hour
)

// EduGAIN metadata URLs
const (
	// EduGAINProductionMetadataURL is the production federation metadata URL
	EduGAINProductionMetadataURL = "https://mds.edugain.org/edugain-v2.xml"

	// EduGAINTestMetadataURL is the test federation metadata URL
	EduGAINTestMetadataURL = "https://mds-test.edugain.org/edugain-v2.xml"
)

// SAML binding constants
const (
	// SAMLBindingHTTPRedirect is the HTTP-Redirect binding
	SAMLBindingHTTPRedirect = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Redirect"

	// SAMLBindingHTTPPOST is the HTTP-POST binding
	SAMLBindingHTTPPOST = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"

	// SAMLBindingHTTPArtifact is the HTTP-Artifact binding
	SAMLBindingHTTPArtifact = "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-Artifact"
)

// SAML NameID format constants
const (
	// NameIDFormatPersistent is the persistent NameID format
	NameIDFormatPersistent = "urn:oasis:names:tc:SAML:2.0:nameid-format:persistent"

	// NameIDFormatTransient is the transient NameID format
	NameIDFormatTransient = "urn:oasis:names:tc:SAML:2.0:nameid-format:transient"

	// NameIDFormatEmailAddress is the email address NameID format
	NameIDFormatEmailAddress = "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"

	// NameIDFormatUnspecified is the unspecified NameID format
	NameIDFormatUnspecified = "urn:oasis:names:tc:SAML:1.1:nameid-format:unspecified"
)

// SAML AuthnContext constants
const (
	// AuthnContextPasswordProtected is the password protected transport context
	//nolint:gosec // G101: This is a SAML URN constant, not a credential
	AuthnContextPasswordProtected = "urn:oasis:names:tc:SAML:2.0:ac:classes:PasswordProtectedTransport"

	// AuthnContextMFA is the REFEDS MFA profile context
	AuthnContextMFA = "https://refeds.org/profile/mfa"

	// AuthnContextSFA is the REFEDS SFA (single factor) profile context
	AuthnContextSFA = "https://refeds.org/profile/sfa"
)

// ============================================================================
// Enums
// ============================================================================

// FederationStatus represents the status of federation metadata
type FederationStatus string

const (
	// FederationStatusUnknown indicates unknown status
	FederationStatusUnknown FederationStatus = "unknown"

	// FederationStatusLoading indicates metadata is being loaded
	FederationStatusLoading FederationStatus = "loading"

	// FederationStatusActive indicates metadata is active and valid
	FederationStatusActive FederationStatus = "active"

	// FederationStatusExpired indicates metadata has expired
	FederationStatusExpired FederationStatus = "expired"

	// FederationStatusError indicates metadata load error
	FederationStatusError FederationStatus = "error"
)

// AllFederationStatuses returns all valid federation statuses
func AllFederationStatuses() []FederationStatus {
	return []FederationStatus{
		FederationStatusUnknown,
		FederationStatusLoading,
		FederationStatusActive,
		FederationStatusExpired,
		FederationStatusError,
	}
}

// IsValidFederationStatus checks if a status is valid
func IsValidFederationStatus(s FederationStatus) bool {
	for _, valid := range AllFederationStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// InstitutionType represents the type of institution
type InstitutionType string

const (
	// InstitutionTypeUniversity is a university
	InstitutionTypeUniversity InstitutionType = "university"

	// InstitutionTypeResearchInstitute is a research institute
	InstitutionTypeResearchInstitute InstitutionType = "research_institute"

	// InstitutionTypeSchool is a school (K-12)
	InstitutionTypeSchool InstitutionType = "school"

	// InstitutionTypeLibrary is a library
	InstitutionTypeLibrary InstitutionType = "library"

	// InstitutionTypeGovernment is a government organization
	InstitutionTypeGovernment InstitutionType = "government"

	// InstitutionTypeOther is any other institution type
	InstitutionTypeOther InstitutionType = "other"
)

// AllInstitutionTypes returns all valid institution types
func AllInstitutionTypes() []InstitutionType {
	return []InstitutionType{
		InstitutionTypeUniversity,
		InstitutionTypeResearchInstitute,
		InstitutionTypeSchool,
		InstitutionTypeLibrary,
		InstitutionTypeGovernment,
		InstitutionTypeOther,
	}
}

// IsValidInstitutionType checks if an institution type is valid
func IsValidInstitutionType(t InstitutionType) bool {
	for _, valid := range AllInstitutionTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// SessionStatus represents the status of an EduGAIN session
type SessionStatus string

const (
	// SessionStatusActive indicates an active session
	SessionStatusActive SessionStatus = "active"

	// SessionStatusExpired indicates an expired session
	SessionStatusExpired SessionStatus = "expired"

	// SessionStatusRevoked indicates a revoked session
	SessionStatusRevoked SessionStatus = "revoked"

	// SessionStatusPending indicates a pending session (awaiting IdP response)
	SessionStatusPending SessionStatus = "pending"
)

// AllSessionStatuses returns all valid session statuses
func AllSessionStatuses() []SessionStatus {
	return []SessionStatus{
		SessionStatusActive,
		SessionStatusExpired,
		SessionStatusRevoked,
		SessionStatusPending,
	}
}

// IsValidSessionStatus checks if a session status is valid
func IsValidSessionStatus(s SessionStatus) bool {
	for _, valid := range AllSessionStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// AffiliationType represents eduPersonAffiliation values
type AffiliationType string

const (
	// AffiliationStudent is a student
	AffiliationStudent AffiliationType = "student"

	// AffiliationFaculty is faculty/academic staff
	AffiliationFaculty AffiliationType = "faculty"

	// AffiliationStaff is non-academic staff
	AffiliationStaff AffiliationType = "staff"

	// AffiliationEmployee is an employee
	AffiliationEmployee AffiliationType = "employee"

	// AffiliationMember is a member
	AffiliationMember AffiliationType = "member"

	// AffiliationAffiliate is an affiliate
	AffiliationAffiliate AffiliationType = "affiliate"

	// AffiliationAlum is an alumnus
	AffiliationAlum AffiliationType = "alum"

	// AffiliationLibraryWalkIn is a library walk-in user
	AffiliationLibraryWalkIn AffiliationType = "library-walk-in"
)

// AllAffiliationTypes returns all valid affiliation types
func AllAffiliationTypes() []AffiliationType {
	return []AffiliationType{
		AffiliationStudent,
		AffiliationFaculty,
		AffiliationStaff,
		AffiliationEmployee,
		AffiliationMember,
		AffiliationAffiliate,
		AffiliationAlum,
		AffiliationLibraryWalkIn,
	}
}

// IsValidAffiliationType checks if an affiliation type is valid
func IsValidAffiliationType(t AffiliationType) bool {
	for _, valid := range AllAffiliationTypes() {
		if t == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Core Types
// ============================================================================

// Institution represents an identity provider (IdP) in the EduGAIN federation
type Institution struct {
	// EntityID is the unique SAML entity identifier
	EntityID string `json:"entity_id"`

	// DisplayName is the human-readable institution name
	DisplayName string `json:"display_name"`

	// Description is an optional description
	Description string `json:"description,omitempty"`

	// Type is the institution type
	Type InstitutionType `json:"type"`

	// Country is the ISO 3166-1 alpha-2 country code
	Country string `json:"country"`

	// Federation is the national/regional federation name
	Federation string `json:"federation"`

	// LogoURL is the institution logo URL
	LogoURL string `json:"logo_url,omitempty"`

	// InformationURL is the institution information page URL
	InformationURL string `json:"information_url,omitempty"`

	// PrivacyStatementURL is the privacy statement URL
	PrivacyStatementURL string `json:"privacy_statement_url,omitempty"`

	// SupportedBindings lists supported SAML bindings
	SupportedBindings []string `json:"supported_bindings"`

	// SSOEndpoints are the SSO endpoint URLs by binding type
	SSOEndpoints map[string]string `json:"sso_endpoints"`

	// SLOEndpoints are the Single Logout endpoint URLs by binding type
	SLOEndpoints map[string]string `json:"slo_endpoints,omitempty"`

	// Certificates are the IdP signing certificates (PEM encoded)
	Certificates []string `json:"certificates"`

	// NameIDFormats are the supported NameID formats
	NameIDFormats []string `json:"name_id_formats"`

	// Attributes lists the attributes the IdP can provide
	Attributes []string `json:"attributes,omitempty"`

	// SupportsMFA indicates if the IdP supports REFEDS MFA profile
	SupportsMFA bool `json:"supports_mfa"`

	// SupportsECP indicates if the IdP supports Enhanced Client/Proxy
	SupportsECP bool `json:"supports_ecp"`

	// LastUpdated is when the metadata was last updated
	LastUpdated time.Time `json:"last_updated"`

	// MetadataValidUntil is when the metadata expires
	MetadataValidUntil time.Time `json:"metadata_valid_until"`
}

// Validate validates the institution data
func (i *Institution) Validate() error {
	if i.EntityID == "" {
		return errors.New("entity_id is required")
	}
	if i.DisplayName == "" {
		return errors.New("display_name is required")
	}
	if len(i.Certificates) == 0 {
		return errors.New("at least one certificate is required")
	}
	if len(i.SSOEndpoints) == 0 {
		return errors.New("at least one SSO endpoint is required")
	}
	return nil
}

// GetSSOEndpoint returns the SSO endpoint for the preferred binding
func (i *Institution) GetSSOEndpoint(preferredBinding string) (string, error) {
	// Try preferred binding first
	if endpoint, ok := i.SSOEndpoints[preferredBinding]; ok {
		return endpoint, nil
	}
	// Fall back to HTTP-Redirect
	if endpoint, ok := i.SSOEndpoints[SAMLBindingHTTPRedirect]; ok {
		return endpoint, nil
	}
	// Fall back to HTTP-POST
	if endpoint, ok := i.SSOEndpoints[SAMLBindingHTTPPOST]; ok {
		return endpoint, nil
	}
	return "", ErrUnsupportedBinding
}

// FederationMetadata represents parsed EduGAIN federation metadata
type FederationMetadata struct {
	// FederationID is the federation entity ID
	FederationID string `json:"federation_id"`

	// Name is the federation name
	Name string `json:"name"`

	// ValidUntil is when the metadata expires
	ValidUntil time.Time `json:"valid_until"`

	// CacheDuration is the recommended cache duration
	CacheDuration time.Duration `json:"cache_duration"`

	// Institutions is the list of identity providers
	Institutions []Institution `json:"institutions"`

	// TotalEntities is the total entity count
	TotalEntities int `json:"total_entities"`

	// LastRefresh is when the metadata was last refreshed
	LastRefresh time.Time `json:"last_refresh"`

	// Status is the current federation status
	Status FederationStatus `json:"status"`

	// SignatureValid indicates if metadata signature was verified
	SignatureValid bool `json:"signature_valid"`

	// SignerCertificateHash is the hash of the signing certificate
	SignerCertificateHash string `json:"signer_certificate_hash,omitempty"`
}

// IsValid checks if the metadata is currently valid
func (m *FederationMetadata) IsValid() bool {
	return m.Status == FederationStatusActive && time.Now().Before(m.ValidUntil)
}

// FindInstitution finds an institution by entity ID
func (m *FederationMetadata) FindInstitution(entityID string) (*Institution, error) {
	for i := range m.Institutions {
		if m.Institutions[i].EntityID == entityID {
			return &m.Institutions[i], nil
		}
	}
	return nil, ErrInstitutionNotFound
}

// EduPersonAttributes represents the eduPerson schema attributes
// See: https://wiki.refeds.org/display/STAN/eduPerson
type EduPersonAttributes struct {
	// PrincipalName is the eduPersonPrincipalName (e.g., "user@institution.edu")
	// This is the primary identifier, scoped to the institution
	PrincipalName string `json:"principal_name"`

	// PrincipalNameHash is SHA-256 hash of the principal name for privacy
	PrincipalNameHash string `json:"principal_name_hash"`

	// Affiliation lists the eduPersonAffiliation values (student, faculty, staff, etc.)
	Affiliation []AffiliationType `json:"affiliation"`

	// ScopedAffiliation lists the eduPersonScopedAffiliation values
	// (e.g., "student@institution.edu")
	ScopedAffiliation []string `json:"scoped_affiliation,omitempty"`

	// Entitlement lists the eduPersonEntitlement URIs (access rights)
	Entitlement []string `json:"entitlement,omitempty"`

	// Assurance lists the eduPersonAssurance values (identity assurance level)
	Assurance []string `json:"assurance,omitempty"`

	// TargetedID is the eduPersonTargetedID (opaque, pairwise identifier)
	TargetedID string `json:"targeted_id,omitempty"`

	// TargetedIDHash is SHA-256 hash of the targeted ID
	TargetedIDHash string `json:"targeted_id_hash,omitempty"`

	// UniqueID is the eduPersonUniqueId (globally unique, persistent)
	UniqueID string `json:"unique_id,omitempty"`

	// OrcidID is the eduPersonOrcid (ORCID researcher identifier)
	OrcidID string `json:"orcid_id,omitempty"`

	// EPPN is eduPersonPrincipalName when scope validated
	EPPN string `json:"eppn,omitempty"`

	// EPPNHash is SHA-256 hash of EPPN
	EPPNHash string `json:"eppn_hash,omitempty"`
}

// SchacAttributes represents SCHAC (Schema for Academia) attributes
// See: https://wiki.refeds.org/display/STAN/SCHAC
type SchacAttributes struct {
	// HomeOrganization is the schacHomeOrganization (e.g., "mit.edu")
	HomeOrganization string `json:"home_organization"`

	// HomeOrganizationType is the schacHomeOrganizationType
	HomeOrganizationType string `json:"home_organization_type,omitempty"`

	// PersonalUniqueCode is the schacPersonalUniqueCode (student ID, etc.)
	PersonalUniqueCode string `json:"personal_unique_code,omitempty"`

	// PersonalUniqueCodeHash is SHA-256 hash of the unique code
	PersonalUniqueCodeHash string `json:"personal_unique_code_hash,omitempty"`

	// UserStatus is the schacUserStatus
	UserStatus string `json:"user_status,omitempty"`

	// CountryOfCitizenship is schacCountryOfCitizenship
	CountryOfCitizenship []string `json:"country_of_citizenship,omitempty"`
}

// UserAttributes represents all user attributes from SAML assertion
type UserAttributes struct {
	// EduPerson contains eduPerson schema attributes
	EduPerson EduPersonAttributes `json:"edu_person"`

	// Schac contains SCHAC attributes
	Schac SchacAttributes `json:"schac"`

	// DisplayName is the user's display name
	DisplayName string `json:"display_name,omitempty"`

	// GivenName is the user's first name
	GivenName string `json:"given_name,omitempty"`

	// Surname is the user's family name
	Surname string `json:"surname,omitempty"`

	// Email is the user's email address
	Email string `json:"email,omitempty"`

	// EmailHash is SHA-256 hash of email for privacy
	EmailHash string `json:"email_hash,omitempty"`

	// CommonName is the user's full name
	CommonName string `json:"common_name,omitempty"`

	// Raw contains raw attribute values for debugging
	Raw map[string][]string `json:"raw,omitempty"`
}

// HasAffiliation checks if user has a specific affiliation
func (a *UserAttributes) HasAffiliation(affiliation AffiliationType) bool {
	for _, aff := range a.EduPerson.Affiliation {
		if aff == affiliation {
			return true
		}
	}
	return false
}

// HasAnyAffiliation checks if user has any of the specified affiliations
func (a *UserAttributes) HasAnyAffiliation(affiliations ...AffiliationType) bool {
	for _, target := range affiliations {
		if a.HasAffiliation(target) {
			return true
		}
	}
	return false
}

// SAMLAssertion represents a verified SAML assertion
type SAMLAssertion struct {
	// ID is the assertion ID
	ID string `json:"id"`

	// IssuerEntityID is the IdP entity ID that issued the assertion
	IssuerEntityID string `json:"issuer_entity_id"`

	// SubjectNameID is the NameID value
	SubjectNameID string `json:"subject_name_id"`

	// SubjectNameIDFormat is the NameID format
	SubjectNameIDFormat string `json:"subject_name_id_format"`

	// Audience is the SP entity ID (audience restriction)
	Audience string `json:"audience"`

	// Attributes are the parsed user attributes
	Attributes UserAttributes `json:"attributes"`

	// AuthnInstant is when authentication occurred
	AuthnInstant time.Time `json:"authn_instant"`

	// AuthnContextClassRef is the authentication context class
	AuthnContextClassRef string `json:"authn_context_class_ref"`

	// SessionIndex is the IdP session index (for SLO)
	SessionIndex string `json:"session_index,omitempty"`

	// NotBefore is the earliest time the assertion is valid
	NotBefore time.Time `json:"not_before"`

	// NotOnOrAfter is when the assertion expires
	NotOnOrAfter time.Time `json:"not_on_or_after"`

	// IsMFA indicates if MFA was used (REFEDS MFA profile)
	IsMFA bool `json:"is_mfa"`

	// SignatureVerified indicates if signature was verified
	SignatureVerified bool `json:"signature_verified"`

	// RawXML is the raw assertion XML (for audit)
	// SECURITY: This should be encrypted at rest
	RawXML []byte `json:"raw_xml,omitempty"`
}

// Validate validates the assertion timing
func (a *SAMLAssertion) Validate(clockSkew time.Duration) error {
	now := time.Now()

	// Check NotBefore
	if !a.NotBefore.IsZero() && now.Add(clockSkew).Before(a.NotBefore) {
		return ErrAssertionNotYetValid
	}

	// Check NotOnOrAfter
	if !a.NotOnOrAfter.IsZero() && now.Add(-clockSkew).After(a.NotOnOrAfter) {
		return ErrAssertionExpired
	}

	return nil
}

// AuthnRequest represents a SAML authentication request
type AuthnRequest struct {
	// ID is the request ID
	ID string `json:"id"`

	// IssueInstant is when the request was created
	IssueInstant time.Time `json:"issue_instant"`

	// Destination is the IdP SSO endpoint URL
	Destination string `json:"destination"`

	// IssuerEntityID is the SP entity ID
	IssuerEntityID string `json:"issuer_entity_id"`

	// AssertionConsumerServiceURL is where the response should be sent
	AssertionConsumerServiceURL string `json:"assertion_consumer_service_url"`

	// ProtocolBinding is the response binding type
	ProtocolBinding string `json:"protocol_binding"`

	// NameIDPolicy specifies the NameID format preference
	NameIDPolicy string `json:"name_id_policy,omitempty"`

	// RequestedAuthnContext specifies authentication requirements
	RequestedAuthnContext []string `json:"requested_authn_context,omitempty"`

	// ForceAuthn requires fresh authentication
	ForceAuthn bool `json:"force_authn"`

	// IsPassive allows passive authentication only
	IsPassive bool `json:"is_passive"`

	// RelayState is the SP state to preserve across SSO
	RelayState string `json:"relay_state,omitempty"`
}

// AuthnRequestParams contains parameters for creating an AuthnRequest
type AuthnRequestParams struct {
	// InstitutionID is the IdP entity ID
	InstitutionID string `json:"institution_id"`

	// RelayState is the state to preserve
	RelayState string `json:"relay_state"`

	// RequireMFA forces REFEDS MFA profile
	RequireMFA bool `json:"require_mfa"`

	// ForceAuthn forces fresh authentication
	ForceAuthn bool `json:"force_authn"`

	// PreferredBinding is the preferred response binding
	PreferredBinding string `json:"preferred_binding,omitempty"`

	// NameIDFormat is the preferred NameID format
	NameIDFormat string `json:"name_id_format,omitempty"`
}

// Session represents an EduGAIN session
type Session struct {
	// ID is the session ID
	ID string `json:"id"`

	// WalletAddress is the associated wallet address
	WalletAddress string `json:"wallet_address"`

	// InstitutionID is the IdP entity ID
	InstitutionID string `json:"institution_id"`

	// InstitutionName is the IdP display name
	InstitutionName string `json:"institution_name"`

	// Attributes are the user attributes
	Attributes UserAttributes `json:"attributes"`

	// AuthnInstant is when authentication occurred
	AuthnInstant time.Time `json:"authn_instant"`

	// CreatedAt is when the session was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when the session expires
	ExpiresAt time.Time `json:"expires_at"`

	// Status is the session status
	Status SessionStatus `json:"status"`

	// IsMFA indicates if MFA was used
	IsMFA bool `json:"is_mfa"`

	// SessionIndex is the IdP session index (for SLO)
	SessionIndex string `json:"session_index,omitempty"`

	// VEIDScopeID is the associated VEID scope ID
	VEIDScopeID string `json:"veid_scope_id,omitempty"`

	// AssertionID is the original assertion ID
	AssertionID string `json:"assertion_id"`
}

// IsValid checks if the session is valid
func (s *Session) IsValid() bool {
	return s.Status == SessionStatusActive && time.Now().Before(s.ExpiresAt)
}

// TimeToExpiry returns duration until session expires
func (s *Session) TimeToExpiry() time.Duration {
	return time.Until(s.ExpiresAt)
}

// InstitutionSearchQuery represents search parameters for institutions
type InstitutionSearchQuery struct {
	// Query is the search text (name, entity ID, etc.)
	Query string `json:"query,omitempty"`

	// Country filters by country code
	Country string `json:"country,omitempty"`

	// Federation filters by federation name
	Federation string `json:"federation,omitempty"`

	// Type filters by institution type
	Type InstitutionType `json:"type,omitempty"`

	// SupportsMFA filters to only MFA-capable IdPs
	SupportsMFA *bool `json:"supports_mfa,omitempty"`

	// Limit is the maximum results to return
	Limit int `json:"limit,omitempty"`

	// Offset is the pagination offset
	Offset int `json:"offset,omitempty"`
}

// InstitutionSearchResult represents search results
type InstitutionSearchResult struct {
	// Institutions are the matching institutions
	Institutions []Institution `json:"institutions"`

	// TotalCount is the total matching count
	TotalCount int `json:"total_count"`

	// Query was the search query
	Query InstitutionSearchQuery `json:"query"`
}

// FederationStats contains federation statistics
type FederationStats struct {
	// TotalInstitutions is the total IdP count
	TotalInstitutions int `json:"total_institutions"`

	// ByCountry is counts by country code
	ByCountry map[string]int `json:"by_country"`

	// ByFederation is counts by federation name
	ByFederation map[string]int `json:"by_federation"`

	// ByType is counts by institution type
	ByType map[InstitutionType]int `json:"by_type"`

	// MFACapableCount is IdPs supporting REFEDS MFA
	MFACapableCount int `json:"mfa_capable_count"`

	// LastRefresh is when stats were last computed
	LastRefresh time.Time `json:"last_refresh"`
}
