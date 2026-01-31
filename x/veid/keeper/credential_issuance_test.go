package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

// ============================================================================
// Test Suite Setup
// ============================================================================

type CredentialIssuanceTestSuite struct {
	suite.Suite
	ctx           sdk.Context
	keeper        keeper.Keeper
	cdc           codec.Codec
	stateStore    store.CommitMultiStore
	issuerPubKey  ed25519.PublicKey
	issuerPrivKey ed25519.PrivateKey
	issuerValAddr sdk.ValAddress
	subjectAddr   sdk.AccAddress
}

func TestCredentialIssuanceTestSuite(t *testing.T) {
	suite.Run(t, new(CredentialIssuanceTestSuite))
}

func (s *CredentialIssuanceTestSuite) SetupTest() {
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

	// Generate issuer key pair
	s.issuerPubKey, s.issuerPrivKey, err = ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	// Create test addresses
	s.issuerValAddr = sdk.ValAddress([]byte("test-validator-addr"))
	s.subjectAddr = sdk.AccAddress([]byte("test-subject-address"))
}

func (s *CredentialIssuanceTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	s.stateStore = stateStore

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// TearDownTest closes the IAVL store to stop background pruning goroutines
func (s *CredentialIssuanceTestSuite) TearDownTest() {
	CloseStoreIfNeeded(s.stateStore)
}

// ============================================================================
// Credential Issuance Tests
// ============================================================================

func (s *CredentialIssuanceTestSuite) TestIssueCredential_Success() {
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365, // 1 year
		Claims: map[string]interface{}{
			"scopes_verified": 3,
			"model_version":   "v1.0.0",
		},
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)

	s.Require().NoError(err)
	s.Require().NotNil(credential)

	// Verify credential structure
	s.Assert().Contains(credential.Context, types.ContextW3CCredentials)
	s.Assert().Contains(credential.Context, types.ContextVirtEngine)
	s.Assert().Contains(credential.Type, types.TypeVerifiableCredential)
	s.Assert().Contains(credential.Type, types.TypeVEIDCredential)
	s.Assert().NotEmpty(credential.ID)

	// Verify issuer
	s.Assert().Contains(credential.Issuer.ID, s.issuerValAddr.String())
	s.Assert().Equal("Test Validator", credential.Issuer.Name)

	// Verify subject
	s.Assert().Contains(credential.CredentialSubject.ID, s.subjectAddr.String())
	s.Assert().Equal(types.TypeIdentityVerification, credential.CredentialSubject.VerificationType)
	s.Assert().Equal(3, credential.CredentialSubject.VerificationLevel)
	s.Assert().Equal(0.85, credential.CredentialSubject.TrustScore)

	// Verify proof exists
	s.Assert().NotEmpty(credential.Proof.ProofValue)
	s.Assert().Equal(types.ProofTypeEd25519Signature2020, credential.Proof.Type)
	s.Assert().Equal(types.ProofPurposeAssertion, credential.Proof.ProofPurpose)

	// Verify credential was stored
	record, found := s.keeper.GetCredentialRecord(s.ctx, credential.ID)
	s.Require().True(found)
	s.Assert().Equal(credential.ID, record.CredentialID)
	s.Assert().Equal(s.subjectAddr.String(), record.SubjectAddress)
	s.Assert().Equal(types.CredentialStatusActive, record.Status)
}

