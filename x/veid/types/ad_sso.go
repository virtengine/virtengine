// Package types provides types for the VEID module.
//
// VE-907: Active Directory SSO verification scope
// This file defines types for Active Directory integration supporting:
// - Azure AD via OIDC (OpenID Connect)
// - SAML 2.0 for enterprise federation
// - On-premises AD via LDAP
// - Wallet binding to AD identity
package types

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// ============================================================================
// AD SSO Constants
// ============================================================================

// ADSSOVersion is the current version of the AD SSO verification format
const ADSSOVersion uint32 = 1

// ADAuthMethod defines the authentication method used
type ADAuthMethod string

const (
	// ADAuthMethodOIDC represents Azure AD authentication via OpenID Connect
	ADAuthMethodOIDC ADAuthMethod = "oidc"

	// ADAuthMethodSAML represents SAML 2.0 authentication
	ADAuthMethodSAML ADAuthMethod = "saml"

	// ADAuthMethodLDAP represents on-premises AD via LDAP
	ADAuthMethodLDAP ADAuthMethod = "ldap"
)

// AllADAuthMethods returns all valid AD authentication methods
func AllADAuthMethods() []ADAuthMethod {
	return []ADAuthMethod{
		ADAuthMethodOIDC,
		ADAuthMethodSAML,
		ADAuthMethodLDAP,
	}
}

// IsValidADAuthMethod checks if an authentication method is valid
func IsValidADAuthMethod(m ADAuthMethod) bool {
	for _, valid := range AllADAuthMethods() {
		if m == valid {
			return true
		}
	}
	return false
}

// ADSSOStatus represents the status of an AD SSO verification
type ADSSOStatus string

const (
	// ADSSOStatusPending indicates verification is pending
	ADSSOStatusPending ADSSOStatus = "pending"

	// ADSSOStatusVerified indicates verification is complete
	ADSSOStatusVerified ADSSOStatus = "verified"

	// ADSSOStatusFailed indicates verification failed
	ADSSOStatusFailed ADSSOStatus = "failed"

	// ADSSOStatusRevoked indicates verification was revoked
	ADSSOStatusRevoked ADSSOStatus = "revoked"

	// ADSSOStatusExpired indicates verification has expired
	ADSSOStatusExpired ADSSOStatus = "expired"
)

// AllADSSOStatuses returns all valid AD SSO statuses
func AllADSSOStatuses() []ADSSOStatus {
	return []ADSSOStatus{
		ADSSOStatusPending,
		ADSSOStatusVerified,
		ADSSOStatusFailed,
		ADSSOStatusRevoked,
		ADSSOStatusExpired,
	}
}

// IsValidADSSOStatus checks if a status is valid
func IsValidADSSOStatus(s ADSSOStatus) bool {
	for _, valid := range AllADSSOStatuses() {
		if s == valid {
			return true
		}
	}
	return false
}

// ============================================================================
// Azure AD OIDC Types
// ============================================================================

// AzureADOIDCConfig holds configuration for Azure AD OIDC authentication
// SECURITY: Sensitive fields (client_secret) are never stored on-chain
type AzureADOIDCConfig struct {
	// TenantID is the Azure AD tenant identifier
	// Stored as hash for privacy
	TenantIDHash string `json:"tenant_id_hash"`

	// Issuer is the OIDC issuer URL (e.g., "https://login.microsoftonline.com/{tenant}/v2.0")
	Issuer string `json:"issuer"`

	// ClientIDHash is a hash of the registered application client ID
	ClientIDHash string `json:"client_id_hash"`

	// AllowedAudiences specifies valid token audiences
	AllowedAudiencesHash []string `json:"allowed_audiences_hash,omitempty"`
}

// NewAzureADOIDCConfig creates a new Azure AD OIDC configuration
// Parameters are hashed before storage
func NewAzureADOIDCConfig(tenantID, issuer, clientID string) *AzureADOIDCConfig {
	return &AzureADOIDCConfig{
		TenantIDHash: HashADIdentifier(tenantID),
		Issuer:       issuer,
		ClientIDHash: HashADIdentifier(clientID),
	}
}

