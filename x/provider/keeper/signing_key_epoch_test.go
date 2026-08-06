// Copyright 2026 VirtEngine contributors.
// SPDX-License-Identifier: Apache-2.0

package keeper_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"math/big"
	"testing"
	"time"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	decredsecp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
	"github.com/stretchr/testify/require"

	types "github.com/virtengine/virtengine/sdk/go/node/provider/v1beta4"
	"github.com/virtengine/virtengine/sdk/go/testutil"
	"github.com/virtengine/virtengine/x/provider/keeper"
)

func TestProviderSigningKeyEpochRotationAndRevocation(t *testing.T) {
	ctx, k := setupKeeper(t)
	ctx = ctx.WithChainID("virtengine-test-1").WithBlockHeight(100).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	provider := testutil.Provider(t)
	require.NoError(t, k.Create(ctx, provider))
	owner, err := sdk.AccAddressFromBech32(provider.Owner)
	require.NoError(t, err)

	oldPublic, oldPrivate, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	require.NoError(t, k.SetProviderPublicKey(ctx, owner, oldPublic, types.PublicKeyTypeEd25519))

	active, found := k.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, uint64(1), active.Epoch)
	require.NotEmpty(t, active.KeyID)
	require.Equal(t, int64(100), active.ActivatedAtHeight)
	preactivation := active
	preactivation.ActivatedAtHeight = ctx.BlockHeight() + 1
	preactivation.ActivatedAtUnix = ctx.BlockTime().Add(time.Second).Unix()
	require.False(t, preactivation.IsValidAt(ctx.BlockHeight(), ctx.BlockTime()))
	expired := active
	expired.ExpiresAtHeight = ctx.BlockHeight() - 1
	expired.ExpiresAtUnix = ctx.BlockTime().Add(-time.Second).Unix()
	require.False(t, expired.IsValidAt(ctx.BlockHeight(), ctx.BlockTime()))

	newPublic, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	rotationBytes, err := types.ProviderKeyRotationSignBytes(types.ProviderKeyRotationPayload{
		ChainID:          ctx.ChainID(),
		Provider:         owner.String(),
		OldKeyID:         active.KeyID,
		OldEpoch:         active.Epoch,
		NewKeyType:       types.PublicKeyTypeEd25519,
		NewPublicKey:     newPublic,
		NewEpoch:         2,
		ActivationHeight: 101,
		ActivationUnix:   ctx.BlockTime().Unix() + 1,
		OverlapEndHeight: 101 + types.ProviderSigningKeyOverlapBlocks,
		OverlapEndUnix:   ctx.BlockTime().Unix() + 1 + types.ProviderSigningKeyOverlapSeconds,
		SignatureVersion: types.ProviderKeyRotationSignatureVersionV1,
	})
	require.NoError(t, err)
	proof := ed25519.Sign(oldPrivate, rotationBytes)

	rotationCtx := ctx.WithBlockHeight(101).WithBlockTime(ctx.BlockTime().Add(time.Second))
	require.NoError(t, k.RotateProviderPublicKey(rotationCtx, owner, newPublic, types.PublicKeyTypeEd25519, proof))

	rotated, found := k.GetProviderPublicKeyRecord(rotationCtx, owner)
	require.True(t, found)
	require.Equal(t, uint64(2), rotated.Epoch)
	require.Equal(t, active.KeyID, rotated.PreviousKeyID)

	oldEpoch, found := k.GetProviderSigningKey(rotationCtx, owner, active.KeyID, active.Epoch)
	require.True(t, found)
	require.Equal(t, int64(101)+types.ProviderSigningKeyOverlapBlocks, oldEpoch.RetiredAtHeight)
	require.True(t, oldEpoch.IsValidAt(rotationCtx.BlockHeight(), rotationCtx.BlockTime()))

	afterOverlap := rotationCtx.WithBlockHeight(oldEpoch.RetiredAtHeight + 1).WithBlockTime(time.Unix(oldEpoch.RetiredAtUnix+1, 0).UTC())
	require.False(t, oldEpoch.IsValidAt(afterOverlap.BlockHeight(), afterOverlap.BlockTime()))

	require.NoError(t, k.RevokeProviderSigningKey(rotationCtx, owner, rotated.KeyID))
	revoked, found := k.GetProviderSigningKey(rotationCtx, owner, rotated.KeyID, rotated.Epoch)
	require.True(t, found)
	require.NotZero(t, revoked.RevokedAtHeight)
	require.False(t, revoked.IsValidAt(rotationCtx.BlockHeight(), rotationCtx.BlockTime()))
}