func (s *CredentialIssuanceTestSuite) TestIssueCredential_InvalidRequest() {
	testCases := []struct {
		name    string
		request types.CredentialIssuanceRequest
		errMsg  string
	}{
		{
			name: "missing subject address",
			request: types.CredentialIssuanceRequest{
				VerificationType:  types.TypeIdentityVerification,
				VerificationLevel: 3,
				TrustScore:        0.85,
			},
			errMsg: "subject address is required",
		},
		{
			name: "invalid subject address",
			request: types.CredentialIssuanceRequest{
				SubjectAddress:    "invalid-address",
				VerificationType:  types.TypeIdentityVerification,
				VerificationLevel: 3,
				TrustScore:        0.85,
			},
			errMsg: "invalid subject address",
		},
		{
			name: "missing verification type",
			request: types.CredentialIssuanceRequest{
				SubjectAddress:    s.subjectAddr.String(),
				VerificationLevel: 3,
				TrustScore:        0.85,
			},
			errMsg: "verification type is required",
		},
		{
			name: "invalid verification level",
			request: types.CredentialIssuanceRequest{
				SubjectAddress:    s.subjectAddr.String(),
				VerificationType:  types.TypeIdentityVerification,
				VerificationLevel: 5, // Max is 4
				TrustScore:        0.85,
			},
			errMsg: "verification level must be 0-4",
		},
		{
			name: "invalid trust score",
			request: types.CredentialIssuanceRequest{
				SubjectAddress:    s.subjectAddr.String(),
				VerificationType:  types.TypeIdentityVerification,
				VerificationLevel: 3,
				TrustScore:        1.5, // Max is 1.0
			},
			errMsg: "trust score must be 0.0-1.0",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			_, err := s.keeper.IssueCredential(
				s.ctx,
				tc.request,
				s.issuerValAddr,
				"Test Validator",
				s.issuerPrivKey,
			)
			s.Require().Error(err)
			s.Assert().Contains(err.Error(), tc.errMsg)
		})
	}
}

func (s *CredentialIssuanceTestSuite) TestIssueCredential_WithClaims() {
	claims := map[string]interface{}{
		"country":        "US",
		"document_type":  "passport",
		"document_valid": true,
		"confidence":     0.95,
	}

	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeDocumentVerification,
		VerificationLevel: 4,
		TrustScore:        0.95,
		ValidityDuration:  24 * time.Hour * 30, // 30 days
		Claims:            claims,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)

	s.Require().NoError(err)
	s.Assert().Equal(claims, credential.CredentialSubject.Claims)
}

// ============================================================================
// Credential Verification Tests
// ============================================================================

func (s *CredentialIssuanceTestSuite) TestVerifyCredential_Success() {
	// Issue a credential first
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Verify the credential
	err = s.keeper.VerifyCredential(s.ctx, credential, s.issuerPubKey)
	s.Require().NoError(err)
}

func (s *CredentialIssuanceTestSuite) TestVerifyCredential_InvalidSignature() {
	// Issue a credential
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Generate a different key pair
	wrongPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	// Verify with wrong public key should fail
	err = s.keeper.VerifyCredential(s.ctx, credential, wrongPubKey)
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "signature verification failed")
}

func (s *CredentialIssuanceTestSuite) TestVerifyCredential_ExpiredCredential() {
	// Issue a credential with very short validity
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  1 * time.Nanosecond,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Advance time past expiration
	futureCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Hour))

	// Verify should fail due to expiration
	err = s.keeper.VerifyCredential(futureCtx, credential, s.issuerPubKey)
	s.Require().Error(err)
	s.Assert().Equal(types.ErrCredentialExpired, err)
}

func (s *CredentialIssuanceTestSuite) TestVerifyCredential_TamperedCredential() {
	// Issue a credential
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Tamper with the credential
	credential.CredentialSubject.TrustScore = 1.0 // Changed from 0.85

	// Verify should fail
	err = s.keeper.VerifyCredential(s.ctx, credential, s.issuerPubKey)
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "signature verification failed")
}

// ============================================================================
// Credential Revocation Tests
// ============================================================================

func (s *CredentialIssuanceTestSuite) TestRevokeCredential_Success() {
	// Issue a credential first
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Revoke the credential
	err = s.keeper.RevokeCredential(
		s.ctx,
		credential.ID,
		s.issuerValAddr,
		"Fraudulent activity detected",
	)
	s.Require().NoError(err)

	// Verify credential is revoked
	s.Assert().True(s.keeper.IsCredentialRevoked(s.ctx, credential.ID))

	// Verify record is updated
	record, found := s.keeper.GetCredentialRecord(s.ctx, credential.ID)
	s.Require().True(found)
	s.Assert().Equal(types.CredentialStatusRevoked, record.Status)
	s.Assert().NotNil(record.RevokedAt)
	s.Assert().Equal("Fraudulent activity detected", record.RevocationReason)
}

