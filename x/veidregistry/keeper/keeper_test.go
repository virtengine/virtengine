package keeper

import (
	"bytes"
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
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veidregistry/types"
)

func setupRegistryKeeper(t *testing.T) (Keeper, sdk.Context) {
	t.Helper()

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() { closeStoreIfNeeded(stateStore) })
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)

	k := NewKeeper(cdc, storeKey, "authority")
	ctx := sdk.NewContext(stateStore, cmtproto.Header{
		Height: 100,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	require.NoError(t, k.SetParams(ctx, types.Params{
		MinimumReadyValidators:            2,
		MinimumIndependentImplementations: 2,
		AllowLegacyMirroring:              true,
	}))

	return k, ctx
}

func TestActivateReadyVerifiers(t *testing.T) {
	k, ctx := setupRegistryKeeper(t)

	require.NoError(t, k.SetVerifierVersion(ctx, types.VerifierVersion{
		VerifierID:        "verifier-1",
		SpecVersion:       "veid-1.2",
		WeightsSHA256:     registryHashA,
		TestVectorsSHA256: registryHashB,
		ImageHash:         registryHashC,
		ModelManifestHash: registryHashA,
		ActivationHeight:  ctx.BlockHeight(),
		Status:            string(types.VerifierStatusApproved),
	}))

	require.NoError(t, k.SetValidatorReadiness(ctx, types.ValidatorReadiness{
		ValidatorAddress:  "virtvaloper1",
		VerifierID:        "verifier-1",
		ConformancePassed: true,
		ImplementationID:  "impl-a",
		Organization:      "org-a",
		ReportedHeight:    ctx.BlockHeight(),
	}))
	require.NoError(t, k.SetValidatorReadiness(ctx, types.ValidatorReadiness{
		ValidatorAddress:  "virtvaloper2",
		VerifierID:        "verifier-1",
		ConformancePassed: true,
		ImplementationID:  "impl-b",
		Organization:      "org-b",
		ReportedHeight:    ctx.BlockHeight(),
	}))

	require.NoError(t, k.ActivateReadyVerifiers(ctx))

	active, found := k.GetActiveVerifier(ctx)
	require.True(t, found)
	require.Equal(t, "verifier-1", active.VerifierID)
	require.Equal(t, "veid-1.2", active.SpecVersion)

	info, found := k.GetActiveVerifierInfo(ctx)
	require.True(t, found)
	require.Equal(t, registryHashA, info.WeightsSHA256)
}

func TestGetActiveVerifierInfoStrictProjection(t *testing.T) {
	t.Run("no pointer", func(t *testing.T) {
		k, ctx := setupRegistryKeeper(t)
		info, found, err := k.GetActiveVerifierInfoStrict(ctx)
		require.NoError(t, err)
		require.False(t, found)
		require.Empty(t, info)
	})

	t.Run("exact active projection supports prefixed and raw commitments", func(t *testing.T) {
		k, ctx := setupRegistryKeeper(t)
		verifier := activeProjectionVerifier(ctx)
		verifier.ImageHash = registryHashC[len("sha256:"):]
		require.NoError(t, k.SetVerifierVersion(ctx, verifier))
		require.NoError(t, k.SetActiveVerifier(ctx, types.ActiveVerifierPointer{
			VerifierID: verifier.VerifierID, SpecVersion: verifier.SpecVersion, ActivatedAtHeight: ctx.BlockHeight(),
		}))
		beforePointer := bytes.Clone(ctx.KVStore(k.skey).Get(types.ActiveVerifierKey()))
		beforeRecord := bytes.Clone(ctx.KVStore(k.skey).Get(types.VerifierVersionKey(verifier.VerifierID)))

		info, found, err := k.GetActiveVerifierInfoStrict(ctx)

		require.NoError(t, err)
		require.True(t, found)
		require.Equal(t, string(types.VerifierStatusActive), info.Status)
		require.Equal(t, verifier.ImageHash, info.ImageHash)
		require.Equal(t, verifier.ModelManifestHash, info.ModelManifestHash)
		require.Equal(t, beforePointer, ctx.KVStore(k.skey).Get(types.ActiveVerifierKey()))
		require.Equal(t, beforeRecord, ctx.KVStore(k.skey).Get(types.VerifierVersionKey(verifier.VerifierID)))
	})

	tests := []struct {
		name   string
		mutate func(Keeper, sdk.Context, types.VerifierVersion)
	}{
		{
			name: "missing record",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				ctx.KVStore(k.skey).Delete(types.VerifierVersionKey(verifier.VerifierID))
			},
		},
		{
			name: "inactive record",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				verifier.Status = string(types.VerifierStatusDeprecated)
				require.NoError(t, k.SetVerifierVersion(ctx, verifier))
			},
		},
		{
			name: "pointer spec mismatch",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				pointer := types.ActiveVerifierPointer{VerifierID: verifier.VerifierID, SpecVersion: "veid-9.9", ActivatedAtHeight: verifier.ActivationHeight}
				bz, err := json.Marshal(pointer)
				require.NoError(t, err)
				ctx.KVStore(k.skey).Set(types.ActiveVerifierKey(), bz)
			},
		},
		{
			name: "future activation",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				verifier.ActivationHeight = ctx.BlockHeight() + 1
				require.NoError(t, k.SetVerifierVersion(ctx, verifier))
				pointer := types.ActiveVerifierPointer{VerifierID: verifier.VerifierID, SpecVersion: verifier.SpecVersion, ActivatedAtHeight: verifier.ActivationHeight}
				bz, err := json.Marshal(pointer)
				require.NoError(t, err)
				ctx.KVStore(k.skey).Set(types.ActiveVerifierKey(), bz)
			},
		},
		{
			name: "malformed commitment",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				verifier.ImageHash = "sha256:not-a-digest"
				require.NoError(t, k.SetVerifierVersion(ctx, verifier))
			},
		},
		{
			name: "uppercase prefix",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				verifier.ImageHash = "SHA256:" + registryHashC[len("sha256:"):]
				require.NoError(t, k.SetVerifierVersion(ctx, verifier))
			},
		},
		{
			name: "uppercase hex",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				verifier.ImageHash = "sha256:AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
				require.NoError(t, k.SetVerifierVersion(ctx, verifier))
			},
		},
		{
			name: "surrounding whitespace",
			mutate: func(k Keeper, ctx sdk.Context, verifier types.VerifierVersion) {
				verifier.ImageHash = " " + registryHashC
				require.NoError(t, k.SetVerifierVersion(ctx, verifier))
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k, ctx := setupRegistryKeeper(t)
			verifier := activeProjectionVerifier(ctx)
			require.NoError(t, k.SetVerifierVersion(ctx, verifier))
			require.NoError(t, k.SetActiveVerifier(ctx, types.ActiveVerifierPointer{
				VerifierID: verifier.VerifierID, SpecVersion: verifier.SpecVersion, ActivatedAtHeight: ctx.BlockHeight(),
			}))
			storedVerifier, found := k.GetVerifierVersion(ctx, verifier.VerifierID)
			require.True(t, found)
			test.mutate(k, ctx, *storedVerifier)
			beforePointer := bytes.Clone(ctx.KVStore(k.skey).Get(types.ActiveVerifierKey()))
			beforeRecord := bytes.Clone(ctx.KVStore(k.skey).Get(types.VerifierVersionKey(storedVerifier.VerifierID)))

			_, found, err := k.GetActiveVerifierInfoStrict(ctx)

			require.Error(t, err)
			require.True(t, found)
			require.Equal(t, beforePointer, ctx.KVStore(k.skey).Get(types.ActiveVerifierKey()))
			require.Equal(t, beforeRecord, ctx.KVStore(k.skey).Get(types.VerifierVersionKey(storedVerifier.VerifierID)))
		})
	}
}

