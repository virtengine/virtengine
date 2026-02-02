// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"context"
	"time"
)

// ============================================================================
// Service Interface
// ============================================================================

// Service defines the main EduGAIN service interface
type Service interface {
	// Lifecycle methods

	// Start starts the service (metadata refresh, etc.)
	Start(ctx context.Context) error

	// Stop stops the service gracefully
	Stop(ctx context.Context) error

	// IsHealthy returns true if the service is healthy
	IsHealthy() bool

	// GetStatus returns the service status
	GetStatus() ServiceStatus

	// Metadata methods

	// RefreshMetadata forces a metadata refresh
	RefreshMetadata(ctx context.Context) error

	// GetMetadata returns the current federation metadata
	GetMetadata() (*FederationMetadata, error)

	// GetMetadataStatus returns the metadata status
	GetMetadataStatus() MetadataStatus

	// Discovery methods

	// DiscoverInstitutions searches for institutions
	DiscoverInstitutions(ctx context.Context, query InstitutionSearchQuery) (*InstitutionSearchResult, error)

	// GetInstitution returns a specific institution by entity ID
	GetInstitution(ctx context.Context, entityID string) (*Institution, error)

	// GetFederationStats returns federation statistics
	GetFederationStats(ctx context.Context) (*FederationStats, error)

	// SAML methods

	// CreateAuthnRequest creates a SAML AuthnRequest
	CreateAuthnRequest(ctx context.Context, params AuthnRequestParams) (*AuthnRequestResult, error)

	// VerifyResponse verifies a SAML response and returns the assertion
	VerifyResponse(ctx context.Context, samlResponseBase64 string) (*SAMLAssertion, error)

	// CreateLogoutRequest creates a SAML LogoutRequest
	CreateLogoutRequest(ctx context.Context, sessionID string) (*LogoutRequestResult, error)

	// Session methods

	// CreateSession creates a session from a verified assertion
	CreateSession(ctx context.Context, assertion *SAMLAssertion, walletAddress string) (*Session, error)

	// GetSession returns a session by ID
	GetSession(ctx context.Context, sessionID string) (*Session, error)

	// ValidateSession validates a session token
	ValidateSession(ctx context.Context, sessionToken string) (*Session, error)

	// RevokeSession revokes a session
	RevokeSession(ctx context.Context, sessionID string) error

	// ListSessions lists sessions for a wallet
	ListSessions(ctx context.Context, walletAddress string) ([]Session, error)

	// VEID integration methods

	// CreateVEIDScope creates a VEID scope from a session
	CreateVEIDScope(ctx context.Context, session *Session) (*VEIDScope, error)

	// EnrichVEID enriches an existing VEID with EduGAIN attributes
	EnrichVEID(ctx context.Context, session *Session, veidID string) error
}

// ============================================================================
// Metadata Service Interface
// ============================================================================

// MetadataService defines the metadata management interface
type MetadataService interface {
	// Load loads metadata from the configured source
	Load(ctx context.Context) error

	// Refresh refreshes metadata (respects cache settings)
	Refresh(ctx context.Context) error

	// ForceRefresh forces a metadata refresh ignoring cache
	ForceRefresh(ctx context.Context) error

	// Get returns the current metadata
	Get() (*FederationMetadata, error)

	// GetStatus returns the metadata status
	GetStatus() MetadataStatus

	// IsValid returns true if metadata is valid
	IsValid() bool

	// FindInstitution finds an institution by entity ID
	FindInstitution(entityID string) (*Institution, error)

	// SearchInstitutions searches institutions
	SearchInstitutions(query InstitutionSearchQuery) (*InstitutionSearchResult, error)

	// GetStats returns federation statistics
	GetStats() *FederationStats

	// SetCertificate sets the certificate for signature verification
	SetCertificate(certPEM []byte) error

	// OnRefresh registers a callback for refresh events
	OnRefresh(callback MetadataRefreshCallback)
}

// MetadataRefreshCallback is called when metadata is refreshed
type MetadataRefreshCallback func(metadata *FederationMetadata, err error)

// MetadataStatus contains metadata status information
type MetadataStatus struct {
	// Status is the current status
	Status FederationStatus `json:"status"`

	// LastRefresh is when metadata was last refreshed
	LastRefresh time.Time `json:"last_refresh"`

	// NextRefresh is when the next refresh is scheduled
	NextRefresh time.Time `json:"next_refresh"`

	// ValidUntil is when the metadata expires
	ValidUntil time.Time `json:"valid_until"`

	// InstitutionCount is the number of institutions
	InstitutionCount int `json:"institution_count"`

	// LastError is the last error (if any)
	LastError string `json:"last_error,omitempty"`

	// SignatureValid indicates if signature was verified
	SignatureValid bool `json:"signature_valid"`
}

// ============================================================================
// SAML Provider Interface
// ============================================================================

