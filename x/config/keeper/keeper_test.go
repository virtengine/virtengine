package keeper

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	dbm "github.com/cosmos/cosmos-db"
	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/config/types"
)

// Test constants
const (
	testAuthority   = "virtengine1authority"
	testAdmin1      = "virtengine1admin1"
	testAdmin2      = "virtengine1admin2"
	testClientID    = "test-mobile-app"
	testClientID2   = "test-web-portal"
	testClientName  = "Test Mobile App"
	testMinVersion  = "1.0.0"
	testMaxVersion  = "2.0.0"
)

// KeeperTestSuite is the test suite for the config keeper
type KeeperTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper Keeper
	cdc    codec.BinaryCodec

	testPublicKey  []byte
	testPrivateKey ed25519.PrivateKey
}

// SetupTest sets up the test suite
func (s *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create store
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(s.T(), stateStore.LoadLatestVersion())

	// Create context
	s.ctx = sdk.NewContext(stateStore, false, log.NewNopLogger()).WithBlockTime(time.Now())

	// Create keeper
	s.keeper = NewKeeper(s.cdc, runtime.NewKVStoreService(storeKey), testAuthority)

	// Initialize with default params including admin
	params := types.DefaultParams()
	params.AdminAddresses = []string{testAdmin1, testAdmin2}
	require.NoError(s.T(), s.keeper.SetParams(s.ctx, params))

	// Generate test key pair
	s.testPublicKey, s.testPrivateKey, _ = ed25519.GenerateKey(rand.Reader)
}

// TestKeeperTestSuite runs the test suite
func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

// TestRegisterClient tests client registration
func (s *KeeperTestSuite) TestRegisterClient() {
	testCases := []struct {
		name      string
		client    types.ApprovedClient
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid registration",
			client: *types.NewApprovedClient(
				testClientID,
				testClientName,
				"Test mobile app for identity verification",
				s.testPublicKey,
				types.KeyTypeEd25519,
				testMinVersion,
				testMaxVersion,
				[]string{"selfie", "document_front", "document_back"},
				testAdmin1,
				s.ctx.BlockTime(),
			),
			expectErr: false,
		},
		{
			name: "duplicate client ID",
			client: *types.NewApprovedClient(
				testClientID,
				"Duplicate Client",
				"Should fail",
				s.testPublicKey,
				types.KeyTypeEd25519,
				testMinVersion,
				"",
				nil,
				testAdmin1,
				s.ctx.BlockTime(),
			),
			expectErr: true,
			errMsg:    "already exists",
		},
		{
			name: "empty client ID",
			client: types.ApprovedClient{
				ClientID:     "",
				Name:         "Invalid Client",
				PublicKey:    s.testPublicKey,
				KeyType:      types.KeyTypeEd25519,
				MinVersion:   testMinVersion,
				Status:       types.ClientStatusActive,
				RegisteredBy: testAdmin1,
				RegisteredAt: s.ctx.BlockTime(),
			},
			expectErr: true,
			errMsg:    "client_id cannot be empty",
		},
		{
			name: "invalid semver",
			client: types.ApprovedClient{
				ClientID:     "another-client",
				Name:         "Another Client",
				PublicKey:    s.testPublicKey,
				KeyType:      types.KeyTypeEd25519,
				MinVersion:   "invalid",
				Status:       types.ClientStatusActive,
				RegisteredBy: testAdmin1,
				RegisteredAt: s.ctx.BlockTime(),
			},
			expectErr: true,
			errMsg:    "invalid min_version",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.RegisterClient(s.ctx, tc.client)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)

				// Verify client was stored
				client, found := s.keeper.GetClient(s.ctx, tc.client.ClientID)
				s.Require().True(found)
				s.Require().Equal(tc.client.ClientID, client.ClientID)
				s.Require().Equal(tc.client.Name, client.Name)
				s.Require().Equal(types.ClientStatusActive, client.Status)
			}
		})
	}
}

