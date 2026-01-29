package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
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
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
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

type SignatureCryptoTestSuite struct {
	suite.Suite
	ctx    sdk.Context
	keeper keeper.Keeper
	cdc    codec.Codec
}

func TestSignatureCryptoTestSuite(t *testing.T) {
	suite.Run(t, new(SignatureCryptoTestSuite))
}

func (s *SignatureCryptoTestSuite) SetupTest() {
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

	// Set default params with signatures required
	params := types.DefaultParams()
	params.RequireClientSignature = true
	params.RequireUserSignature = true
	err := s.keeper.SetParams(s.ctx, params)
	s.Require().NoError(err)
}

func (s *SignatureCryptoTestSuite) createContextWithStore(storeKey *storetypes.KVStoreKey) sdk.Context {
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	err := stateStore.LoadLatestVersion()
	if err != nil {
		s.T().Fatalf("failed to load latest version: %v", err)
	}

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx
}

// ============================================================================
// Ed25519 Signature Verification Tests
// ============================================================================

func TestEd25519SignatureVerification(t *testing.T) {
	t.Run("valid signature", func(t *testing.T) {
		// Generate a key pair
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		// Create message and sign it
		message := []byte("test message for ed25519 signing")
		signature := ed25519.Sign(privKey, message)

		// Verify should succeed
		err = keeper.VerifyEd25519Signature(pubKey, message, signature)
		assert.NoError(t, err)
	})

	t.Run("invalid signature - wrong message", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		message := []byte("original message")
		signature := ed25519.Sign(privKey, message)

		// Verify with different message should fail
		wrongMessage := []byte("different message")
		err = keeper.VerifyEd25519Signature(pubKey, wrongMessage, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})

	t.Run("invalid signature - wrong key", func(t *testing.T) {
		_, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		otherPubKey, _, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		message := []byte("test message")
		signature := ed25519.Sign(privKey, message)

		// Verify with different public key should fail
		err = keeper.VerifyEd25519Signature(otherPubKey, message, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})

	t.Run("invalid public key length", func(t *testing.T) {
		shortKey := make([]byte, 16) // Too short
		message := []byte("test message")
		signature := make([]byte, 64)

		err := keeper.VerifyEd25519Signature(shortKey, message, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid public key length")
	})

	t.Run("invalid signature length", func(t *testing.T) {
		pubKey, _, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		message := []byte("test message")
		shortSig := make([]byte, 32) // Too short

		err = keeper.VerifyEd25519Signature(pubKey, message, shortSig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature length")
	})

	t.Run("empty message", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		message := []byte{}
		signature := ed25519.Sign(privKey, message)

		// Empty message should still work
		err = keeper.VerifyEd25519Signature(pubKey, message, signature)
		assert.NoError(t, err)
	})
}

// ============================================================================
// Secp256k1 Signature Verification Tests
// ============================================================================

func TestSecp256k1SignatureVerification(t *testing.T) {
	t.Run("valid signature", func(t *testing.T) {
		// Generate a key pair
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().(*secp256k1.PubKey)

		// Create message and sign it (hash first, then sign)
		message := []byte("test message for secp256k1 signing")
		hash := sha256.Sum256(message)
		signature, err := privKey.Sign(hash[:])
		require.NoError(t, err)

		// Verify should succeed
		err = keeper.VerifySecp256k1Signature(pubKey, message, signature)
		assert.NoError(t, err)
	})

	t.Run("invalid signature - wrong message", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().(*secp256k1.PubKey)

		message := []byte("original message")
		hash := sha256.Sum256(message)
		signature, err := privKey.Sign(hash[:])
		require.NoError(t, err)

		// Verify with different message should fail
		wrongMessage := []byte("different message")
		err = keeper.VerifySecp256k1Signature(pubKey, wrongMessage, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})

	t.Run("invalid signature - wrong key", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		otherPubKey := secp256k1.GenPrivKey().PubKey().(*secp256k1.PubKey)

		message := []byte("test message")
		hash := sha256.Sum256(message)
		signature, err := privKey.Sign(hash[:])
		require.NoError(t, err)

		// Verify with different public key should fail
		err = keeper.VerifySecp256k1Signature(otherPubKey, message, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "signature verification failed")
	})

	t.Run("nil public key", func(t *testing.T) {
		message := []byte("test message")
		signature := make([]byte, 64)

		err := keeper.VerifySecp256k1Signature(nil, message, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "public key is nil")
	})

	t.Run("invalid signature length", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().(*secp256k1.PubKey)

		message := []byte("test message")
		shortSig := make([]byte, 32) // Too short

		err := keeper.VerifySecp256k1Signature(pubKey, message, shortSig)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid signature length")
	})
}