// SAMLProvider defines the SAML 2.0 service provider interface
type SAMLProvider interface {
	// GetEntityID returns the SP entity ID
	GetEntityID() string

	// GetMetadata returns the SP metadata XML
	GetMetadata() ([]byte, error)

	// CreateAuthnRequest creates an AuthnRequest
	CreateAuthnRequest(ctx context.Context, idp *Institution, params AuthnRequestParams) (*AuthnRequestResult, error)

	// VerifyResponse verifies a SAML response
	VerifyResponse(ctx context.Context, samlResponseBase64 string, idp *Institution) (*SAMLAssertion, error)

	// CreateLogoutRequest creates a LogoutRequest
	CreateLogoutRequest(ctx context.Context, session *Session, idp *Institution) (*LogoutRequestResult, error)

	// VerifyLogoutResponse verifies a LogoutResponse
	VerifyLogoutResponse(ctx context.Context, samlResponseBase64 string) error

	// DecryptAssertion decrypts an encrypted assertion
	DecryptAssertion(encryptedXML []byte) ([]byte, error)

	// SetSigningCredentials sets the signing certificate and key
	SetSigningCredentials(certPEM, keyPEM []byte) error

	// SetEncryptionCredentials sets the encryption certificate and key
	SetEncryptionCredentials(certPEM, keyPEM []byte) error
}

// AuthnRequestResult contains the result of creating an AuthnRequest
type AuthnRequestResult struct {
	// Request is the AuthnRequest
	Request *AuthnRequest `json:"request"`

	// URL is the redirect URL (for HTTP-Redirect binding)
	URL string `json:"url,omitempty"`

	// SAMLRequest is the encoded request (for HTTP-POST binding)
	SAMLRequest string `json:"saml_request,omitempty"`

	// RelayState is the relay state
	RelayState string `json:"relay_state"`

	// Binding is the binding type used
	Binding string `json:"binding"`

	// PostFormHTML is the HTML form for HTTP-POST binding
	PostFormHTML string `json:"post_form_html,omitempty"`
}

// LogoutRequestResult contains the result of creating a LogoutRequest
type LogoutRequestResult struct {
	// URL is the redirect URL
	URL string `json:"url,omitempty"`

	// SAMLRequest is the encoded request
	SAMLRequest string `json:"saml_request,omitempty"`

	// RelayState is the relay state
	RelayState string `json:"relay_state"`

	// Binding is the binding type used
	Binding string `json:"binding"`
}

// ============================================================================
// Session Manager Interface
// ============================================================================

// SessionManager defines the session management interface
type SessionManager interface {
	// Create creates a new session
	Create(ctx context.Context, assertion *SAMLAssertion, walletAddress string) (*Session, error)

	// Get returns a session by ID
	Get(ctx context.Context, sessionID string) (*Session, error)

	// ValidateToken validates a session token and returns the session
	ValidateToken(ctx context.Context, token string) (*Session, error)

	// GenerateToken generates a session token
	GenerateToken(session *Session) (string, error)

	// Revoke revokes a session
	Revoke(ctx context.Context, sessionID string) error

	// RevokeAll revokes all sessions for a wallet
	RevokeAll(ctx context.Context, walletAddress string) error

	// List lists sessions for a wallet
	List(ctx context.Context, walletAddress string) ([]Session, error)

	// Cleanup removes expired sessions
	Cleanup(ctx context.Context) (int, error)

	// GetStats returns session statistics
	GetStats(ctx context.Context) (*SessionStats, error)

	// TrackAssertionID tracks an assertion ID for replay detection
	TrackAssertionID(ctx context.Context, assertionID string, expiry time.Time) error

	// IsAssertionReplayed checks if an assertion ID has been seen
	IsAssertionReplayed(ctx context.Context, assertionID string) (bool, error)
}

// SessionStats contains session statistics
type SessionStats struct {
	// TotalSessions is the total session count
	TotalSessions int `json:"total_sessions"`

	// ActiveSessions is the active session count
	ActiveSessions int `json:"active_sessions"`

	// ExpiredSessions is the expired session count
	ExpiredSessions int `json:"expired_sessions"`

	// UniqueWallets is the unique wallet count
	UniqueWallets int `json:"unique_wallets"`

	// ByInstitution is counts by institution
	ByInstitution map[string]int `json:"by_institution"`
}

// ============================================================================
// Attribute Mapper Interface
// ============================================================================

// AttributeMapper defines the attribute mapping interface
type AttributeMapper interface {
	// MapAttributes maps raw SAML attributes to UserAttributes
	MapAttributes(rawAttrs map[string][]string) (*UserAttributes, error)

	// GetRequiredAttributes returns required attribute OIDs/names
	GetRequiredAttributes() []string

	// GetOptionalAttributes returns optional attribute OIDs/names
	GetOptionalAttributes() []string

	// ValidateAttributes validates required attributes are present
	ValidateAttributes(attrs *UserAttributes) error

	// HashSensitiveData hashes PII fields
	HashSensitiveData(attrs *UserAttributes) *UserAttributes

	// ExtractAffiliations extracts affiliation types from attributes
	ExtractAffiliations(attrs *UserAttributes) []AffiliationType
}