// Validate validates the Azure AD OIDC configuration
func (c *AzureADOIDCConfig) Validate() error {
	if c.TenantIDHash == "" {
		return ErrInvalidADSSO.Wrap("tenant_id_hash cannot be empty")
	}
	if len(c.TenantIDHash) != 64 {
		return ErrInvalidADSSO.Wrap("tenant_id_hash must be a valid SHA256 hex string")
	}
	if c.Issuer == "" {
		return ErrInvalidADSSO.Wrap("issuer cannot be empty")
	}
	if c.ClientIDHash == "" {
		return ErrInvalidADSSO.Wrap("client_id_hash cannot be empty")
	}
	if len(c.ClientIDHash) != 64 {
		return ErrInvalidADSSO.Wrap("client_id_hash must be a valid SHA256 hex string")
	}
	return nil
}

// ============================================================================
// SAML 2.0 Types
// ============================================================================

// SAMLConfig holds configuration for SAML 2.0 authentication
type SAMLConfig struct {
	// EntityID is the identity provider entity ID
	EntityIDHash string `json:"entity_id_hash"`

	// MetadataURL is the IdP metadata URL (optional, for reference)
	MetadataURL string `json:"metadata_url,omitempty"`

	// CertificateFingerprintHash is a hash of the IdP certificate fingerprint
	CertificateFingerprintHash string `json:"certificate_fingerprint_hash"`

	// NameIDFormat specifies the expected NameID format
	NameIDFormat string `json:"name_id_format"`
}

// NewSAMLConfig creates a new SAML configuration
func NewSAMLConfig(entityID, metadataURL, certFingerprint, nameIDFormat string) *SAMLConfig {
	return &SAMLConfig{
		EntityIDHash:               HashADIdentifier(entityID),
		MetadataURL:                metadataURL,
		CertificateFingerprintHash: HashADIdentifier(certFingerprint),
		NameIDFormat:               nameIDFormat,
	}
}

// Validate validates the SAML configuration
func (c *SAMLConfig) Validate() error {
	if c.EntityIDHash == "" {
		return ErrInvalidADSSO.Wrap("entity_id_hash cannot be empty")
	}
	if len(c.EntityIDHash) != 64 {
		return ErrInvalidADSSO.Wrap("entity_id_hash must be a valid SHA256 hex string")
	}
	if c.CertificateFingerprintHash == "" {
		return ErrInvalidADSSO.Wrap("certificate_fingerprint_hash cannot be empty")
	}
	if len(c.CertificateFingerprintHash) != 64 {
		return ErrInvalidADSSO.Wrap("certificate_fingerprint_hash must be a valid SHA256 hex string")
	}
	if c.NameIDFormat == "" {
		return ErrInvalidADSSO.Wrap("name_id_format cannot be empty")
	}
	return nil
}

// ============================================================================
// LDAP Types
// ============================================================================

// LDAPConfig holds configuration for on-premises AD via LDAP
// SECURITY: No credentials stored - only connection verification metadata
type LDAPConfig struct {
	// ServerHash is a hash of the LDAP server address
	ServerHash string `json:"server_hash"`

	// BaseDNHash is a hash of the base distinguished name
	BaseDNHash string `json:"base_dn_hash"`

	// DomainHash is a hash of the AD domain
	DomainHash string `json:"domain_hash"`

	// UseSSL indicates if LDAPS was used
	UseSSL bool `json:"use_ssl"`

	// UseTLS indicates if StartTLS was used
	UseTLS bool `json:"use_tls"`
}

// NewLDAPConfig creates a new LDAP configuration
func NewLDAPConfig(server, baseDN, domain string, useSSL, useTLS bool) *LDAPConfig {
	return &LDAPConfig{
		ServerHash: HashADIdentifier(server),
		BaseDNHash: HashADIdentifier(baseDN),
		DomainHash: HashADIdentifier(domain),
		UseSSL:     useSSL,
		UseTLS:     useTLS,
	}
}

