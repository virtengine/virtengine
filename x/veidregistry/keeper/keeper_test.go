package keeper

import (
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
		VerifierID:       "verifier-1",
		SpecVersion:      "veid-1.2",
		WeightsSHA256:    "sha256:abc",
		ActivationHeight: ctx.BlockHeight(),
		Status:           string(types.VerifierStatusApproved),
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
	require.Equal(t, "sha256:abc", info.WeightsSHA256)
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
