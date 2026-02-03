package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veid/types"
)

// authMethodInvalid is a test constant for invalid auth methods
const authMethodInvalid = "invalid"

// ============================================================================
// AD SSO Tests (VE-907: Active Directory SSO)
// ============================================================================

// TestAllADAuthMethods verifies all valid AD authentication methods
func TestAllADAuthMethods(t *testing.T) {
	methods := types.AllADAuthMethods()
	require.Len(t, methods, 3)
	assert.Contains(t, methods, types.ADAuthMethodOIDC)
	assert.Contains(t, methods, types.ADAuthMethodSAML)
	assert.Contains(t, methods, types.ADAuthMethodLDAP)
}

// TestIsValidADAuthMethod tests AD authentication method validation
func TestIsValidADAuthMethod(t *testing.T) {
	tests := []struct {
		name   string
		method types.ADAuthMethod
		want   bool
	}{
		{"OIDC is valid", types.ADAuthMethodOIDC, true},
		{"SAML is valid", types.ADAuthMethodSAML, true},
		{"LDAP is valid", types.ADAuthMethodLDAP, true},
		{"empty is invalid", types.ADAuthMethod(""), false},
		{"unknown is invalid", types.ADAuthMethod("kerberos"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.IsValidADAuthMethod(tt.method)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestAllADSSOStatuses verifies all valid AD SSO statuses
func TestAllADSSOStatuses(t *testing.T) {
	statuses := types.AllADSSOStatuses()
	require.Len(t, statuses, 5)
	assert.Contains(t, statuses, types.ADSSOStatusPending)
	assert.Contains(t, statuses, types.ADSSOStatusVerified)
	assert.Contains(t, statuses, types.ADSSOStatusFailed)
	assert.Contains(t, statuses, types.ADSSOStatusRevoked)
	assert.Contains(t, statuses, types.ADSSOStatusExpired)
}

// TestIsValidADSSOStatus tests AD SSO status validation
func TestIsValidADSSOStatus(t *testing.T) {
	tests := []struct {
		name   string
		status types.ADSSOStatus
		want   bool
	}{
		{"pending is valid", types.ADSSOStatusPending, true},
		{"verified is valid", types.ADSSOStatusVerified, true},
		{"failed is valid", types.ADSSOStatusFailed, true},
		{"revoked is valid", types.ADSSOStatusRevoked, true},
		{"expired is valid", types.ADSSOStatusExpired, true},
		{"empty is invalid", types.ADSSOStatus(""), false},
		{"unknown is invalid", types.ADSSOStatus("unknown"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.IsValidADSSOStatus(tt.status)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestHashADIdentifier tests AD identifier hashing
func TestHashADIdentifier(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		wantLen    int
	}{
		{"normal identifier", "user@contoso.com", 64},
		{"tenant ID", "12345678-1234-1234-1234-123456789abc", 64},
		{"empty returns empty", "", 0},
		{"special characters", "CN=John Doe,OU=Users,DC=contoso,DC=com", 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := types.HashADIdentifier(tt.identifier)
			assert.Len(t, got, tt.wantLen)
		})
	}

	// Verify determinism
	hash1 := types.HashADIdentifier("test@example.com")
	hash2 := types.HashADIdentifier("test@example.com")
	assert.Equal(t, hash1, hash2, "hash should be deterministic")

	// Verify different inputs produce different hashes
	hash3 := types.HashADIdentifier("other@example.com")
	assert.NotEqual(t, hash1, hash3, "different inputs should produce different hashes")
}

// ============================================================================
// Azure AD OIDC Config Tests
// ============================================================================

func TestAzureADOIDCConfig_NewAndValidate(t *testing.T) {
	config := types.NewAzureADOIDCConfig(
		"12345678-1234-1234-1234-123456789abc",
		"https://login.microsoftonline.com/12345678-1234-1234-1234-123456789abc/v2.0",
		"app-client-id-12345",
	)

	require.NotNil(t, config)
	assert.Len(t, config.TenantIDHash, 64)
	assert.Len(t, config.ClientIDHash, 64)
	assert.Equal(t, "https://login.microsoftonline.com/12345678-1234-1234-1234-123456789abc/v2.0", config.Issuer)

	err := config.Validate()
	require.NoError(t, err)
}

func TestAzureADOIDCConfig_Validate(t *testing.T) {
	now := time.Now()

	validConfig := types.NewAzureADOIDCConfig(
		"tenant-123",
		"https://login.microsoftonline.com/tenant-123/v2.0",
		"client-456",
	)

	tests := []struct {
		name    string
		config  *types.AzureADOIDCConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  validConfig,
			wantErr: false,
		},
		{
			name: "empty tenant_id_hash",
			config: &types.AzureADOIDCConfig{
				TenantIDHash: "",
				Issuer:       "https://login.microsoftonline.com/test",
				ClientIDHash: types.HashADIdentifier("client"),
			},
			wantErr: true,
		},
		{
			name: "invalid tenant_id_hash length",
			config: &types.AzureADOIDCConfig{
				TenantIDHash: "too-short",
				Issuer:       "https://login.microsoftonline.com/test",
				ClientIDHash: types.HashADIdentifier("client"),
			},
			wantErr: true,
		},
		{
			name: "empty issuer",
			config: &types.AzureADOIDCConfig{
				TenantIDHash: types.HashADIdentifier("tenant"),
				Issuer:       "",
				ClientIDHash: types.HashADIdentifier("client"),
			},
			wantErr: true,
		},
		{
			name: "empty client_id_hash",
			config: &types.AzureADOIDCConfig{
				TenantIDHash: types.HashADIdentifier("tenant"),
				Issuer:       "https://login.microsoftonline.com/test",
				ClientIDHash: "",
			},
			wantErr: true,
		},
	}

	_ = now // Silence unused variable

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// SAML Config Tests
// ============================================================================

func TestSAMLConfig_NewAndValidate(t *testing.T) {
	config := types.NewSAMLConfig(
		"https://idp.contoso.com/adfs",
		"https://idp.contoso.com/adfs/metadata",
		"AB:CD:EF:12:34:56:78:90",
		"urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress",
	)

	require.NotNil(t, config)
	assert.Len(t, config.EntityIDHash, 64)
	assert.Len(t, config.CertificateFingerprintHash, 64)
	assert.Equal(t, "https://idp.contoso.com/adfs/metadata", config.MetadataURL)
	assert.Equal(t, "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress", config.NameIDFormat)

	err := config.Validate()
	require.NoError(t, err)
}

func TestSAMLConfig_Validate(t *testing.T) {
	validConfig := types.NewSAMLConfig(
		"https://idp.example.com",
		"https://idp.example.com/metadata",
		"fingerprint123",
		"urn:oasis:names:tc:SAML:2.0:nameid-format:persistent",
	)

	tests := []struct {
		name    string
		config  *types.SAMLConfig
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  validConfig,
			wantErr: false,
		},
		{
			name: "empty entity_id_hash",
			config: &types.SAMLConfig{
				EntityIDHash:               "",
				CertificateFingerprintHash: types.HashADIdentifier("cert"),
				NameIDFormat:               "format",
			},
			wantErr: true,
		},
		{
			name: "empty certificate_fingerprint_hash",
			config: &types.SAMLConfig{
				EntityIDHash:               types.HashADIdentifier("entity"),
				CertificateFingerprintHash: "",
				NameIDFormat:               "format",
			},
			wantErr: true,
		},
		{
			name: "empty name_id_format",
			config: &types.SAMLConfig{
				EntityIDHash:               types.HashADIdentifier("entity"),
				CertificateFingerprintHash: types.HashADIdentifier("cert"),
				NameIDFormat:               "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// LDAP Config Tests
// ============================================================================

func TestLDAPConfig_NewAndValidate(t *testing.T) {
	config := types.NewLDAPConfig(
		"ldaps://ad.contoso.com:636",
		"DC=contoso,DC=com",
		"contoso.com",
		true,
		false,
	)

	require.NotNil(t, config)
	assert.Len(t, config.ServerHash, 64)
	assert.Len(t, config.BaseDNHash, 64)
	assert.Len(t, config.DomainHash, 64)
	assert.True(t, config.UseSSL)
	assert.False(t, config.UseTLS)

	err := config.Validate()
	require.NoError(t, err)
}

func TestLDAPConfig_Validate(t *testing.T) {
	validConfigSSL := types.NewLDAPConfig(
		"ldaps://ad.example.com",
		"DC=example,DC=com",
		"example.com",
		true,
		false,
	)

	validConfigTLS := types.NewLDAPConfig(
		"ldap://ad.example.com",
		"DC=example,DC=com",
		"example.com",
		false,
		true,
	)

	tests := []struct {
		name    string
		config  *types.LDAPConfig
		wantErr bool
	}{
		{
			name:    "valid config with SSL",
			config:  validConfigSSL,
			wantErr: false,
		},
		{
			name:    "valid config with TLS",
			config:  validConfigTLS,
			wantErr: false,
		},
		{
			name: "empty server_hash",
			config: &types.LDAPConfig{
				ServerHash: "",
				BaseDNHash: types.HashADIdentifier("base"),
				DomainHash: types.HashADIdentifier("domain"),
				UseSSL:     true,
			},
			wantErr: true,
		},
		{
			name: "empty base_dn_hash",
			config: &types.LDAPConfig{
				ServerHash: types.HashADIdentifier("server"),
				BaseDNHash: "",
				DomainHash: types.HashADIdentifier("domain"),
				UseSSL:     true,
			},
			wantErr: true,
		},
		{
			name: "empty domain_hash",
			config: &types.LDAPConfig{
				ServerHash: types.HashADIdentifier("server"),
				BaseDNHash: types.HashADIdentifier("base"),
				DomainHash: "",
				UseSSL:     true,
			},
			wantErr: true,
		},
		{
			name: "no SSL or TLS",
			config: &types.LDAPConfig{
				ServerHash: types.HashADIdentifier("server"),
				BaseDNHash: types.HashADIdentifier("base"),
				DomainHash: types.HashADIdentifier("domain"),
				UseSSL:     false,
				UseTLS:     false,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// AD SSO Linkage Metadata Tests
// ============================================================================

func TestADSSOLinkageMetadata_NewAndValidate(t *testing.T) {
	now := time.Now()
	metadata := types.NewADSSOLinkageMetadata(
		"linkage-123",
		"cosmos1abc123",
		types.ADAuthMethodOIDC,
		"user@contoso.com",
		"12345678-1234-1234-1234-123456789abc",
		"nonce-xyz",
		now,
	)

	require.NotNil(t, metadata)
	assert.Equal(t, types.ADSSOVersion, metadata.Version)
	assert.Equal(t, "linkage-123", metadata.LinkageID)
	assert.Equal(t, "cosmos1abc123", metadata.AccountAddress)
	assert.Equal(t, types.ADAuthMethodOIDC, metadata.AuthMethod)
	assert.Len(t, metadata.SubjectHash, 64)
	assert.Len(t, metadata.TenantHash, 64)
	assert.Equal(t, "nonce-xyz", metadata.Nonce)
	assert.Equal(t, types.ADSSOStatusVerified, metadata.Status)
}

func TestADSSOLinkageMetadata_Validate(t *testing.T) {
	now := time.Now()

	// Create valid base metadata with OIDC config
	createValidOIDC := func() *types.ADSSOLinkageMetadata {
		m := types.NewADSSOLinkageMetadata(
			"linkage-123",
			"cosmos1abc123",
			types.ADAuthMethodOIDC,
			"user@contoso.com",
			"tenant-123",
			"nonce-xyz",
			now,
		)
		m.OIDCConfig = types.NewAzureADOIDCConfig("tenant-123", "https://login.microsoftonline.com/tenant-123/v2.0", "client-456")
		m.WalletBindingSignature = []byte("valid-signature")
		return m
	}

	// Create valid SAML metadata
	createValidSAML := func() *types.ADSSOLinkageMetadata {
		m := types.NewADSSOLinkageMetadata(
			"linkage-456",
			"cosmos1def456",
			types.ADAuthMethodSAML,
			"user@example.com",
			"example.com",
			"nonce-abc",
			now,
		)
		m.SAMLConfig = types.NewSAMLConfig("https://idp.example.com", "", "fingerprint", "format")
		m.WalletBindingSignature = []byte("valid-signature")
		return m
	}

	// Create valid LDAP metadata
	createValidLDAP := func() *types.ADSSOLinkageMetadata {
		m := types.NewADSSOLinkageMetadata(
			"linkage-789",
			"cosmos1ghi789",
			types.ADAuthMethodLDAP,
			"CN=User,DC=corp,DC=com",
			"corp.com",
			"nonce-def",
			now,
		)
		m.LDAPConfig = types.NewLDAPConfig("ldaps://ad.corp.com", "DC=corp,DC=com", "corp.com", true, false)
		m.WalletBindingSignature = []byte("valid-signature")
		return m
	}

	tests := []struct {
		name     string
		metadata *types.ADSSOLinkageMetadata
		wantErr  bool
	}{
		{
			name:     "valid OIDC linkage",
			metadata: createValidOIDC(),
			wantErr:  false,
		},
		{
			name:     "valid SAML linkage",
			metadata: createValidSAML(),
			wantErr:  false,
		},
		{
			name:     "valid LDAP linkage",
			metadata: createValidLDAP(),
			wantErr:  false,
		},
		{
			name: "invalid version",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.Version = 999
				return m
			}(),
			wantErr: true,
		},
		{
			name: "empty linkage_id",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.LinkageID = ""
				return m
			}(),
			wantErr: true,
		},
		{
			name: "empty account_address",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.AccountAddress = ""
				return m
			}(),
			wantErr: true,
		},
		{
			name: "invalid auth_method",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.AuthMethod = authMethodInvalid
				return m
			}(),
			wantErr: true,
		},
		{
			name: "empty subject_hash",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.SubjectHash = ""
				return m
			}(),
			wantErr: true,
		},
		{
			name: "invalid subject_hash length",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.SubjectHash = "too-short"
				return m
			}(),
			wantErr: true,
		},
		{
			name: "empty tenant_hash",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.TenantHash = ""
				return m
			}(),
			wantErr: true,
		},
		{
			name: "empty nonce",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.Nonce = ""
				return m
			}(),
			wantErr: true,
		},
		{
			name: "invalid status",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.Status = "unknown"
				return m
			}(),
			wantErr: true,
		},
		{
			name: "zero verified_at",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.VerifiedAt = time.Time{}
				return m
			}(),
			wantErr: true,
		},
		{
			name: "zero created_at",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.CreatedAt = time.Time{}
				return m
			}(),
			wantErr: true,
		},
		{
			name: "missing wallet_binding_signature when verified",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.WalletBindingSignature = nil
				return m
			}(),
			wantErr: true,
		},
		{
			name: "OIDC without oidc_config",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidOIDC()
				m.OIDCConfig = nil
				return m
			}(),
			wantErr: true,
		},
		{
			name: "SAML without saml_config",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidSAML()
				m.SAMLConfig = nil
				return m
			}(),
			wantErr: true,
		},
		{
			name: "LDAP without ldap_config",
			metadata: func() *types.ADSSOLinkageMetadata {
				m := createValidLDAP()
				m.LDAPConfig = nil
				return m
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.metadata.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestADSSOLinkageMetadata_IsActive(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name   string
		setup  func() *types.ADSSOLinkageMetadata
		expect bool
	}{
		{
			name: "verified with no expiry",
			setup: func() *types.ADSSOLinkageMetadata {
				m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)
				return m
			},
			expect: true,
		},
		{
			name: "verified with future expiry",
			setup: func() *types.ADSSOLinkageMetadata {
				m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)
				m.ExpiresAt = &future
				return m
			},
			expect: true,
		},
		{
			name: "verified with past expiry",
			setup: func() *types.ADSSOLinkageMetadata {
				m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)
				m.ExpiresAt = &past
				return m
			},
			expect: false,
		},
		{
			name: "pending status",
			setup: func() *types.ADSSOLinkageMetadata {
				m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)
				m.Status = types.ADSSOStatusPending
				return m
			},
			expect: false,
		},
		{
			name: "revoked status",
			setup: func() *types.ADSSOLinkageMetadata {
				m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)
				m.Status = types.ADSSOStatusRevoked
				return m
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.setup()
			assert.Equal(t, tt.expect, m.IsActive())
		})
	}
}