func TestSetActiveVerifierRejectsInvalidPointerWithoutMutation(t *testing.T) {
	tests := []struct {
		name    string
		pointer func(types.VerifierVersion, sdk.Context) types.ActiveVerifierPointer
	}{
		{
			name: "spec mismatch",
			pointer: func(verifier types.VerifierVersion, ctx sdk.Context) types.ActiveVerifierPointer {
				return types.ActiveVerifierPointer{VerifierID: verifier.VerifierID, SpecVersion: "veid-9.9.9", ActivatedAtHeight: ctx.BlockHeight()}
			},
		},
		{
			name: "zero height",
			pointer: func(verifier types.VerifierVersion, ctx sdk.Context) types.ActiveVerifierPointer {
				return types.ActiveVerifierPointer{VerifierID: verifier.VerifierID, SpecVersion: verifier.SpecVersion}
			},
		},
		{
			name: "future height",
			pointer: func(verifier types.VerifierVersion, ctx sdk.Context) types.ActiveVerifierPointer {
				return types.ActiveVerifierPointer{VerifierID: verifier.VerifierID, SpecVersion: verifier.SpecVersion, ActivatedAtHeight: ctx.BlockHeight() + 1}
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			k, ctx := setupRegistryKeeper(t)
			previous := activeProjectionVerifier(ctx)
			previous.VerifierID = "previous-verifier"
			require.NoError(t, k.SetVerifierVersion(ctx, previous))
			require.NoError(t, k.SetActiveVerifier(ctx, types.ActiveVerifierPointer{
				VerifierID: previous.VerifierID, SpecVersion: previous.SpecVersion, ActivatedAtHeight: ctx.BlockHeight(),
			}))
			candidate := activeProjectionVerifier(ctx)
			candidate.VerifierID = "candidate-verifier"
			require.NoError(t, k.SetVerifierVersion(ctx, candidate))
			beforePointer := bytes.Clone(ctx.KVStore(k.skey).Get(types.ActiveVerifierKey()))
			beforePrevious := bytes.Clone(ctx.KVStore(k.skey).Get(types.VerifierVersionKey(previous.VerifierID)))
			beforeCandidate := bytes.Clone(ctx.KVStore(k.skey).Get(types.VerifierVersionKey(candidate.VerifierID)))

			err := k.SetActiveVerifier(ctx, test.pointer(candidate, ctx))

			require.Error(t, err)
			require.Equal(t, beforePointer, ctx.KVStore(k.skey).Get(types.ActiveVerifierKey()))
			require.Equal(t, beforePrevious, ctx.KVStore(k.skey).Get(types.VerifierVersionKey(previous.VerifierID)))
			require.Equal(t, beforeCandidate, ctx.KVStore(k.skey).Get(types.VerifierVersionKey(candidate.VerifierID)))
		})
	}
}