// Validate validates the LDAP configuration
func (c *LDAPConfig) Validate() error {
	if c.ServerHash == "" {
		return ErrInvalidADSSO.Wrap("server_hash cannot be empty")
	}
	if len(c.ServerHash) != 64 {
		return ErrInvalidADSSO.Wrap("server_hash must be a valid SHA256 hex string")
	}
	if c.BaseDNHash == "" {
		return ErrInvalidADSSO.Wrap("base_dn_hash cannot be empty")
	}
	if len(c.BaseDNHash) != 64 {
		return ErrInvalidADSSO.Wrap("base_dn_hash must be a valid SHA256 hex string")
	}
	if c.DomainHash == "" {
		return ErrInvalidADSSO.Wrap("domain_hash cannot be empty")
	}
	if len(c.DomainHash) != 64 {
		return ErrInvalidADSSO.Wrap("domain_hash must be a valid SHA256 hex string")
	}
	// At least one security layer should be enabled
	if !c.UseSSL && !c.UseTLS {
		return ErrInvalidADSSO.Wrap("at least UseSSL or UseTLS must be enabled")
	}
	return nil
}

// ============================================================================
// AD SSO Linkage Metadata
// ============================================================================

// ADSSOLinkageMetadata contains the minimal on-chain metadata for AD SSO linkage
// SECURITY: No sensitive credentials, tokens, or PII stored
type ADSSOLinkageMetadata struct {
	// Version is the format version
	Version uint32 `json:"version"`

	// LinkageID is a unique identifier for this linkage
	LinkageID string `json:"linkage_id"`

	// AccountAddress is the blockchain account bound to this AD identity
	AccountAddress string `json:"account_address"`

	// AuthMethod is the authentication method used
	AuthMethod ADAuthMethod `json:"auth_method"`

	// SubjectHash is a SHA256 hash of the AD subject/user identifier
	// This allows verification without exposing the raw subject ID
	SubjectHash string `json:"subject_hash"`

	// UPNHash is a SHA256 hash of the User Principal Name (optional)
	UPNHash string `json:"upn_hash,omitempty"`

	// TenantHash is a SHA256 hash of the tenant/domain identifier
	TenantHash string `json:"tenant_hash"`

	// GroupHashes contains SHA256 hashes of group memberships (optional)
	GroupHashes []string `json:"group_hashes,omitempty"`

	// Nonce is the challenge nonce used during verification
	Nonce string `json:"nonce"`

	// Status is the current status of this linkage
	Status ADSSOStatus `json:"status"`

	// VerifiedAt is when this linkage was verified
	VerifiedAt time.Time `json:"verified_at"`

	// ExpiresAt is when this linkage expires
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RevokedAt is when this linkage was revoked (if applicable)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// RevokedReason is the reason for revocation
	RevokedReason string `json:"revoked_reason,omitempty"`

	// WalletBindingSignature is the signature proving wallet authorization
	WalletBindingSignature []byte `json:"wallet_binding_signature"`

	// OIDCConfig holds OIDC-specific configuration (if OIDC method)
	OIDCConfig *AzureADOIDCConfig `json:"oidc_config,omitempty"`

	// SAMLConfig holds SAML-specific configuration (if SAML method)
	SAMLConfig *SAMLConfig `json:"saml_config,omitempty"`

	// LDAPConfig holds LDAP-specific configuration (if LDAP method)
	LDAPConfig *LDAPConfig `json:"ldap_config,omitempty"`

	// CreatedAt is when this linkage was created
	CreatedAt time.Time `json:"created_at"`

	// UpdatedAt is when this linkage was last updated
	UpdatedAt time.Time `json:"updated_at"`
}

