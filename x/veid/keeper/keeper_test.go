package keeper_test

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"
	"time"

	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// Test constants for address and scope identifiers
const (
	testAddress1        = testAddress1
	testScopeID1        = testScopeID1
	testScopeToRevoke   = testScopeToRevoke
	testScopeVerifyTest = testScopeVerifyTest
)

type KeeperTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) SetupTest() {
	// Create codec
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	s.cdc = codec.NewProtoCodec(interfaceRegistry)

	// Create store key
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	// Create context with store
	s.ctx = s.createContextWithStore(storeKey)

	// Create keeper
	s.keeper = keeper.NewKeeper(s.cdc, storeKey, "authority")

	// Set default params
	err := s.keeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func (s *KeeperTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	// This is a simplified context creation for testing
	// In production, this would use a full CMS store
	db := runtime.NewKVStoreService(storeKey)
	_ = db // Use in actual implementation
	
	// For unit tests, we create a minimal context
	ctx := sdk.Context{}.WithBlockTime(time.Now()).WithBlockHeight(1)
	return ctx
}

func (s *KeeperTestSuite) generateSalt() []byte {
	salt := make([]byte, 32)
	_, _ = rand.Read(salt)
	return salt
}

func (s *KeeperTestSuite) generateSignature() []byte {
	sig := make([]byte, 64)
	_, _ = rand.Read(sig)
	return sig
}

func (s *KeeperTestSuite) createTestPayload() encryptiontypes.EncryptedPayloadEnvelope {
	nonce := make([]byte, 24)
	_, _ = rand.Read(nonce)
	
	ciphertext := make([]byte, 128)
	_, _ = rand.Read(ciphertext)
	
	pubKey := make([]byte, 32)
	_, _ = rand.Read(pubKey)
	
	return encryptiontypes.EncryptedPayloadEnvelope{
		Version:         1,
		AlgorithmID:     "X25519-XSALSA20-POLY1305",
		RecipientKeyIDs: []string{"recipient1"},
		Nonce:           nonce,
		Ciphertext:      ciphertext,
		SenderPubKey:    pubKey,
	}
}

func (s *KeeperTestSuite) createTestUploadMetadata() types.UploadMetadata {
	salt := s.generateSalt()
	payload := s.createTestPayload()
	payloadHash := sha256.Sum256(payload.Ciphertext)
	
	return types.UploadMetadata{
		Salt:              salt,
		SaltHash:          types.ComputeSaltHash(salt),
		DeviceFingerprint: "test-device-fp",
		ClientID:          "test-client",
		ClientSignature:   s.generateSignature(),
		UserSignature:     s.generateSignature(),
		PayloadHash:       payloadHash[:],
	}
}

// Test: Scope upload with valid signatures
func (s *KeeperTestSuite) TestUploadScopeValid() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Register an approved client first (relaxed params for testing)
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Create test scope
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		testScopeID1,
		types.ScopeTypeIDDocument,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	
	// Upload scope should succeed
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)
	
	// Verify scope was stored
	storedScope, found := s.keeper.GetScope(s.ctx, address, testScopeID1)
	s.Require().True(found)
	s.Require().Equal(testScopeID1, storedScope.ScopeID)
	s.Require().Equal(types.ScopeTypeIDDocument, storedScope.ScopeType)
	s.Require().Equal(types.VerificationStatusPending, storedScope.Status)
	
	// Verify identity record was created/updated
	record, found := s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().True(found)
	s.Require().Len(record.ScopeRefs, 1)
	s.Require().Equal(testScopeID1, record.ScopeRefs[0].ScopeID)
}

// Test: Scope upload with invalid client signature (reject)
func (s *KeeperTestSuite) TestUploadScopeInvalidClientSignature() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Enable client signature requirement
	params := types.DefaultParams()
	params.RequireClientSignature = true
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Create test scope with unknown client
	metadata := s.createTestUploadMetadata()
	metadata.ClientID = "unknown-client"
	
	scope := types.NewIdentityScope(
		"scope-invalid-client",
		types.ScopeTypeIDDocument,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	
	// Upload should fail - client not approved
	err = s.keeper.ValidateUploadSignatures(s.ctx, address, &metadata)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "not approved")
}

