package issuancepolicy

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

	"github.com/virtengine/virtengine/x/issuancepolicy/keeper"
	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

func TestGenesisRoundTripIncludesProofRecords(t *testing.T) {
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
	ctx := sdk.NewContext(stateStore, cmtproto.Header{Height: 250, Time: time.Now().UTC()}, false, log.NewNopLogger())

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))
	policy := types.DefaultIssuancePolicy()
	policy.PolicyID = "policy-genesis"
	policy.Status = string(types.PolicyStatusActive)
	policy.MintUnitsPerProof = 33
	policy.CreatedAtHeight = ctx.BlockHeight()
	require.NoError(t, k.SetPolicy(ctx, policy))
	require.NoError(t, k.SetActivePolicy(ctx, policy.PolicyID))
	_, err := k.RecordVerifiedProof(ctx, "proof-genesis", "virt1genesis", "*", "veid-1.5", 88)
	require.NoError(t, err)

	exported := ExportGenesis(ctx, k)
	require.Len(t, exported.Policies, 1)
	require.Len(t, exported.ProofRecords, 1)
	require.Equal(t, policy.PolicyID, exported.ActivePolicyID)

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
	ctx2 := sdk.NewContext(stateStore2, cmtproto.Header{Height: 250, Time: time.Now().UTC()}, false, log.NewNopLogger())

	InitGenesis(ctx2, k2, exported)

	roundTrip := ExportGenesis(ctx2, k2)
	require.Equal(t, exported.Params, roundTrip.Params)
	require.Equal(t, exported.Policies, roundTrip.Policies)
	require.Equal(t, exported.ProofRecords, roundTrip.ProofRecords)
	require.Equal(t, exported.ActivePolicyID, roundTrip.ActivePolicyID)
}