func TestADSSOLinkageMetadata_GroupHashes(t *testing.T) {
	now := time.Now()
	m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)

	// Initially no groups
	assert.Empty(t, m.GroupHashes)

	// Add groups
	m.AddGroupHash("group-admins")
	m.AddGroupHash("group-users")
	assert.Len(t, m.GroupHashes, 2)

	// Adding same group again should not duplicate
	m.AddGroupHash("group-admins")
	assert.Len(t, m.GroupHashes, 2)

	// Check group existence
	assert.True(t, m.HasGroupHash("group-admins"))
	assert.True(t, m.HasGroupHash("group-users"))
	assert.False(t, m.HasGroupHash("group-nonexistent"))
}

func TestADSSOLinkageMetadata_Revoke(t *testing.T) {
	now := time.Now()
	m := types.NewADSSOLinkageMetadata("id", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)

	assert.Equal(t, types.ADSSOStatusVerified, m.Status)
	assert.Nil(t, m.RevokedAt)
	assert.Empty(t, m.RevokedReason)

	m.Revoke("Security concern")

	assert.Equal(t, types.ADSSOStatusRevoked, m.Status)
	assert.NotNil(t, m.RevokedAt)
	assert.Equal(t, "Security concern", m.RevokedReason)
	assert.False(t, m.IsActive())
}