func (s *CredentialIssuanceTestSuite) TestRevokeCredential_Unauthorized() {
	// Issue a credential
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Try to revoke with different issuer
	wrongIssuer := sdk.ValAddress([]byte("wrong-validator-addr"))
	err = s.keeper.RevokeCredential(
		s.ctx,
		credential.ID,
		wrongIssuer,
		"Unauthorized attempt",
	)
	s.Require().Error(err)
	s.Assert().Contains(err.Error(), "only issuer can revoke credential")
}

func (s *CredentialIssuanceTestSuite) TestRevokeCredential_AlreadyRevoked() {
	// Issue and revoke a credential
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// First revocation should succeed
	err = s.keeper.RevokeCredential(s.ctx, credential.ID, s.issuerValAddr, "First revocation")
	s.Require().NoError(err)

	// Second revocation should fail
	err = s.keeper.RevokeCredential(s.ctx, credential.ID, s.issuerValAddr, "Second revocation")
	s.Require().Error(err)
	s.Assert().Equal(types.ErrCredentialAlreadyRevoked, err)
}

func (s *CredentialIssuanceTestSuite) TestRevokeCredential_NotFound() {
	err := s.keeper.RevokeCredential(
		s.ctx,
		"non-existent-credential",
		s.issuerValAddr,
		"Reason",
	)
	s.Require().Error(err)
	s.Assert().Equal(types.ErrCredentialNotFound, err)
}

func (s *CredentialIssuanceTestSuite) TestVerifyCredential_RevokedCredential() {
	// Issue a credential
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Revoke it
	err = s.keeper.RevokeCredential(s.ctx, credential.ID, s.issuerValAddr, "Revoked for testing")
	s.Require().NoError(err)

	// Verification should fail
	err = s.keeper.VerifyCredential(s.ctx, credential, s.issuerPubKey)
	s.Require().Error(err)
	s.Assert().Equal(types.ErrCredentialRevoked, err)
}

// ============================================================================
// Credential Query Tests
// ============================================================================

func (s *CredentialIssuanceTestSuite) TestListCredentialsForSubject() {
	// Issue multiple credentials for the same subject
	for i := 0; i < 3; i++ {
		request := types.CredentialIssuanceRequest{
			SubjectAddress:    s.subjectAddr.String(),
			VerificationType:  types.TypeIdentityVerification,
			VerificationLevel: i + 1,
			TrustScore:        float64(i+1) * 0.25,
			ValidityDuration:  24 * time.Hour * 365,
		}

		_, err := s.keeper.IssueCredential(
			s.ctx,
			request,
			s.issuerValAddr,
			"Test Validator",
			s.issuerPrivKey,
		)
		s.Require().NoError(err)

		// Advance block time/height to get unique IDs
		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Second))
	}

	// List credentials
	credentials, err := s.keeper.ListCredentialsForSubject(s.ctx, s.subjectAddr)
	s.Require().NoError(err)
	s.Assert().Len(credentials, 3)
}

func (s *CredentialIssuanceTestSuite) TestListActiveCredentialsForSubject() {
	// Issue credentials
	for i := 0; i < 3; i++ {
		request := types.CredentialIssuanceRequest{
			SubjectAddress:    s.subjectAddr.String(),
			VerificationType:  types.TypeIdentityVerification,
			VerificationLevel: i + 1,
			TrustScore:        float64(i+1) * 0.25,
			ValidityDuration:  24 * time.Hour * 365,
		}

		credential, err := s.keeper.IssueCredential(
			s.ctx,
			request,
			s.issuerValAddr,
			"Test Validator",
			s.issuerPrivKey,
		)
		s.Require().NoError(err)

		// Revoke the first credential
		if i == 0 {
			err = s.keeper.RevokeCredential(s.ctx, credential.ID, s.issuerValAddr, "Revoked")
			s.Require().NoError(err)
		}

		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Second))
	}

	// List active credentials (should be 2, not 3)
	credentials, err := s.keeper.ListActiveCredentialsForSubject(s.ctx, s.subjectAddr)
	s.Require().NoError(err)
	s.Assert().Len(credentials, 2)
}