// NewADSSOLinkageMetadata creates a new AD SSO linkage metadata record
func NewADSSOLinkageMetadata(
	linkageID string,
	accountAddress string,
	authMethod ADAuthMethod,
	subjectID string,
	tenantID string,
	nonce string,
	verifiedAt time.Time,
) *ADSSOLinkageMetadata {
	now := verifiedAt
	return &ADSSOLinkageMetadata{
		Version:        ADSSOVersion,
		LinkageID:      linkageID,
		AccountAddress: accountAddress,
		AuthMethod:     authMethod,
		SubjectHash:    HashADIdentifier(subjectID),
		TenantHash:     HashADIdentifier(tenantID),
		Nonce:          nonce,
		Status:         ADSSOStatusVerified,
		VerifiedAt:     verifiedAt,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

// HashADIdentifier creates a SHA256 hash of an AD identifier
func HashADIdentifier(identifier string) string {
	if identifier == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(identifier))
	return hex.EncodeToString(hash[:])
}

// Validate validates the AD SSO linkage metadata
func (m *ADSSOLinkageMetadata) Validate() error {
	if m.Version == 0 || m.Version > ADSSOVersion {
		return ErrInvalidADSSO.Wrapf("unsupported version: %d", m.Version)
	}

	if m.LinkageID == "" {
		return ErrInvalidADSSO.Wrap("linkage_id cannot be empty")
	}

	if m.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address cannot be empty")
	}

	if !IsValidADAuthMethod(m.AuthMethod) {
		return ErrInvalidADSSO.Wrapf("invalid auth_method: %s", m.AuthMethod)
	}

	if m.SubjectHash == "" {
		return ErrInvalidADSSO.Wrap("subject_hash cannot be empty")
	}

	if len(m.SubjectHash) != 64 {
		return ErrInvalidADSSO.Wrap("subject_hash must be a valid SHA256 hex string")
	}

	if m.TenantHash == "" {
		return ErrInvalidADSSO.Wrap("tenant_hash cannot be empty")
	}

	if len(m.TenantHash) != 64 {
		return ErrInvalidADSSO.Wrap("tenant_hash must be a valid SHA256 hex string")
	}

	if m.Nonce == "" {
		return ErrInvalidADSSO.Wrap("nonce cannot be empty")
	}

	if !IsValidADSSOStatus(m.Status) {
		return ErrInvalidADSSO.Wrapf("invalid status: %s", m.Status)
	}

	if m.VerifiedAt.IsZero() {
		return ErrInvalidADSSO.Wrap("verified_at cannot be zero")
	}

	if m.CreatedAt.IsZero() {
		return ErrInvalidADSSO.Wrap("created_at cannot be zero")
	}

	if len(m.WalletBindingSignature) == 0 && m.Status == ADSSOStatusVerified {
		return ErrInvalidADSSO.Wrap("wallet_binding_signature required for verified status")
	}

	// Validate method-specific config
	switch m.AuthMethod {
	case ADAuthMethodOIDC:
		if m.OIDCConfig == nil {
			return ErrInvalidADSSO.Wrap("oidc_config required for OIDC method")
		}
		if err := m.OIDCConfig.Validate(); err != nil {
			return err
		}
	case ADAuthMethodSAML:
		if m.SAMLConfig == nil {
			return ErrInvalidADSSO.Wrap("saml_config required for SAML method")
		}
		if err := m.SAMLConfig.Validate(); err != nil {
			return err
		}
	case ADAuthMethodLDAP:
		if m.LDAPConfig == nil {
			return ErrInvalidADSSO.Wrap("ldap_config required for LDAP method")
		}
		if err := m.LDAPConfig.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IsActive returns true if the linkage is currently valid
func (m *ADSSOLinkageMetadata) IsActive() bool {
	if m.Status != ADSSOStatusVerified {
		return false
	}
	if m.ExpiresAt != nil && time.Now().After(*m.ExpiresAt) {
		return false
	}
	return true
}

// SetUPNHash sets the User Principal Name hash
func (m *ADSSOLinkageMetadata) SetUPNHash(upn string) {
	m.UPNHash = HashADIdentifier(upn)
	m.UpdatedAt = time.Now()
}

// AddGroupHash adds a group membership hash
func (m *ADSSOLinkageMetadata) AddGroupHash(groupID string) {
	hash := HashADIdentifier(groupID)
	for _, existing := range m.GroupHashes {
		if existing == hash {
			return // Already exists
		}
	}
	m.GroupHashes = append(m.GroupHashes, hash)
	m.UpdatedAt = time.Now()
}

// HasGroupHash checks if a group hash exists
func (m *ADSSOLinkageMetadata) HasGroupHash(groupID string) bool {
	hash := HashADIdentifier(groupID)
	for _, existing := range m.GroupHashes {
		if existing == hash {
			return true
		}
	}
	return false
}

// SetOIDCConfig sets the OIDC configuration
func (m *ADSSOLinkageMetadata) SetOIDCConfig(config *AzureADOIDCConfig) {
	m.OIDCConfig = config
	m.UpdatedAt = time.Now()
}

// SetSAMLConfig sets the SAML configuration
func (m *ADSSOLinkageMetadata) SetSAMLConfig(config *SAMLConfig) {
	m.SAMLConfig = config
	m.UpdatedAt = time.Now()
}

// SetLDAPConfig sets the LDAP configuration
func (m *ADSSOLinkageMetadata) SetLDAPConfig(config *LDAPConfig) {
	m.LDAPConfig = config
	m.UpdatedAt = time.Now()
}

// SetWalletBinding sets the wallet binding signature
func (m *ADSSOLinkageMetadata) SetWalletBinding(signature []byte) {
	m.WalletBindingSignature = signature
	m.UpdatedAt = time.Now()
}

// Revoke marks the linkage as revoked
func (m *ADSSOLinkageMetadata) Revoke(reason string) {
	now := time.Now()
	m.Status = ADSSOStatusRevoked
	m.RevokedAt = &now
	m.RevokedReason = reason
	m.UpdatedAt = now
}

// String returns a string representation (non-sensitive)
func (m *ADSSOLinkageMetadata) String() string {
	return fmt.Sprintf("ADSSOLinkage{ID: %s, Method: %s, Status: %s}",
		m.LinkageID, m.AuthMethod, m.Status)
}

// ============================================================================
// AD SSO Verification Challenge
// ============================================================================

// ADSSOChallenge represents a pending AD SSO verification challenge
type ADSSOChallenge struct {
	// ChallengeID is a unique identifier for this challenge
	ChallengeID string `json:"challenge_id"`

	// AccountAddress is the account requesting AD SSO linkage
	AccountAddress string `json:"account_address"`

	// AuthMethod is the authentication method to use
	AuthMethod ADAuthMethod `json:"auth_method"`

	// Nonce is the challenge nonce (state parameter for OIDC, RelayState for SAML)
	Nonce string `json:"nonce"`

	// CreatedAt is when this challenge was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this challenge expires
	ExpiresAt time.Time `json:"expires_at"`

	// Status indicates the challenge status
	Status ADSSOStatus `json:"status"`

	// RedirectURI is where the auth flow should redirect (OIDC/SAML)
	RedirectURI string `json:"redirect_uri,omitempty"`

	// CompletedAt is when this challenge was completed
	CompletedAt *time.Time `json:"completed_at,omitempty"`

	// TenantHint is an optional tenant hint for multi-tenant scenarios
	TenantHint string `json:"tenant_hint,omitempty"`
}

// ChallengeTTLSeconds is the default TTL for AD SSO challenges (10 minutes)
const ChallengeTTLSeconds int64 = 600

// NewADSSOChallenge creates a new AD SSO verification challenge
func NewADSSOChallenge(
	challengeID string,
	accountAddress string,
	authMethod ADAuthMethod,
	nonce string,
	createdAt time.Time,
	ttlSeconds int64,
) *ADSSOChallenge {
	if ttlSeconds <= 0 {
		ttlSeconds = ChallengeTTLSeconds
	}
	expiresAt := createdAt.Add(time.Duration(ttlSeconds) * time.Second)
	return &ADSSOChallenge{
		ChallengeID:    challengeID,
		AccountAddress: accountAddress,
		AuthMethod:     authMethod,
		Nonce:          nonce,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         ADSSOStatusPending,
	}
}

// Validate validates the AD SSO challenge
func (c *ADSSOChallenge) Validate() error {
	if c.ChallengeID == "" {
		return ErrInvalidADSSO.Wrap("challenge_id cannot be empty")
	}
	if c.AccountAddress == "" {
		return ErrInvalidAddress.Wrap("account_address cannot be empty")
	}
	if !IsValidADAuthMethod(c.AuthMethod) {
		return ErrInvalidADSSO.Wrapf("invalid auth_method: %s", c.AuthMethod)
	}
	if c.Nonce == "" {
		return ErrInvalidADSSO.Wrap("nonce cannot be empty")
	}
	if c.CreatedAt.IsZero() {
		return ErrInvalidADSSO.Wrap("created_at cannot be zero")
	}
	if c.ExpiresAt.IsZero() {
		return ErrInvalidADSSO.Wrap("expires_at cannot be zero")
	}
	if !c.ExpiresAt.After(c.CreatedAt) {
		return ErrInvalidADSSO.Wrap("expires_at must be after created_at")
	}
	return nil
}

// IsExpired returns true if the challenge has expired
func (c *ADSSOChallenge) IsExpired(now time.Time) bool {
	return now.After(c.ExpiresAt)
}

// Complete marks the challenge as completed
func (c *ADSSOChallenge) Complete() {
	now := time.Now()
	c.Status = ADSSOStatusVerified
	c.CompletedAt = &now
}

// ============================================================================
// AD SSO Wallet Binding
// ============================================================================

// ADWalletBinding represents the binding between an AD identity and a wallet
type ADWalletBinding struct {
	// BindingID is a unique identifier for this binding
	BindingID string `json:"binding_id"`

	// WalletAddress is the blockchain wallet address
	WalletAddress string `json:"wallet_address"`

	// LinkageID references the AD SSO linkage
	LinkageID string `json:"linkage_id"`

	// SubjectHash is the hashed AD subject identifier
	SubjectHash string `json:"subject_hash"`

	// TenantHash is the hashed tenant identifier
	TenantHash string `json:"tenant_hash"`

	// AuthMethod is the authentication method used
	AuthMethod ADAuthMethod `json:"auth_method"`

	// BindingSignature is the cryptographic signature proving authorization
	BindingSignature []byte `json:"binding_signature"`

	// BindingMessage is the signed message content (for verification)
	BindingMessage string `json:"binding_message"`

	// CreatedAt is when this binding was created
	CreatedAt time.Time `json:"created_at"`

	// ExpiresAt is when this binding expires (optional)
	ExpiresAt *time.Time `json:"expires_at,omitempty"`

	// RevokedAt is when this binding was revoked (if applicable)
	RevokedAt *time.Time `json:"revoked_at,omitempty"`

	// Active indicates if this binding is currently active
	Active bool `json:"active"`
}

// NewADWalletBinding creates a new AD wallet binding
func NewADWalletBinding(
	bindingID string,
	walletAddress string,
	linkageID string,
	subjectID string,
	tenantID string,
	authMethod ADAuthMethod,
	bindingMessage string,
	signature []byte,
	createdAt time.Time,
) *ADWalletBinding {
	return &ADWalletBinding{
		BindingID:        bindingID,
		WalletAddress:    walletAddress,
		LinkageID:        linkageID,
		SubjectHash:      HashADIdentifier(subjectID),
		TenantHash:       HashADIdentifier(tenantID),
		AuthMethod:       authMethod,
		BindingSignature: signature,
		BindingMessage:   bindingMessage,
		CreatedAt:        createdAt,
		Active:           true,
	}
}

// Validate validates the AD wallet binding
func (b *ADWalletBinding) Validate() error {
	if b.BindingID == "" {
		return ErrInvalidADSSO.Wrap("binding_id cannot be empty")
	}
	if b.WalletAddress == "" {
		return ErrInvalidAddress.Wrap("wallet_address cannot be empty")
	}
	if b.LinkageID == "" {
		return ErrInvalidADSSO.Wrap("linkage_id cannot be empty")
	}
	if b.SubjectHash == "" {
		return ErrInvalidADSSO.Wrap("subject_hash cannot be empty")
	}
	if len(b.SubjectHash) != 64 {
		return ErrInvalidADSSO.Wrap("subject_hash must be a valid SHA256 hex string")
	}
	if b.TenantHash == "" {
		return ErrInvalidADSSO.Wrap("tenant_hash cannot be empty")
	}
	if len(b.TenantHash) != 64 {
		return ErrInvalidADSSO.Wrap("tenant_hash must be a valid SHA256 hex string")
	}
	if !IsValidADAuthMethod(b.AuthMethod) {
		return ErrInvalidADSSO.Wrapf("invalid auth_method: %s", b.AuthMethod)
	}
	if len(b.BindingSignature) == 0 {
		return ErrInvalidBindingSignature.Wrap("binding_signature cannot be empty")
	}
	if b.BindingMessage == "" {
		return ErrInvalidADSSO.Wrap("binding_message cannot be empty")
	}
	if b.CreatedAt.IsZero() {
		return ErrInvalidADSSO.Wrap("created_at cannot be zero")
	}
	return nil
}

// IsActive returns true if the binding is currently active
func (b *ADWalletBinding) IsActive() bool {
	if !b.Active {
		return false
	}
	if b.RevokedAt != nil {
		return false
	}
	if b.ExpiresAt != nil && time.Now().After(*b.ExpiresAt) {
		return false
	}
	return true
}

// Revoke marks the binding as revoked
func (b *ADWalletBinding) Revoke() {
	now := time.Now()
	b.RevokedAt = &now
	b.Active = false
}

// ============================================================================
// AD SSO Scoring
// ============================================================================

// ADSSOScoringWeight defines the weight of AD SSO verification in VEID scoring
type ADSSOScoringWeight struct {
	// AuthMethod is the authentication method
	AuthMethod ADAuthMethod `json:"auth_method"`

	// Weight is the score weight in basis points (out of 10000)
	Weight uint32 `json:"weight"`

	// RequireGroupMembership if true, requires at least one group hash
	RequireGroupMembership bool `json:"require_group_membership,omitempty"`

	// MinGroupCount is the minimum number of groups for full weight
	MinGroupCount int `json:"min_group_count,omitempty"`
}

// DefaultADSSOScoringWeights returns default scoring weights for AD SSO methods
func DefaultADSSOScoringWeights() []ADSSOScoringWeight {
	return []ADSSOScoringWeight{
		{AuthMethod: ADAuthMethodOIDC, Weight: 400}, // 4.0% weight (Azure AD cloud)
		{AuthMethod: ADAuthMethodSAML, Weight: 350}, // 3.5% weight (SAML federation)
		{AuthMethod: ADAuthMethodLDAP, Weight: 300}, // 3.0% weight (on-premises)
	}
}

// GetADSSOScoringWeight returns the scoring weight for an auth method
func GetADSSOScoringWeight(method ADAuthMethod) uint32 {
	weights := DefaultADSSOScoringWeights()
	for _, w := range weights {
		if w.AuthMethod == method {
			return w.Weight
		}
	}
	return 0
}

// ============================================================================
// AD SSO Verification Result
// ============================================================================

// ADSSOVerificationResult represents the result of an AD SSO verification
type ADSSOVerificationResult struct {
	// ChallengeID is the challenge that was verified
	ChallengeID string `json:"challenge_id"`

	// LinkageID is the resulting linkage ID
	LinkageID string `json:"linkage_id"`

	// Success indicates if verification was successful
	Success bool `json:"success"`

	// AuthMethod is the authentication method used
	AuthMethod ADAuthMethod `json:"auth_method"`

	// VerifiedAt is when verification completed
	VerifiedAt time.Time `json:"verified_at"`

	// SubjectHash is the verified subject hash
	SubjectHash string `json:"subject_hash"`

	// TenantHash is the verified tenant hash
	TenantHash string `json:"tenant_hash"`

	// ErrorCode is the error code if verification failed
	ErrorCode string `json:"error_code,omitempty"`

	// ErrorMessage is the error message if verification failed
	ErrorMessage string `json:"error_message,omitempty"`

	// ScoreContribution is the calculated score contribution
	ScoreContribution uint32 `json:"score_contribution"`
}

// NewADSSOVerificationResult creates a new successful verification result
func NewADSSOVerificationResult(
	challengeID string,
	linkageID string,
	authMethod ADAuthMethod,
	subjectID string,
	tenantID string,
	verifiedAt time.Time,
) *ADSSOVerificationResult {
	return &ADSSOVerificationResult{
		ChallengeID:       challengeID,
		LinkageID:         linkageID,
		Success:           true,
		AuthMethod:        authMethod,
		VerifiedAt:        verifiedAt,
		SubjectHash:       HashADIdentifier(subjectID),
		TenantHash:        HashADIdentifier(tenantID),
		ScoreContribution: GetADSSOScoringWeight(authMethod),
	}
}

// NewADSSOVerificationFailure creates a failed verification result
func NewADSSOVerificationFailure(
	challengeID string,
	authMethod ADAuthMethod,
	errorCode string,
	errorMessage string,
) *ADSSOVerificationResult {
	return &ADSSOVerificationResult{
		ChallengeID:  challengeID,
		Success:      false,
		AuthMethod:   authMethod,
		VerifiedAt:   time.Now(),
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}
}