func TestADSSOLinkageMetadata_String(t *testing.T) {
	now := time.Now()
	m := types.NewADSSOLinkageMetadata("linkage-123", "addr", types.ADAuthMethodOIDC, "sub", "tenant", "nonce", now)

	str := m.String()
	assert.Contains(t, str, "linkage-123")
	assert.Contains(t, str, "oidc")
	assert.Contains(t, str, "verified")
}

// ============================================================================
// AD SSO Challenge Tests
// ============================================================================

func TestADSSOChallenge_NewAndValidate(t *testing.T) {
	now := time.Now()
	challenge := types.NewADSSOChallenge(
		"challenge-123",
		"cosmos1abc123",
		types.ADAuthMethodOIDC,
		"nonce-xyz",
		now,
		600, // 10 minutes
	)

	require.NotNil(t, challenge)
	assert.Equal(t, "challenge-123", challenge.ChallengeID)
	assert.Equal(t, "cosmos1abc123", challenge.AccountAddress)
	assert.Equal(t, types.ADAuthMethodOIDC, challenge.AuthMethod)
	assert.Equal(t, "nonce-xyz", challenge.Nonce)
	assert.Equal(t, types.ADSSOStatusPending, challenge.Status)
	assert.True(t, challenge.ExpiresAt.After(challenge.CreatedAt))

	err := challenge.Validate()
	require.NoError(t, err)
}