func (s *CredentialIssuanceTestSuite) TestListCredentialsForIssuer() {
	// Issue credentials
	for i := 0; i < 2; i++ {
		request := types.CredentialIssuanceRequest{
			SubjectAddress:    s.subjectAddr.String(),
			VerificationType:  types.TypeIdentityVerification,
			VerificationLevel: 3,
			TrustScore:        0.85,
			ValidityDuration:  24 * time.Hour * 365,
		}

		_, err := s.keeper.IssueCredential(
			s.ctx,
			request,
			s.issuerValAddr,
			"Test Validator",
			s.issuerPrivKey,
		)
		s.Require().NoError(err)

		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Second))
	}

	// List by issuer
	credentials, err := s.keeper.ListCredentialsForIssuer(s.ctx, s.issuerValAddr)
	s.Require().NoError(err)
	s.Assert().Len(credentials, 2)
}

func (s *CredentialIssuanceTestSuite) TestListCredentialsByType() {
	// Issue credentials of different types
	types1 := []string{types.TypeIdentityVerification, types.TypeDocumentVerification}

	for _, vType := range types1 {
		request := types.CredentialIssuanceRequest{
			SubjectAddress:    s.subjectAddr.String(),
			VerificationType:  vType,
			VerificationLevel: 3,
			TrustScore:        0.85,
			ValidityDuration:  24 * time.Hour * 365,
		}

		_, err := s.keeper.IssueCredential(
			s.ctx,
			request,
			s.issuerValAddr,
			"Test Validator",
			s.issuerPrivKey,
		)
		s.Require().NoError(err)

		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Second))
	}

	// List by type - VEIDCredential should be in all
	credentials, err := s.keeper.ListCredentialsByType(s.ctx, types.TypeVEIDCredential)
	s.Require().NoError(err)
	s.Assert().Len(credentials, 2)

	// List by specific type
	docCredentials, err := s.keeper.ListCredentialsByType(s.ctx, types.TypeDocumentVerification)
	s.Require().NoError(err)
	s.Assert().Len(docCredentials, 1)
}

func (s *CredentialIssuanceTestSuite) TestCountCredentialsForSubject() {
	// Issue credentials
	for i := 0; i < 5; i++ {
		request := types.CredentialIssuanceRequest{
			SubjectAddress:    s.subjectAddr.String(),
			VerificationType:  types.TypeIdentityVerification,
			VerificationLevel: 3,
			TrustScore:        0.85,
			ValidityDuration:  24 * time.Hour * 365,
		}

		_, err := s.keeper.IssueCredential(
			s.ctx,
			request,
			s.issuerValAddr,
			"Test Validator",
			s.issuerPrivKey,
		)
		s.Require().NoError(err)

		s.ctx = s.ctx.WithBlockHeight(s.ctx.BlockHeight() + 1)
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(time.Second))
	}

	count := s.keeper.CountCredentialsForSubject(s.ctx, s.subjectAddr)
	s.Assert().Equal(5, count)
}

// ============================================================================
// JSON-LD Serialization Tests
// ============================================================================

func (s *CredentialIssuanceTestSuite) TestCredentialJSONSerialization() {
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
		Claims: map[string]interface{}{
			"document_verified": true,
		},
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Serialize to JSON
	jsonBytes, err := credential.ToJSON()
	s.Require().NoError(err)
	s.Assert().NotEmpty(jsonBytes)

	// Verify JSON structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	s.Require().NoError(err)

	// Verify @context is present
	s.Assert().Contains(parsed, "@context")

	// Deserialize back
	deserialized, err := types.FromJSON(jsonBytes)
	s.Require().NoError(err)
	s.Assert().Equal(credential.ID, deserialized.ID)
	s.Assert().Equal(credential.Issuer.ID, deserialized.Issuer.ID)
}

