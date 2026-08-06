//go:build security && integration

package security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"io"
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
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	mfapb "github.com/virtengine/virtengine/sdk/go/node/mfa/v1"
	veidv1 "github.com/virtengine/virtengine/sdk/go/node/veid/v1"
	veidkeeper "github.com/virtengine/virtengine/x/veid/keeper"
	veidtypes "github.com/virtengine/virtengine/x/veid/types"
	mfatypes "github.com/virtengine/virtengine/x/mfa/types"
)

type auditWalletEnv struct {
	ctx        sdk.Context
	keeper     veidkeeper.Keeper
	stateStore store.CommitMultiStore
	pubKey     ed25519.PublicKey
	privKey    ed25519.PrivateKey
	address    sdk.AccAddress
}

func setupAuditWalletEnv(t *testing.T) *auditWalletEnv {
	t.Helper()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	veidtypes.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(veidtypes.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())
	t.Cleanup(func() {
		if closer, ok := stateStore.(io.Closer); ok {
			_ = closer.Close()
		}
	})

	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Unix(1_700_000_000, 0).UTC(),
		Height: 200,
	}, false, log.NewNopLogger())

	keeper := veidkeeper.NewKeeper(cdc, storeKey, "authority")
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	return &auditWalletEnv{
		ctx:        ctx,
		keeper:     keeper,
		stateStore: stateStore,
		pubKey:     pubKey,
		privKey:    privKey,
		address:    sdk.AccAddress(pubKey[:20]),
	}
}

func (e *auditWalletEnv) signWalletBinding(walletID string, privKey ed25519.PrivateKey) []byte {
	msg := veidtypes.GetWalletBindingMessage(walletID, e.address.String())
	return ed25519.Sign(privKey, msg)
}

func (e *auditWalletEnv) signRebindAuthorization(newPubKey []byte, privKey ed25519.PrivateKey) []byte {
	return ed25519.Sign(privKey, newPubKey)
}

func TestIdentitySecurity_RebindWalletFlowRequiresLinkedSignatures(t *testing.T) {
	t.Run("valid_old_and_new_signatures_rebind_wallet", func(t *testing.T) {
		env := setupAuditWalletEnv(t)

		walletID := veidkeeper.GenerateWalletID(env.address.String())
		bindingSignature := env.signWalletBinding(walletID, env.privKey)
		_, err := env.keeper.CreateWallet(env.ctx, env.address, bindingSignature, env.pubKey)
		require.NoError(t, err)

		newPubKey, newPrivKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		oldSignature := env.signRebindAuthorization(newPubKey, env.privKey)
		newBindingSignature := env.signWalletBinding(walletID, newPrivKey)

		err = env.keeper.RebindWallet(env.ctx, env.address, newBindingSignature, newPubKey, oldSignature)
		require.NoError(t, err)

		wallet, found := env.keeper.GetWallet(env.ctx, env.address)
		require.True(t, found)
		require.Equal(t, []byte(newPubKey), wallet.BindingPubKey)
		require.Equal(t, newBindingSignature, wallet.BindingSignature)
	})

	t.Run("invalid_old_signature_is_rejected", func(t *testing.T) {
		env := setupAuditWalletEnv(t)

		walletID := veidkeeper.GenerateWalletID(env.address.String())
		bindingSignature := env.signWalletBinding(walletID, env.privKey)
		_, err := env.keeper.CreateWallet(env.ctx, env.address, bindingSignature, env.pubKey)
		require.NoError(t, err)

		newPubKey, newPrivKey, err := ed25519.GenerateKey(rand.Reader)
		require.NoError(t, err)

		invalidOldSignature := make([]byte, ed25519.SignatureSize)
		_, err = rand.Read(invalidOldSignature)
		require.NoError(t, err)

		newBindingSignature := env.signWalletBinding(walletID, newPrivKey)
		err = env.keeper.RebindWallet(env.ctx, env.address, newBindingSignature, newPubKey, invalidOldSignature)
		require.Error(t, err)
		require.Contains(t, err.Error(), "old signature verification failed")
	})
}

func TestIdentitySecurity_MFAGatingEnforcesSerializedRebindProof(t *testing.T) {
	env := newMFATestEnv(t)
	sender := sdk.AccAddress([]byte("security-rebind-sender"))

	enrollActiveFactor(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP, "totp-security-rebind")
	setMFAPolicy(t, env.keeper, env.ctx, sender, mfatypes.FactorTypeTOTP)
	setSensitiveTxConfig(t, env.keeper, env.ctx, mfatypes.SensitiveTxKeyRotation, "", mfatypes.FactorTypeTOTP)

	newBindingSigHash := sha256.Sum256([]byte("new-binding-signature"))
	newBindingPubKeyHash := sha256.Sum256([]byte("new-binding-pubkey"))
	oldSignatureHash := sha256.Sum256([]byte("old-binding-signature"))

	msg := &veidv1.MsgRebindWallet{
		Sender:              sender.String(),
		NewBindingSignature: newBindingSigHash[:],
		NewBindingPubKey:    newBindingPubKeyHash[:],
		OldSignature:        oldSignatureHash[:],
		DeviceFingerprint:   "audit-wallet-device",
	}

	err := runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, msg)
	require.Error(t, err)
	require.ErrorContains(t, err, mfatypes.SensitiveTxKeyRotation.String())

	createSession(t, env.keeper, env.ctx, sender, mfatypes.SensitiveTxKeyRotation, "audit-rebind-session", mfatypes.FactorTypeTOTP)

	rawProof, marshalErr := proto.Marshal(&mfapb.MFAProof{
		SessionId:       "audit-rebind-session",
		VerifiedFactors: []mfapb.FactorType{mfapb.FactorTypeTOTP},
		Timestamp:       env.ctx.BlockTime().Unix(),
	})
	require.NoError(t, marshalErr)

	msg.MfaProof = rawProof

	require.NoError(t, runMFADecorator(t, env.decorator, env.ctx, []sdk.AccAddress{sender}, msg))
}