// Test: Scope upload with approved client
func (s *KeeperTestSuite) TestUploadScopeWithApprovedClient() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Register an approved client
	client := types.NewApprovedClient(
		"approved-client",
		"Test Client",
		make([]byte, 32), // Ed25519 public key
		"ed25519",
		time.Now().Unix(),
	)
	err := s.keeper.SetApprovedClient(s.ctx, *client)
	s.Require().NoError(err)
	
	// Verify client is approved
	isApproved := s.keeper.IsClientApproved(s.ctx, "approved-client")
	s.Require().True(isApproved)
	
	// Unknown client should not be approved
	isApproved = s.keeper.IsClientApproved(s.ctx, "unknown-client")
	s.Require().False(isApproved)
}

// Test: Scope revocation
func (s *KeeperTestSuite) TestRevokeScope() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Upload a scope first
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		testScopeToRevoke,
		types.ScopeTypeSelfie,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)
	
	// Revoke the scope
	err = s.keeper.RevokeScope(s.ctx, address, testScopeToRevoke, "user requested")
	s.Require().NoError(err)
	
	// Verify scope is revoked
	revokedScope, found := s.keeper.GetScope(s.ctx, address, testScopeToRevoke)
	s.Require().True(found)
	s.Require().True(revokedScope.Revoked)
	s.Require().NotNil(revokedScope.RevokedAt)
	s.Require().Equal("user requested", revokedScope.RevokedReason)
	
	// Should not be able to revoke again
	err = s.keeper.RevokeScope(s.ctx, address, testScopeToRevoke, "second revoke")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "already revoked")
}

// Test: Verification status transitions
func (s *KeeperTestSuite) TestVerificationStatusTransitions() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Upload a scope
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		testScopeVerifyTest,
		types.ScopeTypeFaceVideo,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)
	
	// Initial status should be Pending
	storedScope, _ := s.keeper.GetScope(s.ctx, address, testScopeVerifyTest)
	s.Require().Equal(types.VerificationStatusPending, storedScope.Status)
	
	// Transition to InProgress
	err = s.keeper.UpdateVerificationStatus(s.ctx, address, testScopeVerifyTest, 
		types.VerificationStatusInProgress, "verification started", "validator1")
	s.Require().NoError(err)
	
	storedScope, _ = s.keeper.GetScope(s.ctx, address, testScopeVerifyTest)
	s.Require().Equal(types.VerificationStatusInProgress, storedScope.Status)
	
	// Transition to Verified
	err = s.keeper.UpdateVerificationStatus(s.ctx, address, testScopeVerifyTest,
		types.VerificationStatusVerified, "verification complete", "validator1")
	s.Require().NoError(err)
	
	storedScope, _ = s.keeper.GetScope(s.ctx, address, testScopeVerifyTest)
	s.Require().Equal(types.VerificationStatusVerified, storedScope.Status)
	s.Require().NotNil(storedScope.VerifiedAt)
}

// Test: Invalid verification status transition
func (s *KeeperTestSuite) TestInvalidStatusTransition() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Upload a scope
	metadata := s.createTestUploadMetadata()
	scope := types.NewIdentityScope(
		"scope-invalid-transition",
		types.ScopeTypeBiometric,
		s.createTestPayload(),
		metadata,
		time.Now(),
	)
	
	err = s.keeper.UploadScope(s.ctx, address, scope)
	s.Require().NoError(err)
	
	// Try to transition directly from Pending to Verified (should fail)
	err = s.keeper.UpdateVerificationStatus(s.ctx, address, "scope-invalid-transition",
		types.VerificationStatusVerified, "skip steps", "validator1")
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "cannot transition")
}

// Test: Score update and tier calculation
func (s *KeeperTestSuite) TestScoreUpdateAndTier() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Create identity record
	_, err := s.keeper.CreateIdentityRecord(s.ctx, address)
	s.Require().NoError(err)
	
	// Initial tier should be Unverified
	record, _ := s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().Equal(types.IdentityTierUnverified, record.Tier)
	s.Require().Equal(uint32(0), record.CurrentScore)
	
	// Update score to 25 (Basic tier)
	err = s.keeper.UpdateScore(s.ctx, address, 25, "v1.0")
	s.Require().NoError(err)
	
	record, _ = s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().Equal(types.IdentityTierBasic, record.Tier)
	s.Require().Equal(uint32(25), record.CurrentScore)
	
	// Update score to 50 (Standard tier)
	err = s.keeper.UpdateScore(s.ctx, address, 50, "v1.0")
	s.Require().NoError(err)
	
	record, _ = s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().Equal(types.IdentityTierStandard, record.Tier)
	
	// Update score to 75 (Verified tier)
	err = s.keeper.UpdateScore(s.ctx, address, 75, "v1.0")
	s.Require().NoError(err)
	
	record, _ = s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().Equal(types.IdentityTierVerified, record.Tier)
	
	// Update score to 90 (Trusted tier)
	err = s.keeper.UpdateScore(s.ctx, address, 90, "v1.0")
	s.Require().NoError(err)
	
	record, _ = s.keeper.GetIdentityRecord(s.ctx, address)
	s.Require().Equal(types.IdentityTierTrusted, record.Tier)
}

