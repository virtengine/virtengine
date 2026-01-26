package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestApprovedClientValidate tests ApprovedClient validation
func TestApprovedClientValidate(t *testing.T) {
	validKey := make([]byte, 32) // Valid ed25519 key length

	testCases := []struct {
		name      string
		client    ApprovedClient
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid client",
			client: ApprovedClient{
				ClientID:     "valid-client",
				Name:         "Valid Client",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: false,
		},
		{
			name: "empty client ID",
			client: ApprovedClient{
				ClientID:     "",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "client_id cannot be empty",
		},
		{
			name: "client ID too long",
			client: ApprovedClient{
				ClientID:     "this-is-a-very-long-client-id-that-exceeds-the-maximum-allowed-length-of-64-characters",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "cannot exceed 64 characters",
		},
		{
			name: "invalid client ID format",
			client: ApprovedClient{
				ClientID:     "123-invalid",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "must start with a letter",
		},
		{
			name: "empty name",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "name cannot be empty",
		},
		{
			name: "empty public key",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    []byte{},
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "public_key cannot be empty",
		},
		{
			name: "invalid key type",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      "invalid",
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "invalid key type",
		},
		{
			name: "wrong ed25519 key length",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    make([]byte, 16), // Wrong length
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "ed25519 public key must be 32 bytes",
		},
		{
			name: "empty min version",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "min_version cannot be empty",
		},
		{
			name: "invalid min version format",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "invalid",
				Status:       ClientStatusActive,
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "invalid min_version",
		},
		{
			name: "invalid status",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       "invalid",
				RegisteredBy: "virtengine1test",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "invalid status",
		},
		{
			name: "empty registered by",
			client: ApprovedClient{
				ClientID:     "test-client",
				Name:         "Test",
				PublicKey:    validKey,
				KeyType:      KeyTypeEd25519,
				MinVersion:   "1.0.0",
				Status:       ClientStatusActive,
				RegisteredBy: "",
				RegisteredAt: time.Now(),
			},
			expectErr: true,
			errMsg:    "registered_by cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.client.Validate()
			if tc.expectErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestClientStatusTransitions tests status transition logic
func TestClientStatusTransitions(t *testing.T) {
	testCases := []struct {
		name     string
		from     ClientStatus
		to       ClientStatus
		canTrans bool
	}{
		{"active to suspended", ClientStatusActive, ClientStatusSuspended, true},
		{"active to revoked", ClientStatusActive, ClientStatusRevoked, true},
		{"active to active", ClientStatusActive, ClientStatusActive, false},
		{"suspended to active", ClientStatusSuspended, ClientStatusActive, true},
		{"suspended to revoked", ClientStatusSuspended, ClientStatusRevoked, true},
		{"suspended to suspended", ClientStatusSuspended, ClientStatusSuspended, false},
		{"revoked to active", ClientStatusRevoked, ClientStatusActive, false},
		{"revoked to suspended", ClientStatusRevoked, ClientStatusSuspended, false},
		{"revoked to revoked", ClientStatusRevoked, ClientStatusRevoked, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.from.CanTransitionTo(tc.to)
			require.Equal(t, tc.canTrans, result)
		})
	}
}

// TestApprovedClientCanSubmitScope tests scope permission checking
func TestApprovedClientCanSubmitScope(t *testing.T) {
	client := &ApprovedClient{
		Status:        ClientStatusActive,
		AllowedScopes: []string{"selfie", "document_front"},
	}

	// Allowed scopes
	require.True(t, client.CanSubmitScope("selfie"))
	require.True(t, client.CanSubmitScope("document_front"))

	// Disallowed scope
	require.False(t, client.CanSubmitScope("document_back"))

	// Empty allowed scopes means all allowed
	clientAllScopes := &ApprovedClient{
		Status:        ClientStatusActive,
		AllowedScopes: []string{},
	}
	require.True(t, clientAllScopes.CanSubmitScope("anything"))

	// Wildcard scope
	clientWildcard := &ApprovedClient{
		Status:        ClientStatusActive,
		AllowedScopes: []string{"*"},
	}
	require.True(t, clientWildcard.CanSubmitScope("anything"))

	// Inactive client cannot submit any scope
	clientInactive := &ApprovedClient{
		Status:        ClientStatusSuspended,
		AllowedScopes: []string{"selfie"},
	}
	require.False(t, clientInactive.CanSubmitScope("selfie"))
}

// TestApprovedClientStatusMethods tests status helper methods
func TestApprovedClientStatusMethods(t *testing.T) {
	activeClient := &ApprovedClient{Status: ClientStatusActive}
	require.True(t, activeClient.IsActive())
	require.False(t, activeClient.IsSuspended())
	require.False(t, activeClient.IsRevoked())

	suspendedClient := &ApprovedClient{Status: ClientStatusSuspended}
	require.False(t, suspendedClient.IsActive())
	require.True(t, suspendedClient.IsSuspended())
	require.False(t, suspendedClient.IsRevoked())

	revokedClient := &ApprovedClient{Status: ClientStatusRevoked}
	require.False(t, revokedClient.IsActive())
	require.False(t, revokedClient.IsSuspended())
	require.True(t, revokedClient.IsRevoked())
}

// TestApprovedClientSuspend tests client suspension
func TestApprovedClientSuspend(t *testing.T) {
	now := time.Now()
	client := &ApprovedClient{Status: ClientStatusActive}

	err := client.Suspend("Security concern", "admin", now)
	require.NoError(t, err)
	require.Equal(t, ClientStatusSuspended, client.Status)
	require.Equal(t, "Security concern", client.StatusReason)
	require.NotNil(t, client.SuspendedAt)
	require.Equal(t, "admin", client.LastUpdatedBy)

	// Cannot suspend already suspended client
	err = client.Suspend("Another reason", "admin", now)
	require.Error(t, err)
}

// TestApprovedClientRevoke tests client revocation
func TestApprovedClientRevoke(t *testing.T) {
	now := time.Now()
	client := &ApprovedClient{Status: ClientStatusActive}

	err := client.Revoke("Compromised", "admin", now)
	require.NoError(t, err)
	require.Equal(t, ClientStatusRevoked, client.Status)
	require.Equal(t, "Compromised", client.StatusReason)
	require.NotNil(t, client.RevokedAt)

	// Cannot revoke already revoked client
	err = client.Revoke("Another reason", "admin", now)
	require.Error(t, err)
}

// TestApprovedClientReactivate tests client reactivation
func TestApprovedClientReactivate(t *testing.T) {
	now := time.Now()
	client := &ApprovedClient{Status: ClientStatusSuspended}
	suspendedAt := now.Add(-time.Hour)
	client.SuspendedAt = &suspendedAt

	err := client.Reactivate("Issue resolved", "admin", now)
	require.NoError(t, err)
	require.Equal(t, ClientStatusActive, client.Status)
	require.Equal(t, "Issue resolved", client.StatusReason)
	require.Nil(t, client.SuspendedAt) // Should be cleared

	// Cannot reactivate active client
	err = client.Reactivate("Another reason", "admin", now)
	require.Error(t, err)

	// Cannot reactivate revoked client
	revokedClient := &ApprovedClient{Status: ClientStatusRevoked}
	err = revokedClient.Reactivate("Try to reactivate", "admin", now)
	require.Error(t, err)
}

// TestApprovedClientUpdate tests client update
func TestApprovedClientUpdate(t *testing.T) {
	now := time.Now()
	client := &ApprovedClient{
		Name:          "Original",
		Description:   "Original description",
		MinVersion:    "1.0.0",
		MaxVersion:    "2.0.0",
		AllowedScopes: []string{"scope1"},
	}

	err := client.Update("Updated", "New description", "1.5.0", "3.0.0", []string{"scope1", "scope2"}, "admin", now)
	require.NoError(t, err)
	require.Equal(t, "Updated", client.Name)
	require.Equal(t, "New description", client.Description)
	require.Equal(t, "1.5.0", client.MinVersion)
	require.Equal(t, "3.0.0", client.MaxVersion)
	require.Equal(t, []string{"scope1", "scope2"}, client.AllowedScopes)
	require.Equal(t, "admin", client.LastUpdatedBy)

	// Partial update (empty strings should not update)
	err = client.Update("", "", "", "", nil, "admin2", now)
	require.NoError(t, err)
	require.Equal(t, "Updated", client.Name) // Should not change
	require.Equal(t, "admin2", client.LastUpdatedBy)
}

// TestKeyTypeValid tests key type validation
func TestKeyTypeValid(t *testing.T) {
	require.True(t, KeyTypeEd25519.IsValid())
	require.True(t, KeyTypeSecp256k1.IsValid())
	require.False(t, KeyType("invalid").IsValid())
}