// TestUpdateClient tests client update
func (s *KeeperTestSuite) TestUpdateClient() {
	// First register a client
	client := types.NewApprovedClient(
		testClientID2,
		"Original Name",
		"Original description",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		[]string{"selfie"},
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	testCases := []struct {
		name        string
		clientID    string
		newName     string
		newVersion  string
		newScopes   []string
		expectErr   bool
		errMsg      string
	}{
		{
			name:       "update name",
			clientID:   testClientID2,
			newName:    "Updated Name",
			newVersion: "",
			newScopes:  nil,
			expectErr:  false,
		},
		{
			name:       "update version constraint",
			clientID:   testClientID2,
			newName:    "",
			newVersion: "1.5.0",
			newScopes:  nil,
			expectErr:  false,
		},
		{
			name:       "update allowed scopes",
			clientID:   testClientID2,
			newName:    "",
			newVersion: "",
			newScopes:  []string{"selfie", "document_front"},
			expectErr:  false,
		},
		{
			name:       "client not found",
			clientID:   "nonexistent",
			newName:    "Test",
			newVersion: "",
			newScopes:  nil,
			expectErr:  true,
			errMsg:     "not found",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.UpdateClient(s.ctx, tc.clientID, tc.newName, "", tc.newVersion, "", tc.newScopes, testAdmin1)
			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.errMsg)
			} else {
				s.Require().NoError(err)

				// Verify update
				updated, found := s.keeper.GetClient(s.ctx, tc.clientID)
				s.Require().True(found)

				if tc.newName != "" {
					s.Require().Equal(tc.newName, updated.Name)
				}
				if tc.newVersion != "" {
					s.Require().Equal(tc.newVersion, updated.MinVersion)
				}
				if tc.newScopes != nil {
					s.Require().Equal(tc.newScopes, updated.AllowedScopes)
				}
			}
		})
	}
}

// TestSuspendClient tests client suspension
func (s *KeeperTestSuite) TestSuspendClient() {
	// Register a fresh client for suspension tests
	client := types.NewApprovedClient(
		"suspend-test-client",
		"Suspend Test Client",
		"For testing suspension",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Test suspension
	err := s.keeper.SuspendClient(s.ctx, "suspend-test-client", "Security concern", testAdmin1)
	s.Require().NoError(err)

	// Verify suspension
	suspended, found := s.keeper.GetClient(s.ctx, "suspend-test-client")
	s.Require().True(found)
	s.Require().Equal(types.ClientStatusSuspended, suspended.Status)
	s.Require().Equal("Security concern", suspended.StatusReason)
	s.Require().NotNil(suspended.SuspendedAt)

	// Test that suspended client is not approved
	s.Require().False(s.keeper.IsClientApproved(s.ctx, "suspend-test-client"))
}

// TestRevokeClient tests client revocation
func (s *KeeperTestSuite) TestRevokeClient() {
	// Register a fresh client for revocation tests
	client := types.NewApprovedClient(
		"revoke-test-client",
		"Revoke Test Client",
		"For testing revocation",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Test revocation
	err := s.keeper.RevokeClient(s.ctx, "revoke-test-client", "Compromised keys", testAdmin1)
	s.Require().NoError(err)

	// Verify revocation
	revoked, found := s.keeper.GetClient(s.ctx, "revoke-test-client")
	s.Require().True(found)
	s.Require().Equal(types.ClientStatusRevoked, revoked.Status)
	s.Require().Equal("Compromised keys", revoked.StatusReason)
	s.Require().NotNil(revoked.RevokedAt)

	// Test that revoked client cannot be reactivated
	err = s.keeper.ReactivateClient(s.ctx, "revoke-test-client", "Attempting reactivation", testAdmin1)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "cannot reactivate")
}

// TestReactivateClient tests client reactivation
func (s *KeeperTestSuite) TestReactivateClient() {
	// Register and suspend a client
	client := types.NewApprovedClient(
		"reactivate-test-client",
		"Reactivate Test Client",
		"For testing reactivation",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))
	s.Require().NoError(s.keeper.SuspendClient(s.ctx, "reactivate-test-client", "Temporary issue", testAdmin1))

	// Test reactivation
	err := s.keeper.ReactivateClient(s.ctx, "reactivate-test-client", "Issue resolved", testAdmin1)
	s.Require().NoError(err)

	// Verify reactivation
	reactivated, found := s.keeper.GetClient(s.ctx, "reactivate-test-client")
	s.Require().True(found)
	s.Require().Equal(types.ClientStatusActive, reactivated.Status)
	s.Require().Equal("Issue resolved", reactivated.StatusReason)
	s.Require().Nil(reactivated.SuspendedAt)
}