func TestADSSOChallenge_DefaultTTL(t *testing.T) {
	now := time.Now()
	challenge := types.NewADSSOChallenge(
		"challenge-123",
		"cosmos1abc123",
		types.ADAuthMethodSAML,
		"nonce-xyz",
		now,
		0, // Should use default TTL
	)

	expectedExpiry := now.Add(time.Duration(types.ChallengeTTLSeconds) * time.Second)
	assert.WithinDuration(t, expectedExpiry, challenge.ExpiresAt, time.Second)
}

func TestADSSOChallenge_Validate(t *testing.T) {
	now := time.Now()

	createValid := func() *types.ADSSOChallenge {
		return types.NewADSSOChallenge("challenge-123", "cosmos1abc123", types.ADAuthMethodOIDC, "nonce-xyz", now, 600)
	}

	tests := []struct {
		name      string
		challenge *types.ADSSOChallenge
		wantErr   bool
	}{
		{
			name:      "valid challenge",
			challenge: createValid(),
			wantErr:   false,
		},
		{
			name: "empty challenge_id",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.ChallengeID = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "empty account_address",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.AccountAddress = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "invalid auth_method",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.AuthMethod = authMethodInvalid
				return c
			}(),
			wantErr: true,
		},
		{
			name: "empty nonce",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.Nonce = ""
				return c
			}(),
			wantErr: true,
		},
		{
			name: "zero created_at",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.CreatedAt = time.Time{}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "zero expires_at",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.ExpiresAt = time.Time{}
				return c
			}(),
			wantErr: true,
		},
		{
			name: "expires_at before created_at",
			challenge: func() *types.ADSSOChallenge {
				c := createValid()
				c.ExpiresAt = c.CreatedAt.Add(-time.Hour)
				return c
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.challenge.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestADSSOChallenge_IsExpired(t *testing.T) {
	now := time.Now()
	challenge := types.NewADSSOChallenge("id", "addr", types.ADAuthMethodOIDC, "nonce", now, 600)

	// Not expired at creation
	assert.False(t, challenge.IsExpired(now))

	// Not expired 5 minutes later
	assert.False(t, challenge.IsExpired(now.Add(5*time.Minute)))

	// Expired 15 minutes later (after 10 minute TTL)
	assert.True(t, challenge.IsExpired(now.Add(15*time.Minute)))
}

func TestADSSOChallenge_Complete(t *testing.T) {
	now := time.Now()
	challenge := types.NewADSSOChallenge("id", "addr", types.ADAuthMethodOIDC, "nonce", now, 600)

	assert.Equal(t, types.ADSSOStatusPending, challenge.Status)
	assert.Nil(t, challenge.CompletedAt)

	challenge.Complete()

	assert.Equal(t, types.ADSSOStatusVerified, challenge.Status)
	assert.NotNil(t, challenge.CompletedAt)
}

// ============================================================================
// AD Wallet Binding Tests
// ============================================================================

func TestADWalletBinding_NewAndValidate(t *testing.T) {
	now := time.Now()
	binding := types.NewADWalletBinding(
		"binding-123",
		"cosmos1abc123",
		"linkage-456",
		"user@contoso.com",
		"contoso.com",
		types.ADAuthMethodOIDC,
		"I authorize this wallet binding",
		[]byte("signature-bytes"),
		now,
	)

	require.NotNil(t, binding)
	assert.Equal(t, "binding-123", binding.BindingID)
	assert.Equal(t, "cosmos1abc123", binding.WalletAddress)
	assert.Equal(t, "linkage-456", binding.LinkageID)
	assert.Len(t, binding.SubjectHash, 64)
	assert.Len(t, binding.TenantHash, 64)
	assert.Equal(t, types.ADAuthMethodOIDC, binding.AuthMethod)
	assert.Equal(t, "I authorize this wallet binding", binding.BindingMessage)
	assert.Equal(t, []byte("signature-bytes"), binding.BindingSignature)
	assert.True(t, binding.Active)

	err := binding.Validate()
	require.NoError(t, err)
}

func TestADWalletBinding_Validate(t *testing.T) {
	now := time.Now()

	createValid := func() *types.ADWalletBinding {
		return types.NewADWalletBinding(
			"binding-123",
			"cosmos1abc123",
			"linkage-456",
			"user@contoso.com",
			"contoso.com",
			types.ADAuthMethodOIDC,
			"message",
			[]byte("signature"),
			now,
		)
	}

	tests := []struct {
		name    string
		binding *types.ADWalletBinding
		wantErr bool
	}{
		{
			name:    "valid binding",
			binding: createValid(),
			wantErr: false,
		},
		{
			name: "empty binding_id",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.BindingID = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "empty wallet_address",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.WalletAddress = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "empty linkage_id",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.LinkageID = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "empty subject_hash",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.SubjectHash = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "invalid subject_hash length",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.SubjectHash = "too-short"
				return b
			}(),
			wantErr: true,
		},
		{
			name: "empty tenant_hash",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.TenantHash = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "invalid auth_method",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.AuthMethod = authMethodInvalid
				return b
			}(),
			wantErr: true,
		},
		{
			name: "empty binding_signature",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.BindingSignature = nil
				return b
			}(),
			wantErr: true,
		},
		{
			name: "empty binding_message",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.BindingMessage = ""
				return b
			}(),
			wantErr: true,
		},
		{
			name: "zero created_at",
			binding: func() *types.ADWalletBinding {
				b := createValid()
				b.CreatedAt = time.Time{}
				return b
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.binding.Validate()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestADWalletBinding_IsActive(t *testing.T) {
	now := time.Now()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name   string
		setup  func() *types.ADWalletBinding
		expect bool
	}{
		{
			name: "active binding with no expiry",
			setup: func() *types.ADWalletBinding {
				return types.NewADWalletBinding("id", "addr", "link", "sub", "tenant", types.ADAuthMethodOIDC, "msg", []byte("sig"), now)
			},
			expect: true,
		},
		{
			name: "active binding with future expiry",
			setup: func() *types.ADWalletBinding {
				b := types.NewADWalletBinding("id", "addr", "link", "sub", "tenant", types.ADAuthMethodOIDC, "msg", []byte("sig"), now)
				b.ExpiresAt = &future
				return b
			},
			expect: true,
		},
		{
			name: "binding with past expiry",
			setup: func() *types.ADWalletBinding {
				b := types.NewADWalletBinding("id", "addr", "link", "sub", "tenant", types.ADAuthMethodOIDC, "msg", []byte("sig"), now)
				b.ExpiresAt = &past
				return b
			},
			expect: false,
		},
		{
			name: "inactive binding",
			setup: func() *types.ADWalletBinding {
				b := types.NewADWalletBinding("id", "addr", "link", "sub", "tenant", types.ADAuthMethodOIDC, "msg", []byte("sig"), now)
				b.Active = false
				return b
			},
			expect: false,
		},
		{
			name: "revoked binding",
			setup: func() *types.ADWalletBinding {
				b := types.NewADWalletBinding("id", "addr", "link", "sub", "tenant", types.ADAuthMethodOIDC, "msg", []byte("sig"), now)
				b.Revoke()
				return b
			},
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setup()
			assert.Equal(t, tt.expect, b.IsActive())
		})
	}
}