func TestProviderSigningKeyRejectsX25519AndUnprovedOverwrite(t *testing.T) {
	ctx, k := setupKeeper(t)
	ctx = ctx.WithChainID("virtengine-test-1").WithBlockHeight(10).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	provider := testutil.Provider(t)
	require.NoError(t, k.Create(ctx, provider))
	owner, err := sdk.AccAddressFromBech32(provider.Owner)
	require.NoError(t, err)

	first, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	second, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	require.NoError(t, k.SetProviderPublicKey(ctx, owner, first, types.PublicKeyTypeEd25519))
	require.Error(t, k.SetProviderPublicKey(ctx, owner, second, types.PublicKeyTypeEd25519))
	require.Error(t, k.RotateProviderPublicKey(ctx, owner, second, types.PublicKeyTypeEd25519, []byte("not-a-signature")))

	x25519 := make([]byte, 32)
	require.Error(t, k.RotateProviderPublicKey(ctx, owner, x25519, types.PublicKeyTypeX25519, make([]byte, ed25519.SignatureSize)))
}

func TestMigrateLegacyProviderSigningKeyEpoch(t *testing.T) {
	ctx, k := setupKeeper(t)
	ctx = ctx.WithChainID("virtengine-test-1").WithBlockHeight(50).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	provider := testutil.Provider(t)
	require.NoError(t, k.Create(ctx, provider))
	owner, err := sdk.AccAddressFromBech32(provider.Owner)
	require.NoError(t, err)
	publicKey, _, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	legacy := types.ProviderPublicKeyRecord{
		PublicKey:     publicKey,
		KeyType:       types.PublicKeyTypeEd25519,
		UpdatedAt:     40,
		RotationCount: 2,
	}
	store := ctx.KVStore(k.StoreKey())
	legacyBytes, err := json.Marshal(&legacy)
	require.NoError(t, err)
	store.Set(keeper.ProviderPublicKeyKey(owner), legacyBytes)

	require.NoError(t, k.MigrateSigningKeyEpochs(ctx))
	current, found := k.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	require.Equal(t, uint64(3), current.Epoch)
	require.NotEmpty(t, current.KeyID)
	require.Equal(t, int64(40), current.ActivatedAtHeight)
	historical, found := k.GetProviderSigningKey(ctx, owner, current.KeyID, current.Epoch)
	require.True(t, found)
	require.Equal(t, current, historical)
}

func TestProviderSigningKeySecp256k1RotationRejectsHighS(t *testing.T) {
	ctx, k := setupKeeper(t)
	ctx = ctx.WithChainID("virtengine-test-1").WithBlockHeight(100).WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	provider := testutil.Provider(t)
	require.NoError(t, k.Create(ctx, provider))
	owner, err := sdk.AccAddressFromBech32(provider.Owner)
	require.NoError(t, err)

	oldPrivate := cosmossecp256k1.GenPrivKey()
	require.NoError(t, k.SetProviderPublicKey(ctx, owner, oldPrivate.PubKey().Bytes(), types.PublicKeyTypeSecp256k1))
	active, found := k.GetProviderPublicKeyRecord(ctx, owner)
	require.True(t, found)
	newPrivate := cosmossecp256k1.GenPrivKey()
	rotationCtx := ctx.WithBlockHeight(101).WithBlockTime(ctx.BlockTime().Add(time.Second))
	rotationBytes, err := types.ProviderKeyRotationSignBytes(types.ProviderKeyRotationPayload{
		ChainID:          rotationCtx.ChainID(),
		Provider:         owner.String(),
		OldKeyID:         active.KeyID,
		OldEpoch:         active.Epoch,
		NewKeyType:       types.PublicKeyTypeSecp256k1,
		NewPublicKey:     newPrivate.PubKey().Bytes(),
		NewEpoch:         active.Epoch + 1,
		ActivationHeight: rotationCtx.BlockHeight(),
		ActivationUnix:   rotationCtx.BlockTime().Unix(),
		OverlapEndHeight: rotationCtx.BlockHeight() + types.ProviderSigningKeyOverlapBlocks,
		OverlapEndUnix:   rotationCtx.BlockTime().Unix() + types.ProviderSigningKeyOverlapSeconds,
		SignatureVersion: types.ProviderKeyRotationSignatureVersionV1,
	})
	require.NoError(t, err)
	proof, err := oldPrivate.Sign(rotationBytes)
	require.NoError(t, err)

	highS := append([]byte(nil), proof...)
	sValue := new(big.Int).SetBytes(highS[32:])
	sValue.Sub(decredsecp256k1.S256().N, sValue)
	sBytes := sValue.Bytes()
	for i := 32; i < 64; i++ {
		highS[i] = 0
	}
	copy(highS[64-len(sBytes):], sBytes)
	require.ErrorIs(t, k.RotateProviderPublicKey(rotationCtx, owner, newPrivate.PubKey().Bytes(), types.PublicKeyTypeSecp256k1, highS), types.ErrInvalidRotationSignature)
	require.NoError(t, k.RotateProviderPublicKey(rotationCtx, owner, newPrivate.PubKey().Bytes(), types.PublicKeyTypeSecp256k1, proof))
}
