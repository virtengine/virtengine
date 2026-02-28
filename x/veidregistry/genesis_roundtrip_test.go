package veidregistry

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
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veidregistry/keeper"
	"github.com/virtengine/virtengine/x/veidregistry/types"
)

func TestGenesisRoundTripPreservesRegistryState(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	db := dbm.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() {
		if closer, ok := stateStore.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	})
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	require.NoError(t, stateStore.LoadLatestVersion())

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	k := keeper.NewKeeper(cdc, storeKey, "authority")
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 120, Time: time.Now().UTC()}, false, log.NewNopLogger())

	require.NoError(t, k.SetParams(ctx, types.Params{
		MinimumReadyValidators:            2,
		MinimumIndependentImplementations: 2,
		AllowLegacyMirroring:              false,
	}))
	require.NoError(t, k.SetVerifierVersion(ctx, types.VerifierVersion{
		VerifierID:           "verifier-genesis",
		SpecVersion:          "veid-1.5",
		WeightsSHA256:        "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		ActivationHeight:     ctx.BlockHeight(),
		Status:               string(types.VerifierStatusApproved),
		GovernanceProposalID: 42,
	}))
	require.NoError(t, k.SetValidatorReadiness(ctx, types.ValidatorReadiness{
		ValidatorAddress:  "virtvaloper1genesis",
		VerifierID:        "verifier-genesis",
		ConformancePassed: true,
		ImplementationID:  "impl-a",
		Organization:      "org-a",
		ReportedHeight:    ctx.BlockHeight(),
	}))
	require.NoError(t, k.SetActiveVerifier(ctx, types.ActiveVerifierPointer{
		VerifierID:        "verifier-genesis",
		SpecVersion:       "veid-1.5",
		ActivatedAtHeight: ctx.BlockHeight(),
	}))

	exported := ExportGenesis(ctx, k)
	require.Len(t, exported.Verifiers, 1)
	require.Len(t, exported.ValidatorReadiness, 1)
	require.NotNil(t, exported.ActiveVerifier)

	storeKey2 := storetypes.NewKVStoreKey(types.StoreKey)
	db2 := dbm.NewMemDB()
	stateStore2 := store.NewCommitMultiStore(db2, log.NewNopLogger(), storemetrics.NewNoOpMetrics())
	t.Cleanup(func() {
		if closer, ok := stateStore2.(interface{ Close() error }); ok {
			_ = closer.Close()
		}
	})
	stateStore2.MountStoreWithDB(storeKey2, storetypes.StoreTypeIAVL, db2)
	require.NoError(t, stateStore2.LoadLatestVersion())
	k2 := keeper.NewKeeper(cdc, storeKey2, "authority")
	ctx2 := sdk.NewContext(stateStore2, cmtproto.Header{Height: 120, Time: time.Now().UTC()}, false, log.NewNopLogger())

	InitGenesis(ctx2, k2, exported)

	roundTrip := ExportGenesis(ctx2, k2)
	require.Equal(t, exported.Params, roundTrip.Params)
	require.Equal(t, exported.Verifiers, roundTrip.Verifiers)
	require.Equal(t, exported.ValidatorReadiness, roundTrip.ValidatorReadiness)
	require.Equal(t, exported.ActiveVerifier, roundTrip.ActiveVerifier)
}