func TestADWalletBinding_Revoke(t *testing.T) {
	now := time.Now()
	binding := types.NewADWalletBinding("id", "addr", "link", "sub", "tenant", types.ADAuthMethodOIDC, "msg", []byte("sig"), now)

	assert.True(t, binding.Active)
	assert.Nil(t, binding.RevokedAt)

	binding.Revoke()

	assert.False(t, binding.Active)
	assert.NotNil(t, binding.RevokedAt)
	assert.False(t, binding.IsActive())
}

// ============================================================================
// AD SSO Scoring Tests
// ============================================================================

func TestDefaultADSSOScoringWeights(t *testing.T) {
	weights := types.DefaultADSSOScoringWeights()

	require.Len(t, weights, 3)

	// Find each weight
	var oidcWeight, samlWeight, ldapWeight uint32
	for _, w := range weights {
		switch w.AuthMethod {
		case types.ADAuthMethodOIDC:
			oidcWeight = w.Weight
		case types.ADAuthMethodSAML:
			samlWeight = w.Weight
		case types.ADAuthMethodLDAP:
			ldapWeight = w.Weight
		}
	}

	// OIDC should have highest weight (Azure AD cloud = most trust)
	assert.Greater(t, oidcWeight, samlWeight)
	assert.Greater(t, samlWeight, ldapWeight)
	assert.Greater(t, ldapWeight, uint32(0))
}