func activeProjectionVerifier(ctx sdk.Context) types.VerifierVersion {
	return types.VerifierVersion{
		VerifierID:        "verifier-runtime",
		SpecVersion:       "veid-1.2.0",
		WeightsSHA256:     registryHashA,
		TestVectorsSHA256: registryHashB,
		ImageHash:         registryHashC,
		ModelManifestHash: registryHashA,
		ActivationHeight:  ctx.BlockHeight(),
		Status:            string(types.VerifierStatusApproved),
	}
}

func TestProposalHandlerAndQuery(t *testing.T) {
	k, ctx := setupRegistryKeeper(t)

	handler := NewProposalHandler(k)
	proposal := &types.AddVerifierVersionProposal{
		Title:       "Add verifier",
		Description: "Approve veid verifier v1.3",
		Verifier: types.VerifierVersion{
			VerifierID:       "verifier-2",
			SpecVersion:      "veid-1.3",
			WeightsSHA256:    "sha256:def",
			ActivationHeight: ctx.BlockHeight() + 10,
			Status:           string(types.VerifierStatusProposed),
		},
	}

	require.NoError(t, handler(ctx, govv1beta1.Content(proposal)))

	querier := GRPCQuerier{Keeper: k}
	resp, err := querier.Verifier(ctx, &types.QueryVerifierRequest{VerifierID: "verifier-2"})
	require.NoError(t, err)
	require.Equal(t, string(types.VerifierStatusApproved), resp.Verifier.Status)

	paramsProposal := &types.UpdateParamsProposal{
		Title:       "Tune registry",
		Description: "Raise readiness threshold",
		Params: types.Params{
			MinimumReadyValidators:            3,
			MinimumIndependentImplementations: 2,
			AllowLegacyMirroring:              false,
		},
	}
	require.NoError(t, handler(ctx, govv1beta1.Content(paramsProposal)))

	paramsResp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, uint32(3), paramsResp.Params.MinimumReadyValidators)
	require.False(t, paramsResp.Params.AllowLegacyMirroring)
}

func closeStoreIfNeeded(stateStore store.CommitMultiStore) {
	t, ok := any(stateStore).(interface{ Close() error })
	if ok {
		_ = t.Close()
	}
}