// ============================================================================
// VEID Integration Interface
// ============================================================================

// VEIDIntegrator defines the VEID integration interface
type VEIDIntegrator interface {
	// CreateScope creates a VEID scope from EduGAIN session
	CreateScope(ctx context.Context, session *Session) (*VEIDScope, error)

	// EnrichIdentity enriches an existing VEID identity
	EnrichIdentity(ctx context.Context, session *Session, veidID string) error

	// GetExistingScope returns an existing EduGAIN scope for a wallet
	GetExistingScope(ctx context.Context, walletAddress string) (*VEIDScope, error)

	// UpdateScope updates an existing EduGAIN scope
	UpdateScope(ctx context.Context, scope *VEIDScope, session *Session) error

	// RevokeScope revokes an EduGAIN scope
	RevokeScope(ctx context.Context, scopeID string) error

	// ComputeScoreContribution computes identity score contribution
	ComputeScoreContribution(attrs *UserAttributes) uint32
}

// VEIDScope represents an EduGAIN scope in VEID
type VEIDScope struct {
	// ID is the scope ID
	ID string `json:"id"`

	// WalletAddress is the associated wallet
	WalletAddress string `json:"wallet_address"`

	// InstitutionID is the IdP entity ID
	InstitutionID string `json:"institution_id"`

	// InstitutionName is the IdP display name
	InstitutionName string `json:"institution_name"`

	// Federation is the federation name
	Federation string `json:"federation"`

	// Country is the institution country
	Country string `json:"country"`

	// PrincipalNameHash is hash of eduPersonPrincipalName
	PrincipalNameHash string `json:"principal_name_hash"`

	// HomeOrganization is schacHomeOrganization
	HomeOrganization string `json:"home_organization"`

	// Affiliations are the user's affiliations
	Affiliations []AffiliationType `json:"affiliations"`

	// IsMFA indicates if MFA was used
	IsMFA bool `json:"is_mfa"`

	// ScoreContribution is the identity score contribution
	ScoreContribution uint32 `json:"score_contribution"`

	// AuthnInstant is when authentication occurred
	AuthnInstant time.Time `json:"authn_instant"`

	// ExpiresAt is when the scope expires
	ExpiresAt time.Time `json:"expires_at"`

	// CreatedAt is when the scope was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when the scope was last updated
	UpdatedAt time.Time `json:"updated_at"`

	// Status is the scope status
	Status string `json:"status"`
}

// ============================================================================
// Discovery Service Interface
// ============================================================================

// DiscoveryService defines the institution discovery (WAYF) interface
type DiscoveryService interface {
	// Search searches for institutions by query
	Search(ctx context.Context, query InstitutionSearchQuery) (*InstitutionSearchResult, error)

	// GetByEntityID returns an institution by entity ID
	GetByEntityID(ctx context.Context, entityID string) (*Institution, error)

	// GetByCountry returns institutions by country code
	GetByCountry(ctx context.Context, countryCode string, limit int) ([]Institution, error)

	// GetByFederation returns institutions by federation name
	GetByFederation(ctx context.Context, federationName string, limit int) ([]Institution, error)

	// GetPopular returns popular/frequently used institutions
	GetPopular(ctx context.Context, limit int) ([]Institution, error)

	// GetRecent returns recently used institutions for a wallet
	GetRecent(ctx context.Context, walletAddress string, limit int) ([]Institution, error)

	// RecordUsage records institution usage for popularity tracking
	RecordUsage(ctx context.Context, entityID string, walletAddress string) error

	// GetStats returns discovery statistics
	GetStats(ctx context.Context) (*DiscoveryStats, error)
}

// DiscoveryStats contains discovery statistics
type DiscoveryStats struct {
	// TotalSearches is the total search count
	TotalSearches int64 `json:"total_searches"`

	// TopInstitutions are the most popular institutions
	TopInstitutions []InstitutionUsage `json:"top_institutions"`

	// TopCountries are the most popular countries
	TopCountries map[string]int64 `json:"top_countries"`
}

// InstitutionUsage tracks institution usage
type InstitutionUsage struct {
	// EntityID is the institution entity ID
	EntityID string `json:"entity_id"`

	// DisplayName is the institution name
	DisplayName string `json:"display_name"`

	// UsageCount is the usage count
	UsageCount int64 `json:"usage_count"`
}

// ============================================================================
// Service Status
// ============================================================================

// ServiceStatus contains service status information
type ServiceStatus struct {
	// Healthy indicates if the service is healthy
	Healthy bool `json:"healthy"`

	// Version is the service version
	Version string `json:"version"`

	// StartedAt is when the service started
	StartedAt time.Time `json:"started_at"`

	// Uptime is the service uptime
	Uptime time.Duration `json:"uptime"`

	// Metadata contains metadata status
	Metadata MetadataStatus `json:"metadata"`

	// Sessions contains session statistics
	Sessions *SessionStats `json:"sessions,omitempty"`

	// LastError is the last error (if any)
	LastError string `json:"last_error,omitempty"`
}