// Test: Salt binding - prevent salt reuse
func (s *KeeperTestSuite) TestSaltBindingPreventReuse() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Disable signature requirements for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Create a specific salt
	salt := s.generateSalt()
	
	// First upload with this salt should succeed
	metadata1 := s.createTestUploadMetadata()
	metadata1.Salt = salt
	metadata1.SaltHash = types.ComputeSaltHash(salt)
	
	scope1 := types.NewIdentityScope(
		"scope-salt-1",
		types.ScopeTypeEmailProof,
		s.createTestPayload(),
		metadata1,
		time.Now(),
	)
	
	err = s.keeper.UploadScope(s.ctx, address, scope1)
	s.Require().NoError(err)
	
	// Second upload with same salt should fail
	metadata2 := s.createTestUploadMetadata()
	metadata2.Salt = salt
	metadata2.SaltHash = types.ComputeSaltHash(salt)
	
	scope2 := types.NewIdentityScope(
		"scope-salt-2",
		types.ScopeTypeSMSProof,
		s.createTestPayload(),
		metadata2,
		time.Now(),
	)
	
	err = s.keeper.UploadScope(s.ctx, address, scope2)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "salt")
}

// Test: Max scopes per account
func (s *KeeperTestSuite) TestMaxScopesPerAccount() {
	address := sdk.AccAddress([]byte(testAddress1))
	
	// Set low max scopes for testing
	params := types.DefaultParams()
	params.RequireClientSignature = false
	params.RequireUserSignature = false
	params.MaxScopesPerAccount = 2
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
	
	// Upload first scope
	err = s.keeper.UploadScope(s.ctx, address, types.NewIdentityScope(
		"scope-max-1",
		types.ScopeTypeIDDocument,
		s.createTestPayload(),
		s.createTestUploadMetadata(),
		time.Now(),
	))
	s.Require().NoError(err)
	
	// Upload second scope
	err = s.keeper.UploadScope(s.ctx, address, types.NewIdentityScope(
		"scope-max-2",
		types.ScopeTypeSelfie,
		s.createTestPayload(),
		s.createTestUploadMetadata(),
		time.Now(),
	))
	s.Require().NoError(err)
	
	// Third scope should fail
	err = s.keeper.UploadScope(s.ctx, address, types.NewIdentityScope(
		"scope-max-3",
		types.ScopeTypeFaceVideo,
		s.createTestPayload(),
		s.createTestUploadMetadata(),
		time.Now(),
	))
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "maximum")
}

// Additional test function for running individual tests
func TestScopeTypeValidation(t *testing.T) {
	// Valid scope types
	require.True(t, types.IsValidScopeType(types.ScopeTypeIDDocument))
	require.True(t, types.IsValidScopeType(types.ScopeTypeSelfie))
	require.True(t, types.IsValidScopeType(types.ScopeTypeFaceVideo))
	require.True(t, types.IsValidScopeType(types.ScopeTypeBiometric))
	require.True(t, types.IsValidScopeType(types.ScopeTypeSSOMetadata))
	require.True(t, types.IsValidScopeType(types.ScopeTypeEmailProof))
	require.True(t, types.IsValidScopeType(types.ScopeTypeSMSProof))
	require.True(t, types.IsValidScopeType(types.ScopeTypeDomainVerify))
	
	// Invalid scope type
	require.False(t, types.IsValidScopeType(types.ScopeType("invalid")))
}

func TestVerificationStatusValidation(t *testing.T) {
	// Valid statuses
	require.True(t, types.IsValidVerificationStatus(types.VerificationStatusUnknown))
	require.True(t, types.IsValidVerificationStatus(types.VerificationStatusPending))
	require.True(t, types.IsValidVerificationStatus(types.VerificationStatusInProgress))
	require.True(t, types.IsValidVerificationStatus(types.VerificationStatusVerified))
	require.True(t, types.IsValidVerificationStatus(types.VerificationStatusRejected))
	require.True(t, types.IsValidVerificationStatus(types.VerificationStatusExpired))
	
	// Invalid status
	require.False(t, types.IsValidVerificationStatus(types.VerificationStatus("invalid")))
}