// TestIsClientApproved tests client approval status check
func (s *KeeperTestSuite) TestIsClientApproved() {
	// Register an active client
	client := types.NewApprovedClient(
		"approval-test-client",
		"Approval Test Client",
		"For testing approval status",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Test active client is approved
	s.Require().True(s.keeper.IsClientApproved(s.ctx, "approval-test-client"))

	// Test non-existent client is not approved
	s.Require().False(s.keeper.IsClientApproved(s.ctx, "nonexistent"))

	// Suspend and check not approved
	s.Require().NoError(s.keeper.SuspendClient(s.ctx, "approval-test-client", "Testing", testAdmin1))
	s.Require().False(s.keeper.IsClientApproved(s.ctx, "approval-test-client"))
}

// TestIsAdmin tests admin authorization check
func (s *KeeperTestSuite) TestIsAdmin() {
	// Authority is always admin
	authorityAddr, _ := sdk.AccAddressFromBech32(testAuthority)
	s.Require().True(s.keeper.IsAdmin(s.ctx, authorityAddr))

	// Configured admins
	admin1Addr, _ := sdk.AccAddressFromBech32(testAdmin1)
	s.Require().True(s.keeper.IsAdmin(s.ctx, admin1Addr))

	admin2Addr, _ := sdk.AccAddressFromBech32(testAdmin2)
	s.Require().True(s.keeper.IsAdmin(s.ctx, admin2Addr))

	// Random address is not admin
	randomAddr := sdk.AccAddress("random-address")
	s.Require().False(s.keeper.IsAdmin(s.ctx, randomAddr))
}

// TestValidateClientVersion tests version constraint validation
func (s *KeeperTestSuite) TestValidateClientVersion() {
	// Register a client with version constraints
	client := types.NewApprovedClient(
		"version-test-client",
		"Version Test Client",
		"For testing version validation",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"2.0.0",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	testCases := []struct {
		name      string
		version   string
		expectErr bool
	}{
		{"at minimum", "1.0.0", false},
		{"at maximum", "2.0.0", false},
		{"in range", "1.5.0", false},
		{"below minimum", "0.9.0", true},
		{"above maximum", "2.1.0", true},
		{"invalid format", "invalid", true},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			err := s.keeper.ValidateClientVersion(s.ctx, "version-test-client", tc.version)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
			}
		})
	}
}

// TestValidateScopePermission tests scope permission validation
func (s *KeeperTestSuite) TestValidateScopePermission() {
	// Register a client with specific allowed scopes
	client := types.NewApprovedClient(
		"scope-test-client",
		"Scope Test Client",
		"For testing scope permissions",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		[]string{"selfie", "document_front"},
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Test allowed scopes
	s.Require().NoError(s.keeper.ValidateScopePermission(s.ctx, "scope-test-client", "selfie"))
	s.Require().NoError(s.keeper.ValidateScopePermission(s.ctx, "scope-test-client", "document_front"))

	// Test disallowed scope
	err := s.keeper.ValidateScopePermission(s.ctx, "scope-test-client", "document_back")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "scope type not allowed")
}

// TestValidateClientSignature tests signature validation
func (s *KeeperTestSuite) TestValidateClientSignature() {
	// Register a client with ed25519 key
	client := types.NewApprovedClient(
		"sig-test-client",
		"Signature Test Client",
		"For testing signature validation",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Create a message and sign it
	message := []byte("test payload hash")
	messageHash := sha256.Sum256(message)
	signature := ed25519.Sign(s.testPrivateKey, messageHash[:])

	// Valid signature
	err := s.keeper.ValidateClientSignature(s.ctx, "sig-test-client", signature, messageHash[:])
	s.Require().NoError(err)

	// Invalid signature
	invalidSignature := make([]byte, 64)
	rand.Read(invalidSignature)
	err = s.keeper.ValidateClientSignature(s.ctx, "sig-test-client", invalidSignature, messageHash[:])
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "verification failed")

	// Non-existent client
	err = s.keeper.ValidateClientSignature(s.ctx, "nonexistent", signature, messageHash[:])
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not found")
}

// TestListClients tests listing all clients
func (s *KeeperTestSuite) TestListClients() {
	// Clear existing clients by creating new context (simulating fresh state)
	clients := s.keeper.ListClients(s.ctx)
	initialCount := len(clients)

	// Register multiple clients
	for i := 0; i < 3; i++ {
		client := types.NewApprovedClient(
			"list-test-client-"+string(rune('a'+i)),
			"List Test Client "+string(rune('A'+i)),
			"",
			s.testPublicKey,
			types.KeyTypeEd25519,
			"1.0.0",
			"",
			nil,
			testAdmin1,
			s.ctx.BlockTime(),
		)
		s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))
	}

	// Verify all clients are listed
	clients = s.keeper.ListClients(s.ctx)
	s.Require().Equal(initialCount+3, len(clients))
}

