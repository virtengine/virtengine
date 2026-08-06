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

	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

func setupIssuancePolicyKeeper(t *testing.T) (Keeper, sdk.Context) {
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
		Height: 250,
		Time:   time.Now().UTC(),
	}, false, log.NewNopLogger())

	require.NoError(t, k.SetParams(ctx, types.DefaultParams()))

	return k, ctx
}

func TestRecordVerifiedProof(t *testing.T) {
	k, ctx := setupIssuancePolicyKeeper(t)

	policy := types.DefaultIssuancePolicy()
	policy.PolicyID = "issuance-1"
	policy.MintUnitsPerProof = 117667
	policy.DailyCap = 500000
	policy.EpochCap = 500000
	policy.CreatedAtHeight = ctx.BlockHeight()

	require.NoError(t, k.SetPolicy(ctx, policy))
	require.NoError(t, k.SetActivePolicy(ctx, policy.PolicyID))

	units, err := k.RecordVerifiedProof(ctx, "proof-1", "virt1proofacct", "verifier-1", "veid-1.2", 92)
	require.NoError(t, err)
	require.Equal(t, uint64(117667), units)

	record, found := k.GetProofMintRecord(ctx, "proof-1")
	require.True(t, found)
	require.Equal(t, string(types.ProofMintStatusRecorded), record.Status)

	counters := k.GetCounters(ctx)
	require.Equal(t, uint64(117667), counters.MintedToday)
	require.Equal(t, uint64(117667), counters.MintedThisEpoch)

	replayedUnits, err := k.RecordVerifiedProof(ctx, "proof-1", "virt1proofacct", "verifier-1", "veid-1.2", 92)
	require.NoError(t, err)
	require.Equal(t, units, replayedUnits)

	replayedCounters := k.GetCounters(ctx)
	require.Equal(t, counters, replayedCounters)
}

func TestRecordVerifiedProofVerifierScopeMismatch(t *testing.T) {
	k, ctx := setupIssuancePolicyKeeper(t)

	policy := types.DefaultIssuancePolicy()
	policy.PolicyID = "issuance-2"
	policy.ActiveVerifierScope = "verifier-1"
	policy.MintUnitsPerProof = 10
	policy.CreatedAtHeight = ctx.BlockHeight()

	require.NoError(t, k.SetPolicy(ctx, policy))
	require.NoError(t, k.SetActivePolicy(ctx, policy.PolicyID))

	units, err := k.RecordVerifiedProof(ctx, "proof-2", "virt1proofacct", "verifier-2", "veid-1.2", 90)
	require.NoError(t, err)
	require.Zero(t, units)

	record, found := k.GetProofMintRecord(ctx, "proof-2")
	require.True(t, found)
	require.Equal(t, string(types.ProofMintStatusVerifierMismatch), record.Status)
}

func TestProposalHandlerAndQuery(t *testing.T) {
	k, ctx := setupIssuancePolicyKeeper(t)

	handler := NewProposalHandler(k)
	policy := types.IssuancePolicy{
		PolicyID:            "policy-governed",
		Status:              string(types.PolicyStatusActive),
		ActiveVerifierScope: "verifier-3",
		MintUnitsPerProof:   25,
		DailyCap:            1000,
		EpochCap:            1000,
	}

	require.NoError(t, handler(ctx, govv1beta1.Content(&types.UpsertPolicyProposal{
		Title:       "Add policy",
		Description: "Create a governed issuance policy",
		Policy:      policy,
	})))
	require.NoError(t, handler(ctx, govv1beta1.Content(&types.SetActivePolicyProposal{
		Title:       "Activate policy",
		Description: "Switch active issuance policy",
		PolicyID:    policy.PolicyID,
	})))

	querier := GRPCQuerier{Keeper: k}
	activeResp, err := querier.ActivePolicy(ctx, &types.QueryActivePolicyRequest{})
	require.NoError(t, err)
	require.NotNil(t, activeResp.Policy)
	require.Equal(t, policy.PolicyID, activeResp.Policy.PolicyID)

	require.NoError(t, handler(ctx, govv1beta1.Content(&types.UpdateParamsProposal{
		Title:       "Tighten params",
		Description: "Adjust issuance caps",
		Params: types.Params{
			EpochLengthBlocks:     200,
			MaxMintUnitsPerProof:  500,
			MaxDailyCap:           10000,
			MaxEpochCap:           10000,
			EmergencyPauseEnabled: false,
		},
	})))

	paramsResp, err := querier.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, int64(200), paramsResp.Params.EpochLengthBlocks)
	require.False(t, paramsResp.Params.EmergencyPauseEnabled)
}

func closeStoreIfNeeded(stateStore store.CommitMultiStore) {
	t, ok := any(stateStore).(interface{ Close() error })
	if ok {
		_ = t.Close()
	}
}