func TestVerificationStatusTransitions(t *testing.T) {
	// Valid transitions from Pending
	require.True(t, types.VerificationStatusPending.CanTransitionTo(types.VerificationStatusInProgress))
	require.True(t, types.VerificationStatusPending.CanTransitionTo(types.VerificationStatusRejected))
	
	// Invalid transitions from Pending
	require.False(t, types.VerificationStatusPending.CanTransitionTo(types.VerificationStatusVerified))
	
	// Valid transitions from InProgress
	require.True(t, types.VerificationStatusInProgress.CanTransitionTo(types.VerificationStatusVerified))
	require.True(t, types.VerificationStatusInProgress.CanTransitionTo(types.VerificationStatusRejected))
	
	// Valid transitions from Verified
	require.True(t, types.VerificationStatusVerified.CanTransitionTo(types.VerificationStatusExpired))
	
	// Invalid transitions from Verified
	require.False(t, types.VerificationStatusVerified.CanTransitionTo(types.VerificationStatusPending))
}

func TestIdentityTierCalculation(t *testing.T) {
	require.Equal(t, types.IdentityTierUnverified, types.ComputeTierFromScore(0))
	require.Equal(t, types.IdentityTierBasic, types.ComputeTierFromScore(1))
	require.Equal(t, types.IdentityTierBasic, types.ComputeTierFromScore(29))
	require.Equal(t, types.IdentityTierStandard, types.ComputeTierFromScore(30))
	require.Equal(t, types.IdentityTierStandard, types.ComputeTierFromScore(59))
	require.Equal(t, types.IdentityTierVerified, types.ComputeTierFromScore(60))
	require.Equal(t, types.IdentityTierVerified, types.ComputeTierFromScore(84))
	require.Equal(t, types.IdentityTierTrusted, types.ComputeTierFromScore(85))
	require.Equal(t, types.IdentityTierTrusted, types.ComputeTierFromScore(100))
}

func TestScopeTypeWeights(t *testing.T) {
	// ID document has highest weight
	require.Equal(t, uint32(30), types.ScopeTypeWeight(types.ScopeTypeIDDocument))
	
	// Face video is high weight
	require.Equal(t, uint32(25), types.ScopeTypeWeight(types.ScopeTypeFaceVideo))
	
	// SSO metadata is lowest weight
	require.Equal(t, uint32(5), types.ScopeTypeWeight(types.ScopeTypeSSOMetadata))
	
	// Unknown type has zero weight
	require.Equal(t, uint32(0), types.ScopeTypeWeight(types.ScopeType("unknown")))
}

func TestUploadMetadataValidation(t *testing.T) {
	// Valid metadata
	salt := make([]byte, 32)
	_, _ = rand.Read(salt)
	payloadHash := sha256.Sum256([]byte("test payload"))
	
	metadata := types.UploadMetadata{
		Salt:              salt,
		SaltHash:          types.ComputeSaltHash(salt),
		DeviceFingerprint: "test-device",
		ClientID:          "test-client",
		ClientSignature:   make([]byte, 64),
		UserSignature:     make([]byte, 64),
		PayloadHash:       payloadHash[:],
	}
	
	err := metadata.Validate()
	require.NoError(t, err)
	
	// Invalid - empty salt
	invalidMetadata := metadata
	invalidMetadata.Salt = nil
	err = invalidMetadata.Validate()
	require.Error(t, err)
	
	// Invalid - short salt
	invalidMetadata = metadata
	invalidMetadata.Salt = make([]byte, 8)
	err = invalidMetadata.Validate()
	require.Error(t, err)
	
	// Invalid - empty client ID
	invalidMetadata = metadata
	invalidMetadata.ClientID = ""
	err = invalidMetadata.Validate()
	require.Error(t, err)
	
	// Invalid - wrong payload hash size
	invalidMetadata = metadata
	invalidMetadata.PayloadHash = make([]byte, 16)
	err = invalidMetadata.Validate()
	require.Error(t, err)
}

func TestParamsValidation(t *testing.T) {
	// Valid params
	params := types.DefaultParams()
	err := params.Validate()
	require.NoError(t, err)
	
	// Invalid - zero max scopes
	invalidParams := params
	invalidParams.MaxScopesPerAccount = 0
	err = invalidParams.Validate()
	require.Error(t, err)
	
	// Invalid - salt max < salt min
	invalidParams = params
	invalidParams.SaltMinBytes = 64
	invalidParams.SaltMaxBytes = 32
	err = invalidParams.Validate()
	require.Error(t, err)
}

