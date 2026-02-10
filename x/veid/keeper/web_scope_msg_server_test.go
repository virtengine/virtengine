package keeper_test

import (
	"encoding/base64"
	"encoding/hex"
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
	"github.com/stretchr/testify/require"

	encryptiontypes "github.com/virtengine/virtengine/x/encryption/types"
	"github.com/virtengine/virtengine/x/veid/keeper"
	"github.com/virtengine/virtengine/x/veid/types"
)

func TestSubmitSocialMediaScopeIntegration(t *testing.T) {
	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	ctx, stateStore := createContextWithStore(t, storeKey)
	defer CloseStoreIfNeeded(stateStore)

	k := keeper.NewKeeper(cdc, storeKey, "authority")
	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	accountAddress := sdk.AccAddress([]byte("social_scope_addr")).String()
	wallet := types.NewIdentityWallet(
		"wallet-social-1",
		accountAddress,
		ctx.BlockTime(),
		[]byte("binding-sig"),
		[]byte("binding-pub"),
	)

	nameHash := types.HashSocialMediaField("Jane Doe")
	nameHashBytes, err := hex.DecodeString(nameHash)
	require.NoError(t, err)
	wallet.DerivedFeatures.DocFieldHashes[types.DocFieldNameHash] = nameHashBytes
	wallet.DerivedFeatures.LastComputedAt = ctx.BlockTime()
	require.NoError(t, k.SetWallet(ctx, wallet))

	issuer := types.AttestationIssuer{
		ID:             "did:virtengine:signer:social",
		KeyFingerprint: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	}
	subject := types.AttestationSubject{
		ID:             "did:virtengine:" + accountAddress,
		AccountAddress: accountAddress,
	}
	att := types.NewVerificationAttestation(
		issuer,
		subject,
		types.AttestationTypeSocialMediaVerification,
		[]byte("nonce-social-32------------"),
		ctx.BlockTime(),
		time.Hour,
		95,
		90,
	)
	att.SetProof(types.AttestationProof{
		Type:               types.ProofTypeEd25519,
		Created:            ctx.BlockTime(),
		VerificationMethod: "did:virtengine:signer:social#keys-1",
		ProofPurpose:       "assertionMethod",
		ProofValue:         base64.StdEncoding.EncodeToString([]byte("sig")),
		Nonce:              "nonce",
	})
	attestationBytes, err := att.ToJSON()
	require.NoError(t, err)

	alg := encryptiontypes.DefaultAlgorithm()
	algInfo, err := encryptiontypes.GetAlgorithmInfo(alg)
	require.NoError(t, err)
	recipientPubKey := make([]byte, encryptiontypes.X25519PublicKeySize)
	recipientKeyID := encryptiontypes.ComputeKeyFingerprint(recipientPubKey)
	payload := types.EncryptedPayloadEnvelope{
		Version:             encryptiontypes.EnvelopeVersion,
		AlgorithmId:         alg,
		AlgorithmVersion:    algInfo.Version,
		RecipientKeyIds:     []string{recipientKeyID},
		RecipientPublicKeys: [][]byte{recipientPubKey},
		EncryptedKeys:       [][]byte{[]byte("encrypted-key")},
		Nonce:               make([]byte, algInfo.NonceSize),
		Ciphertext:          []byte("ciphertext"),
		SenderPubKey:        make([]byte, encryptiontypes.X25519PublicKeySize),
		SenderSignature:     []byte("sender-sig"),
	}

	msg := &types.MsgSubmitSocialMediaScope{
		AccountAddress:         accountAddress,
		ScopeId:                "social-1",
		Provider:               types.SocialMediaProviderToProto(types.SocialMediaProviderGoogle),
		ProfileNameHash:        nameHash,
		EmailHash:              types.HashSocialMediaField("jane@example.com"),
		AccountAgeDays:         365,
		IsVerified:             true,
		AttestationData:        attestationBytes,
		AccountSignature:       []byte("account-sig"),
		EncryptedPayload:       payload,
		EvidenceStorageBackend: "ipfs",
		EvidenceStorageRef:     "vault://social/1",
	}

	srv := keeper.NewMsgServerImpl(k)
	resp, err := srv.SubmitSocialMediaScope(ctx, msg)
	require.NoError(t, err)
	require.Equal(t, "social-1", resp.ScopeId)

	record, found := k.GetSocialMediaScope(ctx, "social-1")
	require.True(t, found)
	require.Equal(t, types.SocialMediaProviderGoogle, record.Provider)

	score, _, scoreFound := k.GetScore(ctx, accountAddress)
	require.True(t, scoreFound)
	require.NotZero(t, score)
}

func createContextWithStore(t *testing.T, storeKey *storetypes.KVStoreKey) (sdk.Context, store.CommitMultiStore) {
	t.Helper()
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	if err := stateStore.LoadLatestVersion(); err != nil {
		t.Fatalf("failed to load latest version: %v", err)
	}
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Time:   time.Now().UTC(),
		Height: 100,
	}, false, log.NewNopLogger())
	return ctx, stateStore
}