func TestGetADSSOScoringWeight(t *testing.T) {
	tests := []struct {
		method types.ADAuthMethod
		expect uint32
	}{
		{types.ADAuthMethodOIDC, 400},
		{types.ADAuthMethodSAML, 350},
		{types.ADAuthMethodLDAP, 300},
		{types.ADAuthMethod("unknown"), 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.method), func(t *testing.T) {
			got := types.GetADSSOScoringWeight(tt.method)
			assert.Equal(t, tt.expect, got)
		})
	}
}

// ============================================================================
// AD SSO Verification Result Tests
// ============================================================================

func TestNewADSSOVerificationResult(t *testing.T) {
	now := time.Now()
	result := types.NewADSSOVerificationResult(
		"challenge-123",
		"linkage-456",
		types.ADAuthMethodOIDC,
		"user@contoso.com",
		"contoso.com",
		now,
	)

	require.NotNil(t, result)
	assert.Equal(t, "challenge-123", result.ChallengeID)
	assert.Equal(t, "linkage-456", result.LinkageID)
	assert.True(t, result.Success)
	assert.Equal(t, types.ADAuthMethodOIDC, result.AuthMethod)
	assert.Len(t, result.SubjectHash, 64)
	assert.Len(t, result.TenantHash, 64)
	assert.Empty(t, result.ErrorCode)
	assert.Empty(t, result.ErrorMessage)
	assert.Equal(t, types.GetADSSOScoringWeight(types.ADAuthMethodOIDC), result.ScoreContribution)
}

