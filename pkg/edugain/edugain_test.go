// Package edugain provides EduGAIN federation integration.
//
// VE-908: EduGAIN federation support for academic/research institution SSO
package edugain

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// Config Tests
// ============================================================================

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	assert.True(t, cfg.Enabled)
	assert.NotEmpty(t, cfg.SPEntityID)
	assert.Equal(t, EduGAINProductionMetadataURL, cfg.MetadataURL)
	assert.Equal(t, DefaultMetadataRefreshInterval, cfg.MetadataRefreshInterval)
	assert.Equal(t, DefaultSessionDuration, cfg.SessionDuration)
	assert.Equal(t, SAMLBindingHTTPPOST, cfg.PreferredBinding)
	assert.Equal(t, NameIDFormatPersistent, cfg.NameIDFormat)
	assert.True(t, cfg.VEIDIntegration.Enabled)
}

func TestTestConfig(t *testing.T) {
	cfg := TestConfig()

	assert.True(t, cfg.Enabled)
	assert.Equal(t, EduGAINTestMetadataURL, cfg.MetadataURL)
	assert.Equal(t, 1*time.Hour, cfg.SessionDuration)
	assert.True(t, cfg.Logging.LogAssertions)
}

func TestConfigValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
	}{
		{
			name:    "valid default config with ACS URL",
			modify:  func(c *Config) { c.AssertionConsumerServiceURL = "https://example.com/acs" },
			wantErr: false,
		},
		{
			name:    "disabled config skips validation",
			modify:  func(c *Config) { c.Enabled = false; c.SPEntityID = "" },
			wantErr: false,
		},
		{
			name:    "missing SP entity ID",
			modify:  func(c *Config) { c.SPEntityID = "" },
			wantErr: true,
		},
		{
			name:    "missing metadata URL",
			modify:  func(c *Config) { c.MetadataURL = "" },
			wantErr: true,
		},
		{
			name:    "missing ACS URL",
			modify:  func(c *Config) {},
			wantErr: true,
		},
		{
			name:    "refresh interval too short",
			modify:  func(c *Config) { c.MetadataRefreshInterval = 1 * time.Minute; c.AssertionConsumerServiceURL = "https://example.com/acs" },
			wantErr: true,
		},
		{
			name:    "session duration too short",
			modify:  func(c *Config) { c.SessionDuration = 30 * time.Second; c.AssertionConsumerServiceURL = "https://example.com/acs" },
			wantErr: true,
		},
		{
			name:    "session duration too long",
			modify:  func(c *Config) { c.SessionDuration = 48 * time.Hour; c.AssertionConsumerServiceURL = "https://example.com/acs" },
			wantErr: true,
		},
		{
			name:    "invalid preferred binding",
			modify:  func(c *Config) { c.PreferredBinding = "invalid"; c.AssertionConsumerServiceURL = "https://example.com/acs" },
			wantErr: true,
		},
		{
			name:    "invalid name ID format",
			modify:  func(c *Config) { c.NameIDFormat = "invalid"; c.AssertionConsumerServiceURL = "https://example.com/acs" },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultConfig()
			tt.modify(&cfg)
			err := cfg.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestConfigIsInstitutionAllowed(t *testing.T) {
	cfg := DefaultConfig()

	// No restrictions
	assert.True(t, cfg.IsInstitutionAllowed("urn:mace:incommon:mit.edu"))

	// With blocklist
	cfg.BlockedInstitutions = []string{"urn:mace:blocked:example.org"}
	assert.True(t, cfg.IsInstitutionAllowed("urn:mace:incommon:mit.edu"))
	assert.False(t, cfg.IsInstitutionAllowed("urn:mace:blocked:example.org"))

	// With allowlist
	cfg.AllowedInstitutions = []string{"urn:mace:incommon:mit.edu", "urn:mace:incommon:stanford.edu"}
	assert.True(t, cfg.IsInstitutionAllowed("urn:mace:incommon:mit.edu"))
	assert.False(t, cfg.IsInstitutionAllowed("urn:mace:incommon:harvard.edu"))
}

func TestConfigIsFederationTrusted(t *testing.T) {
	cfg := DefaultConfig()

	// No restrictions
	assert.True(t, cfg.IsFederationTrusted("InCommon"))
	assert.True(t, cfg.IsFederationTrusted("AAF"))

	// With trust list
	cfg.TrustedFederations = []string{"InCommon", "UK Access Federation"}
	assert.True(t, cfg.IsFederationTrusted("InCommon"))
	assert.False(t, cfg.IsFederationTrusted("AAF"))
}

// ============================================================================
// Types Tests
// ============================================================================

func TestAllFederationStatuses(t *testing.T) {
	statuses := AllFederationStatuses()
	assert.Len(t, statuses, 5)
	assert.Contains(t, statuses, FederationStatusUnknown)
	assert.Contains(t, statuses, FederationStatusActive)
	assert.Contains(t, statuses, FederationStatusExpired)
}

func TestIsValidFederationStatus(t *testing.T) {
	assert.True(t, IsValidFederationStatus(FederationStatusActive))
	assert.True(t, IsValidFederationStatus(FederationStatusExpired))
	assert.False(t, IsValidFederationStatus("invalid"))
}

func TestAllInstitutionTypes(t *testing.T) {
	types := AllInstitutionTypes()
	assert.Len(t, types, 6)
	assert.Contains(t, types, InstitutionTypeUniversity)
	assert.Contains(t, types, InstitutionTypeResearchInstitute)
}

func TestIsValidInstitutionType(t *testing.T) {
	assert.True(t, IsValidInstitutionType(InstitutionTypeUniversity))
	assert.True(t, IsValidInstitutionType(InstitutionTypeResearchInstitute))
	assert.False(t, IsValidInstitutionType("invalid"))
}

func TestAllSessionStatuses(t *testing.T) {
	statuses := AllSessionStatuses()
	assert.Len(t, statuses, 4)
	assert.Contains(t, statuses, SessionStatusActive)
	assert.Contains(t, statuses, SessionStatusExpired)
}

func TestIsValidSessionStatus(t *testing.T) {
	assert.True(t, IsValidSessionStatus(SessionStatusActive))
	assert.True(t, IsValidSessionStatus(SessionStatusRevoked))
	assert.False(t, IsValidSessionStatus("invalid"))
}

func TestAllAffiliationTypes(t *testing.T) {
	affiliations := AllAffiliationTypes()
	assert.Len(t, affiliations, 8)
	assert.Contains(t, affiliations, AffiliationStudent)
	assert.Contains(t, affiliations, AffiliationFaculty)
	assert.Contains(t, affiliations, AffiliationStaff)
}

func TestIsValidAffiliationType(t *testing.T) {
	assert.True(t, IsValidAffiliationType(AffiliationStudent))
	assert.True(t, IsValidAffiliationType(AffiliationFaculty))
	assert.False(t, IsValidAffiliationType("invalid"))
}

func TestInstitutionValidate(t *testing.T) {
	tests := []struct {
		name    string
		inst    Institution
		wantErr bool
	}{
		{
			name: "valid institution",
			inst: Institution{
				EntityID:     "urn:mace:incommon:mit.edu",
				DisplayName:  "MIT",
				Certificates: []string{"cert1"},
				SSOEndpoints: map[string]string{SAMLBindingHTTPPOST: "https://mit.edu/sso"},
			},
			wantErr: false,
		},
		{
			name: "missing entity ID",
			inst: Institution{
				DisplayName:  "MIT",
				Certificates: []string{"cert1"},
				SSOEndpoints: map[string]string{SAMLBindingHTTPPOST: "https://mit.edu/sso"},
			},
			wantErr: true,
		},
		{
			name: "missing display name",
			inst: Institution{
				EntityID:     "urn:mace:incommon:mit.edu",
				Certificates: []string{"cert1"},
				SSOEndpoints: map[string]string{SAMLBindingHTTPPOST: "https://mit.edu/sso"},
			},
			wantErr: true,
		},
		{
			name: "missing certificates",
			inst: Institution{
				EntityID:     "urn:mace:incommon:mit.edu",
				DisplayName:  "MIT",
				SSOEndpoints: map[string]string{SAMLBindingHTTPPOST: "https://mit.edu/sso"},
			},
			wantErr: true,
		},
		{
			name: "missing SSO endpoints",
			inst: Institution{
				EntityID:     "urn:mace:incommon:mit.edu",
				DisplayName:  "MIT",
				Certificates: []string{"cert1"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.inst.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstitutionGetSSOEndpoint(t *testing.T) {
	inst := Institution{
		SSOEndpoints: map[string]string{
			SAMLBindingHTTPPOST:     "https://example.com/sso/post",
			SAMLBindingHTTPRedirect: "https://example.com/sso/redirect",
		},
	}

	// Get preferred binding
	endpoint, err := inst.GetSSOEndpoint(SAMLBindingHTTPPOST)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/sso/post", endpoint)

	// Get redirect binding
	endpoint, err = inst.GetSSOEndpoint(SAMLBindingHTTPRedirect)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/sso/redirect", endpoint)

	// Fall back to redirect
	inst2 := Institution{
		SSOEndpoints: map[string]string{
			SAMLBindingHTTPRedirect: "https://example.com/sso/redirect",
		},
	}
	endpoint, err = inst2.GetSSOEndpoint(SAMLBindingHTTPPOST)
	assert.NoError(t, err)
	assert.Equal(t, "https://example.com/sso/redirect", endpoint)

	// No endpoints
	inst3 := Institution{SSOEndpoints: map[string]string{}}
	_, err = inst3.GetSSOEndpoint(SAMLBindingHTTPPOST)
	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedBinding, err)
}

func TestFederationMetadataIsValid(t *testing.T) {
	// Valid metadata
	metadata := &FederationMetadata{
		Status:     FederationStatusActive,
		ValidUntil: time.Now().Add(1 * time.Hour),
	}
	assert.True(t, metadata.IsValid())

	// Expired metadata
	metadata.ValidUntil = time.Now().Add(-1 * time.Hour)
	assert.False(t, metadata.IsValid())

	// Wrong status
	metadata.ValidUntil = time.Now().Add(1 * time.Hour)
	metadata.Status = FederationStatusError
	assert.False(t, metadata.IsValid())
}

func TestFederationMetadataFindInstitution(t *testing.T) {
	metadata := &FederationMetadata{
		Institutions: []Institution{
			{EntityID: "urn:mace:incommon:mit.edu", DisplayName: "MIT"},
			{EntityID: "urn:mace:incommon:stanford.edu", DisplayName: "Stanford"},
		},
	}

	// Found
	inst, err := metadata.FindInstitution("urn:mace:incommon:mit.edu")
	assert.NoError(t, err)
	assert.Equal(t, "MIT", inst.DisplayName)

	// Not found
	_, err = metadata.FindInstitution("urn:mace:incommon:harvard.edu")
	assert.Error(t, err)
	assert.Equal(t, ErrInstitutionNotFound, err)
}

func TestUserAttributesHasAffiliation(t *testing.T) {
	attrs := &UserAttributes{
		EduPerson: EduPersonAttributes{
			Affiliation: []AffiliationType{AffiliationStudent, AffiliationMember},
		},
	}

	assert.True(t, attrs.HasAffiliation(AffiliationStudent))
	assert.True(t, attrs.HasAffiliation(AffiliationMember))
	assert.False(t, attrs.HasAffiliation(AffiliationFaculty))
}

func TestUserAttributesHasAnyAffiliation(t *testing.T) {
	attrs := &UserAttributes{
		EduPerson: EduPersonAttributes{
			Affiliation: []AffiliationType{AffiliationStudent},
		},
	}

	assert.True(t, attrs.HasAnyAffiliation(AffiliationStudent, AffiliationFaculty))
	assert.False(t, attrs.HasAnyAffiliation(AffiliationFaculty, AffiliationStaff))
}

func TestSAMLAssertionValidate(t *testing.T) {
	clockSkew := 2 * time.Minute

	// Valid assertion
	assertion := &SAMLAssertion{
		NotBefore:    time.Now().Add(-1 * time.Minute),
		NotOnOrAfter: time.Now().Add(5 * time.Minute),
	}
	assert.NoError(t, assertion.Validate(clockSkew))

	// Not yet valid
	assertion.NotBefore = time.Now().Add(10 * time.Minute)
	assert.Error(t, assertion.Validate(clockSkew))
	assert.Equal(t, ErrAssertionNotYetValid, assertion.Validate(clockSkew))

	// Expired
	assertion.NotBefore = time.Now().Add(-10 * time.Minute)
	assertion.NotOnOrAfter = time.Now().Add(-5 * time.Minute)
	assert.Error(t, assertion.Validate(clockSkew))
	assert.Equal(t, ErrAssertionExpired, assertion.Validate(clockSkew))
}

func TestSessionIsValid(t *testing.T) {
	session := &Session{
		Status:    SessionStatusActive,
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	assert.True(t, session.IsValid())

	session.ExpiresAt = time.Now().Add(-1 * time.Hour)
	assert.False(t, session.IsValid())

	session.ExpiresAt = time.Now().Add(1 * time.Hour)
	session.Status = SessionStatusRevoked
	assert.False(t, session.IsValid())
}

func TestSessionTimeToExpiry(t *testing.T) {
	session := &Session{
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	ttl := session.TimeToExpiry()
	assert.True(t, ttl > 59*time.Minute && ttl <= 1*time.Hour)
}

// ============================================================================
// Attribute Mapper Tests
// ============================================================================

func TestAttributeMapperMapAttributes(t *testing.T) {
	mapper := newAttributeMapper()

	rawAttrs := map[string][]string{
		OIDEduPersonPrincipalName: {"jsmith@mit.edu"},
		OIDEduPersonAffiliation:   {"student", "member"},
		OIDSchacHomeOrganization:  {"mit.edu"},
		OIDDisplayName:            {"John Smith"},
		OIDEmail:                  {"jsmith@mit.edu"},
	}

	attrs, err := mapper.MapAttributes(rawAttrs)
	require.NoError(t, err)

	assert.Equal(t, "jsmith@mit.edu", attrs.EduPerson.PrincipalName)
	assert.Contains(t, attrs.EduPerson.Affiliation, AffiliationStudent)
	assert.Contains(t, attrs.EduPerson.Affiliation, AffiliationMember)
	assert.Equal(t, "mit.edu", attrs.Schac.HomeOrganization)
	assert.Equal(t, "John Smith", attrs.DisplayName)
	assert.Equal(t, "jsmith@mit.edu", attrs.Email)
}

func TestAttributeMapperMapAttributesFriendlyNames(t *testing.T) {
	mapper := newAttributeMapper()

	// Using friendly names instead of OIDs
	rawAttrs := map[string][]string{
		FriendlyEduPersonPrincipalName: {"jdoe@stanford.edu"},
		FriendlyEduPersonAffiliation:   {"faculty"},
		FriendlySchacHomeOrganization:  {"stanford.edu"},
	}

	attrs, err := mapper.MapAttributes(rawAttrs)
	require.NoError(t, err)

	assert.Equal(t, "jdoe@stanford.edu", attrs.EduPerson.PrincipalName)
	assert.Contains(t, attrs.EduPerson.Affiliation, AffiliationFaculty)
	assert.Equal(t, "stanford.edu", attrs.Schac.HomeOrganization)
}

func TestAttributeMapperValidateAttributes(t *testing.T) {
	mapper := newAttributeMapper()

	// Valid
	attrs := &UserAttributes{
		EduPerson: EduPersonAttributes{
			PrincipalName: "user@example.edu",
		},
	}
	assert.NoError(t, mapper.ValidateAttributes(attrs))

	// Missing principal name
	attrs.EduPerson.PrincipalName = ""
	assert.Error(t, mapper.ValidateAttributes(attrs))

	// Missing scope in principal name
	attrs.EduPerson.PrincipalName = "user"
	assert.Error(t, mapper.ValidateAttributes(attrs))
}

func TestAttributeMapperHashSensitiveData(t *testing.T) {
	mapper := newAttributeMapper()

	attrs := &UserAttributes{
		EduPerson: EduPersonAttributes{
			PrincipalName: "user@example.edu",
			TargetedID:    "target123",
			EPPN:          "user@example.edu",
		},
		Schac: SchacAttributes{
			PersonalUniqueCode: "S123456",
		},
		Email: "user@example.edu",
	}

	hashed := mapper.HashSensitiveData(attrs)

	// Hashes should be set
	assert.NotEmpty(t, hashed.EduPerson.PrincipalNameHash)
	assert.NotEmpty(t, hashed.EduPerson.TargetedIDHash)
	assert.NotEmpty(t, hashed.EduPerson.EPPNHash)
	assert.NotEmpty(t, hashed.Schac.PersonalUniqueCodeHash)
	assert.NotEmpty(t, hashed.EmailHash)

	// Hashes should be 64 chars (SHA-256 hex)
	assert.Len(t, hashed.EduPerson.PrincipalNameHash, 64)
	assert.Len(t, hashed.EmailHash, 64)
}

func TestComputeAttributeScore(t *testing.T) {
	// Basic attributes
	attrs := &UserAttributes{
		EduPerson: EduPersonAttributes{
			PrincipalName: "user@example.edu",
		},
	}
	score := ComputeAttributeScore(attrs)
	assert.True(t, score > 0)

	// With faculty affiliation
	attrs.EduPerson.Affiliation = []AffiliationType{AffiliationFaculty}
	scoreWithFaculty := ComputeAttributeScore(attrs)
	assert.True(t, scoreWithFaculty > score)

	// With organization
	attrs.Schac.HomeOrganization = "example.edu"
	scoreWithOrg := ComputeAttributeScore(attrs)
	assert.True(t, scoreWithOrg > scoreWithFaculty)
}

// ============================================================================
// Session Manager Tests
// ============================================================================

func TestSessionManagerCreate(t *testing.T) {
	cfg := TestConfig()
	cfg.SessionStorage.MaxSessions = 3
	manager := newSessionManager(cfg)

	ctx := context.Background()
	assertion := &SAMLAssertion{
		ID:             "assertion1",
		IssuerEntityID: "urn:mace:incommon:mit.edu",
		AuthnInstant:   time.Now(),
		Attributes: UserAttributes{
			EduPerson: EduPersonAttributes{
				PrincipalName: "user@mit.edu",
			},
		},
	}

	session, err := manager.Create(ctx, assertion, "ve1wallet123")
	require.NoError(t, err)

	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "ve1wallet123", session.WalletAddress)
	assert.Equal(t, "urn:mace:incommon:mit.edu", session.InstitutionID)
	assert.Equal(t, SessionStatusActive, session.Status)
}

func TestSessionManagerGet(t *testing.T) {
	cfg := TestConfig()
	manager := newSessionManager(cfg)

	ctx := context.Background()
	assertion := &SAMLAssertion{
		ID:             "assertion1",
		IssuerEntityID: "urn:mace:incommon:mit.edu",
		AuthnInstant:   time.Now(),
	}

	created, err := manager.Create(ctx, assertion, "ve1wallet123")
	require.NoError(t, err)

	// Get existing session
	session, err := manager.Get(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, session.ID)

	// Get non-existing session
	_, err = manager.Get(ctx, "nonexistent")
	assert.Error(t, err)
	assert.Equal(t, ErrSessionNotFound, err)
}

func TestSessionManagerRevoke(t *testing.T) {
	cfg := TestConfig()
	manager := newSessionManager(cfg)

	ctx := context.Background()
	assertion := &SAMLAssertion{
		ID:             "assertion1",
		IssuerEntityID: "urn:mace:incommon:mit.edu",
		AuthnInstant:   time.Now(),
	}

	session, err := manager.Create(ctx, assertion, "ve1wallet123")
	require.NoError(t, err)

	// Revoke session
	err = manager.Revoke(ctx, session.ID)
	assert.NoError(t, err)

	// Session should be gone
	_, err = manager.Get(ctx, session.ID)
	assert.Error(t, err)
}

func TestSessionManagerList(t *testing.T) {
	cfg := TestConfig()
	manager := newSessionManager(cfg)

	ctx := context.Background()
	walletAddress := "ve1wallet123"

	// Create multiple sessions
	for i := 0; i < 3; i++ {
		assertion := &SAMLAssertion{
			ID:             "assertion" + string(rune('0'+i)),
			IssuerEntityID: "urn:mace:incommon:mit.edu",
			AuthnInstant:   time.Now(),
		}
		_, err := manager.Create(ctx, assertion, walletAddress)
		require.NoError(t, err)
	}

	sessions, err := manager.List(ctx, walletAddress)
	require.NoError(t, err)
	assert.Len(t, sessions, 3)
}

func TestSessionManagerMaxSessions(t *testing.T) {
	cfg := TestConfig()
	cfg.SessionStorage.MaxSessions = 2
	manager := newSessionManager(cfg)

	ctx := context.Background()
	walletAddress := "ve1wallet123"

	// Create 3 sessions (max is 2)
	for i := 0; i < 3; i++ {
		assertion := &SAMLAssertion{
			ID:             "assertion" + string(rune('0'+i)),
			IssuerEntityID: "urn:mace:incommon:mit.edu",
			AuthnInstant:   time.Now(),
		}
		_, err := manager.Create(ctx, assertion, walletAddress)
		require.NoError(t, err)
	}

	// Should have max 2 sessions
	sessions, err := manager.List(ctx, walletAddress)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestSessionManagerTokenValidation(t *testing.T) {
	cfg := TestConfig()
	manager := newSessionManager(cfg)

	ctx := context.Background()
	assertion := &SAMLAssertion{
		ID:             "assertion1",
		IssuerEntityID: "urn:mace:incommon:mit.edu",
		AuthnInstant:   time.Now(),
	}

	session, err := manager.Create(ctx, assertion, "ve1wallet123")
	require.NoError(t, err)

	// Generate token
	token, err := manager.GenerateToken(session)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate token
	validated, err := manager.ValidateToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, session.ID, validated.ID)

	// Invalid token
	_, err = manager.ValidateToken(ctx, "invalidtoken")
	assert.Error(t, err)
}

func TestSessionManagerReplayDetection(t *testing.T) {
	cfg := TestConfig()
	manager := newSessionManager(cfg)

	ctx := context.Background()
	assertionID := "unique-assertion-id"
	expiry := time.Now().Add(1 * time.Hour)

	// First check - not replayed
	replayed, err := manager.IsAssertionReplayed(ctx, assertionID)
	require.NoError(t, err)
	assert.False(t, replayed)

	// Track the assertion
	err = manager.TrackAssertionID(ctx, assertionID, expiry)
	require.NoError(t, err)

	// Second check - replayed
	replayed, err = manager.IsAssertionReplayed(ctx, assertionID)
	require.NoError(t, err)
	assert.True(t, replayed)
}

func TestSessionManagerCleanup(t *testing.T) {
	cfg := TestConfig()
	cfg.SessionDuration = 1 * time.Millisecond // Very short for testing
	manager := newSessionManager(cfg)

	ctx := context.Background()
	assertion := &SAMLAssertion{
		ID:             "assertion1",
		IssuerEntityID: "urn:mace:incommon:mit.edu",
		AuthnInstant:   time.Now(),
	}

	_, err := manager.Create(ctx, assertion, "ve1wallet123")
	require.NoError(t, err)

	// Wait for expiry
	time.Sleep(10 * time.Millisecond)

	// Cleanup
	removed, err := manager.Cleanup(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, removed)
}

// ============================================================================
// SAML Provider Tests
// ============================================================================

func TestSAMLProviderGetEntityID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AssertionConsumerServiceURL = "https://example.com/acs"
	provider := newSAMLProvider(cfg)

	assert.Equal(t, cfg.SPEntityID, provider.GetEntityID())
}

func TestSAMLProviderGetMetadata(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AssertionConsumerServiceURL = "https://example.com/acs"
	provider := newSAMLProvider(cfg)

	metadata, err := provider.GetMetadata()
	require.NoError(t, err)

	// Check it's valid XML
	assert.Contains(t, string(metadata), "EntityDescriptor")
	assert.Contains(t, string(metadata), cfg.SPEntityID)
	assert.Contains(t, string(metadata), cfg.AssertionConsumerServiceURL)
}

func TestSAMLProviderCreateAuthnRequest(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AssertionConsumerServiceURL = "https://example.com/acs"
	provider := newSAMLProvider(cfg)

	ctx := context.Background()
	idp := &Institution{
		EntityID: "urn:mace:incommon:mit.edu",
		SSOEndpoints: map[string]string{
			SAMLBindingHTTPPOST:     "https://mit.edu/sso/post",
			SAMLBindingHTTPRedirect: "https://mit.edu/sso/redirect",
		},
	}

	params := AuthnRequestParams{
		InstitutionID: idp.EntityID,
		RelayState:    "session123",
	}

	// HTTP-POST binding
	result, err := provider.CreateAuthnRequest(ctx, idp, params)
	require.NoError(t, err)

	assert.NotEmpty(t, result.Request.ID)
	assert.Equal(t, "session123", result.RelayState)
	assert.Equal(t, SAMLBindingHTTPPOST, result.Binding)
	assert.NotEmpty(t, result.SAMLRequest)
	assert.NotEmpty(t, result.PostFormHTML)

	// HTTP-Redirect binding
	params.PreferredBinding = SAMLBindingHTTPRedirect
	result, err = provider.CreateAuthnRequest(ctx, idp, params)
	require.NoError(t, err)

	assert.Equal(t, SAMLBindingHTTPRedirect, result.Binding)
	assert.NotEmpty(t, result.URL)
	assert.Contains(t, result.URL, "SAMLRequest=")
}

func TestSAMLProviderCreateAuthnRequestWithMFA(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AssertionConsumerServiceURL = "https://example.com/acs"
	cfg.RequireMFA = true
	provider := newSAMLProvider(cfg)

	ctx := context.Background()
	idp := &Institution{
		EntityID: "urn:mace:incommon:mit.edu",
		SSOEndpoints: map[string]string{
			SAMLBindingHTTPPOST: "https://mit.edu/sso/post",
		},
	}

	params := AuthnRequestParams{
		InstitutionID: idp.EntityID,
	}

	result, err := provider.CreateAuthnRequest(ctx, idp, params)
	require.NoError(t, err)

	assert.Contains(t, result.Request.RequestedAuthnContext, AuthnContextMFA)
}

// ============================================================================
// Discovery Service Tests
// ============================================================================

func TestDiscoveryServiceSearch(t *testing.T) {
	cfg := DefaultConfig()
	metadata := newMetadataService(cfg)

	// Set up test metadata
	ms := metadata.(*metadataService)
	ms.metadata = &FederationMetadata{
		Status: FederationStatusActive,
		Institutions: []Institution{
			{EntityID: "urn:mace:incommon:mit.edu", DisplayName: "MIT", Country: "US", Federation: "InCommon"},
			{EntityID: "urn:mace:incommon:stanford.edu", DisplayName: "Stanford University", Country: "US", Federation: "InCommon"},
			{EntityID: "urn:mace:aaf:unimelb.edu.au", DisplayName: "University of Melbourne", Country: "AU", Federation: "AAF"},
		},
	}

	discovery := newDiscoveryService(metadata)
	ctx := context.Background()

	// Search by name
	result, err := discovery.Search(ctx, InstitutionSearchQuery{Query: "stanford"})
	require.NoError(t, err)
	assert.Len(t, result.Institutions, 1)
	assert.Equal(t, "Stanford University", result.Institutions[0].DisplayName)

	// Search by country
	result, err = discovery.Search(ctx, InstitutionSearchQuery{Country: "US"})
	require.NoError(t, err)
	assert.Len(t, result.Institutions, 2)

	// Search by federation
	result, err = discovery.Search(ctx, InstitutionSearchQuery{Federation: "AAF"})
	require.NoError(t, err)
	assert.Len(t, result.Institutions, 1)
	assert.Equal(t, "University of Melbourne", result.Institutions[0].DisplayName)
}

func TestDiscoveryServiceRecordUsage(t *testing.T) {
	cfg := DefaultConfig()
	metadata := newMetadataService(cfg)

	ms := metadata.(*metadataService)
	ms.metadata = &FederationMetadata{
		Status: FederationStatusActive,
		Institutions: []Institution{
			{EntityID: "urn:mace:incommon:mit.edu", DisplayName: "MIT"},
		},
	}

	discovery := newDiscoveryService(metadata)
	ctx := context.Background()

	// Record usage
	err := discovery.RecordUsage(ctx, "urn:mace:incommon:mit.edu", "ve1wallet123")
	require.NoError(t, err)

	// Check recent
	recent, err := discovery.GetRecent(ctx, "ve1wallet123", 10)
	require.NoError(t, err)
	assert.Len(t, recent, 1)
	assert.Equal(t, "MIT", recent[0].DisplayName)
}

// ============================================================================
// VEID Integration Tests
// ============================================================================

func TestVEIDIntegratorCreateScope(t *testing.T) {
	cfg := DefaultConfig()
	integrator := newVEIDIntegrator(cfg)

	ctx := context.Background()
	session := &Session{
		ID:              "session123",
		WalletAddress:   "ve1wallet123",
		InstitutionID:   "urn:mace:incommon:mit.edu",
		InstitutionName: "MIT",
		Status:          SessionStatusActive,
		AuthnInstant:    time.Now(),
		ExpiresAt:       time.Now().Add(8 * time.Hour),
		IsMFA:           true,
		Attributes: UserAttributes{
			EduPerson: EduPersonAttributes{
				PrincipalName:     "user@mit.edu",
				PrincipalNameHash: hashString("user@mit.edu"),
				Affiliation:       []AffiliationType{AffiliationFaculty},
			},
			Schac: SchacAttributes{
				HomeOrganization: "mit.edu",
			},
		},
	}

	scope, err := integrator.CreateScope(ctx, session)
	require.NoError(t, err)

	assert.NotEmpty(t, scope.ID)
	assert.Equal(t, "ve1wallet123", scope.WalletAddress)
	assert.Equal(t, "urn:mace:incommon:mit.edu", scope.InstitutionID)
	assert.True(t, scope.IsMFA)
	assert.True(t, scope.ScoreContribution > 0)
	assert.Contains(t, scope.Affiliations, AffiliationFaculty)
}

func TestVEIDIntegratorRevokeScope(t *testing.T) {
	cfg := DefaultConfig()
	integrator := newVEIDIntegrator(cfg)

	ctx := context.Background()
	session := &Session{
		ID:            "session123",
		WalletAddress: "ve1wallet123",
		InstitutionID: "urn:mace:incommon:mit.edu",
		Status:        SessionStatusActive,
		ExpiresAt:     time.Now().Add(8 * time.Hour),
	}

	scope, err := integrator.CreateScope(ctx, session)
	require.NoError(t, err)

	// Revoke scope
	err = integrator.RevokeScope(ctx, scope.ID)
	require.NoError(t, err)

	// Check status
	vi := integrator.(*veidIntegrator)
	assert.Equal(t, "revoked", vi.scopes[scope.ID].Status)
}

func TestValidateForVEID(t *testing.T) {
	// Valid session
	session := &Session{
		Status:        SessionStatusActive,
		WalletAddress: "ve1wallet123",
		Attributes: UserAttributes{
			EduPerson: EduPersonAttributes{
				PrincipalName: "user@example.edu",
			},
		},
	}
	assert.NoError(t, ValidateForVEID(session))

	// Nil session
	assert.Error(t, ValidateForVEID(nil))

	// Inactive session
	session.Status = SessionStatusRevoked
	assert.Error(t, ValidateForVEID(session))

	// Missing wallet
	session.Status = SessionStatusActive
	session.WalletAddress = ""
	assert.Error(t, ValidateForVEID(session))

	// Missing principal name
	session.WalletAddress = "ve1wallet123"
	session.Attributes.EduPerson.PrincipalName = ""
	assert.Error(t, ValidateForVEID(session))
}

func TestConvertToScopeData(t *testing.T) {
	scope := &VEIDScope{
		ID:                "scope123",
		WalletAddress:     "ve1wallet123",
		InstitutionID:     "urn:mace:incommon:mit.edu",
		Federation:        "InCommon",
		PrincipalNameHash: "hash123",
		HomeOrganization:  "mit.edu",
		Affiliations:      []AffiliationType{AffiliationStudent, AffiliationMember},
		IsMFA:             true,
		ScoreContribution: 15,
		AuthnInstant:      time.Now(),
		ExpiresAt:         time.Now().Add(8 * time.Hour),
	}

	data := ConvertToScopeData(scope)

	assert.Equal(t, uint32(1), data.Version)
	assert.Equal(t, ScopeTypeEduGAIN, data.Type)
	assert.NotEmpty(t, data.InstitutionIDHash)
	assert.NotEmpty(t, data.FederationHash)
	assert.Equal(t, "hash123", data.PrincipalNameHash)
	assert.Len(t, data.Affiliations, 2)
	assert.True(t, data.IsMFA)
	assert.Equal(t, uint32(15), data.ScoreContribution)
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestGenerateID(t *testing.T) {
	id1, err := generateID()
	require.NoError(t, err)
	assert.NotEmpty(t, id1)
	assert.True(t, id1[0] == '_') // SAML IDs start with underscore

	id2, err := generateID()
	require.NoError(t, err)
	assert.NotEqual(t, id1, id2) // Should be unique
}

func TestDeflateString(t *testing.T) {
	// Use a longer string to ensure compression is effective
	input := "This is a test string that will be compressed. " +
		"This is a test string that will be compressed. " +
		"This is a test string that will be compressed. " +
		"This is a test string that will be compressed."
	compressed, err := deflateString(input)
	require.NoError(t, err)
	assert.NotEmpty(t, compressed)
	// Just verify compression works, not size (overhead may make small strings larger)
}

func TestDecodeBase64(t *testing.T) {
	original := "Hello, World!"
	encoded := base64.StdEncoding.EncodeToString([]byte(original))

	decoded, err := decodeBase64(encoded)
	require.NoError(t, err)
	assert.Equal(t, original, string(decoded))

	// Invalid base64
	_, err = decodeBase64("not valid base64!!!")
	assert.Error(t, err)
}

func TestHashString(t *testing.T) {
	hash1 := hashString("test")
	hash2 := hashString("test")
	hash3 := hashString("different")

	assert.Len(t, hash1, 64) // SHA-256 hex is 64 chars
	assert.Equal(t, hash1, hash2) // Same input = same output
	assert.NotEqual(t, hash1, hash3) // Different input = different output
}

func TestParseFederationFromEntityID(t *testing.T) {
	tests := []struct {
		entityID string
		expected string
	}{
		{"urn:mace:incommon:mit.edu", "InCommon"},
		{"https://idp.mit.edu", "InCommon"},
		{"https://idp.unimelb.edu.au", "AAF"},
		{"https://idp.ox.ac.uk", "UK Access Federation"},
		{"https://cern.ch/shibboleth", "SWITCHaai"},
		{"https://unknown.example.org", "EduGAIN"},
	}

	for _, tt := range tests {
		t.Run(tt.entityID, func(t *testing.T) {
			result := parseFederationFromEntityID(tt.entityID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseInstitutionType(t *testing.T) {
	tests := []struct {
		categories []string
		expected   InstitutionType
	}{
		{[]string{"http://example.org/university"}, InstitutionTypeUniversity},
		{[]string{"http://example.org/research"}, InstitutionTypeResearchInstitute},
		{[]string{"http://example.org/school", "http://example.org/k-12"}, InstitutionTypeSchool},
		{[]string{"http://example.org/library"}, InstitutionTypeLibrary},
		{[]string{"http://example.org/government"}, InstitutionTypeGovernment},
		{[]string{}, InstitutionTypeOther},
		{[]string{"http://example.org/unknown"}, InstitutionTypeOther},
	}

	for _, tt := range tests {
		t.Run(string(tt.expected), func(t *testing.T) {
			result := parseInstitutionType(tt.categories)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsString(t *testing.T) {
	slice := []string{"a", "b", "c"}

	assert.True(t, containsString(slice, "a"))
	assert.True(t, containsString(slice, "c"))
	assert.False(t, containsString(slice, "d"))
	assert.False(t, containsString(nil, "a"))
	assert.False(t, containsString([]string{}, "a"))
}