func TestSecp256k1SignatureVerificationRaw(t *testing.T) {
	t.Run("valid signature with raw bytes", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().(*secp256k1.PubKey)
		pubKeyBytes := pubKey.Bytes()

		message := []byte("test message for raw verification")
		hash := sha256.Sum256(message)
		signature, err := privKey.Sign(hash[:])
		require.NoError(t, err)

		err = keeper.VerifySecp256k1SignatureRaw(pubKeyBytes, message, signature)
		assert.NoError(t, err)
	})

	t.Run("invalid public key length", func(t *testing.T) {
		shortKey := make([]byte, 16)
		message := []byte("test message")
		signature := make([]byte, 64)

		err := keeper.VerifySecp256k1SignatureRaw(shortKey, message, signature)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid public key length")
	})
}

// ============================================================================
// Salt Binding Verification Tests
// ============================================================================

func TestSaltBindingVerification(t *testing.T) {
	t.Run("valid salt binding with ed25519", func(t *testing.T) {
		pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		salt := make([]byte, 32)
		_, err = rand.Read(salt)
		require.NoError(t, err)

		address := sdk.AccAddress([]byte("test_address_12345"))
		scopeID := "scope-123"
		timestamp := time.Now()
		currentTime := timestamp.Add(1 * time.Minute) // Within valid range

		// Create binding payload and sign it
		bindingData := &keeper.SaltBindingData{
			Salt:      salt,
			Address:   address,
			ScopeID:   scopeID,
			Timestamp: timestamp,
		}
		payload := bindingData.Payload()
		signature := ed25519.Sign(privKey, payload)

		err = keeper.VerifySaltBinding(
			salt, address, scopeID, timestamp,
			signature, pubKey, keeper.AlgorithmEd25519, currentTime,
		)
		assert.NoError(t, err)
	})

	t.Run("valid salt binding with secp256k1", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().(*secp256k1.PubKey)

		salt := make([]byte, 32)
		_, err := rand.Read(salt)
		require.NoError(t, err)

		address := sdk.AccAddress([]byte("test_address_12345"))
		scopeID := "scope-456"
		timestamp := time.Now()
		currentTime := timestamp.Add(1 * time.Minute)

		// Create binding payload, hash it, and sign
		bindingData := &keeper.SaltBindingData{
			Salt:      salt,
			Address:   address,
			ScopeID:   scopeID,
			Timestamp: timestamp,
		}
		payload := bindingData.Payload()
		hash := sha256.Sum256(payload)
		signature, err := privKey.Sign(hash[:])
		require.NoError(t, err)

		err = keeper.VerifySaltBinding(
			salt, address, scopeID, timestamp,
			signature, pubKey.Bytes(), keeper.AlgorithmSecp256k1, currentTime,
		)
		assert.NoError(t, err)
	})

	t.Run("invalid - empty salt", func(t *testing.T) {
		pubKey, _, _ := ed25519.GenerateKey(rand.Reader)
		address := sdk.AccAddress([]byte("test"))
		signature := make([]byte, 64)

		err := keeper.VerifySaltBinding(
			[]byte{}, address, "scope", time.Now(),
			signature, pubKey, keeper.AlgorithmEd25519, time.Now(),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "salt cannot be empty")
	})

	t.Run("invalid - empty address", func(t *testing.T) {
		pubKey, _, _ := ed25519.GenerateKey(rand.Reader)
		salt := []byte("test-salt")
		signature := make([]byte, 64)

		err := keeper.VerifySaltBinding(
			salt, sdk.AccAddress{}, "scope", time.Now(),
			signature, pubKey, keeper.AlgorithmEd25519, time.Now(),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("invalid - empty scope ID", func(t *testing.T) {
		pubKey, _, _ := ed25519.GenerateKey(rand.Reader)
		salt := []byte("test-salt")
		address := sdk.AccAddress([]byte("test"))
		signature := make([]byte, 64)

		err := keeper.VerifySaltBinding(
			salt, address, "", time.Now(),
			signature, pubKey, keeper.AlgorithmEd25519, time.Now(),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "scope ID cannot be empty")
	})

	t.Run("invalid - timestamp too old", func(t *testing.T) {
		pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
		salt := []byte("test-salt")
		address := sdk.AccAddress([]byte("test_address"))
		scopeID := "scope"
		timestamp := time.Now().Add(-10 * time.Minute) // Too old
		currentTime := time.Now()

		bindingData := &keeper.SaltBindingData{
			Salt:      salt,
			Address:   address,
			ScopeID:   scopeID,
			Timestamp: timestamp,
		}
		payload := bindingData.Payload()
		signature := ed25519.Sign(privKey, payload)

		err := keeper.VerifySaltBinding(
			salt, address, scopeID, timestamp,
			signature, pubKey, keeper.AlgorithmEd25519, currentTime,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp is too old")
	})

	t.Run("invalid - timestamp in future", func(t *testing.T) {
		pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
		salt := []byte("test-salt")
		address := sdk.AccAddress([]byte("test_address"))
		scopeID := "scope"
		timestamp := time.Now().Add(5 * time.Minute) // Too far in future
		currentTime := time.Now()

		bindingData := &keeper.SaltBindingData{
			Salt:      salt,
			Address:   address,
			ScopeID:   scopeID,
			Timestamp: timestamp,
		}
		payload := bindingData.Payload()
		signature := ed25519.Sign(privKey, payload)

		err := keeper.VerifySaltBinding(
			salt, address, scopeID, timestamp,
			signature, pubKey, keeper.AlgorithmEd25519, currentTime,
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timestamp is in the future")
	})

	t.Run("invalid - unsupported algorithm", func(t *testing.T) {
		salt := []byte("test-salt")
		address := sdk.AccAddress([]byte("test"))
		signature := make([]byte, 64)
		pubKey := make([]byte, 32)

		err := keeper.VerifySaltBinding(
			salt, address, "scope", time.Now(),
			signature, pubKey, "unknown-algorithm", time.Now(),
		)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported signature algorithm")
	})
}

// ============================================================================
// Address Matching Tests
// ============================================================================

func TestAddressMatchesPubKey(t *testing.T) {
	t.Run("matching address and pubkey", func(t *testing.T) {
		privKey := secp256k1.GenPrivKey()
		pubKey := privKey.PubKey().(*secp256k1.PubKey)
		expectedAddr := sdk.AccAddress(pubKey.Address())

		err := keeper.VerifyAddressMatchesPubKey(pubKey, expectedAddr)
		assert.NoError(t, err)
	})

	t.Run("mismatched address and pubkey", func(t *testing.T) {
		pubKey := secp256k1.GenPrivKey().PubKey().(*secp256k1.PubKey)
		wrongAddr := sdk.AccAddress([]byte("wrong_address_here"))

		err := keeper.VerifyAddressMatchesPubKey(pubKey, wrongAddr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not match expected address")
	})

	t.Run("nil public key", func(t *testing.T) {
		addr := sdk.AccAddress([]byte("test"))

		err := keeper.VerifyAddressMatchesPubKey(nil, addr)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "public key is nil")
	})
}

// ============================================================================
// ValidateClientSignature Integration Tests
// ============================================================================

func (s *SignatureCryptoTestSuite) TestValidateClientSignature() {
	// Generate Ed25519 key pair for the client
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	clientID := "test-client-ed25519"

	// Register the client
	client := types.ApprovedClient{
		ClientID:     clientID,
		Name:         "Test Ed25519 Client",
		PublicKey:    pubKey,
		Algorithm:    "ed25519",
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}
	err = s.keeper.SetApprovedClient(s.ctx, client)
	s.Require().NoError(err)

	s.Run("valid ed25519 signature", func() {
		payload := []byte("test payload for client signature")
		signature := ed25519.Sign(privKey, payload)

		err := s.keeper.ValidateClientSignature(s.ctx, clientID, signature, payload)
		s.NoError(err)
	})

	s.Run("invalid signature - wrong payload", func() {
		payload := []byte("original payload")
		signature := ed25519.Sign(privKey, payload)

		wrongPayload := []byte("different payload")
		err := s.keeper.ValidateClientSignature(s.ctx, clientID, signature, wrongPayload)
		s.Error(err)
		s.Contains(err.Error(), "invalid client signature")
	})

	s.Run("client not found", func() {
		err := s.keeper.ValidateClientSignature(s.ctx, "unknown-client", []byte{}, []byte{})
		s.Error(err)
		s.Contains(err.Error(), "client not approved")
	})

	s.Run("inactive client", func() {
		inactiveClientID := "inactive-client"
		inactiveClient := types.ApprovedClient{
			ClientID:     inactiveClientID,
			Name:         "Inactive Client",
			PublicKey:    pubKey,
			Algorithm:    "ed25519",
			Active:       false,
			RegisteredAt: time.Now().Unix(),
		}
		err := s.keeper.SetApprovedClient(s.ctx, inactiveClient)
		s.Require().NoError(err)

		err = s.keeper.ValidateClientSignature(s.ctx, inactiveClientID, []byte{}, []byte{})
		s.Error(err)
		s.Contains(err.Error(), "not active")
	})
}

func (s *SignatureCryptoTestSuite) TestValidateClientSignatureSecp256k1() {
	// Generate secp256k1 key pair for the client
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(*secp256k1.PubKey)

	clientID := "test-client-secp256k1"

	// Register the client
	client := types.ApprovedClient{
		ClientID:     clientID,
		Name:         "Test Secp256k1 Client",
		PublicKey:    pubKey.Bytes(),
		Algorithm:    "secp256k1",
		Active:       true,
		RegisteredAt: time.Now().Unix(),
	}
	err := s.keeper.SetApprovedClient(s.ctx, client)
	s.Require().NoError(err)

	s.Run("valid secp256k1 signature", func() {
		payload := []byte("test payload for secp256k1 client signature")
		hash := sha256.Sum256(payload)
		signature, err := privKey.Sign(hash[:])
		s.Require().NoError(err)

		err = s.keeper.ValidateClientSignature(s.ctx, clientID, signature, payload)
		s.NoError(err)
	})

	s.Run("invalid secp256k1 signature", func() {
		payload := []byte("original payload")
		hash := sha256.Sum256(payload)
		signature, err := privKey.Sign(hash[:])
		s.Require().NoError(err)

		wrongPayload := []byte("different payload")
		err = s.keeper.ValidateClientSignature(s.ctx, clientID, signature, wrongPayload)
		s.Error(err)
	})
}

// ============================================================================
// ValidateUserSignature Integration Tests
// ============================================================================

func (s *SignatureCryptoTestSuite) TestValidateUserSignature() {
	privKey := secp256k1.GenPrivKey()
	address := sdk.AccAddress(privKey.PubKey().Address())

	s.Run("empty signature rejected", func() {
		err := s.keeper.ValidateUserSignature(s.ctx, address, []byte{}, []byte("payload"))
		s.Error(err)
		s.Contains(err.Error(), "cannot be empty")
	})

	s.Run("invalid signature length rejected", func() {
		shortSig := make([]byte, 32) // Wrong length
		err := s.keeper.ValidateUserSignature(s.ctx, address, shortSig, []byte("payload"))
		s.Error(err)
		s.Contains(err.Error(), "invalid signature length")
	})

	s.Run("valid signature length accepted", func() {
		validLengthSig := make([]byte, 64)
		err := s.keeper.ValidateUserSignature(s.ctx, address, validLengthSig, []byte("payload"))
		// Should pass basic validation (full verification needs pubkey)
		s.NoError(err)
	})
}

func (s *SignatureCryptoTestSuite) TestValidateUserSignatureWithPubKey() {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(*secp256k1.PubKey)
	address := sdk.AccAddress(pubKey.Address())

	s.Run("valid signature with pubkey", func() {
		payload := []byte("test payload for user signature")
		hash := sha256.Sum256(payload)
		signature, err := privKey.Sign(hash[:])
		s.Require().NoError(err)

		err = s.keeper.ValidateUserSignatureWithPubKey(s.ctx, address, pubKey.Bytes(), signature, payload)
		s.NoError(err)
	})

	s.Run("invalid signature - wrong payload", func() {
		payload := []byte("original payload")
		hash := sha256.Sum256(payload)
		signature, err := privKey.Sign(hash[:])
		s.Require().NoError(err)

		wrongPayload := []byte("different payload")
		err = s.keeper.ValidateUserSignatureWithPubKey(s.ctx, address, pubKey.Bytes(), signature, wrongPayload)
		s.Error(err)
		s.Contains(err.Error(), "signature verification failed")
	})

	s.Run("invalid - address mismatch", func() {
		payload := []byte("test payload")
		hash := sha256.Sum256(payload)
		signature, err := privKey.Sign(hash[:])
		s.Require().NoError(err)

		wrongAddress := sdk.AccAddress([]byte("wrong_address_value"))
		err = s.keeper.ValidateUserSignatureWithPubKey(s.ctx, wrongAddress, pubKey.Bytes(), signature, payload)
		s.Error(err)
		s.Contains(err.Error(), "address mismatch")
	})

	s.Run("invalid pubkey length", func() {
		shortKey := make([]byte, 16)
		err := s.keeper.ValidateUserSignatureWithPubKey(s.ctx, address, shortKey, []byte{}, []byte{})
		s.Error(err)
		s.Contains(err.Error(), "invalid public key length")
	})
}

// ============================================================================
// ValidateSaltBinding Integration Tests
// ============================================================================

func (s *SignatureCryptoTestSuite) TestValidateSaltBinding() {
	s.Run("valid salt passes", func() {
		salt := make([]byte, 32)
		_, err := rand.Read(salt)
		s.Require().NoError(err)

		err = s.keeper.ValidateSaltBinding(s.ctx, salt)
		s.NoError(err)
	})

	s.Run("empty salt rejected", func() {
		err := s.keeper.ValidateSaltBinding(s.ctx, []byte{})
		s.Error(err)
		s.Contains(err.Error(), "cannot be empty")
	})

	s.Run("salt too short rejected", func() {
		shortSalt := make([]byte, 8)
		err := s.keeper.ValidateSaltBinding(s.ctx, shortSalt)
		s.Error(err)
		s.Contains(err.Error(), "at least")
	})
}

func (s *SignatureCryptoTestSuite) TestValidateSaltBindingWithSignature() {
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	s.Require().NoError(err)

	s.Run("valid salt binding with signature", func() {
		salt := make([]byte, 32)
		_, err := rand.Read(salt)
		s.Require().NoError(err)

		address := sdk.AccAddress([]byte("test_address_12345678"))
		scopeID := "scope-test-123"
		timestamp := s.ctx.BlockTime().Unix()

		// Create and sign binding
		bindingData := &keeper.SaltBindingData{
			Salt:      salt,
			Address:   address,
			ScopeID:   scopeID,
			Timestamp: time.Unix(timestamp, 0),
		}
		payload := bindingData.Payload()
		signature := ed25519.Sign(privKey, payload)

		err = s.keeper.ValidateSaltBindingWithSignature(
			s.ctx, salt, address, scopeID, timestamp, signature, pubKey, "ed25519",
		)
		s.NoError(err)
	})

	s.Run("invalid signature rejected", func() {
		salt := make([]byte, 32)
		_, err := rand.Read(salt)
		s.Require().NoError(err)

		address := sdk.AccAddress([]byte("test_address_12345678"))
		scopeID := "scope-test-456"
		timestamp := s.ctx.BlockTime().Unix()

		// Create wrong signature
		wrongSignature := make([]byte, 64)
		_, _ = rand.Read(wrongSignature)

		err = s.keeper.ValidateSaltBindingWithSignature(
			s.ctx, salt, address, scopeID, timestamp, wrongSignature, pubKey, "ed25519",
		)
		s.Error(err)
		s.Contains(err.Error(), "salt binding verification failed")
	})
}

// ============================================================================
// Composite Signature Verification Tests
// ============================================================================

func TestVerifyClientAndUserSignatures(t *testing.T) {
	// Setup client keys (Ed25519)
	clientPubKey, clientPrivKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	// Setup user keys (secp256k1)
	userPrivKey := secp256k1.GenPrivKey()
	userPubKey := userPrivKey.PubKey().(*secp256k1.PubKey)
	userAddress := sdk.AccAddress(userPubKey.Address())

	t.Run("both signatures valid", func(t *testing.T) {
		clientPayload := []byte("client payload")
		clientSignature := ed25519.Sign(clientPrivKey, clientPayload)

		userPayload := []byte("user payload")
		userHash := sha256.Sum256(userPayload)
		userSignature, err := userPrivKey.Sign(userHash[:])
		require.NoError(t, err)

		clientResult, userResult := keeper.VerifyClientAndUserSignatures(
			clientPubKey, keeper.AlgorithmEd25519, clientSignature, clientPayload,
			userPubKey, userSignature, userPayload, userAddress,
		)

		assert.True(t, clientResult.Verified)
		assert.Nil(t, clientResult.Error)
		assert.True(t, userResult.Verified)
		assert.Nil(t, userResult.Error)
	})

	t.Run("client signature invalid", func(t *testing.T) {
		wrongSignature := make([]byte, 64)
		clientPayload := []byte("client payload")

		userPayload := []byte("user payload")
		userHash := sha256.Sum256(userPayload)
		userSignature, err := userPrivKey.Sign(userHash[:])
		require.NoError(t, err)

		clientResult, userResult := keeper.VerifyClientAndUserSignatures(
			clientPubKey, keeper.AlgorithmEd25519, wrongSignature, clientPayload,
			userPubKey, userSignature, userPayload, userAddress,
		)

		assert.False(t, clientResult.Verified)
		assert.NotNil(t, clientResult.Error)
		assert.True(t, userResult.Verified)
		assert.Nil(t, userResult.Error)
	})

	t.Run("user signature invalid", func(t *testing.T) {
		clientPayload := []byte("client payload")
		clientSignature := ed25519.Sign(clientPrivKey, clientPayload)

		userPayload := []byte("user payload")
		wrongSignature := make([]byte, 64)

		clientResult, userResult := keeper.VerifyClientAndUserSignatures(
			clientPubKey, keeper.AlgorithmEd25519, clientSignature, clientPayload,
			userPubKey, wrongSignature, userPayload, userAddress,
		)

		assert.True(t, clientResult.Verified)
		assert.Nil(t, clientResult.Error)
		assert.False(t, userResult.Verified)
		assert.NotNil(t, userResult.Error)
	})

	t.Run("user address mismatch", func(t *testing.T) {
		clientPayload := []byte("client payload")
		clientSignature := ed25519.Sign(clientPrivKey, clientPayload)

		userPayload := []byte("user payload")
		userHash := sha256.Sum256(userPayload)
		userSignature, err := userPrivKey.Sign(userHash[:])
		require.NoError(t, err)

		wrongAddress := sdk.AccAddress([]byte("wrong_address_here"))

		clientResult, userResult := keeper.VerifyClientAndUserSignatures(
			clientPubKey, keeper.AlgorithmEd25519, clientSignature, clientPayload,
			userPubKey, userSignature, userPayload, wrongAddress,
		)

		assert.True(t, clientResult.Verified)
		assert.False(t, userResult.Verified)
		assert.Contains(t, userResult.Error.Error(), "does not match")
	})
}

// ============================================================================
// Wrong Key Type Tests
// ============================================================================

func TestWrongKeyType(t *testing.T) {
	t.Run("ed25519 signature with secp256k1 key fails", func(t *testing.T) {
		// Create Ed25519 signature
		_, privKey, _ := ed25519.GenerateKey(rand.Reader)
		message := []byte("test message")
		signature := ed25519.Sign(privKey, message)

		// Try to verify with secp256k1 key (wrong type)
		secp256k1Key := secp256k1.GenPrivKey().PubKey().(*secp256k1.PubKey)

		err := keeper.VerifySecp256k1Signature(secp256k1Key, message, signature)
		assert.Error(t, err)
	})

	t.Run("secp256k1 signature with ed25519 key fails", func(t *testing.T) {
		// Create secp256k1 signature
		privKey := secp256k1.GenPrivKey()
		message := []byte("test message")
		hash := sha256.Sum256(message)
		signature, _ := privKey.Sign(hash[:])

		// Try to verify with Ed25519 key (wrong type)
		ed25519PubKey, _, _ := ed25519.GenerateKey(rand.Reader)

		err := keeper.VerifyEd25519Signature(ed25519PubKey, message, signature)
		assert.Error(t, err)
	})
}

// ============================================================================
// Benchmark Tests
// ============================================================================

func BenchmarkEd25519Verification(b *testing.B) {
	pubKey, privKey, _ := ed25519.GenerateKey(rand.Reader)
	message := []byte("benchmark message for ed25519")
	signature := ed25519.Sign(privKey, message)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = keeper.VerifyEd25519Signature(pubKey, message, signature)
	}
}

func BenchmarkSecp256k1Verification(b *testing.B) {
	privKey := secp256k1.GenPrivKey()
	pubKey := privKey.PubKey().(*secp256k1.PubKey)
	message := []byte("benchmark message for secp256k1")
	hash := sha256.Sum256(message)
	signature, _ := privKey.Sign(hash[:])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = keeper.VerifySecp256k1Signature(pubKey, message, signature)
	}
}
