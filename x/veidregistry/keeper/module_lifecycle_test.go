package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/virtengine/virtengine/x/veidregistry/types"
)

const (
	registryHashA = "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	registryHashB = "sha256:bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
	registryHashC = "sha256:cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
)

func TestMsgServerLifecycleAndQueries(t *testing.T) {
	k, ctx := setupRegistryKeeper(t)
	msgServer := NewMsgServerImpl(k)
	queryServer := GRPCQuerier{Keeper: k}

	_, err := msgServer.UpsertVerifierVersion(ctx, &types.MsgUpsertVerifierVersion{
		Authority: k.GetAuthority(),
		Verifier: types.VerifierVersion{
			VerifierID:    "verifier-main",
			SpecVersion:   "veid-1.3",
			WeightsSHA256: registryHashA,
			Status:        string(types.VerifierStatusProposed),
		},
	})
	require.NoError(t, err)

	_, err = msgServer.ApproveVerifierVersion(ctx, &types.MsgApproveVerifierVersion{
		Authority:            k.GetAuthority(),
		VerifierID:           "verifier-main",
		GovernanceProposalID: 9,
		ActivationHeight:     ctx.BlockHeight() + 5,
	})
	require.NoError(t, err)

	pending, err := queryServer.QueuedVerifiers(ctx, &types.QueryQueuedVerifiersRequest{})
	require.NoError(t, err)
	require.Len(t, pending.Verifiers, 1)
	require.Equal(t, string(types.VerifierStatusApproved), pending.Verifiers[0].Status)

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 5)
	_, err = msgServer.ReportValidatorReadiness(ctx, &types.MsgReportValidatorReadiness{
		ValidatorAddress:  "virtvaloper1alpha",
		VerifierID:        "verifier-main",
		ConformancePassed: true,
		ImplementationID:  "impl-a",
		Organization:      "org-a",
	})
	require.NoError(t, err)
	_, err = msgServer.ReportValidatorReadiness(ctx, &types.MsgReportValidatorReadiness{
		ValidatorAddress:  "virtvaloper1beta",
		VerifierID:        "verifier-main",
		ConformancePassed: true,
		ImplementationID:  "impl-b",
		Organization:      "org-b",
	})
	require.NoError(t, err)

	eligible, err := queryServer.EligibleVerifiers(ctx, &types.QueryEligibleVerifiersRequest{})
	require.NoError(t, err)
	require.Len(t, eligible.Verifiers, 1)
	require.Equal(t, "verifier-main", eligible.Verifiers[0].VerifierID)

	require.NoError(t, k.ActivateReadyVerifiers(ctx))

	active, err := queryServer.ActiveVerifier(ctx, &types.QueryActiveVerifierRequest{})
	require.NoError(t, err)
	require.NotNil(t, active.ActiveVerifier)
	require.Equal(t, "verifier-main", active.ActiveVerifier.VerifierID)

	info, found := k.GetActiveVerifierInfo(ctx)
	require.True(t, found)
	require.Equal(t, registryHashA, info.WeightsSHA256)

	_, err = msgServer.UpsertVerifierVersion(ctx, &types.MsgUpsertVerifierVersion{
		Authority: k.GetAuthority(),
		Verifier: types.VerifierVersion{
			VerifierID:    "verifier-cancel",
			SpecVersion:   "veid-1.4",
			WeightsSHA256: registryHashB,
			Status:        string(types.VerifierStatusProposed),
		},
	})
	require.NoError(t, err)
	_, err = msgServer.CancelVerifierVersion(ctx, &types.MsgCancelVerifierVersion{
		Authority:  k.GetAuthority(),
		VerifierID: "verifier-cancel",
	})
	require.NoError(t, err)

	cancelled, found := k.GetVerifierVersion(ctx, "verifier-cancel")
	require.True(t, found)
	require.Equal(t, string(types.VerifierStatusCancelled), cancelled.Status)
}

func TestActivateReadyVerifiersChoosesEarliestEligibleVersion(t *testing.T) {
	k, ctx := setupRegistryKeeper(t)

	require.NoError(t, k.SetVerifierVersion(ctx, types.VerifierVersion{
		VerifierID:       "verifier-late",
		SpecVersion:      "veid-2.0",
		WeightsSHA256:    registryHashB,
		ActivationHeight: ctx.BlockHeight() + 20,
		Status:           string(types.VerifierStatusApproved),
	}))
	require.NoError(t, k.SetVerifierVersion(ctx, types.VerifierVersion{
		VerifierID:       "verifier-early",
		SpecVersion:      "veid-1.9",
		WeightsSHA256:    registryHashA,
		ActivationHeight: ctx.BlockHeight() + 10,
		Status:           string(types.VerifierStatusApproved),
	}))

	ctx = ctx.WithBlockHeight(ctx.BlockHeight() + 25)
	for _, report := range []types.ValidatorReadiness{
		{ValidatorAddress: "virtvaloper1", VerifierID: "verifier-late", ConformancePassed: true, ImplementationID: "impl-a", Organization: "org-a", ReportedHeight: ctx.BlockHeight()},
		{ValidatorAddress: "virtvaloper2", VerifierID: "verifier-late", ConformancePassed: true, ImplementationID: "impl-b", Organization: "org-b", ReportedHeight: ctx.BlockHeight()},
		{ValidatorAddress: "virtvaloper3", VerifierID: "verifier-early", ConformancePassed: true, ImplementationID: "impl-a", Organization: "org-a", ReportedHeight: ctx.BlockHeight()},
		{ValidatorAddress: "virtvaloper4", VerifierID: "verifier-early", ConformancePassed: true, ImplementationID: "impl-b", Organization: "org-b", ReportedHeight: ctx.BlockHeight()},
	} {
		require.NoError(t, k.SetValidatorReadiness(ctx, report))
	}

	eligible := k.EligibleVerifierVersions(ctx)
	require.Len(t, eligible, 2)
	require.Equal(t, "verifier-early", eligible[0].VerifierID)

	require.NoError(t, k.ActivateReadyVerifiers(ctx))

	active, found := k.GetActiveVerifier(ctx)
	require.True(t, found)
	require.Equal(t, "verifier-early", active.VerifierID)
}
