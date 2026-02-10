package keeper

import (
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	storemetrics "cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/encryption/crypto"
	encryptionkeeper "github.com/virtengine/virtengine/x/encryption/keeper"
	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/market/types/marketplace"
)

func setupEncryptionTest(t *testing.T) (sdk.Context, *encryptionkeeper.Keeper, *EncryptionEnforcer) {
	t.Helper()

	ctx, ek := setupKeeperForEncryption(t)
	enforcer := NewEncryptionEnforcer(ek)

	return ctx, ek, enforcer
}

// setupKeeperForEncryption creates a minimal encryption keeper for testing
func setupKeeperForEncryption(t testing.TB) (sdk.Context, *encryptionkeeper.Keeper) {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(interfaceRegistry)

	key := storetypes.NewKVStoreKey(encryptiontypes.StoreKey)
	db := dbm.NewMemDB()

	ms := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	ms.MountStoreWithDB(key, storetypes.StoreTypeIAVL, db)
	err := ms.LoadLatestVersion()
	require.NoError(t, err)

	ctx := sdk.NewContext(ms, tmproto.Header{Time: time.Unix(1000, 0)}, false, log.NewNopLogger())
	k := encryptionkeeper.NewKeeper(cdc, key, "authority")
	require.NoError(t, k.SetParams(ctx, encryptiontypes.DefaultParams()))

	return ctx, &k
}

// Helper to create a valid encrypted bid payload for testing
func createTestEncryptedBidPayload(t *testing.T, bidderPubKey, customerPubKey []byte) (*marketplace.EncryptedBidPayload, error) {
	t.Helper()

	plaintext := []byte(`{"price":"100uakt","resources":{}}`)

	senderKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	envelope, err := crypto.CreateMultiRecipientEnvelope(
		plaintext,
		[][]byte{bidderPubKey, customerPubKey},
		senderKeyPair,
	)
	if err != nil {
		return nil, err
	}

	bidderFingerprint := encryptiontypes.ComputeKeyFingerprint(bidderPubKey)
	customerFingerprint := encryptiontypes.ComputeKeyFingerprint(customerPubKey)

	// Update envelope with correct fingerprints
	envelope.RecipientKeyIDs = []string{bidderFingerprint, customerFingerprint}

	payload := &marketplace.EncryptedBidPayload{
		Envelope:      envelope,
		BidderKeyID:   bidderFingerprint,
		CustomerKeyID: customerFingerprint,
	}
	payload.EnsureEnvelopeHash()

	return payload, nil
}

func TestNewEncryptionEnforcer(t *testing.T) {
	ctx, ek, _ := setupEncryptionTest(t)
	enforcer := NewEncryptionEnforcer(ek)
	require.NotNil(t, enforcer)
	require.NotNil(t, ctx)
}

func TestValidateEncryptedBidPayload(t *testing.T) {
	ctx, ek, enforcer := setupEncryptionTest(t)

	bidderKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)
	customerKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	bidderAddr := sdk.AccAddress([]byte("bidder_address_____"))
	customerAddr := sdk.AccAddress([]byte("customer_address___"))

	// Register keys
	bidderFingerprint, err := ek.RegisterRecipientKey(
		ctx, bidderAddr, bidderKeyPair.PublicKey[:],
		encryptiontypes.DefaultAlgorithm(), "bidder",
	)
	require.NoError(t, err)

	customerFingerprint, err := ek.RegisterRecipientKey(
		ctx, customerAddr, customerKeyPair.PublicKey[:],
		encryptiontypes.DefaultAlgorithm(), "customer",
	)
	require.NoError(t, err)

	t.Run("valid encrypted bid payload", func(t *testing.T) {
		payload, err := createTestEncryptedBidPayload(t, bidderKeyPair.PublicKey[:], customerKeyPair.PublicKey[:])
		require.NoError(t, err)

		// Update with registered fingerprints
		payload.Envelope.RecipientKeyIDs = []string{bidderFingerprint, customerFingerprint}
		payload.BidderKeyID = bidderFingerprint
		payload.CustomerKeyID = customerFingerprint

		err = enforcer.ValidateEncryptedBidPayload(ctx, payload, bidderAddr, customerAddr)
		require.NoError(t, err)
	})

	t.Run("nil payload returns error", func(t *testing.T) {
		err := enforcer.ValidateEncryptedBidPayload(ctx, nil, bidderAddr, customerAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "encrypted bid payload is required")
	})

	t.Run("missing envelope returns error", func(t *testing.T) {
		payload := &marketplace.EncryptedBidPayload{}
		err := enforcer.ValidateEncryptedBidPayload(ctx, payload, bidderAddr, customerAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "encrypted bid payload is required")
	})
}