func TestNewADSSOVerificationFailure(t *testing.T) {
	result := types.NewADSSOVerificationFailure(
		"challenge-123",
		types.ADAuthMethodSAML,
		"auth_failed",
		"SAML assertion validation failed",
	)

	require.NotNil(t, result)
	assert.Equal(t, "challenge-123", result.ChallengeID)
	assert.False(t, result.Success)
	assert.Equal(t, types.ADAuthMethodSAML, result.AuthMethod)
	assert.Equal(t, "auth_failed", result.ErrorCode)
	assert.Equal(t, "SAML assertion validation failed", result.ErrorMessage)
	assert.Empty(t, result.LinkageID)
	assert.Empty(t, result.SubjectHash)
	assert.Equal(t, uint32(0), result.ScoreContribution)
}

// ============================================================================
// Integration Tests
// ============================================================================

func TestADSSO_EndToEndFlow_OIDC(t *testing.T) {
	now := time.Now()

	// Step 1: Create a challenge
	challenge := types.NewADSSOChallenge(
		"challenge-oidc-123",
		"cosmos1wallet123",
		types.ADAuthMethodOIDC,
		"nonce-secure-random",
		now,
		600,
	)
	challenge.RedirectURI = "https://app.virtengine.com/callback"
	require.NoError(t, challenge.Validate())

	// Step 2: Complete the challenge (simulating OAuth callback)
	challenge.Complete()
	assert.Equal(t, types.ADSSOStatusVerified, challenge.Status)

	// Step 3: Create linkage metadata
	linkage := types.NewADSSOLinkageMetadata(
		"linkage-oidc-456",
		"cosmos1wallet123",
		types.ADAuthMethodOIDC,
		"john.doe@contoso.com",
		"12345678-1234-1234-1234-123456789abc",
		challenge.Nonce,
		*challenge.CompletedAt,
	)
	linkage.SetOIDCConfig(types.NewAzureADOIDCConfig(
		"12345678-1234-1234-1234-123456789abc",
		"https://login.microsoftonline.com/12345678-1234-1234-1234-123456789abc/v2.0",
		"app-client-id",
	))
	linkage.SetUPNHash("john.doe@contoso.com")
	linkage.AddGroupHash("domain-admins")
	linkage.AddGroupHash("azure-users")
	linkage.SetWalletBinding([]byte("wallet-signature"))
	require.NoError(t, linkage.Validate())

	// Step 4: Create wallet binding
	binding := types.NewADWalletBinding(
		"binding-789",
		"cosmos1wallet123",
		linkage.LinkageID,
		"john.doe@contoso.com",
		"12345678-1234-1234-1234-123456789abc",
		types.ADAuthMethodOIDC,
		"I authorize VirtEngine to link my Azure AD identity",
		[]byte("wallet-signature"),
		now,
	)
	require.NoError(t, binding.Validate())

	// Step 5: Create verification result
	result := types.NewADSSOVerificationResult(
		challenge.ChallengeID,
		linkage.LinkageID,
		types.ADAuthMethodOIDC,
		"john.doe@contoso.com",
		"12345678-1234-1234-1234-123456789abc",
		*challenge.CompletedAt,
	)
	assert.True(t, result.Success)
	assert.Equal(t, uint32(400), result.ScoreContribution)

	// Verify all hashes are consistent
	assert.Equal(t, linkage.SubjectHash, result.SubjectHash)
	assert.Equal(t, linkage.TenantHash, result.TenantHash)
	assert.Equal(t, binding.SubjectHash, result.SubjectHash)
}