func (s *CredentialIssuanceTestSuite) TestCredentialHash() {
	request := types.CredentialIssuanceRequest{
		SubjectAddress:    s.subjectAddr.String(),
		VerificationType:  types.TypeIdentityVerification,
		VerificationLevel: 3,
		TrustScore:        0.85,
		ValidityDuration:  24 * time.Hour * 365,
	}

	credential, err := s.keeper.IssueCredential(
		s.ctx,
		request,
		s.issuerValAddr,
		"Test Validator",
		s.issuerPrivKey,
	)
	s.Require().NoError(err)

	// Hash should be deterministic
	hash1, err := credential.Hash()
	s.Require().NoError(err)

	hash2, err := credential.Hash()
	s.Require().NoError(err)

	s.Assert().Equal(hash1, hash2)
	s.Assert().Len(hash1, 32) // SHA256
}

// ============================================================================
// Standalone Unit Tests
// ============================================================================

func TestVerifiableCredential_Validate(t *testing.T) {
	now := time.Now()
	expiry := now.Add(24 * time.Hour)

	tests := []struct {
		name      string
		vc        types.VerifiableCredential
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid credential",
			vc: types.VerifiableCredential{
				Context: []string{types.ContextW3CCredentials, types.ContextVirtEngine},
				ID:      "urn:uuid:test-credential",
				Type:    []string{types.TypeVerifiableCredential, types.TypeVEIDCredential},
				Issuer: types.CredentialIssuer{
					ID:   "did:virtengine:validator1",
					Name: "Test Validator",
				},
				IssuanceDate:   now,
				ExpirationDate: &expiry,
				CredentialSubject: types.CredentialSubject{
					ID:                "did:virtengine:user1",
					VerificationType:  types.TypeIdentityVerification,
					VerificationLevel: 3,
					TrustScore:        0.85,
				},
			},
			expectErr: false,
		},
		{
			name: "missing context",
			vc: types.VerifiableCredential{
				Context: []string{},
				ID:      "test",
				Type:    []string{types.TypeVerifiableCredential},
			},
			expectErr: true,
			errMsg:    "context is required",
		},
		{
			name: "missing W3C context",
			vc: types.VerifiableCredential{
				Context: []string{types.ContextVirtEngine},
				ID:      "test",
				Type:    []string{types.TypeVerifiableCredential},
			},
			expectErr: true,
			errMsg:    "W3C credentials context is required",
		},
		{
			name: "missing VerifiableCredential type",
			vc: types.VerifiableCredential{
				Context: []string{types.ContextW3CCredentials},
				ID:      "test",
				Type:    []string{types.TypeVEIDCredential},
			},
			expectErr: true,
			errMsg:    "VerifiableCredential type is required",
		},
		{
			name: "expiration before issuance",
			vc: types.VerifiableCredential{
				Context: []string{types.ContextW3CCredentials},
				ID:      "test",
				Type:    []string{types.TypeVerifiableCredential},
				Issuer: types.CredentialIssuer{
					ID: "did:virtengine:validator1",
				},
				IssuanceDate:   now,
				ExpirationDate: func() *time.Time { t := now.Add(-1 * time.Hour); return &t }(),
				CredentialSubject: types.CredentialSubject{
					ID:                "did:virtengine:user1",
					VerificationType:  "test",
					VerificationLevel: 1,
					TrustScore:        0.5,
				},
			},
			expectErr: true,
			errMsg:    "expiration date must be after issuance date",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.vc.Validate()
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCredentialSubject_Validate(t *testing.T) {
	tests := []struct {
		name      string
		cs        types.CredentialSubject
		expectErr bool
		errMsg    string
	}{
		{
			name: "valid subject",
			cs: types.CredentialSubject{
				ID:                "did:virtengine:user1",
				VerificationType:  "IdentityVerification",
				VerificationLevel: 3,
				TrustScore:        0.85,
			},
			expectErr: false,
		},
		{
			name: "missing ID",
			cs: types.CredentialSubject{
				VerificationType:  "test",
				VerificationLevel: 1,
				TrustScore:        0.5,
			},
			expectErr: true,
			errMsg:    "subject ID is required",
		},
		{
			name: "verification level too high",
			cs: types.CredentialSubject{
				ID:                "did:virtengine:user1",
				VerificationType:  "test",
				VerificationLevel: 5,
				TrustScore:        0.5,
			},
			expectErr: true,
			errMsg:    "verification level must be 0-4",
		},
		{
			name: "trust score too high",
			cs: types.CredentialSubject{
				ID:                "did:virtengine:user1",
				VerificationType:  "test",
				VerificationLevel: 3,
				TrustScore:        1.5,
			},
			expectErr: true,
			errMsg:    "trust score must be 0.0-1.0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cs.Validate()
			if tc.expectErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestCredentialTypeFromScopeType(t *testing.T) {
	tests := []struct {
		scopeType      types.ScopeType
		expectedType   string
	}{
		{types.ScopeTypeIDDocument, types.TypeDocumentVerification},
		{types.ScopeTypeSelfie, types.TypeFacialVerification},
		{types.ScopeTypeFaceVideo, types.TypeLivenessVerification},
		{types.ScopeTypeBiometric, types.TypeBiometricVerification},
		{types.ScopeTypeEmailProof, types.TypeEmailVerification},
		{types.ScopeTypeSMSProof, types.TypeSMSVerification},
		{types.ScopeTypeDomainVerify, types.TypeDomainVerification},
		{types.ScopeTypeSSOMetadata, types.TypeSSOVerification},
		{types.ScopeTypeADSSO, types.TypeSSOVerification},
	}

	for _, tc := range tests {
		t.Run(string(tc.scopeType), func(t *testing.T) {
			result := types.CredentialTypeFromScopeType(tc.scopeType)
			assert.Equal(t, tc.expectedType, result)
		})
	}
}

func TestVerificationLevelFromScore(t *testing.T) {
	tests := []struct {
		score         uint32
		expectedLevel int
	}{
		{100, 4},
		{95, 4},
		{90, 4},
		{89, 3},
		{70, 3},
		{69, 2},
		{50, 2},
		{49, 1},
		{30, 1},
		{29, 0},
		{0, 0},
	}

	for _, tc := range tests {
		result := types.VerificationLevelFromScore(tc.score)
		assert.Equal(t, tc.expectedLevel, result, "score %d", tc.score)
	}
}

func TestTrustScoreFromScore(t *testing.T) {
	tests := []struct {
		score         uint32
		expectedScore float64
	}{
		{100, 1.0},
		{50, 0.5},
		{0, 0.0},
		{85, 0.85},
		{150, 1.0}, // capped at 100
	}

	for _, tc := range tests {
		result := types.TrustScoreFromScore(tc.score)
		assert.Equal(t, tc.expectedScore, result, "score %d", tc.score)
	}
}

func TestCredentialRecord_IsActive(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	tests := []struct {
		name     string
		record   types.CredentialRecord
		expected bool
	}{
		{
			name: "active credential",
			record: types.CredentialRecord{
				Status:    types.CredentialStatusActive,
				ExpiresAt: &future,
			},
			expected: true,
		},
		{
			name: "revoked credential",
			record: types.CredentialRecord{
				Status:    types.CredentialStatusRevoked,
				RevokedAt: &now,
			},
			expected: false,
		},
		{
			name: "expired credential",
			record: types.CredentialRecord{
				Status:    types.CredentialStatusActive,
				ExpiresAt: &past,
			},
			expected: false,
		},
		{
			name: "no expiry",
			record: types.CredentialRecord{
				Status: types.CredentialStatusActive,
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.record.IsActive(now)
			assert.Equal(t, tc.expected, result)
		})
	}
}