// TestListClientsByStatus tests listing clients by status
func (s *KeeperTestSuite) TestListClientsByStatus() {
	// Register clients with different statuses
	activeClient := types.NewApprovedClient(
		"status-active-client",
		"Active Client",
		"",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *activeClient))

	suspendClient := types.NewApprovedClient(
		"status-suspend-client",
		"To Be Suspended",
		"",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *suspendClient))
	s.Require().NoError(s.keeper.SuspendClient(s.ctx, "status-suspend-client", "Testing", testAdmin1))

	// List by status
	activeClients := s.keeper.ListClientsByStatus(s.ctx, types.ClientStatusActive)
	suspendedClients := s.keeper.ListClientsByStatus(s.ctx, types.ClientStatusSuspended)

	// Check that we have at least one of each
	foundActive := false
	for _, c := range activeClients {
		if c.ClientID == "status-active-client" {
			foundActive = true
			break
		}
	}
	s.Require().True(foundActive, "Active client not found in active list")

	foundSuspended := false
	for _, c := range suspendedClients {
		if c.ClientID == "status-suspend-client" {
			foundSuspended = true
			break
		}
	}
	s.Require().True(foundSuspended, "Suspended client not found in suspended list")
}

// TestGenesisExportImport tests genesis state export and import
func (s *KeeperTestSuite) TestGenesisExportImport() {
	// Register some clients
	client := types.NewApprovedClient(
		"genesis-test-client",
		"Genesis Test Client",
		"For testing genesis",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		[]string{"selfie"},
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Export genesis
	exported := s.keeper.ExportGenesis(s.ctx)
	s.Require().NotNil(exported)

	// Verify exported data contains our client
	found := false
	for _, c := range exported.ApprovedClients {
		if c.ClientID == "genesis-test-client" {
			found = true
			s.Require().Equal("Genesis Test Client", c.Name)
			break
		}
	}
	s.Require().True(found, "Client not found in exported genesis")
}

// TestVerifyUploadSignatures tests the comprehensive signature verification
func (s *KeeperTestSuite) TestVerifyUploadSignatures() {
	// Register a client
	client := types.NewApprovedClient(
		"upload-sig-test",
		"Upload Signature Test",
		"",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"2.0.0",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Create test data
	salt := make([]byte, 32)
	rand.Read(salt)
	payloadHash := sha256.Sum256([]byte("test payload"))

	// Compute signing payload (salt + payload hash)
	signingPayload := computeSigningPayload(salt, payloadHash[:])

	// Create valid client signature
	clientSignature := ed25519.Sign(s.testPrivateKey, signingPayload)

	// Create user signing payload
	userPayload := computeUserSigningPayload(clientSignature, payloadHash[:])

	// Create mock user signature (in real scenario this would be validated differently)
	userSignature := make([]byte, 64)
	rand.Read(userSignature)

	userAddr := sdk.AccAddress("test-user-address")

	// Test successful verification
	err := s.keeper.VerifyUploadSignatures(
		s.ctx,
		"upload-sig-test",
		"1.5.0",
		clientSignature,
		userSignature,
		payloadHash[:],
		salt,
		userAddr,
	)
	s.Require().NoError(err)

	// Test with invalid version
	err = s.keeper.VerifyUploadSignatures(
		s.ctx,
		"upload-sig-test",
		"3.0.0", // Outside allowed range
		clientSignature,
		userSignature,
		payloadHash[:],
		salt,
		userAddr,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "outside allowed range")

	// Test with non-existent client
	err = s.keeper.VerifyUploadSignatures(
		s.ctx,
		"nonexistent",
		"1.0.0",
		clientSignature,
		userSignature,
		payloadHash[:],
		salt,
		userAddr,
	)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not found")
}

// TestAuditHistory tests audit log functionality
func (s *KeeperTestSuite) TestAuditHistory() {
	// Register a client
	client := types.NewApprovedClient(
		"audit-test-client",
		"Audit Test Client",
		"",
		s.testPublicKey,
		types.KeyTypeEd25519,
		"1.0.0",
		"",
		nil,
		testAdmin1,
		s.ctx.BlockTime(),
	)
	s.Require().NoError(s.keeper.RegisterClient(s.ctx, *client))

	// Perform some operations
	s.Require().NoError(s.keeper.UpdateClient(s.ctx, "audit-test-client", "Updated Name", "", "", "", nil, testAdmin1))
	s.Require().NoError(s.keeper.SuspendClient(s.ctx, "audit-test-client", "Testing audit", testAdmin2))
	s.Require().NoError(s.keeper.ReactivateClient(s.ctx, "audit-test-client", "Audit complete", testAdmin1))

	// Check audit history
	history := s.keeper.GetAuditHistory(s.ctx, "audit-test-client")
	s.Require().GreaterOrEqual(len(history), 3, "Should have at least 3 audit entries")

	// Verify audit entries contain expected actions
	actions := make(map[string]bool)
	for _, entry := range history {
		actions[entry.Action] = true
	}
	s.Require().True(actions["update"], "Should have update action")
	s.Require().True(actions["suspend"], "Should have suspend action")
	s.Require().True(actions["reactivate"], "Should have reactivate action")
}