func TestADSSO_EndToEndFlow_LDAP(t *testing.T) {
	now := time.Now()

	// Step 1: Create LDAP challenge
	challenge := types.NewADSSOChallenge(
		"challenge-ldap-123",
		"cosmos1onprem456",
		types.ADAuthMethodLDAP,
		"ldap-nonce-xyz",
		now,
		600,
	)
	require.NoError(t, challenge.Validate())

	// Step 2: Complete the challenge
	challenge.Complete()

	// Step 3: Create LDAP linkage
	linkage := types.NewADSSOLinkageMetadata(
		"linkage-ldap-789",
		"cosmos1onprem456",
		types.ADAuthMethodLDAP,
		"CN=Jane Smith,OU=Users,DC=corp,DC=example,DC=com",
		"corp.example.com",
		challenge.Nonce,
		*challenge.CompletedAt,
	)
	linkage.SetLDAPConfig(types.NewLDAPConfig(
		"ldaps://ad.corp.example.com:636",
		"DC=corp,DC=example,DC=com",
		"corp.example.com",
		true,
		false,
	))
	linkage.SetWalletBinding([]byte("ldap-wallet-sig"))
	require.NoError(t, linkage.Validate())

	// Verify LDAP scoring
	result := types.NewADSSOVerificationResult(
		challenge.ChallengeID,
		linkage.LinkageID,
		types.ADAuthMethodLDAP,
		"CN=Jane Smith,OU=Users,DC=corp,DC=example,DC=com",
		"corp.example.com",
		*challenge.CompletedAt,
	)
	assert.True(t, result.Success)
	assert.Equal(t, uint32(300), result.ScoreContribution) // LDAP weight
}

func TestADSSO_SecurityNoSensitiveData(t *testing.T) {
	now := time.Now()

	// Create a complete linkage
	linkage := types.NewADSSOLinkageMetadata(
		"linkage-security-test",
		"cosmos1security123",
		types.ADAuthMethodOIDC,
		"sensitive.user@secret-org.com", // Sensitive!
		"secret-tenant-id-12345",        // Sensitive!
		"nonce-value",
		now,
	)
	linkage.SetOIDCConfig(types.NewAzureADOIDCConfig(
		"secret-tenant-id-12345", // Sensitive!
		"https://login.microsoftonline.com/secret-tenant-id-12345/v2.0",
		"secret-client-id", // Sensitive!
	))
	linkage.SetWalletBinding([]byte("sig"))

	// Verify no sensitive data appears in plain text
	str := linkage.String()
	assert.NotContains(t, str, "sensitive.user@secret-org.com")
	assert.NotContains(t, str, "secret-tenant-id-12345")
	assert.NotContains(t, str, "secret-client-id")

	// Verify hashes are stored, not raw values
	assert.NotEqual(t, "sensitive.user@secret-org.com", linkage.SubjectHash)
	assert.NotEqual(t, "secret-tenant-id-12345", linkage.TenantHash)
	assert.Len(t, linkage.SubjectHash, 64)
	assert.Len(t, linkage.TenantHash, 64)
}