func TestCheckBidAccess(t *testing.T) {
	ctx, ek, enforcer := setupEncryptionTest(t)

	bidderKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)
	customerKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)
	nonPartyKeyPair, err := crypto.GenerateKeyPair()
	require.NoError(t, err)

	bidderAddr := sdk.AccAddress([]byte("bidder_address_____"))
	customerAddr := sdk.AccAddress([]byte("customer_address___"))
	nonPartyAddr := sdk.AccAddress([]byte("non_party_address__"))

	bidderFingerprint, err := ek.RegisterRecipientKey(
		ctx, bidderAddr, bidderKeyPair.PublicKey[:],
		encryptiontypes.DefaultAlgorithm(), "bidder",
	)
	require.NoError(t, err)

	customerFingerprint, err := ek.RegisterRecipientKey(
		ctx, customerAddr, customerKeyPair.PublicKey[:],
		encryptiontypes.DefaultAlgorithm(), "customer",
	)
	require.NoError(t, err)

	_, err = ek.RegisterRecipientKey(
		ctx, nonPartyAddr, nonPartyKeyPair.PublicKey[:],
		encryptiontypes.DefaultAlgorithm(), "non_party",
	)
	require.NoError(t, err)

	payload, err := createTestEncryptedBidPayload(t, bidderKeyPair.PublicKey[:], customerKeyPair.PublicKey[:])
	require.NoError(t, err)
	payload.Envelope.RecipientKeyIDs = []string{bidderFingerprint, customerFingerprint}
	payload.BidderKeyID = bidderFingerprint
	payload.CustomerKeyID = customerFingerprint

	t.Run("bidder has access", func(t *testing.T) {
		err := enforcer.CheckBidAccess(ctx, payload, bidderAddr)
		require.NoError(t, err)
	})

	t.Run("customer has access", func(t *testing.T) {
		err := enforcer.CheckBidAccess(ctx, payload, customerAddr)
		require.NoError(t, err)
	})

	t.Run("non-party denied access", func(t *testing.T) {
		err := enforcer.CheckBidAccess(ctx, payload, nonPartyAddr)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is not a recipient")
	})
}

func TestEnforceEncryptionForSensitiveFields(t *testing.T) {
	_, _, enforcer := setupEncryptionTest(t)

	t.Run("required field missing", func(t *testing.T) {
		err := enforcer.EnforceEncryptionForSensitiveFields("bid_details", nil, true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "required but missing")
	})

	t.Run("optional field missing", func(t *testing.T) {
		err := enforcer.EnforceEncryptionForSensitiveFields("optional_metadata", nil, false)
		require.NoError(t, err)
	})

	t.Run("valid envelope", func(t *testing.T) {
		envelope := encryptiontypes.NewEncryptedPayloadEnvelope()
		envelope.Ciphertext = []byte("encrypted")
		envelope.RecipientKeyIDs = []string{"fingerprint"}

		err := enforcer.EnforceEncryptionForSensitiveFields("bid_details", envelope, true)
		require.NoError(t, err)
	})
}
