package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/issuancepolicy/types"
)

func TestMsgServerLifecycleAndQueries(t *testing.T) {
	k, ctx := setupIssuancePolicyKeeper(t)
	msgServer := NewMsgServerImpl(k)
	queryServer := GRPCQuerier{Keeper: k}

	policy := types.IssuancePolicy{
		PolicyID:             "policy-main",
		Status:               string(types.PolicyStatusPaused),
		ActiveVerifierScope:  "verifier-main",
		MintUnitsPerProof:    25,
		DailyCap:             1000,
		EpochCap:             1000,
		CreatedAtHeight:      ctx.BlockHeight(),
		GovernanceProposalID: 11,
	}

	_, err := msgServer.UpsertPolicy(sdk.WrapSDKContext(ctx), &types.MsgUpsertPolicy{
		Authority: k.GetAuthority(),
		Policy:    policy,
	})
	require.NoError(t, err)

	_, err = msgServer.SetActivePolicy(sdk.WrapSDKContext(ctx), &types.MsgSetActivePolicy{
		Authority: k.GetAuthority(),
		PolicyID:  policy.PolicyID,
	})
	require.NoError(t, err)

	activeResp, err := queryServer.ActivePolicy(ctx, &types.QueryActivePolicyRequest{})
	require.NoError(t, err)
	require.NotNil(t, activeResp.Policy)
	require.Equal(t, policy.PolicyID, activeResp.Policy.PolicyID)

	_, err = msgServer.PausePolicy(sdk.WrapSDKContext(ctx), &types.MsgPausePolicy{
		Authority: k.GetAuthority(),
		PolicyID:  policy.PolicyID,
	})
	require.NoError(t, err)
	require.True(t, k.IsMintingPaused(ctx))

	_, err = msgServer.ResumePolicy(sdk.WrapSDKContext(ctx), &types.MsgResumePolicy{
		Authority: k.GetAuthority(),
		PolicyID:  policy.PolicyID,
	})
	require.NoError(t, err)
	require.False(t, k.IsMintingPaused(ctx))

	units, err := k.RecordVerifiedProof(ctx, "proof-main", "virt1proofacct", "verifier-main", "veid-1.3", 91)
	require.NoError(t, err)
	require.Equal(t, uint64(25), units)

	recordResp, err := queryServer.ProofMintRecord(ctx, &types.QueryProofMintRecordRequest{ProofID: "proof-main"})
	require.NoError(t, err)
	require.Equal(t, string(types.ProofMintStatusRecorded), recordResp.Record.Status)

	recordsResp, err := queryServer.ProofMintRecords(ctx, &types.QueryProofMintRecordsRequest{})
	require.NoError(t, err)
	require.Len(t, recordsResp.Records, 1)

	_, err = msgServer.DeprecatePolicy(sdk.WrapSDKContext(ctx), &types.MsgDeprecatePolicy{
		Authority: k.GetAuthority(),
		PolicyID:  policy.PolicyID,
	})
	require.NoError(t, err)

	_, found := k.GetActivePolicy(ctx)
	require.False(t, found)
}
