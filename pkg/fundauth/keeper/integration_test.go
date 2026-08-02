package keeper_test

import (
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/metrics"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/pkg/fundauth"
	fundauthkeeper "github.com/virtengine/virtengine/pkg/fundauth/keeper"
)

type integrationKeyResolver struct {
	key fundauth.ResolvedPossessionKey
}

func (resolver integrationKeyResolver) ResolveEd25519(_ context.Context, accountID, keyID string, epoch uint64) (fundauth.ResolvedPossessionKey, error) {
	if accountID != resolver.key.AccountID || keyID != resolver.key.KeyID || epoch != resolver.key.Epoch {
		return fundauth.ResolvedPossessionKey{}, errors.New("key coordinates do not match")
	}
	return resolver.key, nil
}

func TestVerifyPolicyAndConsumeIntegration(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey("fundauth-integration")
	database := dbm.NewMemDB()
	stores := rootmulti.NewStore(database, log.NewNopLogger(), metrics.NewNoOpMetrics())
	stores.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, database)
	require.NoError(t, stores.LoadLatestVersion())
	sdkCtx := sdk.NewContext(stores, tmproto.Header{Height: 150, Time: time.Unix(1_800_000_000, 0)}, false, log.NewNopLogger())
	ctx := sdk.WrapSDKContext(sdkCtx)
	consumer, err := fundauthkeeper.NewKeeper(storeKey)
	require.NoError(t, err)

	digestHex := func(label string) string {
		digest := sha256.Sum256([]byte(label))
		return hex.EncodeToString(digest[:])
	}
	auth := fundauth.FundAuthorization{
		Domain:               fundauth.AuthorizationDomain,
		Version:              fundauth.AuthorizationVersion,
		ChainID:              "virtengine-integration-1",
		AccountID:            "account:alice",
		SignerKeyID:          "device-key:primary",
		SignerKeyEpoch:       7,
		SourceID:             "/cosmos.bank.v1beta1.MsgSend",
		TypeURL:              "/cosmos.bank.v1beta1.MsgSend",
		Phase:                fundauth.PhaseImmediate,
		Effect:               fundauth.EffectTransfer,
		MessageDigestHex:     digestHex("message"),
		Amounts:              []fundauth.Amount{{Denom: "uve", MinorUnits: "42"}},
		Parties:              []fundauth.PartyBinding{{Role: fundauth.PartyRoleSender, AccountID: "account:alice"}, {Role: fundauth.PartyRoleRecipient, AccountID: "account:bob"}},
		CaseDigestHex:        digestHex("case"),
		OrderDigestHex:       digestHex("order"),
		ReferenceDigestHex:   digestHex("reference"),
		MFAMode:              fundauth.MFAEvidenceRequired,
		MFADigestHex:         digestHex("mfa"),
		EligibilityMode:      fundauth.EligibilityEvidenceRequired,
		EligibilityDigestHex: digestHex("eligibility"),
		PolicyDigestHex:      digestHex("policy"),
		NonceDigestHex:       digestHex("nonce-commit"),
		IssuedAtBlock:        100,
		IssuedAtUnix:         1_799_999_900,
		LowerBlock:           100,
		UpperBlock:           200,
		Expiry:               fundauth.ExpiryCoordinate{Kind: fundauth.ExpiryAtBlock, Value: 180},
	}
	bindingFor := func(value fundauth.FundAuthorization) fundauth.TransactionBinding {
		return fundauth.TransactionBinding{
			Domain: value.Domain, Version: value.Version, ChainID: value.ChainID, AccountID: value.AccountID,
			SignerKeyID: value.SignerKeyID, SignerKeyEpoch: value.SignerKeyEpoch, SourceID: value.SourceID, TypeURL: value.TypeURL,
			Phase: value.Phase, Effect: value.Effect, MessageDigestHex: value.MessageDigestHex,
			Amounts: append([]fundauth.Amount(nil), value.Amounts...), Parties: append([]fundauth.PartyBinding(nil), value.Parties...),
			CaseDigestHex: value.CaseDigestHex, OrderDigestHex: value.OrderDigestHex, ReferenceDigestHex: value.ReferenceDigestHex,
			MFAMode: value.MFAMode, MFADigestHex: value.MFADigestHex, EligibilityMode: value.EligibilityMode,
			EligibilityDigestHex: value.EligibilityDigestHex, PolicyDigestHex: value.PolicyDigestHex, NonceDigestHex: value.NonceDigestHex,
			IssuedAtBlock: value.IssuedAtBlock, IssuedAtUnix: value.IssuedAtUnix, LowerBlock: value.LowerBlock, UpperBlock: value.UpperBlock, Expiry: value.Expiry,
			CurrentBlock: 150, CurrentTime: time.Unix(1_800_000_000, 0), MaxClockSkew: 30 * time.Second, MaxLifetime: time.Hour,
		}
	}
	policy := fundauth.AuthorizationPolicyContext{
		AccountID:                 auth.AccountID,
		AccountState:              fundauth.AuthorizationAccountActive,
		AccountRevision:           11,
		AuthorizedAccountRevision: 11,
		KeyEpoch:                  auth.SignerKeyEpoch,
		PolicyDigestHex:           auth.PolicyDigestHex,
		PolicyRevision:            13,
		AuthorizedPolicyRevision:  13,
		FactorMode:                fundauth.FactorPossessionPlusMFA,
		MFADigestHex:              auth.MFADigestHex,
		EligibilityMode:           fundauth.EligibilityEvidenceRequired,
		EligibilityState:          fundauth.EligibilityEligible,
		EligibilityDigestHex:      auth.EligibilityDigestHex,
		EligibilityFreshAtBlock:   150,
		EligibilityFreshAtTime:    time.Unix(1_800_000_000, 0),
		CurrentBlock:              150,
		CurrentTime:               time.Unix(1_800_000_000, 0),
	}
	seed := sha256.Sum256([]byte("fundauth keeper integration key"))
	privateKey := ed25519.NewKeyFromSeed(seed[:])
	resolver := integrationKeyResolver{key: fundauth.ResolvedPossessionKey{
		AccountID: auth.AccountID, KeyID: auth.SignerKeyID, Epoch: auth.SignerKeyEpoch,
		PublicKey: privateKey.Public().(ed25519.PublicKey), Active: true,
	}}
	sign := func(value fundauth.FundAuthorization) (fundauth.SignedAuthorization, fundauth.Digest) {
		t.Helper()
		signBytes, authDigest, signErr := fundauth.CanonicalSignBytes(value)
		require.NoError(t, signErr)
		return fundauth.SignedAuthorization{Authorization: value, Signature: ed25519.Sign(privateKey, signBytes)}, authDigest
	}
	decodeDigest := func(value string) fundauth.Digest {
		t.Helper()
		decoded, decodeErr := hex.DecodeString(value)
		require.NoError(t, decodeErr)
		var digest fundauth.Digest
		copy(digest[:], decoded)
		return digest
	}

	signed, authDigest := sign(auth)
	binding := bindingFor(auth)
	callbackCalls := 0
	tampered := signed
	tampered.Signature = append([]byte(nil), signed.Signature...)
	tampered.Signature[0] ^= 0xff
	require.ErrorIs(t, fundauth.VerifyPolicyAndConsume(ctx, tampered, fundauth.DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error {
		callbackCalls++
		return nil
	}), fundauth.ErrInvalidSignature)
	wrongBinding := binding
	wrongBinding.MessageDigestHex = digestHex("wrong-message")
	require.Error(t, fundauth.VerifyPolicyAndConsume(ctx, signed, fundauth.DefaultRegistry(), resolver, wrongBinding, policy, consumer, func(context.Context) error {
		callbackCalls++
		return nil
	}))
	require.Zero(t, callbackCalls)

	protectedKey := []byte("protected/committed")
	require.NoError(t, fundauth.VerifyPolicyAndConsume(ctx, signed, fundauth.DefaultRegistry(), resolver, binding, policy, consumer, func(callbackCtx context.Context) error {
		callbackCalls++
		sdk.UnwrapSDKContext(callbackCtx).KVStore(storeKey).Set(protectedKey, []byte("committed"))
		return nil
	}))
	require.Equal(t, []byte("committed"), sdkCtx.KVStore(storeKey).Get(protectedKey))
	storedDigest, found, err := consumer.AuthorizationDigest(ctx, auth.AccountID, decodeDigest(auth.NonceDigestHex))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, authDigest, storedDigest)

	require.ErrorIs(t, fundauth.VerifyPolicyAndConsume(ctx, signed, fundauth.DefaultRegistry(), resolver, binding, policy, consumer, func(context.Context) error {
		callbackCalls++
		return nil
	}), fundauthkeeper.ErrAuthorizationReplay)
	require.Equal(t, 1, callbackCalls)

	rollbackAuth := auth
	rollbackAuth.NonceDigestHex = digestHex("nonce-rollback")
	rollbackSigned, rollbackAuthDigest := sign(rollbackAuth)
	rollbackBinding := bindingFor(rollbackAuth)
	rollbackKey := []byte("protected/rollback")
	callbackErr := errors.New("protected operation failed")
	require.ErrorIs(t, fundauth.VerifyPolicyAndConsume(ctx, rollbackSigned, fundauth.DefaultRegistry(), resolver, rollbackBinding, policy, consumer, func(callbackCtx context.Context) error {
		sdk.UnwrapSDKContext(callbackCtx).KVStore(storeKey).Set(rollbackKey, []byte("discarded"))
		return callbackErr
	}), callbackErr)
	require.Nil(t, sdkCtx.KVStore(storeKey).Get(rollbackKey))
	_, found, err = consumer.AuthorizationDigest(ctx, rollbackAuth.AccountID, decodeDigest(rollbackAuth.NonceDigestHex))
	require.NoError(t, err)
	require.False(t, found)

	require.NoError(t, fundauth.VerifyPolicyAndConsume(ctx, rollbackSigned, fundauth.DefaultRegistry(), resolver, rollbackBinding, policy, consumer, func(callbackCtx context.Context) error {
		sdk.UnwrapSDKContext(callbackCtx).KVStore(storeKey).Set(rollbackKey, []byte("committed-after-retry"))
		return nil
	}))
	require.Equal(t, []byte("committed-after-retry"), sdkCtx.KVStore(storeKey).Get(rollbackKey))
	storedDigest, found, err = consumer.AuthorizationDigest(ctx, rollbackAuth.AccountID, decodeDigest(rollbackAuth.NonceDigestHex))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, rollbackAuthDigest, storedDigest)
}
